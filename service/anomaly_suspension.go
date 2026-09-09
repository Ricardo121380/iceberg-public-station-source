package service

import (
	"context"
	"errors"
	"github.com/QuantumNous/new-api/common"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

// TG pauses are separate from evidence-based safety state and never change login.
func AnomalySuspendedUntil(ctx context.Context, user int) (int64, error) {
	if !common.RedisEnabled || common.RDB == nil {
		return 0, nil
	}
	raw, err := common.RDB.Get(ctx, anomalyPrefix+"pause:"+strconv.Itoa(user)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	var p struct {
		Until int64 `json:"until"`
	}
	err = common.UnmarshalJsonStr(raw, &p)
	return p.Until, err
}

// op is a server-created random confirmation ID, reused on transport retries.
func ApplyAnomalyAction(ctx context.Context, eventID, op, action string, user int, operator int64) (int64, error) {
	if !common.RedisEnabled || common.RDB == nil {
		return 0, ErrAnomalyUnavailable
	}
	now := time.Now().Unix()
	until := now + 1800
	audit, _ := common.Marshal(map[string]any{"id": op, "operator_id": 0, "telegram_operator_id": operator, "action": "tg_" + action, "event_id": eventID, "user_id": user, "created_at": now})
	pause, _ := common.Marshal(map[string]any{"until": until, "event_id": eventID, "operation_id": op})
	n, err := common.RDB.Eval(ctx, `
 local previous=redis.call('GET',KEYS[3]);if previous then return tonumber(previous) end
 local raw=redis.call('GET',KEYS[1]);if not raw then return -1 end
 local e=cjson.decode(raw)
 if e.user_id~=tonumber(ARGV[1]) then return -1 end
 local now=tonumber(ARGV[2])
 local result=0
 if ARGV[3]=='pause' then
  if not e.alert or (e.kind~='invalid_schema' and e.kind~='rate_limit') or now-e.last_seen>1800 then return -1 end
  if redis.call('EXISTS',KEYS[2])==1 then return -2 end
  if redis.call('EXISTS',KEYS[5])==1 then return -1 end
  result=tonumber(ARGV[4]);redis.call('SET',KEYS[2],ARGV[5],'EX',1800)
  redis.call('SET',KEYS[5],'1','EX',604800)
 elseif ARGV[3]=='release' then
  local p=redis.call('GET',KEYS[2]);if not p then return -1 end
  if cjson.decode(p).event_id~=e.id then return -1 end
  redis.call('DEL',KEYS[2])
 else return -1 end
 redis.call('SET',KEYS[3],result,'EX',604800)
 redis.call('LPUSH',KEYS[4],ARGV[6]);redis.call('LTRIM',KEYS[4],0,199);redis.call('EXPIRE',KEYS[4],604800)
 return result`, []string{anomalyPrefix + "event:" + eventID, anomalyPrefix + "pause:" + strconv.Itoa(user), anomalyPrefix + "operation:" + op, anomalyPrefix + "audit", anomalyPrefix + "acted:" + eventID}, user, now, action, until, string(pause), string(audit)).Int64()
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return 0, errors.New("event expired, already handled, or user already paused")
	}
	return n, nil
}
