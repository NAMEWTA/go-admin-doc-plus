---
schema_version: 6
artifact: goal-plan
change: 2026-08-29-productization-and-ui-reconstruction
status: ready
modes: [migration, high-assurance, reference-conformance, release-coordination]
orchestration: lead-directed
lead: codex-root@2026-08-29-productization-and-ui-reconstruction/epoch-1
implementation_agent_limit: 3
integration_attempt_limit: 6
ticket_workspace_policy: required
integration_gate: candidate-merge
ready_for_execution: true
---

# Goal Plan: Go Admin Plus 产品化与 UI 重构

- **Goal Plan：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/goal-plan.md</Path>`
- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **Tickets Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **Ticket 目录：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/</Path>`
- **Evidence 目录：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/</Path>`

## 1. Outcome and Authority

### Outcome

在不再次搬迁目录、不恢复旧兼容层、不引入租户/Redis/JWT/Casbin 的前提下，将当前 `main` 演进为可从空库安全初始化、可管理、可恢复、可观测并有真实发布证据的 Go 管理脚手架。最终候选必须同时交付 Server SQLite、Server PostgreSQL 与 Desktop SQLite 三个 Profile；Web 与 Desktop 共享当前 OpenAPI、领域语义和管理体验，后端仍是身份、授权、数据范围和不可逆操作的事实源。

### Success and False Completion

成功只在以下事实同时成立时成立：

- `AC-001~039` 均有 Lead 核对的 Evidence，21 个非取消 Ticket 均有 source commit、通过的 parent-candidate、父分支 `result_sha` 和路径核对。
- Bootstrap/Recovery、Session/CSRF、数据范围、账号生命周期、文件容量、CLI/Profile、Router/UI 与 Generator 在 SQLite/PostgreSQL/Web/Desktop 的规定接缝上成立。
- T-18 的真实 PostgreSQL/安全供应链、T-19 的真实浏览器、T-20 的真实 macOS Tauri、T-21 的三 Profile clean-room 均实际运行；缺环境、skip、not-run 或 source-worktree 自报不得算通过。
- `main` 只由 Lead 在候选验证通过且 HEAD 未漂移后推进；所有 Speculo 工件、Git checkpoint 和实际状态一致。

以下均属于伪完成：只编译不运行、只测 SQLite、用 mock 代替 required E2E、以隐藏按钮代替后端授权、保留固定管理员密码或旧初始化路径、让 API/worker 暗迁移 PostgreSQL、把签名/公证的 `not-required` 写成 `passed`、用 Evidence-only 或空 commit 关闭 Ticket。

### Non-goals

- 不重新命名或搬迁 `go-admin-plus/`、`go-admin-plus-ui/`、`deploy/`、`release/`、`database/` 与根治理布局。
- 不兼容旧目录、旧 API/schema/config/命令、默认账号、租户、Redis、JWT、Casbin 或旧 UI 状态模型。
- 不复制 Backplane/RuoYi 的依赖栈、默认凭据或弱安全实现；只吸收运维易用性与后台交互成熟度。
- 不新增公开未认证初始化/恢复 API，不引入 Desktop 之外的第三个交付 App。
- 本计划不授权 push、PR、remote merge、部署、生产迁移、发布、签名、公证、归档或 branch/worktree/stash 清理。

### Authoritative Inputs

| 优先级 | 来源 | 负责内容 | 冲突处理 |
|---|---|---|---|
| 1 | 用户最新决定：推荐方案全部接受；2026-08-31“都批准” | required worktree、本地 commit、candidate integration、`main` 更新、required runner 本地准备，以及 DEV-09-001 联合候选/T-09 提前组合检查点授权 | 新决定先更新真正 owner，再修订本 Plan；不包含 push、部署、发布、生产迁移或清理 |
| 2 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ADR.md</Path>` 与 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/CONTEXT.md</Path>` | 本 change 的身份、数据、Profile、UI 与恢复决定 | 返回 `<Path>{roots.workflows}/specdev/G-grill-with-docs/G-grill-with-docs.md</Path>`，不得由实现者改写 |
| 3 | `<Path>{roots.state}/specdev/adr/</Path>` 与 `<Path>{roots.state}/specdev/context/</Path>` | 已毕业的模块化单体、pnpm/Tauri/OpenAPI、无租户/Redis与质量门禁 | 本 change 替代时须在 ADR/LOG 明示 |
| 4 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>` | `US-001~017`、`AC-001~039`、范围与外部行为 | 下游只能实现或触发偏差，不能改写验收 |
| 5 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/</Path>` | 单 Ticket 的局部实现、路径、验证与恢复合同 | Ticket frontmatter 是单 Ticket 权威 |
| 6 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/PLAN.md</Path>`、固定研究 URL 与参考项目快照 | 审查来源、分阶段理由与 reference-conformance | 仅作约束来源，不覆盖 ADR/Spec |
| 7 | `main@904d66a5afaff58037798d4d17b96bc69c1988f5` 与执行时实测 | 当前代码、工具、工作树与可运行性 | 冲突时暂停受影响 Wave并返回权威 owner |

## 2. Execution Graph

### DAG and Critical Path

