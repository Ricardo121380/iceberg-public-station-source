package router

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	appI18n "github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestAnomalyRoutesEnforceRolesAndAuditAcknowledgements(t *testing.T) {
	require.NoError(t, appI18n.Init())
	env := setupRegistrationInviteRouterTest(t)
	old, enabled := common.RDB, common.RedisEnabled
	r := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: r.Addr()})
	common.RDB = client
	common.RedisEnabled = true
	t.Cleanup(func() { client.Close(); common.RDB = old; common.RedisEnabled = enabled })
	registerAnomalyRoutes(env.router.Group("/api"))
	for _, path := range []string{"/events", "/settings", "/audit"} {
		res := performRegistrationInviteRequest(t, env.router, "GET", "/api/anomalies"+path, env.commonToken, "")
		require.Equal(t, 403, res.Code)
		res = performRegistrationInviteRequest(t, env.router, "GET", "/api/anomalies"+path, env.adminToken, "")
		require.Equal(t, 200, res.Code)
		require.Contains(t, res.Header().Get("Cache-Control"), "no-store")
	}
	body := `{"enabled":true,"window_minutes":5,"schema_threshold":10,"rate_threshold":100,"upstream_threshold":5,"version":0}`
	res := performRegistrationInviteRequest(t, env.router, "PUT", "/api/anomalies/settings", env.adminToken, body)
	require.Equal(t, 403, res.Code)
	res = performRegistrationInviteRequest(t, env.router, "PUT", "/api/anomalies/settings", env.rootToken, body)
	require.Equal(t, 200, res.Code, res.Body.String())
	res = performRegistrationInviteRequest(t, env.router, "PUT", "/api/anomalies/settings", env.rootToken, body)
	require.Equal(t, 409, res.Code)
	ctx := context.Background()
	require.NoError(t, service.RecordAnomaly(ctx, 305, 1, "test", "invalid_schema"))
	events, err := service.ListAnomalies(ctx)
	require.NoError(t, err)
	path := fmt.Sprintf("/api/anomalies/events/%s/acknowledge", events[0].ID)
	res = performRegistrationInviteRequest(t, env.router, "POST", path, env.commonToken, `{"count":1}`)
	require.Equal(t, 403, res.Code)
	res = performRegistrationInviteRequest(t, env.router, "POST", path, env.adminToken, `{"count":1}`)
	require.Equal(t, 200, res.Code)
	res = performRegistrationInviteRequest(t, env.router, "GET", "/api/anomalies/events?p=bad&user_id=305", env.rootToken, "")
	require.Equal(t, 400, res.Code)
	res = performRegistrationInviteRequest(t, env.router, "GET", "/api/anomalies/events?user_id=305&state=acknowledged", env.rootToken, "")
	require.Equal(t, 200, res.Code)
	require.Contains(t, res.Body.String(), `"total":1`)
	res = performRegistrationInviteRequest(t, env.router, "GET", "/api/anomalies/audit", env.rootToken, "")
	require.Contains(t, res.Body.String(), "acknowledged")
}

func TestAnomalyBotSecretAndProtectedUsers(t *testing.T) {
	require.NoError(t, appI18n.Init())
	env := setupRegistrationInviteRouterTest(t)
	t.Setenv("ANOMALY_BOT_SECRET", "01234567890123456789012345678901")
	t.Setenv("ANOMALY_BOT_OPERATOR_ID", "900")
	registerAnomalyRoutes(env.router.Group("/api"))
	req := httptest.NewRequest("GET", "/api/internal/anomaly-bot/settings", nil)
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	require.Equal(t, 403, rec.Code)
	req = httptest.NewRequest("POST", "/api/internal/anomaly-bot/action", strings.NewReader(`{"event_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","operation_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","action":"pause","user_id":1,"operator_id":900}`))
	req.Header.Set("X-Anomaly-Bot-Secret", os.Getenv("ANOMALY_BOT_SECRET"))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	require.Equal(t, 409, rec.Code)
}

func TestAnomalyBotPauseBlocksAllTokensButAllowsDashboard(t *testing.T) {
	require.NoError(t, appI18n.Init())
	setupRelayRouterTestDB(t)
	env := setupRegistrationInviteRouterTest(t)
	require.NoError(t, model.MigrateAbuse(env.database))
	require.NoError(t, env.database.AutoMigrate(&model.Token{}))
	old, enabled := common.RDB, common.RedisEnabled
	r := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: r.Addr()})
	common.RDB = client
	common.RedisEnabled = true
	t.Cleanup(func() { client.Close(); common.RDB = old; common.RedisEnabled = enabled })
	var user model.User
	require.NoError(t, env.database.Where("username = ?", "invite-common").First(&user).Error)
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		require.NoError(t, service.RecordAnomaly(ctx, user.Id, 1, "test", "invalid_schema"))
	}
	events, err := service.ListAnomalies(ctx)
	require.NoError(t, err)
	_, err = service.ApplyAnomalyAction(ctx, events[0].ID, "pause-test", "pause", user.Id, 900)
	require.NoError(t, err)
	server := gin.New()
	SetRelayRouter(server)
	SetTaskPluginProtocolRouter(server)
	for _, key := range []string{"anomalytesttokenone", "anomalytesttokentwo"} {
		require.NoError(t, env.database.Create(&model.Token{UserId: user.Id, Key: key, Status: 1, ExpiredTime: -1, UnlimitedQuota: true}).Error)
		req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(`{"model":"test"}`))
		req.Header.Set("Authorization", "Bearer sk-"+key)
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)
		require.Equal(t, 403, res.Code)
		require.Contains(t, res.Body.String(), "account_temporarily_suspended")
	}
	registerAbuseRoutes(env.router.Group("/api"))
	res := performRegistrationInviteRequest(t, env.router, "GET", "/api/user/self/abuse", env.commonToken, "")
	require.Equal(t, 200, res.Code)
	require.Contains(t, res.Body.String(), `"suspended":true`)
	_, err = service.ApplyAnomalyAction(ctx, events[0].ID, "release-test", "release", user.Id, 900)
	require.NoError(t, err)
	res = performRegistrationInviteRequest(t, env.router, "GET", "/api/user/self/abuse", env.commonToken, "")
	require.Equal(t, 200, res.Code)
	require.Contains(t, res.Body.String(), `"suspended":false`)
}
