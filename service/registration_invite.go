package service

import (
	"crypto/hmac"
	cryptorand "crypto/rand"
	"errors"
	"io"
	"math/big"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"gorm.io/gorm"
)

const (
	RegistrationInviteBatchMax          = 100
	RegistrationInviteListDefaultLimit  = 20
	RegistrationInviteListMaximumLimit  = 100
	RegistrationInviteDefaultExpiry     = 7 * 24 * time.Hour
	registrationInviteGenerationRetries = 5
)

var (
	ErrRegistrationInviteBatchSizeInvalid    = errors.New("registration invite batch size is invalid")
	ErrRegistrationInviteExpiryInvalid       = errors.New("registration invite expiry is invalid")
	ErrRegistrationInviteNoteInvalid         = errors.New("registration invite note is invalid")
	ErrRegistrationInviteCreatorInvalid      = errors.New("registration invite creator is invalid")
	ErrRegistrationInviteOperatorInvalid     = errors.New("registration invite operator is invalid")
	ErrRegistrationInviteUserInvalid         = errors.New("registration invite user is invalid")
	ErrRegistrationInviteTransactionInvalid  = errors.New("registration invite transaction is invalid")
	ErrRegistrationInviteNotFound            = errors.New("registration invite was not found")
	ErrRegistrationInviteExpired             = errors.New("registration invite has expired")
	ErrRegistrationInviteAlreadyUsed         = errors.New("registration invite has already been used")
	ErrRegistrationInviteRevoked             = errors.New("registration invite has been revoked")
	ErrRegistrationInviteGenerationExhausted = errors.New("registration invite generation retries were exhausted")
	ErrRegistrationInviteStatusInvalid       = errors.New("registration invite status is invalid")
	ErrRegistrationInvitePaginationInvalid   = errors.New("registration invite pagination is invalid")
	ErrRegistrationInviteFilterInvalid       = errors.New("registration invite filter is invalid")
	ErrRegistrationInviteUnavailable         = errors.New("registration invite is unavailable")
)

type RegistrationInviteStatus string

const (
	RegistrationInviteStatusActive  RegistrationInviteStatus = "active"
	RegistrationInviteStatusExpired RegistrationInviteStatus = "expired"
	RegistrationInviteStatusUsed    RegistrationInviteStatus = "used"
	RegistrationInviteStatusRevoked RegistrationInviteStatus = "revoked"
)

// GeneratedRegistrationInvite is the sole domain result that carries a raw
// invite code. Its JSON representation deliberately omits Code so later HTTP
// handlers must explicitly decide when the one-time display is appropriate.
type GeneratedRegistrationInvite struct {
	Id         int    `json:"id"`
	Code       string `json:"-"`
	CodePrefix string `json:"code_prefix"`
	ExpiresAt  int64  `json:"expires_at"`
}

// RegistrationInviteValidation contains only what the registration flow needs
// after a raw code has been checked. It never exposes its hash or raw code.
type RegistrationInviteValidation struct {
	InviteID  int   `json:"invite_id"`
	ExpiresAt int64 `json:"expires_at"`
}

// RegistrationInviteView is safe for list and admin-detail DTOs.
type RegistrationInviteView struct {
	Id         int                      `json:"id"`
	CodePrefix string                   `json:"code_prefix"`
	Note       string                   `json:"note"`
	CreatedBy  int                      `json:"created_by"`
	CreatedAt  int64                    `json:"created_at"`
	ExpiresAt  int64                    `json:"expires_at"`
	UsedBy     *int                     `json:"used_by,omitempty"`
	UsedAt     *int64                   `json:"used_at,omitempty"`
	RevokedBy  *int                     `json:"revoked_by,omitempty"`
	RevokedAt  *int64                   `json:"revoked_at,omitempty"`
	Status     RegistrationInviteStatus `json:"status"`
}

type RegistrationInviteListFilters struct {
	Status        RegistrationInviteStatus
	CreatedBy     int
	UsedBy        int
	CreatedAfter  int64
	CreatedBefore int64
	Note          string
}

type RegistrationInvitePagination struct {
	Offset int
	Limit  int
}

type RegistrationInviteListResult struct {
	Items []RegistrationInviteView `json:"items"`
	Total int64                    `json:"total"`
}

