package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

type abuseSettingsRequest struct {
	Mode          string   `json:"mode"`
	Limit10m      int      `json:"limit_10m"`
	Limit24h      int      `json:"limit_24h"`
	FreezeMinutes int      `json:"freeze_minutes"`
	EnabledRules  []string `json:"enabled_rules"`
}

func GetAbuseSettings(c *gin.Context) {
	p, err := model.GetAbusePolicy()
	if err != nil {
		abuseDatabaseError(c)
		return
	}
	var enabled []string
	if err = common.UnmarshalJsonStr(p.EnabledRules, &enabled); err != nil {
		abuseDatabaseError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"mode": p.Mode, "limit_10m": p.Limit10m, "limit_24h": p.Limit24h, "freeze_minutes": p.FreezeMinutes, "enabled_rules": enabled, "rules": service.SafetyRules(), "generation": p.Generation}})
}

func UpdateAbuseSettings(c *gin.Context) {
	var req abuseSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.invalid_settings")})
		return
	}
	for _, id := range req.EnabledRules {
		valid := false
		for _, r := range service.SafetyRules() {
			if r.ID == id && r.Verified && r.Evidence != "" {
				valid = true
			}
		}
		if !valid {
			c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.unverified_rule")})
			return
		}
	}
	if req.Mode != "off" && req.Mode != "observe" && req.Mode != "enforce" || req.Limit10m < 1 || req.Limit10m > 1000 || req.Limit24h < 1 || req.Limit24h > 10000 || req.FreezeMinutes < 1 || req.FreezeMinutes > 1440 {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.invalid_settings")})
		return
	}
	rules, err := common.Marshal(req.EnabledRules)
	if err != nil {
		abuseDatabaseError(c)
		return
	}
	err = model.SaveAbusePolicy(model.AbusePolicy{Mode: req.Mode, Limit10m: req.Limit10m, Limit24h: req.Limit24h, FreezeMinutes: req.FreezeMinutes, EnabledRules: string(rules)}, c.GetInt("id"))
	if err != nil {
		abuseDatabaseError(c)
		return
	}
	GetAbuseSettings(c)
}

func abuseDatabaseError(c *gin.Context) {
	common.SysError("abuse_control: admin_state_unavailable")
	c.JSON(503, gin.H{"success": false, "message": i18n.T(c, "abuse.unavailable")})
}

func abusePagination(c *gin.Context) (int, int, bool) {
	page, err := strconv.Atoi(c.DefaultQuery("p", "1"))
	if err != nil || page < 1 || page > 100000 {
		return 0, 0, false
	}
	size, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil || size < 1 || size > 100 {
		return 0, 0, false
	}
	return (page - 1) * size, size, true
}

func GetAbuseEvents(c *gin.Context) {
	offset, size, ok := abusePagination(c)
	if !ok {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.invalid_pagination")})
		return
	}
	q := model.DB.Model(&model.AbuseEvent{})
	for _, field := range []string{"user_id", "start", "end"} {
		if raw := c.Query(field); raw != "" {
			n, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || n < 0 {
				c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.invalid_filter")})
				return
			}
			switch field {
			case "user_id":
				q = q.Where("user_id = ?", n)
			case "start":
				q = q.Where("created_at >= ?", n)
			case "end":
				q = q.Where("created_at <= ?", n)
			}
		}
	}
	for _, field := range []string{"category", "rule_id", "action"} {
		if value := c.Query(field); value != "" {
			if len(value) > 80 {
				c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.invalid_filter")})
				return
			}
			q = q.Where(field+" = ?", value)
		}
	}
	var total int64
	var events []model.AbuseEvent
	if q.Count(&total).Error != nil || q.Order("id DESC").Limit(size).Offset(offset).Find(&events).Error != nil {
		abuseDatabaseError(c)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"items": events, "total": total}})
}

func GetAbuseUsers(c *gin.Context) {
	offset, size, ok := abusePagination(c)
	if !ok {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.invalid_pagination")})
		return
	}
	q := model.DB.Model(&model.AbuseState{}).Where("blocked_until > ?", time.Now().Unix())
	var total int64
	var states []model.AbuseState
	if q.Count(&total).Error != nil || q.Order("blocked_until DESC").Limit(size).Offset(offset).Find(&states).Error != nil {
		abuseDatabaseError(c)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"items": states, "total": total}})
}

func UnfreezeAbuseUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil || userID <= 0 {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "common.invalid_id")})
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if c.ShouldBindJSON(&req) != nil || len([]rune(strings.TrimSpace(req.Reason))) == 0 || len([]rune(req.Reason)) > 300 {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.reason_required")})
		return
	}
	if err = model.UnfreezeAbuseUser(userID, c.GetInt("id"), strings.TrimSpace(req.Reason)); err != nil {
		c.JSON(409, gin.H{"success": false, "message": i18n.T(c, "abuse.release_failed")})
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func GetSelfAbuse(c *gin.Context) {
	state, err := model.GetAbuseState(c.GetInt("id"))
	if err != nil {
		abuseDatabaseError(c)
		return
	}
	blocked := state.BlockedUntil > time.Now().Unix()
	c.JSON(200, gin.H{"success": true, "data": gin.H{"suspended": blocked, "blocked_until": state.BlockedUntil}})
}
