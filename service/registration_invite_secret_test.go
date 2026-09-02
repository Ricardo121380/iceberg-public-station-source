package service

import (
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistrationInviteHMACSecretFromEnv(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		previous, existed := os.LookupEnv(RegistrationInviteHMACSecretEnv)
		require.NoError(t, os.Unsetenv(RegistrationInviteHMACSecretEnv))
		t.Cleanup(func() {
			if existed {
				_ = os.Setenv(RegistrationInviteHMACSecretEnv, previous)
				return
			}
			_ = os.Unsetenv(RegistrationInviteHMACSecretEnv)
		})

		_, err := RegistrationInviteHMACSecretFromEnv()
		assert.ErrorIs(t, err, ErrRegistrationInviteHMACSecretInvalid)
	})

	secret := strings.Repeat("s", RegistrationInviteHMACSecretMinBytes)
	t.Setenv(RegistrationInviteHMACSecretEnv, secret)

	configured, err := RegistrationInviteHMACSecretFromEnv()
	require.NoError(t, err)
	assert.Equal(t, secret, configured)

	t.Setenv(RegistrationInviteHMACSecretEnv, strings.Repeat("s", RegistrationInviteHMACSecretMinBytes-1))
	_, err = RegistrationInviteHMACSecretFromEnv()
	assert.ErrorIs(t, err, ErrRegistrationInviteHMACSecretInvalid)

	t.Setenv(RegistrationInviteHMACSecretEnv, " "+secret)
	_, err = RegistrationInviteHMACSecretFromEnv()
	assert.ErrorIs(t, err, ErrRegistrationInviteHMACSecretInvalid)
}

func TestHashRegistrationInviteCodeUsesValidatedInputs(t *testing.T) {
	secret := strings.Repeat("s", RegistrationInviteHMACSecretMinBytes)
	code := "inv_0123456789ABCDEFGHJK"

	first, err := HashRegistrationInviteCode(code, secret)
	require.NoError(t, err)
	second, err := HashRegistrationInviteCode(code, secret)
	require.NoError(t, err)
	assert.Len(t, first, 64)
	assert.Equal(t, first, second)
	assert.NotContains(t, first, code)

	_, err = HashRegistrationInviteCode("inv_0123456789ABCDEFGHJI", secret)
	assert.ErrorIs(t, err, model.ErrRegistrationInviteCodeInvalid)
	_, err = HashRegistrationInviteCode(code, strings.Repeat("s", RegistrationInviteHMACSecretMinBytes-1))
	assert.ErrorIs(t, err, ErrRegistrationInviteHMACSecretInvalid)
}
