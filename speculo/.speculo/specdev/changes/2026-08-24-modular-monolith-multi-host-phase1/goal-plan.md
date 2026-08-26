---
schema_version: 6
artifact: goal-plan
change: 2026-08-24-modular-monolith-multi-host-phase1
status: completed
modes: [migration, high-assurance, reference-conformance, release-coordination]
orchestration: lead-directed
lead: root
implementation_agent_limit: 1
integration_attempt_limit: 3
ticket_workspace_policy: current
integration_gate: direct-parent
ready_for_execution: false
---

# Goal Plan: Phase 1 模块化单体与多 Host 管理端

- **Goal Plan：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/goal-plan.md</Path>`
- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **Tickets Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **Ticket 目录：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/</Path>`
- **Evidence 目录：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/</Path>`

## 1. Outcome and Authority

### Outcome

交付一个 Host 无关的 Go 模块化单体 Application，由 ServerHost 和 Wails DesktopHost 适配；交付一个 workspace 化 Admin 前端，由 app-core、runtime、api-client、ui 与业务 Domain 组合，并在 Web HTTP 与 Desktop loopback 两种 Runtime 下消费同一 OpenAPI 合同。最终形成 Linux AMD64 Web/API Compose、macOS ARM64 离线 DMG 和 Windows AMD64 离线 NSIS 三类同版本产物。

### Success and False Completion

成功必须同时满足：13 个 Ticket 按 DAG 和 Gate 完成；`AC-001` 至 `AC-018` 均有 Lead 核验的 Evidence；Server/Desktop 使用同一 Application 和业务 API；Admin 只维护一份业务客户端与 Domain；迁移、文件、loopback、单实例和离线行为通过目标环境验证；三类产物具有同版本 manifest、校验和与可复现来源。

以下均属于 false completion：只移动目录但没有 characterization；新增 Desktop 壳却另写一套业务 API；只在开发机启动但未验证断网、升级和失败恢复；桌面后端绑定非 loopback 或缺少启动 token/Origin 校验；用 skip、删除断言或真实云凭据掩盖默认测试失败；生成安装包但未验证 checksum、trust state、自授权安装和原生离线核心流程；把 self-use 包伪称平台可信签名；生成客户端与 OpenAPI 漂移；保留旧请求入口或路径兼容层却宣称收缩完成。

### Non-goals

- Phase 1 不新增 Web 用户端、Mobile 或第二个管理端，不拆微服务，不更换现有业务语义。
- 不在本计划内执行生产部署、生产数据库迁移、远程 PR/merge、正式发布或凭据管理；用户最新明确要求的当前 reviewed commits push 是一次显式例外。
- 不把 PostgreSQL 与 SQLite 强行统一为相同实现；只统一 Application 端口与可验证合同。
- 不承诺 x86 macOS、Windows ARM64、Linux 桌面或移动端产物。

### Authoritative Inputs

| 优先级 | 来源 | 负责内容 | 冲突处理 |
|---|---|---|---|
| 1 | 用户最新明确决定 | 产品取舍、current workspace、逐 Ticket 本地提交与批准 | 更新真正拥有该决策的工件 |
| 2 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ADR.md</Path>` 与 `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/CONTEXT.md</Path>` | 当前 change 架构决定与领域语义；文件不存在时不构成输入 | 返回 `<Path>{roots.workflows}/specdev/G-grill-with-docs/G-grill-with-docs.md</Path>` 更新真正 owner |
| 3 | `<Path>{roots.state}/specdev/adr/</Path>` 与 `<Path>{roots.state}/specdev/context/</Path>` | 已毕业的永久决定与领域知识 | 当前 change 替代时在 change ADR/LOG 明示 |
| 4 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>` | 外部行为、范围、故事与 `AC-001..018` | 下游不得改写，行为变化返回 Spec |
| 5 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/</Path>` | 单 Ticket 路径、验证、依赖与完成合同 | Goal Plan 只编排；局部缺口同步 Ticket/Map |
| 6 | 当前代码、基线命令与运行事实 | 可行性、既有失败和实现约束 | 冲突时触发 deviation control 并返回真正 owner |

## 2. Execution Graph

### DAG and Critical Path

