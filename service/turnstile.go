package service

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/google/uuid"
)

const (
	TurnstileSiteKeyEnv          = "TURNSTILE_SITE_KEY"
	TurnstileSecretKeyEnv        = "TURNSTILE_SECRET_KEY"
	TurnstileExpectedHostnameEnv = "TURNSTILE_EXPECTED_HOSTNAME"
	TurnstileExpectedActionEnv   = "TURNSTILE_EXPECTED_ACTION"

	turnstileSiteverifyURL          = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	turnstileTokenMaxBytes          = 2048
	turnstileValidationTimeout      = 5 * time.Second
	turnstileValidationMaxAttempts  = 2
	turnstileValidationRetryDelay   = 100 * time.Millisecond
	turnstileValidationMaxBodyBytes = 64 * 1024
)

var ErrTurnstileVerificationUnavailable = errors.New("turnstile verification is unavailable")

// TurnstileValidationRequest contains only the transient values Siteverify
// accepts. Callers must not persist or log Token.
type TurnstileValidationRequest struct {
	Token    string
	RemoteIP string
}

type turnstileValidationConfig struct {
	secret           string
	expectedHostname string
	expectedAction   string
}

type turnstileSiteverifyResponse struct {
	Success  bool   `json:"success"`
	Hostname string `json:"hostname"`
	Action   string `json:"action"`
}

// TurnstileSiteKeyFromEnv returns the public key for the registration UI.
// The private verification secret intentionally has no equivalent accessor.
func TurnstileSiteKeyFromEnv() string {
	return strings.TrimSpace(os.Getenv(TurnstileSiteKeyEnv))
}

// TurnstileExpectedActionFromEnv returns the optional public action the
// registration widget must request for server-side validation.
func TurnstileExpectedActionFromEnv() string {
	return strings.TrimSpace(os.Getenv(TurnstileExpectedActionEnv))
}

// VerifyTurnstile validates one Turnstile token with Cloudflare Siteverify.
// It is fail-closed: callers receive the same public-safe error for a missing
// token, configuration problem, rejected token, or upstream failure.
func VerifyTurnstile(ctx context.Context, request TurnstileValidationRequest) error {
	config, err := turnstileValidationConfigFromEnv()
	if err != nil {
		return ErrTurnstileVerificationUnavailable
	}
	return verifyTurnstileWithHTTPClient(ctx, http.DefaultClient, turnstileSiteverifyURL, config, request)
}

func turnstileValidationConfigFromEnv() (turnstileValidationConfig, error) {
	secret, exists := os.LookupEnv(TurnstileSecretKeyEnv)
	if !exists || secret == "" || strings.TrimSpace(secret) != secret {
		return turnstileValidationConfig{}, ErrTurnstileVerificationUnavailable
	}
	return turnstileValidationConfig{
		secret:           secret,
		expectedHostname: strings.TrimSpace(os.Getenv(TurnstileExpectedHostnameEnv)),
		expectedAction:   TurnstileExpectedActionFromEnv(),
	}, nil
}

func verifyTurnstileWithHTTPClient(ctx context.Context, client *http.Client, endpoint string, config turnstileValidationConfig, input TurnstileValidationRequest) error {
	if client == nil || strings.TrimSpace(input.Token) == "" || len(input.Token) > turnstileTokenMaxBytes {
		return ErrTurnstileVerificationUnavailable
	}

	verificationContext, cancel := context.WithTimeout(ctx, turnstileValidationTimeout)
	defer cancel()
	idempotencyKey := uuid.NewString()
	for attempt := 0; attempt < turnstileValidationMaxAttempts; attempt++ {
		retry, err := verifyTurnstileOnce(verificationContext, client, endpoint, config, input, idempotencyKey)
		if err == nil {
			return nil
		}
		if !retry || attempt+1 == turnstileValidationMaxAttempts {
			return ErrTurnstileVerificationUnavailable
		}
		if err := waitForTurnstileRetry(verificationContext); err != nil {
			return ErrTurnstileVerificationUnavailable
		}
	}
	return ErrTurnstileVerificationUnavailable
}

func verifyTurnstileOnce(ctx context.Context, client *http.Client, endpoint string, config turnstileValidationConfig, input TurnstileValidationRequest, idempotencyKey string) (bool, error) {
	form := url.Values{
		"secret":          {config.secret},
		"response":        {input.Token},
		"idempotency_key": {idempotencyKey},
	}
	if net.ParseIP(input.RemoteIP) != nil {
		form.Set("remoteip", input.RemoteIP)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return ctx.Err() == nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, turnstileValidationMaxBodyBytes))
		return response.StatusCode >= http.StatusInternalServerError, ErrTurnstileVerificationUnavailable
	}

	var payload turnstileSiteverifyResponse
	if err := common.DecodeJson(io.LimitReader(response.Body, turnstileValidationMaxBodyBytes), &payload); err != nil {
		return false, err
	}
	if !payload.Success || (config.expectedHostname != "" && !strings.EqualFold(payload.Hostname, config.expectedHostname)) || (config.expectedAction != "" && payload.Action != config.expectedAction) {
		return false, ErrTurnstileVerificationUnavailable
	}
	return false, nil
}

func waitForTurnstileRetry(ctx context.Context) error {
	timer := time.NewTimer(turnstileValidationRetryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
