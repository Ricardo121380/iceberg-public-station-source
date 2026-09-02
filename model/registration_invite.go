package model

import (
	"errors"
	"strings"
)

const (
	RegistrationInviteTableName        = "registration_invites"
	RegistrationInviteCodePrefix       = "inv_"
	RegistrationInviteCodeRandomLength = 20
	RegistrationInviteCodeLength       = len(RegistrationInviteCodePrefix) + RegistrationInviteCodeRandomLength
	RegistrationInviteCodePrefixLength = len(RegistrationInviteCodePrefix) + 4
	RegistrationInviteCodeAlphabet     = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	registrationInviteCodeHashIndex    = "uk_registration_invites_code_hash"
	registrationInviteExpiresAtIndex   = "idx_registration_invites_expires_at"
	registrationInviteUsedAtIndex      = "idx_registration_invites_used_at"
	registrationInviteRevokedAtIndex   = "idx_registration_invites_revoked_at"
	registrationInviteCreatedByIndex   = "idx_registration_invites_created_by"
	registrationInviteUsedByIndex      = "idx_registration_invites_used_by"
)

var ErrRegistrationInviteCodeInvalid = errors.New("registration invite code is invalid")

// RegistrationInvite records one auditable, one-time registration capability.
// CodeHash is deliberately omitted from JSON so callers can never receive the
// database lookup value through ordinary model serialization.
type RegistrationInvite struct {
	Id         int    `json:"id"`
	CodeHash   string `json:"-" gorm:"type:char(64);not null;uniqueIndex:uk_registration_invites_code_hash"`
	CodePrefix string `json:"code_prefix" gorm:"type:varchar(8);not null"`
	Note       string `json:"note" gorm:"type:varchar(255)"`
	CreatedBy  int    `json:"created_by" gorm:"not null;index:idx_registration_invites_created_by"`
	CreatedAt  int64  `json:"created_at" gorm:"bigint;autoCreateTime;not null"`
	ExpiresAt  int64  `json:"expires_at" gorm:"bigint;not null;index:idx_registration_invites_expires_at"`
	UsedBy     *int   `json:"used_by,omitempty" gorm:"index:idx_registration_invites_used_by"`
	UsedAt     *int64 `json:"used_at,omitempty" gorm:"bigint;index:idx_registration_invites_used_at"`
	RevokedAt  *int64 `json:"revoked_at,omitempty" gorm:"bigint;index:idx_registration_invites_revoked_at"`
}

func (RegistrationInvite) TableName() string {
	return RegistrationInviteTableName
}

func ValidateRegistrationInviteCode(code string) error {
	if len(code) != RegistrationInviteCodeLength ||
		!strings.HasPrefix(code, RegistrationInviteCodePrefix) ||
		strings.TrimSpace(code) != code {
		return ErrRegistrationInviteCodeInvalid
	}
	for _, character := range code[len(RegistrationInviteCodePrefix):] {
		if !strings.ContainsRune(RegistrationInviteCodeAlphabet, character) {
			return ErrRegistrationInviteCodeInvalid
		}
	}
	return nil
}
