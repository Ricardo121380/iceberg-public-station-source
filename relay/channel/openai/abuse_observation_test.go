package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAbuseResponsesWireObservation(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() { constant.StreamingTimeout = oldTimeout })
	for _, stream := range []bool{false, true} {
		t.Run(map[bool]string{false: "json", true: "sse"}[stream], func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			require.NoError(t, err)
			require.NoError(t, model.MigrateAbuse(db))
			require.NoError(t, db.AutoMigrate(&model.User{}))
			require.NoError(t, db.Create(&model.User{Id: 1, Username: "wire", Role: 1, Status: 1}).Error)
			prev := model.DB
			model.DB = db
			t.Cleanup(func() { model.DB = prev; sqlDB, _ := db.DB(); _ = sqlDB.Close() })
			payload := `{"id":"resp_fixture","status":"completed","output":[{"type":"message","content":[{"type":"refusal","refusal":"fixture secret: sk-private"}]}],"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}}`
			if stream {
				payload = "data: {\"type\":\"response.refusal.delta\",\"delta\":\"fixture secret: sk-private\"}\n\ndata: {\"type\":\"response.completed\",\"response\":" + payload + "}\n\ndata: [DONE]\n\n"
			}
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if stream {
					w.Header().Set("Content-Type", "text/event-stream")
				} else {
					w.Header().Set("Content-Type", "application/json")
				}
				_, _ = io.WriteString(w, payload)
			}))
			defer upstream.Close()
			resp, err := http.Get(upstream.URL)
			require.NoError(t, err)
			out := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(out)
			c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
			c.Set("id", 1)
			c.Set("token_id", 1)
			c.Set("channel_id", 1)
			c.Set("channel_type", constant.ChannelTypeNewAPI)
			c.Set(common.RequestIdKey, "wire-fixture")
			finish, err := service.BeginSafetyObservation(c)
			require.NoError(t, err)
			info := &relaycommon.RelayInfo{OriginModelName: "fixture", IsStream: stream, DisablePing: true, ChannelMeta: &relaycommon.ChannelMeta{UpstreamModelName: "fixture"}}
			if stream {
				_, apiErr := OaiResponsesStreamHandler(c, info, resp)
				require.Nil(t, apiErr)
			} else {
				_, apiErr := OaiResponsesHandler(c, info, resp)
				require.Nil(t, apiErr)
				assert.JSONEq(t, payload, out.Body.String())
			}
			finish()
			var events []model.AbuseEvent
			require.NoError(t, db.Find(&events).Error)
			require.Len(t, events, 1)
			assert.False(t, events[0].Actionable)
			assert.NotContains(t, events[0].Summary, "sk-private")
			assert.NotContains(t, events[0].Attempts, "sk-private")
			assert.True(t, strings.Contains(out.Body.String(), "refusal"))
		})
	}
}
