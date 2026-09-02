package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRegistrationInviteCode(t *testing.T) {
	valid := "inv_0123456789ABCDEFGHJK"
	require.Len(t, valid, RegistrationInviteCodeLength)
	require.NoError(t, ValidateRegistrationInviteCode(valid))

	for _, invalid := range []string{
		"",
		"INV_0123456789ABCDEFGHJK",
		"inv_0123456789ABCDEFGHJ",
		"inv_0123456789ABCDEFGHJI",
		" inv_0123456789ABCDEFGHJK",
	} {
		assert.ErrorIs(t, ValidateRegistrationInviteCode(invalid), ErrRegistrationInviteCodeInvalid, invalid)
	}
}
