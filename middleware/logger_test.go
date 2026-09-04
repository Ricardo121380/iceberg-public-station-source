package middleware

import "testing"

func TestAccessLogPathOmitsQuery(t *testing.T) {
	tests := map[string]string{
		"/api/status": "/api/status",
		"/api/oauth/linuxdo?code=secret&state=secret-state": "/api/oauth/linuxdo",
		"/api/search?q=one%20two":                           "/api/search",
	}

	for rawPath, expected := range tests {
		if actual := accessLogPath(rawPath); actual != expected {
			t.Fatalf("accessLogPath(%q) = %q, want %q", rawPath, actual, expected)
		}
	}
}
