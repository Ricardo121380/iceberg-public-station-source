package middleware

import (
	"context"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// Observation never changes a response or account access. Redis failure must
// not turn an otherwise valid relay request into a failure.
func AnomalyObservation() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if c.Request.Method != "POST" || c.Writer.Status() < 400 || c.GetInt("id") <= 0 || !common.RedisEnabled || common.RDB == nil {
			return
		}
		kind := c.GetString("anomaly_kind")
		if kind == "" {
			kind = service.ClassifyAnomaly(c.Writer.Status(), "", false)
		}
		if kind == "" {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		// Errors are exposed by the admin API when storage is unavailable; avoid
		// per-request error logging that could itself become a log flood.
		_ = service.RecordAnomaly(ctx, c.GetInt("id"), c.GetInt("channel_id"), c.GetString("original_model"), kind)
	}
}
