---
schema_version: 6
artifact: goal-plan
change: 2026-08-26-project-architecture-reconstruction
status: in_progress
modes: [migration, high-assurance, reference-conformance, release-coordination]
orchestration: lead-directed
lead: codex-root
implementation_agent_limit: 3
integration_attempt_limit: 4
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

### Unified E2E Deferral Amendment（2026-08-27）

用户最新决定将尚未完成 Ticket 的 E2E 统一延后：T-10、T-12、T-13、T-15 至 T-21 先完成源码、parent-candidate、全部非 E2E 门禁和本地 `main` 集成；这些实现结果可解除下游实现依赖，但 Ticket 保持 `in_progress` / `implemented-pending-final-e2e`。全部 T-01 至 T-21 实现集成后，Lead 基于唯一最终系统候选执行一次统一 E2E，覆盖三 profile、Web/Desktop、权限/Session、可靠运行时、文件系统、生成器、原生宿主、平台发行和原子收缩。只有该统一 E2E 通过后，尚未完成 Ticket 与 change 才能标记 `done` / `complete`。

已完成 T-03 至 T-11、T-14 的既有 E2E Evidence 保持有效历史证据，不重跑也不作废。延后不等于豁免：任何新 source/candidate 仍必须通过其全部静态、单元、集成、race、vet、type、lint、合同、双方言、编译、构建、治理、路径和资源清理门禁；统一 E2E 若发现缺陷，只修复受影响实现并重跑最终系统 E2E。该决定不新增远端、生产、签名、公证、发布或 source cleanup 授权。

### Unified E2E Pause Amendment（2026-08-28）

用户在 G8 Desktop native repair 候选运行期间明确要求“不再 E2E，先完成所有构建和编码”。Lead 已终止正在运行的原生场景，清理测试 Keychain/进程/临时资源，重建并扫描 production 制品；R04 通过全部非 E2E candidate Gate 后进入本地 `main`。从该决定起暂停创建或运行新的 G8 browser、fault、native 和发行 E2E，直到用户重新明确要求恢复。

暂停不修改 Success、False Completion 或发行授权边界：默认 skip、中断结果、source pass 和历史候选部分 pass 均不得记为最终 G8 pass；Ticket/change 继续保持 `implemented-pending-final-e2e` / `in_progress`。恢复时必须基于当时最新 `main` 新建唯一候选并从头重跑完整统一 E2E，不复用失效 attempt 的部分通过结果。

### Worktree Cleanup Amendment（2026-08-28）

用户随后明确授权清理所有已经进入 `main` 的 worktree，并要求没有进入 `main` 的内容先回收再清理。Lead 对 59 个次级 worktree 逐一确认 clean，并核对所有分支相对 `main` 的新增补丁数均为 0；其中 9 个非祖先分支只包含已被等价提交吸收的旧 source 或失败 merge 节点，不存在需要再次应用的独立功能补丁。随后删除全部 source、integration、candidate worktree 及其本地分支，执行 `git worktree prune`，并移除空的 `specdev-worktree/`、`specdev-candidate/` 父目录。

清理后 Git 只登记根工作区和本地 `main`，工作区 clean 且 `main@c308ba7d03fc9ca10156d3adac0849052629cae8` 与 `origin/main` 一致。该清理不撤销任何 base/source/candidate/result SHA、验证或 E2E 事实：11 个已完成 Ticket workspace 进入 `removed`；10 个统一 G8 尚未通过的 Ticket 受 schema 完成门约束继续保持 `review`，但其物理 worktree/branch 已删除，历史 locator 仅作为 Evidence。未运行立即对象回收，远端分支和远端仓库未改动。

### Success and False Completion

成功必须同时满足：

- `AC-001` 至 `AC-036` 均有 Lead 核对的通过 Evidence；
- T-01 至 T-21 每个非 cancelled Ticket 都有非空 source commit、通过全部非 E2E 门禁的 parent-candidate、父分支 `result_sha` 和路径核对；
- 三 profile、Web/Desktop、权限/Session、迁移/Outbox、架构边界、原生制品和供应链证据通过对应 Gate；
- T-21 后 allowlist 外旧目录、旧合同、Wails、JWT、refresh token、Casbin、Redis、tenant、MySQL、SQL Server、AutoMigrate 和临时迁移标记零命中；
- 最终统一系统 E2E 在全部实现集成后的唯一候选上通过；Ticket、Map、Goal Plan、Evidence、change status 与 Git checkpoint 一致，无未集成 source/candidate。

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
| T-10 | Settings | T-07 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-10` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-10.md</Path>` |
| T-11 | Audit | T-06,T-08 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-11` | Lead / execution-time dynamic | required：事件到审计 UI | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-11.md</Path>` |
| T-12 | Scheduler | T-07,T-08 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-12` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-12.md</Path>` |
| T-13 | Files | T-07 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-13` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-13.md</Path>` |
| T-14 | Demo tracer | T-04,T-05,T-07 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-14` | Lead / execution-time dynamic | required：双方言 Web CRUD | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |
| T-15 | Generator | T-14 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-15` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-15.md</Path>` |
| T-16 | Tauri 2 Desktop | T-06,T-08,T-14 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-16` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-16.md</Path>` |
| T-17 | 完整产品组合 | T-09,T-10,T-11,T-12,T-13,T-15,T-16 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-17` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-17.md</Path>` |
| T-18 | Linux OCI/Compose | T-17 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-18` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-18.md</Path>` |
| T-19 | macOS Universal DMG | T-17 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-19` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E；正式签名仍需授权 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-19.md</Path>` |
| T-20 | Windows x64 NSIS | T-17 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-20` | Lead / execution-time dynamic | deferred：全部实现集成后统一系统 E2E；正式签名仍需授权 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-20.md</Path>` |
| T-21 | 原子切换/归零 | T-18,T-19,T-20 | `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-21` | Lead / execution-time dynamic | deferred：作为统一系统 E2E 的最终实现基线 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` |

## 3. Gates and Completion Evidence

### Overall Definition of Done

- 所有非 cancelled Ticket 都完成 source commit -> parent-candidate verification -> parent result SHA；
- shared owner 先稳定并集成，消费者 worktree 只基于包含该 result 的父分支 checkpoint 创建或刷新；
- normal/failure/regression、架构边界、双方言迁移、缓存禁用、Session/权限、Desktop 和原生发行 Evidence 全部由 Lead 核对；
- 尚未完成 Ticket 的 E2E 不在各自 parent-candidate 运行；全部实现集成后只在唯一最终系统候选上统一运行一次，source worktree 的 E2E 声明不计入；
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
| G4 Vertical Modules | G3 关闭并满足各 Ticket 依赖 | T-09 至 T-16 各自实现 result；模块全部非 E2E 门禁通过 | T-17 实现 | Lead | 单模块回退到 source；其他已集成模块保持父结果 |
| G5 Product Integration | G4 全部实现 result | T-17 composition、三 profile/双 App 静态合同、架构、缓存和构建门禁通过并形成实现 result | 三平台发行实现 | Lead | 父分支不推进；修复 T-17 composition 或返回模块 owner |
| G6 Protected Release Implementation | T-17 实现 result | T-18/T-19/T-20 本地实现 result、构建/policy/SBOM/provenance 门禁；未授权保护动作显式 pending | T-21 实现 | Lead；保护动作仍需独立批准 | 未授权凭据/远端动作保持 pending；不伪造签名或发行 pass |
| G7 Atomic Contract Implementation | 三平台实现 result；旧路径 inventory 冻结 | T-21 实现 result、全量 root 非 E2E Gate、allowlist 外零兼容命中、删除清单 | G8 | Lead；删除范围来自已批准 T-21 | candidate 失败时父分支不动；修正 source 或整体放弃候选 |
| G8 Unified System E2E | T-01 至 T-21 全部实现 result 已进入 `main` | 唯一最终系统候选完成三 profile、Web/Desktop、模块、可靠运行时、原生宿主、发行与零兼容统一 E2E；Evidence 与状态一致 | change completion | Lead；受保护动作按原授权边界 | 修复受影响实现后重建最终系统候选并完整重跑统一 E2E |

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
| Integration attempts | 每个候选最多 `4` 次 | 用户 `Q2A` 批准将 config 与本计划快照从 3 提升到 4 |
| Read-only agents | 无 SpecDev 数字上限 | review/research/test-observation，只读且不竞争可变环境 |
| Dispatch | execution-time dynamic | provider、模型、自行实现或派单按 Ticket 事实选择，不静态预分配 |
| SpecDev writes | 仅 Lead | subagent 不写 Ticket/Map/Goal Plan/Evidence/status |
| E2E | 仅 Lead；全部实现集成后统一执行一次 | implementation 和逐 Ticket candidate 只返回非 E2E 事实；最终系统候选集中验证 |

动态 Dispatch Packet 遵循 `<Path>{roots.workflows}/specdev/common/skills/subagent-delivery/SKILL.md</Path>` 的 `operation=plan` 合同：允许 `implementation/review/research/test-observation`；每次 dispatch 固定 Ticket、依赖 Evidence、base SHA、workspace、路径、授权、非 E2E 检查、停止条件和返回格式。外部 provider 需要另行的数据发送授权；本计划不授权 external-web 源码交付。

### Ticket Workspace and Integration

