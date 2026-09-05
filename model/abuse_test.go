package model

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func abuseDB(t *testing.T, dialector gorm.Dialector) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.False(t, db.Migrator().HasTable(&AbusePolicy{}), "use a fresh isolated test database")
	require.False(t, db.Migrator().HasTable(&User{}), "refusing to use a database containing users")
	prevDB, prevRedis := DB, common.RDB
	oldType := common.MainDatabaseType()
	switch dialector.Name() {
	case "mysql":
		common.SetMainDatabaseType(common.DatabaseTypeMySQL)
	case "postgres":
		common.SetMainDatabaseType(common.DatabaseTypePostgreSQL)
	default:
		common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	}
	DB = db
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	server := miniredis.RunT(t)
	redisAddress := server.Addr()
	if configured := os.Getenv("ABUSE_TEST_REDIS_ADDR"); configured != "" {
		redisAddress = configured
	}
	common.RDB = redis.NewClient(&redis.Options{Addr: redisAddress})
	require.NoError(t, common.RDB.Ping(t.Context()).Err())
	t.Cleanup(func() {
		_ = common.RDB.Close()
		DB = prevDB
		common.SetMainDatabaseType(oldType)
		common.RDB = prevRedis
		for _, v := range []any{&AbuseEvent{}, &AbuseState{}, &AbusePolicy{}, &Token{}, &Log{}, &User{}} {
			_ = db.Migrator().DropTable(v)
		}
		_ = sqlDB.Close()
	})
	// First exercise a genuinely empty database, including repeated migration.
	require.NoError(t, MigrateAbuse(db))
	require.NoError(t, MigrateAbuse(db))
	var fresh AbusePolicy
	require.NoError(t, db.First(&fresh, 1).Error)
	assert.Equal(t, "observe", fresh.Mode)
	for _, table := range []any{&AbuseEvent{}, &AbuseState{}, &AbusePolicy{}} {
		require.NoError(t, db.Migrator().DropTable(table))
	}
	var version string
	query := "SELECT version()"
	if db.Dialector.Name() == "sqlite" {
		query = "SELECT sqlite_version()"
	}
	require.NoError(t, db.Raw(query).Scan(&version).Error)
	t.Logf("database version: %s", version)
	// Representative released schema and data exist before installing abuse tables.
	require.NoError(t, db.AutoMigrate(&User{}, &Token{}, &Log{}))
	require.NoError(t, db.Create(&User{Id: 1, Username: "abuse-test", Role: common.RoleCommonUser, Status: common.UserStatusEnabled, Quota: 123, AuthVersion: 1}).Error)
	require.NoError(t, db.Create(&Token{UserId: 1, Key: "isolated-fixture", Status: 1, RemainQuota: 100}).Error)
	require.NoError(t, db.Create(&Log{UserId: 1, RequestId: "legacy-log", Content: "legacy", CreatedAt: time.Now().Unix()}).Error)
	require.NoError(t, MigrateAbuse(db))
	require.NoError(t, MigrateAbuse(db))
	var user User
	require.NoError(t, db.First(&user, 1).Error)
	assert.Equal(t, 123, user.Quota)
	var n int64
	require.NoError(t, db.Model(&Token{}).Count(&n).Error)
	assert.EqualValues(t, 1, n)
	require.NoError(t, db.Model(&Log{}).Count(&n).Error)
	assert.EqualValues(t, 1, n)
	return db
}

func abuseFixture(id string, p AbusePolicy, round int64) AbuseEvent {
	return AbuseEvent{UserID: 1, TokenID: 1, RequestID: id, ChannelID: 1, Protocol: "error", RuleID: "test_verified", RuleVersion: "fixture", Category: "sexual_explicit", Signal: "fixture", Summary: "Synthetic input policy block", Actionable: true, Generation: p.Generation, Round: round, CreatedAt: time.Now().Unix()}
}

