---
schema_version: 3
artifact: tickets-map
change: 2026-08-29-productization-and-ui-reconstruction
status: ready
---

# Tickets Map: 产品化补强与管理端 UI 重构

- **Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **Ticket 目录：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/</Path>`
- **Evidence 目录：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/</Path>`
- **后续 Goal Plan：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/goal-plan.md</Path>`

## 1. 目标与拆分策略

本 Map 将 `US-001~017` 与 `AC-001~039` 切成 21 个可独立验收的 Ticket，目标是把项目从“架构基础较好但存在生产阻断”推进到三 Profile 可从空库安装、可安全运行、可恢复、可观测且有真实发布证据的产品化脚手架。拆分遵循以下策略：

- 先修复依赖漏洞、Bootstrap、Session、数据范围、账号生命周期、文件容量和可观测性等安全/数据不变量，再汇合公共 OpenAPI 与产品 CLI。
- UI 采用 `设计系统 -> Router/Shell -> Session client -> 业务页面` 的顺序；后端仍是授权事实源，页面不复制安全判定。
- schema 与 wire contract 使用 expand-migrate-contract：T-02~06 建立领域能力，T-08 原子收敛公共合同，T-09 收敛运行拓扑，T-21 才移除旧文档与部署描述。
- 共享合同、产品组合根、lockfile、App Shell、CI、E2E 和发布文档均设置唯一写 owner；依赖边同时承担路径串行化。
- T-18~21 是不可跳过的证据 Gate：真实 PostgreSQL、安全供应链、Web 浏览器、Desktop 原生和三 Profile clean-room 不得用 skip 或 not-run 伪装通过。
- 当前规模包含 18 个 deep Ticket、多个 critical 风险与 10 个候选 Wave，正式实现前必须运行 P-goal-plan；本 Map 不授权实现、集成、提交、推送或清理。

## 2. 执行清单

| ID | Ticket | 可观察产出 | Blocked By | Depth | Risk | Ready | Owner | Contract IDs | Wave/Gate | Status |
|---|---|---|---|---|---|---|---|---|---|---|
| T-01 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/01-dependency-security-baseline.md</Path>` | GO-2026-6112 不再可达 | — | standard | high | yes | codex-root | AC-037 | A / Security Prefactor | done |
| T-02 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/02-bootstrap-and-admin-recovery.md</Path>` | 空库首管理员与离线恢复 | — | deep | critical | yes | codex-root | AC-001, AC-002, AC-003, AC-006, AC-007 | A / Identity Prefactor | done |
| T-03 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/03-login-session-hardening.md</Path>` | 持久限流与稳定 CSRF Session | — | deep | critical | yes | codex-root | AC-008~013 | A / Session Prefactor | done |
| T-04 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/04-organization-data-scopes.md</Path>` | 组织归属与五种数据范围 | — | deep | critical | yes | codex-root | AC-019~022 | A / Authorization Prefactor | done |
| T-05 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/05-account-deletion-lifecycle.md</Path>` | Tombstone、会话撤销与文件处置 | T-02, T-04 | deep | critical | yes | codex-root | AC-013, AC-023~026 | B / Lifecycle | done |
| T-06 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/06-file-capacity-governance.md</Path>` | 配额、磁盘水位与崩溃对账 | — | deep | high | yes | codex-root | AC-027~029 | A / Storage Prefactor | done |
| T-07 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/07-observability-and-doctor.md</Path>` | 结构化日志与 Doctor | — | deep | high | yes | codex-root | AC-029, AC-032, AC-033 | A / Operations Prefactor | done |
| T-08 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/08-public-contract-integration.md</Path>` | OpenAPI、生成类型与 HTTP adapter 一致 | T-02, T-03, T-04, T-05, T-06 | deep | critical | yes | codex-root | AC-009~013, AC-019~029, AC-038 | C / Contract Gate | done |
| T-09 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/09-product-cli-runtime-topology.md</Path>` | 单一产品 CLI 与三 Profile 拓扑 | T-07, T-08（closure）；DEV-09-001 early checkpoint | deep | critical | yes | codex-root | AC-029~033 | D / Runtime Gate | done |
| T-10 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/10-desktop-first-setup.md</Path>` | Desktop 原生首次设置与 Session 转换 | T-02, T-09 | deep | critical | yes | unassigned | AC-004, AC-005, AC-031 | E / Desktop Setup | ready |
| T-11 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/11-ui-design-system.md</Path>` | Element Plus token 与共享管理组件 | — | standard | high | yes | codex-root | AC-017, AC-018 | A / UI Prefactor | done |
| T-12 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/12-router-workspace-shell.md</Path>` | Vue Router 单一事实源与双宿主 Shell | T-11 | deep | high | yes | unassigned | AC-014~018 | B / Shell Gate | ready |
| T-13 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/13-cross-tab-session-clients.md</Path>` | Web 跨标签与 Desktop Session client | T-03, T-08 | deep | high | yes | unassigned | AC-011, AC-013 | D / Client Auth | ready |
| T-14 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/14-iam-organization-ui.md</Path>` | IAM/Organization 完整管理体验 | T-04, T-05, T-08, T-11, T-12, T-13 | deep | high | yes | unassigned | AC-014~026, AC-038 | E / Security UI | ready |
| T-15 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/15-operations-pages-ui.md</Path>` | Audit/Settings/Scheduler 成熟页面 | T-11, T-12 | standard | medium | yes | unassigned | AC-014~018, AC-038 | D / Operations UI | ready |
| T-16 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/16-tools-pages-ui.md</Path>` | Files/Generator/Demo 成熟页面 | T-06, T-08, T-11, T-12 | standard | high | yes | unassigned | AC-014~018, AC-027~029, AC-038 | D / Tools UI | ready |
| T-17 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/17-generator-current-scaffold.md</Path>` | 当前架构全垂直生成脚手架 | T-04, T-05, T-06, T-09, T-11, T-14, T-16 | deep | high | yes | unassigned | AC-034 | F / Generator Gate | ready |
| T-18 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/18-postgres-security-ci-gates.md</Path>` | 真实 PostgreSQL 与安全供应链 CI | T-01, T-09, T-17 | deep | high | yes | unassigned | AC-002, AC-006, AC-008, AC-019, AC-023, AC-027, AC-030, AC-035, AC-037 | G / Required CI | ready |
| T-19 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/19-web-browser-e2e.md</Path>` | SQLite/PostgreSQL 真实 Web E2E | T-13, T-14, T-15, T-16, T-18 | deep | high | yes | unassigned | AC-011, AC-013~018, AC-036, AC-038 | H / Required Web E2E | ready |
| T-20 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/20-desktop-native-e2e.md</Path>` | macOS Tauri 原生 E2E | T-10, T-14, T-15, T-16, T-19 | deep | high | yes | unassigned | AC-004, AC-005, AC-016~018, AC-031, AC-036 | I / Required Native E2E | ready |
| T-21 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/21-release-docs-clean-room.md</Path>` | 文档收敛与三 Profile clean-room | T-17, T-18, T-19, T-20 | deep | high | yes | unassigned | AC-030~039 | J / Release Gate | ready |

