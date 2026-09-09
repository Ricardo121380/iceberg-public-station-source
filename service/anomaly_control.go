package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/go-redis/redis/v8"
)

const anomalyPrefix = "anomaly:v1:"
const anomalyRetention = 7 * 24 * time.Hour

var ErrAnomalyUnavailable = errors.New("anomaly storage unavailable")
var ErrAnomalyConflict = errors.New("anomaly changed; refresh before saving")

type AnomalyPolicy struct {
	Enabled           bool  `json:"enabled"`
	WindowMinutes     int   `json:"window_minutes"`
	SchemaThreshold   int   `json:"schema_threshold"`
	RateThreshold     int   `json:"rate_threshold"`
	UpstreamThreshold int   `json:"upstream_threshold"`
	Version           int64 `json:"version"`
}

type AnomalyEvent struct {
	ID                string `json:"id"`
	UserID            int    `json:"user_id"`
	ChannelID         int    `json:"channel_id"`
	Model             string `json:"model"`
	Kind              string `json:"kind"`
	Count             int    `json:"count"`
	Threshold         int    `json:"threshold"`
	FirstSeen         int64  `json:"first_seen"`
	LastSeen          int64  `json:"last_seen"`
	Alert             bool   `json:"alert"`
	AcknowledgedBy    int    `json:"acknowledged_by"`
	AcknowledgedAt    int64  `json:"acknowledged_at"`
	AcknowledgedCount int    `json:"acknowledged_count"`
}

type AnomalyAudit struct {
	TelegramOperatorID int64          `json:"telegram_operator_id,omitempty"`
	UserID             int            `json:"user_id,omitempty"`
	ID                 string         `json:"id"`
	OperatorID         int            `json:"operator_id"`
	Action             string         `json:"action"`
	EventID            string         `json:"event_id,omitempty"`
	CreatedAt          int64          `json:"created_at"`
	Policy             *AnomalyPolicy `json:"policy,omitempty"`
}

func GetAnomalyPolicy(ctx context.Context) (AnomalyPolicy, error) {
	p := AnomalyPolicy{Enabled: true, WindowMinutes: 5, SchemaThreshold: 10, RateThreshold: 100, UpstreamThreshold: 5}
	if !common.RedisEnabled || common.RDB == nil {
		return p, ErrAnomalyUnavailable
	}
	raw, err := common.RDB.Get(ctx, anomalyPrefix+"policy").Result()
	if err == redis.Nil {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	err = common.UnmarshalJsonStr(raw, &p)
	return p, err
}

func ValidateAnomalyPolicy(p AnomalyPolicy) bool {
	return p.WindowMinutes >= 1 && p.WindowMinutes <= 60 && p.SchemaThreshold >= 2 && p.SchemaThreshold <= 1000 && p.RateThreshold >= 2 && p.RateThreshold <= 10000 && p.UpstreamThreshold >= 2 && p.UpstreamThreshold <= 1000 && p.Version >= 0
}

func SaveAnomalyPolicy(ctx context.Context, p AnomalyPolicy, operator int) error {
	if !ValidateAnomalyPolicy(p) {
		return errors.New("invalid anomaly policy")
	}
	if !common.RedisEnabled || common.RDB == nil {
		return ErrAnomalyUnavailable
	}
	expected := p.Version
	p.Version++
	raw, err := common.Marshal(p)
	if err != nil {
		return err
	}
	audit, _ := common.Marshal(AnomalyAudit{ID: common.GetUUID(), OperatorID: operator, Action: "settings_updated", CreatedAt: time.Now().Unix(), Policy: &p})
	result, err := common.RDB.Eval(ctx, `
 local current=redis.call('GET',KEYS[1])
 local version=0
 if current then version=cjson.decode(current).version end
 if version~=tonumber(ARGV[1]) then return 0 end
 redis.call('SET',KEYS[1],ARGV[2])
 redis.call('LPUSH',KEYS[2],ARGV[3]); redis.call('LTRIM',KEYS[2],0,199); redis.call('EXPIRE',KEYS[2],604800)
 return 1`, []string{anomalyPrefix + "policy", anomalyPrefix + "audit"}, expected, string(raw), string(audit)).Int()
	if err != nil {
		return err
	}
	if result == 0 {
		return ErrAnomalyConflict
	}
	return nil
}

// Only static categories leave this classifier. Raw provider messages, headers,
// prompts, tokens and arbitrary error text are never stored in this subsystem.
func ClassifyAnomaly(status int, message string, upstream bool) string {
	if status == 400 && strings.Contains(message, "Invalid schema for response_format") {
		return "invalid_schema"
	}
	if upstream && (status == 429 || status >= 500) {
		return "upstream"
	}
	if status == 429 {
		return "rate_limit"
	}
	if status == 403 && (strings.Contains(message, "额度") || strings.Contains(message, "quota")) {
		return "quota"
	}
	if status == 503 && strings.Contains(message, "No available channel") {
		return "unsupported_model"
	}
	if status == 400 {
		return "invalid_request"
	}
	return ""
}

// Fixed, aligned windows prevent unbounded per-request storage. A request is
// recorded once, after its final status; successful retries do not create alerts.
func RecordAnomaly(ctx context.Context, userID, channelID int, model, kind string) error {
	p, err := GetAnomalyPolicy(ctx)
	if err != nil {
		return err
	}
	if !p.Enabled || userID <= 0 {
		return nil
	}
	threshold := 0
	switch kind {
	case "invalid_schema":
		threshold = p.SchemaThreshold
	case "rate_limit":
		threshold = p.RateThreshold
	case "upstream":
		threshold = p.UpstreamThreshold
	case "quota", "unsupported_model", "invalid_request":
	default:
		return nil
	}
	// Model labels are untrusted input; bound them and exclude control characters.
	if len(model) > 128 {
		model = model[:128]
	}
	model = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, model)
	now := time.Now().Unix()
	window := now / int64(p.WindowMinutes*60)
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%d|%s|%s|%d|%d", userID, channelID, model, kind, window, p.Version)))
	id := fmt.Sprintf("%x", sum[:16])
	e := AnomalyEvent{ID: id, UserID: userID, ChannelID: channelID, Model: model, Kind: kind, Threshold: threshold, FirstSeen: now, LastSeen: now}
	raw, err := common.Marshal(e)
	if err != nil {
		return err
	}
	return common.RDB.Eval(ctx, `
 local raw=redis.call('GET',KEYS[1])
 local event=cjson.decode(ARGV[1])
 if raw then event=cjson.decode(raw) end
 event.count=event.count+1; event.last_seen=tonumber(ARGV[2])
 event.alert=event.threshold>0 and event.count>=event.threshold
 redis.call('SET',KEYS[1],cjson.encode(event),'EX',604800)
 redis.call('ZADD',KEYS[2],ARGV[2],event.id)
 redis.call('ZREMRANGEBYSCORE',KEYS[2],'-inf',tonumber(ARGV[2])-604800)
 local excess=redis.call('ZCARD',KEYS[2])-2000
 if excess>0 then
  local removed=redis.call('ZRANGE',KEYS[2],0,excess-1)
  for _,id in ipairs(removed) do redis.call('DEL',ARGV[3]..id) end
  redis.call('ZREMRANGEBYRANK',KEYS[2],0,excess-1)
 end
 redis.call('EXPIRE',KEYS[2],604800)
 return 1`, []string{anomalyPrefix + "event:" + id, anomalyPrefix + "events"}, string(raw), now, anomalyPrefix+"event:").Err()
}

