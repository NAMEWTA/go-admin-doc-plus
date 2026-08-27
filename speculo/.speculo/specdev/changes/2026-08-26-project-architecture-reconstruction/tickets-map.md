---
schema_version: 3
artifact: tickets-map
change: 2026-08-26-project-architecture-reconstruction
status: in_progress
---

# Tickets Map: Go Admin Plus 自主产品架构重构

- **Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **Ticket 目录：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/</Path>`
- **Evidence 目录：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/</Path>`
- **下一 Goal Plan：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/goal-plan.md</Path>`

## 1. 目标与拆分策略

本 Map 将 `US-001` 至 `US-019` 和 `AC-001` 至 `AC-036` 拆成 21 个 Deep Ticket。T-01 至 T-05 是解除根治理、合同、运行时、数据库和前端共享阻塞的 prefactor；T-06 至 T-16 以可观察业务或宿主行为形成垂直切片；T-17 是唯一产品组合汇合点；T-18 至 T-20 独立验证三个发行平台；T-21 在全部 Gate 后原子收缩旧体系。

施工采用 expand -> vertical migrate -> product observe -> platform verify -> contract。模块 Ticket 只写模块自有 OpenAPI fragment、migration、后端模块和前端 domain/web-domain；共享根、合同、锁文件、产品注册和最终 CI 均有专属 owner，避免并行任务修改相同路径。

## 2. 执行清单

| ID | Ticket | 可观察产出 | Blocked By | Depth | Risk | Ready | Owner | Contract IDs | Wave/Gate | Status |
|---|---|---|---|---|---|---|---|---|---|---|
| T-01 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/01-root-governance-command-plane.md</Path>` | 根治理与统一任务入口 | — | deep | high | yes | codex-root | AC-001, AC-002 | W0 / G0 | done |
| T-02 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/02-openapi-contract-foundation.md</Path>` | OpenAPI 与双方生成基座 | T-01 | deep | high | yes | codex-root | AC-023, AC-025 | W1 / G1 | done |
| T-03 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/03-runtime-profiles-observability.md</Path>` | 三 Profile 配置和生命周期 | T-02 | deep | high | yes | codex-t03-runtime | AC-005, AC-026, AC-027 | W2 / G2 | done |
| T-04 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/04-dual-dialect-persistence-migrations.md</Path>` | 双方言迁移与 Database 基座 | T-03 | deep | high | yes | codex-t04-database | AC-003, AC-004, AC-024, AC-034 | W3 / G2 | done |
| T-05 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/05-frontend-workspace-shell.md</Path>` | Workspace、Shell 和交互基座 | T-02 | deep | high | yes | codex-root | AC-035, AC-036 | W2 / G2 | done |
| T-06 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/06-iam-session-account.md</Path>` | Web 安全登录与账户闭环 | T-02, T-03, T-04, T-05 | deep | critical | yes | codex-t06-iam | AC-008, AC-010, AC-012, AC-025, AC-036 | W4 / G3 | done |
| T-07 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/07-iam-administration-authorization.md</Path>` | IAM 管理与最终授权 | T-06 | deep | critical | yes | codex-t07-iam | AC-011, AC-013, AC-035, AC-036 | W5 / G3 | done |
| T-08 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/08-reliable-runtime-no-redis.md</Path>` | 无 Redis Outbox 与唯一 executor | T-03, T-04 | deep | critical | yes | codex-t08-reliability | AC-018, AC-034 | W4 / G3 | done |
| T-09 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/09-organization-module.md</Path>` | Organization 完整管理闭环 | T-07 | deep | high | yes | codex-t09-organization | AC-014, AC-035 | W6 / G4 | done |
| T-10 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/10-settings-module.md</Path>` | Settings 完整管理闭环 | T-07 | deep | high | yes | codex-t10-settings | AC-015, AC-035 | W6 / G4 | in_progress |
| T-11 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/11-audit-module.md</Path>` | Audit 可靠脱敏闭环 | T-06, T-08 | deep | high | yes | codex-t11-audit | AC-016, AC-035 | W5 / G4 | done |
| T-12 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/12-scheduler-module.md</Path>` | Scheduler 受控执行闭环 | T-07, T-08 | deep | critical | yes | codex-t12-scheduler | AC-017, AC-018, AC-035 | W6 / G4 | in_progress |
| T-13 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/13-files-module.md</Path>` | Files 安全读写闭环 | T-07 | deep | critical | yes | codex-t13-files | AC-020, AC-035 | W6 / G4 | in_progress |
| T-14 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/14-demo-tracer.md</Path>` | Demo 双方言 CRUD 曳光弹 | T-04, T-05, T-07 | deep | medium | yes | codex-t14-demo | AC-021, AC-024, AC-035 | W6 / G4 priority | done |
| T-15 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/15-generator-module.md</Path>` | Generator 预览生成编译闭环 | T-14 | deep | high | yes | codex-root | AC-019, AC-023, AC-028, AC-035 | W7 / G4 | in_progress |
| T-16 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/16-tauri-desktop-host.md</Path>` | Tauri 2 Desktop 安全闭环 | T-06, T-08, T-14 | deep | critical | yes | codex-t16-desktop | AC-006, AC-007, AC-009, AC-021, AC-027, AC-036 | W7 / G4 priority | in_progress |
| T-17 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/17-product-composition-matrix.md</Path>` | 三 Profile、双 App 完整产品 | T-09, T-10, T-11, T-12, T-13, T-15, T-16 | deep | critical | yes | unassigned | AC-003, AC-004, AC-011, AC-021, AC-022, AC-024, AC-025, AC-028, AC-034, AC-035, AC-036 | W8 / G5 | ready |
| T-18 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/18-linux-oci-compose-release.md</Path>` | Linux OCI/Compose 候选 | T-17 | deep | high | yes | unassigned | AC-030, AC-033 | W9 / G6 | ready |
| T-19 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/19-macos-universal-release.md</Path>` | macOS Universal DMG 候选 | T-17 | deep | critical | yes | unassigned | AC-031, AC-033 | W9 / G6 | ready |
| T-20 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/20-windows-nsis-release.md</Path>` | Windows x64 NSIS 候选 | T-17 | deep | critical | yes | unassigned | AC-032, AC-033 | W9 / G6 | ready |
| T-21 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/21-atomic-cutover-contraction.md</Path>` | 原子切换与旧体系归零 | T-18, T-19, T-20 | deep | critical | yes | unassigned | AC-001, AC-002, AC-023, AC-028, AC-029, AC-033 | W10 / G7 | ready |