Ticket frontmatter 是状态、依赖、深度和路径访问契约的权威；本表是同步投影，不得独立修改出另一套真相。

## 3. 依赖 DAG

```text
Wave A / independent prefactors
T-01 ───────────────────────────────────────────────┐
T-02 ───────┬───────────────┐                       │
T-03 ───────┼───────────┐   │                       │
T-04 ──┬────┼───────┐   │   │                       │
T-06 ──┼────┼───┐   │   │   │                       │
T-07 ──┼────┼───┼───┼───┼───┼──→ T-09 ──→ T-10    │
T-11 ──┼────┼───┼───┼───┼───┼──→ T-12             │
        │    │   │   │   │   │                      │
        │    └──→T-05┘   │   │                      │
        │                 │   │                      │
        └─────────────────┴───┴──→ T-08 ──┬─→ T-13 │
                                           ├─→ T-09 │
T-11 ─→ T-12 ───────────────┬──────────────┼─→ T-15 │
T-06 + T-08 + T-11 + T-12 ──┼──────────────┴─→ T-16│
T-04 + T-05 + T-08 + T-11 + T-12 + T-13 ───→ T-14 │
                                                      │
T-04 + T-05 + T-06 + T-11 + T-14 + T-16 ─→ T-17    │
T-01 + T-09 + T-17 ────────────────────────→ T-18 ←─┘
T-13 + T-14 + T-15 + T-16 + T-18 ─────────→ T-19
T-10 + T-14 + T-15 + T-16 + T-19 ─────────→ T-20
T-17 + T-18 + T-19 + T-20 ─────────────────→ T-21
```

