---
schema_version: 6
artifact: goal-plan
change: 2026-08-26-project-architecture-reconstruction
status: blocked
modes: [migration, high-assurance, reference-conformance, release-coordination]
orchestration: lead-directed
lead: codex-root
implementation_agent_limit: 3
integration_attempt_limit: 3
ticket_workspace_policy: required
integration_gate: candidate-merge
ready_for_execution: false
---

# Goal Plan: Go Admin Plus 自主产品架构重构

- **Goal Plan：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/goal-plan.md</Path>`
- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **Tickets Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **Ticket 目录：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/</Path>`
- **Evidence 目录：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/</Path>`

## 1. Outcome and Authority

### Outcome

把当前自主维护产品重构为唯一根治理边界、Go 原生模块化单体、真实 pnpm Workspace 和 Tauri 2 Desktop。最终产品在 Server PostgreSQL、Server SQLite、Desktop SQLite 三个 profile 上具备八个业务模块、统一 OpenAPI 3.1 合同、安全数据库 Session、无 Redis 可靠运行时、双 App 一致能力和 Linux/macOS/Windows 可验证发行链。

执行采用每 Ticket 独立 source worktree、Lead-owned parent-candidate 集成和最多 3 个 implementation owner。新结构先 expand 并逐个形成可验证垂直切片；只有产品与三平台 Gate 完整关闭后，T-21 才执行旧体系原子 contract。

### Success and False Completion

成功必须同时满足：

- `AC-001` 至 `AC-036` 均有 Lead 核对的通过 Evidence；
- T-01 至 T-21 每个非 cancelled Ticket 都有非空 source commit、通过的 parent-candidate、父分支 `result_sha` 和路径核对；
- 三 profile、Web/Desktop、权限/Session、迁移/Outbox、架构边界、原生制品和供应链证据通过对应 Gate；
- T-21 后 allowlist 外旧目录、旧合同、Wails、JWT、refresh token、Casbin、Redis、tenant、MySQL、SQL Server、AutoMigrate 和临时迁移标记零命中；
- Ticket、Map、Goal Plan、Evidence、change status 与 Git checkpoint 一致，无未集成 source/candidate。

以下均属于伪完成：仅编译通过、仅在 source worktree 自报 E2E、只移动目录未闭合行为、以 mock 代替真实数据库/原生安装、保留 compatibility shim、Evidence 没有不可变 SHA、父分支未包含候选、或用 empty commit/Evidence-only 关闭 Ticket。

### Non-goals

- 不兼容、迁移或读取旧生产数据、旧 API、旧 schema、旧配置和旧运行方式；
- 不提供 Linux Desktop、Desktop PostgreSQL、MySQL、SQL Server、Redis、tenant、JWT、Casbin 或 Wails；
- 不授权 push、PR、远端 merge、部署、生产迁移、远端制品发布或生产签名凭据使用；
- 不自动清理成功集成后的 source branch/worktree；
- Goal Plan 不替代 Ticket 的局部实现契约，也不预分配具体 provider、模型或 implementation subagent。

### Authoritative Inputs

| 优先级 | 来源 | 负责内容 | 冲突处理 |
|---|---|---|---|
| 1 | 用户最新明确决定，包括 `Q1A/Q2A` | 产品取舍、worktree 策略与本地执行授权 | 更新真正拥有该决定的工件并暂停受影响 Gate |
| 2 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` 与 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/CONTEXT.md</Path>` | 当前架构决定与领域语义 | 返回 `<Path>{roots.workflows}/specdev/G-grill-with-docs/G-grill-with-docs.md</Path>` |
| 3 | `<Path>{roots.state}/specdev/adr/</Path>` 与 `<Path>{roots.state}/specdev/context/</Path>` | 已毕业永久知识 | 当前 change 替代时在当前 ADR/LOG 明示 |
| 4 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>` | 外部行为、范围和验收合同 | 下游不得改写 |
| 5 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/</Path>` | 单 Ticket 目标、路径、验证和恢复 | Goal Plan 只编排，不扩大 writable paths |
| 6 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>` | DAG、合同覆盖和共享 owner 投影 | 与 Ticket 冲突时以 Ticket frontmatter 为准并修复 Map |
| 7 | `main@888900751ab2b04d2db5fa8d5e6a6b811599ce93` 及当前工作区实测 | 代码、测试、已审计 baseline 和工具可行性 | 新事实冲突时触发 deviation，不静默修改合同 |

## 2. Execution Graph

### DAG and Critical Path

```text
G-BASELINE -> T-01 -> T-02 -> T-03 -> T-04 -> T-06 -> T-07 -> T-14
                       |       |       |       |       |       +-> T-15
                       |       |       |       |       +----------> T-16
                       |       |       |       +-> T-08 ----------> T-16
                       |       |       +-> T-05 -> T-06
                       |       +----------------> T-05
                       |
                       +-> shared contract fan-out

T-07 -> T-09, T-10, T-12, T-13, T-14
T-06 + T-08 -> T-11
T-07 + T-08 -> T-12

T-09 + T-10 + T-11 + T-12 + T-13 + T-15 + T-16
  -> T-17 -> T-18, T-19, T-20 -> T-21
```

关键路径优先级为 `G-BASELINE -> T-01 -> T-02 -> T-03 -> T-04 -> T-06 -> T-07 -> T-14 -> T-16 -> T-17 -> T-19/T-20 -> T-21`。T-14 完成后应优先打开 T-15/T-16；模块池中的其他 Ticket 可以继续并行，但不得占用导致关键路径长期饥饿的全部实现槽位。

### Waves and Ownership

Wave 表示依赖满足后的调度池；同一 Wave 最多同时运行 3 个 implementation owner。Lead 始终串行执行 candidate integration。

