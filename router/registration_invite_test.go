package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const registrationInviteAPITestSecret = "registration-invite-api-test-secret-0123456789"

type registrationInviteRouterTestEnvironment struct {
	database    *gorm.DB
	router      *gin.Engine
	root        *model.User
	rootToken   string
	commonToken string
	adminToken  string
}

func setupRegistrationInviteRouterTest(t *testing.T) registrationInviteRouterTestEnvironment {
	t.Helper()
	previousDB, previousLogDB := model.DB, model.LOG_DB
	previousRedisEnabled := common.RedisEnabled
	previousMainDatabaseType, previousLogDatabaseType := common.MainDatabaseType(), common.LogDatabaseType()
	previousGinMode := gin.Mode()

	gin.SetMode(gin.TestMode)
	common.RedisEnabled = false
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	database, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(&model.User{}, &model.Log{}, &model.RegistrationInvite{}))
	model.DB, model.LOG_DB = database, database
	t.Setenv(service.RegistrationInviteHMACSecretEnv, registrationInviteAPITestSecret)
	t.Cleanup(func() {
		model.DB, model.LOG_DB = previousDB, previousLogDB
		common.RedisEnabled = previousRedisEnabled
		common.SetDatabaseTypes(previousMainDatabaseType, previousLogDatabaseType)
		gin.SetMode(previousGinMode)
		if sqlDB, sqlErr := database.DB(); sqlErr == nil {
			_ = sqlDB.Close()
		}
	})

	root, rootToken := createRegistrationInviteTestUser(t, "invite-root", common.RoleRootUser)
	_, commonToken := createRegistrationInviteTestUser(t, "invite-common", common.RoleCommonUser)
	_, adminToken := createRegistrationInviteTestUser(t, "invite-admin", common.RoleAdminUser)

	router := gin.New()
	apiRouter := router.Group("/api")
	registerRegistrationInviteRoutes(apiRouter)
	return registrationInviteRouterTestEnvironment{
		database:    database,
		router:      router,
		root:        root,
		rootToken:   rootToken,
		commonToken: commonToken,
		adminToken:  adminToken,
	}
}

func createRegistrationInviteTestUser(t *testing.T, username string, role int) (*model.User, string) {
	t.Helper()
	token := username + "-dashboard-token"
	user := &model.User{
		Username:    username,
		Password:    "password-placeholder",
		Role:        role,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AccessToken: &token,
		AuthVersion: 1,
		AffCode:     username + "-aff",
	}
	require.NoError(t, model.DB.Create(user).Error)
	return user, token
}

func performRegistrationInviteRequest(t *testing.T, router *gin.Engine, method string, path string, token string, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestRegistrationInviteRoutesRequireRoot(t *testing.T) {
	environment := setupRegistrationInviteRouterTest(t)
	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "unauthenticated", wantStatus: http.StatusUnauthorized},
		{name: "common user", token: environment.commonToken, wantStatus: http.StatusForbidden},
		{name: "admin user", token: environment.adminToken, wantStatus: http.StatusForbidden},
		{name: "root user", token: environment.rootToken, wantStatus: http.StatusOK},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performRegistrationInviteRequest(t, environment.router, http.MethodGet, "/api/registration-invites", test.token, "")
			assert.Equal(t, test.wantStatus, response.Code)
		})
	}
}

func TestRegistrationInviteCreateEnforcesBoundsAndKeepsRawCodesOneTime(t *testing.T) {
	environment := setupRegistrationInviteRouterTest(t)

	first := performRegistrationInviteRequest(t, environment.router, http.MethodPost, "/api/registration-invites", environment.rootToken, `{"count":1,"valid_days":7,"note":"first batch"}`)
	require.Equal(t, http.StatusOK, first.Code)
	assert.Contains(t, first.Header().Get("Cache-Control"), "no-store")
	assert.NotContains(t, first.Body.String(), "code_hash")
	var firstPayload struct {
		Success bool `json:"success"`
		Data    struct {
			Invites []struct {
				Id   int    `json:"id"`
				Code string `json:"code"`
			} `json:"invites"`
		} `json:"data"`
	}
	require.NoError(t, common.Unmarshal(first.Body.Bytes(), &firstPayload))
	require.True(t, firstPayload.Success)
	require.Len(t, firstPayload.Data.Invites, 1)
	rawCode := firstPayload.Data.Invites[0].Code
	assert.NoError(t, model.ValidateRegistrationInviteCode(rawCode))

	maximum := performRegistrationInviteRequest(t, environment.router, http.MethodPost, "/api/registration-invites", environment.rootToken, `{"count":100,"valid_days":1,"note":"maximum batch"}`)
	require.Equal(t, http.StatusOK, maximum.Code)
	assert.Contains(t, maximum.Header().Get("Cache-Control"), "no-store")
	assert.Contains(t, maximum.Body.String(), `"success":true`)

	overLimit := performRegistrationInviteRequest(t, environment.router, http.MethodPost, "/api/registration-invites", environment.rootToken, `{"count":101,"valid_days":1,"note":"over limit"}`)
	require.Equal(t, http.StatusOK, overLimit.Code)
	assert.Contains(t, overLimit.Body.String(), `"success":false`)
	assert.Contains(t, overLimit.Body.String(), "invalid registration invite request")
	for _, body := range []string{
		`{"count":1,"valid_days":366,"note":"overlong lifetime"}`,
		`{"count":1,"expires_at":1,"note":"expired"}`,
		`{"count":1,"valid_days":1,"expires_at":4102444800,"note":"ambiguous lifetime"}`,
	} {
		response := performRegistrationInviteRequest(t, environment.router, http.MethodPost, "/api/registration-invites", environment.rootToken, body)
		require.Equal(t, http.StatusOK, response.Code)
		assert.Contains(t, response.Body.String(), `"success":false`)
		assert.Contains(t, response.Body.String(), "invalid registration invite request")
	}

	var count int64
	require.NoError(t, environment.database.Model(&model.RegistrationInvite{}).Count(&count).Error)
	assert.Equal(t, int64(101), count)

	list := performRegistrationInviteRequest(t, environment.router, http.MethodGet, "/api/registration-invites?status=active&page_size=100", environment.rootToken, "")
	require.Equal(t, http.StatusOK, list.Code)
	assert.NotContains(t, list.Body.String(), rawCode)
	assert.NotContains(t, list.Body.String(), "code_hash")

	var logs []model.Log
	require.NoError(t, environment.database.Where("type = ?", model.LogTypeManage).Find(&logs).Error)
	require.NotEmpty(t, logs)
	for _, auditLog := range logs {
		assert.NotContains(t, auditLog.Content, rawCode)
		assert.NotContains(t, auditLog.Other, rawCode)
	}
}