```text
Wave A: T-01  T-02  T-03  T-04  T-06  T-07  T-11
                   \     /                 |
Wave B:             T-05                  T-12
                      \                    /
Wave C:               +------ T-08 -------+
                               |
Wave D:                T-09  T-13  T-15  T-16
                         |     |      |      |
Wave E:                T-10   +---- T-14 ----+
                                      |
Wave F:                            T-17
                                      |
Wave G: T-01 + T-09 + T-17 ------ T-18
                                      |
Wave H: T-13 + T-14 + T-15 + T-16 -> T-19
                                      |
Wave I: T-10 + T-14 + T-15 + T-16 -> T-20
                                      |
Wave J: T-17 + T-18 + T-19 + T-20 -> T-21
```

关键路径为 `T-02/T-04 -> T-05 -> T-08 -> T-13 -> T-14 -> T-17 -> T-18 -> T-19 -> T-20 -> T-21`。T-08 是公共 wire contract 汇合点，T-14 是安全数据模型与前端组合汇合点，T-17 是当前脚手架冻结点，T-18~21 是不可跳过的连续证据链。Wave 表示依赖和路径允许并行，不要求填满并发上限。

### Approved Sequencing Checkpoint (DEV-09-001)

为解除“领域 migration 必须先进入产品 registry 才能通过 parent-candidate full suite，而 registry owner T-09 又等待这些领域 result”的闭环，用户于 `2026-08-31T22:50:26+08:00` 明确批准 ticket/release 级执行偏差：

- Lead 可从最新 `main` 构建一个可审计的 Wave A 后端联合候选，按固定顺序合入已完成双轴审查的 T-01/T-02/T-03/T-04/T-06/T-07 source commit；每个 source SHA、父 SHA、冲突和候选树仍须逐项记录。
- T-09 shared-path owner 可在该联合候选上提前建立仅限 `internal/app/product` migration/provider registry、probe/readiness 所需最小组合接线的 source checkpoint；不得借此提前实现 CLI、scripts、Compose 或改写 T-08 wire contract。
- DEV-03-001 允许 T-03 精确修正 2 个 Audit disposable fixture；T-04 继续按既有目录所有权修正 2 个 Authorization fixture。四者仅显式应用 0040 migration；旧 rotate-on-read HTTP 断言保持为 T-08 红灯，不得在领域服务中恢复 GET 写入。
- T-09 的 `blocked_by: [T-07, T-08]` 继续约束完整 Ticket closure/result；提前检查点不是 T-09 result，也不解锁 T-10/T-18。
- 联合候选仍必须补齐 required race 与真实 PostgreSQL 证据后才能晋升。既有 Windows 平台失败只能按 owning Ticket 保留为明确失败证据，不得改写为 passed；G5~G8、T-18~21 的 required 门禁完全不变。
- **Wave A bridge promotion contract：** 用户“都批准”同时批准修订本轮 full-suite 进入条件。T-01/T-02/T-03/T-04/T-06/T-07 只有在全部改动包、产品 migration matrix、`go vet`、architecture、required race、逐 Ticket 独立真实 PostgreSQL 和 reachable vulnerability 检查通过，且 `go test ./...` 已同时在最新 `main` 与候选实际运行并完成差异归因时，才可把联合 parent-candidate 记为 passed。允许携带的红灯仅限两类：两侧完全一致的 Windows 平台基线；以及 T-03 稳定 CSRF/GET 零写必然触发、且由 T-08 明确拥有的旧 rotate-on-read HTTP 断言。它们必须写入对应 owning Ticket，不得记为自身通过，并须在 G4/G5 前清零。

### Waves and Ownership

| Wave | Ticket | 前置条件 | 项目写路径 | Shared owner | Gate/集成序号 |
|---|---|---|---|---|---|
| A | T-01 | G0；固定漏洞基线 | Go module graph | — | G1 / A1 |
| A | T-02 | G0；导入 database 受保护快照 | IAM Bootstrap/Recovery、database bootstrap | — | G1 / A2 |
| A | T-03 | G0；Session 现状基线 | IAM Session/protection | — | G1 / A3 |
| A | T-04 | G0；Organization/IAM port 基线 | IAM scope、Organization projection | — | G1 / A4 |
| A | T-06 | G0；Files 容量基线 | Files service/quota/reconcile | — | G1 / A5 |
| A | T-07 | G0；config/logging 基线 | logging/Doctor | — | G1 / A6 |
| A | T-11 | G0；pnpm lock 基线 | workspace deps、UI package | T-11 | G1 / A7 |
| B | T-05 | T-02、T-04 result | account/file lifecycle | — | G1 / B1 |
| B | T-12 | T-11 result | App Shell 与双宿主入口 | T-12 | G1 / B2 |
| C | T-08 | T-02~06 所列依赖 result | OpenAPI、Go/TS generated、HTTP adapters | T-08 | G2 / C1 |
| D | T-09 | DEV-09-001 允许 migration/provider 组合检查点；完整 result 仍需 T-07、T-08 result | product root、CLI、Task、scripts、compose | T-09 | G3 / D1 |
| D | T-13 | T-03、T-08 result | Browser/Desktop Session adapters | — | G3 / D2 |
| D | T-15 | T-11、T-12 result | Audit/Settings/Scheduler domains | — | G3 / D3 |
| D | T-16 | T-06、T-08、T-11、T-12 result | Files/Generator/Demo domains | — | G3 / D4 |
| E | T-10 | T-02、T-09、T-13 result；导入 `main.rs` 快照 | Desktop setup/native host | T-10 | G3 / E1 |
| E | T-14 | T-04、T-05、T-08、T-11~13 result | IAM/Organization UI | — | G3 / E2 |
| F | T-17 | T-04~06、T-11、T-14、T-16 result | Generator 与 architecture checks | — | G4 / F1 |
| G | T-18 | T-01、T-09、T-17 result | CI、PG/security scripts | T-18 | G5 / G1 |
| H | T-19 | T-13~16、T-18 result | Web E2E root | T-19 | G6 / H1 |
| I | T-20 | T-10、T-14~16、T-19 result | Desktop E2E + DEV-20-002 theme composition + DEV-20-003/004 Web verification | T-20；精确接管 App Shell theme composition 与单一 Web driver 接缝 | G7 / I1 |
| J | T-21 | T-17~20 result | README/docs/deploy/release | T-21 | G8 / J1 |

