package service

import (
	"errors"
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

const (
	RegistrationInviteHMACSecretEnv      = "INVITE_HMAC_SECRET"
	RegistrationInviteHMACSecretMinBytes = 32
)

var ErrRegistrationInviteHMACSecretInvalid = errors.New("registration invite HMAC secret is invalid")

// RegistrationInviteHMACSecretFromEnv is intentionally called by the feature
// activation boundary in Task 05, rather than global initialization. Existing
// deployments therefore remain bootable until invitation registration is
// explicitly enabled.
func RegistrationInviteHMACSecretFromEnv() (string, error) {
	secret, exists := os.LookupEnv(RegistrationInviteHMACSecretEnv)
	if !exists {
		return "", ErrRegistrationInviteHMACSecretInvalid
	}
	if err := ValidateRegistrationInviteHMACSecret(secret); err != nil {
		return "", err
	}
	return secret, nil
}

func ValidateRegistrationInviteHMACSecret(secret string) error {
	if secret == "" || strings.TrimSpace(secret) != secret || len([]byte(secret)) < RegistrationInviteHMACSecretMinBytes {
		return ErrRegistrationInviteHMACSecretInvalid
	}
	return nil
}

func HashRegistrationInviteCode(code string, secret string) (string, error) {
	if err := model.ValidateRegistrationInviteCode(code); err != nil {
		return "", err
	}
	if err := ValidateRegistrationInviteHMACSecret(secret); err != nil {
		return "", err
	}
	return common.HmacSha256(code, secret), nil
}