Ticket frontmatter 是状态、依赖、深度和路径访问契约的权威；本表是同步投影。

## 3. 依赖 DAG

```text
T-01 [root prefactor]
  -> T-02 [shared contract]
       +-> T-03 [runtime]
       |    +-> T-04 [database]
       |    |    +-> T-08 [reliability]
       |    |    +-> T-14 [Demo]
       |    +-> T-06 [IAM session] <- T-04, T-05
       +-> T-05 [frontend]

T-06 -> T-07 [IAM authorization]
T-06 + T-08 -> T-11 [Audit]
T-07 -> T-09 [Organization], T-10 [Settings], T-13 [Files]
T-07 + T-08 -> T-12 [Scheduler]
T-04 + T-05 + T-07 -> T-14 [Demo]
T-14 -> T-15 [Generator]
T-06 + T-08 + T-14 -> T-16 [Desktop]

T-09 + T-10 + T-11 + T-12 + T-13 + T-15 + T-16
  -> T-17 [product integration]
       +-> T-18 [Linux release]
       +-> T-19 [macOS release]
       +-> T-20 [Windows release]
T-18 + T-19 + T-20 -> T-21 [atomic contract]
```

关键路径为 T-01 -> T-02 -> T-03 -> T-04 -> T-06 -> T-07 -> T-14 -> T-16 -> T-17 -> 原生发行 -> T-21。所有边均表示开始所需的实际接口、schema、宿主或发行证据。

## 4. 合同覆盖矩阵