同一 Wave 内最多激活 3 个 implementation subagent；Lead 可根据测试资源、上下文和写路径降低并发。T-08、T-09、T-10、T-11、T-12、T-18、T-19、T-20、T-21 是其 frontmatter 所列 shared path 的唯一写 owner。

### Ticket Quick Reference

| ID | 可观察产出 | Dependencies | Workspace | Implementation owner | E2E disposition | Evidence |
|---|---|---|---|---|---|---|
| T-01 | 可达漏洞清零 | — | `WT/T-01` | Lead / dynamic dispatch | not-required：扫描与合同测试 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-01.md</Path>` |
| T-02 | Bootstrap/Recovery | — | `WT/T-02` | Lead / dynamic dispatch | not-required：T-09/T-20/T-21 覆盖外链路 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-02.md</Path>` |
| T-03 | 持久限流与稳定 Session | — | `WT/T-03` | Lead / dynamic dispatch | not-required：T-19/T-20 覆盖客户端 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-03.md</Path>` |
| T-04 | 五种数据范围 | — | `WT/T-04` | Lead / dynamic dispatch | not-required：双方言合同直接证明 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-04.md</Path>` |
| T-05 | Tombstone/文件处置 | T-02, T-04 | `WT/T-05` | Lead / dynamic dispatch | not-required：故障注入直接证明 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-05.md</Path>` |
| T-06 | 文件容量治理 | — | `WT/T-06` | Lead / dynamic dispatch | not-required：服务/磁盘故障测试 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-06.md</Path>` |
| T-07 | 日志与 Doctor | — | `WT/T-07` | Lead / dynamic dispatch | not-required：进程输出归 T-09/T-20 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-07.md</Path>` |
| T-08 | 公共 wire contract | T-02~06 | `WT/T-08` | Lead / dynamic dispatch | not-required：HTTP contract integration | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-08.md</Path>` |
| T-09 | CLI/Profile 拓扑 | T-07, T-08（closure）；DEV-09-001 early checkpoint | `WT/T-09` | Lead / dynamic dispatch | not-required：真实进程合同 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-09.md</Path>` |
| T-10 | Desktop first setup | T-02, T-09, T-13 | `WT/T-10` | Lead / dynamic dispatch | not-required：原生 E2E 归 T-20 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-10.md</Path>` |
| T-11 | Element Plus 设计系统 | — | `WT/T-11` | Lead / dynamic dispatch | not-required：真实视口归 T-19/T-20 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-11.md</Path>` |
| T-12 | Router/Shell 单一事实源 | T-11 | `WT/T-12` | Lead / dynamic dispatch | not-required：刷新/history 归 T-19/T-20 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-12.md</Path>` |
| T-13 | 跨标签 Session clients | T-03, T-08 | `WT/T-13` | Lead / dynamic dispatch | not-required：真实双标签归 T-19 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-13.md</Path>` |
| T-14 | IAM/Organization UI | T-04, T-05, T-08, T-11~13 | `WT/T-14` | Lead / dynamic dispatch | not-required：真实安全流程归 T-19 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-14.md</Path>` |
| T-15 | Operations pages | T-11, T-12 | `WT/T-15` | Lead / dynamic dispatch | not-required：真实 route/API 归 T-19 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-15.md</Path>` |
| T-16 | Tools pages | T-06, T-08, T-11, T-12 | `WT/T-16` | Lead / dynamic dispatch | not-required：真实上传/生成归 T-19 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-16.md</Path>` |
| T-17 | 当前全垂直 Generator | T-04~06, T-11, T-14, T-16 | `WT/T-17` | Lead / dynamic dispatch | not-required：隔离生成编译为直接证据 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-17.md</Path>` |
| T-18 | PG/security required CI | T-01, T-09, T-17 | `WT/T-18` | Lead / dynamic dispatch | not-required：自身是数据库/供应链 Gate | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-18.md</Path>` |
| T-19 | Web browser E2E | T-13~16, T-18 | `WT/T-19` | Lead / dynamic dispatch | required：Lead 在 parent-candidate 运行 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-19.md</Path>` |
| T-20 | macOS native E2E | T-10, T-14~16, T-19 | `WT/T-20` | Lead / dynamic dispatch | required：Lead 在 parent-candidate 运行 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-20.md</Path>` |
| T-21 | 三 Profile clean-room | T-17~20 | `WT/T-21` | Lead / dynamic dispatch | required：Lead 在 parent-candidate 运行 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-21.md</Path>` |

`WT/T-XX` 是执行期解析为 `specdev-worktree/2026-08-29-productization-and-ui-reconstruction/T-XX` 的唯一 source worktree locator；实际绝对路径、branch 与 `base_sha` 在 Dispatch 时写入 change worktree record，不在计划中绑定机器路径。

## 3. Gates and Completion Evidence

### Overall Definition of Done

- 21 个 Ticket 均为 `done` 或有用户批准且记录原因的 `cancelled`；任何取消不得留下未覆盖 AC。
- 每个非取消 Ticket 都有非空 source commit、source-worktree 非 E2E 检查、Lead 双轴审查、通过的 parent-candidate、父分支 `result_sha` 与 Evidence。
- 所有 shared owner、OpenAPI/generated、双方言 migration、旧调用点归零、不可逆边界和恢复合同闭合。
- `task test`、`task lint`、`task build TARGET=all PROFILE=server-sqlite`、`pnpm --dir go-admin-plus-ui typecheck` 以及 Ticket/Gate 指定检查在适用 checkpoint 通过。
- T-18~21 required Gate 真实通过；父分支无未集成 source checkpoint、活动 candidate、未决 critical/high deviation 或伪绿色。
- Ticket、Map、Goal Plan、Evidence、change worktree records 与 Git 可达关系一致；签名/公证仅记 `not-required`。

### Gates

| Gate | 开启条件 | 关闭证据 | 阻塞范围 | Lead/批准人 | 失败恢复 |
|---|---|---|---|---|---|
| G0 Baseline Protection | Goal Plan ready；用户授权 required/candidate | `main@904d66a`、dirty path/hash、具名可恢复 stash/快照与非删除声明；Speculo 工件可定位 | 所有 implementation | Lead；用户授权已取得 | 不创建 source worktree；从原工作树/具名快照恢复 |
| G1 Domain/UI Foundations | G0 关闭；A/B 依赖满足 | T-01~07、T-11/12 的 result；安全/迁移/组件定向检查；无路径越界 | T-05、T-08、全部后续 UI | Lead | 保留失败 source worktree；父分支不动，最多 4 次 candidate 尝试 |
| G2 Public Contract Stable | T-02~06 所需 result 完整 | T-08 OpenAPI lint、确定性 generate、Go/TS compile、HTTP 正反合同和 result | T-09、T-13、T-16 及消费者 | Lead | 回到 T-08；暂停所有消费者，不让其改 generated |
| G3 Runtime and Product UI | G2；T-07/11/12 等依赖 result | T-09/10/13/14/15/16 result；CLI/profile、route、页面、native setup 的非 E2E 集成检查 | T-17、T-19、T-20 | Lead | 只回退失败候选；合同冲突返回 T-08/ADR owner |
| G4 Generator Current | G3；当前后端/UI/权限形态稳定 | T-17 隔离生成完整 vertical slice 并通过 schema/compile/test/architecture checks | T-18、最终收缩 | Lead | 修正模板/门禁；不得给旧模板加兼容分支 |
| G5 PostgreSQL and Supply Chain | G4；T-01/T-09/T-17 result | T-18 真实 PG service、双方言并发/race、govuln/secret/SBOM 门禁；缺环境失败 | T-19~21 | Lead | 父分支不动；恢复 PG/工具环境或修正产品，禁止 skip |
| G6 Web System | G5；所有 Web 业务页面 result | T-19 在 SQLite+PG 真实后端运行浏览器、双标签、route、权限、Files/Generator 流程 | T-20、T-21 | Lead | 保存 trace/screenshot/log；修正 owning Ticket 后重建 candidate |
| G7 Native System | G6；T-10 与共享页面 result | T-20 在 macOS 原生 Tauri/sidecar/SQLite/Keychain/window 环境完成首次设置、重启和失败清理 | T-21 | Lead | 保留 native 诊断；父分支不动；不得用 Web build 替代 |
| G8 Release Candidate | G4~G7 全关闭；T-21 candidate | 三 Profile 空库 migrate/setup/login/core/restart、全量治理/构建/安全/docs checks、全部 AC Evidence、旧模式零命中 | change completion | Lead；发布/归档另需用户授权 | 不发布；清理 disposable roots，恢复备份并前向修复后重建 |

Gate 的关闭依据是行为和不可变 checkpoint，不是“完成了若干 Ticket”。Required E2E 只由 Lead 在 parent-candidate 环境判定。

### Contract and Reference Coverage

| 合同或参考要求 | 覆盖 Ticket | 验证接缝 | Evidence | 状态 |
|---|---|---|---|---|
| AC-001~007：Bootstrap/Recovery/Desktop 初始化 | T-02, T-10, T-18, T-20 | 双方言事务、原生 setup/session | T-02, T-10, T-18, T-20 | planned |
| AC-008~013：限流、Session、CSRF、撤销 | T-03, T-05, T-08, T-13, T-18, T-19 | 数据库竞争、HTTP、双标签 | T-03, T-05, T-08, T-13, T-18, T-19 | planned |
| AC-014~018：Router、Shell、成熟 UI | T-11~16, T-19, T-20 | component、history、视口、键盘/native | T-11~16, T-19, T-20 | planned |
| AC-019~022：组织与五种数据范围 | T-04, T-08, T-14, T-18 | 双方言授权集合、直接 API 越权 | T-04, T-08, T-14, T-18 | planned |
| AC-023~026：账号删除与文件处置 | T-05, T-08, T-14, T-18 | Tombstone/outbox、取消竞态、UI 确认 | T-05, T-08, T-14, T-18 | planned |
| AC-027~029：配额、磁盘与对账 | T-06~09, T-16, T-18 | 并发容量、低水位、readiness/reconcile | T-06~09, T-16, T-18 | planned |
| AC-030~033：迁移、Profile、日志、Doctor | T-07, T-09, T-10, T-18, T-20, T-21 | 进程合同、schema mismatch、三 Profile | T-07, T-09, T-10, T-18, T-20, T-21 | planned |
| AC-034~039：Generator、CI/E2E、安全、文档 | T-01, T-08, T-14~21 | 隔离生成、PG/Web/native/clean-room | T-01, T-08, T-14~21 | planned |
| 当前 ADR-001~016 与永久 ADR-0001~0021 | 全体，重点 T-02~10、T-17~21 | 架构/compatibility checks、双轴审查 | 各 Ticket + G8 汇总 | planned |
| Backplane/RuoYi/Plus UI 参考符合性 | T-07, T-09, T-11~17, T-21 | 逐项吸收运维/交互优点，反向拒绝弱安全/复制架构 | 对应 Ticket + T-21 | planned |

完整 `AC-001~039` 单项映射仍以 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>` 为权威；本表只投影跨 Ticket Gate。

