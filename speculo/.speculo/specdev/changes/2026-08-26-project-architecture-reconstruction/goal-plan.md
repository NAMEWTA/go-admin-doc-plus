---
schema_version: 6
artifact: goal-plan
change: 2026-08-26-project-architecture-reconstruction
status: in_progress
modes: [migration, high-assurance, reference-conformance, release-coordination]
orchestration: lead-directed
lead: codex-root
implementation_agent_limit: 3
integration_attempt_limit: 3
ticket_workspace_policy: required
integration_gate: candidate-merge
ready_for_execution: true
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
| 7 | `main@0391c8816cb97e8c68e61ec7ef56715046f8115f` 及当前工作区实测 | 代码、测试、dirty baseline 和工具可行性 | 新事实冲突时触发 deviation，不静默修改合同 |

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
- 当前父分支为 `main@0391c8816cb97e8c68e61ec7ef56715046f8115f`；发现 170 项 tracked 改动和 6 个 untracked 条目。
- 基线 Go test、前端 unit/type/lint/build 与 Tickets validator 已实测通过；尚未形成 G-BASELINE commit。
- 当前 Gate 为 `G-BASELINE in-progress`；尚未创建任何 Ticket source worktree、source commit、candidate 或 result SHA。
- implementation commit 和 Local candidate integration and parent update 已由用户 `Q2A` 授权；source cleanup、远端和生产动作未授权。

### Pending Decisions and Blockers

- 无阻止 T-01 开始的产品决定。
- G6 的生产签名、公证、受保护 runner、远端制品发布仍需到达该 Gate 后逐项批准；在此之前 T-19/T-20 不能以模拟或未签名制品关闭最终发行验收。
- source branch/worktree cleanup 未授权，不阻止集成与 change 本地完成，但所有保留 locator 必须写入状态。

### Resume Protocol

恢复时依次读取本 Goal Plan、当前 Ticket、`<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/.status.json</Path>`、Tickets Map 和最新 Evidence。先核对 `main` 当前 HEAD 是否等于最后通过的 `result_sha`；若存在 active source 或 candidate，从其不可变 checkpoint 继续，不重建已通过工作。若父 HEAD 漂移，暂停派单并重算所有未开始 worktree 的 base；若超过 3 次集成尝试或需要新产品决定，标记 blocked 并返回用户。

## Assumptions

- 低影响假设：父分支逻辑名称继续使用 `main`；实际 Git 分支经实测为 `main`。
- 低影响假设：implementation provider 和模型在每次 execution-time dispatch 时选择；本计划不要求外部网页交付。
- 低影响假设：正式依赖版本由对应 owner Ticket 以可审计 lock/digest 固定，不能使用浮动的“最新”。
- 低影响假设：当前 25 个 lint warning 只作为旧基线记录，不允许新增，最终由新 Workspace/原子收缩自然归零。