// newRegistrationInviteCode uses crypto/rand and the canonical invite alphabet.
// It stays internal so GenerateRegistrationInvites is the only public domain
// call that returns a raw registration capability.
func newRegistrationInviteCode() (string, error) {
	return generateRegistrationInviteCode(cryptorand.Reader)
}

func generateRegistrationInviteCode(randomSource io.Reader) (string, error) {
	if randomSource == nil {
		return "", errors.New("registration invite random source is invalid")
	}

	randomPart := make([]byte, model.RegistrationInviteCodeRandomLength)
	alphabetLength := big.NewInt(int64(len(model.RegistrationInviteCodeAlphabet)))
	for index := range randomPart {
		number, err := cryptorand.Int(randomSource, alphabetLength)
		if err != nil {
			return "", err
		}
		randomPart[index] = model.RegistrationInviteCodeAlphabet[number.Int64()]
	}
	return model.RegistrationInviteCodePrefix + string(randomPart), nil
}

// GenerateRegistrationInvites writes a batch in one transaction. If the
// bounded collision retry cannot produce every row, the transaction rolls back.
func GenerateRegistrationInvites(count int, expiresAt int64, note string, creatorID int, secret string) ([]GeneratedRegistrationInvite, error) {
	return generateRegistrationInvites(model.DB, count, expiresAt, note, creatorID, secret, newRegistrationInviteCode)
}

func generateRegistrationInvites(db *gorm.DB, count int, expiresAt int64, note string, creatorID int, secret string, generateCode func() (string, error)) ([]GeneratedRegistrationInvite, error) {
	now := common.GetTimestamp()
	if err := validateRegistrationInviteGenerationInput(count, expiresAt, note, creatorID, secret, now); err != nil {
		return nil, err
	}
	if expiresAt == 0 {
		expiresAt = now + int64(RegistrationInviteDefaultExpiry/time.Second)
	}

	generated := make([]GeneratedRegistrationInvite, 0, count)
	err := db.Transaction(func(tx *gorm.DB) error {
		for range count {
			created := false
			for range registrationInviteGenerationRetries {
				rawCode, err := generateCode()
				if err != nil {
					return err
				}
				codeHash, err := HashRegistrationInviteCode(rawCode, secret)
				if err != nil {
					return err
				}

				var existing model.RegistrationInvite
				err = tx.Select("id").Where("code_hash = ?", codeHash).First(&existing).Error
				if err == nil {
					continue
				}
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}

				invite := &model.RegistrationInvite{
					CodeHash:   codeHash,
					CodePrefix: rawCode[:model.RegistrationInviteCodePrefixLength],
					Note:       note,
					CreatedBy:  creatorID,
					ExpiresAt:  expiresAt,
				}
				if err := tx.Create(invite).Error; err != nil {
					return err
				}
				generated = append(generated, GeneratedRegistrationInvite{
					Id:         invite.Id,
					Code:       rawCode,
					CodePrefix: invite.CodePrefix,
					ExpiresAt:  invite.ExpiresAt,
				})
				created = true
				break
			}
			if !created {
				return ErrRegistrationInviteGenerationExhausted
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return generated, nil
}

func validateRegistrationInviteGenerationInput(count int, expiresAt int64, note string, creatorID int, secret string, now int64) error {
	if count <= 0 || count > RegistrationInviteBatchMax {
		return ErrRegistrationInviteBatchSizeInvalid
	}
	if expiresAt != 0 && expiresAt <= now {
		return ErrRegistrationInviteExpiryInvalid
	}
	if len(note) > 255 {
		return ErrRegistrationInviteNoteInvalid
	}
	if creatorID <= 0 {
		return ErrRegistrationInviteCreatorInvalid
	}
	return ValidateRegistrationInviteHMACSecret(secret)
}

// ValidateRegistrationInvite resolves a raw code to its one-time invite ID.
// Detailed errors are for trusted API/OAuth code; public callers must use
// RegistrationInvitePublicError before rendering an error response.
func ValidateRegistrationInvite(rawCode string, secret string) (*RegistrationInviteValidation, error) {
	codeHash, err := HashRegistrationInviteCode(rawCode, secret)
	if err != nil {
		return nil, err
	}

	var invite model.RegistrationInvite
	if err := model.DB.Where("code_hash = ?", codeHash).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegistrationInviteNotFound
		}
		return nil, err
	}
	if !hmac.Equal([]byte(invite.CodeHash), []byte(codeHash)) {
		return nil, ErrRegistrationInviteNotFound
	}
	if err := registrationInviteStateError(&invite, common.GetTimestamp()); err != nil {
		return nil, err
	}
	return &RegistrationInviteValidation{InviteID: invite.Id, ExpiresAt: invite.ExpiresAt}, nil
}

// ConsumeRegistrationInvite must run inside the same transaction as user
// creation. It rechecks the stored state before its compare-and-swap update.
func ConsumeRegistrationInvite(tx *gorm.DB, inviteID int, userID int) error {
	if tx == nil {
		return ErrRegistrationInviteTransactionInvalid
	}
	if inviteID <= 0 {
		return ErrRegistrationInviteNotFound
	}
	if userID <= 0 {
		return ErrRegistrationInviteUserInvalid
	}

	now := common.GetTimestamp()
	invite, err := model.GetRegistrationInviteForUpdate(tx, inviteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRegistrationInviteNotFound
		}
		return err
	}
	if err := registrationInviteStateError(invite, now); err != nil {
		return err
	}

	consumed, err := model.ConsumeRegistrationInviteIfAvailable(tx, inviteID, userID, now)
	if err != nil {
		return err
	}
	if !consumed {
		return ErrRegistrationInviteAlreadyUsed
	}
	return nil
}

