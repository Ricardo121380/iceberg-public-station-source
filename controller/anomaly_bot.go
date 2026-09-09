package controller

import (
	"context"
	"crypto/subtle"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
	"time"
)

// A dedicated random secret, never a user API key. The host bot uses loopback.
func AnomalyBotAuth(c *gin.Context) {
	expected := os.Getenv("ANOMALY_BOT_SECRET")
	if len(expected) < 32 || subtle.ConstantTimeCompare([]byte(c.GetHeader("X-Anomaly-Bot-Secret")), []byte(expected)) != 1 {
		c.AbortWithStatus(403)
		return
	}
	c.Next()
}
func ApplyAnomalyBotAction(c *gin.Context) {
	var req struct {
		EventID     string `json:"event_id"`
		OperationID string `json:"operation_id"`
		Action      string `json:"action"`
		UserID      int    `json:"user_id"`
		OperatorID  int64  `json:"operator_id"`
	}
	allowed, err := strconv.ParseInt(os.Getenv("ANOMALY_BOT_OPERATOR_ID"), 10, 64)
	if err != nil || allowed <= 0 || c.ShouldBindJSON(&req) != nil || req.OperatorID != allowed || !anomalyIDPattern.MatchString(req.EventID) || !anomalyIDPattern.MatchString(req.OperationID) || req.UserID <= 0 || (req.Action != "pause" && req.Action != "release") {
		c.JSON(400, gin.H{"success": false})
		return
	}
	user, err := model.GetUserById(req.UserID, false)
	if err != nil || user.Role >= 10 || user.Status != 1 {
		c.JSON(409, gin.H{"success": false, "message": "User unavailable or protected"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	until, err := service.ApplyAnomalyAction(ctx, req.EventID, req.OperationID, req.Action, req.UserID, req.OperatorID)
	if err != nil {
		c.JSON(409, gin.H{"success": false, "message": "Event expired or already handled"})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"blocked_until": until}})
}
