package service

import (
	"fmt"
	"io"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type abuseExcerpt struct {
	Text         string `json:"text"`
	Source       string `json:"source"`
	Truncated    bool   `json:"truncated"`
	OmittedParts bool   `json:"omitted_parts"`
}

// Only text of the last user message is eligible. No fallback to system,
// assistant, tool messages, attachments, or previous user turns.
func extractAbuseExcerpt(payload []byte, path string) (abuseExcerpt, string) {
	if !gjson.ValidBytes(payload) {
		return abuseExcerpt{}, "invalid_body"
	}
	root := gjson.ParseBytes(payload)
	var content gjson.Result
	var source string
	if path == "/v1/responses" {
		input := root.Get("input")
		if input.Type == gjson.String {
			content = input
			source = "input"
		} else if input.IsArray() {
			items := input.Array()
			for i := len(items) - 1; i >= 0; i-- {
				item := items[i]
				if item.Get("role").String() == "user" && (item.Get("type").String() == "" || item.Get("type").String() == "message") {
					content = item.Get("content")
					source = "input.user.content"
					break
				}
			}
		}
	} else if path == "/v1/chat/completions" || path == "/pg/chat/completions" || path == "/v1/messages" {
		messages := root.Get("messages").Array()
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Get("role").String() == "user" {
				content = messages[i].Get("content")
				source = "messages.user.content"
				break
			}
		}
	} else {
		return abuseExcerpt{}, "unsupported_protocol"
	}
	excerpt := abuseExcerpt{Source: source}
	var text strings.Builder
	if content.Type == gjson.String {
		text.WriteString(content.String())
	} else if content.IsArray() {
		for _, part := range content.Array() {
			kind := part.Get("type").String()
			if (kind == "text" || kind == "input_text") && part.Get("text").Type == gjson.String {
				text.WriteString(part.Get("text").String())
				text.WriteByte('\n')
			} else {
				excerpt.OmittedParts = true
			}
		}
	}
	excerpt.Text, excerpt.Truncated = common.RedactAbuseText(text.String())
	if excerpt.Text == "" {
		return excerpt, "no_user_text"
	}
	return excerpt, "available"
}

// Called only after observing a safety signal, while the original reusable body
// remains alive. An independent reader leaves relay cursor/response semantics intact.
func captureAbuseEvidence(c *gin.Context, e *model.AbuseEvent) {
	e.EvidenceStatus = "unavailable"
	if _, err := common.AbuseEvidenceCipher(); err != nil {
		e.EvidenceStatus = "key_unavailable"
		return
	}
	cached, ok := c.Get(common.KeyBodyStorage)
	storage, okStorage := cached.(common.BodyStorage)
	if !ok || !okStorage || storage == nil {
		e.EvidenceStatus = "body_unavailable"
		return
	}
	if storage.Size() > 1<<20 {
		e.EvidenceStatus = "body_too_large"
		return
	}
	reader, err := storage.NewReader()
	if err != nil {
		e.EvidenceStatus = "body_unavailable"
		return
	}
	defer reader.Close()
	body, err := io.ReadAll(io.LimitReader(reader, (1<<20)+1))
	if err != nil || len(body) > 1<<20 {
		e.EvidenceStatus = "body_unavailable"
		return
	}
	excerpt, status := extractAbuseExcerpt(body, c.Request.URL.Path)
	e.EvidenceStatus = status
	if status != "available" {
		return
	}
	data, err := common.Marshal(excerpt)
	if err != nil {
		e.EvidenceStatus = "capture_failed"
		return
	}
	identity := fmt.Sprintf("%d:%s", e.UserID, e.RequestID)
	ciphertext, err := common.SealAbuseEvidence(data, identity)
	if err != nil {
		e.EvidenceStatus = "capture_failed"
		return
	}
	e.EvidenceExpiresAt = e.CreatedAt + 7*86400
	e.Evidence = &model.AbuseEvidence{Ciphertext: ciphertext, ExpiresAt: e.EvidenceExpiresAt}
}

func ReadAbuseExcerpt(id int64, operator int, now int64) (abuseExcerpt, string, int64, error) {
	// The audit transaction must commit before plaintext can leave this service.
	e, encrypted, err := model.AccessAbuseEvidence(id, operator, now)
	if err != nil {
		return abuseExcerpt{}, "", 0, err
	}
	status := e.EvidenceStatus
	if status == "" {
		status = "not_collected"
	}
	if e.EvidenceExpiresAt > 0 && now >= e.EvidenceExpiresAt {
		return abuseExcerpt{}, "expired", e.EvidenceExpiresAt, nil
	}
	if status != "available" {
		return abuseExcerpt{}, status, e.EvidenceExpiresAt, nil
	}
	plain, err := common.OpenAbuseEvidence(encrypted.Ciphertext, fmt.Sprintf("%d:%s", e.UserID, e.RequestID))
	if err != nil {
		return abuseExcerpt{}, "", 0, err
	}
	var excerpt abuseExcerpt
	if err = common.Unmarshal(plain, &excerpt); err != nil {
		return abuseExcerpt{}, "", 0, err
	}
	return excerpt, "available", e.EvidenceExpiresAt, nil
}
