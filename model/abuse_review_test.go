package model

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"testing"
	"time"
)

type abuseEventBeforeReview struct {
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

func (abuseEventBeforeReview) TableName() string { return "abuse_events" }

func exerciseReviewMigration(t *testing.T, db *gorm.DB) {
	prior := DB
	DB = db
	t.Cleanup(func() { DB = prior })
	require.False(t, db.Migrator().HasTable(&AbuseEvent{}), "isolated empty test database required")
	// Exact released event schema, with data and the existing composite uniqueness index.
	require.NoError(t, db.AutoMigrate(&abuseEventBeforeReview{}))
	old := abuseEventBeforeReview{UserID: 1, RequestID: "legacy-request", Summary: "legacy summary", Action: "recorded", CreatedAt: time.Now().Unix()}
	require.NoError(t, db.Create(&old).Error)
	require.NoError(t, MigrateAbuse(db))
	require.NoError(t, MigrateAbuse(db))
	var event AbuseEvent
	require.NoError(t, db.First(&event, old.ID).Error)
	assert.Equal(t, "legacy summary", event.Summary)
	duplicate := AbuseEvent{UserID: 1, RequestID: "legacy-request"}
	require.Error(t, db.Create(&duplicate).Error)
	require.NoError(t, ReviewAbuseEvent(old.ID, 100, 0, "insufficient_evidence", "generic refusal", 1000))
	require.ErrorIs(t, ReviewAbuseEvent(old.ID, 100, 0, "confirmed", "duplicate", 1001), ErrAbuseReviewConflict)
	require.NoError(t, ReviewAbuseEvent(old.ID, 100, 1, "suspected_false_positive", "password=secret", 1002))
	require.NoError(t, db.First(&event, old.ID).Error)
	assert.EqualValues(t, 2, event.ReviewVersion)
	assert.Equal(t, "recorded", event.Action)
	assert.False(t, event.Actionable)
	assert.False(t, event.Notify)
	var audit []AbuseReviewAudit
	require.NoError(t, db.Order("id").Find(&audit).Error)
	require.Len(t, audit, 2)
	assert.NotContains(t, audit[1].Note, "secret")
	require.ErrorIs(t, ReviewAbuseEvent(old.ID, 100, 2, "invalid", "note", 1003), ErrAbuseReviewInvalid)
	require.NoError(t, db.Create(&AbuseEvidence{EventID: old.ID, Ciphertext: "encrypted-test", ExpiresAt: 1100}).Error)
	require.NoError(t, CleanupAbuseEvidence(1099))
	var count int64
	require.NoError(t, db.Model(&AbuseEvidence{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)
	require.NoError(t, CleanupAbuseEvidence(1100))
	require.NoError(t, db.Model(&AbuseEvidence{}).Count(&count).Error)
	assert.Zero(t, count)
	require.NoError(t, MigrateAbuse(db))
	require.NoError(t, db.First(&event, old.ID).Error)
	assert.EqualValues(t, 2, event.ReviewVersion)
	require.NoError(t, CleanupAbuseEvents(time.Now().Unix()+31*86400))
	require.NoError(t, db.Model(&AbuseReviewAudit{}).Count(&count).Error)
	assert.Zero(t, count)
	// Recreate only the isolated test schema and verify a fresh install twice.
	require.NoError(t, db.Migrator().DropTable(&AbuseEvidence{}, &AbuseReviewAudit{}, &AbuseEvent{}, &AbuseState{}, &AbusePolicy{}))
	// A fresh installation starts a new process/connection, not cached prepared
	// SELECT * plans from the deliberately dropped legacy test schema.
	if db.Dialector.Name() != "sqlite" {
		freshDB, err := gorm.Open(db.Dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		require.NoError(t, err)
		conn, err := freshDB.DB()
		require.NoError(t, err)
		conn.SetMaxOpenConns(1)
		t.Cleanup(func() { _ = conn.Close() })
		db = freshDB
		DB = freshDB
	}
	require.NoError(t, MigrateAbuse(db))
	require.NoError(t, MigrateAbuse(db))
	fresh := AbuseEvent{UserID: 2, RequestID: "fresh-request", CreatedAt: time.Now().Unix(), EvidenceStatus: "available", EvidenceExpiresAt: time.Now().Unix() + 604800}
	require.NoError(t, db.Create(&fresh).Error)
	require.NoError(t, db.Create(&AbuseEvidence{EventID: fresh.ID, Ciphertext: "fresh encrypted data", ExpiresAt: fresh.EvidenceExpiresAt}).Error)
	_, evidence, err := AccessAbuseEvidence(fresh.ID, 100, time.Now().Unix())
	require.NoError(t, err)
	assert.Equal(t, "fresh encrypted data", evidence.Ciphertext)
	require.NoError(t, ReviewAbuseEvent(fresh.ID, 100, 0, "confirmed", "verified separate evidence", time.Now().Unix()))
	require.NoError(t, MigrateAbuse(db))
	require.NoError(t, db.First(&fresh, fresh.ID).Error)
	assert.Equal(t, "confirmed", fresh.ReviewStatus)
}

func TestAbuseReviewMigrationMatrix(t *testing.T) {
	for _, engine := range []string{"SQLITE", "MYSQL", "POSTGRES"} {
		t.Run(engine, func(t *testing.T) {
			var dialect gorm.Dialector = sqlite.Open(":memory:")
			if engine != "SQLITE" {
				dsn := os.Getenv("ABUSE_REVIEW_TEST_" + engine + "_DSN")
				if dsn == "" {
					t.Skip("isolated test DSN not configured")
				}
				require.Contains(t, dsn, "abuse_review_test")
				if engine == "MYSQL" {
					dialect = mysql.Open(dsn)
				} else {
					dialect = postgres.Open(dsn)
				}
			}
			db, err := gorm.Open(dialect, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			sqlDB.SetMaxOpenConns(1)
			t.Cleanup(func() { _ = sqlDB.Close() })
			oldType := common.MainDatabaseType()
			t.Cleanup(func() { common.SetMainDatabaseType(oldType) })
			if engine == "MYSQL" {
				common.SetMainDatabaseType(common.DatabaseTypeMySQL)
			} else if engine == "POSTGRES" {
				common.SetMainDatabaseType(common.DatabaseTypePostgreSQL)
			} else {
				common.SetMainDatabaseType(common.DatabaseTypeSQLite)
			}
			var version string
			query := "SELECT version()"
			if engine == "SQLITE" {
				query = "SELECT sqlite_version()"
			}
			require.NoError(t, db.Raw(query).Scan(&version).Error)
			t.Logf("database version: %s", version)
			exerciseReviewMigration(t, db)
		})
	}
}
