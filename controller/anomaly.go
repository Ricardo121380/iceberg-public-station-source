package controller

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func anomalyError(c *gin.Context, err error) {
	status := 503
	message := i18n.T(c, "anomaly.unavailable")
	if errors.Is(err, service.ErrAnomalyConflict) {
		status = 409
		message = i18n.T(c, "anomaly.conflict")
	}
	c.JSON(status, gin.H{"success": false, "message": message})
}
func GetAnomalySettings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	p, err := service.GetAnomalyPolicy(ctx)
	if err != nil {
		anomalyError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": p})
}
func UpdateAnomalySettings(c *gin.Context) {
	var p service.AnomalyPolicy
	if c.ShouldBindJSON(&p) != nil || !service.ValidateAnomalyPolicy(p) {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "anomaly.invalid_settings")})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := service.SaveAnomalyPolicy(ctx, p, c.GetInt("id")); err != nil {
		anomalyError(c, err)
		return
	}
	GetAnomalySettings(c)
}
func GetAnomalyEvents(c *gin.Context) {
	page, pageErr := strconv.Atoi(c.DefaultQuery("p", "1"))
	var err error
	user := c.Query("user_id")
	uid := 0
	if user != "" {
		uid, err = strconv.Atoi(user)
		if err != nil || uid < 1 {
			c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "anomaly.invalid_user")})
			return
		}
	}
	if pageErr != nil || page < 1 || page > 100 {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "anomaly.invalid_page")})
		return
	}
	kind := c.Query("kind")
	state := c.Query("state")
	if kind != "" && kind != "invalid_schema" && kind != "rate_limit" && kind != "upstream" && kind != "quota" && kind != "unsupported_model" && kind != "invalid_request" {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "anomaly.invalid_category")})
		return
	}
	if state != "" && state != "pending" && state != "acknowledged" {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "anomaly.invalid_state")})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	all, err := service.ListAnomalies(ctx)
	if err != nil {
		anomalyError(c, err)
		return
	}
	items := make([]service.AnomalyEvent, 0)
	pending := 0
	for _, e := range all {
		if e.Alert && e.Count > e.AcknowledgedCount {
			pending++
		}
		if uid != 0 && uid != e.UserID || kind != "" && kind != e.Kind {
			continue
		}
		if state == "pending" && (!e.Alert || e.Count <= e.AcknowledgedCount) || state == "acknowledged" && (e.AcknowledgedCount == 0 || e.Count > e.AcknowledgedCount) {
			continue
		}
		items = append(items, e)
	}
	total := len(items)
	start := (page - 1) * 20
	if start > total {
		start = total
	}
	end := start + 20
	if end > total {
		end = total
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"items": items[start:end], "total": total, "pending": pending, "retained": len(all)}})
}

var anomalyIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)

func AcknowledgeAnomaly(c *gin.Context) {
	var req struct {
		Count int `json:"count"`
	}
	if !anomalyIDPattern.MatchString(c.Param("id")) || c.ShouldBindJSON(&req) != nil || req.Count < 1 {
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "anomaly.invalid_acknowledgement")})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := service.AcknowledgeAnomaly(ctx, c.Param("id"), req.Count, c.GetInt("id")); err != nil {
		anomalyError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true})
}
func GetAnomalyAudit(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	data, err := service.ListAnomalyAudit(ctx)
	if err != nil {
		anomalyError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": data})
}
