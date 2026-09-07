package controller

import (
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUsageQualityRequestScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	previous := model.LOG_DB
	model.LOG_DB = db
	t.Cleanup(func() { model.LOG_DB = previous; _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&model.Log{}))
	require.NoError(t, db.Create(&[]model.Log{
		{UserId: 1, Username: "one", ModelName: "a", CreatedAt: 3601, Type: model.LogTypeConsume, PromptTokens: 100, Other: `{"cache_tokens":80}`},
		{UserId: 2, Username: "two", ModelName: "a", CreatedAt: 3601, Type: model.LogTypeConsume, PromptTokens: 900, Other: `{"cache_tokens":90}`},
	}).Error)
	tests := []struct {
		name, path, query string
		role, id          int
		success           bool
		tokens            int64
		status            int
	}{
		{"admin all", "/api/data/quality", "start_timestamp=3600&end_timestamp=7200", 10, 1, true, 1000, 200},
		{"admin username filter", "/api/data/quality", "start_timestamp=3600&end_timestamp=7200&username=two", 10, 1, true, 900, 200},
		{"self ignores forged username", "/api/data/quality/self", "start_timestamp=3600&end_timestamp=7200&username=two&user_id=2", 1, 1, true, 100, 200},
		{"admin self stays personal", "/api/data/quality/self", "start_timestamp=3600&end_timestamp=7200", 10, 1, true, 100, 200},
		{"missing identity rejected", "/api/data/quality/self", "start_timestamp=3600&end_timestamp=7200", 1, 0, false, 0, 401},
		{"inverted range", "/api/data/quality", "start_timestamp=7200&end_timestamp=3600", 10, 1, false, 0, 200},
		{"invalid start", "/api/data/quality", "start_timestamp=bad&end_timestamp=7200", 10, 1, false, 0, 200},
		{"unbounded range rejected", "/api/data/quality", "start_timestamp=1&end_timestamp=9999999", 10, 1, false, 0, 200},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set("role", test.role); c.Set("id", test.id) })
			router.GET(test.path, GetUsageQuality)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", test.path+"?"+test.query, nil))
			require.Equal(t, test.status, response.Code)
			if response.Code == 401 {
				return
			}
			var result struct {
				Success bool                       `json:"success"`
				Data    []model.UsageQualityBucket `json:"data"`
			}
			require.NoError(t, common.Unmarshal(response.Body.Bytes(), &result))
			assert.Equal(t, test.success, result.Success)
			if test.success {
				require.Len(t, result.Data, 1)
				assert.Equal(t, test.tokens, result.Data[0].InputTokens)
			} else {
				assert.Empty(t, result.Data)
			}
		})
	}
}
