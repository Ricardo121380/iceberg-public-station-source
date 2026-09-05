package service

import (
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	appI18n "github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func safetyTestContext(t *testing.T) (*gin.Context, *gorm.DB) {
	require.NoError(t, appI18n.Init())
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, model.MigrateAbuse(db))
	require.NoError(t, db.AutoMigrate(&model.User{}))
	require.NoError(t, db.Create(&model.User{Id: 1, Username: "safety", Status: 1, Role: 1}).Error)
	old := model.DB
	model.DB = db
	t.Cleanup(func() { model.DB = old; s, _ := db.DB(); _ = s.Close() })
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	c.Set("id", 1)
	c.Set("token_id", 2)
	c.Set("channel_type", constant.ChannelTypeNewAPI)
	c.Set("channel_id", 1)
	c.Set(common.RequestIdKey, "safety-test-request")
	return c, db
}

func TestAbuseStructuralSignalsAndRedaction(t *testing.T) {
	tests := []struct {
		name, payload, protocol string
		found                   bool
	}{
		{"chat-output", `{"choices":[{"finish_reason":"content_filter"}]}`, "chat", true},
		{"responses-incomplete", `{"type":"response.incomplete","response":{"incomplete_details":{"reason":"content_filter"}}}`, "responses", true},
		{"responses-refusal", `{"type":"response.refusal.delta","delta":"sk-private user text"}`, "responses", true},
		{"claude-refusal", `{"delta":{"stop_reason":"refusal"}}`, "claude", true},
		{"error-code", `{"error":{"code":"content_policy_violation","message":"prompt sk-private"}}`, "error", true},
		{"text-only", `{"error":{"code":"invalid_request","message":"NSFW safety sk-private"}}`, "error", false},
		{"normal-text", `{"choices":[{"message":{"content":"I cannot help with that"}}]}`, "chat", false},
		{"fee-wrapper", `{"error":{"code":"violation_fee.grok.csam"}}`, "error", false},
		{"truncated", `{"error":`, "error", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, db := safetyTestContext(t)
			finish, err := BeginSafetyObservation(c)
			require.NoError(t, err)
			NextSafetyAttempt(c)
			ObserveSafetyPayload(c, []byte(tt.payload), tt.protocol)
			ObserveSafetyPayload(c, []byte(tt.payload), tt.protocol)
			assert.False(t, SafetyInputBlocked(c))
			finish()
			var events []model.AbuseEvent
			require.NoError(t, db.Find(&events).Error)
			if !tt.found {
				assert.Empty(t, events)
				return
			}
			require.Len(t, events, 1)
			assert.False(t, events[0].Actionable)
			assert.NotContains(t, events[0].Summary, "sk-private")
			assert.NotContains(t, events[0].Attempts, "sk-private")
			var attempts []safetyAttempt
			require.NoError(t, common.UnmarshalJsonStr(events[0].Attempts, &attempts))
			assert.Len(t, attempts, 1)
		})
	}
}

func TestAbuseSuspensionAndDatabaseFailure(t *testing.T) {
	c, db := safetyTestContext(t)
	require.NoError(t, db.Create(&model.AbuseState{UserID: 1, Round: 2, BlockedUntil: time.Now().Unix() + 3600}).Error)
	finish, err := BeginSafetyObservation(c)
	require.NoError(t, err)
	require.NotNil(t, finish)
	assert.True(t, c.IsAborted())
	assert.Equal(t, 403, c.Writer.Status())
	require.NoError(t, db.Migrator().DropTable(&model.AbuseState{}))
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	c.Set("id", 1)
	_, err = BeginSafetyObservation(c)
	require.Error(t, err)
}

func TestAbuseVerifiedFixtureFreezesAfterThreeRequests(t *testing.T) {
	c, db := safetyTestContext(t)
	previousRules := verifiedSafetyRules
	verifiedSafetyRules = []SafetyRule{{ID: "fixture-only", Version: "test", ChannelID: 1, Protocol: "error", Field: "error.code", Value: "fixture_explicit_input", Category: "sexual_explicit", Verified: true, Evidence: "synthetic test only; never production"}}
	t.Cleanup(func() { verifiedSafetyRules = previousRules })
	previousRedis := common.RDB
	server := miniredis.RunT(t)
	common.RDB = redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = common.RDB.Close(); common.RDB = previousRedis })
	p, err := model.GetAbusePolicy()
	require.NoError(t, err)
	p.Mode = "enforce"
	p.EnabledRules = `["fixture-only"]`
	require.NoError(t, model.SaveAbusePolicy(p, 1))
	for i := 0; i < 3; i++ {
		c, _ = gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
		c.Set("id", 1)
		c.Set("token_id", i+1)
		c.Set("channel_type", constant.ChannelTypeNewAPI)
		c.Set("channel_id", 1)
		c.Set(common.RequestIdKey, fmt.Sprintf("fixture-%d", i))
		finish, err := BeginSafetyObservation(c)
		require.NoError(t, err)
		require.False(t, c.IsAborted())
		NextSafetyAttempt(c)
		ObserveSafetyPayload(c, []byte(`{"error":{"code":"fixture_explicit_input","message":"not persisted"}}`), "error")
		assert.True(t, SafetyInputBlocked(c))
		finish()
	}
	state, err := model.GetAbuseState(1)
	require.NoError(t, err)
	assert.Greater(t, state.BlockedUntil, time.Now().Unix())
	var freezes int64
	require.NoError(t, db.Model(&model.AbuseEvent{}).Where("action = ?", "frozen").Count(&freezes).Error)
	assert.EqualValues(t, 1, freezes)
	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	c.Set("id", 1)
	_, err = BeginSafetyObservation(c)
	require.NoError(t, err)
	assert.True(t, c.IsAborted())
}
