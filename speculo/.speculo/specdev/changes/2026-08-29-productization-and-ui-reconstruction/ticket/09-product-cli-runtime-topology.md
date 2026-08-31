---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-09
title: 统一产品 CLI、运行拓扑与 Profile Migration
status: done
planning_depth: deep
planning_depth_reason: 修改产品组合根、运行角色、迁移所有权、共享命令面和生产 readiness，事故半径覆盖所有 profile
ready: true
risk: critical
blocked_by: [T-07, T-08]
contract_ids: [AC-029, AC-030, AC-031, AC-032, AC-033]
owner: codex-root
expected_changes: ["<Path>go-admin-plus/cmd/go-admin-plus/**</Path>", "<Path>go-admin-plus/cmd/migrate/**</Path>", "<Path>go-admin-plus/cmd/config-check/**</Path>", "<Path>go-admin-plus/internal/app/product/runtime.go</Path>", "<Path>go-admin-plus/internal/app/product/server.go</Path>", "<Path>go-admin-plus/internal/app/product/workers.go</Path>", "<Path>go-admin-plus/internal/app/product/migration.go</Path>", "<Path>go-admin-plus/internal/app/product/registry.go</Path>", "<Path>go-admin-plus/internal/app/product/*_test.go</Path>", "<Path>go-admin-plus/internal/host/server/**</Path>", "<Path>go-admin-plus/internal/application/architecture_test.go</Path>", "<Path>Taskfile.yml</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>scripts/quality/architecture-check.mjs</Path>", "<Path>deploy/compose/compose.yml</Path>", "<Path>.github/workflows/ci.yml</Path>", "<Path>.github/workflows/codeql.yml</Path>"]
writable_paths: ["<Path>go-admin-plus/cmd/go-admin-plus/**</Path>", "<Path>go-admin-plus/cmd/migrate/**</Path>", "<Path>go-admin-plus/cmd/config-check/**</Path>", "<Path>go-admin-plus/internal/app/product/runtime.go</Path>", "<Path>go-admin-plus/internal/app/product/server.go</Path>", "<Path>go-admin-plus/internal/app/product/workers.go</Path>", "<Path>go-admin-plus/internal/app/product/migration.go</Path>", "<Path>go-admin-plus/internal/app/product/registry.go</Path>", "<Path>go-admin-plus/internal/app/product/*_test.go</Path>", "<Path>go-admin-plus/internal/host/server/**</Path>", "<Path>go-admin-plus/internal/application/architecture_test.go</Path>", "<Path>Taskfile.yml</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>scripts/quality/architecture-check.mjs</Path>", "<Path>deploy/compose/compose.yml</Path>", "<Path>.github/workflows/ci.yml</Path>", "<Path>.github/workflows/codeql.yml</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/platform/logging/**</Path>", "<Path>go-admin-plus/internal/application/operations/doctor/**</Path>", "<Path>go-admin-plus/internal/modules/**</Path>", "<Path>release/linux/**</Path>"]
shared_paths: ["<Path>go-admin-plus/internal/app/product/**</Path>", "<Path>go-admin-plus/cmd/go-admin-plus/**</Path>", "<Path>Taskfile.yml</Path>", "<Path>scripts/go-admin-plus/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus/internal/app/product/**</Path> => T-09", "<Path>go-admin-plus/cmd/go-admin-plus/**</Path> => T-09", "<Path>Taskfile.yml</Path> => T-09", "<Path>scripts/go-admin-plus/**</Path> => T-09"]
---

