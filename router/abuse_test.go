package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	appI18n "github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAbuseRootRoutesAndOwnStatus(t *testing.T) {
	require.NoError(t, appI18n.Init())
	env := setupRegistrationInviteRouterTest(t)
	require.NoError(t, model.MigrateAbuse(env.database))
	old := common.SessionCookieSecure
	common.SessionCookieSecure = false
	t.Cleanup(func() { common.SessionCookieSecure = old })
	registerAbuseRoutes(env.router.Group("/api"))
	for _, path := range []string{"/api/abuse/settings", "/api/abuse/events", "/api/abuse/users"} {
		for _, token := range []string{env.commonToken, env.adminToken} {
			resp := performRegistrationInviteRequest(t, env.router, "GET", path, token, "")
			assert.Equal(t, 403, resp.Code)
		}
		resp := performRegistrationInviteRequest(t, env.router, "GET", path, env.rootToken, "")
		assert.Equal(t, 200, resp.Code, resp.Body.String())
	}
	resp := performRegistrationInviteRequest(t, env.router, "GET", "/api/user/self/abuse", env.commonToken, "")
	assert.Equal(t, 200, resp.Code)
	assert.NotContains(t, resp.Body.String(), "rule_id")
	resp = performRegistrationInviteRequest(t, env.router, "PUT", "/api/abuse/settings", env.rootToken, `{"mode":"enforce","limit_10m":3,"limit_24h":8,"freeze_minutes":60,"enabled_rules":["fake-verified"]}`)
	assert.Equal(t, 400, resp.Code)
	resp = performRegistrationInviteRequest(t, env.router, "GET", "/api/abuse/events?p=-1", env.rootToken, "")
	assert.Equal(t, 400, resp.Code)
	resp = performRegistrationInviteRequest(t, env.router, "POST", "/api/abuse/users/1/unfreeze", env.rootToken, `{"reason":" "}`)
	assert.Equal(t, 400, resp.Code)
	require.Eventually(t, func() bool {
		var n int64
		return env.database.Model(&model.Log{}).Where("type = ?", model.LogTypeManage).Count(&n).Error == nil && n >= 2
	}, 3*time.Second, 5*time.Millisecond)
}

func TestAbuseSuspensionBlocksEveryTokenAndPlayground(t *testing.T) {
	require.NoError(t, appI18n.Init())
	setupRelayRouterTestDB(t)
	env := setupRegistrationInviteRouterTest(t)
	require.NoError(t, model.MigrateAbuse(env.database))
	require.NoError(t, env.database.AutoMigrate(&model.Token{}))
	var user model.User
	require.NoError(t, env.database.Where("username = ?", "invite-common").First(&user).Error)
	for _, key := range []string{"abuse-token-one", "abuse-token-two"} {
		require.NoError(t, env.database.Create(&model.Token{UserId: user.Id, Key: strings.ReplaceAll(key, "-", ""), Status: 1, ExpiredTime: -1, UnlimitedQuota: true}).Error)
	}
	require.NoError(t, env.database.Create(&model.AbuseState{UserID: user.Id, Round: 2, BlockedUntil: time.Now().Unix() + 3600}).Error)
	r := gin.New()
	SetRelayRouter(r)
	SetTaskPluginProtocolRouter(r)
	for _, key := range []string{"abusetokenone", "abusetokentwo"} {
		req := httptest.NewRequest("POST", "/v1/responses", strings.NewReader(`{"model":"test"}`))
		req.Header.Set("Authorization", "Bearer sk-"+key)
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		r.ServeHTTP(recorder, req)
		assert.Equal(t, 403, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "account_temporarily_suspended")
	}
	// Dashboard auth uses the same distributor boundary; no token is required.
	r = gin.New()
	r.POST("/pg/chat/completions", middleware.UserAuth(), middleware.Distribute(), controller.Playground)
	resp := performRegistrationInviteRequest(t, r, "POST", "/pg/chat/completions", env.commonToken, `{"model":"test"}`)
	assert.Equal(t, http.StatusForbidden, resp.Code)
	assert.Contains(t, resp.Body.String(), "account_temporarily_suspended")
	// With state unavailable, do not silently forward.
	require.NoError(t, env.database.Migrator().DropTable(&model.AbuseState{}))
	resp = performRegistrationInviteRequest(t, r, "POST", "/pg/chat/completions", env.commonToken, `{"model":"test"}`)
	assert.Equal(t, 503, resp.Code)
}

func TestAbuseReviewRequiresRootAndNeverChangesAccountAccess(t *testing.T) {
	require.NoError(t, appI18n.Init())
	env := setupRegistrationInviteRouterTest(t)
	require.NoError(t, model.MigrateAbuse(env.database))
	registerAbuseRoutes(env.router.Group("/api"))
	event := model.AbuseEvent{UserID: 1, RequestID: "review-request", CreatedAt: time.Now().Unix(), Action: "recorded"}
	require.NoError(t, env.database.Create(&event).Error)
	path := fmt.Sprintf("/api/abuse/events/%d", event.ID)
	for _, token := range []string{env.commonToken, env.adminToken} {
		for _, endpoint := range []string{"/excerpt", "/review"} {
			res := performRegistrationInviteRequest(t, env.router, "POST", path+endpoint, token, `{"decision":"confirmed","note":"test","version":0}`)
			assert.Equal(t, 403, res.Code)
		}
		res := performRegistrationInviteRequest(t, env.router, "GET", path+"/history", token, "")
		assert.Equal(t, 403, res.Code)
	}
	res := performRegistrationInviteRequest(t, env.router, "POST", path+"/excerpt", env.rootToken, "")
	assert.Equal(t, 200, res.Code)
	assert.Contains(t, res.Body.String(), "not_collected")
	assert.Contains(t, res.Header().Get("Cache-Control"), "no-store")
	res = performRegistrationInviteRequest(t, env.router, "POST", path+"/review", env.rootToken, `{"decision":"confirmed","note":"reviewed category evidence","version":0}`)
	assert.Equal(t, 200, res.Code, res.Body.String())
	res = performRegistrationInviteRequest(t, env.router, "POST", path+"/review", env.rootToken, `{"decision":"confirmed","note":"same submission","version":0}`)
	assert.Equal(t, 409, res.Code)
	res = performRegistrationInviteRequest(t, env.router, "GET", path+"/history", env.rootToken, "")
	assert.Equal(t, 200, res.Code)
	assert.Contains(t, res.Body.String(), "excerpt_access")
	assert.Contains(t, res.Body.String(), "reviewed category evidence")
	require.NoError(t, env.database.First(&event, event.ID).Error)
	assert.Equal(t, "confirmed", event.ReviewStatus)
	assert.Equal(t, "recorded", event.Action)
	assert.False(t, event.Actionable)
	assert.False(t, event.Notify)
	res = performRegistrationInviteRequest(t, env.router, "POST", path+"/review", env.rootToken, `{"decision":"confirmed","note":" ","version":1}`)
	assert.Equal(t, 400, res.Code)
	require.NoError(t, env.database.Migrator().DropTable(&model.AbuseReviewAudit{}))
	res = performRegistrationInviteRequest(t, env.router, "POST", path+"/excerpt", env.rootToken, "")
	assert.Equal(t, 503, res.Code)
}
