package model

import (
	"context"
	"math"
	"sort"

	"github.com/QuantumNous/new-api/common"
)

// UsageQualityBucket contains additive counters, never averages of averages.
// It is a response type only; existing retained consume logs are the source.
type UsageQualityBucket struct {
	ModelName       string  `json:"model_name"`
	CreatedAt       int64   `json:"created_at"`
	TtftSumMs       float64 `json:"ttft_sum_ms"`
	TtftCount       int64   `json:"ttft_count"`
	CacheReadTokens int64   `json:"cache_read_tokens"`
	InputTokens     int64   `json:"input_tokens"`
	CacheCount      int64   `json:"cache_count"`
}

func (bucket *UsageQualityBucket) addLog(log *Log) {
	var other struct {
		Frt                 *float64 `json:"frt"`
		CacheTokens         *int64   `json:"cache_tokens"`
		InputTokensTotal    *int64   `json:"input_tokens_total"`
		CacheWriteTokens    *int64   `json:"cache_write_tokens"`
		CacheCreationTokens int64    `json:"cache_creation_tokens"`
		CacheCreation5m     int64    `json:"cache_creation_tokens_5m"`
		CacheCreation1h     int64    `json:"cache_creation_tokens_1h"`
		Claude              bool     `json:"claude"`
		UsageSemantic       string   `json:"usage_semantic"`
		AdminInfo           struct {
			LocalCountTokens bool `json:"local_count_tokens"`
		} `json:"admin_info"`
		Stream struct {
			Status string `json:"status"`
		} `json:"stream"`
	}
	if common.UnmarshalJsonStr(log.Other, &other) != nil {
		return
	}
	if log.IsStream && other.Stream.Status != "error" && other.Frt != nil && *other.Frt > 0 && !math.IsNaN(*other.Frt) && !math.IsInf(*other.Frt, 0) {
		bucket.TtftSumMs += *other.Frt
		bucket.TtftCount++
	}
	if other.CacheTokens == nil || other.AdminInfo.LocalCountTokens {
		return
	}
	// Bound per-log arithmetic before adding provider-controlled token counts.
	const maxTokens = int64(1<<53 - 1)
	read := *other.CacheTokens
	input := int64(log.PromptTokens)
	if read < 0 || read > maxTokens || input < 0 || input > maxTokens {
		return
	}
	if other.InputTokensTotal != nil {
		input = *other.InputTokensTotal
	} else if other.Claude || other.UsageSemantic == "anthropic" {
		write := other.CacheCreationTokens
		if other.CacheCreation5m < 0 || other.CacheCreation1h < 0 || other.CacheCreation5m > maxTokens || other.CacheCreation1h > maxTokens {
			return
		}
		if other.CacheCreation5m > 0 || other.CacheCreation1h > 0 {
			write = other.CacheCreation5m + other.CacheCreation1h
		}
		if other.CacheWriteTokens != nil {
			write = *other.CacheWriteTokens
		}
		if write < 0 || write > maxTokens {
			return
		}
		input += read + write
	}
	if input <= 0 || input > maxTokens || read > input {
		return
	}
	bucket.CacheReadTokens += read
	bucket.InputTokens += input
	bucket.CacheCount++
}

// GetUsageQuality reads only the necessary fields from the configured log DB.
// Rows are streamed to avoid loading request metadata for an entire period into memory.
func GetUsageQuality(ctx context.Context, start, end int64, username string, userID int) ([]UsageQualityBucket, error) {
	query := LOG_DB.WithContext(ctx).Model(&Log{}).
		Select("model_name, created_at, prompt_tokens, is_stream, other").
		Where("type = ? AND created_at >= ? AND created_at <= ?", LogTypeConsume, start, end)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	} else if username != "" {
		query = query.Where("username = ?", username)
	}
	rows, err := query.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type bucketKey struct {
		model string
		hour  int64
	}
	buckets := make(map[bucketKey]*UsageQualityBucket)
	for rows.Next() {
		var log Log
		if err := query.ScanRows(rows, &log); err != nil {
			return nil, err
		}
		key := bucketKey{log.ModelName, log.CreatedAt - log.CreatedAt%3600}
		bucket := buckets[key]
		if bucket == nil {
			bucket = &UsageQualityBucket{ModelName: key.model, CreatedAt: key.hour}
			buckets[key] = bucket
		}
		bucket.addLog(&log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make([]UsageQualityBucket, 0, len(buckets))
	for _, bucket := range buckets {
		if bucket.TtftCount > 0 || bucket.CacheCount > 0 {
			result = append(result, *bucket)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt == result[j].CreatedAt {
			return result[i].ModelName < result[j].ModelName
		}
		return result[i].CreatedAt < result[j].CreatedAt
	})
	return result, nil
}
