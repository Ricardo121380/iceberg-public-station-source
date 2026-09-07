# 脱敏输入片段与人工审核验证

基线：`350b229`，包含生产 .17 源码 `8fedcf6` 及发布说明更新。独立分支 `codex/abuse-review`，未改动其他工作树。

## 已通过

- `go test ./common ./model ./service ./router -run 'TestAbuse|TestSafetyEvent|TestNormalRequests' -count=1`：38 项测试/子测试通过；本地 MySQL/PostgreSQL DSN 未配置的测试显式跳过，真实三库验证另列。
- `go build -o /tmp/iceberg-abuse-review-app .`：通过。
- `bun run test src/features/abuse/__tests__/review.test.tsx src/features/abuse/__tests__/management.test.tsx`：10 项通过。
- `bun run typecheck`、受影响 TS/TSX 文件 oxlint、oxfmt：通过。
- `bun run build`：通过。未更改依赖锁文件，复用同基线工作树的前端依赖，仅作为本地开发环境。
- SQLite/服务测试覆盖：AES-GCM 加密及账号/请求绑定、缺失密钥失败关闭、脱敏后截断、最后用户文字选择、附件省略、正常请求无正文、off 不采集、1 MiB 上限、原始 body cursor 不变、7 天到期及清理、审计失败拒绝返回、Root 权限隔离、审核并发版本冲突、不修改处罚语义。
- 前端交互覆盖：明确点击才读取、关闭弹窗清除片段、过期/服务失败显示、备注校验、审核保存及冲突提示、不发账号处置请求。
- ego-browser 本地模拟预览：中文片段与表单可读；390px 屏幕无横向溢出（scrollWidth=390，dialogWidth=358），内容可纵向滚动至提交按钮。所有预览数据为模拟内容，未读取真实用户正文。
- Go 模块 `relaykit` 未改动，不涉及其独立构建接口。

## 真实数据库验证

在用户明确允许的 Oracle 隔离容器执行；生产数据库未访问或修改。测试命令在 `../abuse-review-matrix.sh`，使用独立测试容器、loopback 端口及一次性测试数据。

首次测试发现 MySQL 临时初始化实例不能以 mysqladmin ping 代表可用，已改用 TCP 认证查询等待。
PostgreSQL 在测试中删除旧表并重建后的同连接查询出现 cached plan result type 冲突；已将“全新安装”阶段改为新连接，与真实新进程启动一致。旧版升级、重复迁移阶段本身通过。

最终复验全部通过（2026-09-07）：

| 引擎 | 实际版本 | 结果 |
|---|---|---|
| SQLite | 3.50.4 | PASS |
| MySQL | 8.0.46 | PASS |
| PostgreSQL | 17.11（Debian 17.11-1.pgdg12+2，ARM64） | PASS |

命令：`CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go test -c ./model`，上传测试二进制后执行 `bash /tmp/iceberg-abuse-review-matrix.sh`；测试二进制参数 `-test.run '^TestAbuseReviewMigrationMatrix$' -test.v`。

每个引擎覆盖：.17 的原始事件 schema 与真实旧记录、保留复合唯一约束、升级与重复 AutoMigrate、审核与审计数据保留、版本冲突、到期密文删除、30 天关联清理，以及空库重复初始化。使用事务验证审核不改变 action/actionable/notify。此路径仅使用主数据库，不影响单独配置的日志数据库。


## 未执行

- 未部署生产镜像、未生成或设置生产 `ABUSE_EVIDENCE_KEY`。
- 未向 Telegram 发送测试消息，未新增正文通知，也未调用真实模型。
- 未将任何用户标为违规或冻结账号。旧事件无法补回正文。
- 此验证不证明正则脱敏可以识别任意自然语言隐私，也不证明所选片段就是拦截原因。