| Wave | Ticket | 前置条件 | 项目写路径 | Shared owner | Gate/集成序号 |
|---|---|---|---|---|---|
| W-BASE | baseline checkpoint | 当前改动清单、secret/大文件检查、基线测试完成 | 当前已存在的用户/规划改动 | Lead | G-BASELINE |
| W0 | T-01 | baseline result SHA | 根治理、Task、Hook、分类脚本 | T-01 | G0 / I01 |
| W1 | T-02 | T-01 result | 公共 OpenAPI、合同工具、双方公共 transport | T-02 | G1 / I02 |
| W2 | T-03, T-05 | T-02 result | 后端 runtime；前端 workspace/core | T-03, T-05 | G2 / I03-I04 |
| W3 | T-04 | T-03 result | Database、migration API、Go 依赖 | T-04 | G2 / I05 |
| W4 | T-06, T-08 | T-03/T-04/T-05 对应依赖 | IAM session；reliable-runtime | T-06, T-08 | G3 / I06-I07 |
| W5 | T-07, T-11 | T-06/T-08 对应依赖 | IAM admin；Audit 模块 | 各 Ticket | G3/G4 / I08-I09 |
| W6 | T-09, T-10, T-12, T-13, T-14 | T-07/T-08 对应依赖 | 五个模块独占路径 | 各 Ticket；T-14 最高调度优先级 | G4 / I10-I14 |
| W7 | T-15, T-16 | T-14 result，另满足各自依赖 | Generator；Desktop/sidecar | T-16 拥有 shared sidecar | G4 / I15-I16 |
| W8 | T-17 | T-09 至 T-16 全部依赖 result | product OpenAPI/composition/manifest | T-17 | G5 / I17 |
| W9 | T-18, T-19, T-20 | T-17 result；对应原生 runner 可用 | 三平台独占 release/script/workflow | 各平台 Ticket | G6 / I18-I20 |
| W10 | T-21 | 三平台 Gate 关闭且旧路径 inventory 冻结 | 根 CI/docs/quality 与精确旧路径 | T-21 | G7 / I21 |

### Ticket Quick Reference

| ID | 可观察产出 | Dependencies | Workspace | Implementation owner | E2E disposition | Evidence |
|---|---|---|---|---|---|---|
| T-01 | 根治理与统一任务 | — | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-01` | Lead / execution-time dynamic | not-required：静态治理合同 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-01.md</Path>` |
| T-02 | OpenAPI/生成基座 | T-01 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-02` | Lead / execution-time dynamic | not-required：生成 conformance 覆盖 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-02.md</Path>` |
| T-03 | Profile/observability | T-02 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-03` | Lead / execution-time dynamic | required：真实进程状态 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-03.md</Path>` |
| T-04 | 双方言迁移基座 | T-03 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-04` | Lead / execution-time dynamic | required：真实 PostgreSQL/SQLite | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-04.md</Path>` |
| T-05 | Workspace/App Shell | T-02 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-05` | Lead / execution-time dynamic | required：浏览器导航/交互 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-05.md</Path>` |
| T-06 | 安全登录/账户 | T-03,T-04,T-05 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-06` | Lead / execution-time dynamic | required：Cookie/CSRF/Session | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-06.md</Path>` |
| T-07 | IAM 授权管理 | T-06 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-07` | Lead / execution-time dynamic | required：UI/API 权限矩阵 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-07.md</Path>` |
| T-08 | Outbox/唯一 executor | T-03,T-04 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-08` | Lead / execution-time dynamic | required：多进程故障恢复 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-08.md</Path>` |
| T-09 | Organization | T-07 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-09` | Lead / execution-time dynamic | required：API/UI/引用冲突 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-09.md</Path>` |
| T-10 | Settings | T-07 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-10` | Lead / execution-time dynamic | required：API/UI/secret 拒绝 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-10.md</Path>` |
| T-11 | Audit | T-06,T-08 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-11` | Lead / execution-time dynamic | required：事件到审计 UI | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-11.md</Path>` |
| T-12 | Scheduler | T-07,T-08 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-12` | Lead / execution-time dynamic | required：多实例/停止 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-12.md</Path>` |
| T-13 | Files | T-07 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-13` | Lead / execution-time dynamic | required：真实文件系统 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-13.md</Path>` |
| T-14 | Demo tracer | T-04,T-05,T-07 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-14` | Lead / execution-time dynamic | required：双方言 Web CRUD | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |
| T-15 | Generator | T-14 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-15` | Lead / execution-time dynamic | required：真实生成/编译 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-15.md</Path>` |
| T-16 | Tauri 2 Desktop | T-06,T-08,T-14 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-16` | Lead / execution-time dynamic | required：原生 sidecar/Stronghold | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-16.md</Path>` |
| T-17 | 完整产品组合 | T-09,T-10,T-11,T-12,T-13,T-15,T-16 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-17` | Lead / execution-time dynamic | required：三 profile/双 App | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-17.md</Path>` |
| T-18 | Linux OCI/Compose | T-17 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-18` | Lead / execution-time dynamic | required：双架构容器 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-18.md</Path>` |
| T-19 | macOS Universal DMG | T-17 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-19` | Lead / execution-time dynamic | required：签名/公证/安装 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-19.md</Path>` |
| T-20 | Windows x64 NSIS | T-17 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-20` | Lead / execution-time dynamic | required：签名/安装/卸载 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-20.md</Path>` |
| T-21 | 原子切换/归零 | T-18,T-19,T-20 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-21` | Lead / execution-time dynamic | required：最终全矩阵 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` |

## 3. Gates and Completion Evidence

### Overall Definition of Done

- 所有非 cancelled Ticket 都完成 source commit -> parent-candidate verification -> parent result SHA；
- shared owner 先稳定并集成，消费者 worktree 只基于包含该 result 的父分支 checkpoint 创建或刷新；
- normal/failure/regression、架构边界、双方言迁移、缓存禁用、Session/权限、Desktop 和原生发行 Evidence 全部由 Lead 核对；
- 任何 required E2E 只在 parent-candidate 运行，source worktree 的 E2E 声明不计入；
- T-21 contract 后新产品全量回归通过，零兼容扫描成立，没有未决高影响偏差；
- 远端 reconcile 未授权时不阻止本地实现完成，但必须在归档前明确 reconcile 或 waive。

### Gates

| Gate | 开启条件 | 关闭证据 | 阻塞范围 | Lead/批准人 | 失败恢复 |
|---|---|---|---|---|---|
| G-BASELINE | Q1A/Q2A 已记录；当前 workspace 实测 | Lead 审计 170 tracked/6 untracked、secret/大文件/生成物边界；基线测试通过；在 `main` 形成 baseline SHA | 所有 Ticket worktree | Lead；授权来自用户 Q2A | 不提交可疑内容；保留现状，修正 inventory 后重试 |
| G0 Root Command | baseline SHA 稳定 | T-01 result、任务合同、Hook 委派和治理扫描 | T-02 及全部下游 | Lead | 保留 T-01 source，修复后重建 candidate |
| G1 Contract Stable | T-01 result | T-02 result、OpenAPI lint/bundle/generate/conformance、clean-tree | T-03/T-05 及所有合同消费者 | Lead | 暂停消费者；只在 T-02 owner 修订共享合同 |
| G2 Runtime/Data/UI Foundations | T-02 result | T-03/T-05 result；随后 T-04 双方言/锁/迁移 result | IAM、模块、Desktop | Lead | 回到失败 owner；已通过父结果不回退 |
| G3 Identity/Reliability | G2 关闭 | T-06/T-08 result 与 E2E；T-07 授权矩阵 result | 全部业务模块 | Lead | revoke 测试会话、停 dispatcher，修复后重跑候选 |
| G4 Vertical Modules | G3 关闭并满足各 Ticket 依赖 | T-09 至 T-16 各自 result；模块 API/UI/双方言/故障/原生证据 | T-17 | Lead | 单模块回退到 source；其他已集成模块保持父结果 |
| G5 Product Integration | G4 全部依赖 result | T-17 三 profile、双 App、权限、架构、合同、缓存 E2E 和 result | 三平台发行 | Lead | 父分支不推进；修复 T-17 composition 或返回模块 owner |
| G6 Protected Release | T-17 result；对应 runner/批准可用 | T-18/T-19/T-20 result，原生安装、签名/公证、SBOM/provenance | T-21 | Lead + 受保护发行批准人 | 未授权凭据/远端动作保持 pending；不伪造发行 pass |
| G7 Atomic Contract | G6 三平台证据完整；旧路径 inventory 冻结 | T-21 result、全量 root Gate、allowlist 外零兼容命中、删除清单 | change completion | Lead；删除范围来自已批准 T-21 | candidate 失败时父分支不动；修正 source 或整体放弃候选 |

### Contract and Reference Coverage

| 合同或参考要求 | 覆盖 Ticket | 验证接缝 | Evidence | 状态 |
|---|---|---|---|---|
| 根治理、统一任务、分层 CI | T-01,T-18,T-19,T-20,T-21 | Task/Hook/CI policy 与平台 Gate | 各 Ticket Evidence | planned |
| OpenAPI 3.1、稳定错误、生成 conformance | T-02,T-06,T-15,T-17,T-21 | lint/bundle/generate/negative/clean-tree | T-02/T-15/T-17/T-21 | planned |
| 三 profile、双方言迁移、无缓存事实源 | T-03,T-04,T-08,T-14,T-17 | profile/migration/cache-disabled matrix | T-03/T-04/T-08/T-14/T-17 | planned |
| Session、Permission Code、数据范围 | T-06,T-07,T-16,T-17 | Web/Desktop security 与权限矩阵 | T-06/T-07/T-16/T-17 | planned |
| 八业务模块完整能力 | T-07,T-09,T-10,T-11,T-12,T-13,T-14,T-15,T-17 | module API/UI/failure/E2E | 对应模块 Evidence | planned |
| Tauri 2、Stronghold、sidecar、Desktop SQLite | T-16,T-17,T-19,T-20 | native tracer 与安装包 Gate | T-16/T-17/T-19/T-20 | planned |
| Linux/macOS/Windows 发行与供应链 | T-18,T-19,T-20 | native/container protected Gate | T-18/T-19/T-20 | pending external approval |
| 零兼容与最终架构 | T-15,T-17,T-21 | architecture/compatibility scan | T-15/T-17/T-21 | planned |

详细 `AC-001` 至 `AC-036` 映射以 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>` 为投影权威。

