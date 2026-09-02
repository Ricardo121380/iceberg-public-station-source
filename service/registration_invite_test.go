package service

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const registrationInviteTestSecret = "registration-invite-test-secret-0123456789"

func setupRegistrationInviteDomainTest(t *testing.T, db *gorm.DB, databaseType common.DatabaseType) {
	t.Helper()
	require.False(t, db.Migrator().HasTable(&model.RegistrationInvite{}), "registration invite domain tests require an empty database")

	previousDB := model.DB
	previousMainDatabaseType := common.MainDatabaseType()
	previousLogDatabaseType := common.LogDatabaseType()
	model.DB = db
	common.SetDatabaseTypes(databaseType, databaseType)
	require.NoError(t, db.AutoMigrate(&model.RegistrationInvite{}))

	t.Cleanup(func() {
		require.NoError(t, db.Migrator().DropTable(&model.RegistrationInvite{}))
		model.DB = previousDB
		common.SetDatabaseTypes(previousMainDatabaseType, previousLogDatabaseType)
	})
}

func openRegistrationInviteSQLiteTestDatabase(t *testing.T, gormLogger logger.Interface) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf(
		"file:%s?mode=memory&cache=shared&_pragma=busy_timeout(30000)&_txlock=immediate",
		strings.ReplaceAll(t.Name(), "/", "_"),
	)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gormLogger})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func clearRegistrationInvites(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.RegistrationInvite{}).Error)
}

func registrationInviteTestCode(character string) string {
	return model.RegistrationInviteCodePrefix + strings.Repeat(character, model.RegistrationInviteCodeRandomLength)
}

func createRegistrationInvite(t *testing.T, db *gorm.DB, rawCode string, expiresAt int64, mutate func(*model.RegistrationInvite)) *model.RegistrationInvite {
	t.Helper()
	codeHash, err := HashRegistrationInviteCode(rawCode, registrationInviteTestSecret)
	require.NoError(t, err)
	invite := &model.RegistrationInvite{
		CodeHash:   codeHash,
		CodePrefix: rawCode[:model.RegistrationInviteCodePrefixLength],
		Note:       "domain test",
		CreatedBy:  7,
		ExpiresAt:  expiresAt,
	}
	if mutate != nil {
		mutate(invite)
	}
	require.NoError(t, db.Create(invite).Error)
	return invite
}

func consumeRegistrationInvite(db *gorm.DB, inviteID int, userID int) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return ConsumeRegistrationInvite(tx, inviteID, userID)
	})
}

