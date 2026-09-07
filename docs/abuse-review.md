# 安全事件脱敏片段与人工审核

## 使用方式

Root 在“安全事件”展开记录后点击“人工审核安全事件”。点击“查看脱敏片段”后才发送请求，成功提交查看审计后才返回明文。默认不加载片段；关闭弹窗、隐藏或到期后移除界面中的片段，不写 localStorage，不进入事件列表响应或 Telegram。

审核结论为“确认违规 / 疑似误判 / 证据不足”，备注必填 1–300 字，经过敏感字段清理。保存仅更新审核状态和审计，不修改安全分类、处罚次数、冻结状态或用户权限。每次修改带版本号，冲突返回 409，需刷新再提交；全部审核和查看历史按事件分页查询。

片段只是最后一条用户消息的文字，不保证它导致上游拦截。历史消息、附件、模型输出也可能是原因。不能只凭拒答次数封号。旧事件没有采集正文，无法补回。

## 采集和密钥

- 仅已有结构化安全信号的请求会执行片段提取；`off` 模式不采集。
- 支持 `/v1/responses` 字符串 input、最后一个用户 message 的文字，以及 `/v1/chat/completions`、`/pg/chat/completions`、`/v1/messages` 最后一条用户消息的文字块。
- 不回退到上一条用户消息，不采集 system/developer/assistant/tool 正文，不采集图片、文件、音频或其 URL。大于 1 MiB 的请求不解析；明确显示“不支持/正文过大/无用户文字”等原因。
- 脱敏在截断前进行；最多保存 300 个 Unicode 字符，记录来源、截断和省略非文字部分的标志。规则清理常见密钥、密码赋值、邮箱、电话号码、URL、JWT、长不透明标识和控制字符；规则无法保证识别所有自然语言个人信息，不作为通用隐私识别器。
- 设置 `ABUSE_EVIDENCE_KEY` 为独立的 32 字节随机密钥的标准 Base64 编码，通过现有主机秘密配置传入业务容器。不要复用 SESSION_SECRET，不写数据库 option、前端或 Git。**本次未在生产生成或配置密钥，也未部署应用。**
- AES-256-GCM 使用随机 nonce，账号 ID 和 Request ID 作为附加认证数据；密文存独立 `abuse_evidences` 表，事件列表不返回密文。密钥缺失/非法时只记录不可用原因，不保存明文。
- 片段接口在事件发生 7 天后立即拒绝提供正文；每小时清除过期密文。已存在的备份副本随备份保留周期过期，不能声称备份也在第 7 天即时擦除。密钥需单独安全备份；直接更换密钥会使现存 7 天内密文无法解密，应在计划轮换前处理此保留窗口。
- 30 天事件清理同时删除对应片段和审核审计。普通请求和原有计费、响应、渠道重试语义不变。捕获失败不能阻断上游已完成响应，但会显示片段不可用状态。

## 接口与表

全部接口复用 RootAuth、DisableCache；写入与读取片段的 POST 经过 SessionCookieOriginGuard 和用户关键操作限流。

- `POST /api/abuse/events/:id/excerpt`：记录一次访问后返回可用片段或不可用状态；审计或解密失败返回 503，不返回正文。响应 no-store。
- `POST /api/abuse/events/:id/review`：`{decision, note, version}`；decision 为 confirmed、suspected_false_positive、insufficient_evidence。
- `GET /api/abuse/events/:id/history?p=1&page_size=20`：分页访问/审核历史。
- 事件列表增加 evidence_status、evidence_expires_at、review_status、review_version、reviewed_by、reviewed_at；支持 review_status 筛选。
- 安全设置返回 evidence_capture_ready 布尔值，不返回密钥。

主数据库新表 `abuse_evidences`、`abuse_review_audits`；`abuse_events` 新增元数据字段。独立日志数据库不用于此功能，也不写入片段。

## 发布与回滚准备

1. 在批准发布后，备份主数据库，按现有 .17 之后的 Oracle 原生 ARM64 流程构建、扫描和部署；不要使用已经停用的 GHCR 路径。
2. 将独立密钥安全注入现有运行配置，核对后台显示采集已就绪。禁止打印密钥或完整请求。
3. 用专用测试账号和模拟上游安全信号验证采集、脱敏、查看审计、审核、到期和权限；不要向真实上游发送违规请求。
4. 回滚旧应用前可移除采集密钥以停止新增片段，再回滚镜像。保留新增表和已有审计；旧版不提供查看/审核或 7 天密文清理，需要在回滚期间单独维持过期清理或在批准后清除剩余密文，不能声称旧版保持本功能。