## 4. Execution and Integration Protocol

### Lead Orchestration

| 项目 | 决定 | 事实依据 |
|---|---|---|
| Lead | `codex-root` | 唯一 SpecDev 状态、Evidence、E2E 与父分支 owner |
| Implementation subagents | 最多 `3`，Lead 不计入 | config 上限 3、平台 4 个总槽位、用户选择 Q1A |
| Integration attempts | 每个候选最多 `3` 次 | config 快照 `max_integration_attempts=3` |
| Read-only agents | 无 SpecDev 数字上限 | review/research/test-observation，只读且不竞争可变环境 |
| Dispatch | execution-time dynamic | provider、模型、自行实现或派单按 Ticket 事实选择，不静态预分配 |
| SpecDev writes | 仅 Lead | subagent 不写 Ticket/Map/Goal Plan/Evidence/status |
| E2E | 仅 Lead | implementation 返回只含非 E2E 候选事实 |

动态 Dispatch Packet 遵循 `<Path>{roots.workflows}/specdev/common/skills/subagent-delivery/SKILL.md</Path>` 的 `operation=plan` 合同：允许 `implementation/review/research/test-observation`；每次 dispatch 固定 Ticket、依赖 Evidence、base SHA、workspace、路径、授权、非 E2E 检查、停止条件和返回格式。外部 provider 需要另行的数据发送授权；本计划不授权 external-web 源码交付。

### Ticket Workspace and Integration

所有 Ticket 使用 `required` 策略。每个 source branch/worktree 基于开始时最新父分支 result SHA；source worktree 不运行 E2E。implementation owner 只写 Ticket frontmatter 允许的项目路径并形成非空 source commit。Lead 冻结 `parent_before_sha`，在 Lead-owned candidate 中集成 source checkpoint，运行集成检查和 required E2E，确认父 HEAD 未漂移后才更新 `main` 并记录 `result_sha`。

