package model

import (
	"os"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUsageQualityLogSemantics(t *testing.T) {
	tests := []struct {
		name               string
		prompt             int
		stream             bool
		other              string
		ttft               float64
		input, read, count int64
	}{
		{"openai inclusive input", 100, true, `{"frt":1000,"cache_tokens":80}`, 1000, 100, 80, 1},
		{"anthropic exclusive input", 20, true, `{"frt":2000,"claude":true,"cache_tokens":60,"cache_creation_tokens":20}`, 2000, 100, 60, 1},
		{"split writes replace total", 20, true, `{"frt":1500,"usage_semantic":"anthropic","cache_tokens":50,"cache_creation_tokens":30,"cache_creation_tokens_5m":10,"cache_creation_tokens_1h":20}`, 1500, 100, 50, 1},
		{"normalized input wins", 100, true, `{"frt":900,"claude":true,"input_tokens_total":100,"cache_tokens":80,"cache_write_tokens":20}`, 900, 100, 80, 1},
		{"normalized writes win", 20, false, `{"claude":true,"cache_tokens":50,"cache_creation_tokens":999,"cache_write_tokens":30}`, 0, 100, 50, 1},
		{"zero hit is real", 100, false, `{"frt":1000,"cache_tokens":0}`, 0, 100, 0, 1},
		{"missing cache is unknown", 100, true, `{"frt":1000}`, 1000, 0, 0, 0},
		{"invalid ttft excluded", 100, true, `{"frt":-1,"cache_tokens":80}`, 0, 100, 80, 1},
		{"stream error excluded from ttft", 100, true, `{"frt":1000,"stream":{"status":"error"},"cache_tokens":80}`, 0, 100, 80, 1},
		{"local counts excluded from cache", 100, true, `{"frt":1000,"cache_tokens":0,"admin_info":{"local_count_tokens":true}}`, 1000, 0, 0, 0},
		{"read greater than input excluded", 100, false, `{"cache_tokens":101}`, 0, 0, 0, 0},
		{"negative writes excluded", 20, false, `{"claude":true,"cache_tokens":50,"cache_write_tokens":-1}`, 0, 0, 0, 0},
		{"zero input excluded", 0, false, `{"cache_tokens":0}`, 0, 0, 0, 0},
		{"bad json excluded", 100, true, `{broken`, 0, 0, 0, 0},
		{"no metadata excluded", 100, true, `{}`, 0, 0, 0, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var bucket UsageQualityBucket
			bucket.addLog(&Log{PromptTokens: test.prompt, IsStream: test.stream, Other: test.other})
			assert.Equal(t, test.ttft, bucket.TtftSumMs)
			expectedTtftCount := int64(0)
			if test.ttft > 0 {
				expectedTtftCount = 1
			}
			assert.Equal(t, expectedTtftCount, bucket.TtftCount)
			assert.Equal(t, test.input, bucket.InputTokens)
			assert.Equal(t, test.read, bucket.CacheReadTokens)
			assert.Equal(t, test.count, bucket.CacheCount)
		})
	}
}

func TestUsageQualityDatabase(t *testing.T) {
	tests := []struct {
		name, env string
		open      func(string) gorm.Dialector
	}{
		{"sqlite", "", func(string) gorm.Dialector { return sqlite.Open(":memory:") }},
		{"mysql", "TEST_MYSQL_DSN", func(dsn string) gorm.Dialector { return mysql.Open(dsn) }},
		{"postgres", "TEST_POSTGRES_DSN", func(dsn string) gorm.Dialector {
			return postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dsn := os.Getenv(test.env)
			if test.env != "" && dsn == "" {
				t.Skip("isolated test database not configured")
			}
			db, err := gorm.Open(test.open(dsn), &gorm.Config{})
			require.NoError(t, err)
			require.False(t, db.Migrator().HasTable(&Log{}), "use an empty isolated database")
			require.False(t, db.Migrator().HasTable(&User{}), "refusing database with users")
			sqlDB, err := db.DB()
			require.NoError(t, err)
			sqlDB.SetMaxOpenConns(1)
			previous := LOG_DB
			LOG_DB = db
			t.Cleanup(func() { LOG_DB = previous; _ = db.Migrator().DropTable(&Log{}); _ = sqlDB.Close() })
			require.NoError(t, db.AutoMigrate(&Log{}))
			var version string
			versionQuery := "SELECT version()"
			if test.name == "sqlite" {
				versionQuery = "SELECT sqlite_version()"
			}
			require.NoError(t, db.Raw(versionQuery).Scan(&version).Error)
			t.Logf("database version: %s", version)
			logs := []Log{
				{UserId: 1, Username: "one", ModelName: "a", CreatedAt: 3601, Type: LogTypeConsume, PromptTokens: 100, IsStream: true, Other: `{"frt":1000,"cache_tokens":80}`},
				{UserId: 2, Username: "two", ModelName: "a", CreatedAt: 3602, Type: LogTypeConsume, PromptTokens: 900, IsStream: true, Other: `{"frt":3000,"cache_tokens":90}`},
				{UserId: 1, Username: "one", ModelName: "b", CreatedAt: 7200, Type: LogTypeConsume, PromptTokens: 100, IsStream: false, Other: `{"cache_tokens":0}`},
				{UserId: 1, Username: "one", ModelName: "ignored", CreatedAt: 3603, Type: LogTypeError, PromptTokens: 100, Other: `{"cache_tokens":90}`},
				{UserId: 1, Username: "one", ModelName: "outside", CreatedAt: 7201, Type: LogTypeConsume, PromptTokens: 100, Other: `{"cache_tokens":90}`},
			}
			require.NoError(t, db.Create(&logs).Error)
			all, err := GetUsageQuality(t.Context(), 3601, 7200, "", 0)
			require.NoError(t, err)
			require.Len(t, all, 2)
			assert.Equal(t, UsageQualityBucket{ModelName: "a", CreatedAt: 3600, TtftSumMs: 4000, TtftCount: 2, InputTokens: 1000, CacheReadTokens: 170, CacheCount: 2}, all[0])
			assert.Equal(t, "b", all[1].ModelName)
			own, err := GetUsageQuality(t.Context(), 3601, 7200, "two", 1)
			require.NoError(t, err)
			require.Len(t, own, 2)
			assert.EqualValues(t, 100, own[0].InputTokens)
			named, err := GetUsageQuality(t.Context(), 3601, 7200, "two", 0)
			require.NoError(t, err)
			require.Len(t, named, 1)
			assert.EqualValues(t, 900, named[0].InputTokens)
			empty, err := GetUsageQuality(t.Context(), 1, 2, "", 0)
			require.NoError(t, err)
			assert.Empty(t, empty)
			payload, err := common.Marshal(all)
			require.NoError(t, err)
			assert.NotContains(t, string(payload), "username")
			assert.NotContains(t, string(payload), "other")
			require.NoError(t, db.Migrator().DropTable(&Log{}))
			_, err = GetUsageQuality(t.Context(), 3601, 7200, "", 0)
			require.Error(t, err)
		})
	}
}
