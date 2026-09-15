# 公益站仅允许流式生成

本站在完成身份认证后、渠道选择及上游请求和预扣费之前拒绝非流式生成。
这是本站固定策略，无新增数据库字段或后台开关。发布包含本改动的服务版本后生效。

## 受限接口

以下 POST 接口必须在 JSON 请求体中显式设置布尔值 `"stream": true`：

- `/v1/chat/completions`
- `/v1/completions`
- `/v1/responses`（包括任务插件处理入口；`background: true` 不豁免流式要求）
- `/v1/messages`
- `/pg/chat/completions`

缺失、`false`、`null` 均返回 HTTP 400。字符串 `"true"`、数字、无效 JSON
也会被拒绝。校验不修改客户端请求体；不把非流式请求自动转换为流式请求。
JSON 重复键或大小写变体不能利用普通 DTO 与插件解析差异绕过限制。

OpenAI 兼容接口的非流式错误示例：

```json
{
  "error": {
    "type": "invalid_request_error",
    "code": "stream_required",
    "param": "stream",
    "message": "This station only supports streaming generation. Set stream=true."
  }
}
```

实际 message 会附带 request id。Claude Messages 使用 `type: error` 外层及
`error.type: invalid_request_error`，不要求客户端解析 OpenAI 的 code 字段。

Gemini `/v1/models/{model}:generateContent` 和
`/v1beta/models/{model}:generateContent` 返回 HTTP 400、`INVALID_ARGUMENT`，
提示改用 `:streamGenerateContent`。Gemini 流式由 URL 操作名决定，不能通过
给 `:generateContent` 添加 `stream: true` 绕过。

## 保持原有语义的接口

模型列表、Responses 检索及 `/v1/responses/compact`、`/v1/alpha/search`、
Embeddings、Moderations、Rerank、图片、音频、视频、Realtime，以及 Gemini
embedding/countTokens 操作不受本策略拦截；其原有可用性及其他校验仍适用。

仅支持非流式的客户端需要开启流式才能使用受限接口。被本策略拒绝的请求不调用上游，
不预扣 token 额度；有效流式请求仍受现有限流、安全和额度规则约束。
非流式请求在模型 RPM 中间件之前拒绝，本策略不代替边缘防刷或并发限制。

## 本地验证

```sh
go test ./middleware ./router -count=1
```

测试覆盖各协议拒绝及放行、真实路由注册、认证顺序、用户和令牌余额未扣减、
Playground、插件入口、gzip、磁盘请求体、413 大小限制、请求体重读和 SSE 透传。
现有暂冻和异常暂停测试使用合法流式请求，继续验证暂停账号无法访问模型。
测试不调用生产上游、不更改线上配置；生产生效需另行授权发布及线上验收。