| Contract ID | 覆盖 Ticket | 验证接缝 | 状态 | 说明 |
|---|---|---|---|---|
| AC-001 | T-01, T-21 | 根治理扫描 | covered | expand 后最终 contract |
| AC-002 | T-01, T-21 | 根 Task 合同 | covered | 命令建立与最终 CI 复核 |
| AC-003 | T-04, T-17 | PostgreSQL profile E2E | covered | 基座与产品场景 |
| AC-004 | T-04, T-17 | Server SQLite E2E | covered | 迁移、业务与单实例 |
| AC-005 | T-03 | 配置失败/脱敏 suite | covered | 三 profile 启动前失败 |
| AC-006 | T-16 | Desktop native tracer | covered | Tauri/sidecar/开窗 Gate |
| AC-007 | T-16 | Desktop migration/restart | covered | 备份与幂等升级 |
| AC-008 | T-06 | Web IAM E2E | covered | Cookie/CSRF/Session hash |
| AC-009 | T-16 | Desktop Session tracer | covered | Stronghold/transport proxy |
| AC-010 | T-06 | Session lifecycle suite | covered | revoke/timeout/rotation |
| AC-011 | T-07, T-17 | 权限/数据范围矩阵 | covered | 模块与产品复核 |
| AC-012 | T-06 | IAM account integration | covered | 资料/密码/退出 |
| AC-013 | T-07 | IAM API/Web E2E | covered | 用户/角色/授权管理 |
| AC-014 | T-09 | Organization suite | covered | 树/岗位/引用不变量 |
| AC-015 | T-10 | Settings suite | covered | 参数/字典/secret 边界 |
| AC-016 | T-11 | Audit/redaction suite | covered | 可靠消费和脱敏 |
| AC-017 | T-12 | Scheduler clock/E2E | covered | 注册类型与停止语义 |
| AC-018 | T-08, T-12 | worker/outbox recovery | covered | 平台与调度消费者 |
| AC-019 | T-15 | generator golden/compile | covered | 预览与生成门禁 |
| AC-020 | T-13 | Files security suite | covered | 授权和路径逃逸 |
| AC-021 | T-14, T-16, T-17 | Demo Web/Desktop tracer | covered | 双方言与双 App |
| AC-022 | T-17 | shared manifest E2E | covered | 双 App 能力等价 |
| AC-023 | T-02, T-15, T-21 | generate/conformance | covered | 基座、消费者和最终扫描 |
| AC-024 | T-04, T-14, T-17 | migration matrix | covered | 平台、参考模块与全产品 |
| AC-025 | T-02, T-06, T-17 | negative contract suite | covered | 错误模型与产品复核 |
| AC-026 | T-03 | observability host suite | covered | 生命周期语义 |
| AC-027 | T-03, T-16 | config/profile suite | covered | Server 与 Desktop 来源 |
| AC-028 | T-15, T-17, T-21 | architecture boundary | covered | 生成、组合和最终 Gate |
| AC-029 | T-21 | zero compatibility scan | covered | 原子收缩 |
| AC-030 | T-18 | Linux native/container Gate | covered | OCI/Compose/supply chain |
| AC-031 | T-19 | macOS protected Gate | covered | Universal DMG |
| AC-032 | T-20 | Windows protected Gate | covered | x64 NSIS |
| AC-033 | T-18, T-19, T-20, T-21 | CI policy contract | covered | 平台 Gate 与根汇总 |
| AC-034 | T-04, T-08, T-17 | cache-disabled matrix | covered | 平台、可靠性和产品 |
| AC-035 | T-05, T-07, T-09, T-10, T-11, T-12, T-13, T-14, T-15, T-17 | component/Web E2E | covered | 共享行为与模块证明 |
| AC-036 | T-05, T-06, T-07, T-16, T-17 | App Shell/双 App E2E | covered | Shell 建立到产品复核 |

不存在 `uncovered` 或 `deferred` 合同。

## 5. 并行与路径所有权