// RevokeRegistrationInvite records the responsible operator. Repeating a
// completed revoke is idempotent; consuming a code remains a hard failure.
func RevokeRegistrationInvite(inviteID int, operatorID int) (bool, error) {
	if inviteID <= 0 {
		return false, ErrRegistrationInviteNotFound
	}
	if operatorID <= 0 {
		return false, ErrRegistrationInviteOperatorInvalid
	}

	var revoked bool
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		invite, err := model.GetRegistrationInviteForUpdate(tx, inviteID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRegistrationInviteNotFound
			}
			return err
		}
		if invite.UsedAt != nil {
			return ErrRegistrationInviteAlreadyUsed
		}
		if invite.RevokedAt != nil {
			return nil
		}

		revoked, err = model.RevokeRegistrationInviteIfUnused(tx, inviteID, operatorID, common.GetTimestamp())
		if err != nil {
			return err
		}
		if !revoked {
			return ErrRegistrationInviteAlreadyUsed
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return revoked, nil
}

// ListRegistrationInvites returns only safe, derived invite fields. The code
// hash is selected neither by SQL nor by the returned DTO.
func ListRegistrationInvites(filters RegistrationInviteListFilters, pagination RegistrationInvitePagination) (*RegistrationInviteListResult, error) {
	if err := validateRegistrationInviteListInput(filters, pagination); err != nil {
		return nil, err
	}
	if pagination.Limit == 0 {
		pagination.Limit = RegistrationInviteListDefaultLimit
	}

	now := common.GetTimestamp()
	query := model.DB.Model(&model.RegistrationInvite{})
	if filters.CreatedBy > 0 {
		query = query.Where("created_by = ?", filters.CreatedBy)
	}
	if filters.UsedBy > 0 {
		query = query.Where("used_by = ?", filters.UsedBy)
	}
	if filters.CreatedAfter > 0 {
		query = query.Where("created_at >= ?", filters.CreatedAfter)
	}
	if filters.CreatedBefore > 0 {
		query = query.Where("created_at <= ?", filters.CreatedBefore)
	}
	if filters.Note != "" {
		query = query.Where("note LIKE ?", "%"+filters.Note+"%")
	}
	query, err := applyRegistrationInviteStatusFilter(query, filters.Status, now)
	if err != nil {
		return nil, err
	}

	result := &RegistrationInviteListResult{}
	if err := query.Count(&result.Total).Error; err != nil {
		return nil, err
	}

	var invites []model.RegistrationInvite
	if err := query.
		Select("id", "code_prefix", "note", "created_by", "created_at", "expires_at", "used_by", "used_at", "revoked_by", "revoked_at").
		Order("id DESC").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&invites).Error; err != nil {
		return nil, err
	}

	result.Items = make([]RegistrationInviteView, 0, len(invites))
	for index := range invites {
		result.Items = append(result.Items, registrationInviteView(&invites[index], now))
	}
	return result, nil
}