## 4. Execution and Integration Protocol

### Lead Orchestration

| 项目 | 决定 | 事实依据 |
|---|---|---|
| Lead | `codex-root@2026-08-29-productization-and-ui-reconstruction/epoch-1` | 唯一 SpecDev 状态、Evidence、E2E 与父分支 owner |
| Implementation subagents | 上限 3，Lead 不计入；Wave 可进一步降并发 | config=3，平台最多 3 个非 Lead slot，required worktree 隔离 |
| Integration attempts | 每 Ticket 最多 4 次 | config 快照 `max_integration_attempts=4` |
| Read-only agents | 无 SpecDev 数字上限 | review/research/test-observation，不写项目、状态或 Evidence |
| Dispatch | execution-time dynamic | provider、模型、是否派单按 Ticket 当时事实选择，不预分配 |
| Speculo 写入 | Lead-only | subagent 不写 Ticket/Map/Plan/Evidence/change status |
| E2E | Lead-only parent-candidate | source worktree 不运行 E2E，也不能自报 required pass |

每次 implementation Dispatch Packet 必须绑定 Ticket、Goal Plan、依赖 Evidence、最新父分支 `base_sha`、source branch/worktree、writable/read-only/shared owner、允许动作、非 E2E 检查、停止条件和返回格式。原生 implementation 返回必须包含最终 source commit、dirty 状态、实际路径、命令/结果与未验证项；外部 provider 未另行授权，不得接收源码或项目数据。