- **关键汇合 T-08：** 领域能力完成后才原子修改 OpenAPI、生成物与 HTTP adapter，避免 wire contract 多头写入。
- **关键汇合 T-09：** 日志/Doctor 与公共合同稳定后再替换产品命令、迁移所有权和运行角色。
- **关键汇合 T-14：** 安全数据模型、Session client、设计系统与 Router/Shell 全部稳定后构建高风险 IAM UI。
- **关键汇合 T-17：** 以实际完成后的后端、前端和权限形态升级 Generator，禁止生成旧架构。
- **证据链 T-18 -> T-19 -> T-20 -> T-21：** required gate 串行吸收 candidate，最终 clean-room 才能判定整体完成。

## 4. 合同覆盖矩阵

| Contract ID | 覆盖 Ticket | 验证接缝 | 状态 | 说明 |
|---|---|---|---|---|
| AC-001 | T-02 | Bootstrap command/use case | covered | 空库一次性首管理员 |
| AC-002 | T-02, T-18 | SQLite/PostgreSQL concurrency | covered | 仅一个初始化成功 |
| AC-003 | T-02 | secret input/output audit | covered | 无固定或回显凭据 |
| AC-004 | T-10, T-20 | native first setup | covered | Desktop SQLite 首次设置 |
| AC-005 | T-10, T-20 | setup-to-session transition | covered | 凭据不进入 WebView |
| AC-006 | T-02, T-18 | dual-dialect migrate/bootstrap | covered | 双方言一致 |
| AC-007 | T-02 | offline recovery | covered | 恢复可审计且受限 |
| AC-008 | T-03, T-18 | persistent login protection | covered | 跨进程/实例限流 |
| AC-009 | T-03, T-08 | account/IP policy contract | covered | 失败预算与状态明确 |
| AC-010 | T-03, T-08 | retry/error contract | covered | 客户端可解释退避 |
| AC-011 | T-03, T-13, T-19 | cross-tab session flow | covered | 标签页不互相制造 403 |
| AC-012 | T-03, T-08 | session renewal transaction | covered | 续期并发正确 |
| AC-013 | T-03, T-05, T-13, T-19 | revocation propagation | covered | 删除/恢复后立即失效 |
| AC-014 | T-12, T-14, T-15, T-16, T-19 | route deep-link | covered | URL、标题和页面一致 |
| AC-015 | T-12, T-14, T-15, T-16, T-19 | history/navigation | covered | 页内切换同步 history |
| AC-016 | T-12, T-14, T-15, T-16, T-19, T-20 | Web/Desktop shell | covered | 双宿主行为一致 |
| AC-017 | T-11, T-12, T-14, T-15, T-16, T-19, T-20 | visual/UI contract | covered | Element Plus 管理体验 |
| AC-018 | T-11, T-12, T-14, T-15, T-16, T-19, T-20 | responsive/accessibility | covered | 无溢出遮挡且可操作 |
| AC-019 | T-04, T-08, T-14, T-18 | five data scopes | covered | 五种范围全链路 |
| AC-020 | T-04, T-08, T-14 | organization membership | covered | 归属实时生效 |
| AC-021 | T-04, T-08, T-14 | scoped query enforcement | covered | 查询不可绕过范围 |
| AC-022 | T-04, T-08, T-14 | capability/scope separation | covered | 能力与范围职责清晰 |
| AC-023 | T-05, T-08, T-14, T-18 | account tombstone | covered | 删除不用物理级联冒险 |
| AC-024 | T-05, T-08, T-14 | last-admin invariant | covered | 最后系统管理员受保护 |
| AC-025 | T-05, T-08, T-14 | file disposition workflow | covered | 保留/转移/清除明确 |
| AC-026 | T-05, T-08, T-14 | deletion audit/idempotency | covered | 重试与审计成立 |
| AC-027 | T-06, T-08, T-16, T-18 | account/global quotas | covered | 并发容量限制 |
| AC-028 | T-06, T-08, T-16 | disk watermark response | covered | 低空间拒绝可解释 |
| AC-029 | T-06, T-07, T-08, T-09, T-16 | reconcile/readiness | covered | 元数据/文件/容量可恢复 |
| AC-030 | T-09, T-18, T-21 | explicit migration ownership | covered | serve/worker 不暗迁移 PG |
| AC-031 | T-09, T-10, T-20, T-21 | three runtime profiles | covered | Server 双库、Desktop SQLite |
| AC-032 | T-07, T-09, T-21 | structured logging | covered | level 生效且具关联字段 |
| AC-033 | T-07, T-09, T-21 | Doctor/readiness | covered | 诊断稳定且脱敏 |
| AC-034 | T-17, T-21 | generated module compile gate | covered | 生成当前全垂直架构 |
| AC-035 | T-18, T-21 | real PostgreSQL CI | covered | 缺依赖必须失败 |
| AC-036 | T-19, T-20, T-21 | real Web/native E2E | covered | 真实用户流程证据 |
| AC-037 | T-01, T-18, T-21 | vulnerability/security gate | covered | 可达漏洞和供应链门禁 |
| AC-038 | T-08, T-14, T-15, T-16, T-19, T-21 | API/UI compatibility | covered | UI 始终对齐当前后端 |
| AC-039 | T-21 | clean-room and docs scan | covered | 新用户只靠当前文档完成安装 |