func exerciseAbuseLifecycle(t *testing.T, db *gorm.DB) {
	p, err := GetAbusePolicy()
	require.NoError(t, err)
	assert.Equal(t, "observe", p.Mode)
	for i := 0; i < 3; i++ {
		e := abuseFixture(fmt.Sprintf("observe-%d", i), p, 1)
		e.TokenID = i + 1
		require.NoError(t, RecordAbuseEvent(&e))
		if i == 2 {
			assert.Equal(t, "threshold_observed", e.Action)
			assert.True(t, e.Notify)
		}
	}
	state, err := GetAbuseState(1)
	require.NoError(t, err)
	assert.Zero(t, state.BlockedUntil)
	// Policy update starts a new epoch: observation counts are not punished.
	p.Mode = "enforce"
	require.NoError(t, SaveAbusePolicy(p, 1))
	p, err = GetAbusePolicy()
	require.NoError(t, err)
	first := abuseFixture("enforce-1", p, 1)
	require.NoError(t, RecordAbuseEvent(&first))
	assert.Equal(t, 1, first.Count10m)
	duplicate := abuseFixture("enforce-1", p, 1)
	require.NoError(t, RecordAbuseEvent(&duplicate))
	assert.Equal(t, first.ID, duplicate.ID)
	// Clear Redis to prove reconstruction from durable events.
	require.NoError(t, common.RDB.FlushDB(t.Context()).Err())
	second := abuseFixture("enforce-2", p, 1)
	require.NoError(t, RecordAbuseEvent(&second))
	assert.Equal(t, 2, second.Count10m)
	third := abuseFixture("enforce-3", p, 1)
	require.NoError(t, RecordAbuseEvent(&third))
	assert.Equal(t, "frozen", third.Action)
	state, err = GetAbuseState(1)
	require.NoError(t, err)
	assert.Greater(t, state.BlockedUntil, time.Now().Unix())
	assert.EqualValues(t, 2, state.Round)
	late := abuseFixture("in-flight", p, 1)
	require.NoError(t, RecordAbuseEvent(&late))
	assert.Equal(t, "in_flight", late.Action)
	assert.False(t, late.Actionable)
	until := state.BlockedUntil
	require.NoError(t, common.RDB.FlushDB(t.Context()).Err())
	state, err = GetAbuseState(1)
	require.NoError(t, err)
	assert.Equal(t, until, state.BlockedUntil)
	require.NoError(t, UnfreezeAbuseUser(1, 100, "reviewed fixture"))
	state, err = GetAbuseState(1)
	require.NoError(t, err)
	assert.Zero(t, state.BlockedUntil)
	e := abuseFixture("after-release", p, state.Round)
	require.NoError(t, RecordAbuseEvent(&e))
	assert.Equal(t, 1, e.Count10m)
	// Counter errors preserve evidence, without a suspension.
	rdb := common.RDB
	common.RDB = nil
	e = abuseFixture("redis-down", p, state.Round)
	err = RecordAbuseEvent(&e)
	common.RDB = rdb
	require.NoError(t, err)
	assert.Equal(t, "counter_unavailable", e.Action)
	assert.True(t, e.Notify)
	// Old policy generation is never retroactively actionable.
	e = abuseFixture("old-epoch", p, state.Round)
	e.Generation--
	require.NoError(t, RecordAbuseEvent(&e))
	assert.False(t, e.Actionable)
	// Cleaning events leaves existing state and unrelated legacy data intact.
	old := AbuseEvent{UserID: 1, RequestID: "old", CreatedAt: time.Now().Unix() - 31*86400}
	require.NoError(t, db.Create(&old).Error)
	require.NoError(t, CleanupAbuseEvents(time.Now().Unix()))
	assert.ErrorIs(t, db.First(&AbuseEvent{}, old.ID).Error, gorm.ErrRecordNotFound)
	require.NoError(t, MigrateAbuse(db))
	persisted, err := GetAbusePolicy()
	require.NoError(t, err)
	assert.Equal(t, "enforce", persisted.Mode)
	if db.Dialector.Name() != "sqlite" {
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.SetMaxOpenConns(5)
		state, err := GetAbuseState(1)
		require.NoError(t, err)
		var wg sync.WaitGroup
		errs := make([]error, 8)
		for i := range errs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				e := abuseFixture(fmt.Sprintf("matrix-concurrent-%d", i), persisted, state.Round)
				errs[i] = RecordAbuseEvent(&e)
			}(i)
		}
		wg.Wait()
		for _, err := range errs {
			require.NoError(t, err)
		}
		var frozen int64
		require.NoError(t, db.Model(&AbuseEvent{}).Where("request_id LIKE ? AND action = ?", "matrix-concurrent-%", "frozen").Count(&frozen).Error)
		assert.EqualValues(t, 1, frozen)
	}
}

