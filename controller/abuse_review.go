package controller

import (
	"errors"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func abuseReviewError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(404, gin.H{"success": false, "message": i18n.T(c, "abuse.event_missing")})
	case errors.Is(err, model.ErrAbuseReviewConflict):
		c.JSON(409, gin.H{"success": false, "message": i18n.T(c, "abuse.review_conflict")})
	case errors.Is(err, model.ErrAbuseReviewInvalid):
		c.JSON(400, gin.H{"success": false, "message": i18n.T(c, "abuse.review_invalid")})
	default:
		abuseDatabaseError(c)
	}
}

func ReadAbuseEvidence(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		abuseReviewError(c, model.ErrAbuseReviewInvalid)
		return
	}
	excerpt, status, expires, err := service.ReadAbuseExcerpt(id, c.GetInt("id"), time.Now().Unix())
	if err != nil {
		abuseReviewError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"status": status, "expires_at": expires, "excerpt": excerpt}})
}

func SaveAbuseReview(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		abuseReviewError(c, model.ErrAbuseReviewInvalid)
		return
	}
	var body struct {
		Decision string `json:"decision"`
		Note     string `json:"note"`
		Version  *int64 `json:"version"`
	}
	if c.ShouldBindJSON(&body) != nil || body.Version == nil {
		abuseReviewError(c, model.ErrAbuseReviewInvalid)
		return
	}
	if err = model.ReviewAbuseEvent(id, c.GetInt("id"), *body.Version, body.Decision, body.Note, time.Now().Unix()); err != nil {
		abuseReviewError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func GetAbuseReviewHistory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		abuseReviewError(c, model.ErrAbuseReviewInvalid)
		return
	}
	offset, size, ok := abusePagination(c)
	if !ok {
		abuseReviewError(c, model.ErrAbuseReviewInvalid)
		return
	}
	var event model.AbuseEvent
	if err = model.DB.First(&event, id).Error; err != nil {
		abuseReviewError(c, err)
		return
	}
	var items []model.AbuseReviewAudit
	var total int64
	q := model.DB.Model(&model.AbuseReviewAudit{}).Where("event_id = ?", id)
	if err = q.Count(&total).Error; err == nil {
		err = q.Order("id DESC").Offset(offset).Limit(size).Find(&items).Error
	}
	if err != nil {
		abuseReviewError(c, err)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"items": items, "total": total}})
}

func abuseCaptureReady() bool {
	_, err := common.AbuseEvidenceCipher()
	return err == nil
}
