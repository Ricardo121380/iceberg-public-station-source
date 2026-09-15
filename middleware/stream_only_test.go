package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamingOnlyGenerationProtocols(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, path := range []string{
		"/v1/chat/completions", "/v1/completions", "/v1/responses", "/v1/messages", "/pg/chat/completions",
	} {
		for _, body := range []string{
			`{"model":"test"}`, `{"stream":false}`, `{"stream":null}`, `{"stream":"true"}`, `{"stream":1}`,
			`{"background":true}`, `{"Stream":true}`, `{"stream":true,"Stream":false}`,
			`{"stream":false,"Stream":true}`, `{"stream":true,"stream":false}`, `{"stream":false,"stream":true}`,
			`{"metadata":{"stream":true}}`, `{"stream":true`, `null`, `[]`,
		} {
			t.Run(path+"/"+body, func(t *testing.T) {
				called := false
				router := gin.New()
				router.Use(BodyStorageCleanup(), StreamingOnly())
				router.POST(path, func(c *gin.Context) { called = true })
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				assert.False(t, called, "must stop before upstream dispatch and billing")
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				var payload map[string]any
				require.NoError(t, common.Unmarshal(recorder.Body.Bytes(), &payload))
				detail, ok := payload["error"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "invalid_request_error", detail["type"])
				assert.Contains(t, detail["message"], "stream=true")
				if path == "/v1/messages" {
					assert.Equal(t, "error", payload["type"])
				} else {
					assert.Equal(t, "stream", detail["param"])
				}
			})
		}
	}
}

func TestStreamingOnlyPreservesStreamingBodyAndSSE(t *testing.T) {
	for _, path := range []string{"/v1/chat/completions", "/v1/completions", "/v1/responses", "/v1/messages", "/pg/chat/completions"} {
		for _, compressed := range []bool{false, true} {
			t.Run(path+"/gzip="+strconv.FormatBool(compressed), func(t *testing.T) {
				body := `{"model":"test","stream":true,"messages":[{"role":"user","content":"hello"}],"input":"hello"}`
				router := gin.New()
				router.Use(DecompressRequestMiddleware(), BodyStorageCleanup(), StreamingOnly())
				router.POST(path, func(c *gin.Context) {
					got, err := io.ReadAll(c.Request.Body)
					require.NoError(t, err)
					assert.Equal(t, body, string(got))
					var parsed map[string]any
					require.NoError(t, common.UnmarshalBodyReusable(c, &parsed))
					assert.Equal(t, true, parsed["stream"])
					c.Data(http.StatusOK, "text/event-stream", []byte("data: [DONE]\n\n"))
				})
				var input io.Reader = strings.NewReader(body)
				if compressed {
					var buf bytes.Buffer
					writer := gzip.NewWriter(&buf)
					_, err := writer.Write([]byte(body))
					require.NoError(t, err)
					require.NoError(t, writer.Close())
					input = &buf
				}
				req := httptest.NewRequest(http.MethodPost, path, input)
				req.Header.Set("Content-Type", "application/json")
				if compressed {
					req.Header.Set("Content-Encoding", "gzip")
				}
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				assert.Equal(t, http.StatusOK, recorder.Code)
				assert.Equal(t, "text/event-stream", recorder.Header().Get("Content-Type"))
				assert.Equal(t, "data: [DONE]\n\n", recorder.Body.String())
			})
		}
	}
}

func TestStreamingOnlyGeminiAndUnrelatedOperations(t *testing.T) {
	for _, tc := range []struct {
		method string
		path   string
		status int
	}{
		{"POST", "/v1/models/gemini:generateContent", 400},
		{"POST", "/v1beta/models/gemini:generateContent", 400},
		{"POST", "/v1/models/gemini:streamGenerateContent", 204},
		{"POST", "/v1beta/models/gemini:streamGenerateContent", 204},
		{"POST", "/v1beta/models/gemini:embedContent", 204},
		{"POST", "/v1beta/models/gemini:batchEmbedContents", 204},
		{"POST", "/v1beta/models/gemini:countTokens", 204},
		{"POST", "/v1/responses/compact", 204},
		{"POST", "/v1/embeddings", 204},
		{"POST", "/v1/images/generations", 204},
		{"POST", "/v1/audio/speech", 204},
		{"POST", "/v1/alpha/search", 204},
		{"POST", "/v1/moderations", 204},
		{"POST", "/v1/rerank", 204},
		{"GET", "/v1/responses/resp_test", 204},
		{"GET", "/v1/models", 204},
		{"GET", "/v1/realtime", 204},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			router := gin.New()
			router.Use(StreamingOnly())
			router.Handle(tc.method, tc.path, func(c *gin.Context) { c.Status(http.StatusNoContent) })
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{"stream":true}`)))
			assert.Equal(t, tc.status, recorder.Code)
			if tc.status == 400 {
				assert.Contains(t, recorder.Body.String(), ":streamGenerateContent")
				assert.Contains(t, recorder.Body.String(), `"status":"INVALID_ARGUMENT"`)
			}
		})
	}
}

func TestStreamingOnlyPreservesBodySizeLimit(t *testing.T) {
	original := constant.MaxRequestBodyMB
	constant.MaxRequestBodyMB = 1
	t.Cleanup(func() { constant.MaxRequestBodyMB = original })
	router := gin.New()
	router.Use(DecompressRequestMiddleware(), BodyStorageCleanup(), StreamingOnly())
	router.POST("/v1/responses", func(c *gin.Context) { t.Error("oversized request reached relay") })
	body := `{"stream":true,"input":"` + strings.Repeat("x", 1<<20) + `"}`
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(body)))
	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "request_body_too_large")
}

func TestStreamingOnlyDiskBackedBody(t *testing.T) {
	original := common.GetDiskCacheConfig()
	common.SetDiskCacheConfig(common.DiskCacheConfig{Enabled: true, ThresholdMB: 0, MaxSizeMB: 16, Path: t.TempDir()})
	t.Cleanup(func() { common.SetDiskCacheConfig(original) })
	for _, tc := range []struct {
		body   string
		status int
	}{
		{`{"stream":true,"input":"hello"}`, 204},
		{`{"stream":true} {"stream":false}`, 400},
	} {
		t.Run(tc.body, func(t *testing.T) {
			router := gin.New()
			router.Use(BodyStorageCleanup(), StreamingOnly())
			router.POST("/v1/responses", func(c *gin.Context) {
				storage, err := common.GetBodyStorage(c)
				require.NoError(t, err)
				require.True(t, storage.IsDisk())
				body, err := io.ReadAll(c.Request.Body)
				require.NoError(t, err)
				assert.Equal(t, tc.body, string(body))
				c.Status(http.StatusNoContent)
			})
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(tc.body)))
			assert.Equal(t, tc.status, recorder.Code, recorder.Body.String())
			entries, err := os.ReadDir(common.GetDiskCacheDir())
			require.NoError(t, err)
			assert.Empty(t, entries, "request storage must be cleaned on success and rejection")
		})
	}
}
