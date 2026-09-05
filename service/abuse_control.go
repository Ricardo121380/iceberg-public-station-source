package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// Rules can only become verified through a reviewed source change accompanied
// by a real, sanitized upstream fixture. Settings cannot forge verification.
type SafetyRule struct {
	ID        string `json:"id"`
	Version   string `json:"version"`
	ChannelID int    `json:"channel_id"`
	Protocol  string `json:"protocol"`
	Field     string `json:"field"`
	Value     string `json:"value"`
	Category  string `json:"category"`
	Verified  bool   `json:"verified"`
	Evidence  string `json:"evidence"`
}

// Immutable in production; only fixture tests replace this package-private catalog.
var verifiedSafetyRules = []SafetyRule{}

func SafetyRules() []SafetyRule { return append([]SafetyRule{}, verifiedSafetyRules...) }

type safetyAttempt struct {
	ChannelID  int    `json:"channel_id"`
	Attempt    int    `json:"attempt"`
	Protocol   string `json:"protocol"`
	Field      string `json:"field"`
	Signal     string `json:"signal"`
	RuleID     string `json:"rule_id"`
	Version    string `json:"version"`
	Category   string `json:"category"`
	Actionable bool   `json:"actionable"`
}

type safetyObservation struct {
	mu       sync.Mutex
	policy   model.AbusePolicy
	round    int64
	attempt  int
	finished bool
	attempts []safetyAttempt
}

const safetyContextKey = "abuse_observation"

// BeginSafetyObservation runs after authentication, before channel selection.
// Its completion is deferred by the distributor to cover errors and 200/SSE.
func BeginSafetyObservation(c *gin.Context) (func(), error) {
	if _, exists := c.Get(safetyContextKey); exists {
		return func() {}, nil
	}
	p, err := model.GetAbusePolicy()
	if err != nil {
		return nil, err
	}
	s, err := model.GetAbuseState(c.GetInt("id"))
	if err != nil {
		return nil, err
	}
	if s.BlockedUntil > time.Now().Unix() {
		c.AbortWithStatusJSON(403, gin.H{"error": gin.H{"type": "access_denied", "code": "account_temporarily_suspended", "message": i18n.T(c, "abuse.suspended"), "blocked_until": s.BlockedUntil}})
		return func() {}, nil
	}
	obs := &safetyObservation{policy: p, round: s.Round}
	c.Set(safetyContextKey, obs)
	return func() { finishSafetyObservation(c, obs) }, nil
}

func NextSafetyAttempt(c *gin.Context) {
	if value, ok := c.Get(safetyContextKey); ok {
		obs := value.(*safetyObservation)
		obs.mu.Lock()
		obs.attempt++
		obs.mu.Unlock()
	}
}