| Ticket | Parent/base | Workspace/branch | Source checks | Implementation commit | Integration checks/E2E | Parent result |
|---|---|---|---|---|---|---|
| T-01 | baseline result | worktree T-01 / `specdev/T-01` | governance/task/static | required non-empty | candidate governance；E2E N/A | passed candidate SHA |
| T-02 | T-01 result | worktree T-02 / `specdev/T-02` | lint/generate/conformance | required non-empty | candidate clean-tree；E2E N/A | passed candidate SHA |
| T-03 | T-02 result | worktree T-03 / `specdev/T-03` | Go unit/race/static | required non-empty | process startup/status E2E | passed candidate SHA |
| T-04 | T-03 result | worktree T-04 / `specdev/T-04` | Go/migration fixtures | required non-empty | PostgreSQL/SQLite/lock E2E | passed candidate SHA |
| T-05 | T-02 result | worktree T-05 / `specdev/T-05` | unit/type/lint/build | required non-empty | browser shell E2E | passed candidate SHA |
| T-06 | latest deps result | worktree T-06 / `specdev/T-06` | IAM unit/API/security | required non-empty | Cookie/CSRF/Session E2E | passed candidate SHA |
| T-07 | T-06 result | worktree T-07 / `specdev/T-07` | IAM permission/API/UI unit | required non-empty | browser/direct API matrix | passed candidate SHA |
| T-08 | latest deps result | worktree T-08 / `specdev/T-08` | outbox/race/fault unit | required non-empty | multi-process recovery E2E | passed candidate SHA |
| T-09 | T-07 result | worktree T-09 / `specdev/T-09` | module unit/API/type | required non-empty | Organization UI/API E2E | passed candidate SHA |
| T-10 | T-07 result | worktree T-10 / `specdev/T-10` | module unit/API/type | required non-empty | Settings UI/API E2E | passed candidate SHA |
| T-11 | latest deps result | worktree T-11 / `specdev/T-11` | audit/redaction/retry | required non-empty | event-to-UI E2E | passed candidate SHA |
| T-12 | latest deps result | worktree T-12 / `specdev/T-12` | clock/registry/unit | required non-empty | multi-worker/UI E2E | passed candidate SHA |
| T-13 | T-07 result | worktree T-13 / `specdev/T-13` | file security/unit | required non-empty | filesystem/API/UI E2E | passed candidate SHA |
| T-14 | latest deps result | worktree T-14 / `specdev/T-14` | Demo dialect/unit/type | required non-empty | PostgreSQL/SQLite Web E2E | passed candidate SHA |
| T-15 | T-14 result | worktree T-15 / `specdev/T-15` | golden/type/compile | required non-empty | generate-to-temp E2E | passed candidate SHA |
| T-16 | latest deps result | worktree T-16 / `specdev/T-16` | Go/Rust/TS/build | required non-empty | native Desktop E2E | passed candidate SHA |
| T-17 | all module results | worktree T-17 / `specdev/T-17` | full non-E2E gates | required non-empty | three-profile/two-App E2E | passed candidate SHA |
| T-18 | T-17 result | worktree T-18 / `specdev/T-18` | OCI/Compose policy/build | required non-empty | native/container release Gate | passed candidate SHA |
| T-19 | T-17 result | worktree T-19 / `specdev/T-19` | macOS policy/build | required non-empty | protected macOS Gate | passed candidate SHA |
| T-20 | T-17 result | worktree T-20 / `specdev/T-20` | Windows policy/build | required non-empty | protected Windows Gate | passed candidate SHA |
| T-21 | all release results | worktree T-21 / `specdev/T-21` | inventory/zero/full static | required non-empty | final full E2E and zero scan | passed candidate SHA |

Candidate 集成队列严格串行。若父分支已包含 source checkpoint，可 fast-forward；否则使用独立 merge commit。候选验证失败时父分支保持不变，同一 source worktree 修正；父 HEAD 漂移则 candidate 标记 stale 并从最新父结果重建。成功集成不自动清理 source branch/worktree。

### Authorization Matrix

| 动作 | 状态 | 目标与条件 |
|---|---|---|
| Current workspace Ticket changes | not-authorized | 本计划使用 required，不在当前 workspace 实现 Ticket |
| Ticket worktree local changes | allowed | Q1A/Q2A；仅对应 Ticket writable/shared-owner 路径 |
| Implementation commit | allowed | Q2A；包含经审计 baseline commit 及 T-01 至 T-21 非空 source commit |
| Local direct-parent verification and parent update | not-authorized | 不适用 required 策略 |
| Local candidate integration and parent update | allowed | Q2A；Lead-only、通过 candidate checks/E2E 且父 HEAD 未漂移 |
| Push / PR / remote merge | not-authorized | 本地授权不继承到远端 |
| Branch/worktree cleanup | not-authorized | source worktree 集成后保留，另行确认 |
| Deploy / migration / production actions | not-authorized | 仅允许隔离测试数据库/fixture；生产目标逐动作授权 |
| Production signing/notarization/release credentials | not-authorized | G6 可实现流程和本地非生产检查；正式 Gate 等待受保护批准 |

### Evidence Return