### Ticket Workspace and Integration

| Ticket | Parent/base | Workspace/branch | Source checks | Implementation commit | Integration checks/E2E | Parent result |
|---|---|---|---|---|---|---|
| T-01 | dispatch 时最新 `main` | `WT/T-01` / `specdev/.../T-01` | Go、contract、govuln | required/non-empty | candidate Go+contract；E2E N/R | `result_sha` 后解锁 G5 |
| T-02 | 最新 `main` + G0 database snapshot | `WT/T-02` / `specdev/.../T-02` | IAM、双方言、race/secret | required/non-empty | candidate IAM/DB；E2E N/R | result 后解锁 T-05/T-08/T-10 |
| T-03 | dispatch 时最新 `main` | `WT/T-03` / `specdev/.../T-03` | Session 双方言/race/time | required/non-empty | candidate IAM/session；E2E N/R | result 后解锁 T-08/T-13 |
| T-04 | dispatch 时最新 `main` | `WT/T-04` / `specdev/.../T-04` | scope/organization 双方言 | required/non-empty | candidate authorization；E2E N/R | result 后解锁 T-05/T-08/T-14 |
| T-05 | T-02+T-04 result 的最新 `main` | `WT/T-05` / `specdev/.../T-05` | lifecycle/outbox/fault | required/non-empty | candidate destructive-state tests；E2E N/R | result 后解锁 T-08/T-14/T-17 |
| T-06 | dispatch 时最新 `main` | `WT/T-06` / `specdev/.../T-06` | quota/reconcile 双方言 | required/non-empty | candidate Files/fault；E2E N/R | result 后解锁 T-08/T-16/T-17 |
| T-07 | dispatch 时最新 `main` | `WT/T-07` / `specdev/.../T-07` | logging/Doctor/config | required/non-empty | candidate process integration；E2E N/R | result 后解锁 T-09 |
| T-08 | T-02~06 所需 result 的最新 `main` | `WT/T-08` / `specdev/.../T-08` | OpenAPI/generate/Go+TS | required/non-empty | candidate contract/conformance；E2E N/R | result 后关闭 G2 |
| T-09 | T-07+T-08 result | `WT/T-09` / `specdev/.../T-09` | CLI/process/profile/schema | required/non-empty | candidate runtime/Task；E2E N/R | result 后解锁 T-10/T-18 |
| T-10 | T-02+T-09+T-13 result + G0 native snapshot | `WT/T-10` / `specdev/.../T-10` | Go/Rust/native integration | required/non-empty | candidate Rust/build；E2E N/R | result 后解锁 T-20 |
| T-11 | dispatch 时最新 `main` | `WT/T-11` / `specdev/.../T-11` | pnpm/UI component/visual contract | required/non-empty | candidate lint/type/build；E2E N/R | result 后解锁 T-12/14/15/16 |
| T-12 | T-11 result | `WT/T-12` / `specdev/.../T-12` | router/Shell component | required/non-empty | candidate dual-host build；E2E N/R | result 后解锁 UI consumers |
| T-13 | T-03+T-08 result | `WT/T-13` / `specdev/.../T-13` | adapter/domain deterministic | required/non-empty | candidate client integration；E2E N/R | result 后解锁 T-14/T-19 |
| T-14 | 所列安全/UI依赖 result | `WT/T-14` / `specdev/.../T-14` | domain/component/security | required/non-empty | candidate UI/type/build；E2E N/R | result 后解锁 T-17/T-19/T-20 |
| T-15 | T-11+T-12 result | `WT/T-15` / `specdev/.../T-15` | three Web Domains | required/non-empty | candidate UI/type/build；E2E N/R | result 后解锁 T-19/T-20 |
| T-16 | 所列 Files/UI依赖 result | `WT/T-16` / `specdev/.../T-16` | tools domain/component | required/non-empty | candidate UI/type/build；E2E N/R | result 后解锁 T-17/T-19/T-20 |
| T-17 | 所列当前架构 result | `WT/T-17` / `specdev/.../T-17` | isolated generate/compile/test | required/non-empty | candidate architecture/generator；E2E N/R | result 后关闭 G4 |
| T-18 | T-01+T-09+T-17 result | `WT/T-18` / `specdev/.../T-18` | CI scripts/negative policy | required/non-empty | candidate real PG/race/security/SBOM；E2E N/R | result 后关闭 G5 |
| T-19 | T-13~16+T-18 result | `WT/T-19` / `specdev/.../T-19` | runner/unit/type checks only | required/non-empty | candidate required SQLite+PG browser E2E | result 后关闭 G6 |
| T-20 | T-10+T-14~16+T-19 result | `WT/T-20` / `specdev/.../T-20` | runner/Rust/script checks + DEV-20-002 theme composition + DEV-20-003/004 Web verification | required/non-empty | candidate required macOS native E2E | result 后关闭 G7 |
| T-21 | T-17~20 result | `WT/T-21` / `specdev/.../T-21` | docs/release policy checks | required/non-empty | candidate required full regression + three-profile clean-room | final `result_sha` 后评估 completion |

