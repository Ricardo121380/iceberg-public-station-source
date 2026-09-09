# 异常提醒控制与独立安全防护入口

本分支 `codex/anomaly-console` 基于 v1.0.0.18 开发，发布目标 v1.0.0.19。

## 页面与权限

- `/anomalies`：异常提醒控制。管理员及 Root 可查看、筛选和确认事件；仅 Root 可修改提醒规则。
- `/safety`：违规行为安全防护。保留原有 Root 专属的配置、敏感证据查看、人工复核和解冻权限，不向普通管理员开放。
- 原 `/system-settings/security/abuse` 跳转至 `/safety`；系统设置侧边栏不再列出该页面。
- 普通用户看不到入口，直接访问页面及异常 API 会被拒绝。所有写接口保留 SessionCookieOriginGuard、用户操作限速；读取接口 no-store。

## 本版功能与边界

新增功能为**站内异常观察提醒**，不是自动封号系统。默认启用采集：固定对齐 5 分钟窗口，结构错误 10 次、本站限流 100 次、上游故障 5 次触发待关注状态。阈值可由 Root 调整。额度不足、模型未配置及其他参数错误只记录，不因这些错误产生处罚。

按用户 ID、渠道 ID、模型、失败类别和规则版本聚合。不是按 Schema 指纹聚合，不能把汇总事件当作同一个 Schema 的确证。所有 Token 的请求会按所属用户聚合，但 Token 信息不写入异常记录。

覆盖 `/v1`、`/pg` 的已认证用户失败 POST 请求，每次只统计最终 HTTP 失败一次；内部重试成功不产生失败事件。本站限流与上游限流分别分类；早于模型选择的拒绝可能没有模型和渠道信息，界面明确展示未知。不采集认证失败、GET/WebSocket、其他协议路由及 HTTP 200 后的流式中断。不是全站计费或成功率统计。

本版不改变现有限流，不自动处罚，不回填历史日志。v1.0.0.19 加入 Telegram 提醒及人工确认临时暂停，详见末尾上线补充；Schema 自动冷却和渐进自动处罚未开启。

## 数据与运行

使用已有 Redis 客户端，共享 namespace `anomaly:v1:`；不增加 SQL 表、不运行 SQL 迁移。规则持久化为无 TTL 的独立 Redis key；事件保留 7 天且至多 2,000 组；处理审计保留 7 天内最近 200 条。保留的是容量受限的运行记录，不作为不可变合规审计。

Redis 的持久性取决于现有部署配置；清空 Redis 将删除事件、审计和规则，恢复默认观察规则。不要把这些 key 当作可任意清理的普通缓存。

Lua 脚本保证计数和确认写入原子性。设置使用版本冲突保护；确认使用已见计数保护，记录更新后返回 409，避免误确认新增失败。确认后出现新失败会重新变为待关注，不会无限延长任何封禁，因为本版不执行封禁。

采集是尽力记录，单请求 Redis 操作限时 150ms，Redis 异常不改变原请求响应；数据故障期间可能漏记。后台 Redis 不可用时返回 503，不展示伪造的空成功状态。未记录原始错误正文、提示词、密钥、IP 或 Token 内容；模型标签限长并移除控制字符。

## API

| 接口 | 权限 | 用途 |
|---|---|---|
| GET `/api/anomalies/events` | Admin | 分页筛选，参数 p、user_id、kind、state |
| GET `/api/anomalies/settings` | Admin | 读取采集开关、窗口、提醒阈值、版本 |
| PUT `/api/anomalies/settings` | Root | 校验范围、版本后保存并审计 |
| POST `/api/anomalies/events/:id/acknowledge` | Admin | 提交 count，确认当前计数并记录操作人 |
| GET `/api/anomalies/audit` | Admin | 获取保留期内处理记录 |

## 本地验证

```sh
go test ./service ./middleware ./controller ./router -run 'Test(Anomaly|Abuse)' -count=1
cd web
bun run typecheck
bun run test src/features/anomalies/__tests__ src/hooks/__tests__/anomaly-navigation.test.tsx src/features/abuse/__tests__
bun run build
```

后端测试使用项目现有 miniredis（执行 Lua）与现有 SQLite 路由测试夹具。未新增或修改关系数据库行为，未执行三数据库迁移矩阵；未连接生产 Redis 做写入验证。部署前应在隔离 Redis 环境验证该部署版本的脚本执行和持久化，再进行经授权的发布验证。

页面预览：源码根目录运行 `python3 docs/verification/anomaly-preview.py`，打开 `http://127.0.0.1:3340/anomalies`。这是仅绑定回环地址的**合成数据 fixture**，展示的用户和计数不代表实时生产数据，不能部署此脚本或将它当作生产后端。

## 本次验证结果

- Go：`go test ./service ./middleware ./controller ./router -run 'Test(Anomaly|Abuse)' -count=1` 通过（controller 编译通过，无匹配测试）。覆盖原安全防护与新提醒路由权限、计数并发、阈值、关闭采集、容量淘汰、过期、确认冲突、再次提醒、存储故障及响应保持。
- 前端：4 个测试文件、22 项测试通过；TypeScript 类型检查通过；变更 TS/TSX 文件 oxlint 无错误；生产构建通过。
- Impeccable 检测无机械问题。ego-browser 使用本地合成数据，验证两个入口、旧路径跳转与桌面/手机页面。独立视觉审查子任务被平台拦截，未取得其结论。
- CodeGraph 已在独立工作树初始化并同步，AnomalyConsole 符号查询成功。
- 没有部署、修改线上配置、暂停用户、发送通知或调用外部模型。

## v1.0.0.19 上线补充

增加独立主机TG通知进程，通过专用随机秘密访问internal/anomaly-bot接口；只允许配置的TG私聊用户。首次按钮请求发送二次确认，确认有效期5分钟；原告警按钮15分钟。按事件限制一次暂停，重复operation_id返回原结果，不延长。暂停30分钟，保留登录、覆盖全部API Token和playground，允许TG提前解除。仅结构错误/本站限流阈值事件提供暂停，管理员及非正常状态用户受保护，上游故障不允许此操作。

暂停状态在Redis，和原有AbuseState独立；清空Redis会解除TG暂停。操作记录保留TG操作人，不混用站点管理员ID。站内个人暂停提示同时显示TG暂停。监控器持续检查机器人服务。机器人每30秒查询一次，3秒轮询按钮，异常传输退避，不将密钥URL写入日志。

补齐独立responses协议路由的采集。后台和日志不回填旧数据，发布后开始采集。自动封号仍未开启，所有暂停必须人工确认。
