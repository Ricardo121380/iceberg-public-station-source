# 上游安全事件与账号暂冻

## 当前发布边界

默认 observe；站主至少观察 7 天后自行切换 enforce。普通 refusal/content_filter
仅为分类信息不足的观察事件，不计入自动处罚。2026-09-05 核查的生产渠道是
NewAPI 类型 60，模型 gpt-5.6-sol、gpt-6-astra；最近 7 天现存错误/reject_reason
聚合没有可用于验证色情/破限输入类别的样例。故生产 verified rule catalog 为空。
模拟上游只证明代码链路，不证明真实上游分类准确度，不可据此启用规则。

## 操作

Root：系统设置 → 安全与限制 → 安全风控。可修改模式、10 分钟/24 小时阈值、
暂停时长，查询事件/冻结账号并填写原因解冻。默认 3/10 分钟、8/24 小时、60 分钟。
普通用户只能查看自己的暂停状态；仍能登录。所有经过 Distribute 的模型入口，
包括 Responses、Chat、Claude-compatible、Playground 和任务入口，共用冻结检查。
已冻结账号的新 Token、重新启用 Token 不绕过；已发往上游的在途请求不取消。

切换模式开启新 generation，历史事件不追罚。暂停创建时 round 增加，直到到期
前的请求不能进入新一轮；到期后新请求采用新 round，人工解冻再推进 round。
切回 observe/off 不解除已有冻结；解除须通过 Root 操作或等待到期。
管理员/Root 只观察、不自动冻结。永封使用现有用户管理入口。

## 接口

均使用现有 success/data 包装；错误不返回数据库详情。Root 的写接口检查同源。

- GET/PUT /api/abuse/settings：mode、limit_10m、limit_24h、freeze_minutes、enabled_rules。
  规则验证状态只由代码目录及真实证据确定，接口不能将任意规则标记 verified。
- GET /api/abuse/events：p/page_size（最多 100），user_id、category、rule_id、action、start/end（Unix 秒）。
- GET /api/abuse/users：分页列出仍处于暂停期限的账号。
- POST /api/abuse/users/:id/unfreeze：reason，1–300 字；记录操作人。
- GET /api/user/self/abuse：suspended、blocked_until，不返回证据。

暂停调用返回 403，error.code=account_temporarily_suspended，blocked_until 为 Unix 秒。
状态数据库不可用返回 503，safety_state_unavailable；Redis 不可用停止新处罚、保留事件。

## 证据与存储

主库 abuse_policies、abuse_states、abuse_events；与 LOG_SQL_DSN 分离，事务始终在主库。
每账号/Request ID 唯一，最多保存 32 个结构化尝试标记；不保存原始响应、提示词、Key。
目前所有上游描述都未验证为不回显用户内容，因此使用固定安全摘要，长度小于 300 字。
新增经过审核的上游摘要字段前，必须增加不泄露凭据/提示词的样例测试。

Redis ZSET 使用账号/generation/round 分组。每次可处置事件在数据库决策锁内从
当前 24 小时持久事件重建派生索引，Lua 原子去重、剪裁、统计，TTL 25 小时。
针对当前小站流量，选择一致性优先的全局策略行写锁；可处置事件量增长后再优化。
请求完成后写证据失败以 abuse_control: persistence_failed 报错，主机监控告警。

每 24 小时分批删除 30 天之前的事件，不删除冻结状态；备份随原保留周期过期。

## 通知

仓库外层 ops/monitoring/new-api-public-station-monitor.sh 读取主库待通知事件，
每分钟最多 20 条。Telegram 密钥继续只在主机 0600 文件，不放入应用容器。
发送成功后才确认数据库 outbox；失败保留待通知事件，复用传输失败队列。
Telegram 无事务确认协议：发送成功但确认数据库前崩溃可能重复发送同一 event ID，
不承诺跨外部服务的严格 exactly-once。平稳运行按事件去重。
同账号/规则/动作观察和计数故障通知最多每小时一次；冻结及解冻即时进入 outbox。

## 协议覆盖

| 链路 | 捕获位置 | 可处置 |
|---|---|---|
| Chat JSON/SSE | 原始 choices.finish_reason、refusal 字段 | 通用信号仅观察 |
| Responses JSON/SSE | incomplete_details、refusal 事件及 output 项 | 通用信号仅观察 |
| Chat 经 Responses 转换 | 转换前原始 JSON/SSE（含缓冲转换） | 通用信号仅观察 |
| Claude-compatible JSON/SSE | stop_reason=refusal | 仅观察 |
| 非 2xx | 标准化前 error.code/type | 通用安全码仅观察 |
| 其他供应商、图片/音频/异步任务内容 | 冻结检查覆盖，未专门适配分类 | 不承诺内容识别 |

新增规则需把渠道、协议、字段、值、类别、版本和真实脱敏证据路径加入
service/abuse_control.go 的只读规则目录，并通过样例回归。禁止以消息关键词替代。

## 验证与恢复

本地：GOWORK=off make test；go test -race ./model ./service -run '^TestAbuse'。
前端：bun run typecheck；bun run test；bun run build；变更文件 oxlint/oxfmt。
三数据库：ABUSE_TEST_MYSQL_DSN / ABUSE_TEST_POSTGRES_DSN 指向独立空 abuse_test 库，
go test ./model -run '^TestAbuseDatabaseMatrix$' -v。测试拒绝含既有 users 表的库。
发布工作流自带 PostgreSQL 16、MySQL 8 服务和该测试；SQLite 始终执行。

快速停用新处罚：后台改 observe。数据库备份后按现有不可变镜像流程发布。
回滚旧版本会失去冻结检查，即使 abuse_states 表仍在也不会执行限制；回滚前
必须核对在冻账号并决定是否通过旧用户禁用流程保持阻断。保留新表，不反向删表。