- implementation subagent 上限为 3，来自 `<Path>{roots.state}/specdev/config.json</Path>`；Lead 不计入。
- 当前 Goal Plan 固定 `ticket_workspace_policy: required` 与 `integration_gate: candidate-merge`；每个 Ticket 使用独立 source worktree，Lead 串行集成 candidate。
- Ticket frontmatter 是可写路径权威；并行模块只写自有 fragment、module、domain、web-domain 和测试。
- Lead 唯一修改 SpecDev 状态、Evidence 和父分支；required E2E 不在 source-worktree 执行。

| 并行组 | Ticket | Writable 交集 | 共享 owner/处理 |
|---|---|---|---|
| P2 | T-03, T-05 | 无 | 后端 runtime 与前端 workspace 分离 |
| P4 | T-06, T-08 | 无 | IAM 与 reliable-runtime 分离 |
| P5 | T-07, T-11 | Session service/test 与 workspace lock 已按原 result 串行移交；`T11-D04` 仅重开 Audit 自有 capability 文件/测试 | T-07 已 done；T-11 原 result 保留，当前只补 IAM registry 所需 Audit permissions/menu，其他共享路径仍只读 |
| P6 | T-09, T-10, T-12, T-13, T-14 | T-14/T-09 根 verify/lock 接入按 `T14-D01`、`T09-D01` 串行完成；T-12/T-13 现分别以 `T12-D01`/`T13-D01` 只拥有 package-local manifests，T-13 另精确拥有 Browser Files adapter export/manifest；后续根 verify/lock 继续由 Lead 串行分配 | 各模块独占合同/schema/backend/frontend；Scheduler 只读消费 T-08 同一全局 lease，Files 的 Desktop/config composition 延后 |
| P7 | T-15, T-16 | `T15-D01` 只开放 Generator 自有 manifests；`T16-D01` 只把 T-05 已预留的 Desktop adapter 移交 T-16；两者根 verify/lock stage 2 均等待 T-09 result | Generator 与 Desktop 分离，共同只读 Demo |
| P9 | T-18, T-19, T-20 | 无 | 各平台独占 release/script/workflow |

| 共享路径 | 唯一 owner | 消费方式 |
|---|---|---|
| `<Path>Taskfile.yml</Path>`、`<Path>.husky/**</Path>`、根分类脚本 | T-01；T-02 串行拥有 `contract:lint`/`generate:check` 并令 `generate` 唯一委派 canonical generator | 其他 Ticket 只读调用 |
| `<Path>contracts/openapi/openapi.yaml</Path>`、公共 components、合同工具与公共 client | T-02 | 模块写自有 fragment |
| kernel/config/observability | T-03 | 宿主和模块只读依赖 |
| Database/migration API、`<Path>go-admin-plus/go.mod</Path>`、`<Path>go-admin-plus/go.sum</Path>` | T-04；T-02 先串行拥有合同生成器/transport 依赖 | 模块实现自有 Provider，不改依赖清单 |
| Workspace/lock/adapters/app-shell manifest+core/platform/ui | T-05（含 T05-D01、T05-D02）；T-02/T-07/T-11/T-14/T-09 串行接入已完成；现由 `T16-D02` 只拥有 Desktop checks 与两个既有 importers，T-15/T-10 继续等待串行 amendment | 模块只写获批 manifest/importer；browser adapter 与 App Shell core 保持只读 |
| Shared list request state | T-05 基座；T-14 under `T14-D04` 精确修改 `<Path>go-admin-plus-ui/packages/ui/src/list.ts</Path>` 与既有回归 | 请求可规范化/校验，最新成功前不提交 query state；模块只配置/包装，不复制列表状态机 |
| IAM Session request authorization seam | T-06；`T07-D02` 精确追加 `AuthorizeRequest` 与既有测试 | T-07 HTTP 只消费统一 token/CSRF/touch/rotation 结果，不读取 Session 私表或复制 secret/hash 规则 |
| IAM Session login audit seam | T-06；T-07 result 已完成 `T07-D02`，现由 `T11-D02/T11-D03` 追加 fail-closed 模块无关 Login Fact Port 与精确 test-only 调用点 | T-11 Audit adapter 同步记录成功/失败登录；Session 不导入 Audit且无生产 discard，T-17 composition 必须注入 Port，密码/用户名/token/CSRF 不进入事实 |
| IAM Module Capability Registry | T-07 基座；T-14 under `T14-D02/T14-D03` 精确新增 `<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry.go</Path>` 与对应测试 | 模块声明 Permission Code 与 protected menu 并调用 IAM registry；只有 IAM 写 capability/role grant 表，T-17 composition 显式注册完整目录 |
| IAM Organization consumer Port | T-09 under `T09-D01`，精确文件为 `<Path>go-admin-plus/internal/modules/iam/administration/organization_port.go</Path>` | IAM 定义最小投影接口，Organization 提供 adapter，T-17 只负责显式注入；IAM 不导入 Organization |
| Outbox/coordination/local cache | T-08 | Audit/Scheduler 只读消费 |
| Browser Files transfer adapter export/manifest | T-13 under `T13-D01` | 只追加 Files 的流式上传/下载 adapter、测试、export 与直接依赖；不得修改通用 RuntimePort 或其他 adapter |
| `<Path>release/shared/sidecar/**</Path>` | T-16 | macOS/Windows 只读打包 |
| product OpenAPI/composition/manifest | T-17 | 发行 Ticket 只读消费 |
| 根 CI、quality scripts、README/docs | T-21 | 最终收缩专属写入 |

