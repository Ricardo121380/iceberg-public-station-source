package model

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type registrationInviteMigrationCounts struct {
	Users    int64
	Channels int64
	Tokens   int64
	Logs     int64
}

func registrationInviteMigrationCountsFor(t *testing.T, db *gorm.DB) registrationInviteMigrationCounts {
	t.Helper()
	counts := registrationInviteMigrationCounts{}
	require.NoError(t, db.Model(&User{}).Count(&counts.Users).Error)
	require.NoError(t, db.Model(&Channel{}).Count(&counts.Channels).Error)
	require.NoError(t, db.Model(&Token{}).Count(&counts.Tokens).Error)
	require.NoError(t, db.Model(&Log{}).Count(&counts.Logs).Error)
	return counts
}

func requireRegistrationInviteMigrationTestDatabaseEmpty(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, model := range []any{
		&RegistrationInvite{},
		&User{},
		&Channel{},
		&Token{},
		&Log{},
	} {
		if db.Migrator().HasTable(model) {
			t.Skip("registration invite migration tests require an empty database")
		}
	}
}

func testRegistrationInviteMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	requireRegistrationInviteMigrationTestDatabaseEmpty(t, db)
	t.Cleanup(func() {
		_ = db.Migrator().DropTable(&RegistrationInvite{})
		_ = db.Migrator().DropTable(&Log{})
		_ = db.Migrator().DropTable(&Token{})
		_ = db.Migrator().DropTable(&Channel{})
		_ = db.Migrator().DropTable(&User{})
	})

	require.NoError(t, db.AutoMigrate(&User{}, &Channel{}, &Token{}, &Log{}))
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	user := &User{
		Username:    "invite-legacy-" + suffix,
		Password:    "password123",
		DisplayName: "Invite legacy user",
	}
	require.NoError(t, db.Create(user).Error)
	require.NoError(t, db.Create(&Channel{
		Key:  "legacy-channel-" + suffix,
		Name: "legacy-channel-" + suffix,
	}).Error)
	require.NoError(t, db.Create(&Token{
		UserId: user.Id,
		Key:    "legacy-token-" + suffix,
		Name:   "legacy-token-" + suffix,
	}).Error)
	require.NoError(t, db.Create(&Log{
		UserId:  user.Id,
		Type:    LogTypeSystem,
		Content: "legacy log",
	}).Error)

	before := registrationInviteMigrationCountsFor(t, db)
	for range 2 {
		require.NoError(t, migrateRegistrationInvites(db))
	}
	assert.True(t, db.Migrator().HasTable(&RegistrationInvite{}))
	for _, indexName := range []string{
		registrationInviteCodeHashIndex,
		registrationInviteExpiresAtIndex,
		registrationInviteUsedAtIndex,
		registrationInviteRevokedAtIndex,
		registrationInviteCreatedByIndex,
		registrationInviteUsedByIndex,
	} {
		assert.True(t, db.Migrator().HasIndex(&RegistrationInvite{}, indexName), indexName)
	}

	rawCode := "inv_0123456789ABCDEFGHJK"
	codeHash := common.HmacSha256(rawCode, strings.Repeat("h", 32))
	invite := &RegistrationInvite{
		CodeHash:   codeHash,
		CodePrefix: rawCode[:RegistrationInviteCodePrefixLength],
		Note:       "migration test",
		CreatedBy:  user.Id,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	require.NoError(t, db.Create(invite).Error)

	duplicate := *invite
	duplicate.Id = 0
	require.Error(t, db.Create(&duplicate).Error, "code_hash must be unique")

	var stored RegistrationInvite
	require.NoError(t, db.Where("id = ?", invite.Id).First(&stored).Error)
	assert.Equal(t, codeHash, stored.CodeHash)
	assert.NotContains(t, stored.CodeHash, rawCode)

	serialized, err := common.Marshal(stored)
	require.NoError(t, err)
	assert.NotContains(t, string(serialized), stored.CodeHash)
	assert.NotContains(t, string(serialized), rawCode)
	assert.Equal(t, before, registrationInviteMigrationCountsFor(t, db))
}

func TestRegistrationInviteMigrationSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	testRegistrationInviteMigration(t, db)
}

func TestRegistrationInviteMigrationMySQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not configured")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testRegistrationInviteMigration(t, db)
}

func TestRegistrationInviteMigrationPostgreSQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not configured")
	}
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testRegistrationInviteMigration(t, db)
}