```text
T-01 baseline
  ├── T-02 Application ── T-03 adapters ── T-04 ServerHost ── T-06 contract/runtime ─┐
  └── T-05 workspace ────────────────────────────────────────────────────────────────┤
                                                                                     ├── T-07 domains ─┐
                                               T-03 + T-06 ── T-08 desktop tracer ───┴── T-09 hardening ─┬── T-11 macOS ─┐
                                               T-04 + T-07 ── T-10 Linux Compose ──────────────────────────┤              ├── T-13 release/contract
                                                                                                           └── T-12 Windows ┘
```

DAG 的结构关键路径经 T-01、Application/adapter、Server/OpenAPI、Domain/Desktop、桌面加固、平台发布到 T-13。由于用户选择 current workspace，调度关键链被收紧为 `T-01 -> T-02 -> T-05 -> T-03 -> T-04 -> T-06 -> T-07 -> T-08 -> T-09 -> T-10 -> T-11 -> T-12 -> T-13`；不得利用图中的独立分支并行写入。

### Waves and Ownership

| Wave | Ticket | 前置条件 | 项目写路径摘要 | Shared owner | Gate/集成序号 |
|---|---|---|---|---|---|
| W0 | T-01 | 根/子仓库基线已记录且无用户脏改动 | 后端 characterization/file_store tests；前端 E2E/fixtures | 无 | G0 / 01 |
| W1 | T-02 | G0 | `internal/application`、`internal/modules`、module 装配、`cmd/api` | T-02: `cmd/api` | G1 / 02 |
| W1 | T-05 | T-01 result 已包含于父分支 | `apps/admin`、workspace/config、根前端 manifests | T-05: 前端 manifests | G1 / 03 |
| W2 | T-03 | T-02 result | platform/profile/tenant/jobs/migrate | T-03: `cmd/migrate` | G2 / 04 |
| W3 | T-04 | T-02、T-03 results | ServerHost、health、`cmd/go-admin`、Makefile | T-04: Makefile | G3 / 05 |
| W4 | T-06 | T-04、T-05 results | OpenAPI、contracts/runtime/api-client | T-06: OpenAPI/contracts | G4 / 06 |
| W5 | T-07 | T-05、T-06 results | app-core/ui/domains/Admin src/menu migration | T-07: migration version | G5 / 07 |
| W5 | T-08 | T-03、T-06 results | DesktopHost、desktop runtime、Go manifests | T-08: go.mod/go.sum | G5 / 08 |
| W6 | T-09 | T-07、T-08 results | desktop host/platform/tests | T-09: desktop host | G6 / 09 |
| W6 | T-10 | T-04、T-07 results | Compose/Linux release/Docker/Nginx | T-10: Compose/Linux release | G6 / 10 |
| W7 | T-11 | T-09 result；macOS ARM64 runner 可用 | macOS self-use release/build/workflow | T-11: macOS release | G7 / 11 |
| W7 | T-12 | T-09 result；Windows AMD64 runner 可用 | Windows self-use release/build/workflow | T-12: Windows release | G7 / 12 |
| W8 | T-13 | T-10、T-11、T-12 results | product workflow/manifest/兼容层收缩 | T-13: root release orchestration | G8 / 13 |

Wave 是能力阶段，不是并行许可。Shared owner 只在对应 Ticket 活跃时写该路径；后续 Ticket 对前序 shared path 的写入必须已经在自身 frontmatter 明确授权。

### Ticket Quick Reference

