---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-03
title: 重构持久登录保护与 Session 续期模型
status: ready
planning_depth: deep
planning_depth_reason: 改变认证协议、CSRF 安全语义、双方言并发状态和 Session 数据迁移
ready: true
risk: critical
blocked_by: []
contract_ids: [AC-008, AC-009, AC-010, AC-011, AC-012, AC-013]
owner: unassigned
expected_changes: ["<Path>go-admin-plus/internal/modules/iam/account/password_budget.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/service.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/*_test.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/protection/**</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0040-session-protection/**</Path>", "<Path>go-admin-plus/test/iam/session/**</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/modules/iam/account/password_budget.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/service.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/*_test.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/protection/**</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0040-session-protection/**</Path>", "<Path>go-admin-plus/test/iam/session/**</Path>"]
read_only_paths: ["<Path>contracts/openapi/modules/iam-session.yaml</Path>", "<Path>go-admin-plus/internal/modules/iam/session/http.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/transport/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-03: 重构持久登录保护与 Session 续期模型

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/03-login-session-hardening.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-03.md</Path>`

## 1. 战略与来源

- **目标：** 让登录防护跨重启/实例成立，并让认证读取、稳定 CSRF 和显式续期消除跨标签互相失效。
- **可观察产出：** 持久账号/来源双桶给出模糊限流反馈；认证 GET 零写入；heartbeat/renew 延长 idle 而不延长 absolute expiry。
- **来源：** `US-004~006`、`AC-008~013`、`ADR-003`、`ADR-006`、`ADR-011`。
- **当前事实：** `<Path>go-admin-plus/internal/modules/iam/session/service.go</Path>` 在所有认证请求旋转 Session/CSRF；password budget 仅限制进程内并发。
- **Planning Depth 原因：** 认证、CSRF、数据库竞态和迁移任一错误都可造成提权或大面积登出。

## 2. 决策状态

### 已锁定决策

- 账号桶与可信来源桶由正式数据库原子维护，两者都必须通过且保护不可关闭。
- 不存在账号执行 dummy Argon2；普通失败统一，限流仅返回粗粒度 Retry-After。
- Session 家族内 CSRF 稳定；GET 只读，heartbeat/业务写/提前 renew 更新活跃期。

### 已采用的低影响假设

- 限流阈值和时间窗采用类型化安全范围内的保守默认值；精确数值不进入公共 wire contract。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 0040 migration、双桶、可信来源 port、Session family、heartbeat/renew use case | 现有不透明 Cookie、Argon2、数据库 Session 实时授权 | OpenAPI/HTTP adapter（T-08）、跨标签客户端（T-13）、Redis |

## 4. 要构建什么

登录请求先在数据库中原子消费账号和来源预算，再执行真实或 dummy 密码工作。成功建立带稳定 CSRF 的 Session 家族。普通认证读取只验证当前数据库事实；显式 heartbeat、renew 和受保护业务写更新 idle。退出、撤销、禁用、密码变更与 absolute expiry 在下一请求即时生效。

## 5. 实现契约

- **入口或接缝：** Login、Current、Heartbeat、Renew、Logout 与 Revoke use cases。
- **输入与输出：** 规范化账号标识、可信来源、Cookie/CSRF；输出稳定认证结果、粗粒度 Retry-After 或统一 Problem 类别。
- **公共接口变化：** 模块用例先固定；wire contract 由 T-08 独占发布。
- **不变量：** GET 不写 Session；renew 不换 CSRF；absolute expiry 不延长；数据库是真实认证源。
- **状态或数据流：** active -> renewed -> revoked/idle-expired/absolute-expired；限流桶跨进程持久。
- **错误与失败行为：** 账号存在与否公开响应等价；双桶任一拒绝即限流；竞态不得超发预算。
- **兼容要求：** 旧滚动 CSRF/每请求续期无兼容窗口；切换时旧 Session 统一撤销。
- **安全与隐私要求：** token、CSRF、来源原值和账号探测信息不得记录；代理信任显式配置。

## 6. 执行路线

1. 建立 GET 零写入、双桶竞争、账号枚举等价和 renew 竞态红灯测试。
2. 增加 0040 migration、限流 repository 与受验证 policy。
3. 重构 Session 家族、Current/Heartbeat/Renew 与业务写活跃更新。
4. 统一失败、Retry-After 分桶和脱敏审计 ports。
5. 运行双方言、race、时间控制与 Session 回归。

## 7. 路径访问契约

- **预计修改点/可写范围：** frontmatter 列出的 Session service、protection、0040 migration 与测试。
- **只读上下文：** canonical OpenAPI、HTTP/生成 transport 和前端 adapter。
- **共享路径：** 无；公共合同由 T-08，客户端由 T-13。
- **保留或不动：** 不引入 Redis/JWT，不改其他 IAM administration 文件。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | Session use cases | 定向双方言测试 | 登录、只读 Current、heartbeat/renew 成立 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-03.md</Path>` |
| 失败路径 | 竞争/枚举/超时 | 并发、重启、存在/不存在账号、idle/absolute 边界 | 不超发、不枚举、即时失效 | 同上 |
| 回归 | IAM Session | `cd go-admin-plus && go test ./internal/modules/iam/session/... ./test/iam/session/... -race -count=1` | Session 回归通过 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行单元、双方言集成、race、vet；真实 PG Gate 在 T-18。
- **E2E disposition：** not-required：本 Ticket 的稳定接缝是模块 use case；HTTP、浏览器双标签和原生 E2E 分别归 T-08、T-19、T-20。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；本 Ticket 不在 source 声明 E2E。
- **Integration evidence：** implementation/source commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** 0040 schema -> repository/use case -> T-08 wire -> T-13 client；部署时统一撤销旧 Session。
- **兼容窗口：** 无双协议；旧 Session 与滚动 CSRF 在原子切换点失效。
- **监控信号：** 登录结果分类、限流桶聚合、heartbeat/renew 冲突、Session DB 写率。
- **回滚或前向恢复：** forward-only；协议问题前向修复，必要时撤销全部 Session 并要求重新登录。
- **不可逆操作与批准点：** 旧 Session 撤销前需 Goal Plan Gate 明确批准。
- **收缩条件：** 旧 rotate-on-read 调用为零，GET 写入测试稳定为零。

## 10. 验收标准

- [ ] `AC-008~013` 在双方言、并发和时间边界测试中成立。
- [ ] 认证 GET 零写入，CSRF 家族稳定，外部错误不枚举账号。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-03.md</Path>`。
- [ ] 路径、commit、direct-parent/candidate、父分支 result 与 E2E disposition 完整。

