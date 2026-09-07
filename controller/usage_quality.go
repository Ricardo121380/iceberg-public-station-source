package controller

import (
	"context"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// GetUsageQuality serves both routes; the self route always forces session identity.
func GetUsageQuality(c *gin.Context) {
	start, end, ok := parseFlowQuotaTimeRange(c)
	if !ok {
		return
	}
	if end-start > 30*24*3600 {
		common.ApiErrorMsg(c, "时间跨度不能超过 1 个月")
		return
	}
	userID := 0
	username := c.Query("username")
	if c.FullPath() == "/api/data/quality/self" || c.GetInt("role") < common.RoleAdminUser {
		userID = c.GetInt("id")
		if userID <= 0 {
			c.AbortWithStatus(401)
			return
		}
		username = ""
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	data, err := model.GetUsageQuality(ctx, start, end, username, userID)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
