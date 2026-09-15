package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	appI18n "github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayRoutesRejectNonStreamingBeforeChannelSelection(t *testing.T) {
	require.NoError(t, appI18n.Init())
	setupRelayRouterTestDB(t)
	user := model.User{Username: "stream-only-user", Status: common.UserStatusEnabled, Group: "default", Quota: 100}
	require.NoError(t, model.DB.Create(&user).Error)
	token := model.Token{UserId: user.Id, Key: "streamonlytestkey", Status: common.TokenStatusEnabled, ExpiredTime: -1, RemainQuota: 100}
	require.NoError(t, model.DB.Create(&token).Error)
	engine := gin.New()
	SetRelayRouter(engine)
	SetTaskPluginProtocolRouter(engine)
	// No channels or upstreams exist in this fixture. A stream policy error
	// proves dispatch was not attempted, including the Responses plugin path.
	for _, path := range []string{
		"/v1/chat/completions", "/v1/completions", "/v1/responses", "/v1/messages",
		"/v1/models/test:generateContent", "/v1beta/models/test:generateContent",
	} {
		t.Run(path, func(t *testing.T) {
			for _, key := range []string{"", token.Key} {
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"test","stream":false,"input":"hello"}`))
				req.Header.Set("Content-Type", "application/json")
				if key != "" {
					req.Header.Set("Authorization", "Bearer "+key)
				}
				recorder := httptest.NewRecorder()
				engine.ServeHTTP(recorder, req)
				if key == "" {
					assert.Equal(t, http.StatusUnauthorized, recorder.Code)
					continue
				}
				assert.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
				assert.Contains(t, recorder.Body.String(), "only supports streaming generation")
			}
		})
	}
	var afterUser model.User
	var afterToken model.Token
	require.NoError(t, model.DB.First(&afterUser, user.Id).Error)
	require.NoError(t, model.DB.First(&afterToken, token.Id).Error)
	assert.Equal(t, user.Quota, afterUser.Quota)
	assert.Zero(t, afterUser.UsedQuota)
	assert.Equal(t, token.RemainQuota, afterToken.RemainQuota)
	assert.Zero(t, afterToken.UsedQuota)
}

func TestPlaygroundRouteRejectsNonStreaming(t *testing.T) {
	require.NoError(t, appI18n.Init())
	env := setupRegistrationInviteRouterTest(t)
	engine := gin.New()
	SetRelayRouter(engine)
	resp := performRegistrationInviteRequest(t, engine, "POST", "/pg/chat/completions", env.commonToken, `{"model":"test","stream":false}`)
	assert.Equal(t, http.StatusBadRequest, resp.Code, resp.Body.String())
	assert.Contains(t, resp.Body.String(), "stream_required")
}