所有 Ticket 使用 `required` 策略。每个 source branch/worktree 基于开始时最新父分支 result SHA；source worktree 不运行 E2E。implementation owner 只写 Ticket frontmatter 允许的项目路径并形成非空 source commit。Lead 冻结 `parent_before_sha`，在 Lead-owned candidate 中集成 source checkpoint并运行全部非 E2E 集成检查；通过后可更新 `main`、记录实现 `result_sha` 并解除下游实现依赖，但未完成 Ticket 保持 `implemented-pending-final-e2e`。全部实现 result 完成后，Lead 才在唯一最终系统候选运行统一 E2E。

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
| T-10 | T-07 result | worktree T-10 / `specdev/T-10` | module unit/API/type | required non-empty | full non-E2E module Gate；E2E -> G8 | implementation result SHA |
| T-11 | latest deps result | worktree T-11 / `specdev/T-11` | audit/redaction/retry | required non-empty | event-to-UI E2E | passed candidate SHA |
| T-12 | latest deps result | worktree T-12 / `specdev/T-12` | clock/registry/unit | required non-empty | full non-E2E scheduler Gate；E2E -> G8 | implementation result SHA |
| T-13 | T-07 result | worktree T-13 / `specdev/T-13` | file security/unit | required non-empty | full non-E2E files Gate；E2E -> G8 | implementation result SHA |
| T-14 | latest deps result | worktree T-14 / `specdev/T-14` | Demo dialect/unit/type | required non-empty | PostgreSQL/SQLite Web E2E | passed candidate SHA |
| T-15 | T-14 result | worktree T-15 / `specdev/T-15` | golden/type/compile | required non-empty | real generate/compile non-E2E Gate；Web E2E -> G8 | implementation result SHA |
| T-16 | latest deps result | worktree T-16 / `specdev/T-16` | Go/Rust/TS/release build | required non-empty | full native build/static Gate；native E2E -> G8 | implementation result SHA |
| T-17 | all module implementation results | worktree T-17 / `specdev/T-17` | full composition non-E2E gates | required non-empty | profile/App build/contracts；E2E -> G8 | implementation result SHA |
| T-18 | T-17 implementation result | worktree T-18 / `specdev/T-18` | OCI/Compose policy/build | required non-empty | local release implementation Gate；E2E -> G8 | implementation result SHA |
| T-19 | T-17 implementation result | worktree T-19 / `specdev/T-19` | macOS policy/build | required non-empty | local release implementation Gate；protected actions pending；E2E -> G8 | implementation result SHA |
| T-20 | T-17 implementation result | worktree T-20 / `specdev/T-20` | Windows policy/build | required non-empty | local release implementation Gate；protected actions pending；E2E -> G8 | implementation result SHA |
| T-21 | all release implementation results | worktree T-21 / `specdev/T-21` | inventory/zero/full static | required non-empty | atomic contract non-E2E Gate；final full E2E -> G8 | implementation result SHA |

Candidate 集成队列严格串行。若父分支已包含 source checkpoint，可 fast-forward；否则使用独立 merge commit。候选验证失败时父分支保持不变，同一 source worktree 修正；父 HEAD 漂移则 candidate 标记 stale 并从最新父结果重建。成功集成不自动清理 source branch/worktree。

### Authorization Matrix

| 动作 | 状态 | 目标与条件 |
|---|---|---|
| Current workspace Ticket changes | not-authorized | 本计划使用 required，不在当前 workspace 实现 Ticket |
| Ticket worktree local changes | allowed | Q1A/Q2A；仅对应 Ticket writable/shared-owner 路径 |
| Implementation commit | allowed | Q2A；包含经审计 baseline commit 及 T-01 至 T-21 非空 source commit |
| Local direct-parent verification and parent update | not-authorized | 不适用 required 策略 |
| Local candidate integration and parent update | allowed | Q2A；Lead-only、通过全部非 E2E candidate checks 且父 HEAD 未漂移；最终状态等待统一 E2E |
| Push / PR / remote merge | not-authorized | 本地授权不继承到远端 |
| Branch/worktree cleanup | completed | 用户 2026-08-28 明确授权；59 个 clean source/integration/candidate worktree 及本地分支已删除，Git 仅剩根工作区与本地 `main` |
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

source worktree 只运行 Ticket 非 E2E 检查；任何 source worktree E2E pass 无效。Lead 在逐 Ticket parent-candidate 运行受影响的非 E2E 集成、回归、双方言和构建门禁，并在全部实现集成后的唯一最终系统候选运行统一 E2E。禁止以跳过非 E2E 门禁、放宽 lint/type、更新 golden 而不审查、mock 替代统一 E2E 中的真实 PostgreSQL/SQLite/原生宿主、忽略非零退出码或只检查文件存在制造伪绿色。

### Migration or Release Sequence