func ListAnomalies(ctx context.Context) ([]AnomalyEvent, error) {
	if !common.RedisEnabled || common.RDB == nil {
		return nil, ErrAnomalyUnavailable
	}
	now := time.Now().Unix()
	ids, err := common.RDB.ZRevRangeByScore(ctx, anomalyPrefix+"events", &redis.ZRangeBy{Min: strconv.FormatInt(now-int64(anomalyRetention/time.Second), 10), Max: "+inf", Offset: 0, Count: 2000}).Result()
	if err != nil {
		return nil, err
	}
	events := make([]AnomalyEvent, 0, len(ids))
	if len(ids) == 0 {
		return events, nil
	}
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = anomalyPrefix + "event:" + id
	}
	vals, err := common.RDB.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	for _, v := range vals {
		if v == nil {
			continue
		}
		var e AnomalyEvent
		if err := common.UnmarshalJsonStr(v.(string), &e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func AcknowledgeAnomaly(ctx context.Context, id string, count, operator int) error {
	if !common.RedisEnabled || common.RDB == nil {
		return ErrAnomalyUnavailable
	}
	now := time.Now().Unix()
	audit, _ := common.Marshal(AnomalyAudit{ID: common.GetUUID(), OperatorID: operator, Action: "acknowledged", EventID: id, CreatedAt: now})
	n, err := common.RDB.Eval(ctx, `
 local raw=redis.call('GET',KEYS[1]); if not raw then return 0 end
 local e=cjson.decode(raw)
 if e.count~=tonumber(ARGV[1]) or e.acknowledged_count>=e.count then return 0 end
 local ttl=redis.call('TTL',KEYS[1]); if ttl<=0 then return 0 end
 e.acknowledged_by=tonumber(ARGV[2]); e.acknowledged_at=tonumber(ARGV[3]); e.acknowledged_count=e.count
 redis.call('SET',KEYS[1],cjson.encode(e),'EX',ttl)
 redis.call('LPUSH',KEYS[2],ARGV[4]); redis.call('LTRIM',KEYS[2],0,199); redis.call('EXPIRE',KEYS[2],604800)
 return 1`, []string{anomalyPrefix + "event:" + id, anomalyPrefix + "audit"}, count, operator, now, string(audit)).Int()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrAnomalyConflict
	}
	return nil
}

func ListAnomalyAudit(ctx context.Context) ([]AnomalyAudit, error) {
	if !common.RedisEnabled || common.RDB == nil {
		return nil, ErrAnomalyUnavailable
	}
	rows, err := common.RDB.LRange(ctx, anomalyPrefix+"audit", 0, 199).Result()
	if err != nil {
		return nil, err
	}
	result := make([]AnomalyAudit, 0, len(rows))
	cutoff := time.Now().Add(-anomalyRetention).Unix()
	for _, raw := range rows {
		var a AnomalyAudit
		if err := common.UnmarshalJsonStr(raw, &a); err != nil {
			return nil, err
		}
		if a.CreatedAt >= cutoff {
			result = append(result, a)
		}
	}
	return result, nil
}