对每个 Ticket，Lead 从最新依赖已集成的 `main` 创建唯一 source worktree；source worktree 不运行 E2E。Lead 接收 source commit 后冻结 `parent_before_sha`，在 Lead-owned detached candidate checkout 合并并运行 integration checks/适用 E2E；父 HEAD 漂移则 candidate 标记 stale 并重建。验证通过且 HEAD 未漂移后，Lead 才 fast-forward `main`、重读 tree 并记录 `result_sha`。候选失败时 `main` 不动，Ticket 回到同一 source worktree修正；成功集成不自动清理 branch/worktree。

### Authorization Matrix

| 动作 | 状态 | 目标与条件 |
|---|---|---|
| Current workspace Ticket changes | not-authorized | 已选择 required；当前 workspace 只由 Lead 写 Speculo 工件和执行父分支集成 |
| Ticket worktree local changes | allowed | 每 Ticket 唯一 source worktree，严格限制 writable/shared owner 路径 |
| Implementation commit | allowed | 用户于 2026-08-31 接受推荐；每 Ticket 必须形成非空本地 source commit |
| Local direct-parent verification and parent update | not-authorized | 本计划不使用 current/direct-parent |
| Local candidate integration and parent update | allowed | Lead-only；候选通过、父 HEAD 未漂移后可本地 fast-forward `main` |
| G0 reversible dirty-state snapshot/stash | allowed | 仅四组已知产品路径；记录 hash/locator且在最终核对前不得 drop |
| Push / PR / remote merge | exception-only | 仅 DEV-20-005/006 的同一具名 workflow-only probe branch 可 push；PR、remote merge 与其他远程写入仍未授权 |
| Branch/worktree/stash cleanup | not-authorized | 集成成功只改变生命周期状态；清理另行授权 |
| Local disposable migration/E2E data | allowed | 仅明确临时根/测试数据库；禁止破坏 `<Path>dev_store/**</Path>` |
| Deploy / production migration / publish / archive | not-authorized | 每个外部或不可逆动作另行取得用户批准 |

### Evidence Return

