package model

import (
	"errors"
	"strings"

	"gorm.io/gorm"
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
	registrationInviteRevokedByIndex   = "idx_registration_invites_revoked_by"
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
	RevokedBy  *int   `json:"revoked_by,omitempty" gorm:"index:idx_registration_invites_revoked_by"`
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

// GetRegistrationInviteForUpdate reads an invite while retaining a row lock
// until the caller's transaction ends. SQLite intentionally falls back to the
// transaction and compare-and-swap updates below because it has no FOR UPDATE.
func GetRegistrationInviteForUpdate(tx *gorm.DB, id int) (*RegistrationInvite, error) {
	invite := &RegistrationInvite{}
	if err := lockForUpdate(tx).Where("id = ?", id).First(invite).Error; err != nil {
		return nil, err
	}
	return invite, nil
}

// ConsumeRegistrationInviteIfAvailable is the compare-and-swap write used by
// the invitation registration transaction. It succeeds only once per invite.
func ConsumeRegistrationInviteIfAvailable(tx *gorm.DB, id int, userID int, usedAt int64) (bool, error) {
	result := tx.Model(&RegistrationInvite{}).
		Where("id = ? AND used_at IS NULL AND revoked_at IS NULL AND expires_at > ?", id, usedAt).
		Updates(map[string]any{
			"used_by": userID,
			"used_at": usedAt,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

// RevokeRegistrationInviteIfUnused is the compare-and-swap write used by the
// admin revoke operation. A consumed invite can never become revoked.
func RevokeRegistrationInviteIfUnused(tx *gorm.DB, id int, operatorID int, revokedAt int64) (bool, error) {
	result := tx.Model(&RegistrationInvite{}).
		Where("id = ? AND used_at IS NULL AND revoked_at IS NULL", id).
		Updates(map[string]any{
			"revoked_by": operatorID,
			"revoked_at": revokedAt,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}