func testRegistrationInviteDomain(t *testing.T, db *gorm.DB, databaseType common.DatabaseType) {
	t.Helper()
	setupRegistrationInviteDomainTest(t, db, databaseType)

	t.Run("generates valid codes and returns only a one-time raw result", func(t *testing.T) {
		clearRegistrationInvites(t, db)
		generated, err := GenerateRegistrationInvites(3, 0, "first batch", 7, registrationInviteTestSecret)
		require.NoError(t, err)
		require.Len(t, generated, 3)

		seen := make(map[string]struct{}, len(generated))
		for _, invite := range generated {
			assert.NoError(t, model.ValidateRegistrationInviteCode(invite.Code))
			assert.Equal(t, invite.Code[:model.RegistrationInviteCodePrefixLength], invite.CodePrefix)
			assert.Greater(t, invite.ExpiresAt, common.GetTimestamp())
			_, exists := seen[invite.Code]
			assert.False(t, exists, "a generated batch must not contain duplicate raw codes")
			seen[invite.Code] = struct{}{}

			validated, validateErr := ValidateRegistrationInvite(invite.Code, registrationInviteTestSecret)
			require.NoError(t, validateErr)
			assert.Equal(t, invite.Id, validated.InviteID)
		}

		serialized, err := common.Marshal(generated[0])
		require.NoError(t, err)
		assert.NotContains(t, string(serialized), generated[0].Code)
	})

	t.Run("rejects invalid state and invalid HMAC configuration before consumption", func(t *testing.T) {
		clearRegistrationInvites(t, db)
		now := common.GetTimestamp()
		active := createRegistrationInvite(t, db, registrationInviteTestCode("A"), now+3600, nil)
		expired := createRegistrationInvite(t, db, registrationInviteTestCode("B"), now-1, nil)
		usedAt := now - 10
		usedBy := 11
		used := createRegistrationInvite(t, db, registrationInviteTestCode("C"), now+3600, func(invite *model.RegistrationInvite) {
			invite.UsedAt = &usedAt
			invite.UsedBy = &usedBy
		})
		revokedAt := now - 10
		revokedBy := 12
		revoked := createRegistrationInvite(t, db, registrationInviteTestCode("D"), now+3600, func(invite *model.RegistrationInvite) {
			invite.RevokedAt = &revokedAt
			invite.RevokedBy = &revokedBy
		})

		_, err := ValidateRegistrationInvite("bad", registrationInviteTestSecret)
		require.ErrorIs(t, err, model.ErrRegistrationInviteCodeInvalid)
		_, err = ValidateRegistrationInvite(registrationInviteTestCode("E"), registrationInviteTestSecret)
		require.ErrorIs(t, err, ErrRegistrationInviteNotFound)
		_, err = ValidateRegistrationInvite(registrationInviteTestCode("F"), "short")
		require.ErrorIs(t, err, ErrRegistrationInviteHMACSecretInvalid)
		_, err = ValidateRegistrationInvite(registrationInviteTestCode("B"), registrationInviteTestSecret)
		require.ErrorIs(t, err, ErrRegistrationInviteExpired)
		_, err = ValidateRegistrationInvite(registrationInviteTestCode("C"), registrationInviteTestSecret)
		require.ErrorIs(t, err, ErrRegistrationInviteAlreadyUsed)
		_, err = ValidateRegistrationInvite(registrationInviteTestCode("D"), registrationInviteTestSecret)
		require.ErrorIs(t, err, ErrRegistrationInviteRevoked)

		assert.ErrorIs(t, consumeRegistrationInvite(db, expired.Id, 21), ErrRegistrationInviteExpired)
		assert.ErrorIs(t, consumeRegistrationInvite(db, used.Id, 21), ErrRegistrationInviteAlreadyUsed)
		assert.ErrorIs(t, consumeRegistrationInvite(db, revoked.Id, 21), ErrRegistrationInviteRevoked)
		_, err = RevokeRegistrationInvite(used.Id, 22)
		assert.ErrorIs(t, err, ErrRegistrationInviteAlreadyUsed)

		var stored model.RegistrationInvite
		require.NoError(t, db.First(&stored, active.Id).Error)
		assert.Nil(t, stored.UsedAt, "a validation failure must not consume a different valid invite")
		assert.ErrorIs(t, RegistrationInvitePublicError(ErrRegistrationInviteNotFound), ErrRegistrationInviteUnavailable)
		assert.ErrorIs(t, RegistrationInvitePublicError(ErrRegistrationInviteExpired), ErrRegistrationInviteUnavailable)
	})

	t.Run("consumes exactly once with one hundred concurrent registration transactions", func(t *testing.T) {
		clearRegistrationInvites(t, db)
		invite := createRegistrationInvite(t, db, registrationInviteTestCode("F"), common.GetTimestamp()+3600, nil)

		const workers = 100
		start := make(chan struct{})
		results := make(chan error, workers)
		var waitGroup sync.WaitGroup
		waitGroup.Add(workers)
		for worker := 0; worker < workers; worker++ {
			go func(userID int) {
				defer waitGroup.Done()
				<-start
				results <- consumeRegistrationInvite(db, invite.Id, userID)
			}(worker + 100)
		}
		close(start)
		waitGroup.Wait()
		close(results)

		successes := 0
		for err := range results {
			if err == nil {
				successes++
				continue
			}
			assert.ErrorIs(t, err, ErrRegistrationInviteAlreadyUsed)
		}
		assert.Equal(t, 1, successes)

		var stored model.RegistrationInvite
		require.NoError(t, db.First(&stored, invite.Id).Error)
		assert.NotNil(t, stored.UsedAt)
		assert.NotNil(t, stored.UsedBy)
	})

	t.Run("revoke is idempotent and preserves the responsible operator", func(t *testing.T) {
		clearRegistrationInvites(t, db)
		invite := createRegistrationInvite(t, db, registrationInviteTestCode("G"), common.GetTimestamp()+3600, nil)

		revoked, err := RevokeRegistrationInvite(invite.Id, 88)
		require.NoError(t, err)
		assert.True(t, revoked)
		revoked, err = RevokeRegistrationInvite(invite.Id, 99)
		require.NoError(t, err)
		assert.False(t, revoked)

		var stored model.RegistrationInvite
		require.NoError(t, db.First(&stored, invite.Id).Error)
		require.NotNil(t, stored.RevokedAt)
		require.NotNil(t, stored.RevokedBy)
		assert.Equal(t, 88, *stored.RevokedBy)
		assert.ErrorIs(t, consumeRegistrationInvite(db, invite.Id, 21), ErrRegistrationInviteRevoked)
	})

	t.Run("retries known collisions and rolls back an incomplete batch", func(t *testing.T) {
		clearRegistrationInvites(t, db)
		now := common.GetTimestamp()
		collisionCode := registrationInviteTestCode("H")
		createRegistrationInvite(t, db, collisionCode, now+3600, nil)
		codeA := registrationInviteTestCode("J")
		codeB := registrationInviteTestCode("K")

		sequence := []string{collisionCode, codeA, codeB}
		position := 0
		generated, err := generateRegistrationInvites(db, 2, now+3600, "collision retry", 7, registrationInviteTestSecret, func() (string, error) {
			code := sequence[position]
			position++
			return code, nil
		})
		require.NoError(t, err)
		require.Len(t, generated, 2)
		assert.Equal(t, []string{codeA, codeB}, []string{generated[0].Code, generated[1].Code})

		clearRegistrationInvites(t, db)
		createRegistrationInvite(t, db, collisionCode, now+3600, nil)
		calls := 0
		_, err = generateRegistrationInvites(db, 2, now+3600, "must rollback", 7, registrationInviteTestSecret, func() (string, error) {
			calls++
			if calls == 1 {
				return codeA, nil
			}
			return collisionCode, nil
		})
		require.ErrorIs(t, err, ErrRegistrationInviteGenerationExhausted)

		var count int64
		require.NoError(t, db.Model(&model.RegistrationInvite{}).Count(&count).Error)
		assert.Equal(t, int64(1), count, "the first generated row must roll back with the failed batch")
	})

	t.Run("list returns safe derived data without code hash or raw code", func(t *testing.T) {
		clearRegistrationInvites(t, db)
		rawCode := registrationInviteTestCode("M")
		invite := createRegistrationInvite(t, db, rawCode, common.GetTimestamp()+3600, nil)

		listed, err := ListRegistrationInvites(RegistrationInviteListFilters{Status: RegistrationInviteStatusActive}, RegistrationInvitePagination{Limit: 10})
		require.NoError(t, err)
		require.Len(t, listed.Items, 1)
		assert.Equal(t, invite.Id, listed.Items[0].Id)
		assert.Equal(t, RegistrationInviteStatusActive, listed.Items[0].Status)

		serialized, err := common.Marshal(listed)
		require.NoError(t, err)
		assert.NotContains(t, string(serialized), rawCode)
		assert.NotContains(t, string(serialized), "code_hash")

		var stored model.RegistrationInvite
		require.NoError(t, db.First(&stored, invite.Id).Error)
		assert.NotEqual(t, rawCode, stored.CodeHash)
		assert.NotContains(t, stored.CodeHash, rawCode)
	})
}

func TestRegistrationInviteDomainSQLite(t *testing.T) {
	testRegistrationInviteDomain(t, openRegistrationInviteSQLiteTestDatabase(t, logger.Default.LogMode(logger.Silent)), common.DatabaseTypeSQLite)
}

func TestRegistrationInviteDomainMySQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not configured")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	testRegistrationInviteDomain(t, db, common.DatabaseTypeMySQL)
}

func TestRegistrationInviteDomainPostgreSQL(t *testing.T) {
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
	testRegistrationInviteDomain(t, db, common.DatabaseTypePostgreSQL)
}

func TestRegistrationInviteGenerationDoesNotLogRawCode(t *testing.T) {
	var output bytes.Buffer
	gormLogger := logger.New(log.New(&output, "", 0), logger.Config{LogLevel: logger.Info})
	db := openRegistrationInviteSQLiteTestDatabase(t, gormLogger)
	setupRegistrationInviteDomainTest(t, db, common.DatabaseTypeSQLite)

	generated, err := GenerateRegistrationInvites(1, common.GetTimestamp()+3600, "log capture", 7, registrationInviteTestSecret)
	require.NoError(t, err)
	require.Len(t, generated, 1)

	assert.NotContains(t, output.String(), generated[0].Code)
}