## 6. Gate、Wave 与集成点

- **G-BASELINE：** Lead 审计当前 170 项 tracked 与 6 个 untracked 条目、复核基线门禁并形成获授权 baseline commit；所有 source worktree 基于该结果或后续最新父结果。
- **G0-G2 Foundations：** W0-W3 的 T-01 至 T-05；证明根任务、共享合同、runtime、database 和 frontend 基座。
- **G3 Identity/Reliability：** W4-W5 的 T-06 至 T-08；证明登录、授权和无 Redis 正确性。
- **G4 Modules：** W5-W7 的 T-09 至 T-16；最多 3 个 implementation owner，T-14/T-16 优先解锁关键路径，Lead 串行 candidate integration。
- **G5 Product Integration：** W8 的 T-17；Lead 在 parent-candidate 运行三 profile、双 App 和全边界 E2E。
- **G6 Protected Release：** W9 的 T-18/T-19/T-20；平台 runner 独立阻断，正式凭据和远端动作仍需额外批准。
- **G7 Atomic Contract：** W10 的 T-21；冻结全部 result SHA 后删除旧体系并运行最终零扫描。

由于 Ticket 数量超过 10、全部为 Deep、包含 migration/shared contract/多平台 Gate，本 change 必须进入 Goal Plan，不能直接无编排实施。

## 7. 横切契约与风险

- 不保留旧 API、schema、配置、数据、命名或内部模式兼容；施工期 expand 不是产品兼容窗口。
- Server 支持 PostgreSQL/SQLite，Desktop 仅 SQLite；SQLite 单实例，PostgreSQL 多 API 副本但唯一 worker。
- Redis、JWT、refresh token、Casbin、tenant、Wails、MySQL、SQL Server、AutoMigrate 在 T-21 后零保留。
- OpenAPI fragment、模块 migration/repository、Permission Code 和前端包均由模块 owner 管理；跨模块只用消费者 Port 和 Integration Event。
- Web Session 使用 Cookie+CSRF，Desktop 使用 Stronghold proxy；任何日志、错误、审计和供应链 Evidence 都不得包含 secret。
- 破坏性 schema、生产签名凭据、远端发布和实现阶段删除仍需 I-implement/Goal Plan 对应授权；本 Map 批准不扩大副作用权限。

## 8. 同步规则

- Ticket 状态变化后同步执行清单；Ticket frontmatter 始终为权威。
- Goal Plan 创建后，Wave、Gate、workspace policy、owner 和集成顺序以该 Goal Plan 为编排权威。
- 依赖、合同或路径 owner 变化后运行 `<Path>{roots.workflows}/specdev/common/tools/validate-specdev.mjs</Path>`。
- 每个完成 Ticket 写入 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/{ticket-id}.md</Path>` 并记录 commit/result SHA。
- 内部工件不得使用相对 Markdown 链接；偏差必须停止并更新 Ticket/Map，而不是静默扩展范围。
