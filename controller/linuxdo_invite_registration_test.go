package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/oauth"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const linuxDOInviteRegistrationTestSecret = "linuxdo-invite-registration-test-secret-0123456789"

type linuxDOInviteTestProvider struct {
	oauthUser *oauth.OAuthUser
}

func (*linuxDOInviteTestProvider) GetName() string { return "Linux DO" }
func (*linuxDOInviteTestProvider) IsEnabled() bool { return true }
func (*linuxDOInviteTestProvider) ExchangeToken(context.Context, string, *gin.Context) (*oauth.OAuthToken, error) {
	return &oauth.OAuthToken{AccessToken: "test-access-token"}, nil
}
func (provider *linuxDOInviteTestProvider) GetUserInfo(context.Context, *oauth.OAuthToken) (*oauth.OAuthUser, error) {
	if provider.oauthUser == nil {
		return nil, errors.New("test OAuth user is not configured")
	}
	user := *provider.oauthUser
	return &user, nil
}
func (*linuxDOInviteTestProvider) IsUserIDTaken(providerUserID string) bool {
	return model.IsLinuxDOIdAlreadyTaken(providerUserID)
}
func (*linuxDOInviteTestProvider) FillUserByProviderID(user *model.User, providerUserID string) error {
	user.LinuxDOId = providerUserID
	return user.FillUserByLinuxDOId()
}
func (*linuxDOInviteTestProvider) SetProviderUserID(user *model.User, providerUserID string) {
	user.LinuxDOId = providerUserID
}
func (*linuxDOInviteTestProvider) GetProviderPrefix() string    { return "linuxdo_" }
func (*linuxDOInviteTestProvider) ProviderUserIDColumn() string { return "linux_do_id" }

type linuxDOInviteTestEnvironment struct {
	database *gorm.DB
	router   *gin.Engine
	provider *linuxDOInviteTestProvider
}

func setupLinuxDOInviteRegistrationTest(t *testing.T) linuxDOInviteTestEnvironment {
	t.Helper()
	require.NoError(t, i18n.Init())

	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousMainDatabaseType, previousLogDatabaseType := common.MainDatabaseType(), common.LogDatabaseType()
	previousRedisEnabled := common.RedisEnabled
	previousRegistrationInviteRequired := common.RegistrationInviteRequired
	previousRegisterEnabled := common.RegisterEnabled
	previousLinuxDOTrustLevel := common.LinuxDOMinimumTrustLevel
	previousQuotaForNewUser := common.QuotaForNewUser
	previousSessionSecret := common.SessionSecret
	previousGenerateDefaultToken := constant.GenerateDefaultToken
	previousGinMode := gin.Mode()

	gin.SetMode(gin.TestMode)
	common.RedisEnabled = false
	common.RegistrationInviteRequired = true
	common.RegisterEnabled = true
	common.LinuxDOMinimumTrustLevel = 0
	common.QuotaForNewUser = 0
	common.SessionSecret = "linuxdo-invite-registration-session-secret"
	constant.GenerateDefaultToken = true
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)

	database, err := gorm.Open(sqlite.Open(fmt.Sprintf(
		"file:%s?mode=memory&cache=shared&_pragma=busy_timeout(30000)&_txlock=immediate",
		strings.ReplaceAll(t.Name(), "/", "_"),
	)), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := database.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	require.NoError(t, database.AutoMigrate(
		&model.AuthFlow{},
		&model.Log{},
		&model.RegistrationInvite{},
		&model.Token{},
		&model.User{},
		&model.UserSession{},
	))
	model.DB, model.LOG_DB = database, database
	t.Setenv(service.RegistrationInviteHMACSecretEnv, linuxDOInviteRegistrationTestSecret)

	provider := &linuxDOInviteTestProvider{}
	previousProvider := oauth.GetProvider(linuxDOOAuthProviderName)
	oauth.Register(linuxDOOAuthProviderName, provider)
	t.Cleanup(func() {
		if previousProvider == nil {
			oauth.Unregister(linuxDOOAuthProviderName)
		} else {
			oauth.Register(linuxDOOAuthProviderName, previousProvider)
		}
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMainDatabaseType, previousLogDatabaseType)
		common.RedisEnabled = previousRedisEnabled
		common.RegistrationInviteRequired = previousRegistrationInviteRequired
		common.RegisterEnabled = previousRegisterEnabled
		common.LinuxDOMinimumTrustLevel = previousLinuxDOTrustLevel
		common.QuotaForNewUser = previousQuotaForNewUser
		common.SessionSecret = previousSessionSecret
		constant.GenerateDefaultToken = previousGenerateDefaultToken
		gin.SetMode(previousGinMode)
		_ = sqlDB.Close()
	})

	router := gin.New()
	router.GET("/api/status", GetStatus)
	router.POST("/api/oauth/state", GenerateOAuthCode)
	router.GET("/api/oauth/:provider", HandleOAuth)
	router.POST("/api/user/register", Register)
	return linuxDOInviteTestEnvironment{database: database, router: router, provider: provider}
}