| ID | 可观察产出 | Dependencies | Workspace | Implementation owner | E2E disposition | Evidence |
|---|---|---|---|---|---|---|
| T-01 | 可重复 API/UI/数据/启动基线，默认测试不依赖真实云端 | — | current | Lead / dynamic dispatch | required：定义全 change 基线 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-01.md</Path>` |
| T-02 | Host 无关 Application 生命周期内核 | T-01 | current | Lead / dynamic dispatch | required：启动/关闭与双 Host 接缝 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-02.md</Path>` |
| T-05 | workspace 化 Admin 等价构建与导航 | T-01 | current | Lead / dynamic dispatch | required：登录/菜单/CRUD 等价 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-05.md</Path>` |
| T-03 | Server/Desktop profile、迁移与文件合同 | T-02 | current | Lead / dynamic dispatch | required：双数据库与失败恢复 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-03.md</Path>` |
| T-04 | ServerHost 与 health/capabilities | T-02,T-03 | current | Lead / dynamic dispatch | required：现有 API 与健康语义 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-04.md</Path>` |
| T-06 | OpenAPI、Runtime 与纯 ApiClient | T-04,T-05 | current | Lead / dynamic dispatch | required：HTTP/desktop 双 runtime | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-06.md</Path>` |
| T-07 | app-core、Domain 与稳定 routeKey | T-05,T-06 | current | Lead / dynamic dispatch | required：Web Admin 全链路 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-07.md</Path>` |
| T-08 | Wails DesktopHost 垂直曳光弹 | T-03,T-06 | current | Lead / dynamic dispatch | required：离线登录和 CRUD | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-08.md</Path>` |
| T-09 | 桌面耐久、生命周期与 loopback 安全 | T-07,T-08 | current | Lead / dynamic dispatch | required：升级/单实例/攻击面 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-09.md</Path>` |
| T-10 | Linux Web/API 生产 Compose | T-04,T-07 | current | Lead / dynamic dispatch | required：空卷、升级、回滚 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-10.md</Path>` |
| T-11 | 可校验、自授权、离线 macOS ARM64 DMG | T-09 | current | Lead / dynamic dispatch | required：原生 self-use 安装 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-11.md</Path>` |
| T-12 | 可校验、内含 WebView2 的 Windows x64 self-use NSIS | T-09 | current | Lead / dynamic dispatch | required：原生 self-use 安装 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-12.md</Path>` |
| T-13 | 三类同版本产物与旧兼容合同收缩 | T-10,T-11,T-12 | current | Lead / dynamic dispatch | required：产品级矩阵 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-13.md</Path>` |

## 3. Gates and Completion Evidence

### Overall Definition of Done

- 每个 Ticket 的依赖 result SHA 已包含在当前根分支，修改未超出 writable/shared ownership，且形成非空本地 checkpoint。
- 每个 Ticket 的定向检查、Speculo 配置中的 test/typecheck/lint/build 适用部分，以及标记 required 的 E2E 均由 Lead 核验并写入 Evidence。
- 根仓库 Evidence 同时记录 root parent-before/result SHA、涉及的子仓库 before/result SHA、命令、环境、产物摘要和未验证项。
- `AC-001..018` 无 uncovered、pending 或以 warning 代替验收的条目；Ticket、Map、Goal Plan、Evidence 与 change status 一致。
- G8 完成前，旧兼容入口只有在零调用证据后删除；三类 release manifest 同版本且校验和可追溯。

### Gates

| Gate | 开启条件 | 关闭证据 | 阻塞范围 | Lead/批准人 | 失败恢复 |
|---|---|---|---|---|---|
| G0 基线可信 | 规划工件已提交；root/backend/frontend baseline 已记录 | T-01 characterization、对象存储 hermetic tests、前端 E2E；默认 `go test ./...` 不再依赖真实云端 | 全部实现 | Lead | 仅修复 T-01 测试接缝；不得进入结构迁移 |
| G1 可逆 prefactor | G0；T-02/T-05 依次执行 | Application lifecycle tests、Admin workspace 等价 build/test/E2E、旧行为无漂移 | T-03 以后 | Lead | 回到对应 result 前一提交，保留 characterization |
| G2 双 profile 合同 | T-02 result | PostgreSQL/SQLite migration、FileStore、失败恢复和数据完整性证据 | T-04/T-08 以后 | Lead | fail closed，恢复 fixture/备份，修适配器而非绕开端口 |
| G3 ServerHost 兼容 | G2 | 现有 `/api/v1` characterization、live/ready/capabilities、优雅停止 | T-06/T-10 | Lead | 恢复旧启动入口作为 expand 兼容层，修新 Host |
| G4 单一 API 合同 | G3 + T-05 result | OpenAPI 生成零 diff、type-check、HTTP/Desktop Runtime 合同测试 | T-07/T-08 | Lead | 保留旧 request adapter，禁止复制第二套 client |
| G5 Web/Desktop 曳光弹 | G4 | routeKey、Web Admin、桌面离线登录/CRUD 的 required E2E | T-09/T-10 | Lead | 保留 expand adapter，按 Web 或 Desktop 接缝定向修复 |
| G6 可恢复交付 | G5 | 桌面安全/升级/单实例；Compose 空卷/升级/最小权限/health | T-11/T-12/T-13 | Lead | Desktop 用备份前向恢复；Compose 回滚镜像并恢复卷备份 |
| G7 原生 self-use 发布 | G6 | macOS ad-hoc 完整性+checksum+单 App 授权+离线核心流程；Windows checksum+WebView2+SmartScreen self-use+离线核心流程 | T-13 | Lead；用户已于 2026-08-25 批准 unsigned self-use 合同 | trust state、checksum、安装或 E2E 失败则隔离产物；不通过全局关闭系统安全能力制造绿色 |
| G8 产品汇合 | G7 + T-10 result | 同版本 manifest/checksum/SBOM 来源、三平台 E2E、零旧调用扫描、全量验证 | change completion | Lead + 用户对外部 publish/deploy 的逐项批准 | 兼容层未满足收缩条件则保留并记录，不发布不完整矩阵 |

### Contract and Reference Coverage

| 合同或参考要求 | 覆盖 Ticket | 验证接缝 | Evidence | 状态 |
|---|---|---|---|---|
| AC-001..002 Application/Host 与 API 兼容 | T-01,T-02,T-04,T-07,T-13 | lifecycle + HTTP characterization | 各 Ticket Evidence | covered |
| AC-003..006 Admin、Runtime、OpenAPI、routeKey | T-01,T-05,T-06,T-07,T-13 | build/typecheck/generation/E2E/零调用扫描 | 各 Ticket Evidence | covered |
| AC-007..009 macOS、Windows、loopback 安全 | T-08,T-09,T-11,T-12 | 原生离线 E2E + socket/Origin/token tests | 各 Ticket Evidence | covered |
| AC-010..011 双方言迁移与 lifecycle | T-02,T-03,T-09,T-11,T-12 | fixture upgrade/recovery/single-instance | 各 Ticket Evidence | covered |
| AC-012..015 Compose、health、最小权限、FileStore | T-03,T-04,T-09,T-10 | Compose config/空卷/health/contract | 各 Ticket Evidence | covered |
| AC-016..017 三类产物与断网核心流程 | T-03,T-08,T-09,T-10,T-11,T-12,T-13 | release verification + network isolation | 各 Ticket Evidence | covered |
| AC-018 单一前端基础合同与可证收缩 | T-05,T-06,T-07,T-13 | import/call scan + full regression | 各 Ticket Evidence | covered |

## 4. Execution and Integration Protocol

### Lead Orchestration

| 项目 | 决定 | 事实依据 |
|---|---|---|
| Lead | `root` | change status 中唯一 SpecDev 状态、Evidence 与父分支 owner |
| Implementation subagents | `1`，Lead 不计入 | current workspace 单 writer；低于 config 最大值 3 |
| Integration attempts | `3` | `<Path>{roots.state}/specdev/config.json</Path>` 快照 |
| Read-only agents | 无 SpecDev 数字上限 | 仅 review/research/test observation，不写工作区或状态 |
| Dispatch | execution-time dynamic | 每次只在前一 Ticket Gate 关闭后派单，不预绑定 provider/模型 |

### Ticket Workspace and Integration

| Ticket | Parent/base | Workspace/branch | Source checks | Implementation commit | Integration checks/E2E | Parent result |
|---|---|---|---|---|---|---|
| T-01 | root `1126a76`; backend `750537c`; frontend `1d1af04` | current / `main` | 定向 Go/Playwright + path check | 子仓库 commit(s) 后 1 个 root checkpoint | Lead 跑 G0 全量适用检查和基线 E2E | root checkpoint = `result_sha` |
| T-02 | T-01 result | current / `main` | lifecycle/unit + path check | 同上 | G1 后端 characterization/lifecycle | root checkpoint |
| T-05 | T-02 result | current / `main` | workspace unit/type/build + path check | 同上 | G1 Admin 等价 E2E | root checkpoint |
| T-03 | T-05 result | current / `main` | adapter/migration contract + path check | 同上 | G2 双 profile/恢复 E2E | root checkpoint |
| T-04 | T-03 result | current / `main` | ServerHost/health + path check | 同上 | G3 HTTP/stop E2E | root checkpoint |
| T-06 | T-04 result | current / `main` | generation/type/Runtime contracts | 同上 | G4 双 Runtime E2E | root checkpoint |
| T-07 | T-06 result | current / `main` | Domain/route/import checks | 同上 | G5 Web Admin E2E | root checkpoint |
| T-08 | T-07 result | current / `main` | desktop build/runtime contracts | 同上 | G5 原生桌面 tracer E2E | root checkpoint |
| T-09 | T-08 result | current / `main` | security/upgrade/lifecycle tests | 同上 | G6 桌面故障与攻击面 E2E | root checkpoint |
| T-10 | T-09 result | current / `main` | Compose config/image/health checks | 同上 | G6 Linux 空卷/升级 E2E | root checkpoint |
| T-11 | T-10 result | current / `main` | macOS native build/verify | 同上 | G7 checksum/ad-hoc/self-use/断网安装 E2E | root checkpoint |
| T-12 | T-11 result | current / `main` | Windows native build/verify | 同上 | G7 checksum/WebView2/SmartScreen self-use/断网安装 E2E | root checkpoint |
| T-13 | T-12 result | current / `main` | manifest/零调用/full suite | 同上 | G8 三产物矩阵 E2E | root checkpoint |

current 模式下不得创建 source/candidate worktree。每次只允许一个 implementation owner 写当前 workspace；owner 完成非 E2E 检查后，先在每个被修改的子仓库形成可审计 commit，再在根仓库形成恰好一个 Ticket checkpoint，记录子仓库 gitlink 与根文件。Lead 独立核对 path diff、子仓库 SHA、根 parent-before 和 required E2E；通过后根 checkpoint 成为该 Ticket 的 `result_sha`。由于 checkpoint 不能自记录自身 SHA，允许随后形成一个只包含 SpecDev Evidence/状态投影的 finalization commit；父 HEAD 包含 result 且 workspace clean 后才能开始下一 Ticket。纯根仓库 Ticket 不需要子仓库 commit；跨两个子仓库的 Ticket 必须先分别提交子仓库，禁止以脏 gitlink 进入下一 Ticket。

### Authorization Matrix

| 动作 | 状态 | 目标与条件 |
|---|---|---|
| Current workspace Ticket changes | allowed | 仅本 change 的 Ticket writable/shared owner 合同；严格串行、单一 writer |
| Ticket worktree local changes | not-authorized | current 策略禁止 source/candidate worktree |
| Implementation commit | allowed | 每 Ticket 必需；子仓库 commit(s) + 1 个根 checkpoint；不 push |
| Local direct-parent verification and parent update | allowed | Lead 核对 root checkpoint 后在本地 `main` 继续 |
| Local candidate integration and parent update | not-authorized | current 策略不适用 |
| Push | allowed | 用户 2026-08-25 明确授权当前 Goal Plan 的所有 repository push；仅限本计划经审计提交 |
| PR / remote merge | not-authorized | 本次 push 授权不扩展到 PR 或远程 merge |
| Branch/worktree cleanup | not-authorized | 本计划不创建 worktree；其他清理不自动获权 |
| Deploy / production migration | not-authorized | 需逐环境、逐动作明确授权 |
| Signing/notarization/external publish | signing/notarization not-required；publish not-authorized | Phase 1 desktop 为 unsigned self-use；Developer ID/AuthentiCode/notary 不执行，GitHub Release 或其他外部分发仍需单独批准 |

### Evidence Return

Implementation owner 只返回候选事实：变更摘要、命令与退出码、未验证项、子仓库 SHA 和根 checkpoint。Lead 必须自行检查 `git diff`/包含关系、重跑 Gate 命令并写 Evidence、Ticket/Map/status；owner 的口头结论不能替代 Evidence。

## 5. Constraints, Risk and Recovery

### Non-negotiable Constraints

- 一个 Application，多 Host adapter；Domain 不导入具体数据库、HTTP server、Wails 或全局配置实现。
- `/api/v1`、授权和响应 envelope 在 expand-migrate-contract 窗口保持兼容；OpenAPI 是前端生成合同权威。
- Desktop 仅绑定 loopback 随机端口，并同时实现启动 token、严格 Origin、单实例与受限文件权限。
- SQLite/PostgreSQL migration 只追加且可恢复；不修改已经发布的迁移，不在启动失败后继续提供半初始化服务。
- 前端 Domain 只能通过 app-core/runtime/api-client/ui 公共合同组合；禁止 app-to-app import 和复制业务客户端。
- 所有默认测试 hermetic；需要公网、云服务或签名凭据的检查必须显式分层，缺少配置时给出可判定状态而非 panic。

### Verification Integrity

当前判卷基线（2026-08-24）：root `1126a76e18fe11dcde33e4bb9102545b12182a12`，backend `750537c0b8edde522f9c9dddc2bfcf64169689ea`，frontend `1d1af04e3305faf3e7087fe4503b2fb132009663`。后端 `make build` 通过；前端 220 个 unit tests、type-check、生产 build 通过；lint 0 error/29 个既有 warning。后端 `go test ./...` 因 `TestKODOUpload` 缺少 `test.png`、`TestOBSUpload` 缺 endpoint 后 nil client panic 而失败；这是 G0 已知红线，T-01 必须在不删除断言、不默认 skip、不注入真实云凭据的条件下消除。

每 Ticket 先执行定向 source checks，再由 Lead 在同一 current/direct-parent 状态执行适用集成和 E2E。Speculo 全局命令基线为：

```sh
(cd go-admin-plus && go test ./...) && (cd go-admin-ui-plus && pnpm test:unit)
cd go-admin-ui-plus && pnpm type-check
cd go-admin-ui-plus && pnpm lint
(cd go-admin-plus && make build) && (cd go-admin-ui-plus && pnpm build:prod)
```

允许记录 warning，但 warning 不得覆盖 error；不得把命令改成忽略退出码、缩小到遗漏受影响路径，或以生成缓存/已存在产物代替本次构建。前端还存在 `pnpm.overrides` 位置 warning 和一个旧 CSS pseudo-element 构建 warning，分别由 T-05 workspace manifest 和后续前端迁移中的实际 owner 处理；未造成失败时作为基线风险追踪。

### Migration or Release Sequence

1. T-01 冻结行为和可恢复 fixture；T-02/T-05 只做可逆 prefactor/expand。
2. T-03 建立双方言 migration/FileStore contract，T-04 保持 Server 兼容，T-06 建立生成合同。
3. T-07 迁移 Domain/routeKey，T-08 建立 Desktop tracer；兼容入口持续存在。
4. T-09 验证桌面升级/安全，T-10 验证 Compose 空卷、升级和回滚。
5. T-11/T-12 在各自原生 runner 上生成并验证明确标识 `unsigned-self-use` 的离线安装包；必须附 checksum/SBOM/trust state 和单应用 self-use 安装说明，不把它伪称平台可信签名。
6. T-13 汇合同版本 manifest，先以扫描和 E2E 证明消费者迁移为零旧调用，再收缩兼容层；外部 publish/deploy 仍需另行授权。

### Risks, Monitoring and Recovery

| 风险 | 前置信号/监控 | 恢复策略 |
|---|---|---|
| 宽重构造成行为漂移 | characterization、OpenAPI diff、Admin E2E | 保留 expand adapter，回到最近 result 后定向修复 |
| 双数据库语义分叉或数据损坏 | migration fixture、checksum、备份恢复测试 | fail closed；Desktop 恢复备份，Server 走受控前向 migration |
| Desktop loopback 被同机进程或恶意页面调用 | bind/socket、token、Origin、单实例测试 | 不启动 UI/业务服务；轮换 token 并修复边界 |
| 前端 workspace 形成循环依赖或新复制层 | import graph、生成零 diff、duplicate scan | 回退违规 Domain 迁移，修公共合同 owner |
| Compose 在空卷/升级环境才失败 | health、容器日志、volume/schema 版本 | 回滚镜像，恢复卷备份，禁止自动重复迁移 |
| unsigned 包触发 Gatekeeper/SmartScreen 或 WebView2 离线依赖缺失 | native verification、checksum、self-use 安装与干净机离线矩阵 | 先核验来源，只对单应用建立例外；受管策略/Smart App Control 阻断时报告 unsupported，不关闭全局安全能力 |
| current workspace 前序提交污染后序 | 每 Ticket path diff、parent SHA、clean check | 最多 3 次集成修复；仍失败则标 blocked 并停止后序 |

### Deviation Control

遵循 `<Path>{roots.workflows}/specdev/common/rules/deviation-control.md</Path>`。公共行为、数据、安全、迁移、目标平台或发布合同变化返回 Spec；依赖、路径 owner、验证接缝或 Ticket 局部实现合同变化返回 Ticket/Map；编排顺序、Gate 或 workspace 策略变化更新 Goal Plan。未经同步和 Lead 核验不得继续下一个 Ticket。

## 6. Progress and Decisions

### Current Status

- Plan 状态：`completed`；T-01 至 T-13 已按严格串行完成，G0-G8 全部关闭。T-13 result 为 root `2f01bae` / backend `1859e63` / frontend `4899b36`；Linux/macOS/Windows runs `32880870190`、`32880876001`、`32880882112` 与 Product run `32881713952` 全部通过。
- Workspace：current `main`；implementation limit 1；direct-parent；无 worktree。
- Baseline：规划基线 root `1126a76`，T-01 实施前 root `81106fb`，backend `750537c`，frontend `1d1af04`。
- 最近验证：T-08 macOS ARM64 原生 Wails tracer 与 GitHub Windows Server 2025 AMD64 run `32842185071` 均完成真实 WebView 登录、getinfo 和 demo product create/list/update/delete，并输出唯一 `GO_ADMIN_DESKTOP_E2E_PASS`；Windows run 固定 backend `a51e7dd`、frontend `3bfc10f`，环境为 Wails 2.15.0/WebView2 151.0.4129.86。全量 Go/race/vet、前端 build 与 136 mocked E2E 继续通过，T-08 result 为 root `2cd0139`。
- T-09 result 为 root `321616c` / backend `b901a56` / frontend `3bfc10f`；macOS ARM64 原生 GUI/双开/重启/备份与 Windows Server 2025 run `32852402806` 的 profile/native hardening、Wails build、WebView2 login/getinfo/CRUD 均通过，精确 marker 为 `GO_ADMIN_DESKTOP_E2E_PASS`。
- T-10 result 为 root `085eb94` / backend `0b42bfe` / frontend `7887d32`；Linux AMD64 run `32860819614` 通过空卷 Compose、生产验证码登录、CRUD 持久化、迁移/依赖失败、安全和 release artifact/SBOM gate，artifact 为 `9568298586`。
- T-11 result 为 root `bf45f45` / backend `365ea9c` / frontend `7887d32`；macOS ARM64 run `32868421377` 通过全量 Go、两次真实 WebView 登录/CRUD、持久化重启、正式 ARM64 app、ad-hoc/no-Team/no-Authority、DMG、scoped quarantine gate、SPDX/checksum 和 artifact `9571215915`。artifact 明确为 `unsigned-self-use`，未执行 Developer ID/notary 或外部 publish。
- 未验证：当前 Phase 1 范围内无。字面 Windows 10/11 干净 VM 属于未来 external publish 前人工烟测，当前 Windows Server 2025 AMD64 证据不伪装为该环境。

### Pending Decisions and Blockers

DEV-001（spec + release）由用户在 2026-08-25 明确批准：Phase 1 的 macOS/Windows 正式桌面产物改为 unsigned self-use，以可安装可用优先，允许设备 owner 在校验来源后对单应用确认运行；平台可信签名、公证和 Store 分发延期。Apple 官方确认可在首次尝试后通过 Privacy & Security 的 Open Anyway 对未知开发者 App 建立例外；Microsoft 官方说明 unsigned 文件通常可由 SmartScreen 的 Run anyway 继续，但 Smart App Control 或企业策略可能禁止。实现禁止建议全局关闭平台安全能力。当前无 blocker；外部 publish/deploy 仍需新授权，不影响本 change 完成。

### Resume Protocol

恢复时依次读取 Goal Plan、当前 Ticket、change `.status.json`、最新 Evidence、根与两个子仓库状态。确认根 HEAD 包含最后通过的 `result_sha`、子仓库 gitlink 与 Evidence SHA 一致、工作区 clean；若当前 Ticket 尚未形成 checkpoint，从其 path diff 和 source checks 恢复；若 checkpoint 已形成但 Gate 未关闭，由 Lead 重跑 direct-parent checks/E2E。不得跳过失败 Gate 或从后序 Ticket开始。

## Assumptions

- Phase 1 的 Linux 容器目标为 AMD64，桌面目标仅 macOS ARM64 与 Windows AMD64；均由 Spec 明确，可在各原生 runner 验证。
- current workspace 的本地 `main` 是执行父分支；用户已在 2026-08-25 明确授权三个仓库当前 Goal Plan 的 push 和相关 workflow dispatch。PR merge、签名、公证、外部 Release publish、部署和生产迁移仍不在该授权内。
- Phase 1 桌面正式 artifact class 为 `unsigned-self-use`；必须与 `unsigned-development`/未来 `signed-production` 区分，并在 manifest、文件名、安装说明和 Evidence 中保持一致。