// ObserveSafetyPayload reads only structural fields. It never retains the raw
// response or an upstream message, which may echo prompts or credentials.
func ObserveSafetyPayload(ctx context.Context, payload []byte, protocol string) {
	c, ok := ctx.(*gin.Context)
	if !ok {
		return
	}
	value, ok := c.Get(safetyContextKey)
	if !ok {
		return
	}
	obs := value.(*safetyObservation)
	if obs.policy.Mode == "off" || c.GetInt("channel_type") != constant.ChannelTypeNewAPI {
		return
	}
	if !gjson.ValidBytes(payload) {
		return
	}
	root := gjson.ParseBytes(payload)
	var signals []struct{ field, value string }
	for _, field := range []string{"error.code", "error.type", "response.error.code", "response.error.type"} {
		v := root.Get(field).String()
		switch v {
		case "content_filter", "content_policy_violation", "safety_violation", "prompt_blocked":
			signals = append(signals, struct{ field, value string }{field, v})
		}
	}
	for _, choice := range root.Get("choices").Array() {
		if choice.Get("finish_reason").String() == "content_filter" {
			signals = append(signals, struct{ field, value string }{"choices.finish_reason", "content_filter"})
		}
		for _, f := range []string{"message.refusal", "delta.refusal"} {
			if choice.Get(f).String() != "" {
				signals = append(signals, struct{ field, value string }{"choices.refusal", "refusal"})
			}
		}
	}
	for _, field := range []string{"incomplete_details.reason", "response.incomplete_details.reason"} {
		if root.Get(field).String() == "content_filter" {
			signals = append(signals, struct{ field, value string }{field, "content_filter"})
		}
	}
	if strings.HasPrefix(root.Get("type").String(), "response.refusal.") {
		signals = append(signals, struct{ field, value string }{"type", "refusal"})
	}
	for _, prefix := range []string{"", "response."} {
		for _, item := range root.Get(prefix + "output").Array() {
			for _, part := range item.Get("content").Array() {
				if part.Get("type").String() == "refusal" {
					signals = append(signals, struct{ field, value string }{"output.content.type", "refusal"})
				}
			}
		}
	}
	if root.Get("stop_reason").String() == "refusal" || root.Get("delta.stop_reason").String() == "refusal" {
		signals = append(signals, struct{ field, value string }{"stop_reason", "refusal"})
	}
	// Protocols with explicit categories are only actionable once a reviewed,
	// channel-bound exact rule has a real evidence fixture.
	var enabled []string
	_ = common.UnmarshalJsonStr(obs.policy.EnabledRules, &enabled)
	for _, rule := range SafetyRules() {
		if !rule.Verified || rule.Evidence == "" || rule.ChannelID != c.GetInt("channel_id") || rule.Protocol != protocol {
			continue
		}
		if root.Get(rule.Field).String() != rule.Value {
			continue
		}
		for _, id := range enabled {
			if id == rule.ID {
				signals = append(signals, struct{ field, value string }{rule.Field, rule.Value})
			}
		}
	}
	obs.mu.Lock()
	defer obs.mu.Unlock()
	if obs.finished {
		return
	}
	for _, s := range signals {
		a := safetyAttempt{ChannelID: c.GetInt("channel_id"), Attempt: obs.attempt, Protocol: protocol, Field: s.field, Signal: s.value, RuleID: "upstream_unclassified", Version: "1", Category: "unclassified"}
		for _, rule := range SafetyRules() {
			for _, id := range enabled {
				if rule.ID == id && rule.Verified && rule.Evidence != "" && rule.ChannelID == a.ChannelID && rule.Protocol == protocol && rule.Field == s.field && rule.Value == s.value {
					a.RuleID = rule.ID
					a.Version = rule.Version
					a.Category = rule.Category
					a.Actionable = true
				}
			}
		}
		duplicate := false
		for _, prior := range obs.attempts {
			if prior == a {
				duplicate = true
				break
			}
		}
		// Bound request-local memory even with malicious/repeating stream events.
		if !duplicate && len(obs.attempts) < 32 {
			obs.attempts = append(obs.attempts, a)
		}
	}
}

func SafetyInputBlocked(c *gin.Context) bool {
	v, ok := c.Get(safetyContextKey)
	if !ok {
		return false
	}
	o := v.(*safetyObservation)
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, a := range o.attempts {
		if a.Attempt == o.attempt && a.Actionable {
			return true
		}
	}
	return false
}

func finishSafetyObservation(c *gin.Context, o *safetyObservation) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.finished = true
	if len(o.attempts) == 0 {
		return
	}
	selected := o.attempts[len(o.attempts)-1]
	for _, a := range o.attempts {
		if a.Actionable {
			selected = a
			break
		}
	}
	requestID := c.GetString(common.RequestIdKey)
	if requestID == "" {
		requestID = common.NewRequestId()
	}
	summary := "Upstream safety signal; category information is insufficient."
	if selected.Actionable {
		summary = "Verified upstream input policy block."
	}
	attempts, err := common.Marshal(o.attempts)
	if err != nil {
		common.SysError("abuse_control: serialize_failed")
		return
	}
	e := model.AbuseEvent{UserID: c.GetInt("id"), TokenID: c.GetInt("token_id"), RequestID: requestID, ChannelID: selected.ChannelID, Model: c.GetString("original_model"), Protocol: selected.Protocol, RuleID: selected.RuleID, RuleVersion: selected.Version, Category: selected.Category, Signal: selected.Signal, Summary: summary, Attempts: string(attempts), Actionable: selected.Actionable, Generation: o.policy.Generation, Round: o.round, CreatedAt: time.Now().Unix()}
	// A database failure may be transient or an uncertain commit. Retrying the
	// immutable event is safe because user/request identity is unique.
	for attempt := 0; attempt < 2; attempt++ {
		event := e
		if err = model.RecordAbuseEvent(&event); err == nil {
			return
		}
	}
	common.SysError(fmt.Sprintf("abuse_control: persistence_failed user_id=%d request_id=%s", e.UserID, requestID))
}

// StartAbuseMaintenance owns retention; it never changes suspension state.
func StartAbuseMaintenance() {
	go func() {
		for {
			if err := model.CleanupAbuseEvents(time.Now().Unix()); err != nil {
				common.SysError("abuse_control: cleanup_failed")
			}
			time.Sleep(24 * time.Hour)
		}
	}()
}