func createLinuxDOInviteRegistrationTestInvite(t *testing.T) service.GeneratedRegistrationInvite {
	t.Helper()
	invites, err := service.GenerateRegistrationInvites(
		1,
		common.GetTimestamp()+3600,
		"OAuth registration test",
		1,
		linuxDOInviteRegistrationTestSecret,
	)
	require.NoError(t, err)
	require.Len(t, invites, 1)
	return invites[0]
}

func createLinuxDOOAuthState(t *testing.T, environment linuxDOInviteTestEnvironment, inviteCode string) string {
	t.Helper()
	body, err := common.Marshal(oauthStateRequest{
		Provider:   linuxDOOAuthProviderName,
		Intent:     model.AuthFlowIntentLogin,
		InviteCode: inviteCode,
	})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/api/oauth/state", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	environment.router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)

	var payload struct {
		Success bool `json:"success"`
		Data    struct {
			FlowToken string `json:"flow_token"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	require.True(t, payload.Success, response.Body.String())
	require.NotEmpty(t, payload.Data.FlowToken)
	return payload.Data.FlowToken
}

func executeLinuxDOOAuthCallback(t *testing.T, environment linuxDOInviteTestEnvironment, state string, oauthUser *oauth.OAuthUser) *httptest.ResponseRecorder {
	return executeOAuthCallback(t, environment, linuxDOOAuthProviderName, state, oauthUser)
}

func executeOAuthCallback(t *testing.T, environment linuxDOInviteTestEnvironment, providerName string, state string, oauthUser *oauth.OAuthUser) *httptest.ResponseRecorder {
	t.Helper()
	environment.provider.oauthUser = oauthUser
	request := httptest.NewRequest(http.MethodGet, "/api/oauth/"+providerName+"?state="+state+"&code=test-code", nil)
	response := httptest.NewRecorder()
	environment.router.ServeHTTP(response, request)
	return response
}

func decodeOAuthCallbackResponse(t *testing.T, response *httptest.ResponseRecorder) (bool, string) {
	t.Helper()
	var payload struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	return payload.Success, payload.Message
}

func linuxDOOAuthUser(providerID string, trustLevel int) *oauth.OAuthUser {
	return &oauth.OAuthUser{
		ProviderUserID: providerID,
		Username:       "linuxdo" + providerID,
		DisplayName:    "LinuxDO " + providerID,
		Extra:          map[string]any{"trust_level": trustLevel},
	}
}

func assertLinuxDOInviteRegistrationUserCount(t *testing.T, database *gorm.DB, want int64) {
	t.Helper()
	var count int64
	require.NoError(t, database.Unscoped().Model(&model.User{}).Count(&count).Error)
	assert.Equal(t, want, count)
}

func TestLinuxDOInviteOAuthStateStoresOnlyInviteIDAndCreatesUser(t *testing.T) {
	environment := setupLinuxDOInviteRegistrationTest(t)
	invite := createLinuxDOInviteRegistrationTestInvite(t)
	state := createLinuxDOOAuthState(t, environment, invite.Code)

	flow, err := model.GetAuthFlow(state, model.AuthFlowMatch{
		Purpose:  model.AuthFlowPurposeOAuth,
		Provider: linuxDOOAuthProviderName,
		Intent:   model.AuthFlowIntentLogin,
	})
	require.NoError(t, err)
	assert.NotContains(t, flow.Payload, invite.Code)
	var payload oauthFlowPayload
	require.NoError(t, common.UnmarshalJsonStr(flow.Payload, &payload))
	assert.Equal(t, invite.Id, payload.RegistrationInviteID)
	assert.Empty(t, payload.AffiliateCode)

	response := executeLinuxDOOAuthCallback(t, environment, state, linuxDOOAuthUser("101", 1))
	require.Equal(t, http.StatusOK, response.Code)
	success, message := decodeOAuthCallbackResponse(t, response)
	assert.True(t, success, message)
	_, err = model.GetAuthFlow(state, model.AuthFlowMatch{Purpose: model.AuthFlowPurposeOAuth})
	assert.ErrorIs(t, err, model.ErrAuthFlowConsumed)

	var user model.User
	require.NoError(t, environment.database.Where("linux_do_id = ?", "101").First(&user).Error)
	assert.Equal(t, common.RoleCommonUser, user.Role)
	assert.Equal(t, common.UserStatusEnabled, user.Status)
	assert.Equal(t, publicRegistrationDefaultGroup, user.Group)
	assert.Zero(t, user.Quota)
	var tokenCount int64
	require.NoError(t, environment.database.Model(&model.Token{}).Count(&tokenCount).Error)
	assert.Zero(t, tokenCount, "OAuth invitation registration must not create a default API token")

	var storedInvite model.RegistrationInvite
	require.NoError(t, environment.database.First(&storedInvite, invite.Id).Error)
	require.NotNil(t, storedInvite.UsedAt)
	require.NotNil(t, storedInvite.UsedBy)
	assert.Equal(t, user.Id, *storedInvite.UsedBy)
}

func TestLinuxDOInviteOAuthRejectsUnavailableInvitesWithoutCreatingUsers(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, environment linuxDOInviteTestEnvironment, invite service.GeneratedRegistrationInvite)
	}{
		{
			name:    "missing invitation",
			prepare: func(*testing.T, linuxDOInviteTestEnvironment, service.GeneratedRegistrationInvite) {},
		},
		{
			name: "expired after state issuance",
			prepare: func(t *testing.T, environment linuxDOInviteTestEnvironment, invite service.GeneratedRegistrationInvite) {
				require.NoError(t, environment.database.Model(&model.RegistrationInvite{}).Where("id = ?", invite.Id).Update("expires_at", common.GetTimestamp()-1).Error)
			},
		},
		{
			name: "revoked after state issuance",
			prepare: func(t *testing.T, environment linuxDOInviteTestEnvironment, invite service.GeneratedRegistrationInvite) {
				now := common.GetTimestamp()
				require.NoError(t, environment.database.Model(&model.RegistrationInvite{}).Where("id = ?", invite.Id).Updates(map[string]any{"revoked_by": 7, "revoked_at": now}).Error)
			},
		},
		{
			name: "used after state issuance",
			prepare: func(t *testing.T, environment linuxDOInviteTestEnvironment, invite service.GeneratedRegistrationInvite) {
				now := common.GetTimestamp()
				require.NoError(t, environment.database.Model(&model.RegistrationInvite{}).Where("id = ?", invite.Id).Updates(map[string]any{"used_by": 8, "used_at": now}).Error)
			},
		},
	}

	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			environment := setupLinuxDOInviteRegistrationTest(t)
			invite := createLinuxDOInviteRegistrationTestInvite(t)
			inviteCode := invite.Code
			if index == 0 {
				inviteCode = ""
			}
			state := createLinuxDOOAuthState(t, environment, inviteCode)
			test.prepare(t, environment, invite)

			response := executeLinuxDOOAuthCallback(t, environment, state, linuxDOOAuthUser(fmt.Sprintf("unavailable-%d", index), 1))
			require.Equal(t, http.StatusOK, response.Code)
			success, message := decodeOAuthCallbackResponse(t, response)
			assert.False(t, success)
			assert.Equal(t, i18n.Translate(i18n.DefaultLang, i18n.MsgOAuthRegistrationUnavailable), message)
			assert.NotContains(t, response.Body.String(), invite.Code)
			assertLinuxDOInviteRegistrationUserCount(t, environment.database, 0)

			_, err := model.GetAuthFlow(state, model.AuthFlowMatch{Purpose: model.AuthFlowPurposeOAuth})
			assert.ErrorIs(t, err, model.ErrAuthFlowConsumed)
		})
	}
}

func TestLinuxDOExistingUserSignsInWithoutInviteOrTL1(t *testing.T) {
	environment := setupLinuxDOInviteRegistrationTest(t)
	existing := &model.User{
		Username:    "linuxdo-existing",
		Password:    "password-placeholder",
		DisplayName: "Existing LinuxDO user",
		Role:        common.RoleCommonUser,
		Status:      common.UserStatusEnabled,
		Group:       publicRegistrationDefaultGroup,
		LinuxDOId:   "existing-linuxdo-id",
		AffCode:     "linuxdo-existing-aff",
		AuthVersion: 1,
	}
	require.NoError(t, environment.database.Create(existing).Error)
	invite := createLinuxDOInviteRegistrationTestInvite(t)
	state := createLinuxDOOAuthState(t, environment, "")

	response := executeLinuxDOOAuthCallback(t, environment, state, linuxDOOAuthUser(existing.LinuxDOId, 0))
	require.Equal(t, http.StatusOK, response.Code)
	success, message := decodeOAuthCallbackResponse(t, response)
	assert.True(t, success, message)
	assertLinuxDOInviteRegistrationUserCount(t, environment.database, 1)

	var storedInvite model.RegistrationInvite
	require.NoError(t, environment.database.First(&storedInvite, invite.Id).Error)
	assert.Nil(t, storedInvite.UsedAt)
	assert.Nil(t, storedInvite.UsedBy)
}

func TestLinuxDOInviteRegistrationRequiresTL1(t *testing.T) {
	environment := setupLinuxDOInviteRegistrationTest(t)
	invite := createLinuxDOInviteRegistrationTestInvite(t)
	state := createLinuxDOOAuthState(t, environment, invite.Code)

	response := executeLinuxDOOAuthCallback(t, environment, state, linuxDOOAuthUser("tl0-user", 0))
	require.Equal(t, http.StatusOK, response.Code)
	success, message := decodeOAuthCallbackResponse(t, response)
	assert.False(t, success)
	assert.Equal(t, i18n.Translate(i18n.DefaultLang, i18n.MsgOAuthTrustLevelLow), message)
	assertLinuxDOInviteRegistrationUserCount(t, environment.database, 0)

	var storedInvite model.RegistrationInvite
	require.NoError(t, environment.database.First(&storedInvite, invite.Id).Error)
	assert.Nil(t, storedInvite.UsedAt)
}

func TestLinuxDOInviteOAuthRejectsTamperedExpiredAndReplayedState(t *testing.T) {
	t.Run("tampered state", func(t *testing.T) {
		environment := setupLinuxDOInviteRegistrationTest(t)
		invite := createLinuxDOInviteRegistrationTestInvite(t)
		state := createLinuxDOOAuthState(t, environment, invite.Code)
		response := executeLinuxDOOAuthCallback(t, environment, state+"x", linuxDOOAuthUser("tampered-state", 1))
		assert.Equal(t, http.StatusForbidden, response.Code)
		assertLinuxDOInviteRegistrationUserCount(t, environment.database, 0)
	})

	t.Run("expired state", func(t *testing.T) {
		environment := setupLinuxDOInviteRegistrationTest(t)
		invite := createLinuxDOInviteRegistrationTestInvite(t)
		state := createLinuxDOOAuthState(t, environment, invite.Code)
		require.NoError(t, environment.database.Model(&model.AuthFlow{}).Where("expires_at > 0").Update("expires_at", time.Now().Add(-time.Second)).Error)
		response := executeLinuxDOOAuthCallback(t, environment, state, linuxDOOAuthUser("expired-state", 1))
		assert.Equal(t, http.StatusForbidden, response.Code)
		assertLinuxDOInviteRegistrationUserCount(t, environment.database, 0)
	})

	t.Run("tampered payload", func(t *testing.T) {
		environment := setupLinuxDOInviteRegistrationTest(t)
		invite := createLinuxDOInviteRegistrationTestInvite(t)
		state := createLinuxDOOAuthState(t, environment, invite.Code)
		require.NoError(t, environment.database.Model(&model.AuthFlow{}).Where("expires_at > 0").Update("payload", `{invalid`).Error)
		response := executeLinuxDOOAuthCallback(t, environment, state, linuxDOOAuthUser("tampered-payload", 1))
		require.Equal(t, http.StatusOK, response.Code)
		success, message := decodeOAuthCallbackResponse(t, response)
		assert.False(t, success)
		assert.Equal(t, i18n.Translate(i18n.DefaultLang, i18n.MsgOAuthStateInvalid), message)
		assertLinuxDOInviteRegistrationUserCount(t, environment.database, 0)
		_, err := model.GetAuthFlow(state, model.AuthFlowMatch{Purpose: model.AuthFlowPurposeOAuth})
		assert.ErrorIs(t, err, model.ErrAuthFlowConsumed)
	})

	t.Run("replayed state", func(t *testing.T) {
		environment := setupLinuxDOInviteRegistrationTest(t)
		invite := createLinuxDOInviteRegistrationTestInvite(t)
		state := createLinuxDOOAuthState(t, environment, invite.Code)
		first := executeLinuxDOOAuthCallback(t, environment, state, linuxDOOAuthUser("replayed-state", 1))
		firstSuccess, firstMessage := decodeOAuthCallbackResponse(t, first)
		require.True(t, firstSuccess, firstMessage)
		second := executeLinuxDOOAuthCallback(t, environment, state, linuxDOOAuthUser("replayed-state", 1))
		assert.Equal(t, http.StatusForbidden, second.Code)
		assertLinuxDOInviteRegistrationUserCount(t, environment.database, 1)
	})
}

func TestRegistrationInviteRequiredHidesOtherLoginMethodsAndBlocksNewUsers(t *testing.T) {
	environment := setupLinuxDOInviteRegistrationTest(t)

	statusRequest := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	statusResponse := httptest.NewRecorder()
	environment.router.ServeHTTP(statusResponse, statusRequest)
	require.Equal(t, http.StatusOK, statusResponse.Code)
	var statusPayload struct {
		Success bool           `json:"success"`
		Data    map[string]any `json:"data"`
	}
	require.NoError(t, common.Unmarshal(statusResponse.Body.Bytes(), &statusPayload))
	require.True(t, statusPayload.Success)
	assert.Equal(t, true, statusPayload.Data["registration_invite_required"])
	assert.Equal(t, float64(1), statusPayload.Data["linuxdo_minimum_trust_level"])
	assert.Equal(t, false, statusPayload.Data["password_login_enabled"])
	assert.Equal(t, false, statusPayload.Data["password_register_enabled"])

	passwordRequest := httptest.NewRequest(http.MethodPost, "/api/user/register", strings.NewReader(`{"username":"blocked-user","password":"password123"}`))
	passwordRequest.Header.Set("Content-Type", "application/json")
	passwordResponse := httptest.NewRecorder()
	environment.router.ServeHTTP(passwordResponse, passwordRequest)
	success, message := decodeOAuthCallbackResponse(t, passwordResponse)
	assert.False(t, success)
	assert.Equal(t, i18n.Translate(i18n.DefaultLang, i18n.MsgUserPasswordRegisterDisabled), message)
	assertLinuxDOInviteRegistrationUserCount(t, environment.database, 0)

	const otherProviderName = "task5-other-oauth"
	oauth.Register(otherProviderName, environment.provider)
	t.Cleanup(func() { oauth.Unregister(otherProviderName) })
	body, err := common.Marshal(oauthStateRequest{Provider: otherProviderName, Intent: model.AuthFlowIntentLogin})
	require.NoError(t, err)
	stateRequest := httptest.NewRequest(http.MethodPost, "/api/oauth/state", strings.NewReader(string(body)))
	stateRequest.Header.Set("Content-Type", "application/json")
	stateResponse := httptest.NewRecorder()
	environment.router.ServeHTTP(stateResponse, stateRequest)
	require.Equal(t, http.StatusOK, stateResponse.Code)
	var statePayload struct {
		Success bool `json:"success"`
		Data    struct {
			FlowToken string `json:"flow_token"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(stateResponse.Body.Bytes(), &statePayload))
	require.True(t, statePayload.Success)
	response := executeOAuthCallback(t, environment, otherProviderName, statePayload.Data.FlowToken, linuxDOOAuthUser("other-provider", 1))
	require.Equal(t, http.StatusOK, response.Code)
	success, message = decodeOAuthCallbackResponse(t, response)
	assert.False(t, success)
	assert.Equal(t, i18n.Translate(i18n.DefaultLang, i18n.MsgOAuthRegistrationUnavailable), message)
	assertLinuxDOInviteRegistrationUserCount(t, environment.database, 0)
}

func setupLinuxDOInviteRegistrationTransactionTest(t *testing.T, database *gorm.DB, databaseType common.DatabaseType) {
	t.Helper()
	require.False(t, database.Migrator().HasTable(&model.User{}), "transaction tests require an empty database")

	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousMainDatabaseType, previousLogDatabaseType := common.MainDatabaseType(), common.LogDatabaseType()
	previousRedisEnabled := common.RedisEnabled
	previousRegistrationInviteRequired := common.RegistrationInviteRequired
	previousRegisterEnabled := common.RegisterEnabled
	previousLinuxDOTrustLevel := common.LinuxDOMinimumTrustLevel
	previousQuotaForNewUser := common.QuotaForNewUser
	previousSessionSecret := common.SessionSecret
	common.RedisEnabled = false
	common.RegistrationInviteRequired = true
	common.RegisterEnabled = true
	common.LinuxDOMinimumTrustLevel = 0
	common.QuotaForNewUser = 0
	common.SessionSecret = "linuxdo-invite-registration-transaction-session-secret"
	common.SetDatabaseTypes(databaseType, databaseType)
	model.DB, model.LOG_DB = database, database
	require.NoError(t, database.AutoMigrate(&model.AuthFlow{}, &model.User{}, &model.RegistrationInvite{}))
	t.Cleanup(func() {
		require.NoError(t, database.Migrator().DropTable(&model.AuthFlow{}, &model.RegistrationInvite{}, &model.User{}))
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.SetDatabaseTypes(previousMainDatabaseType, previousLogDatabaseType)
		common.RedisEnabled = previousRedisEnabled
		common.RegistrationInviteRequired = previousRegistrationInviteRequired
		common.RegisterEnabled = previousRegisterEnabled
		common.LinuxDOMinimumTrustLevel = previousLinuxDOTrustLevel
		common.QuotaForNewUser = previousQuotaForNewUser
		common.SessionSecret = previousSessionSecret
	})
}

func testLinuxDOInviteRegistrationTransaction(t *testing.T, database *gorm.DB, databaseType common.DatabaseType) {
	t.Helper()
	setupLinuxDOInviteRegistrationTransactionTest(t, database, databaseType)
	invites, err := service.GenerateRegistrationInvites(1, common.GetTimestamp()+3600, "concurrent OAuth registration", 1, linuxDOInviteRegistrationTestSecret)
	require.NoError(t, err)
	require.Len(t, invites, 1)

	provider := &linuxDOInviteTestProvider{}
	start := make(chan struct{})
	results := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for _, oauthUser := range []*oauth.OAuthUser{
		linuxDOOAuthUser("txn-one", 1),
		linuxDOOAuthUser("txn-two", 1),
	} {
		waitGroup.Add(1)
		go func(user *oauth.OAuthUser) {
			defer waitGroup.Done()
			<-start
			_, createErr := createRegistrationInviteOAuthUser(linuxDOOAuthProviderName, provider, user, invites[0].Id)
			results <- createErr
		}(oauthUser)
	}
	close(start)
	waitGroup.Wait()
	close(results)

	successes := 0
	for result := range results {
		if result == nil {
			successes++
			continue
		}
		var inviteError *OAuthInviteRegistrationError
		assert.ErrorAs(t, result, &inviteError)
	}
	assert.Equal(t, 1, successes)
	assertLinuxDOInviteRegistrationUserCount(t, database, 1)

	var storedInvite model.RegistrationInvite
	require.NoError(t, database.First(&storedInvite, invites[0].Id).Error)
	require.NotNil(t, storedInvite.UsedAt)
	require.NotNil(t, storedInvite.UsedBy)
}

func TestLinuxDOInviteRegistrationTransactionSQLite(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(fmt.Sprintf(
		"file:%s?mode=memory&cache=shared&_pragma=busy_timeout(30000)&_txlock=immediate",
		strings.ReplaceAll(t.Name(), "/", "_"),
	)), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := database.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testLinuxDOInviteRegistrationTransaction(t, database, common.DatabaseTypeSQLite)
}

func testLinuxDOInviteRegistrationProviderIdentityConflict(t *testing.T, database *gorm.DB, databaseType common.DatabaseType) {
	t.Helper()
	setupLinuxDOInviteRegistrationTransactionTest(t, database, databaseType)
	invites, err := service.GenerateRegistrationInvites(2, common.GetTimestamp()+3600, "provider identity conflict", 1, linuxDOInviteRegistrationTestSecret)
	require.NoError(t, err)
	require.Len(t, invites, 2)

	provider := &linuxDOInviteTestProvider{}
	start := make(chan struct{})
	results := make(chan error, len(invites))
	var waitGroup sync.WaitGroup
	for _, invite := range invites {
		waitGroup.Add(1)
		go func(inviteID int) {
			defer waitGroup.Done()
			<-start
			_, createErr := createRegistrationInviteOAuthUser(
				linuxDOOAuthProviderName,
				provider,
				linuxDOOAuthUser("same-linuxdo-identity", 1),
				inviteID,
			)
			results <- createErr
		}(invite.Id)
	}
	close(start)
	waitGroup.Wait()
	close(results)

	successes := 0
	for result := range results {
		if result == nil {
			successes++
			continue
		}
		var inviteError *OAuthInviteRegistrationError
		assert.ErrorAs(t, result, &inviteError)
	}
	assert.Equal(t, 1, successes)
	assertLinuxDOInviteRegistrationUserCount(t, database, 1)

	var usedCount int64
	require.NoError(t, database.Model(&model.RegistrationInvite{}).Where("used_at IS NOT NULL").Count(&usedCount).Error)
	assert.Equal(t, int64(1), usedCount, "the failed identity claim must not consume its invitation")
}

func TestLinuxDOInviteRegistrationProviderIdentityConflictSQLite(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(fmt.Sprintf(
		"file:%s?mode=memory&cache=shared&_pragma=busy_timeout(30000)&_txlock=immediate",
		strings.ReplaceAll(t.Name(), "/", "_"),
	)), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := database.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testLinuxDOInviteRegistrationProviderIdentityConflict(t, database, common.DatabaseTypeSQLite)
}

func TestLinuxDOInviteRegistrationTransactionMySQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not configured")
	}
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := database.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testLinuxDOInviteRegistrationTransaction(t, database, common.DatabaseTypeMySQL)
}

func TestLinuxDOInviteRegistrationProviderIdentityConflictMySQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not configured")
	}
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := database.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testLinuxDOInviteRegistrationProviderIdentityConflict(t, database, common.DatabaseTypeMySQL)
}

func TestLinuxDOInviteRegistrationTransactionPostgreSQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not configured")
	}
	database, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := database.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testLinuxDOInviteRegistrationTransaction(t, database, common.DatabaseTypePostgreSQL)
}

func TestLinuxDOInviteRegistrationProviderIdentityConflictPostgreSQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not configured")
	}
	database, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := database.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testLinuxDOInviteRegistrationProviderIdentityConflict(t, database, common.DatabaseTypePostgreSQL)
}
