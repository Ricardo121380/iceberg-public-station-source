package middleware

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
)

func TestAnomalyMiddlewareObservesFinalResponseOnlyAndNeverEnforces(t *testing.T) {
	old, enabled := common.RDB, common.RedisEnabled
	r := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: r.Addr()})
	common.RDB = client
	common.RedisEnabled = true
	t.Cleanup(func() { client.Close(); common.RDB = old; common.RedisEnabled = enabled })
	router := gin.New()
	router.Use(AnomalyObservation())
	router.POST("/test", func(c *gin.Context) {
		c.Set("id", 42)
		c.Set("anomaly_kind", "invalid_schema")
		c.JSON(400, gin.H{"error": "original"})
	})
	router.POST("/success", func(c *gin.Context) {
		c.Set("id", 42)
		c.Set("anomaly_kind", "invalid_schema")
		c.JSON(200, gin.H{"ok": true})
	})
	for _, path := range []string{"/test", "/success"} {
		req := httptest.NewRequest("POST", path, strings.NewReader(`{}`))
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		if path == "/test" {
			require.Equal(t, 400, resp.Code)
			require.Contains(t, resp.Body.String(), "original")
		} else {
			require.Equal(t, 200, resp.Code)
		}
	}
	events, err := service.ListAnomalies(context.Background())
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, 1, events[0].Count)
	common.RedisEnabled = false
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest("POST", "/success", nil))
	require.Equal(t, 200, resp.Code)
}