func validateRegistrationInviteListInput(filters RegistrationInviteListFilters, pagination RegistrationInvitePagination) error {
	if filters.CreatedBy < 0 || filters.UsedBy < 0 {
		return ErrRegistrationInvitePaginationInvalid
	}
	if filters.CreatedAfter < 0 || filters.CreatedBefore < 0 ||
		(filters.CreatedAfter > 0 && filters.CreatedBefore > 0 && filters.CreatedAfter > filters.CreatedBefore) ||
		len(filters.Note) > 255 {
		return ErrRegistrationInviteFilterInvalid
	}
	if pagination.Offset < 0 || pagination.Limit < 0 || pagination.Limit > RegistrationInviteListMaximumLimit {
		return ErrRegistrationInvitePaginationInvalid
	}
	if filters.Status == "" ||
		filters.Status == RegistrationInviteStatusActive ||
		filters.Status == RegistrationInviteStatusExpired ||
		filters.Status == RegistrationInviteStatusUsed ||
		filters.Status == RegistrationInviteStatusRevoked {
		return nil
	}
	return ErrRegistrationInviteStatusInvalid
}

func applyRegistrationInviteStatusFilter(query *gorm.DB, status RegistrationInviteStatus, now int64) (*gorm.DB, error) {
	switch status {
	case "":
		return query, nil
	case RegistrationInviteStatusActive:
		return query.Where("used_at IS NULL AND revoked_at IS NULL AND expires_at > ?", now), nil
	case RegistrationInviteStatusExpired:
		return query.Where("used_at IS NULL AND revoked_at IS NULL AND expires_at <= ?", now), nil
	case RegistrationInviteStatusUsed:
		return query.Where("used_at IS NOT NULL"), nil
	case RegistrationInviteStatusRevoked:
		return query.Where("revoked_at IS NOT NULL"), nil
	default:
		return nil, ErrRegistrationInviteStatusInvalid
	}
}

func registrationInviteView(invite *model.RegistrationInvite, now int64) RegistrationInviteView {
	return RegistrationInviteView{
		Id:         invite.Id,
		CodePrefix: invite.CodePrefix,
		Note:       invite.Note,
		CreatedBy:  invite.CreatedBy,
		CreatedAt:  invite.CreatedAt,
		ExpiresAt:  invite.ExpiresAt,
		UsedBy:     invite.UsedBy,
		UsedAt:     invite.UsedAt,
		RevokedBy:  invite.RevokedBy,
		RevokedAt:  invite.RevokedAt,
		Status:     registrationInviteStatus(invite, now),
	}
}

func registrationInviteStateError(invite *model.RegistrationInvite, now int64) error {
	switch registrationInviteStatus(invite, now) {
	case RegistrationInviteStatusRevoked:
		return ErrRegistrationInviteRevoked
	case RegistrationInviteStatusUsed:
		return ErrRegistrationInviteAlreadyUsed
	case RegistrationInviteStatusExpired:
		return ErrRegistrationInviteExpired
	default:
		return nil
	}
}

func registrationInviteStatus(invite *model.RegistrationInvite, now int64) RegistrationInviteStatus {
	if invite.RevokedAt != nil {
		return RegistrationInviteStatusRevoked
	}
	if invite.UsedAt != nil {
		return RegistrationInviteStatusUsed
	}
	if invite.ExpiresAt <= now {
		return RegistrationInviteStatusExpired
	}
	return RegistrationInviteStatusActive
}

// RegistrationInvitePublicError hides whether a submitted raw code existed.
// Infrastructure errors remain intact so callers can return a server failure.
func RegistrationInvitePublicError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, model.ErrRegistrationInviteCodeInvalid) ||
		errors.Is(err, ErrRegistrationInviteHMACSecretInvalid) ||
		errors.Is(err, ErrRegistrationInviteNotFound) ||
		errors.Is(err, ErrRegistrationInviteExpired) ||
		errors.Is(err, ErrRegistrationInviteAlreadyUsed) ||
		errors.Is(err, ErrRegistrationInviteRevoked) {
		return ErrRegistrationInviteUnavailable
	}
	return err
}
