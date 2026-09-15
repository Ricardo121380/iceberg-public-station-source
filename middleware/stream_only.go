package middleware

import (
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// StreamingOnly rejects non-streaming text generation before plugin dispatch,
// channel selection and billing. Other API operations keep their own semantics.
func StreamingOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		gemini := (strings.HasPrefix(path, "/v1/models/") || strings.HasPrefix(path, "/v1beta/models/")) && strings.HasSuffix(path, ":generateContent")
		switch path {
		case "/v1/chat/completions", "/v1/completions", "/v1/responses", "/v1/messages", "/pg/chat/completions":
		default:
			if !gemini {
				c.Next()
				return
			}
		}

		status := http.StatusBadRequest
		code := "stream_required"
		message := "This station only supports streaming generation. Set stream=true."
		if gemini {
			message = "This station only supports streaming generation. Use :streamGenerateContent instead of :generateContent."
		} else {
			var request struct {
				Stream *bool `json:"stream,omitempty"`
			}
			storage, err := common.GetBodyStorage(c)
			var body []byte
			if err == nil {
				body, err = storage.Bytes()
			}
			if err == nil {
				err = common.Unmarshal(body, &request)
			}
			if err == nil {
				_, err = storage.Seek(0, io.SeekStart)
				c.Request.Body = io.NopCloser(storage)
			}
			if err != nil {
				code = "invalid_request"
				message = "Invalid request body; expected JSON with stream=true."
				if common.IsRequestBodyTooLargeError(err) {
					status = http.StatusRequestEntityTooLarge
					code = "request_body_too_large"
					message = "Request body is too large."
				}
			} else if request.Stream != nil && *request.Stream && gjson.GetBytes(body, "stream").Type == gjson.True {
				// Plugins use the first exact JSON key, while DTO decoding uses
				// the last case-insensitive match. Both must agree on streaming.
				c.Next()
				return
			}
		}

		message = common.MessageWithRequestId(message, c.GetString(common.RequestIdKey))
		switch {
		case path == "/v1/messages":
			c.AbortWithStatusJSON(status, gin.H{"type": "error", "error": gin.H{"type": "invalid_request_error", "message": message}})
		case gemini:
			c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": status, "status": "INVALID_ARGUMENT", "message": message}})
		default:
			c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"type": "invalid_request_error", "code": code, "param": "stream", "message": message}})
		}
	}
}