# Ticket T-09: 统一产品 CLI、运行拓扑与 Profile Migration

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/09-product-cli-runtime-topology.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-09.md</Path>`

## 1. 战略与来源

- **目标：** 以一个 Server 二进制和明确 profile 拓扑接入全部新模块、日志、Doctor 与 migration 策略。
- **可观察产出：** `serve/worker/migrate/bootstrap/recover-admin/doctor/version` 可发现且失败脱敏；PG API/worker 分离，SQLite/dev 可一体。
- **来源：** `US-013~014`、`AC-029~033`、`ADR-005`、`ADR-012`、`ADR-015`。
- **当前事实：** cmd/go-admin-plus 只有 serve flags；migrate/config-check 是独立二进制；product Build 总是迁移并总是组合 workers。
- **Planning Depth 原因：** 组合根和部署启动顺序错误会使生产 schema 竞态、双 worker 或全服务不可用。

## 2. 决策状态

### 已锁定决策

- 单一 CLI 七个子命令；旧 Server 小命令无 alias。
- PG 生产先 migrate，serve 仅 API、worker 独立；SQLite/dev 可 `serve --with-worker`；Desktop 仍一体。
- PG schema 过旧/过新/未知不 ready 并退出；Server SQLite 自动迁移。

### 已采用的低影响假设

- 使用 Go 标准 flag 的子命令层，不新增 CLI framework；输出保持稳定机器退出码与简洁文本/JSON诊断。

### 已批准执行偏差 DEV-09-001

- **等级：** ticket + release sequencing；用户于 `2026-08-31T22:50:26+08:00` 以“都批准”明确批准。
- **提前范围：** 可在 Lead-owned Wave A 后端联合候选上，先实现 `internal/app/product` 的 migration/provider registry 与所需最小 probe/readiness composition，使 T-02/T-03/T-04/T-06 的领域 migration 能在 parent-candidate 被发现和验证。
- **禁止扩大：** early checkpoint 不得提前改 CLI、Task/scripts/Compose，不得修改 T-08 公共 wire contract，也不得宣称 T-09 result。
- **closure 约束：** frontmatter 的 T-07/T-08 依赖仍约束完整 T-09 验收；只有依赖 result、全部本 Ticket 行为和 required Gate 均成立后才能完成。

### 已批准执行偏差 DEV-09-002

- **等级：** ticket path ownership；用户以“都批准”授权全部必要偏差。
- **触发事实：** 删除旧 `cmd/migrate` 与 `cmd/config-check` 后，架构 canonical path、命令边界测试及 CI/CodeQL build list 仍强制引用旧入口，使 T-09 的零兼容收缩与必过架构门禁互相矛盾。
- **批准范围：** T-09 临时拥有 `internal/application/architecture_test.go`、`scripts/quality/architecture-check.mjs` 及两处 workflow 的旧命令引用，只把 canonical/build list 收敛到 `cmd/go-admin-plus` 与独立 Desktop sidecar。
- **禁止扩大：** 不改变 T-17 的生成器/架构策略，不提前实现 T-18 的 required CI/security gate，不修改 workflow 触发器、权限或安全工具。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| CLI、module/migration 注册、运行角色、readiness、logger/Doctor wiring、Task/scripts/Compose | T-02~08 模块、typed config、lifecycle/coordination | Desktop 首次设置（T-10）、最终 release/docs（T-21）、兼容小命令 |

## 4. 要构建什么

操作者通过一个二进制选择子命令和 profile。每个子命令只构建所需依赖：migrate 独占迁移，serve/worker 验证 schema，Bootstrap/Recovery 不启动 HTTP，Doctor 只读，version 不加载运行时。产品组合按角色注入 logger、业务 modules 和 worker，并通过 readiness 反映真实依赖。

## 5. 实现契约

- **入口或接缝：** `go-admin-plus <subcommand>`、Taskfile 和 Compose commands。
- **输入与输出：** typed profile/config/secret references；稳定 exit code、脱敏 stderr、服务 readiness。
- **公共接口变化：** CLI 破坏性替换；HTTP 使用 T-08 合同。
- **不变量：** PG serve 不跑 worker/migrate；PG worker 不监听 API；SQLite 单实例；version 无副作用。
- **状态或数据流：** parse -> typed snapshot -> role-specific dependency graph -> lifecycle -> exit。
- **错误与失败行为：** schema mismatch、配置、锁、依赖和 shutdown 分层失败，不监听半初始化服务。
- **兼容要求：** 删除 cmd/migrate、cmd/config-check 与旧脚本调用，不留 alias。
- **安全与隐私要求：** password/DSN 不进 argv/log；所有子命令复用 T-07 redaction。

## 6. 执行路线

0. 在 DEV-09-001 联合候选上建立 migration/provider 最小组合红灯与 checkpoint，记录所有输入 source SHA；该 checkpoint 只解除 Wave A 组合闭环。
1. 建立子命令解析、role graph、schema mismatch 和 secret 输出红灯 contract tests。
2. 把 migration/worker/API composition 从通用 Build 中拆成显式角色。
3. 注册 T-02~08 providers/use cases/handlers，并接入 logger/Doctor。
4. 更新 Taskfile、scripts 与 Compose 命令，删除旧小二进制。
5. 运行进程 lifecycle、双方言、readiness、shutdown、架构和构建验证。

## 7. 路径访问契约

- **预计修改点/可写范围：** frontmatter 指定 CLI、product、server host、Task/scripts/Compose。
- **只读上下文：** T-02~08 模块与 release policy。
- **共享路径：** product composition、Server CLI、Taskfile、scripts 由 T-09 唯一拥有。
- **保留或不动：** Desktop Rust `main.rs` 与 Desktop product adapter 由 T-10。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | CLI/process | 各子命令与三个 profile contract tests | 角色、输出、readiness 正确 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-09.md</Path>` |
| 失败路径 | schema/secret/lifecycle | old/new schema、锁、依赖失败、signal shutdown | 不 ready、不泄漏、资源逆序释放 | 同上 |
| 回归 | product command plane | `task task:contract && task architecture:check && task build TARGET=server PROFILE=server-sqlite` | 命令和构建通过 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行 CLI/process/host tests、双方言集成、lint/vet/build、task contract。
- **E2E disposition：** not-required：真实进程 contract 覆盖本 Ticket；浏览器/Desktop/clean-room E2E 由 T-19~21。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；本 Ticket 不在 source 声明 E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA、父分支包含关系、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** 集成所有 providers -> 新 CLI -> 更新 scripts/Compose -> 删除旧命令；PG 先 migrate 后 rollout。
- **兼容窗口：** 无；旧二进制与命令在同一 commit 收缩。
- **监控信号：** schema status、role/profile、readiness checker、worker ownership、shutdown class、Doctor summary。
- **回滚或前向恢复：** forward-only schema；制品回退要求 schema 可被旧制品识别，否则恢复备份；优先前向修复。
- **不可逆操作与批准点：** 生产 migration 与旧命令删除需 Lead Gate；本 Ticket 不执行生产迁移。
- **收缩条件：** 旧 cmd、脚本和部署引用扫描为零，PG serve 进程无 worker/migrate goroutine。

## 10. 验收标准

- [x] `AC-029~033` 的 CLI、角色、迁移、日志与 Doctor 行为成立。
- [x] 旧 Server 小命令和调用点为零，三 profile 构建/进程 contract 通过。
- [x] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-09.md</Path>`。
- [x] shared owner、commit、integration/result 和 E2E disposition 完整。
