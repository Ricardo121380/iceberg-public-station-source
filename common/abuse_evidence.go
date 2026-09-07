package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// Dedicated deployment secret, never the session secret or a database option.
// A missing key disables capture; an invalid key never falls back to plaintext.
func AbuseEvidenceCipher() (cipher.AEAD, error) {
	encoded := os.Getenv("ABUSE_EVIDENCE_KEY")
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(key) != 32 {
		return nil, errors.New("abuse evidence key unavailable")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func SealAbuseEvidence(plain []byte, identity string) (string, error) {
	aead, err := AbuseEvidenceCipher()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(aead.Seal(nonce, nonce, plain, []byte(identity))), nil
}

func OpenAbuseEvidence(encoded, identity string) ([]byte, error) {
	aead, err := AbuseEvidenceCipher()
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(data) < aead.NonceSize()+aead.Overhead() {
		return nil, errors.New("invalid abuse evidence")
	}
	return aead.Open(nil, data[:aead.NonceSize()], data[aead.NonceSize():], []byte(identity))
}

var evidenceSecrets = []*regexp.Regexp{
	regexp.MustCompile(`(?is)-----BEGIN[^\r\n]*PRIVATE KEY-----.*?(?:-----END[^\r\n]*PRIVATE KEY-----|$)`),
	regexp.MustCompile(`(?i)(?:https?://|data:)[^\s<>"']+`),
	regexp.MustCompile(`(?i)\b(?:bearer|basic)\s+[a-z0-9+/=_\-.]+`),
	regexp.MustCompile(`(?i)(?:["']?(?:api[_ -]?key|access[_ -]?token|refresh[_ -]?token|authorization|password|passwd|secret|cookie|密码|口令|密钥|令牌)["']?\s*(?:[:=：]|\bis\b|为|是)\s*)(?:"[^"\r\n]*"|'[^'\r\n]*'|[^\s,;，；]+)`),
	regexp.MustCompile(`(?i)\b(?:sk-|ghp_|github_pat_|AKIA)[a-z0-9_\-]+`),
	regexp.MustCompile(`\b[0-9]{8,12}:[a-zA-Z0-9_-]{20,}\b`),
	regexp.MustCompile(`\beyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\b`),
	regexp.MustCompile(`[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
	regexp.MustCompile(`\+?\d[\d ()-]{6,}\d`),
	regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`),
	regexp.MustCompile(`[a-zA-Z0-9_+/=-]{32,}`),
}

// Redact before truncation so a secret crossing the 300-character boundary is
// not partially leaked. This is best-effort minimization, not PII classification.
func RedactAbuseText(text string) (string, bool) {
	text = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, text)
	for _, pattern := range evidenceSecrets {
		text = pattern.ReplaceAllString(text, "[已脱敏]")
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) > 300 {
		return string(runes[:300]), true
	}
	return string(runes), false
}
