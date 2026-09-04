package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccessLogPathOmitsQuery(t *testing.T) {
	tests := map[string]string{
		"/api/status": "/api/status",
		"/api/oauth/linuxdo?code=secret&state=secret-state": "/api/oauth/linuxdo",
		"/api/search?q=one%20two":                           "/api/search",
	}

	for rawPath, expected := range tests {
		require.Equal(t, expected, accessLogPath(rawPath))
	}
}