原生 implementation 返回必须包含 Ticket ID、workspace locator、base/source commit、dirty 状态、实际修改路径、非 E2E 命令/结果、未运行项和恢复条件。Lead 重新读取 worktree、检查 diff/commit/路径和测试，再写 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/{ticket-id}.md</Path>`。subagent 自报的 Evidence、截图或 E2E pass 均保持 unverified，直到 Lead 在 candidate 复核。

## 5. Constraints, Risk and Recovery

### Non-negotiable Constraints

- 根目录和模块/前端/发行边界严格遵守 Spec、ADR 与 Ticket writable paths；共享合同只有唯一 owner；
- Greenfield 合同，不引入旧 API/schema/config/data shim、双写或 compatibility alias；
- Server PostgreSQL/SQLite、Desktop SQLite；Bun + database/sql、Goose、无 AutoMigrate；
- OpenAPI 3.1 唯一 transport 事实源；生成类型与领域模型显式 mapping；
- 数据库 Session、Web Cookie+CSRF、Desktop Stronghold proxy、Argon2id、Permission Code/data scope；
- 无 Redis、JWT、refresh token、Casbin、tenant、Wails、MySQL、SQL Server；
- 实际依赖/工具版本在 owner Ticket 中以 lockfile、Go module、Cargo lock 或固定 digest 精确锁定，禁止 `latest` 浮动；
- implementation subagent 不写 SpecDev 工件/父分支，不运行 required E2E，不访问未授权 secret/远端/生产系统。

### Verification Integrity

基线于 2026-08-26 在当前 workspace 实测：SpecDev Tickets validator `0 error/0 warning`；Go 全量测试通过；前端 `28` 个文件、`225` 个单元测试通过；typecheck 通过；lint `0 error/25 warning`；后端与前端 production build 通过。25 个 warning 属于待 T-05/T-21 消除的旧结构基线，不得新增或把 error 降级为 warning。

source worktree 只运行 Ticket 非 E2E 检查；任何 source worktree E2E pass 无效。Lead 在 parent-candidate 运行受影响集成、回归和 required E2E。禁止以跳过测试、放宽 lint/type、更新 golden 而不审查、mock 替代真实 PostgreSQL/SQLite/原生安装、忽略非零退出码或只检查文件存在制造伪绿色。

### Migration or Release Sequence

1. **Baseline：** 审计并提交当前用户改动和规划工件，记录父分支基线 SHA；不纳入 secret、临时产物或未解释大文件。
2. **Expand：** T-01 至 T-05 建立根治理、公共合同、runtime/database/frontend 新结构，旧产品保持未发布只读参考。
3. **Vertical migrate：** T-06 至 T-16 逐模块迁入新结构，每个 source/candidate 独立可验证，不做旧数据/API 双轨。
4. **Observe/integrate：** T-17 聚合三 profile 与双 App，关闭共享合同、迁移、权限和可靠性 Gate。
5. **Release verify：** T-18/T-19/T-20 生成各平台实现与本地证据；正式签名、公证、远端/受保护 runner 动作等待独立批准。
6. **Contract：** 三平台 Gate 完整后，T-21 精确删除旧路径并完成零兼容扫描；不在 earlier Wave 提前删除。

### Risks, Monitoring and Recovery

| 风险 | 监控/判卷接缝 | 恢复 |
|---|---|---|
| dirty baseline 混入 secret/生成垃圾 | staged/unstaged inventory、secret/large-file scan、基线 diff | 不提交；保留原文件并修订纳入清单 |
| shared contract 漂移或并行冲突 | owner/path scan、bundle/lock clean-tree | 暂停消费者，返回唯一 owner source 修正 |
| 双方言 schema/行为分叉 | 空库/上一版/幂等 migration matrix | 父分支不推进；保留备份，用前向 migration 修复 |
| Session/权限/文件安全退化 | security matrix、敏感值和路径逃逸测试 | revoke 测试会话、停受影响功能，修复 candidate |
| Outbox/Scheduler 重复副作用 | fault injection、active executor、幂等键 | 停 dispatcher/executor，保留 pending 并幂等重放 |
| Desktop 开窗前失败或凭据泄露 | native launch/Stronghold/JS tracer | 不开窗，保留 DB/备份，修复 sidecar/adapter |
| 原生 runner/凭据不可用 | G6 authorization 与 platform evidence | 保持 Gate pending，不用模拟证据关闭 |
| T-21 删除过宽 | 精确 inventory、candidate diff、zero/full regression | 父分支不推进；丢弃候选或修正 source commit |
| 父 HEAD 漂移/候选冲突 | integration queue 重读 parent HEAD | 标记 stale，最多 3 次重建；之后 blocked |

### Deviation Control

遵循 `<Path>{roots.workflows}/specdev/common/rules/deviation-control.md</Path>`。任何会改变行为、合同、schema、安全、范围、迁移、发行或验收的事实立即暂停受影响 Wave，记录 deviation 并返回 Spec/ADR/Ticket owner；局部实现细节只要不改变外部合同且在 writable paths 内，可由 implementation owner选择并在返回中说明。

## 6. Progress and Decisions

### Current Status

- Goal Plan 已选择 `required/candidate-merge`，Lead 为 `codex-root`，implementation agent 上限 3，integration attempt 上限 3。
- G-BASELINE 已完成：原始 170 项 tracked 改动和 6 个 untracked 条目经 secret/大文件/生成物审计后，形成 `main@888900751ab2b04d2db5fa8d5e6a6b811599ce93` 不可变基线。
- 基线 Go test、前端 unit/type/lint/build 与 Goal Plan validator 已实测通过，validator 结果为 0 error / 0 warning。
- `G0 Root Command` 已关闭：T-01 source 为 `a19199c251a140db399a31fcb3a8ddb7aac4e61e`，parent before 为 `464a4d2c7ef5542b85ec6512837dbb47a6079b20`，通过验证并提升的 candidate/result 为 `249571939f7497fd7c216e6b5985847c0b941e2f`（tree `dad0853a660bc1d63fbe34cb8da758fc302d62a8`）。
- `G1 Contract Stable` 已关闭：T-02 从包含 T-01 result 的 `main@9fb27f69e79b7acb84bf65a7328b2254d2d20359` 创建 source，最终 source 为 `201b4647e15da1e62a3c92012c9847ec9765b706`，通过验证并提升的 candidate/result 为 `9189b89ccc046704166a6e9aa98d0afba1399a36`（tree `e4f8f7307844c7a69e3be28979cc3b8c31edb7fc`）。当前进入 `G2 Runtime/Data/UI Foundations`，T-03/T-05 已解锁。
- T-02 预检登记并批准 Ticket 级偏差 `T02-D01`：为实现既定根验证命令和固定官方生成器，串行补充两个 Task、Go module 生成依赖与 pnpm contract lock；T-03/T-05 尚未开始，无并行 writer 冲突。
- T-02 首轮 source checkpoint 为 `18c10920d0aea9a7c257aa829772505ba0d506e5`；完成首轮审查修正后形成 checkpoint `f9c28705b3d8e485059c43806f6036c459c7ec18`。规范轴复审已 pass；标准轴复审发现 manifest 删除许可与 module path 语法不一致、敏感信息测试在解码后读取空缓冲区两项 medium finding。修正后 checkpoint `c8c35511673ebcdc51f8f79be5f172b16b73c77a` 的标准轴 pass，规范轴发现 Go 最短路径可复用 owner/transport segment 的一项 high finding；增加独立 owner/transport 段约束及 metadata/manifest 拒绝回归后形成干净 checkpoint `201b4647e15da1e62a3c92012c9847ec9765b706`。当前合同 lint、23 个正负向/原子生成工具测试、合法嵌套路径幂等与伪装路径拒删、确定性双方生成、Go race/runtime request conformance、TS client 19 项测试、普通与 SQLite Go suite、前端 225 项测试/lint/build、依赖锁与 peer 完整性检查均通过，等待双轴复审与 parent-candidate Gate。
- T-02 首轮双轴审查在固定点 `9fb27f69e79b7acb84bf65a7328b2254d2d20359...18c10920d0aea9a7c257aa829772505ba0d506e5` 返回 request-changes；Lead 登记并批准 `T02-D02`（最终 UI owner 路径与多 fragment owner 元数据）和 `T02-D03`（canonical generator 接入根 `generate`），其余 finding 在 T-02 source 内修复后形成新 checkpoint 并重审。
- T-02 修复中实测公开 `generate` 的旧 `swag` 链因未受管工具缺失而失败；`T02-D03` 随事实收紧为公开命令只委派 canonical generator，旧脚本不删除但退出公共命令面。
- T-02 最终双轴审查已对固定范围 `9fb27f69e79b7acb84bf65a7328b2254d2d20359...201b4647e15da1e62a3c92012c9847ec9765b706` 返回 pass、0 findings；Lead parent-candidate G1 全门禁通过，父分支已提升至 result，Ticket 状态为 `done`。
- G2 已从 `main@27a186cbda7e77c403b3c64074260a10e924ee92` 并行创建互不重叠的 T-03/T-05 source worktree；implementation owner 分别为 `codex-t03-runtime` 与 `codex-t05-frontend`，Lead 保留状态、required E2E 与 candidate integration 所有权。
- T-05 implementation owner 触发并由 Lead 批准 Ticket 级偏差 `T05-D01`：初始路径合同漏列真实 app-shell workspace package 必需的 manifest，只新增 `<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>`，不扩大其他 app-shell 写入范围。
- T-05 安装验证暴露 T-01 根治理遗漏：根 `<Path>.gitignore</Path>` 只覆盖旧前端名称。Lead 在 T-01 已有可写范围内以 `63ea213008dd8fee591498d2566f15d1a3cf574c` 补齐 `<Path>go-admin-plus-ui</Path>` 的依赖、构建和测试产物规则；旧名称规则按 expand/migrate/contract 合同保留至 T-21。
- T-05 implementation owner 已返回 clean source checkpoint `dd9c816ae8e0cddeb5160a6480a476459bf59384`；固定审查范围为 `27a186cbda7e77c403b3c64074260a10e924ee92..dd9c816ae8e0cddeb5160a6480a476459bf59384`，source 非 E2E workspace/lint/type/test/boundary/build checks 全部通过，required browser E2E 保持 pending。
- T-05 规范轴初审发现 Runtime Adapter 被置于 Admin Web，且 Ticket 漏列 ADR-011 锁定的 `<Path>packages/adapters/{browser,desktop}</Path>`。Lead 批准 `T05-D02` 精确新增 browser adapter package 与 desktop manifest；旧 source checkpoint 失效，修正后的 checkpoint 将重新固定并执行双轴审查。
- T-03 implementation owner 已返回 clean source checkpoint `48a97ff77f0f07501fe7492a6fa4aaee59255f0c`；固定审查范围为 `27a186cbda7e77c403b3c64074260a10e924ee92..48a97ff77f0f07501fe7492a6fa4aaee59255f0c`，12 个文件均在授权路径内，source Go 普通/SQLite/race/vet/build 与 mod verify 全部通过，required process E2E 保持 pending。
- T05-D02 修正后 implementation owner 返回新的 clean source checkpoint `89683cc1e7defdf5cfcc5564e62a43e6296787c0`；browser runtime adapter 已迁入公开 workspace package，Desktop 仅预建 manifest，Admin Web 不再拥有 transport。双轴审查重新固定为 `27a186cbda7e77c403b3c64074260a10e924ee92..89683cc1e7defdf5cfcc5564e62a43e6296787c0`。
- T-03 规范轴初审 pass；标准轴对 `27a186c..48a97ff` 返回 fail：1 high（公开 config input/Desktop material 可被 JSON/slog 结构化序列化泄密）、2 medium（非 GET 在 405 前执行 dependency probe；并发安全合同缺少实际并发/race 场景）。旧 checkpoint 失效并退回原 owner 修正。
- T-05 规范轴在 T05-D02 后 pass；标准轴对 `27a186c..89683cc` 返回 fail：1 high（写入成功后 refresh 失败误报为写失败，可能诱发重复写）、2 medium（乱序导航覆盖新状态；identity 响应未知 credential 字段未 fail closed）、1 low（列表 stale request/error 语义未固定）。旧 checkpoint 失效并退回原 owner 全部修正。
- T-03 原 owner 已闭合全部标准轴 findings 并返回 clean source checkpoint `3801d776001b00d9d2a63e7343ca2f319986848a`；结构化日志脱敏、method-before-probe 和并发生命周期回归均转绿，双轴复审重新固定为 `27a186c..3801d77`。
- T-03 双轴复审对固定范围 `27a186c..3801d77` 返回 pass、0 findings；Lead parent-candidate `cfae1e71cec1d123bc6ce0fca1d8f92938db78f2` 无冲突通过 Go/合同/生成/治理全门禁以及 required process E2E，父分支已快进提升至该 result，Ticket 状态为 `done`。
- T-05 原 owner 已闭合全部标准轴 findings 并返回 clean source checkpoint `5b7b4f43298a6b28ab55190baca07e05ffa50911`；mutation 两阶段结果、导航 sequence/abort/invalidate、identity 精确白名单和 stale list request 语义均转绿，双轴复审重新固定为 `27a186c..5b7b4f4`。
- T-03 集成解除 T-04 阻塞；Lead 已从 `main@fe56a1932d8b9cec8db5f5dd7e24df33ab10853f` 创建 T-04 source worktree，implementation owner 为 `codex-t04-database`，Lead 保留真实 PostgreSQL/SQLite required E2E 与 candidate integration 所有权。
- T-04 implementation owner 已返回 clean source checkpoint `a841a5b97b37236883a9e5f1311784b77ab4eab1`；Bun/Goose、pgx、modernc SQLite、双方言 Provider、只前进事务迁移、SQLite 文件锁与 cache-disabled 基座已通过 source 普通/SQLite/race/vet/build、`CGO_ENABLED=0` 和 module verification。真实 PostgreSQL 与跨进程 SQLite 文件锁仍保持 required E2E pending，当前进入固定点双轴审查。
- T-04 双轴初审对 `fe56a193..a841a5b` 均返回 fail。标准轴发现 3 high（PostgreSQL 多副本 migration 缺数据库级锁、SQLite connection-local PRAGMA 在换连接后失效、Goose annotation 预检与实际 parser 不等价）和 1 medium（脱敏错误丢失 context cancellation sentinel）；规范轴重复确认 SQLite high，并发现 1 medium（无界 `cache.Memory` 违反 ADR-019 且与 T-08 owner 重叠）。checkpoint 已退回原 owner 修正，candidate 尚未创建。
- T-04 原 owner 已返回修正后的 clean checkpoint `69b5ba1b1c091cf6ce887cb96f9abc9ffc10a15a`：Goose PostgreSQL session advisory lock、modernc 每连接 PRAGMA、annotation 等价预检、脱敏 context sentinel 和无界 cache 删除均有回归；Lead 聚焦 unit/race 与强制换连接检查通过，当前固定该范围执行双轴复审。
- T-04 对 `fe56a193..69b5ba1` 的双轴复审确认首轮 findings 已闭合，但继续返回 fail：两轴共同发现 Windows drive/UNC SQLite URI 编码 high；标准轴发现 Up 前 SQL 未被预检拒绝的 medium；规范轴发现允许 ENVSUB 破坏 migration 确定性的 medium。该 checkpoint 再次退回原 owner，candidate attempt 仍为 0。
- T-04 原 owner 以 `5f359381b7a3db09a5bd64e147ce4b22b80b0cda` 完成第二轮修正：Windows drive、空 authority UNC、POSIX 特殊字符/反斜杠 URI，modernc 实际解析、Up 前 SQL 与 ENVSUB fail-closed均转绿。最终标准轴与规范轴对 `fe56a193..5f359381` 均 pass、0 findings，Lead 聚焦 unit/race复核通过，允许进入首次 parent-candidate。
- T-04 merge/result 固定为 `8757554047a828e12712a1faf2d547412bcdf05b`（tree `ada5ca6d19d02441b33dfcbc8e4c0535ee236ae9`）。Lead 创建 candidate 后误在根 workspace merge，立即终止并作废旧 parent 门禁、冻结 `main`，将独立 candidate 快进到同一 SHA/tree 后从零重跑；全 Go/前端/合同/治理/跨编译以及隔离 PostgreSQL 18.3 三 profile/previous/idempotent/双 pool advisory lock、SQLite 跨进程锁 E2E 全部 pass。该 execution incident 不改变代码或验收合同并已写入 Evidence；T-04 状态为 `done`，G2 关闭。
- G2 关闭后，Lead 从最新 `main@05d619d1c4c90378a0138c630fa1aa3bcfa0f942` 并行创建互不重叠的 T-06/T-08 source worktree，implementation owner 分别为 `codex-t06-iam` 与 `codex-t08-reliability`。两者只运行 source 非 E2E checks 并返回非空 clean commit；Cookie/CSRF/browser 与 PostgreSQL/SQLite 多进程 required E2E、candidate integration 仍由 Lead 保留。
- T-06 preflight 在零写入状态发现共享合同实际仍以旧 `<Path>go-admin-ui-plus/</Path>` 为 canonical client 根，且共享生成 manifest 会被每个模块 fragment 改写；这与 T-02 权威路径、ADR-002、T-05 result 和并行所有权冲突。Lead 批准 `T02-D04` 并重新打开 T-02：先在原 T-02 source worktree 合并最新 parent，迁移 canonical client/tool root，令共享 manifest 只管理 canonical outputs，再经新 candidate/result 关闭 G1。同期批准 `T06-D01` 精确补两个 IAM package manifest 与新 Workspace lock；T-06 保持 clean/blocked，待 T02-D04 result 后刷新 base，T-08 独立继续。
- T02-D04 source 的合同/类型/单包测试通过后，全 Workspace 回归发现 T-05 boundary test 将当时 package 数量硬编码为 `23`，合法加入规划内 API client 即失败。Lead 批准 `T05-D03` 并重新打开 T-05，只把该断言修正为“必需 package 集合 + 对全部发现 package 逐一验证”；先独立集成 T-05 correction，再刷新 T-02 source parent 并继续 G1 candidate。
- T-08 implementation owner 已返回 clean source checkpoint `f6499cdccdbf45449b9761b65e9e31c421220405`（base `05d619d1c4c90378a0138c630fa1aa3bcfa0f942`）：事务 Outbox、lease/retry/receipt 幂等、dispatcher 退避/失锁停止、PostgreSQL 专用 advisory 连接、SQLite 单 executor 和 bounded TTL/LRU cache 均完成 source 非 E2E 门禁；未改依赖清单或越界路径。当前固定该范围进入双轴审查，真实 PostgreSQL/SQLite 多进程 E2E 仍仅由 Lead candidate 执行。
- T05-D03 双轴审查对 `12171f0..832aede` 均 pass、0 findings；Lead candidate 以 fast-forward 固定为 `832aede2541ac8dc0ee02c587c59d709889b5aad`（tree `6cf2e4f63211a2b934881a495e176878c4b2a1eb`），frozen install/peers、前端 verify、Go normal/SQLite/vet/mod、合同串行 23/23、治理和 SpecDev 全绿。运行资产相对原 T-05 result 零差异，复用此前真实浏览器 E2E；`main` 已提升到同一 result，T-05 再次 done。
- T02-D04 已在原 source worktree 合并最新 `main@253bc23e28b9cdd20c518585da51812404873538` 并形成 clean checkpoint `24ff115a097305d0577e8d61d7f9dc306187dc81`。canonical API client、生成运行根和公开 Task 已迁至新 Workspace，共享 manifest 只保留 5 个 canonical outputs，每个模块使用自身 Go transport 目录内 manifest 管理精确 5 个 owner outputs；全 Workspace verify、合同 26 项正负向/原子/并发生成测试、API client 7 项、Go 普通/SQLite/vet/race、frozen lock/peer、治理与 Goal Plan validator 均通过。当前固定审查范围为 `253bc23e28b9cdd20c518585da51812404873538..24ff115a097305d0577e8d61d7f9dc306187dc81`；本机未安装 Go Task CLI，故 `task:contract` 仅记录为环境未运行，不安装新系统工具或伪造结果。
- T02-D04 checkpoint `24ff115a097305d0577e8d61d7f9dc306187dc81` 已作废：Lead 隔离探针确认生成目标祖先为 symlink 时会写出 output root；规范轴另确认 fragment 与 manifest 同时删除时，孤儿生成物不进入 stale 集合。两项均不改变合同或路径范围，已退回同一 T-02 source 增加物理路径 fail-closed 与受管理孤儿扫描回归；新 checkpoint 形成前 T-02 保持 `in_progress`。
- T-08 owner 已在原 source 闭合首轮双轴全部 findings 并返回 clean checkpoint `d2311880e7f0b597f5ad4d470ffa316ae47e430d`：固定 advisory lease、真实 Lease-only Dispatcher、数据库内声明式 mutation、claim token/owner/expiry fencing、失锁回滚停止、递归敏感字段拒绝、脱敏 observer、失败时钟 retry、微秒幂等和 cache generation/metrics 均具备聚焦回归；普通/SQLite/全 race/vet/CGO=0/build/module verify 通过。当前对 `05d619d1c4c90378a0138c630fa1aa3bcfa0f942..d2311880e7f0b597f5ad4d470ffa316ae47e430d` 重新执行双轴审查，required 多进程 E2E 仍 pending。
- T02-D04 已闭合 symlink/junction output escape 与无 manifest orphan 漏检：统一 resolver 对写入、检查和 stale 删除执行 lexical root、逐段 `lstat` 与最近存在父路径 `realpath` 证明，外部 sentinel 保持不变；受管理 Go transport 与新 UI domain `src/**/generated` 扫描会发现孤儿且不进入 pnpm `node_modules`。合并最新 Lead 状态后的 clean checkpoint 为 `6fab034640672fef057acad61139aa8b917a85a1`，固定范围 `faa451b47b59bb5c207fe1dcdbcb39bc8474cacd..6fab034640672fef057acad61139aa8b917a85a1`；合同 28/28、lint/generate check、API client typecheck/7 tests 通过，进入新一轮双轴复审。
- T-08 checkpoint `d2311880e7f0b597f5ad4d470ffa316ae47e430d` 的双轴复审再次 request-changes。固定修复合同为：以结构化 mutation 描述替代任意 SQL+黑名单，显式限制非平台 owner 表、操作、业务键、参数和影响行数，并拒绝 Unicode quoted identifier 绕过 advisory 锁；将 Claim/Deliver/Retry 收回不可由 API replica 直接调用且绑定同一 Database/owner 的真实 Lease 状态机；以 topic payload schema/allowlist 和非敏感稳定 BusinessKey 类型取代字段名猜测；补 pending age、claim duration、attempt/retry、active executor/lost-lock 指标，并令探测失锁立即分类。旧 checkpoint 作废，required E2E 未开始。
- T-05 双轴复审对 `27a186c..5b7b4f4` pass、0 findings；首次 parent-candidate `f2bbd5211eada20b8ddc3762cba680bc7944ceae` 的 pnpm 全门禁和功能浏览器断言通过，但 Lead 截图检查发现 390px 视口导航隐式 grid row 被拉伸至近半屏。required visual E2E 判定失败，candidate 未提升并已清理，checkpoint 退回原 owner 修正并补导航高度断言。
- T-05 原 owner 以 `a648baf851d25861ac04eef688b8856fcf439b71` 修正移动 grid rows 并增加静态样式合同；source `pnpm verify` 全绿且 clean，最终双轴复审固定为 `27a186c..a648baf`，第二次 candidate 将把 390px nav 高度上限纳入 required browser E2E。
- T-05 第二次 candidate frozen install 暴露根 `.gitignore` 仍漏掉 pnpm 子 package `node_modules` symlink。Lead 在 T-01 已有所有权内以 `3bc0dd5fe13a6e215caf6f95781482152ccfc279` 补齐 `<Path>go-admin-plus-ui/**/node_modules/</Path>`；该父分支纠正将合入当前 candidate 后继续 clean-tree/browser 门禁。
- T-05 第二次 parent-candidate 已在纳入最新父分支治理修正后固定为 `d5d10dda68fb6b49850c7442bf036b195ba1d202`（tree `0511250c439a3233d9217fb41de598dfdc17dadb`）。24-workspace frozen install、lint/type、22 项 unit、6 项 boundary、build、合同生成、治理、SpecDev 校验与真实 Chromium desktop/mobile E2E 全部通过；390px 导航高度为 71px 且无溢出。`main` 已从 `c2260191c86942f859088d15b9a3ba9af96c975b` 以 `--ff-only` 晋升到同一 result，T-05 状态为 `done`。
- implementation commit 和 Local candidate integration and parent update 已由用户 `Q2A` 授权；source cleanup、远端和生产动作未授权。

### Pending Decisions and Blockers

- 无阻止 T-02 开始的产品决定。
- G6 的生产签名、公证、受保护 runner、远端制品发布仍需到达该 Gate 后逐项批准；在此之前 T-19/T-20 不能以模拟或未签名制品关闭最终发行验收。
- source branch/worktree cleanup 未授权，不阻止集成与 change 本地完成，但所有保留 locator 必须写入状态。

### Resume Protocol

恢复时依次读取本 Goal Plan、当前 Ticket、`<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/.status.json</Path>`、Tickets Map 和最新 Evidence。先核对 `main` 当前 HEAD 是否等于最后通过的 `result_sha`；若存在 active source 或 candidate，从其不可变 checkpoint 继续，不重建已通过工作。若父 HEAD 漂移，暂停派单并重算所有未开始 worktree 的 base；若超过 3 次集成尝试或需要新产品决定，标记 blocked 并返回用户。

## Assumptions

- 低影响假设：父分支逻辑名称继续使用 `main`；实际 Git 分支经实测为 `main`。
- 低影响假设：implementation provider 和模型在每次 execution-time dispatch 时选择；本计划不要求外部网页交付。
- 低影响假设：正式依赖版本由对应 owner Ticket 以可审计 lock/digest 固定，不能使用浮动的“最新”。
- 低影响假设：当前 25 个 lint warning 只作为旧基线记录，不允许新增，最终由新 Workspace/原子收缩自然归零。