1. **Baseline：** 审计并提交当前用户改动和规划工件，记录父分支基线 SHA；不纳入 secret、临时产物或未解释大文件。
2. **Expand：** T-01 至 T-05 建立根治理、公共合同、runtime/database/frontend 新结构，旧产品保持未发布只读参考。
3. **Vertical migrate：** T-06 至 T-16 逐模块迁入新结构，每个 source/candidate 独立可验证，不做旧数据/API 双轨。
4. **Observe/integrate：** T-17 聚合三 profile 与双 App，关闭共享合同、迁移、权限和可靠性 Gate。
5. **Release implementation：** T-18/T-19/T-20 生成各平台实现与本地非 E2E 证据；正式签名、公证、远端/受保护 runner 动作等待独立批准。
6. **Contract implementation：** 三平台实现 result 完成后，T-21 精确删除旧路径并完成零兼容扫描；不在 earlier Wave 提前删除。
7. **Unified validation：** T-01 至 T-21 全部实现集成后创建唯一最终系统候选，集中执行一次完整 E2E；通过后再统一关闭未完成 Ticket 与 change。

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
| 父 HEAD 漂移/候选冲突 | integration queue 重读 parent HEAD | 标记 stale，最多 4 次重建；之后 blocked |

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
- T02-D04 最终双轴复审对 `faa451b47b59bb5c207fe1dcdbcb39bc8474cacd..6fab034640672fef057acad61139aa8b917a85a1` 均 pass、0 findings。Lead candidate/result 为 `7271df83db41c8e7018616f9fee0f68f61c9ffee`（tree `c79ef58b0bd9cfcaf80717db7df76f4feae20467`）；合同 28/28、API client 7/7、新 Workspace 22 unit + 6 boundary、Go 普通/SQLite/vet/race/mod、governance、SpecDev 与 clean-tree 全绿，父分支已从 `d828f74fb8cab349597123ae0ff7371e5aac7b03` 提升。当前机器没有 Go Task CLI，未安装系统工具，Task 底层命令已直接通过。
- T-06 的 T02-D04 依赖已关闭；clean source worktree 从 `05d619d1c4c90378a0138c630fa1aa3bcfa0f942` 无冲突快进到最新依赖记录 `825a1c63fd3afd748e7da7efea10004569dc484c`。T-06 解除阻塞并按已批准 `T06-D01` 恢复实现，required Cookie/CSRF/browser E2E 仍仅由 Lead candidate 执行。
- T-08 checkpoint `d2311880e7f0b597f5ad4d470ffa316ae47e430d` 的双轴复审再次 request-changes。固定修复合同为：以结构化 mutation 描述替代任意 SQL+黑名单，显式限制非平台 owner 表、操作、业务键、参数和影响行数，并拒绝 Unicode quoted identifier 绕过 advisory 锁；将 Claim/Deliver/Retry 收回不可由 API replica 直接调用且绑定同一 Database/owner 的真实 Lease 状态机；以 topic payload schema/allowlist 和非敏感稳定 BusinessKey 类型取代字段名猜测；补 pending age、claim duration、attempt/retry、active executor/lost-lock 指标，并令探测失锁立即分类。旧 checkpoint 作废，required E2E 未开始。
- T-08 第三修正 checkpoint `9a8fdc9d9d4e694ea23a2f147a1e785b4a0723f1` 已闭合结构化 Mutation、Store 私有执行状态机、真实 Lease 数据库/owner 绑定和 observer 指标，但标准轴复审返回 1 个 HIGH：允许的 `PayloadString` 值仍可携带原始 secret，重复 JSON key 还可令 map 校验只看到最后安全值而原始 payload 持久化隐藏对象。该 checkpoint 作废；同一 owner 仅补流式 duplicate-key 拒绝、fail-closed 字符串值合同与无歧义持久化回归，required E2E 仍未开始。
- T-08 checkpoint `90dfc82d4e1f3d83fbe35000fedf386396119534` 已通过重复 key 与 canonical payload 复核，但 `AllowedStrings` 编译仍允许把 `raw-session-secret` 本身登记为闭集成员。该 checkpoint 作废；只退回补严格 domain-label grammar、token boundary 敏感词拒绝及配置级负向测试，不重开已通过的执行状态机边界。
- T-08 checkpoint `47290318ed9a3fc8f4a9ef1bfda56468149f3036` 的标准轴 pass、0 findings；规范轴因固定源码只有同进程双 pool 与 graceful close，缺少可由 Lead 执行的真实多 OS 进程、`pg_terminate_backend`、失锁回滚、takeover、retry/receipt replay harness 而返回 1 个 HIGH。产品逻辑不重开；原 owner 只在 T-08 测试路径补默认 skip 的可复现 PG E2E 资产，required E2E 仍由 parent-candidate 实际运行。
- T-08 最终 clean source checkpoint 为 `212f0fb785ea92480a0185fd0cacf592d844d969`。新增 env-gated 跨 OS 进程 PostgreSQL harness 后，标准轴与规范轴对完整范围 `05d619d1c4c90378a0138c630fa1aa3bcfa0f942..212f0fb785ea92480a0185fd0cacf592d844d969` 均 pass、0 findings；默认 skip 编译、聚焦 unit/race/vet、全量 Go/SQLite 与 clean 已通过，真实 PostgreSQL 18.3 backend termination E2E 保持 Lead parent-candidate pending。
- T-08 parent-candidate/result 为 `503ea063092d9ee595c8adcb59308edca691171b`（tree `44fd724d0764ca3843714b6ede8c9de8ab902dd6`，parent before `08a6535f63ec5da8824f69d1c19fc851afa4f998`，merge attempt 1）。Go 普通/SQLite/全 race/vet/build/mod、UI frozen/peer/22+6/build、合同串行 28/28、API client 7/7、governance/SpecDev/clean 全绿；PostgreSQL 18.3 真实 holder backend termination、事务回滚、跨进程 takeover、retry/receipt replay E2E pass，父分支已 `--ff-only` 提升。首次默认并发合同运行的空 virtual store pnpm 竞争结果已作废并在串行安装后从零重跑。
- T-06 首轮双轴审查对 `b44f90952497a753454eb7a1256d920d40accfb8..5ad57f6142a5165f63a12771d11b278fe082119e` 均 request-changes。固定返修包为：Argon2id 全局有界 fail-fast 工作预算；account 级 epoch/锁序闭合 PostgreSQL rotation 与 revoke 竞态；所有受保护请求执行到期轮换；logout 网络/500 保持已认证可重试，CSRF 拒绝进入重新登录；context sentinel 与测试诊断不泄密；IAM tests 纳入正式 `pnpm verify`；三 profile 提供类型化 idle/absolute/rotation 配置。Lead 批准 `T06-D02` 精确开放 T-03 配置/schema 与根前端测试入口，旧 checkpoint 作废，candidate/E2E 未开始。
- T-06 修正 checkpoint `fafb754fa0c3a8a042ff771c3fa52ed4cec40a5e` 已闭合首轮 Argon2 budget、generation/锁序、全请求轮换、前端失败态、typed policy 和标准 pnpm 门禁，Lead 独立 IAM race/vet、37 tests、6 boundary、build 与合同 28/28 全绿；但双轴复审仍 request-changes：配置 JSON `int64` 在边界检查前乘 `time.Second` 可回绕为合法值；repository/session SQL 包装在 `sanitize` 前丢失 context sentinel；固定源码没有 Lead 可直接执行的同源 HTTPS browser 与 SQLite/PostgreSQL 双 profile 完整 E2E 入口。该 checkpoint 作废，原 owner 只在既有 T-06 writable paths 补前乘法边界、全 SQL sentinel 回归及受跟踪 E2E harness，candidate 尚未创建。
- T-06 第二轮返修 checkpoint `f9375a827f821588f952d99f1813c2f7fbaa98d0`（tree `800d739246e723d1edd7c7379cc9479e3f8def9f`）已闭合前乘法边界、SQL context sentinel 和可追踪双 profile browser harness；规范轴 pass、0 findings，Lead 独立 IAM/config race/vet、前端 37 tests/6 boundary/build、合同 28/28 与 lint/generate check 全绿。标准轴仍返回 3 个 medium findings：required runner 把 PostgreSQL DSN 继承给无关 Vite/Chromium/SQLite 子进程；CDP/child/shutdown/PostgreSQL cleanup 缺少完整有界回收；account create 把所有非 context INSERT 故障误报为 conflict。该 checkpoint 作废，原 owner 只在既有 T-06 writable paths 收紧子进程环境与有界清理，并区分双方言唯一约束和脱敏内部故障；candidate attempt 仍为 0，required E2E 未开始。
- T-06 最终 source/candidate/result 为 `529a6b5ce008390d71930249f1bf806f887d2de7`（tree `6fbff592ba2ff4a7ab37f26c4b91654fe09007fe`，parent before `f4cd993f49bd7930b1d128ce00d128bbda58eb71`，fast-forward attempt 1）。最终双轴审查均 pass、0 findings；Go 普通/SQLite/全 race/vet/build/mod、25-workspace frozen install、前端 5 files/37 tests + 6 boundary/build、合同 28/28、API client 7/7、governance/SpecDev/clean 全绿；Chrome for Testing 151 在真实 SQLite 临时文件和 PostgreSQL 17.11 隔离 schema 上完成 Cookie/CSRF/rotation/旧 token/profile/logout/password/timeout E2E，独立 PostgreSQL generation/revoke 竞态也通过，schema、进程、端口和临时目录均清理。Lead 首次创建 candidate 后误在根 workspace fast-forward，立即作废旧 parent 检查、将 clean `main` 精确恢复至 frozen parent，并在正确独立 candidate 从 frozen install 起重跑全部 Gate 后才正式提升；事故未改变代码/tree 或最终证据。
- T-06/T-08 结果已同时解除 T-07 与 T-11 阻塞；Lead 从最新 `main@c108f8bff07b386cd83356236dc79cd11c301813` 创建两个路径不相交的 required source worktree，implementation owner 分别为 `codex-t07-iam` 与 `codex-t11-audit`。两者只运行 source 非 E2E checks；权限/UI 与 event-to-Audit/UI 的双 profile required E2E、candidate integration 和 SpecDev 写入仍由 Lead 保留。
- T-07/T-11 clean preflight 同时证明四个预建 package manifest 仍是前序占位壳：IAM scripts/exports 只覆盖 Session，Audit 缺 canonical client/Vue 直接依赖和标准门禁；两类管理 UI 还必须显式复用 `@go-admin/ui`。Lead 批准 Ticket 级偏差 `T07-D01` 与 `T11-D01`：各自只写本模块两个 manifest，workspace lock 先由关键路径 T-07 更新 IAM importer；T-11 可并行实现独占源码和 manifest，但必须等待 T-07 result、rebase 后才更新两个 Audit importer。依赖只能复用现有 workspace lock/catalog，其他漂移必须再次停止。
- T-07 backend RED/GREEN 后在 HTTP preflight 证明 T-06 只有私有通用 CSRF mutation fence；若 T-07 直查 Session 私表会复制 token hash、touch 和 rotation 安全规则。Lead 批准 `T07-D02`，只允许在 `<Path>go-admin-plus/internal/modules/iam/session/service.go</Path>` 增加公共 `AuthorizeRequest`，并在既有 Session 测试文件覆盖 token/CSRF/touch/rotation/失败不变性；schema、Cookie/policy 和 Session HTTP 均保持不变。
- T-11 独占实现 preflight 证明 Audit 虽已拥有同步 LoginRecorder，T-06 Session 没有可注入调用点；若只保留未接线 recorder，`AC-016` 的真实登录事实与 event-to-UI E2E 都会形成伪完成。Lead 批准 `T11-D02`，但执行严格排在 T-07 result 之后：T-11 rebase 后才可精确修改同一 Session service/test，增加不依赖 Audit 的结构化 Login Fact Port，证明成功登录与 Session 创建同事务、失败登录同步落事实、审计故障不签发 Session，并保持用户名、密码、token、CSRF 和 raw request 零泄露。操作事件改用 Audit 自有、模块无关的封闭 topic/BusinessKey 合同，禁止 T-11 反向硬编码尚未实现的 Demo 私有 topic。
- T-07 最终 clean source 为 `379d0e2117844fa2ff1ec4197b8707a1e3f2da59`（tree `1088c7f877ea5fd42fc5f416a63abf97a07b3bc8`），标准轴与规范轴均 pass、0 findings。最终 parent-candidate/result 为 `43624643302ac8d34806005761aaae112da2384b`（tree `2e51848dbfa533a4f5c707ea8b84deefe87fda9e`，parent before `75924fc197a7abfe840ca3b7383ca7eeb2ffd45e`，merge attempt 3）。Go 普通/SQLite/全 race/vet/build/mod、25-package frozen/37+6/build、合同 28/28、API client 7/7、governance/SpecDev/clean、PostgreSQL final-authorization concurrency 与 Chrome 151 双 profile required E2E 全绿；全新 PostgreSQL 17.11 验证 `public` 表和 `t07_*` schema 均零残留。
- T-07 attempt 1 的 virtual-time browser runner 与 attempt 2 的响应式/DOM scheduler 修正均未提升。attempt 2 虽通过双 profile 场景，但后续诊断证明 `pgx.ConnConfig.ConnString()` 不会序列化新写入的 RuntimeParams，harness 的 `search_path` 实际丢失并污染 disposable `public`；该结果作废。attempt 3 改用 URL-safe query 参数、补稳定 profile 诊断并在全新集群从 frozen install 起重跑所有 Gate 后才判定通过。T-07 状态为 `done`，现解除 T-09/T-10/T-12/T-13/T-14 的执行依赖，并允许 T-11 按既定串行合同 rebase 后进入第二阶段。
- T-11 第二阶段已在 T-07 result/state `40ab83480f935f4be89b14948afd21e8e9c88192` 后激活：frontmatter 现精确开放 lockfile 的两个 Audit importer，以及 Session `service.go`/既有 `session_test.go` 的模块无关 Login Fact Port；Session HTTP/transport/schema、其他 lock importer、Outbox 与共享 UI 继续只读。stage-1 clean source checkpoint 为 `cb86e2ffa58ccf8ce40df7d03a0f06f7c2dc81ec`，必须先 rebase 到最新 parent，再进入登录事实接线与 required 双 profile E2E。
- 关键路径 T-14 已从 clean `main@18e410d0ac5328a528ce43510705629ae7b21999` 激活，source worktree/branch 固定为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-14` / `speculo/2026-08-26-project-architecture-reconstruction/T-14`，implementation owner 为 `codex-t14-demo`。它只写 Demo 独占合同、后端、前端与测试路径；真实 PostgreSQL/SQLite Web、重启持久化与 candidate integration 仍由 Lead 执行。
- T-09 已从 clean `main@3828a779e824842ec1f5de0bebfd2c69955764bf` 激活，source worktree/branch 固定为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-09` / `speculo/2026-08-26-project-architecture-reconstruction/T-09`，implementation owner 为 `codex-t09-organization`。它只写 Organization 独占合同、后端、前端与测试路径；双方言 Organization API/UI required E2E 与 candidate integration 仍由 Lead 执行。
- T-14 clean preflight 证明 canonical fragment 自动发现和生成目标已满足，但预建 Demo package manifest 缺直接依赖/标准脚本，根 verify 未纳入 Demo tests，且 lock 缺两个 importer。Lead 批准 `T14-D01`：stage 1 只写两个 Demo manifest、根 Demo 聚合脚本和 Vitest Demo include；lockfile 必须等待 T-11 result、rebase 后才由 T-14 stage 2 只更新两个 Demo importer。禁止新增外部版本或修改其他 importer/shared UI/composition。
- T-09 clean preflight 证明预建 Organization manifests 缺直接依赖/标准脚本，且 ADR-012 要求的 IAM consumer Port 尚不存在。Lead 批准 `T09-D01`：stage 1 只增加两个 Organization manifest 与精确 `iam/administration/organization_port.go`；根 verify/Vitest/lock 必须等待 T-14、T-11 对应 result 后再由 Lead amendment。Port 由 IAM 定义最小投影且不得导入 Organization，adapter 归 Organization，显式注入仍归 T-17。
- T-11 fixed source `e027b17d6ca02c7d5852a4a9c213413fd490e817` 闭合 Vue 全局声明与 PostgreSQL `search_path` 隔离后，标准轴仍拒绝：constructor 默认 discard 令登录审计可静默漏接，required browser actor/count fixture 必失败，真实 IAM rotation 未覆盖。Lead 批准 `T11-D03`：constructor 缺 Port 必须 fail-closed；只开放四个精确 test-only 调用点注入本地 noop；修正 fixture 并以可推进时钟在真实 adapter/browser 路径验证 replacement Cookie/CSRF。candidate attempt 仍为 0。
- T-05 双轴复审对 `27a186c..5b7b4f4` pass、0 findings；首次 parent-candidate `f2bbd5211eada20b8ddc3762cba680bc7944ceae` 的 pnpm 全门禁和功能浏览器断言通过，但 Lead 截图检查发现 390px 视口导航隐式 grid row 被拉伸至近半屏。required visual E2E 判定失败，candidate 未提升并已清理，checkpoint 退回原 owner 修正并补导航高度断言。
- T-05 原 owner 以 `a648baf851d25861ac04eef688b8856fcf439b71` 修正移动 grid rows 并增加静态样式合同；source `pnpm verify` 全绿且 clean，最终双轴复审固定为 `27a186c..a648baf`，第二次 candidate 将把 390px nav 高度上限纳入 required browser E2E。
- T-05 第二次 candidate frozen install 暴露根 `.gitignore` 仍漏掉 pnpm 子 package `node_modules` symlink。Lead 在 T-01 已有所有权内以 `3bc0dd5fe13a6e215caf6f95781482152ccfc279` 补齐 `<Path>go-admin-plus-ui/**/node_modules/</Path>`；该父分支纠正将合入当前 candidate 后继续 clean-tree/browser 门禁。
- T-05 第二次 parent-candidate 已在纳入最新父分支治理修正后固定为 `d5d10dda68fb6b49850c7442bf036b195ba1d202`（tree `0511250c439a3233d9217fb41de598dfdc17dadb`）。24-workspace frozen install、lint/type、22 项 unit、6 项 boundary、build、合同生成、治理、SpecDev 校验与真实 Chromium desktop/mobile E2E 全部通过；390px 导航高度为 71px 且无溢出。`main` 已从 `c2260191c86942f859088d15b9a3ba9af96c975b` 以 `--ff-only` 晋升到同一 result，T-05 状态为 `done`。
- T-11 attempt 1 固定在 `11a5d4e01ec7546b88000d7c5fea25e24f803f6d`（tree `05a3818d60e68bf1a10b53b27bba60d051f6f813`）：完整静态门禁通过，但 required browser 因页面合同值 `succeeded` 与 driver 的 `Succeeded` 断言不一致而失败，未提升。诊断增强首版又被独立复审发现跨 pipe chunk 可重组 DSN/Cookie/password 明文，旧 checkpoint 作废；最终以完整行重组后脱敏、未完成/超长行 fail closed 和全 split 回归闭合，双轴复审 pass、0 findings。
- T-11 最终 clean source/candidate/result 为 `ebb8bac4a16b09301593b508549f6bac71f6dba0`（tree `c054434e79ec4b52d33759ffb79d7abbb0d3309c`，parent before `70d277364915f12d757b83b1085c328bc12a1442`，fast-forward attempt 2）。Go 普通/SQLite/全 race/vet/build/CGO=0/mod、25-package frozen/37+6/build、Audit 1+10+7 定向测试、合同 28/28、API client 7/7、governance/SpecDev/clean 全绿；Chrome 151 在真实 SQLite 与 PostgreSQL 17.11 完成登录/操作事实、rotation Cookie/CSRF、筛选/详情/清理、权限撤销与 Session revoke E2E，`public`、`audit_test_*`、进程和临时目录均零残留。T-11 状态为 `done`，现把 lockfile 串行所有权移交 T-14 stage 2。
- T-14 stage 2 已在 T-11 result/state `9bdde5af0ec5e4a57e43490eff19a3e28778b406` 后激活：stage-1 clean checkpoint 为 `64b906758795fa9dfcbd075d50c4e0b11749072b`；frontmatter 现只开放 lockfile 的两个 Demo importer。原 owner 必须先 rebase 到最新 parent，证明其他 importer、共享 UI、composition 与外部版本零漂移，再返回最终 source checkpoint；required 双 profile E2E 仍仅由 Lead candidate 执行。
- T-14 checkpoint `c6dca097c906ff91ae7b6db75d58aa2acb768bee` 的独立双轴审查返回 FAIL：Demo required harness 使用合成身份、错误 Cookie/CSRF 和恒定 `ScopeAll`，Permission Code 无 IAM-owned 生产注册接缝；未知 scope fail open，self 删除 stale revision 误报 denied，Vue 写后投影修复可重复 mutation，PostgreSQL 负向等价与 schema 隔离证据不足。Lead 批准 `T14-D02`，只开放 IAM authorization 的 Permission registry 实现/测试两个精确文件；Demo-owned adapters/harness 必须使用真实 IAM Session/authorization，registry 事务化注册并授权 protected system administrator，Demo migration 禁止写 IAM 表。旧 checkpoint 不进入 candidate，attempt 仍为 0。
- T-09 stage-1 checkpoint `86e21101c86b9672469f3bb5494f745ce2d4b4b3`（tree `50e0ae0e188f98d90568ac1b65e463ccc0406435`）双轴审查返回 FAIL：前端 projection visibility 缺 request sequence fence，Organization authorization adapter 不能证明与业务事务同一 DB/dialect，self scope 页面未隐藏，Unicode 长度语义分叉，PG harness 未实证 `current_schema()`。另证明业务导航菜单与 Permission 一样缺 IAM-owned 生产注册入口；旧 checkpoint 保留但不得 rebase/候选，等待 T-14 registry result 后统一返修。
- Lead 据此批准 `T14-D03`：把未提交的 Permission registry 提升并重命名为 Module Capability Registry，在同一对精确 IAM authorization 文件中事务化注册稳定 Permission 与 protected menu，并授予 protected system-admin；业务模块只声明能力，禁止 migration 跨表。该修订避免 T-09 及后续模块重复建立 IAM 表写入例外，T-14 attempt 仍为 0。
- T-14 后续固定 source 的独立审查证明 Demo 为保留本地校验前的稳定投影而复制了 T-05 filters/page/sort/selection/loading/request-sequence 状态机；共享 controller 又会在请求成功前提交 query state，使纯包装无法取消旧请求并保留最后成功状态。Lead 批准 `T14-D04`，只开放共享 `list.ts` 与既有 `list-form.spec.ts`，加入同步 normalize/validate 与最新成功后原子提交语义；Demo 必须回归配置/包装共享 controller，禁止第二套列表状态机。此前 source 不进入 candidate，attempt 仍为 0。
- T-14 最终 clean source/candidate/result 为 `72c78bd74c6dbc0166df9dfda23df69060719713`（tree `719d9c5da635c7f6e91327337e2de7f84f32b1e0`，parent before `8e6dc8b2a9810c439969359da066e6d0e1cc8158`，fast-forward attempt 1）。第六次独立最终复审 pass、0 findings；Go 普通/SQLite/全 race/vet/build/CGO=0/mod、25-package frozen/65 tests/6 boundary/build、合同 28/28、API client 7/7、governance/SpecDev/clean 全绿。真实 Chrome 在 SQLite 与 PostgreSQL 17.11 完成 IAM Session/CSRF/权限撤销、CRUD、Unicode 字面量搜索、冲突、投影修复与重启持久化，精确返回 `DEMO_E2E_PASS profiles=sqlite,postgres`；`public`、`t14_*` 与 E2E temp 均零残留。T-14 状态为 `done`，现解除 T-15 并完成 T-16 的最后一个执行依赖。
- T-09 stage 2 现以 `T09-D02` 激活：在 T-14 result 后由原 owner rebase 到最新 parent，闭合已登记的 request-sequence、同 DB/dialect 授权、self visibility、Unicode 与 PostgreSQL schema findings，并只更新 Organization 根 checks、Vitest include 和两个 lock importer。T-16 stage 1 不写这些共享根路径。
- 关键路径 T-16 已从 clean `main@ffc53bda0691fa5c64a7400bd17634e9dc28ea4d` 激活，source worktree/branch 为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-16` / `speculo/2026-08-26-project-architecture-reconstruction/T-16`，implementation owner 为 `codex-t16-desktop`。`T16-D01` 只开放 T-05 已预留的 Desktop adapter；当前 registry 预检固定 JS API `2.11.1`、Rust Tauri `2.11.5`、Stronghold `2.3.1`、single-instance `2.4.3`，根 checks/lock 等待 T-09 result 后串行接入。
- T-15 已从 clean `main@c832ca4a4c1b087aaee9fcc2f9e671d7d7e487d3` 激活，source worktree/branch 为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-15` / `speculo/2026-08-26-project-architecture-reconstruction/T-15`，implementation owner 为 `codex-root`。`T15-D01` 只开放两个 Generator 自有 package manifest；根 checks/lock 与 T-16 stage 2 一样等待 T-09 result 后串行接入。
- T-09 最终 clean source 为 `42df51eff0d21617ecf2cf8075fc06723cf29a08`（tree `d835504067a91fe2598166a51dd45259e4eb247b`），parent-candidate/result 为 `b71fb04b3fa2f87b4d21c7529c589b8c79d9223a`（tree `3612509143c3d5bcdbf0dde1ff6d9148c8aec399`，parent before `5064f44e72cc44efa84fe6d5d3ec2768be3926c5`，merge-commit attempt 1）。Go 普通/SQLite/全 race/vet/build/CGO=0/mod、25-package frozen/77 tests/6 boundary/build、合同 28/28、governance/Task contract/SpecDev/clean 全绿；Chrome for Testing 151 在真实 SQLite 与 PostgreSQL 17.11 完成 IAM Session/CSRF、部门与岗位 CRUD、循环/引用/revision 冲突、Unicode 与字面量搜索、权限撤销和 Session revoke，精确返回 `ORGANIZATION_E2E_PASS profiles=sqlite,postgres`，`public`、`t09_*`、进程、端口与临时目录零残留。候选创建时 Lead 首次误把 merge 落到 clean 根 worktree，立即精确恢复 frozen parent，并在正确独立 candidate 从 frozen install 起重跑全部 Gate 后才正式提升；误提交未进入结果或证据。T-09 状态为 `done`。
- T-10 已从 clean `main@cc2ae75c0f2374f7b57c171cc1e819955e1da9fd` 激活，source worktree/branch 固定为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-10` / `speculo/2026-08-26-project-architecture-reconstruction/T-10`，implementation owner 为 `codex-t10-settings`。`T10-D01` 第一阶段只开放两个 Settings manifests 与原独占路径；根 checks、Vitest include、lockfile 和 required 双 profile E2E 继续由 Lead 串行保留。
- T-16 stage 2 以 `T16-D02` 获得 T-09 result 后的下一段共享根所有权：原 owner 先固定 stage-1 clean commit，再 rebase 到 amendment parent；只可追加 Desktop aggregate checks/spec includes，并更新 lockfile 现有 `apps/admin-desktop` 与 `packages/adapters/desktop` 两个 importer。T-15/T-10 在 T-16 result 前不得写这些共享根文件。
- T-16 三次 parent-candidate 均未提升：`bbd9b225` 因无界进程发现输出失败，`7fba6219` 因 migration-failure startup ordering 超时，`a1429111` 的静态门禁通过但 debug 原生宿主 90 秒未开登录窗且异常退出曾留下孤立 sidecar。最终 source `f2d611133a302b912f43cb1d648943e1909762f0`（tree `8e93354e88ae5b1ebbbbdc373252523f509fa4e8`）已改用 release host、父 stdin EOF 自停和精确有界回收并通过全部 source 门禁。用户最新 `Q1A/Q2A` 已批准 `T16-D03` 和第 4 次 candidate，并将 config/本计划 `integration_attempt_limit` 同步提升到 4；父分支将在新 candidate 全部门禁通过前保持不含 T-16 source。
- T-12/T-13 从 `main@9b086f025dd31909edea086727e45c7f301c5ccf` 激活，implementation owner 分别为 `codex-t12-scheduler` 与 `codex-t13-files`。`T12-D01` 只开放两个 Scheduler manifests 并锁定与 Outbox 共用同一全局 lease/事务；`T13-D01` 只开放两个 Files manifests 与 Browser Files adapter 精确接缝。两者根 checks/Vitest/lock、T13 Desktop/config composition 均等待 Lead 后续串行 amendment。
- T-13 source 全量 pnpm gate 证明共享 workspace boundary 把 Browser adapter dependencies 硬编码为仅 platform，与已批准的 Files runtime dependency 冲突。Lead 以 `T13-D02` 只开放该测试的精确 allowlist 断言：必须且只能包含 platform 与 domain-files，并保留额外内部依赖的负向拒绝；其他共享/root 路径不开放。
- T-12 最终 clean source 为 `a9b8a568b17934db78ea2d89d258cf5dfa065fad`（tree `1c4f9613027324cfd7c216f77bbf3db78a4c5d1f`）。独立审查闭合 timeout 误记成功、长任务补跑过期 backlog、软删除后同名不可重建三项问题；source Go/race/vet、SQLite/PostgreSQL、合同、TypeScript/Vitest、boundary/build 均通过，required 多实例/browser E2E 与 parent-candidate 保持 pending。
- T-13 最终 clean source 为 `fd43839c296047ed2e78cc9edb1b106b5b4129ca`（tree `ed8e335f6efc723f3df91a4a8e39e25a64747318`）。独立审查闭合 ready timestamp 响应/持久化漂移和 Unicode 控制字符前后端分叉；Lead 聚焦 Go 与 domain 回归通过，完整 source Go/pnpm/real PostgreSQL/boundary 门禁通过，required filesystem/browser E2E 与 parent-candidate 保持 pending。
- T-15 最终 clean source 为 `8458ceac6a5ce05a1eec811374f60b2f33a83729`（tree `b582cbedbb73260e119b11dc27b22d113c1f3a43`）。独立审查闭合 Config/Preview 绕过 metadata read permission；Generator 全包真实 render/generate/compile gate 与聚焦权限回归通过，required 双 profile 生成/browser E2E 与 parent-candidate 保持 pending。
- T-17 只读 preflight 证明已完成的 T-11 仅声明 Audit permission 常量、未向 IAM Module Capability Registry 注册权限/菜单。Lead 以 `T11-D04` 精确重开 Audit 自有 capability 实现与测试；原登录/操作审计 result 仍保留，补修不得修改 Session、Outbox、前端、root 或 composition。该 preflight 同时确认 IAM Organization consumer Port 只有定义而无实际消费；由于当前 Self/All 数据范围没有可调用组织语义，此项不能以无效 constructor 注入伪闭合，保持为 T-17 前设计 blocker并等待形成真实用例。
- T-11 `T11-D04` 最终 source `772a2c6f7e5023ef26acb2bedf622f3b0e8384ce` 已在 `main@ef69be849415b43b7a83f8873be5dec8338d381d` 上形成 candidate/result `bf5ecdb60f1c861f21110c4ff79467085fd8e39a`（tree `80e876d067cb9cde123ec3fa66960b323edc24b0`，merge-commit attempt 1）。Audit permissions/menu、真实 SQLite capability registry/system-admin/idempotency/tamper 测试、全量 Go/前端/合同/治理、固定 Go Task `task:contract` 与 SQLite/PostgreSQL 17.11/Chrome 151 required E2E 全部通过；`public` 0 表、非系统 schema、进程和 E2E temp 零残留，T-11 恢复为 done。
- 用户最新决定把 T-10、T-12、T-13、T-15 至 T-21 的逐 Ticket E2E 全部延后：各实现通过完整非 E2E candidate Gate 后可进入 `main` 并解除下游实现依赖，状态保持 `implemented-pending-final-e2e`；T-21 实现集成后由 Lead 创建唯一最终系统候选并统一运行一次完整 E2E。既有已完成 Ticket 的 E2E Evidence 保留有效，远端、生产、签名、公证、发布与 source cleanup 权限不变。
- T-16 source 已在同一第 4 次修复链更新到 clean `fdfa4afbe342edbd5b6feb00a8699780dd04caf4`（tree `792f1dc446648f32469054e974e75aa9bab28c1c`）：补齐 release `custom-protocol` 资产、阶段安全诊断、有界 Accessibility 查询、预期宿主退出竞态处理和 macOS 先显示窗口后激活应用的顺序。锁屏不再是当前执行 blocker；同一 retained candidate 只需合入该 checkpoint 与最新父状态并重跑全部非 E2E Gate，native E2E 归入最终统一系统验证。
- T-16 最终 source `fdfa4afbe342edbd5b6feb00a8699780dd04caf4` 已在同一 retained attempt 4 形成 candidate/result `cd50bfdbd24517c8b3e8f3565d75fae8c3f6a22a`（tree `901fdbb2d967164178da02faabba7e1811af6caf`，parent before `ce21e84a432e5d882634399c4e607c94866442f9`）。Go 普通/SQLite/tag/race/vet/CGO0/module、25-package frozen/83+6+6/build、Rust 16+17 tests/clippy/compile guard/locked custom-protocol release、合同 28/28、API client 7/7、governance、`TASK_CONTRACT_PASS`、SpecDev/path/clean 与 production artifact/resource scan 全绿；`main` 已快进到同一 result。T-16 为 `implemented-pending-final-e2e`，native tracer 延后 G8，现解除 T-17 的 T-16 实现依赖。
- T-15 首次 parent-candidate preflight 证明 source 的两个 Generator package manifest 已声明真实依赖与 package-local checks，但根 `verify` 未调用其 typecheck/tests，lockfile 两个预建 importer 仍为空。Lead 批准 `T15-D02`：在 T-16 result 后只开放根 Generator typecheck 聚合、Vitest 两个 Generator include 和 lockfile 两个既有 Generator importer；禁止修改其他 script/include/importer/catalog、引入新版本或触碰 composition。
- T-15 最终 source `8458ceac6a5ce05a1eec811374f60b2f33a83729` 已在 attempt 1 形成 candidate/result `b60d1bc33d702fdaac0cef7a28dad27582e5b975`（tree `38fa0f96eefd74bcbd5ec4e71e7df979d923d7cd`，parent before `b01540bf21d42cdf49270f50b39ca7d253f7b198`）。候选闭合重复 Vite raw 类型与测试复用过期 parent deadline 两项集成问题；真实生成/编译、SQLite full/focused race/vet/CGO0/module、15 files / 90 frontend tests、Web/Desktop build、合同 28/28、API client 7/7、governance、`TASK_CONTRACT_PASS`、path/clean 和 Goal Plan validator 0 error 全部通过，`main` 已快进到同一 result。T-15 为 `implemented-pending-final-e2e`，Web wizard 与跨 profile 场景延后 G8，现解除 T-17 的 T-15 实现依赖。
- T-10 clean source `a630c50fd0f8f22208564a807accfb325ca34f6a`（tree `f36adb6b6e6243fd15d1867b586086939597d536`）独立审查发现 runtime/secret key marker 可绕过、保存失败清空表单、分类/字典切换失败后投影错配，以及根 verify/lock 未接入 Settings。Lead 批准 `T10-D02`：在 T-15 result 后只开放根 Settings typecheck、两个 Vitest include 和两个既有 Settings importer；候选必须闭合前三项产品 finding，禁止修改其他 script/include/importer/catalog、共享 UI 或 composition。
- T-10 最终 source `a630c50fd0f8f22208564a807accfb325ca34f6a` 已在 attempt 1 形成 candidate/result `030eb0b57fd47a4a5381f80a24cd3dafaeb35310`（tree `8cdd579f60f3018be864f55fb54cfa8b9fb8ab3f`，parent before `71108aea07acef377157480c27844096debb6283`）。候选闭合 runtime/secret key 绕过、失败清表单、选择投影错配和 SQLite migration 尾部 whitespace；Go normal/SQLite full/race/vet/CGO0/module、18 files / 103 frontend tests、Web/Desktop build、合同 28/28、API client 7/7、governance、`TASK_CONTRACT_PASS`、41-path/diff/clean 和 Goal Plan validator 0 error 全部通过，`main` 已快进到同一 result。T-10 为 `implemented-pending-final-e2e`，双方言 Settings API/UI 与安全矩阵延后 G8，现解除 T-17 的 T-10 实现依赖。
- T-12 source `a9b8a568b17934db78ea2d89d258cf5dfa065fad` 的 41 个路径与 diff-check clean；既有审查已闭合 timeout 误记成功、长任务补跑 backlog 和软删除名称不可重建。Lead 批准 `T12-D02`：在 T-10 result 后只开放根 Scheduler typecheck、两个 Vitest include 和两个既有 Scheduler importer；禁止修改其他 script/include/importer/catalog、composition、Outbox 或 coordination 实现。
- T-12 最终 source `a9b8a568b17934db78ea2d89d258cf5dfa065fad` 已在 attempt 1 形成 candidate/result `d9e7d5aa5fd973e76a6d3c5b0a8648db04df7b71`（tree `fe2415cac6f73942f4507ed50e24da8fb13907c8`，parent before `3d8bc072241c7e76b4f9baf6793f26fdd5d30896`）。Go normal/SQLite full/race/vet/CGO0/module、21 files / 111 frontend tests、Web/Desktop build、合同 28/28、API client 7/7、governance、`TASK_CONTRACT_PASS`、44-path/diff/clean 和 Goal Plan validator 0 error 全部通过，`main` 已快进到同一 result。T-12 为 `implemented-pending-final-e2e`，多实例/lease-loss/Outbox/browser 矩阵延后 G8，现解除 T-17 的 T-12 实现依赖。
- T-13 source `fd43839c296047ed2e78cc9edb1b106b5b4129ca`（tree `ed8e335f6efc723f3df91a4a8e39e25a64747318`）路径与 diff-check clean；既有审查已闭合 ready timestamp 与 Unicode control 分叉，并保留 Browser adapter 精确依赖边界。Lead 批准 `T13-D03`：在 T-12 result 后只开放根 Files domain/web-domain/Browser adapter typecheck、三组 Vitest include 和三个既有 importer；Desktop/config bridge 与 composition 仍归 T-17。
- T-13 最终 source `fd43839c296047ed2e78cc9edb1b106b5b4129ca` 已在 attempt 1 形成 candidate/result `0f903d89f0e96f2afa0bd67348da3c7698407292`（tree `d575fe1445f117e69d69eab00b6ab8e80e85ab24`，parent before `595fc61463cafa30cd34df23992576c4cc02adc2`）。Go normal/SQLite full/race/vet/CGO0/module、24 files / 128 frontend tests、Web/Desktop build、合同 28/28、API client 7/7、governance、`TASK_CONTRACT_PASS`、46-path/diff/clean 和 Goal Plan validator 0 error 全部通过，`main` 已快进到同一 result。T-13 为 `implemented-pending-final-e2e`，双方言/filesystem/browser/Desktop 联合矩阵延后 G8，现解除 T-17 的 T-13 实现依赖。
- T-17 stage 1 从 clean `main@0cf98457a3eb68b636a45cd17dcc62b24d01d4e1` 激活，source worktree/branch 固定为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-17` / `speculo/2026-08-26-project-architecture-reconstruction/T-17`，implementation/integration owner 均为 `codex-root`。`T17-D01` 只开放 app-shell manifest 的 product subpath export 与 package-local checks；先在独占 product/cmd/test 路径建立完整注册图和失败基线，Apps/adapters/Rust/root/module 路径等待实际证据后串行 amendment。无调用语义的 IAM Organization Port 不以仅构造 adapter 伪闭合。
- T-17 首次双方言 product migration Compose 均被平台 forward-only validator 拒绝；只读定位到 Generator 的 SQLite/PostgreSQL `6400000000000_generator_configs.sql` 仍含 `+goose Down` 与 DROP。Lead 批准 `T17-D02`，只开放两份同版本 SQL 删除 Down 段；Up schema、版本、provider、其余 Generator 实现、根配置、Apps/adapters/Rust 均零漂移，修复后须重跑双方言 Compose、真实 SQLite 全迁移/能力注册及 Generator 聚焦门禁。
- T-17 stage-1 checkpoint 为 `92a0ff8a56226fd317c20e71e1665822c9226595`（tree `50965ffd689dfe1a3e79bd3468ff07af8d1ca4a5`）。只读 App/native preflight 证明 Web 仍占位且 Browser Permission parser 使用冒号，Desktop App/adapter/Rust/sidecar 仅 Demo，双 App 未消费 product manifest。Lead 批准 `T17-D03`：只开放 app-shell 产品工作区、双 App 入口、Browser parser、Desktop transport/Rust allowlist/proxy、sidecar product runtime，以及精确 manifest/Vitest/boundary/lock importer；模块业务实现和最终 E2E 不重开。
- T-17 stage-2 首轮 aggregate test 证明 platform Permission 类型与旧 Web runtime fixture 仍停留在冒号 convention，和 IAM registry、模块合同、数据库能力及 product runtime 的点分格式冲突。Lead 批准 `T17-D04`：只修改 platform PermissionCode 类型和该 fixture 为 `module.resource.action`，禁止双格式兼容与其他 platform/shell 漂移。
- T-17 完整 TypeScript gate 进一步证明 app-shell fixture 也是该类型的直接消费者；`T17-D05` 只把其中 5 个冒号权限字面量改为点分格式，不改变 Shell 行为、断言或其他 fixture。
- T-17 最终 clean source checkpoint 为 `e3abed759ac47f8f154c4638aaeab9eac9a41b7d`（tree `77fe889cbc8fb19286200786df9d312fbc708ba6`）：共享 ProductWorkspace、双薄 App、Desktop product sidecar、精确 Rust product/binary bridge 与点分 Permission 迁移已完成；普通 Go/Generator 真实 compile gate、前端/Rust/合同/治理门禁通过，full race 与 parent-candidate pending。
- T-17 attempt 1 candidate/result 为 `04d7bdfe5aa26583c4101dd4e7dfbe8c67a864e8`（tree `779f538fbae8e983caa2fd67b55821e798ad8f2d`，parent before `ab01846437680619fd55070c52d0bd7d366a582a`）。source/candidate 的完整 Go normal/race/vet/build、25-workspace lint/typecheck/132 tests/7 boundaries、Web/Desktop build、Rust 14 tests/clippy/check、sidecar resource、合同、governance、Task contract、34-path/clean 与 SpecDev Gate 全部通过，`main` 已 `--ff-only` 提升。T-17 为 `implemented-pending-final-e2e`，现解除 T-18/T-19/T-20 的实现依赖。
- T-18 已从 `main@479fcc9190d38d14a185847d132f41b972c830a0` 激活，source worktree/branch 固定为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-18` / `speculo/2026-08-26-project-architecture-reconstruction/T-18`，implementation/integration owner 均为 `codex-root`。只写 Linux release/Compose/script/workflow 专属路径；本机无 Docker，真实双架构运行延后最终 Linux runner，当前必须闭合静态构建、policy、失败阻断和供应链合同。
- T-18 clean source checkpoint 为 `e815b16ae8ac7ca73a72f99200ef004ad9b187e4`（tree `32ea206047c01335be5c887c352f47fef661d685`）：双架构固定 digest OCI、PostgreSQL/SQLite Compose profile、secret refs、Generator 完整工具骨架、non-root/read-only runtime、两架构 x 两 profile protected matrix、checksums/SBOM/provenance 和专属 policy tests 已闭合。25-path contract、本地双架构 Go/Web build、policy/parse/secret/base-manifest Gate 通过；Docker 平台场景按统一 E2E 决策保持 pending。
- T-18 attempt 1 candidate/result 为 `37f989c7c64afb17c05bdeb35818c34416e753ca`（tree `855de1252161a78f933d9df98787815d55008646`，parent before `39f2efe4ce8701165e07a86adb3322c8673459c1`）。完整 Go normal/race/vet/build、前端 132 tests/7 boundaries/双 build、Rust 14 tests/clippy/check、合同 28/28/API 7/7、Linux policy、governance、Task、25-path/clean 与 SpecDev Gate 全部通过，`main` 已快进提升。T-18 为 `implemented-pending-final-e2e`；真实 Docker 双架构矩阵延后 G8。
- T-19 已从 `main@45e37678f695ef9c3b6333738b7cc5eb3f480e86` 激活，source worktree/branch 固定为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-19` / `speculo/2026-08-26-project-architecture-reconstruction/T-19`，implementation/integration owner 均为 `codex-root`。只读预检确认旧 workflow 仍构建 Wails ARM64，identity 与 Tauri `com.goadmin.plus` 不一致；同时 T17 sidecar 会查找源码仓库且 Tauri 清空全部环境，安装包既无法启动完整 Generator，也不能解析 Go/Node/pnpm/git。用户 `Q1A/Q2A` 批准 `T19-D01`：除 macOS release/script/workflow 外，只精确开放 Tauri `main.rs`，让开发态白名单传递本机工具链、发布态固定到 App Resources 内的只读离线 skeleton 与双架构工具链；不改变产品 API、sidecar 生命周期或 Generator 验收合同。
- `T19-D01` 后续只读审计证明 Generator compile gate 会丢弃显式 `GOENV=off`、`GOTOOLCHAIN=local`、`GOPROXY=off` 与 `GOSUMDB=off`，令 macOS/Linux 的离线发布声明失真。该偏差只追加开放 compile gate 实现与既有测试，强制四个固定 policy 值进入子进程；不开放 renderer/writer、公共接口或其他模块路径。
- implementation commit 和 Local candidate integration and parent update 已由用户 `Q2A` 授权；source/integration/candidate cleanup 已由用户 2026-08-28 后续指令授权并完成，远端和生产动作仍未授权。
- T-19 source checkpoint 为 `6c97064f78719b6a9ceaa91d8a6fed5d92b5695d`：Universal Tauri host/sidecar、production-only 签名/公证 Gate、自包含双架构 Generator 工具链、离线 Go/pnpm store、checksums/SBOM/provenance workflow 和失败 policy 已闭合。pnpm 只保留官方 `supportedArchitectures` 的 macOS x64/arm64 原生包，本地 unsigned App 从错误的约 3.5 GB 全平台依赖候选收缩到约 2.2 GB；生产签名、公证、DMG 安装与完整原生 E2E 仍按 G8 pending。attempt 1 `fc48f749c418235464ce2b441c8fe72131ca21b8` 在 clean Rust Gate 证明 workflow 未先构建 host sidecar 后作废；新 checkpoint 已补 sidecar-before-cargo policy，attempt 2 必须完整重建。
- T-19 最终 source `6c97064f78719b6a9ceaa91d8a6fed5d92b5695d` 已在 attempt 2 形成 candidate/result `f18d765246aad076f92c5ae8fda2732283a4186a`（tree `70dfb8f37d17314fd19932b93f11ab3953722a3b`，parent before `53d3d6cba17d3e33c983a6218ad0dc51c5371668`）。Go normal/SQLite/race/vet/CGO0/module、25-workspace lint/typecheck/132 tests/7 boundaries/双 build、Rust 19 tests/clippy/release check、合同 28/28/API 7/7、governance、Task、macOS policy、Universal candidate App/resource/path/clean 与 Goal Plan Gate 全部通过，`main` 已快进提升。T-19 为 `implemented-pending-final-e2e`；正式签名、公证、DMG/Gatekeeper 与完整原生场景延后 G8。
- T-20 已从 `main@af22c2cf3d721937a4c1d8236776b47a309d4a8c` 激活，source worktree/branch 固定为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-20` / `speculo/2026-08-26-project-architecture-reconstruction/T-20`，implementation/integration owner 均为 `codex-root`。只写 Windows release/script/workflow 专属路径；生产 Authenticode、受保护 runner 和原生安装/重启/卸载场景仍按 G8 与独立授权 pending。
- T-20 只读预检证明 T19 发布环境会把 Windows `x86_64` 误映射为 `darwin-amd64`，且 Unix 可执行路径不能启动 Windows packaged Generator。Lead 批准 `T20-D01`：只精确开放 Tauri `main.rs` 的 OS + architecture toolchain 选择、Windows 必需环境白名单与对应测试；产品 transport、vault、sidecar 生命周期、其他 Desktop 文件和 macOS 行为保持不变。
- T-20 clean source checkpoint 为 `faebfeec3bdaad8c7c644b1bd93f3bee4ebc7fa2`（tree `a2f9c568304619d9a1f08ab7fa6dab3ccaa5d6a6`）：Tauri 2 x64 NSIS、production-only Authenticode 精确证书 policy、packaged Windows Generator 工具链、安装/重启/卸载数据边界、checksums/SPDX/provenance 与 protected workflow 已闭合。21-path contract、Go/Windows PE sidecar、25-workspace 132 tests/7 boundaries/双 build、Rust 20 tests/clippy/release check、合同 28/28/API 7/7、governance、Task、Windows policy/PowerShell parse/path/clean 全绿，双轴复审 pass、0 findings；正式签名与原生 Windows E2E 仍按 G8 pending。
- T-20 attempt 1 candidate/result 为 `10aecc19dab8cb2a80c0fc8e16af6ecab2208d10`（tree `4ed4cbb977188ca09331ca191beb324250265122`，parent before `1d2cfd0075de565cd584d7a090e4cf48db41a29a`）。Go normal/SQLite/race/vet/CGO0/module、Windows x64 PE sidecar、25-workspace lint/typecheck/132 tests/7 boundaries/双 build、Rust 20 tests/clippy/release check、合同 28/28/API 7/7、governance、Task、Windows policy/PowerShell AST/path/clean 与 Goal Plan Gate 全部通过，`main` 已快进提升。T-20 为 `implemented-pending-final-e2e`；正式 Authenticode、原生 NSIS 安装/重启/卸载与完整业务场景延后 G8，现解除 T-21 的最后实现依赖。
- T-21 已从 `main@5fb6f86dfd52dcd7aad93ab7380b9089e14c02b9` 激活，source worktree/branch 固定为 `specdev-worktree/2026-08-26-project-architecture-reconstruction/T-21` / `speculo/2026-08-26-project-architecture-reconstruction/T-21`，implementation/integration owner 均为 `codex-root`。G7 只直接修改最终根 CI/docs/quality 与 Ticket 精确旧路径；先冻结 12 项治理删除 inventory 和 allowlist 外 RED 扫描，再实施原子收缩。Speculo 历史、source/candidate worktree、生产/远端动作均不在删除范围。
- T-21 RED 依赖闭包证明三个正式 command 不消费旧 `app/common/api/tenant`，但未被入口消费的 expand scaffolding、根 Task/CI、旧 Windows tracer、聚合 release manifest 和文档仍会令显式目录删除后全仓失败或残留 Wails/旧前端/unsigned-self-use。Lead 批准 `T21-D01` 精确开放 Ticket 新增路径：删除这些无正式消费者的旧包/command/template/test/assets，tidy 后端依赖，并把根治理接到新入口和三平台 policy；新 product/module/platform、三正式 command、前端新 Workspace与平台资产保持只读。
- T-21 首次 compile RED 纠正过删：`internal/application`、`internal/contracts`、`cache/localcache/observability` 均属于新架构并已精确恢复。治理预检另发现前端管理脚本使用不存在的 Desktop package filter，Linux dockerignore 与配置 README 仍引用旧前端；Lead 以 `T21-D02` 只开放这 7 个文件改为当前路径/package 合同，不修改产品源码。
- G8 attempt 1 基于 `main@b72af2cbcc715cb1476dcf6856f5948b18910d43` 创建于 `specdev-candidate/2026-08-26-project-architecture-reconstruction/G8-attempt-1`。IAM Session/Administration、Audit、Demo、Organization 双数据库 browser 组通过后，Settings 暴露 category 投影提前提交、敏感 intent 残留和多 Web client 滚动 CSRF 分裂，候选按 repair loop 整体失效。修复 source `76efe495616470bb985952e130f3253cd5e1745d` 经独立 candidate/result `45ba95b1e2022aba8831ad5b4333f39a7de62b87`（tree `8b09abdfed5e330bac3dbacf6fa627249f0e8b59`）通过 25-workspace lint/typecheck、27 files / 136 tests、7 boundaries、双 build、Settings Go、root governance/architecture/compatibility/docs/Task 门禁后进入 `main`；下一步必须从最新父分支创建 G8 attempt 2 并从头完整重跑，attempt 1 的部分 pass 不累计。
- G8 attempt 2 基于 `main@fe156d7579cd4d8be98eaef6db3229b50d44aeff` 创建于 `specdev-candidate/2026-08-26-project-architecture-reconstruction/G8-attempt-2`，因 IAM PostgreSQL harness 将 `search_path` 写入不会被 `pgx.ConnString()` 序列化的 runtime map、实际污染 `public` 而整体失效。R02 source `f8a664e142c244224ef18bdfed21a7f6a1e79df2` 经独立 candidate/result `51c6dec898983261f1633925762d1e0f5784e27d`（tree `1f5dc8a0aff57c05e24eeb3214bc2ae7292f7156`）闭合四处 URL-safe schema isolation、Scheduler rolling-CSRF client 和首次表单快照初始化；双数据库 IAM/Scheduler、PostgreSQL runtime/concurrency、Go normal/SQLite/race/vet/CGO0/tidy、27 files / 136 tests、7 boundaries、双 build、Tauri 2 Rust 19 tests/clippy/release、合同与全部 root Gate 全绿。命令上下文曾令同 tree source 提前进入本地 `main@c2aae797540f615ac478c5253b0ed2a385b7c2da`，正式 candidate 后以第二父节点集成为 `main@e92a1aecdf1c0c20fc735afeda53440aba6c89b1`；历史未改写、远端未推送。下一步必须从该最新父分支创建 G8 attempt 3 并从 IAM Session 完整重跑，前两次部分 pass 和 source repair E2E 均不累计。
- G8 attempt 3 基于 `main@bb26303f70717235e7f57c6c47e9ffdf66040882` 创建于 `specdev-candidate/2026-08-26-project-architecture-reconstruction/G8-attempt-3`。IAM Session/Administration、Audit、Demo、Organization、Settings、Scheduler 双数据库 browser 组通过后，Files 暴露未共享 rolling-CSRF、fixture 绕过 Vue checkbox 响应式 selection，以及 revoke fixture 绕过页面 revision/sessionRequired 事件链，候选按 repair loop 整体失效且部分 pass 不累计。随后 Generator source 验证继续暴露 SQLite metadata 类型、primary sort 重复 case、条件 import 和 PostgreSQL actor IAM identity 类型漂移。R03 source `614355dce9450e88c0a5758c3909fd20999f6e12` 经独立 candidate/result `d41691a6bbf8e0246e5cf2b82bfa86873716a9b4`（tree `74ca87f723e99857bda4731ed0089c2f588b1690`）闭合全部问题；Files/Generator 双 profile 修复 E2E、Go normal/SQLite/race/vet/CGO0/tidy、27 files / 136 tests、7 boundaries、双 build、Tauri 2 Rust 19 tests/clippy/release、16 release policy tests、合同与全部 root Gate 全绿，`main` 已 `--ff-only` 提升。该最新父分支随后用于 G8 attempt 4；前三次部分 pass 和 source repair E2E 均不累计。
- G8 attempt 4 基于 `main@eb5de678c253c563343d31bba289840464eec60c` 创建于 `specdev-candidate/2026-08-26-project-architecture-reconstruction/G8-attempt-4`。全部 browser module 双 profile 与 PostgreSQL reliable-runtime fault 通过后，Desktop native 预检发现当前架构缺少真实 fixture，候选整体失效且部分 pass 不累计。R04 source 以 `d28e5531666afb54163f036180ed8f2adf6df6e0` 恢复 fixture/native 全链并以 `f49aa4c3e3f3ef5a46c63c57d03f4efa7a568914` 修正 error-only phase 诊断；Attempt 2 candidate/result `8f817bf014e464e5c8d4411851cadc0cee293d48`（tree `80d7603f521d420835f1cff2ee74900e791d7525`）通过全部非 E2E Gate后进入 `main`。用户随后要求暂停 E2E；当前编码/构建完成点为该 SHA，G8 与 change 保持未完成。
- 用户授权的 worktree cleanup 已完成并同步状态。Lead 随后在 `main@c308ba7d03fc9ca10156d3adac0849052629cae8` 当前产品代码上重新执行全部非 E2E Gate：Go normal/SQLite/race/vet/CGO0/module/tidy，25-workspace pnpm lint/typecheck/27 files 136 tests/12 desktop runner tests/7 boundaries/双 build，Tauri 默认 20/native-feature 21 tests/clippy/custom-protocol release，OpenAPI lint/28 contract tests/API client 7 tests/generate-check，root governance/architecture/compatibility/docs/Task 与 16 release policy tests，以及 production artifact test-control 零命中全部通过。该复验没有运行 browser、fault、native 或发行 E2E，G8 继续暂停。
- 最终源码级 UI 审计发现当前 `ProductWorkspace` 与重构前基线的 layout/Navbar/Sidebar/TagsView/design tokens/登录页不一致，违反用户要求的“页面 UI/CSS 保持、只删功能”执行约束。`T05-D04` 以 `main@22d1d65394462bb90bb88366e2244b9ed8715ce2` 为 base，result `4cd5c506d1317e961a57cb58e509c2ad5b2c0259`（tree `cb1e7d3ed4f469948e1308e0d62d5be4f2c7fc63`）在当前 pnpm 边界内恢复共享 design tokens、深色可折叠侧栏、50px 顶栏、面包屑、40px 页签、移动抽屉及终端舞台登录页；验证码、旧 store/router/import 和已删业务能力零恢复。frozen install、lint、typecheck、28 files / 139 tests、7 boundaries、Web/Desktop production build、root 静态治理及 CI 固定 `go-task v3.48.0` 的任务合同全绿；按用户要求未运行任何 E2E，G8 保持暂停。
- `T05-D04` 后的逐页复审继续发现 retained management pages 尚未继承旧版中文与紧凑后台表达：IAM Administration、路由及多个页面仍有英文，局部 scoped CSS 覆盖共享 theme，外层 spacing 和查询/操作/编辑/分页布局也未形成防回退合同。`T05-D05` 以 `main@7a14cd01e2426aca929c8a1baea2506774436b48` 为 base，result `f6f852cf715df2d0c26bb583ecac10639f578b30`（tree `44c0a3e5a1a89791669c47d3d85d41d9209e73ba`）完成 retained routes/pages 中文化、旧 palette 收敛、`12px` PageContainer spacing、共享紧凑管理布局和实际 Vite `5173` 文案，保持 controller/API/Permission/field/test-id 与删减后能力不变。前端 lint/typecheck/28 files 140 tests/7 boundaries/双 build、Go normal/SQLite/vet/双 profile 与 CGO0 build、OpenAPI 28/28/API 7/7/generate-check、root Task/治理/架构/零兼容/文档/16 release policy，以及生产 sidecar、Rust 20 tests/clippy/custom-protocol release 全绿；未运行任何 E2E，G8 继续暂停。
- `T05-D05` 后的交互复审继续发现 retained CRUD pages 把原有“工具栏入口 + 新增/编辑弹窗”退化成常驻编辑器，暂停的 browser/native driver 仍引用旧英文可见标签，Scheduler settle helper 还会丢弃 command result 并阻止成功弹窗关闭。`T05-D06` 以 `main@8165fffae1e659b732e563fe7a10b46dd56089bf` 为 base，result `89f70304c8227d7401d7791b7f7c64e0f1cd2e84`（tree `1cf6552b289ee3112638ee2facc59c594566e447`）在现有无头 controller/Permission/双宿主边界内恢复 IAM、Organization、Demo、Settings、Scheduler 的显式工具栏与响应式弹窗，清空取消态与密码，中文化 Audit/Scheduler 枚举，并同步暂停驱动。前端 lint、全 typecheck、28 files / 143 tests、7 boundaries、7 个 driver 独立编译、19 个 Node 静态测试及 Web/Desktop production build 全绿；按用户要求未运行任何 E2E，G8 继续暂停。
- 最终文档审计发现两份项目 Agent Skills 仍描述已删除架构，且根零兼容与文档检查未扫描 `.agents/skills`。`T21-D04` 以实现提交 `95663eb35d8554f5f66c9b8fdc965e9914943d50`（tree `23427ee92cec831ae1371fcc646fa30dcfc8f460`）重写业务模块/列表页工作流并扩展治理覆盖；两份 Skill 校验和临时打包、compatibility/docs/architecture、质量回归与 diff 检查全绿，产品源码零变更且未运行 E2E。

### Pending Decisions and Blockers

- 无阻止 T-02 开始的产品决定。
- T-16 与 T-15 的 implementation result 均已提升；各自 E2E 延后到最终统一系统候选，不再构成后续实现 blocker。
- G6 的生产签名、公证、受保护 runner、远端制品发布仍需到达该 Gate 后逐项批准；在此之前 T-19/T-20 不能以模拟或未签名制品关闭最终发行验收。
- source、integration 与 candidate branch/worktree cleanup 已获授权并完成；所有历史 locator、不可变 SHA、验证和 E2E disposition 继续保留在状态与 Evidence 中。

### Resume Protocol

恢复时依次读取本 Goal Plan、当前 Ticket、`<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/.status.json</Path>`、Tickets Map 和最新 Evidence。先核对 `main` 当前 HEAD 是否包含最后通过的 `result_sha`；若存在 active source 或 candidate，从其不可变 checkpoint 继续，不重建已通过工作。若父 HEAD 漂移，暂停派单并重算所有未开始 worktree 的 base；若超过 4 次集成尝试或需要新产品决定，标记 blocked 并返回用户。

## Assumptions

- 低影响假设：父分支逻辑名称继续使用 `main`；实际 Git 分支经实测为 `main`。
- 低影响假设：implementation provider 和模型在每次 execution-time dispatch 时选择；本计划不要求外部网页交付。
- 低影响假设：正式依赖版本由对应 owner Ticket 以可审计 lock/digest 固定，不能使用浮动的“最新”。
- 低影响假设：当前 25 个 lint warning 只作为旧基线记录，不允许新增，最终由新 Workspace/原子收缩自然归零。
