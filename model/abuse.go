package model

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AbusePolicy is a singleton. Its write lock orders policy changes and decisions
// across processes, including SQLite (which has no SELECT FOR UPDATE).
type AbusePolicy struct {
	ID            int    `json:"-" gorm:"primaryKey"`
	Mode          string `json:"mode" gorm:"size:16"`
	Limit10m      int    `json:"limit_10m"`
	Limit24h      int    `json:"limit_24h"`
	FreezeMinutes int    `json:"freeze_minutes"`
	EnabledRules  string `json:"-" gorm:"type:text"`
	Generation    int64  `json:"generation"`
	Revision      int64  `json:"revision"`
	UpdatedAt     int64  `json:"updated_at"`
}

type AbuseState struct {
	UserID           int    `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	Round            int64  `json:"round"`
	BlockedUntil     int64  `json:"blocked_until" gorm:"index"`
	TriggerRequestID string `json:"trigger_request_id" gorm:"size:64"`
	UpdatedAt        int64  `json:"updated_at"`
}

type AbuseEvent struct {
	ID           int64  `json:"id" gorm:"primaryKey"`
	UserID       int    `json:"user_id" gorm:"uniqueIndex:idx_abuse_request,priority:1;index:idx_abuse_window,priority:1"`
	RequestID    string `json:"request_id" gorm:"size:64;uniqueIndex:idx_abuse_request,priority:2"`
	TokenID      int    `json:"token_id"`
	ChannelID    int    `json:"channel_id"`
	Model        string `json:"model" gorm:"size:255"`
	Protocol     string `json:"protocol" gorm:"size:32"`
	RuleID       string `json:"rule_id" gorm:"size:80"`
	RuleVersion  string `json:"rule_version" gorm:"size:32"`
	Category     string `json:"category" gorm:"size:32"`
	Signal       string `json:"signal" gorm:"size:100"`
	Summary      string `json:"summary" gorm:"type:text"`
	Attempts     string `json:"attempts" gorm:"type:text"`
	Actionable   bool   `json:"actionable"`
	Generation   int64  `json:"generation" gorm:"index:idx_abuse_window,priority:2"`
	Round        int64  `json:"round" gorm:"index:idx_abuse_window,priority:3"`
	CreatedAt    int64  `json:"created_at" gorm:"index:idx_abuse_window,priority:4;index:idx_abuse_created"`
	Action       string `json:"action" gorm:"size:32"`
	Count10m     int    `json:"count_10m" gorm:"column:count10m"`
	Count24h     int    `json:"count_24h" gorm:"column:count24h"`
	BlockedUntil int64  `json:"blocked_until"`
	OperatorID   int    `json:"operator_id"`
	Notify       bool   `json:"-" gorm:"index"`
	NotifiedAt   int64  `json:"-"`
}

func MigrateAbuse(db *gorm.DB) error {
	if err := db.AutoMigrate(&AbusePolicy{}, &AbuseState{}, &AbuseEvent{}); err != nil {
		return err
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&AbusePolicy{ID: 1, Mode: "observe", Limit10m: 3, Limit24h: 8, FreezeMinutes: 60, EnabledRules: "[]", Generation: 1, Revision: 1}).Error
}

func GetAbusePolicy() (AbusePolicy, error) {
	var p AbusePolicy
	err := DB.First(&p, 1).Error
	return p, err
}

func lockAbusePolicy(tx *gorm.DB) (AbusePolicy, error) {
	// A real write serializes on all supported engines; increment is transactional.
	if err := tx.Model(&AbusePolicy{}).Where("id = ?", 1).UpdateColumn("revision", gorm.Expr("revision + 1")).Error; err != nil {
		return AbusePolicy{}, err
	}
	var p AbusePolicy
	err := tx.First(&p, 1).Error
	return p, err
}

func SaveAbusePolicy(p AbusePolicy, operator int) error {
	if p.Mode != "off" && p.Mode != "observe" && p.Mode != "enforce" {
		return errors.New("invalid abuse mode")
	}
	if p.Limit10m < 1 || p.Limit10m > 1000 || p.Limit24h < 1 || p.Limit24h > 10000 || p.FreezeMinutes < 1 || p.FreezeMinutes > 1440 {
		return errors.New("invalid abuse limits")
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		old, err := lockAbusePolicy(tx)
		if err != nil {
			return err
		}
		// Any mode transition starts a new epoch; outstanding requests cannot be punished retroactively.
		p.ID = 1
		p.Generation = old.Generation
		p.Revision = old.Revision
		p.UpdatedAt = time.Now().Unix()
		if old.Mode != p.Mode || old.EnabledRules != p.EnabledRules || old.Limit10m != p.Limit10m || old.Limit24h != p.Limit24h || old.FreezeMinutes != p.FreezeMinutes {
			p.Generation++
		}
		if err := tx.Save(&p).Error; err != nil {
			return err
		}
		return tx.Create(&AbuseEvent{UserID: operator, OperatorID: operator, RequestID: common.NewRequestId(), CreatedAt: p.UpdatedAt, Action: "settings_changed", Summary: "Safety policy updated", Category: "management"}).Error
	})
}

func GetAbuseState(userID int) (AbuseState, error) {
	s := AbuseState{UserID: userID, Round: 1}
	err := DB.First(&s, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AbuseState{UserID: userID, Round: 1}, nil
	}
	return s, err
}

// Redis is a derived index. Rebuild inside the database decision lock to avoid
// stale rebuilds after restart and to recover from a rolled-back Redis write.
const abuseWindowScript = `
redis.call('DEL', KEYS[1])
for i=2,#ARGV,2 do redis.call('ZADD',KEYS[1],'NX',ARGV[i],ARGV[i+1]) end
local now=tonumber(ARGV[1])
redis.call('ZREMRANGEBYSCORE',KEYS[1],'-inf',now-86400)
local short=redis.call('ZCOUNT',KEYS[1],'('..(now-600),'+inf')
local long=redis.call('ZCARD',KEYS[1])
redis.call('EXPIRE',KEYS[1],90000)
return {short,long}
`

func countAbuseWindow(tx *gorm.DB, e *AbuseEvent) (int, int, error) {
	if common.RDB == nil {
		return 0, 0, errors.New("abuse counter unavailable")
	}
	var recent []AbuseEvent
	if err := tx.Select("request_id", "created_at").Where("user_id = ? AND generation = ? AND round = ? AND actionable = ? AND created_at > ? AND created_at <= ?", e.UserID, e.Generation, e.Round, true, e.CreatedAt-86400, e.CreatedAt).Find(&recent).Error; err != nil {
		return 0, 0, err
	}
	args := []interface{}{e.CreatedAt}
	for _, v := range recent {
		args = append(args, v.CreatedAt, v.RequestID)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := fmt.Sprintf("abuse:safety:{%d}:%d:%d", e.UserID, e.Generation, e.Round)
	values, err := common.RDB.Eval(ctx, abuseWindowScript, []string{key}, args...).Slice()
	if err != nil {
		return 0, 0, err
	}
	if len(values) != 2 {
		return 0, 0, errors.New("invalid abuse counter result")
	}
	short, ok := values[0].(int64)
	if !ok {
		return 0, 0, errors.New("invalid abuse count")
	}
	long, ok := values[1].(int64)
	if !ok {
		return 0, 0, errors.New("invalid abuse count")
	}
	return int(short), int(long), nil
}

// RecordAbuseEvent commits evidence even if the counter fails. No state change
// is made until the current user role and admission epoch have been checked.
func RecordAbuseEvent(e *AbuseEvent) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		p, err := lockAbusePolicy(tx)
		if err != nil {
			return err
		}
		var previous AbuseEvent
		err = tx.Where("user_id = ? AND request_id = ?", e.UserID, e.RequestID).First(&previous).Error
		if err == nil {
			*e = previous
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		e.Action = "recorded"
		var user User
		if err = lockForUpdate(tx).Select("id", "role").First(&user, e.UserID).Error; err != nil {
			return err
		}
		state := AbuseState{UserID: e.UserID, Round: 1}
		if err = tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&state).Error; err != nil {
			return err
		}
		if err = tx.First(&state, "user_id = ?", e.UserID).Error; err != nil {
			return err
		}
		now := time.Now().Unix()
		if e.CreatedAt == 0 {
			e.CreatedAt = now
		}
		eligible := p.Mode != "off" && e.Generation == p.Generation && e.Round == state.Round && state.BlockedUntil <= now
		if !eligible {
			e.Action = "in_flight"
			e.Actionable = false
		}
		if err = tx.Create(e).Error; err != nil {
			return err
		}
		if !e.Actionable {
			return nil
		}
		e.Count10m, e.Count24h, err = countAbuseWindow(tx, e)
		if err != nil {
			e.Action = "counter_unavailable"
			e.Notify = true
			common.SysError("abuse_control: counter_unavailable")
		} else if e.Count10m >= p.Limit10m || e.Count24h >= p.Limit24h {
			e.Action = "threshold_observed"
			if p.Mode == "enforce" && user.Role < common.RoleAdminUser {
				e.Action = "frozen"
				e.Notify = true
				e.BlockedUntil = now + int64(p.FreezeMinutes)*60
				state.BlockedUntil = e.BlockedUntil
				state.TriggerRequestID = e.RequestID
				state.Round++
				state.UpdatedAt = now
				if err = tx.Save(&state).Error; err != nil {
					return err
				}
			}
		}
		if e.Action == "counter_unavailable" || e.Action == "threshold_observed" {
			var count int64
			if err = tx.Model(&AbuseEvent{}).Where("user_id = ? AND rule_id = ? AND action = ? AND notify = ? AND created_at > ?", e.UserID, e.RuleID, e.Action, true, now-3600).Count(&count).Error; err != nil {
				return err
			}
			e.Notify = count == 0
		}
		return tx.Save(e).Error
	})
}

var ErrAbuseNotSuspended = errors.New("account is not suspended")

func UnfreezeAbuseUser(userID, operator int, reason string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		_, err := lockAbusePolicy(tx)
		if err != nil {
			return err
		}
		var state AbuseState
		if err = tx.First(&state, "user_id = ?", userID).Error; err != nil {
			return err
		}
		if state.BlockedUntil <= time.Now().Unix() {
			return ErrAbuseNotSuspended
		}
		state.Round++
		state.BlockedUntil = 0
		state.UpdatedAt = time.Now().Unix()
		if err = tx.Save(&state).Error; err != nil {
			return err
		}
		return tx.Create(&AbuseEvent{UserID: userID, OperatorID: operator, RequestID: common.NewRequestId(), CreatedAt: state.UpdatedAt, Action: "unfrozen", Summary: reason, Category: "management", Notify: true}).Error
	})
}

// CleanupAbuseEvents uses portable bounded deletes. Notification retention is
// also bounded to 30 days; active suspension state is never removed.
func CleanupAbuseEvents(now int64) error {
	for {
		var ids []int64
		if err := DB.Model(&AbuseEvent{}).Where("created_at < ?", now-30*86400).Order("id").Limit(500).Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if err := DB.Where("id IN ?", ids).Delete(&AbuseEvent{}).Error; err != nil {
			return err
		}
	}
}
