package controller

import (
	"errors"
	"strconv"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

const (
	registrationInviteMaximumValidDays = 365
	registrationInviteMaximumNoteBytes = 255
	registrationInviteSecondsPerDay    = int64(24 * 60 * 60)

	registrationInviteInvalidRequestMessage  = "invalid registration invite request"
	registrationInviteUnavailableMessage     = "registration invite service is unavailable"
	registrationInviteOperationFailedMessage = "registration invite operation failed"
	registrationInviteNotFoundMessage        = "registration invite was not found"
	registrationInviteCannotRevokeMessage    = "used registration invites cannot be revoked"
)

type registrationInviteCreateRequest struct {
	Count     int    `json:"count"`
	ExpiresAt *int64 `json:"expires_at"`
	ValidDays *int   `json:"valid_days"`
	Note      string `json:"note"`
}

// registrationInviteCreatedResponse is intentionally the only HTTP DTO that
// contains a raw invite code. It is sent once, on creation, with no-store
// response headers applied by the route.
type registrationInviteCreatedResponse struct {
	Id         int    `json:"id"`
	Code       string `json:"code"`
	CodePrefix string `json:"code_prefix"`
	ExpiresAt  int64  `json:"expires_at"`
}

type registrationInviteCreateResponse struct {
	Invites []registrationInviteCreatedResponse `json:"invites"`
}

type registrationInviteListResponse struct {
	Page     int                              `json:"page"`
	PageSize int                              `json:"page_size"`
	Total    int64                            `json:"total"`
	Items    []service.RegistrationInviteView `json:"items"`
}

type registrationInviteRevokeResponse struct {
	Id      int  `json:"id"`
	Revoked bool `json:"revoked"`
}

func ListRegistrationInvites(c *gin.Context) {
	filters, page, pageSize, ok := registrationInviteListRequest(c)
	if !ok {
		common.ApiErrorMsg(c, registrationInviteInvalidRequestMessage)
		return
	}

	result, err := service.ListRegistrationInvites(filters, service.RegistrationInvitePagination{
		Offset: (page - 1) * pageSize,
		Limit:  pageSize,
	})
	if err != nil {
		if registrationInviteListInputError(err) {
			common.ApiErrorMsg(c, registrationInviteInvalidRequestMessage)
			return
		}
		registrationInviteInternalError(c, "list", err)
		return
	}

	common.ApiSuccess(c, registrationInviteListResponse{
		Page:     page,
		PageSize: pageSize,
		Total:    result.Total,
		Items:    result.Items,
	})
}

func CreateRegistrationInvites(c *gin.Context) {
	var request registrationInviteCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil || !validateRegistrationInviteCreateRequest(request) {
		recordRegistrationInviteCreateAudit(c, request.Count, nil, nil, "invalid_request")
		common.ApiErrorMsg(c, registrationInviteInvalidRequestMessage)
		return
	}

	expiresAt, ok := registrationInviteExpiryFromRequest(request, common.GetTimestamp())
	if !ok {
		recordRegistrationInviteCreateAudit(c, request.Count, nil, nil, "invalid_request")
		common.ApiErrorMsg(c, registrationInviteInvalidRequestMessage)
		return
	}

	secret, err := service.RegistrationInviteHMACSecretFromEnv()
	if err != nil {
		recordRegistrationInviteCreateAudit(c, request.Count, nil, nil, "unavailable")
		common.SysError("registration invite HMAC configuration is invalid")
		common.ApiErrorMsg(c, registrationInviteUnavailableMessage)
		return
	}

	generated, err := service.GenerateRegistrationInvites(request.Count, expiresAt, request.Note, c.GetInt("id"), secret)
	if err != nil {
		if registrationInviteGenerationInputError(err) {
			recordRegistrationInviteCreateAudit(c, request.Count, nil, nil, "invalid_request")
			common.ApiErrorMsg(c, registrationInviteInvalidRequestMessage)
			return
		}
		if errors.Is(err, service.ErrRegistrationInviteHMACSecretInvalid) {
			recordRegistrationInviteCreateAudit(c, request.Count, nil, nil, "unavailable")
			common.ApiErrorMsg(c, registrationInviteUnavailableMessage)
			return
		}
		recordRegistrationInviteCreateAudit(c, request.Count, nil, nil, "error")
		registrationInviteInternalError(c, "create", err)
		return
	}

	invites := make([]registrationInviteCreatedResponse, 0, len(generated))
	inviteIDs := make([]int, 0, len(generated))
	prefixes := make([]string, 0, len(generated))
	for _, invite := range generated {
		invites = append(invites, registrationInviteCreatedResponse{
			Id:         invite.Id,
			Code:       invite.Code,
			CodePrefix: invite.CodePrefix,
			ExpiresAt:  invite.ExpiresAt,
		})
		inviteIDs = append(inviteIDs, invite.Id)
		prefixes = append(prefixes, invite.CodePrefix)
	}
	recordRegistrationInviteCreateAudit(c, len(invites), inviteIDs, prefixes, "created")

	common.ApiSuccess(c, registrationInviteCreateResponse{Invites: invites})
}

func RevokeRegistrationInvite(c *gin.Context) {
	inviteID, err := strconv.Atoi(c.Param("id"))
	if err != nil || inviteID <= 0 {
		recordRegistrationInviteRevokeAudit(c, 0, "invalid_request")
		common.ApiErrorMsg(c, registrationInviteInvalidRequestMessage)
		return
	}

	revoked, err := service.RevokeRegistrationInvite(inviteID, c.GetInt("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRegistrationInviteNotFound):
			recordRegistrationInviteRevokeAudit(c, inviteID, "not_found")
			common.ApiErrorMsg(c, registrationInviteNotFoundMessage)
		case errors.Is(err, service.ErrRegistrationInviteAlreadyUsed):
			recordRegistrationInviteRevokeAudit(c, inviteID, "used")
			common.ApiErrorMsg(c, registrationInviteCannotRevokeMessage)
		default:
			recordRegistrationInviteRevokeAudit(c, inviteID, "error")
			registrationInviteInternalError(c, "revoke", err)
		}
		return
	}

	result := "already_revoked"
	if revoked {
		result = "revoked"
	}
	recordRegistrationInviteRevokeAudit(c, inviteID, result)
	common.ApiSuccess(c, registrationInviteRevokeResponse{Id: inviteID, Revoked: revoked})
}

func registrationInviteListRequest(c *gin.Context) (service.RegistrationInviteListFilters, int, int, bool) {
	page, ok := registrationInviteQueryInt(c, "p", 1, 0)
	if !ok {
		return service.RegistrationInviteListFilters{}, 0, 0, false
	}
	pageSize, ok := registrationInviteQueryInt(c, "page_size", service.RegistrationInviteListDefaultLimit, service.RegistrationInviteListMaximumLimit)
	if !ok {
		return service.RegistrationInviteListFilters{}, 0, 0, false
	}
	maxInt := int(^uint(0) >> 1)
	if page-1 > maxInt/pageSize {
		return service.RegistrationInviteListFilters{}, 0, 0, false
	}

	createdAfter, ok := registrationInviteOptionalTimestamp(c, "created_after")
	if !ok {
		return service.RegistrationInviteListFilters{}, 0, 0, false
	}
	createdBefore, ok := registrationInviteOptionalTimestamp(c, "created_before")
	if !ok || (createdAfter > 0 && createdBefore > 0 && createdAfter > createdBefore) {
		return service.RegistrationInviteListFilters{}, 0, 0, false
	}
	note := c.Query("note")
	if !utf8.ValidString(note) || len(note) > registrationInviteMaximumNoteBytes {
		return service.RegistrationInviteListFilters{}, 0, 0, false
	}

	return service.RegistrationInviteListFilters{
		Status:        service.RegistrationInviteStatus(c.Query("status")),
		CreatedAfter:  createdAfter,
		CreatedBefore: createdBefore,
		Note:          note,
	}, page, pageSize, true
}

func registrationInviteQueryInt(c *gin.Context, key string, defaultValue int, maximum int) (int, bool) {
	raw, present := c.GetQuery(key)
	if !present {
		return defaultValue, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || (maximum > 0 && value > maximum) {
		return 0, false
	}
	return value, true
}

func registrationInviteOptionalTimestamp(c *gin.Context, key string) (int64, bool) {
	raw, present := c.GetQuery(key)
	if !present {
		return 0, true
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, false
	}
	return value, true
}

func validateRegistrationInviteCreateRequest(request registrationInviteCreateRequest) bool {
	return request.Count >= 1 && request.Count <= service.RegistrationInviteBatchMax &&
		utf8.ValidString(request.Note) && len(request.Note) <= registrationInviteMaximumNoteBytes
}

func registrationInviteExpiryFromRequest(request registrationInviteCreateRequest, now int64) (int64, bool) {
	if request.ExpiresAt != nil && request.ValidDays != nil {
		return 0, false
	}
	maximumExpiresAt := now + registrationInviteMaximumValidDays*registrationInviteSecondsPerDay
	if request.ExpiresAt != nil {
		if *request.ExpiresAt <= now || *request.ExpiresAt > maximumExpiresAt {
			return 0, false
		}
		return *request.ExpiresAt, true
	}
	if request.ValidDays != nil {
		if *request.ValidDays < 1 || *request.ValidDays > registrationInviteMaximumValidDays {
			return 0, false
		}
		return now + int64(*request.ValidDays)*registrationInviteSecondsPerDay, true
	}
	return 0, true
}

func registrationInviteListInputError(err error) bool {
	return errors.Is(err, service.ErrRegistrationInviteStatusInvalid) ||
		errors.Is(err, service.ErrRegistrationInvitePaginationInvalid) ||
		errors.Is(err, service.ErrRegistrationInviteFilterInvalid)
}

func registrationInviteGenerationInputError(err error) bool {
	return errors.Is(err, service.ErrRegistrationInviteBatchSizeInvalid) ||
		errors.Is(err, service.ErrRegistrationInviteExpiryInvalid) ||
		errors.Is(err, service.ErrRegistrationInviteNoteInvalid) ||
		errors.Is(err, service.ErrRegistrationInviteCreatorInvalid)
}

func recordRegistrationInviteRevokeAudit(c *gin.Context, inviteID int, result string) {
	params := map[string]interface{}{"result": result}
	if inviteID > 0 {
		params["invite_id"] = inviteID
	}
	recordManageAudit(c, "registration_invite.revoke", params)
}

func recordRegistrationInviteCreateAudit(c *gin.Context, count int, inviteIDs []int, prefixes []string, result string) {
	params := map[string]interface{}{
		"count":  count,
		"result": result,
	}
	if len(inviteIDs) > 0 {
		params["invite_ids"] = inviteIDs
	}
	if len(prefixes) > 0 {
		params["code_prefixes"] = prefixes
	}
	recordManageAudit(c, "registration_invite.create", params)
}

func registrationInviteInternalError(c *gin.Context, operation string, err error) {
	common.SysError("registration invite " + operation + " failed: " + err.Error())
	common.ApiErrorMsg(c, registrationInviteOperationFailedMessage)
}