func TestRegistrationInviteListRejectsUnknownStatusAndInvalidPageSize(t *testing.T) {
	environment := setupRegistrationInviteRouterTest(t)
	for _, path := range []string{
		"/api/registration-invites?status=unknown",
		"/api/registration-invites?page_size=101",
		"/api/registration-invites?created_after=20&created_before=10",
		"/api/registration-invites?note=" + strings.Repeat("x", 256),
	} {
		response := performRegistrationInviteRequest(t, environment.router, http.MethodGet, path, environment.rootToken, "")
		require.Equal(t, http.StatusOK, response.Code)
		assert.Contains(t, response.Body.String(), `"success":false`)
		assert.Contains(t, response.Body.String(), "invalid registration invite request")
	}
}

func TestRegistrationInviteRevokeRejectsUsedAndIsIdempotent(t *testing.T) {
	environment := setupRegistrationInviteRouterTest(t)
	used, err := service.GenerateRegistrationInvites(1, common.GetTimestamp()+3600, "used", environment.root.Id, registrationInviteAPITestSecret)
	require.NoError(t, err)
	require.NoError(t, environment.database.Transaction(func(tx *gorm.DB) error {
		return service.ConsumeRegistrationInvite(tx, used[0].Id, environment.root.Id)
	}))

	usedResponse := performRegistrationInviteRequest(t, environment.router, http.MethodPost, fmt.Sprintf("/api/registration-invites/%d/revoke", used[0].Id), environment.rootToken, "")
	require.Equal(t, http.StatusOK, usedResponse.Code)
	assert.Contains(t, usedResponse.Body.String(), `"success":false`)
	assert.Contains(t, usedResponse.Body.String(), "used registration invites cannot be revoked")

	available, err := service.GenerateRegistrationInvites(1, common.GetTimestamp()+3600, "available", environment.root.Id, registrationInviteAPITestSecret)
	require.NoError(t, err)
	firstRevoke := performRegistrationInviteRequest(t, environment.router, http.MethodPost, fmt.Sprintf("/api/registration-invites/%d/revoke", available[0].Id), environment.rootToken, "")
	require.Equal(t, http.StatusOK, firstRevoke.Code)
	assert.Contains(t, firstRevoke.Body.String(), `"revoked":true`)

	secondRevoke := performRegistrationInviteRequest(t, environment.router, http.MethodPost, fmt.Sprintf("/api/registration-invites/%d/revoke", available[0].Id), environment.rootToken, "")
	require.Equal(t, http.StatusOK, secondRevoke.Code)
	assert.Contains(t, secondRevoke.Body.String(), `"success":true`)
	assert.Contains(t, secondRevoke.Body.String(), `"revoked":false`)

	var stored model.RegistrationInvite
	require.NoError(t, environment.database.First(&stored, available[0].Id).Error)
	require.NotNil(t, stored.RevokedBy)
	assert.Equal(t, environment.root.Id, *stored.RevokedBy)

	var logs []model.Log
	require.NoError(t, environment.database.Where("type = ?", model.LogTypeManage).Find(&logs).Error)
	for _, auditLog := range logs {
		assert.NotContains(t, auditLog.Content, used[0].Code)
		assert.NotContains(t, auditLog.Other, used[0].Code)
		assert.NotContains(t, auditLog.Content, available[0].Code)
		assert.NotContains(t, auditLog.Other, available[0].Code)
	}
}
