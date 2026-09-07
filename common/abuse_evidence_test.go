package common

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestAbuseEvidenceEncryptionAndIdentity(t *testing.T) {
	t.Setenv("ABUSE_EVIDENCE_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("a", 32))))
	encrypted, err := SealAbuseEvidence([]byte("测试输入"), "1:request")
	require.NoError(t, err)
	assert.NotContains(t, encrypted, "测试输入")
	plain, err := OpenAbuseEvidence(encrypted, "1:request")
	require.NoError(t, err)
	assert.Equal(t, "测试输入", string(plain))
	_, err = OpenAbuseEvidence(encrypted, "2:request")
	require.Error(t, err)
	_, err = OpenAbuseEvidence(encrypted[:len(encrypted)-4]+"AAAA", "1:request")
	require.Error(t, err)
	t.Setenv("ABUSE_EVIDENCE_KEY", "")
	_, err = SealAbuseEvidence([]byte("test"), "1")
	require.Error(t, err)
}

func TestAbuseRedactionBeforeUnicodeLimit(t *testing.T) {
	for _, secret := range []string{"sk-privateKEY123", "user@example.com", "+86 13800138000", "password=privateValue", "密码：测试口令", "password is privateValue", "密码是测试口令", "Bearer abc.def", "https://example.com/?token=private", "123456789:abcdefghijklmno01234567890", "-----BEGIN PRIVATE KEY-----\nprivate\n-----END PRIVATE KEY-----"} {
		value, _ := RedactAbuseText("问题 " + secret + " 继续")
		assert.NotContains(t, value, secret)
		assert.Contains(t, value, "[已脱敏]")
	}
	value, truncated := RedactAbuseText(strings.Repeat("文", 295) + " sk-verylongsecret1234567890" + strings.Repeat("字", 20))
	assert.True(t, truncated)
	assert.Len(t, []rune(value), 300)
	assert.NotContains(t, value, "sk-")
}