不存在 `uncovered` 或未经批准的 `deferred` 合同。

## 5. 并行与路径所有权

- implementation subagent 上限已由 Goal Plan 快照为 3，Lead 不计入；Lead 可按测试资源和风险进一步降低。
- review/research/test-observation agent 不设 SpecDev 数字上限，但保持只读。
- shared owner 为专用 Ticket；Lead 是 SpecDev 状态与父分支 integration owner。
- 项目路径契约以 Ticket frontmatter 为准；Ticket 不得顺手修改另一 Ticket 的可写路径。
- 本 change 已选择 `required`：每个实现 Ticket 使用独立 source worktree，Lead 在 parent-candidate 验证并串行推进 `main`；只读调查不进入 I-implement Ticket。
- G0 在任何 source worktree 创建前保护现有 database 与 Desktop `main.rs` 变更，记录 hash 和非丢弃快照；只分别导入 T-02/T-10。

| Ticket A | Ticket B | Writable 交集 | 真实依赖 | 处理 |
|---|---|---|---|---|
| T-01 | T-02/T-03/T-04/T-06/T-07/T-11 | 无 | 否 | Wave A 可并行 |
| T-02 | T-03/T-04/T-06/T-07/T-11 | 无 | 否 | Wave A 可并行；保留用户现有 database 变更 |
| T-05 | T-12 | 无 | 否 | Wave B 可并行 |
| T-08 | 所有合同消费者 | OpenAPI、Go wire、TS generated | 是 | T-08 是唯一 shared owner，消费者先只读 |
| T-09 | T-10/T-18/T-21 | product root、CLI、Task、scripts、compose | 是 | 按 DAG 串行，T-09 先冻结命令面 |
| T-10 | T-20 | Desktop native code | 是 | `main.rs` 仅 T-10 可写；T-20 只写测试/验证脚本 |
| T-11 | T-12/T-14/T-15/T-16 | lockfile、workspace UI package | 是 | T-11 是依赖和设计系统 owner，后续消费 |
| T-12 | T-13/T-14/T-15/T-16 | App Shell | 部分 | T-12 唯一写 Shell；T-13 只写 adapter/domain |
| T-13 | T-15/T-16 | 无 | 否 | Wave D 可并行 |
| T-15 | T-16 | 无 | 否 | 独立 Web Domain 可并行 |
| T-17 | T-18 | architecture checks | 是 | T-17 先升级生成规则，T-18 再接入 CI |
| T-18 | T-19/T-20 | CI/security gate | 是 | T-18 冻结 required CI 后启动 E2E |
| T-19 | T-20 | `tests/e2e/**` 与其 desktop 子树 | 是 | T-19 先建公共 Web runner，T-20 串行接管 desktop 子树 |
| T-21 | T-02/T-09/T-18/T-19/T-20 | database/deploy/release/docs 投影 | 是（直接或传递） | 最终串行收敛，不提前改发布文档 |

唯一 shared owner 汇总：T-08（公共合同/生成物）、T-09（产品组合根/CLI/Task）、T-10（`main.rs`）、T-11（lockfile/UI package）、T-12（App Shell）、T-18（CI/security）、T-19（Web E2E 根）、T-20（Desktop E2E 子树）、T-21（docs/deploy/release）。

## 6. Gate、Wave 与集成点

| Wave | 可启动 Ticket | 进入条件 | 集成/退出 Gate |
|---|---|---|---|
| A | T-01, T-02, T-03, T-04, T-06, T-07, T-11 | G0 dirty-state 快照通过，按最新 `main` 建立独立 source worktree | 各领域 source checks、source commit 与 Lead candidate result |
| B | T-05, T-12 | 对应 A 依赖已集成 | 生命周期不变量、Shell route tests |
| C | T-08 | T-02~06 所需依赖全部集成 | OpenAPI lint/generate drift/Go+TS contract gate |
| D | T-09, T-13, T-15, T-16 | 各自依赖已集成 | runtime、client auth 与业务 UI component tests |
| E | T-10, T-14 | T-09/T-13 及全部 UI/领域依赖已集成 | Desktop setup 与安全管理 UI gate |
| F | T-17 | 当前后端、UI 与权限模式稳定 | isolated generated module compile/test |
| G | T-18 | T-01、T-09、T-17 已集成 | 真实 PG/race/security/SBOM required CI |
| H | T-19 | 业务 UI 与 required CI 已集成 | SQLite+PG 真实浏览器 E2E |
| I | T-20 | Desktop setup、业务 UI、Web runner 已集成 | macOS Tauri 原生 E2E |
| J | T-21 | T-17~20 全部通过 | 三 Profile clean-room、文档扫描、最终 candidate |

