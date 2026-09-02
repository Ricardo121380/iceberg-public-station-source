package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type turnstileRoundTripper func(*http.Request) (*http.Response, error)

func (roundTrip turnstileRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func TestTurnstileValidationConfigFromEnv(t *testing.T) {
	t.Setenv(TurnstileSecretKeyEnv, "test-turnstile-secret")
	t.Setenv(TurnstileSiteKeyEnv, " test-turnstile-site-key ")
	t.Setenv(TurnstileExpectedHostnameEnv, "login.example.test")
	t.Setenv(TurnstileExpectedActionEnv, "linuxdo_login")

	config, err := turnstileValidationConfigFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "test-turnstile-secret", config.secret)
	assert.Equal(t, "test-turnstile-site-key", TurnstileSiteKeyFromEnv())
	assert.Equal(t, "linuxdo_login", TurnstileExpectedActionFromEnv())
	assert.Equal(t, "login.example.test", config.expectedHostname)
	assert.Equal(t, "linuxdo_login", config.expectedAction)

	t.Setenv(TurnstileSecretKeyEnv, " test-turnstile-secret")
	_, err = turnstileValidationConfigFromEnv()
	assert.ErrorIs(t, err, ErrTurnstileVerificationUnavailable)
}

func TestVerifyTurnstileWithHTTPClient(t *testing.T) {
	config := turnstileValidationConfig{
		secret:           "test-turnstile-secret",
		expectedHostname: "login.example.test",
		expectedAction:   "linuxdo_login",
	}
	input := TurnstileValidationRequest{
		Token:    "test-turnstile-token",
		RemoteIP: "198.51.100.24",
	}

	t.Run("sends the expected Siteverify request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			require.Equal(t, http.MethodPost, request.Method)
			if !assert.NoError(t, request.ParseForm()) {
				return
			}
			assert.Equal(t, config.secret, request.PostForm.Get("secret"))
			assert.Equal(t, input.Token, request.PostForm.Get("response"))
			assert.Equal(t, input.RemoteIP, request.PostForm.Get("remoteip"))
			assert.NotEmpty(t, request.PostForm.Get("idempotency_key"))
			writer.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(writer, `{"success":true,"hostname":"login.example.test","action":"linuxdo_login"}`)
		}))
		defer server.Close()

		err := verifyTurnstileWithHTTPClient(context.Background(), server.Client(), server.URL, config, input)
		require.NoError(t, err)
	})

	t.Run("fails closed on rejected token or expected field mismatch", func(t *testing.T) {
		tests := []struct {
			name string
			body string
		}{
			{name: "rejected", body: `{"success":false}`},
			{name: "hostname mismatch", body: `{"success":true,"hostname":"other.example.test","action":"linuxdo_login"}`},
			{name: "action mismatch", body: `{"success":true,"hostname":"login.example.test","action":"other_action"}`},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					calls++
					writer.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(writer, test.body)
				}))
				defer server.Close()

				err := verifyTurnstileWithHTTPClient(context.Background(), server.Client(), server.URL, config, input)
				assert.ErrorIs(t, err, ErrTurnstileVerificationUnavailable)
				assert.Equal(t, 1, calls)
			})
		}
	})

	t.Run("fails closed on client errors and malformed responses", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
			body       string
		}{
			{name: "client error", statusCode: http.StatusBadRequest, body: `{"success":false}`},
			{name: "malformed response", statusCode: http.StatusOK, body: `{`},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				calls := 0
				server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					calls++
					writer.Header().Set("Content-Type", "application/json")
					writer.WriteHeader(test.statusCode)
					_, _ = io.WriteString(writer, test.body)
				}))
				defer server.Close()

				err := verifyTurnstileWithHTTPClient(context.Background(), server.Client(), server.URL, config, input)
				assert.ErrorIs(t, err, ErrTurnstileVerificationUnavailable)
				assert.Equal(t, 1, calls)
			})
		}
	})

	t.Run("fails closed when Siteverify rejects a reused token", func(t *testing.T) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			calls++
			writer.Header().Set("Content-Type", "application/json")
			if calls == 1 {
				_, _ = io.WriteString(writer, `{"success":true,"hostname":"login.example.test","action":"linuxdo_login"}`)
				return
			}
			_, _ = io.WriteString(writer, `{"success":false}`)
		}))
		defer server.Close()

		require.NoError(t, verifyTurnstileWithHTTPClient(context.Background(), server.Client(), server.URL, config, input))
		assert.ErrorIs(t, verifyTurnstileWithHTTPClient(context.Background(), server.Client(), server.URL, config, input), ErrTurnstileVerificationUnavailable)
		assert.Equal(t, 2, calls)
	})

	t.Run("retries temporary failures with one idempotency key", func(t *testing.T) {
		calls := 0
		idempotencyKeys := make([]string, 0, turnstileValidationMaxAttempts)
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			calls++
			if !assert.NoError(t, request.ParseForm()) {
				return
			}
			idempotencyKeys = append(idempotencyKeys, request.PostForm.Get("idempotency_key"))
			if calls == 1 {
				writer.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(writer, `{"success":true,"hostname":"login.example.test","action":"linuxdo_login"}`)
		}))
		defer server.Close()

		err := verifyTurnstileWithHTTPClient(context.Background(), server.Client(), server.URL, config, input)
		require.NoError(t, err)
		assert.Equal(t, turnstileValidationMaxAttempts, calls)
		require.Len(t, idempotencyKeys, turnstileValidationMaxAttempts)
		assert.NotEmpty(t, idempotencyKeys[0])
		assert.Equal(t, idempotencyKeys[0], idempotencyKeys[1])
	})

	t.Run("fails closed on transport and timeout errors", func(t *testing.T) {
		t.Run("transport", func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: turnstileRoundTripper(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("test transport failure")
			})}

			err := verifyTurnstileWithHTTPClient(context.Background(), client, turnstileSiteverifyURL, config, input)
			assert.ErrorIs(t, err, ErrTurnstileVerificationUnavailable)
			assert.Equal(t, turnstileValidationMaxAttempts, calls)
		})

		t.Run("timeout", func(t *testing.T) {
			client := &http.Client{Transport: turnstileRoundTripper(func(request *http.Request) (*http.Response, error) {
				<-request.Context().Done()
				return nil, request.Context().Err()
			})}
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()

			err := verifyTurnstileWithHTTPClient(ctx, client, turnstileSiteverifyURL, config, input)
			assert.ErrorIs(t, err, ErrTurnstileVerificationUnavailable)
		})
	})

	t.Run("rejects blank or oversized tokens before making a request", func(t *testing.T) {
		calls := 0
		client := &http.Client{Transport: turnstileRoundTripper(func(*http.Request) (*http.Response, error) {
			calls++
			return nil, errors.New("must not be called")
		})}

		err := verifyTurnstileWithHTTPClient(context.Background(), client, turnstileSiteverifyURL, config, TurnstileValidationRequest{})
		assert.ErrorIs(t, err, ErrTurnstileVerificationUnavailable)
		err = verifyTurnstileWithHTTPClient(context.Background(), client, turnstileSiteverifyURL, config, TurnstileValidationRequest{Token: strings.Repeat("x", turnstileTokenMaxBytes+1)})
		assert.ErrorIs(t, err, ErrTurnstileVerificationUnavailable)
		assert.Zero(t, calls)
	})

	t.Run("omits an invalid remote IP", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !assert.NoError(t, request.ParseForm()) {
				return
			}
			assert.Empty(t, request.PostForm.Get("remoteip"))
			writer.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(writer, `{"success":true,"hostname":"login.example.test","action":"linuxdo_login"}`)
		}))
		defer server.Close()

		err := verifyTurnstileWithHTTPClient(context.Background(), server.Client(), server.URL, config, TurnstileValidationRequest{
			Token:    input.Token,
			RemoteIP: "not-an-ip",
		})
		require.NoError(t, err)
	})
}
