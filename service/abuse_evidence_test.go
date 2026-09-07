package service

import (
	"encoding/base64"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAbuseExcerptOnlyLastUserText(t *testing.T) {
	cases := []struct {
		name, path, body, text, status string
		omitted                        bool
	}{
		{"responses-string", "/v1/responses", `{"input":"问题 user@example.com"}`, "问题 [已脱敏]", "available", false},
		{"chat-last-user", "/v1/chat/completions", `{"messages":[{"role":"system","content":"system secret"},{"role":"user","content":"old private"},{"role":"user","content":"new question"},{"role":"assistant","content":"output secret"}]}`, "new question", "available", false},
		{"parts", "/v1/responses", `{"input":[{"role":"user","content":[{"type":"input_text","text":"question"},{"type":"input_image","image_url":"private-image"},{"type":"input_file","file_data":"private-file"}]}]}`, "question", "available", true},
		{"latest-attachment", "/v1/messages", `{"messages":[{"role":"user","content":"old private"},{"role":"user","content":[{"type":"image","source":{"data":"secret"}}]}]}`, "", "no_user_text", true},
		{"no-user", "/v1/responses", `{"input":[{"type":"function_call_output","output":"private"}]}`, "", "no_user_text", false},
		{"unsupported", "/v1/audio/transcriptions", `{"input":"private"}`, "", "unsupported_protocol", false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			e, status := extractAbuseExcerpt([]byte(tt.body), tt.path)
			assert.Equal(t, tt.status, status)
			assert.Equal(t, tt.text, e.Text)
			assert.Equal(t, tt.omitted, e.OmittedParts)
		})
	}
}

func TestSafetyEventCapturesEncryptedExcerptAndAuditsRead(t *testing.T) {
	c, db := safetyTestContext(t)
	t.Setenv("ABUSE_EVIDENCE_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))))
	body := `{"input":"test question password=secret-value"}`
	storage, err := common.CreateBodyStorage([]byte(body))
	require.NoError(t, err)
	defer storage.Close()
	c.Set(common.KeyBodyStorage, storage)
	finish, err := BeginSafetyObservation(c)
	require.NoError(t, err)
	ObserveSafetyPayload(c, []byte(`{"error":{"code":"content_policy_violation"}}`), "error")
	finish()
	var event model.AbuseEvent
	require.NoError(t, db.First(&event).Error)
	assert.False(t, event.Actionable)
	assert.Equal(t, "available", event.EvidenceStatus)
	var cipher model.AbuseEvidence
	require.NoError(t, db.First(&cipher).Error)
	assert.NotContains(t, cipher.Ciphertext, "test question")
	bytes, err := common.Marshal(event)
	require.NoError(t, err)
	assert.NotContains(t, string(bytes), "secret-value")
	assert.NotContains(t, string(bytes), cipher.Ciphertext)
	excerpt, status, _, err := ReadAbuseExcerpt(event.ID, 100, time.Now().Unix())
	require.NoError(t, err)
	assert.Equal(t, "available", status)
	assert.Equal(t, "test question [已脱敏]", excerpt.Text)
	var audits []model.AbuseReviewAudit
	require.NoError(t, db.Find(&audits).Error)
	require.Len(t, audits, 1)
	assert.Equal(t, 100, audits[0].OperatorID)
	reader, err := storage.NewReader()
	require.NoError(t, err)
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, body, string(raw))
	_, status, _, err = ReadAbuseExcerpt(event.ID, 100, event.EvidenceExpiresAt)
	require.NoError(t, err)
	assert.Equal(t, "expired", status)
	require.NoError(t, model.CleanupAbuseEvidence(event.EvidenceExpiresAt))
	var count int64
	require.NoError(t, db.Model(&model.AbuseEvidence{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestNormalRequestsDoNotCaptureAndAuditFailureBlocksRead(t *testing.T) {
	c, db := safetyTestContext(t)
	t.Setenv("ABUSE_EVIDENCE_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))))
	finish, err := BeginSafetyObservation(c)
	require.NoError(t, err)
	finish()
	var count int64
	require.NoError(t, db.Model(&model.AbuseEvidence{}).Count(&count).Error)
	assert.Zero(t, count)
	event := model.AbuseEvent{UserID: 1, RequestID: "older", CreatedAt: time.Now().Unix()}
	require.NoError(t, db.Create(&event).Error)
	_, status, _, err := ReadAbuseExcerpt(event.ID, 100, time.Now().Unix())
	require.NoError(t, err)
	assert.Equal(t, "not_collected", status)
	require.NoError(t, db.Migrator().DropTable(&model.AbuseReviewAudit{}))
	_, _, _, err = ReadAbuseExcerpt(event.ID, 100, time.Now().Unix())
	require.Error(t, err)
}

func TestAbuseEvidenceCaptureUnavailableReasons(t *testing.T) {
	for _, tc := range []struct{ name, key, body, want string }{
		{"missing-key", "", `{"input":"private"}`, "key_unavailable"},
		{"oversized", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))), strings.Repeat("x", (1<<20)+1), "body_too_large"},
		{"attachments-only", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))), `{"input":[{"role":"user","content":[{"type":"input_image","image_url":"private"}]}]}`, "no_user_text"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, db := safetyTestContext(t)
			t.Setenv("ABUSE_EVIDENCE_KEY", tc.key)
			storage, err := common.CreateBodyStorage([]byte(tc.body))
			require.NoError(t, err)
			defer storage.Close()
			c.Set(common.KeyBodyStorage, storage)
			finish, err := BeginSafetyObservation(c)
			require.NoError(t, err)
			ObserveSafetyPayload(c, []byte(`{"error":{"code":"content_filter"}}`), "error")
			finish()
			var e model.AbuseEvent
			require.NoError(t, db.First(&e).Error)
			assert.Equal(t, tc.want, e.EvidenceStatus)
			var count int64
			require.NoError(t, db.Model(&model.AbuseEvidence{}).Count(&count).Error)
			assert.Zero(t, count)
		})
	}
}

func TestAbuseEvidenceOffAndBodyCursorPreserved(t *testing.T) {
	c, db := safetyTestContext(t)
	t.Setenv("ABUSE_EVIDENCE_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))))
	storage, err := common.CreateBodyStorage([]byte(`{"input":"user text"}`))
	require.NoError(t, err)
	defer storage.Close()
	c.Set(common.KeyBodyStorage, storage)
	_, err = storage.Seek(5, io.SeekStart)
	require.NoError(t, err)
	event := model.AbuseEvent{UserID: 1, RequestID: "cursor", CreatedAt: time.Now().Unix()}
	captureAbuseEvidence(c, &event)
	position, err := storage.Seek(0, io.SeekCurrent)
	require.NoError(t, err)
	assert.EqualValues(t, 5, position)
	require.NotNil(t, event.Evidence)
	require.NoError(t, db.Model(&model.AbusePolicy{}).Where("id = ?", 1).Update("mode", "off").Error)
	finish, err := BeginSafetyObservation(c)
	require.NoError(t, err)
	ObserveSafetyPayload(c, []byte(`{"error":{"code":"content_filter"}}`), "error")
	finish()
	var eventCount int64
	require.NoError(t, db.Model(&model.AbuseEvent{}).Count(&eventCount).Error)
	assert.Zero(t, eventCount)
	var count int64
	require.NoError(t, db.Model(&model.AbuseEvidence{}).Count(&count).Error)
	assert.Zero(t, count)
}