func TestAbuseDatabaseMatrix(t *testing.T) {
	t.Run("sqlite", func(t *testing.T) { exerciseAbuseLifecycle(t, abuseDB(t, sqlite.Open(":memory:"))) })
	for _, engine := range []string{"MYSQL", "POSTGRES"} {
		t.Run(engine, func(t *testing.T) {
			dsn := os.Getenv("ABUSE_TEST_" + engine + "_DSN")
			if dsn == "" {
				t.Skip("isolated database DSN not configured")
			}
			require.Contains(t, dsn, "abuse_test", "refusing non-test database")
			var dialect gorm.Dialector = postgres.Open(dsn)
			if engine == "MYSQL" {
				dialect = mysql.Open(dsn)
			}
			exerciseAbuseLifecycle(t, abuseDB(t, dialect))
		})
	}
}

func TestAbuseConcurrentFreezeAndAdminExemption(t *testing.T) {
	db := abuseDB(t, sqlite.Open(":memory:"))
	p, err := GetAbusePolicy()
	require.NoError(t, err)
	p.Mode = "enforce"
	require.NoError(t, SaveAbusePolicy(p, 1))
	p, err = GetAbusePolicy()
	require.NoError(t, err)
	var wg sync.WaitGroup
	events := make([]AbuseEvent, 10)
	errs := make([]error, 10)
	for i := range events {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			events[i] = abuseFixture(fmt.Sprintf("concurrent-%d", i), p, 1)
			errs[i] = RecordAbuseEvent(&events[i])
		}(i)
	}
	wg.Wait()
	freezes := 0
	for i, e := range events {
		require.NoError(t, errs[i])
		if e.Action == "frozen" {
			freezes++
		}
	}
	assert.Equal(t, 1, freezes)
	require.NoError(t, UnfreezeAbuseUser(1, 100, "test"))
	require.NoError(t, db.Model(&User{}).Where("id = ?", 1).Update("role", common.RoleRootUser).Error)
	state, err := GetAbuseState(1)
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		e := abuseFixture(fmt.Sprintf("root-%d", i), p, state.Round)
		require.NoError(t, RecordAbuseEvent(&e))
		assert.NotEqual(t, "frozen", e.Action)
	}
	state, err = GetAbuseState(1)
	require.NoError(t, err)
	assert.Zero(t, state.BlockedUntil)
}

func TestAbuseWindowBoundariesAndExpiry(t *testing.T) {
	db := abuseDB(t, sqlite.Open(":memory:"))
	p, err := GetAbusePolicy()
	require.NoError(t, err)
	p.Mode = "enforce"
	require.NoError(t, SaveAbusePolicy(p, 1))
	p, err = GetAbusePolicy()
	require.NoError(t, err)
	now := time.Now().Unix()
	for i := 0; i < 7; i++ {
		e := abuseFixture(fmt.Sprintf("earlier-%d", i), p, 1)
		e.CreatedAt = now - 601 - int64(i)
		require.NoError(t, db.Create(&e).Error)
	}
	expired := abuseFixture("boundary", p, 1)
	expired.CreatedAt = now - 86400
	require.NoError(t, db.Create(&expired).Error)
	e := abuseFixture("eighth", p, 1)
	e.CreatedAt = now
	require.NoError(t, RecordAbuseEvent(&e))
	assert.Equal(t, 1, e.Count10m)
	assert.Equal(t, 8, e.Count24h)
	assert.Equal(t, "frozen", e.Action)
	require.NoError(t, db.Model(&AbuseState{}).Where("user_id = ?", 1).Update("blocked_until", now-1).Error)
	state, err := GetAbuseState(1)
	require.NoError(t, err)
	e = abuseFixture("after-expiry", p, state.Round)
	require.NoError(t, RecordAbuseEvent(&e))
	assert.Equal(t, 1, e.Count24h)
	assert.Equal(t, "recorded", e.Action)
	assert.NotContains(t, strings.ToLower(e.Summary), "sk-")
}