Subagent 只返回候选事实与 source commit，不写 Evidence。Lead 核对实际 diff、pathspec、commit 可达、workspace dirty 状态和非 E2E 结果，随后写 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-XX.md</Path>`。Required E2E、candidate/result SHA、父分支包含关系、偏差与最终验收只由 Lead 记录。

## 5. Constraints, Risk and Recovery

### Non-negotiable Constraints

- 保留模块化单体、消费者 Port/Integration Event、Bun/Goose、OpenAPI 3.1、pnpm monorepo、无头 domain/Web Domain、Tauri 2 Go sidecar 和根 Task 命令面。
- Server 仅 SQLite/PostgreSQL，Desktop 仅 SQLite；租户、Redis、JWT、refresh token、Casbin 和默认管理员保持零保留。
- 无固定 password SQL、未认证 Bootstrap/Recovery HTTP、argv secret、WebView secret、日志/审计 secret。
- 后端是 capability/data scope/删除不变量事实源；前端可见性不能替代授权。
- OpenAPI/generated 仅 T-08 写；`main.rs` 原则上仅 T-10 写，DEV-20-006 只向 T-20 开放 `native-e2e` feature 的 Context data-store 注入；lock/UI、App Shell、CI、Web/Desktop E2E、docs/release 各自由指定 shared owner 写；DEV-20-002 仅向 T-20 开放 test-only native App、ProductWorkspace 主题组合、App Shell manifest 与精确 lock importer，DEV-20-003/004 仅开放 Web Shell 的 lazy-route DOM 等待和主题 reload 验证。
- Migration forward-only；PostgreSQL 生产显式 migrate，Server SQLite 自动 migrate，Desktop 备份成功后自动 migrate。
- 不可逆 purge 只有 worker claim 前可取消；最后系统管理员不可删除；recover-admin 不创建或复活账号。
- required Gate 不允许 skip/allow-failure/静默 retry；签名和公证只记 `not-required`。

### Verification Integrity

执行基线是 `main@904d66a5afaff58037798d4d17b96bc69c1988f5`，每次 Dispatch 再冻结最新父 `base_sha`。判卷接缝来自 Ticket 验证矩阵、根 Taskfile、OpenAPI generator、Go/TS/Rust 工具链和 T-18~21 的真实环境。

禁止通过删除校验、扩大 allowlist、mock required 数据库/浏览器/native、缺环境返回 0、只看截图、把 source-worktree E2E 当通过、把外部 provider 自报当 Evidence 或把未运行项写成 passed 制造绿色。Source-worktree 只做非 E2E；parent-candidate 由 Lead 做集成与适用 E2E。命令、cwd/profile、环境类别、退出状态、关键输出和 checkpoint 必须进入 Evidence。

G0 先对以下用户既有变更记录 path、mode、byte hash 与可恢复 locator，再移入具名且不自动 drop 的 stash/快照，使 `main` 的产品路径可安全集成：`<Path>database/README.md</Path>`、`<Path>database/bootstrap/**</Path>`、`<Path>go-admin-plus/test/database/**</Path>`、`<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>`。前三组只导入 T-02，最后一项只导入 T-10；原快照在两个 result 逐字节/语义核对完成前不得清理。Speculo change 工件和无关用户修改不得进入该快照或被重置。

### Migration or Release Sequence

1. **Expand domain state：** T-02 的 IAM `0030`、T-03 的 `0040`、T-04 的 `0050`、T-05 的 `0060` 与 T-06 的 Files `0020` 按模块所有权分别引入；SQLite/PostgreSQL 同版本语义，旧运行入口尚不消费新 wire。
2. **Migrate and observe：** 各 source/candidate 在 disposable 双方言数据库验证前进迁移、失败原子性、并发和重启；账号/Session/文件不可逆状态机需故障注入。
3. **Publish contract：** T-08 原子更新 OpenAPI、全部生成物和 HTTP adapter；消费者暂停到 G2 关闭。
4. **Switch runtime/clients：** T-09/10/13~16 切换 CLI、Profile、Desktop、Router/Session client 和页面，不保留旧命令/路由/CSRF 模式。
5. **Contract old paths：** T-17 将 Generator 固定到当前架构；T-18 让新门禁 required；T-21 最后删除旧文档、部署描述和固定初始化材料。
6. **Release evidence：** T-19 Web、T-20 native、T-21 三 Profile clean-room 依次完成。公开发布、部署和生产迁移不在本授权内。

### Risks, Monitoring and Recovery

| 风险 | 监控/反向验证 | 恢复 |
|---|---|---|
| 既有 dirty 资产丢失或混入错误 Ticket | G0 hash、path inventory、stash locator、T-02/T-10 diff | 停止所有集成，从非丢弃快照恢复；不 reset/clean |
| 双方言迁移或并发语义漂移 | migration checksum、PG/SQLite contract、race/竞争测试 | 父分支不动；保留 DB/日志，前向修正 migration/use case |
| OpenAPI/generated 多头写入 | generate drift、path owner、consumer pause | 退回 T-08，拒绝消费者手改 generated |
| Session/授权安全回归 | GET 零写、双桶、双标签、直接 API 越权 | 撤销候选；返回 T-03/T-04/T-08 owner |
| purge/账号删除造成不可恢复损失 | 状态机、outbox idempotency、claim 边界、disposable storage | 只在临时数据验证；失败保持 pending/failed 并前向恢复 |
| UI 看似完成但 route/视口不可用 | component + T-19/T-20 real viewport/history | trace 定位 owning Ticket，修正后重建下游候选 |
| candidate 冲突或父 HEAD 漂移 | `parent_before_sha`、重读 HEAD/tree | attempt 记 failed/stale；main 不动，从最新 result 重建，最多 6 次 |
| required 环境缺失或 flaky | 环境探针、退出码、首次失败 artifacts | Gate blocked；修复环境/确定根因，不 skip 或静默 retry |
| 磁盘/测试数据污染用户数据 | disposable root guard、磁盘水位、cleanup manifest | 停止测试，仅清理明确 disposable roots；不触碰 `dev_store` |

### Deviation Control

遵循 `<Path>{roots.workflows}/specdev/common/rules/deviation-control.md</Path>`。路径越界、公共合同变化、迁移编号/语义冲突、安全不变量变化、required E2E 无法运行、超过 6 次候选尝试或需要新外部授权时，Lead 停止受影响 Wave，记录事实与最后可信 checkpoint，并返回 Ticket/Spec/ADR owner；不得由实现者自行扩大范围。低影响实现细节可在 Ticket 锁定边界内决定并写 Evidence。T-19 前五个 candidate 各在首个真实红灯处停止；用户 `USER-DECISION:all-approved` 明确批准将 ticket-local 上限从 4 扩到 6，全部失败 checkpoint 均保留且未重标为通过。

## 6. Progress and Decisions

### Current Status

| 项目 | 当前事实 |
|---|---|
| Plan | `ready`；required worktree + candidate-merge；Lead epoch 1 |
| Parent | 当前 `main@d935d4f` 仍未包含 T-20 产品树；最新 candidate 包含此前治理父节点；`4870670` 误合并由 `8e8855b` 前向撤销并保留审计历史，未声称晋升 |
| Tickets | T-01~T-19 均 `done`；T-20 `in_progress`；T-21 `ready` |
| Gate | G0~G6 已通过；G7/T-20 执行中；G8 尚未开启 |
| Workspace records | T-20 latest source `7c2e7c3`、portable candidate `1e17579`/tree `90cc12f` 已记录并通过 portable checks；probe `b3244c0` 的 DEV-20-018 attempt 2 `login-demo-workspace` 红灯保留，既有 source/candidate、probe、扫描 artifacts 与旧失败候选继续保留 |
| Authorization | local source commits、candidate integration、`main` fast-forward、required runner 本地准备与既有 DEV、DEV-19-001~010 已授权；DEV-20-005/006/008/011/014 各 3 次 probe 均已用尽；DEV-20-007 已用 2 次且剩余 runner-only attempt 不用于产品修正；DEV-20-017 只开放 exact enabled button AXScrollToVisible + 既有 center click 与最多 3 次逐项归因 attempt；其他远程写入/部署/发布/生产迁移/清理未授权 |
| Known dirty state | G0 已将 database 初始化输入固定到 `ee1d7f7`，Desktop 输入固定到 stash object `39480546c2a2e2ff386a176f4278c6183a0e868c`，continuation 输入固定到 stash object `f593f53b2850063f415c9cd521ab6aaa8a99c510`；均未清理 |
| Validation baseline | tickets validator `0 error / 0 warning`；Taskfile 的 test/typecheck/lint/build/contract/generate 入口存在 |

### Pending Decisions and Blockers

当前执行 active；DEV-09-001 已解除 Wave A/T-09 的执行闭环并完成本轮候选晋升：

- WinLibs GCC 16.1 与独立 PostgreSQL 17.11 disposable cluster 已建立，required race 和逐 Ticket PostgreSQL 检查均有通过证据；该 cluster 仅使用显式隔离数据库，未触碰用户数据库或 `dev_store`。
- G7/T-20 的 hosted macOS 15.7.7 arm64 通道已实测 `System Events` UI elements enabled=`true`。DEV-20-014 attempt 3 再次返回 `login-demo-workspace`，且 center click command 未报错。静态产品壳证明 order 600 的 Demo 位于 576px overflow-auto sidebar 的屏外位置而 AX tree 仍可见；DEV-20-017 只在同一 exact enabled button 上先执行 AXScrollToVisible 再 center click，不改变产品、fixture、调用点或等待上限。
- 既有 Windows sidecar、desktop、Generator、UNC、backup、Files `% literal.txt` 与符号链接失败仍归对应 owning Ticket 修复；T-08 已收敛旧 rotate-on-read、product migration count 与 Audit adapter 红灯。

T-20 candidate `be1f74a`（tree `8056abb`，source `1c285f8`）已实现 DEV-20-023，并通过 Vitest 256/256、Node 48/48、Desktop runner 21/21、lint、production build/asset scan、diff-check 与 clean tree；source 完整 typecheck 和 native-e2e build 通过，两项主题测试标记存在，production 重建再次通过零测试字节扫描。前一 attempt 3 probe `0d30fd4` / Actions `33547554824` / job `99988721801` 从 19:13:18Z 至 19:21:48Z 越过 login navigation、Demo page 与 boundary，首个红灯推进到 `theme-dark-toggle`，直接证明 DEV-20-022 有效。DEV-20-023 只新增 production scanner 强制排除的固定测试按钮，由其 exact DOM-click 真实产品主题按钮并保留 aria-label/storage/restart 后置条件，尚余 3 次逐项归因配额。G7 必须在真实 macOS execution 得到 `DESKTOP_NATIVE_E2E_PASS runtime=tauri-native profile=sqlite skipped=0`，该 marker 出现前 candidate 不晋升。所有既有 source/candidate、probe、未跟踪扫描 artifacts 与保护 stash均保留。

### Resume Protocol

恢复时依次读取本 Goal Plan、当前 Ticket、`<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/.status.json</Path>` 的 worktree record 和最新 Evidence；核对实际 `main` HEAD、G0 快照 locator、活动 source/candidate 及其 Git 可达性。从最后已通过的父分支 `result_sha` 或同一 Ticket 最后待修正 `source_checkpoint` 继续，不重新决定 workspace、shared owner、授权或已关闭 Gate。Lead 会话变更时提升 leadership epoch，旧 Lead 不再写状态。

## Assumptions

- 执行平台在 required Gate 到达前仍可提供本机 macOS、SQLite、可启动的 PostgreSQL service、真实浏览器和 Tauri 工具链；若实测不成立，相关 Gate 明确 blocked，不降级验收。
- 逻辑 locator `WT/T-XX` 的实际目录由执行时安全分配；路径选择不改变 branch、owner 或 Evidence 合同。
- 现有 dirty 产品资产是 T-02/T-10 的候选输入而非已验收实现；最终保留、改写或删除取决于对应 Spec/Ticket，并由 G0 快照保证可恢复。
