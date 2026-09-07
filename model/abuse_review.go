package model

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type AbuseEvidence struct {
	EventID    int64  `json:"-" gorm:"primaryKey;autoIncrement:false"`
	Ciphertext string `json:"-" gorm:"type:text"`
	ExpiresAt  int64  `json:"-" gorm:"index"`
}

type AbuseReviewAudit struct {
	ID         int64  `json:"id" gorm:"primaryKey"`
	EventID    int64  `json:"event_id" gorm:"index"`
	OperatorID int    `json:"operator_id"`
	Kind       string `json:"kind" gorm:"size:24"`
	Decision   string `json:"decision" gorm:"size:32"`
	Note       string `json:"note" gorm:"type:text"`
	Version    int64  `json:"version"`
	CreatedAt  int64  `json:"created_at" gorm:"index"`
}

var ErrAbuseReviewConflict = errors.New("abuse review changed")
var ErrAbuseReviewInvalid = errors.New("invalid abuse review")

func AccessAbuseEvidence(id int64, operator int, now int64) (AbuseEvent, AbuseEvidence, error) {
	var event AbuseEvent
	var evidence AbuseEvidence
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&event, id).Error; err != nil {
			return err
		}
		if event.EvidenceStatus == "available" && now < event.EvidenceExpiresAt {
			if err := tx.First(&evidence, "event_id = ?", id).Error; err != nil {
				return err
			}
		}
		return tx.Create(&AbuseReviewAudit{EventID: id, OperatorID: operator, Kind: "excerpt_access", CreatedAt: now}).Error
	})
	return event, evidence, err
}

func ReviewAbuseEvent(id int64, operator int, version int64, decision, note string, now int64) error {
	if decision != "confirmed" && decision != "suspected_false_positive" && decision != "insufficient_evidence" {
		return ErrAbuseReviewInvalid
	}
	note = strings.TrimSpace(note)
	if note == "" || !utf8.ValidString(note) || utf8.RuneCountInString(note) > 300 || version < 0 {
		return ErrAbuseReviewInvalid
	}
	note, _ = common.RedactAbuseText(note)
	return DB.Transaction(func(tx *gorm.DB) error {
		var event AbuseEvent
		if err := tx.First(&event, id).Error; err != nil {
			return err
		}
		// Optimistic conditional update also serializes SQLite writes. A repeated
		// submission with the same revision never adds another audit or changes access.
		q := tx.Model(&AbuseEvent{}).Where("id = ? AND COALESCE(review_version, 0) = ?", id, version).Updates(map[string]interface{}{"review_status": decision, "review_version": version + 1, "reviewed_by": operator, "reviewed_at": now})
		if q.Error != nil {
			return q.Error
		}
		if q.RowsAffected != 1 {
			return ErrAbuseReviewConflict
		}
		return tx.Create(&AbuseReviewAudit{EventID: id, OperatorID: operator, Kind: "review", Decision: decision, Note: note, Version: version + 1, CreatedAt: now}).Error
	})
}

func CleanupAbuseEvidence(now int64) error {
	for {
		var ids []int64
		if err := DB.Model(&AbuseEvidence{}).Where("expires_at <= ?", now).Order("event_id").Limit(500).Pluck("event_id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if err := DB.Where("event_id IN ?", ids).Delete(&AbuseEvidence{}).Error; err != nil {
			return err
		}
	}
}
