package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func anomalyRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	r := miniredis.RunT(t)
	old, enabled := common.RDB, common.RedisEnabled
	client := redis.NewClient(&redis.Options{Addr: r.Addr()})
	common.RDB = client
	common.RedisEnabled = true
	t.Cleanup(func() { client.Close(); common.RDB = old; common.RedisEnabled = enabled })
	return r
}
func TestAnomalyPolicyDefaultsValidationAndConcurrentSave(t *testing.T) {
	anomalyRedis(t)
	ctx := context.Background()
	p, err := GetAnomalyPolicy(ctx)
	require.NoError(t, err)
	require.True(t, p.Enabled)
	require.Equal(t, 10, p.SchemaThreshold)
	require.NoError(t, SaveAnomalyPolicy(ctx, p, 1))
	require.ErrorIs(t, SaveAnomalyPolicy(ctx, p, 2), ErrAnomalyConflict)
	p, err = GetAnomalyPolicy(ctx)
	require.NoError(t, err)
	p.SchemaThreshold = 0
	require.Error(t, SaveAnomalyPolicy(ctx, p, 1))
	p.SchemaThreshold = 10
	p.Enabled = false
	require.NoError(t, SaveAnomalyPolicy(ctx, p, 1))
	require.NoError(t, RecordAnomaly(ctx, 42, 1, "model", "invalid_schema"))
	events, err := ListAnomalies(ctx)
	require.NoError(t, err)
	require.Empty(t, events)
	audit, err := ListAnomalyAudit(ctx)
	require.NoError(t, err)
	require.Len(t, audit, 2)
	require.Equal(t, 1, audit[0].OperatorID)
}
func TestAnomalyConcurrentCountsAcknowledgeAndNewFailures(t *testing.T) {
	r := anomalyRedis(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); require.NoError(t, RecordAnomaly(ctx, 42, 1, "model", "invalid_schema")) }()
	}
	wg.Wait()
	events, err := ListAnomalies(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	e := events[0]
	require.Equal(t, 20, e.Count)
	require.True(t, e.Alert)
	require.ErrorIs(t, AcknowledgeAnomaly(ctx, e.ID, 19, 1), ErrAnomalyConflict)
	require.NoError(t, AcknowledgeAnomaly(ctx, e.ID, 20, 1))
	require.ErrorIs(t, AcknowledgeAnomaly(ctx, e.ID, 20, 1), ErrAnomalyConflict)
	require.NoError(t, RecordAnomaly(ctx, 42, 1, "model", "invalid_schema"))
	events, err = ListAnomalies(ctx)
	require.NoError(t, err)
	require.Equal(t, 20, events[0].AcknowledgedCount)
	require.Equal(t, 21, events[0].Count)
	r.FastForward(8 * 24 * time.Hour)
	events, err = ListAnomalies(ctx)
	require.NoError(t, err)
	require.Empty(t, events)
}
func TestAnomalyClassificationSeparatesUserAndUpstreamFailures(t *testing.T) {
	for _, tc := range []struct {
		status   int
		message  string
		upstream bool
		kind     string
	}{
		{400, "Invalid schema for response_format 'private-name'", true, "invalid_schema"},
		{429, "rate limit", true, "upstream"}, {429, "local limit", false, "rate_limit"},
		{502, "secret body", true, "upstream"}, {403, "用户额度不足", false, "quota"},
		{503, "No available channel for model", false, "unsupported_model"}, {401, "Invalid token", false, ""},
		{200, "context canceled", false, ""}, {400, "invalid", false, "invalid_request"},
	} {
		require.Equal(t, tc.kind, ClassifyAnomaly(tc.status, tc.message, tc.upstream))
	}
	anomalyRedis(t)
	ctx := context.Background()
	require.NoError(t, RecordAnomaly(ctx, 0, 1, "model", "rate_limit"))
	require.NoError(t, RecordAnomaly(ctx, 1, 1, "model", "quota"))
	events, err := ListAnomalies(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.False(t, events[0].Alert)
	raw, err := common.RDB.Get(ctx, anomalyPrefix+"event:"+events[0].ID).Result()
	require.NoError(t, err)
	require.NotContains(t, raw, "secret")
	require.NotContains(t, raw, "private-name")
}
func TestAnomalyUnavailableDoesNotReturnFakeEmptyData(t *testing.T) {
	r := anomalyRedis(t)
	r.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := ListAnomalies(ctx)
	require.Error(t, err)
}

func TestAnomalyRetentionCapEvictsOldestEvent(t *testing.T) {
	anomalyRedis(t)
	ctx := context.Background()
	now := time.Now().Unix()
	for i := 0; i < 2000; i++ {
		id := fmt.Sprintf("%032d", i)
		raw, _ := common.Marshal(AnomalyEvent{ID: id, UserID: 1, Kind: "quota", Count: 1, FirstSeen: now - 100, LastSeen: now - 100})
		require.NoError(t, common.RDB.Set(ctx, anomalyPrefix+"event:"+id, string(raw), anomalyRetention).Err())
		require.NoError(t, common.RDB.ZAdd(ctx, anomalyPrefix+"events", &redis.Z{Score: float64(now - 100), Member: id}).Err())
	}
	require.NoError(t, RecordAnomaly(ctx, 2, 1, "model", "invalid_schema"))
	total, err := common.RDB.ZCard(ctx, anomalyPrefix+"events").Result()
	require.NoError(t, err)
	require.EqualValues(t, 2000, total)
	exists, err := common.RDB.Exists(ctx, anomalyPrefix+"event:"+fmt.Sprintf("%032d", 0)).Result()
	require.NoError(t, err)
	require.Zero(t, exists)
}

func TestAnomalyConfirmedPauseIdempotencyExpiryAndRelease(t *testing.T) {
	r := anomalyRedis(t)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		require.NoError(t, RecordAnomaly(ctx, 42, 1, "model", "invalid_schema"))
	}
	events, err := ListAnomalies(ctx)
	require.NoError(t, err)
	e := events[0]
	until, err := ApplyAnomalyAction(ctx, e.ID, "confirm-one", "pause", 42, 900)
	require.NoError(t, err)
	require.Greater(t, until, time.Now().Unix())
	retry, err := ApplyAnomalyAction(ctx, e.ID, "confirm-one", "pause", 42, 900)
	require.NoError(t, err)
	require.Equal(t, until, retry)
	_, err = ApplyAnomalyAction(ctx, e.ID, "different", "pause", 42, 900)
	require.Error(t, err)
	got, err := AnomalySuspendedUntil(ctx, 42)
	require.NoError(t, err)
	require.Equal(t, until, got)
	_, err = ApplyAnomalyAction(ctx, e.ID, "release", "release", 42, 900)
	require.NoError(t, err)
	got, err = AnomalySuspendedUntil(ctx, 42)
	require.NoError(t, err)
	require.Zero(t, got)
	_, err = ApplyAnomalyAction(ctx, e.ID, "new-pause", "pause", 42, 900)
	require.Error(t, err)
	// Another user can be paused and expires without changing SQL state.
	for i := 0; i < 10; i++ {
		require.NoError(t, RecordAnomaly(ctx, 43, 1, "model", "invalid_schema"))
	}
	events, err = ListAnomalies(ctx)
	require.NoError(t, err)
	for _, item := range events {
		if item.UserID == 43 {
			_, err = ApplyAnomalyAction(ctx, item.ID, "expiry", "pause", 43, 900)
			require.NoError(t, err)
		}
	}
	r.FastForward(31 * time.Minute)
	got, err = AnomalySuspendedUntil(ctx, 43)
	require.NoError(t, err)
	require.Zero(t, got)
}
func TestAnomalyPauseRejectsUpstreamAndStaleEvents(t *testing.T) {
	anomalyRedis(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		require.NoError(t, RecordAnomaly(ctx, 42, 1, "model", "upstream"))
	}
	events, err := ListAnomalies(ctx)
	require.NoError(t, err)
	_, err = ApplyAnomalyAction(ctx, events[0].ID, "op", "pause", 42, 900)
	require.Error(t, err)
	_, err = ApplyAnomalyAction(ctx, events[0].ID, "op", "pause", 43, 900)
	require.Error(t, err)
}