- 每个 Wave 的 source commit、parent-candidate 验证、父分支 result SHA 和 Evidence 由 Lead 按 Goal Plan 集成次序管理；同 Wave 最多 3 个 implementation subagent。
- T-18、T-19、T-20、T-21 是 required Gate；环境缺失、命令未运行、测试 skip 或只完成静态检查均视为失败。
- 签名与公证按用户决策为 `not-required`；不得记为 `passed`，也不得阻止个人自用构建验证。
- T-21 通过前不允许归档该 change；本地 source commit 与 candidate integration 已授权，push、publish、deploy 和 source cleanup 仍未授权。
- 正式编排权威为 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/goal-plan.md</Path>`；Wave、Gate、workspace、动态派单、candidate 顺序与恢复点变化必须同步回本 Map。

## 7. 横切契约与风险

- **数据库：** Server 支持 SQLite/PostgreSQL；Desktop 仅 SQLite。所有 schema 变化双方言同版本语义，PostgreSQL 生产迁移与 serve/worker 解耦。
- **身份：** 无默认账号、默认密码或静态 Bootstrap SQL；首管理员创建和恢复必须一次性、可审计、并发安全，凭据只通过安全外部输入。
- **认证：** Session Cookie 仍是不透明认证根；CSRF 不因其他标签页请求滚动失效；高成本密码校验有持久账户/IP 防护。
- **授权：** 后端 capability 与数据范围是最终事实源；UI 隐藏/禁用仅改善体验，不能成为安全边界；租户能力保持完全删除。
- **删除：** 账号删除采用 tombstone 和显式文件处置，最后系统管理员不能被删除；不可逆清除需显式批准和 Evidence。
- **文件：** 单对象、账户、对象数量与全局磁盘水位共同生效；数据库保留与物理文件写入失败后可对账恢复。
- **前端：** pnpm monorepo、Element Plus 设计系统和 Vue Router 是唯一当前模式；Web 与 Desktop 共享 shell/domain，但 adapter/history/native setup 按宿主隔离。
- **公共合同：** OpenAPI-first；T-08 是合同与生成物唯一写 owner，手写 adapter 不得重新定义 wire 类型。
- **Generator：** 以 T-17 完成时的当前架构生成双方言迁移、OpenAPI、Go vertical slice、capability、domain/Web Domain、组合注册和测试，不保留旧模板。
- **可观测性：** 日志和 Doctor 默认脱敏，稳定输出运行角色、profile、request/job correlation 与 readiness 原因，不泄漏 password、CSRF、Cookie、DSN secret。
- **验证：** required PostgreSQL、浏览器与 native E2E 缺环境即失败；所有 passed 结论必须关联命令、环境、退出状态和 result SHA。
- **恢复：** 数据/协议 Ticket 使用 expand-migrate-contract；发布候选失败优先前向修复，clean-room 只操作明确 disposable roots，不破坏 `<Path>dev_store/**</Path>`。

## 8. 同步规则

- Ticket 状态变化后同步执行清单；`ready` 仅表示实现输入充分，不表示已授权实现。
- Ticket ID、路径、依赖或 frontmatter 不一致时，以 Ticket 文件为权威并修复本 Map。
- Goal Plan 存在时，Wave、Gate、workspace 和 owner 以 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/goal-plan.md</Path>` 为编排权威。
- 依赖、合同覆盖或路径所有权变化后运行 `node speculo/workflows/specdev/common/tools/validate-specdev.mjs --stage tickets speculo/.speculo/specdev/changes/2026-08-29-productization-and-ui-reconstruction`。
- 实现完成时必须创建对应 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-XX.md</Path>`，并同步 Ticket/Map 状态。
- 内部工件不得使用相对 Markdown 链接；项目文件使用项目相对 Path，状态与工作流文件分别使用已注册的 `<Path>{roots.state}/specdev/</Path>` 和 `<Path>{roots.workflows}/specdev/</Path>`。
