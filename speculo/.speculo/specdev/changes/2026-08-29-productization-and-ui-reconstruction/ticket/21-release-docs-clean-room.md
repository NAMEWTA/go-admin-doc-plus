---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-21
title: 收敛文档并完成三 Profile Clean-room 候选
status: in_progress
planning_depth: deep
planning_depth_reason: 最终 Gate 跨数据库迁移、部署、Desktop、本地持久化、运维恢复和发布证据，决定 change 是否可完成
ready: true
risk: high
blocked_by: [T-17, T-18, T-19, T-20]
contract_ids: [AC-030, AC-031, AC-032, AC-033, AC-034, AC-035, AC-036, AC-037, AC-038, AC-039]
owner: codex-root
expected_changes: ["<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>database/README.md</Path>", "<Path>deploy/**</Path>", "<Path>release/**</Path>", "<Path>scripts/release/**</Path>", "<Path>scripts/quality/docs-check.mjs</Path>", "<Path>scripts/quality/docs-check.test.mjs</Path>"]
writable_paths: ["<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>database/README.md</Path>", "<Path>deploy/**</Path>", "<Path>release/**</Path>", "<Path>scripts/release/**</Path>", "<Path>scripts/quality/docs-check.mjs</Path>", "<Path>scripts/quality/docs-check.test.mjs</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>go-admin-plus/**</Path>", "<Path>go-admin-plus-ui/**</Path>", "<Path>.github/workflows/ci.yml</Path>"]
shared_paths: ["<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>deploy/**</Path>", "<Path>release/**</Path>", "<Path>scripts/release/**</Path>"]
shared_path_owners: ["<Path>README.md</Path> => T-21", "<Path>docs/**</Path> => T-21", "<Path>deploy/**</Path> => T-21", "<Path>release/**</Path> => T-21", "<Path>scripts/release/**</Path> => T-21"]
---

# Ticket T-21: 收敛文档并完成三 Profile Clean-room 候选

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/21-release-docs-clean-room.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-21.md</Path>`

## 1. 战略与来源

- **目标：** 让新用户只依赖当前文档即可从空库运行三个 profile，并用完整候选证据结束产品化重构。
- **可观察产出：** Server SQLite、Server PostgreSQL、Desktop SQLite 完成 migrate/setup/login/core management/restart；文档无旧命令/固定密码。
- **来源：** `US-013~017`、`AC-030~039`、`PLAN/P8`。
- **当前事实：** docs/deploy/release 描述当前旧命令和自动迁移，database README/未提交 bootstrap SQL 包含固定凭据说明。
- **Planning Depth 原因：** clean-room 涵盖实际持久数据、部署顺序和最终发布判定，不能只做文案检查。

## 2. 决策状态

### 已锁定决策

- 文档只描述新 CLI、Bootstrap/recovery、profile migration、配额、日志/Doctor、E2E 和故障处理。
- 三 profile 从空库实际演练；required gate 不允许未执行记通过。
- 个人自用签名/公证为 not-required，不伪装 passed。

### 已采用的低影响假设

- clean-room 使用临时/明确 disposable roots；命令由 Taskfile/产品 CLI 直接提供，不建立文档专用脚本分叉。

### 已批准执行偏差 DEV-21-001

- **触发事实：** T-21 source 上 `task release:verify` 的首个诚实红灯来自 `scripts/quality/compatibility-zero.mjs`：Windows `path.relative()` 返回反斜杠路径，而 `ownFiles` 与 `allowedMatches` 使用仓库规范 `/`，导致扫描器自身、反向测试夹具和已审核窄 allowlist 被错误报告为禁用体系残留。同一内容在 POSIX runner 可通过，红灯不代表产品重新引入旧体系。
- **批准范围：** T-21 临时拥有 `scripts/quality/compatibility-zero.mjs` 与 `compatibility-zero.test.mjs`，只把扫描所得相对路径归一化为 `/` 后再执行既有 own-file/allowlist 比较和诊断；测试必须在当前 Windows host 锁定 forward-slash 结果。
- **禁止扩大：** 不修改 removed paths、forbidden regex、allowed match 项、扫描根/扩展名、失败语义或任何产品/CI/release 行为；不新增 skip、平台豁免或宽泛目录排除。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** T-21 source `1b5b7c14504c86cacd7a36decfee1fccd73812b9` 已形成并通过 docs/policy 定向验证；兼容扫描修正尚未形成 checkpoint。

### 已批准执行偏差 DEV-21-002

- **触发事实：** DEV-21-001 路径归一化后，Windows 扫描从大量路径误报收敛到两个真实未登记负例：`go-admin-plus/internal/modules/files/migrations/0020-capacity/provider_test.go` 以 forbidden 列表断言 migration 不含 `tenant`，`go-admin-plus/internal/platform/logging/redaction.go` 以 `mysql://` 标记保证旧 DSN 即使进入日志也被脱敏。两者都是防止旧体系回归的保护代码，不是活动产品能力。
- **批准范围：** 只向既有精确 `allowedMatches` 增加上述路径各自的单一现有名称：Files 测试允许 `tenant feature`，日志 redaction 允许 `MySQL`；现有单测继续要求活动模块中的同名引用失败，并锁定诊断使用 `/`。
- **禁止扩大：** 不允许其他路径、名称、目录或 regex 豁免，不改变被扫描源码、forbidden/removed 集合、扫描根、失败语义或发布行为。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** 已授权，尚未形成修正 checkpoint；`task compatibility:zero` 与 `task release:verify` 保持红灯。

### 已批准执行偏差 DEV-21-003

- **触发事实：** 首个 T-21 integration candidate `07a2e3bcf2df6cfaf1b9afcebefc20ba01dbc70d` 的 `task architecture:check` 精确返回 `frontend root typecheck omits test project tests/e2e/web-shell/tsconfig.json`。T-20 已增加并实际使用该测试项目，但 workspace 根 `typecheck` 仍只串列此前 10 个 E2E tsconfig，导致 required full candidate 不覆盖 Web Shell 类型边界。
- **批准范围：** T-21 临时拥有 `go-admin-plus-ui/package.json` 的 `scripts.typecheck` 单一字符串，只在既有 E2E `vue-tsc -p` 链末尾追加 `tests/e2e/web-shell/tsconfig.json`；architecture check 与完整 typecheck 必须同时通过。
- **禁止扩大：** 不修改依赖、lockfile、任何 tsconfig、T-20 测试/产品代码、lint/test/build 命令或 workspace package；不增加 skip、并行掩盖或类型排除。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** 已授权，尚未形成修正 checkpoint；candidate `07a2e3b` 保留为 architecture 首红。

### 已批准执行偏差 DEV-21-004

- **触发事实：** 当前 T-21 integration candidate `85daa4178cbc7780c93f334667a75bc30b78c2ed` 的完整 `task generate:check` 在 29 项 Generator 契约测试中精确失败 1 项：`scripts/contracts/modules.test.mjs` 使用 POSIX 字面量 `/workspace/product`，Windows 上被实现的 `path.resolve()` 正确解析为当前盘符绝对路径，而测试的 `path.join()` 期望仍是无盘符根路径，因而产生仅测试预期的跨平台差异。
- **批准范围：** T-21 临时拥有 `scripts/contracts/modules.test.mjs`，只允许用 `path.resolve()` 构造跨平台绝对 fixture root，并用相同 `resolve()` 语义断言 Go/TypeScript 输出；完整 29 项 Generator 契约测试与 `task generate:check` 必须重新通过。
- **禁止扩大：** 不修改 `scripts/contracts/modules.mjs`、模块路径语法、生成输出、manifest、负向夹具、测试数量或失败语义；不新增平台分支、skip、allow-failure 或目录豁免。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** 已授权，尚未形成修正 checkpoint；candidate `85daa41` 及本次 `28 pass / 1 fail / 0 skip` 输出保留为 Generator 首红。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| README/docs/database/deploy/release 更新、docs checks、三 profile rehearsal/evidence | T-17~20 成果、Taskfile/CLI、现有 release policy | commit/push/publish/deploy 生产、签名/公证、兼容文档 |

## 4. 要构建什么

操作者按文档选择 profile，使用明确持久根完成配置、迁移、Bootstrap、serve/worker、Doctor、登录、管理和重启；Desktop 走原生首次设置。故障排查覆盖 schema mismatch、管理员恢复、低磁盘、Session 撤销和备份恢复。Lead 在干净候选上逐项执行并把实际命令/环境/结果关联到 AC。

## 5. 实现契约

- **入口或接缝：** README、development/database/deployment/security/troubleshooting docs、release verify/rehearsal。
- **输入与输出：** 当前 Task/CLI/config；输出可复制命令、明确 profile 行为和候选 Evidence。
- **公共接口变化：** 文档/部署配置与产品 CLI 对齐；不新增产品 API。
- **不变量：** 无固定密码/旧命令/自动 PG migration；路径为当前仓库；not-run 不写 passed。
- **状态或数据流：** clean root -> configure/migrate/bootstrap -> run -> login/core flows -> restart -> verify/cleanup。
- **错误与失败行为：** 任一步失败停止候选并分类；不得通过修改数据跳过产品流程。
- **兼容要求：** 零旧文档兼容；过时模式全部删除。
- **安全与隐私要求：** 示例仅 secret file/reference，不写真实 DSN/password；rehearsal 使用 disposable data。

## 6. 执行路线

1. 扫描并建立旧命令、固定凭据、迁移/签名错误描述红灯 docs tests。
2. 更新 README、开发、数据库、部署、安全、恢复和故障排查。
3. 更新 Compose/release policy 与新 CLI/profile 拓扑一致。
4. 在 clean candidate 依次演练 Server SQLite、Server PostgreSQL、Desktop SQLite。
5. 汇总所有 quality/PG/Web/native/security/docs Evidence，确认无未批准偏差。

## 7. 路径访问契约

- **预计修改点/可写范围：** README/docs/database README、deploy/release/scripts/docs checks。
- **只读上下文：** 产品代码、UI、CI、Taskfile。
- **共享路径：** 文档/部署/release 路径仅 T-21 可写；T-09 先完成命令面后本 Ticket最终收敛。
- **保留或不动：** 不提交、推送、发布、部署或归档；用户持久 `dev_store` 不用于 destructive rehearsal。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | clean-room | 三 profile 当前文档流程 | setup/login/core/restart 全通过 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-21.md</Path>` |
| 失败路径 | docs/release reverse | 旧命令/固定 password/schema mismatch/缺 gate | docs check 或候选失败 | 同上 |
| 回归 | full candidate | governance/architecture/zero/docs/generate/Go/pnpm/Rust/PG/Web/native/security | 全部适用 Gate 通过 | 同上 |

- **Workspace checks：** source-worktree 仅运行 docs/policy checks；clean-room 必须由 Lead 在 current-workspace 或 parent-candidate 完成。
- **E2E disposition：** required：三 profile 从空库到重启是最终产品系统验收。
- **E2E owner/environment：** Lead / parent-candidate（required policy）或 current-workspace；使用明确 disposable roots，复用 T-19/T-20 证据并执行 release rehearsal。
- **Integration evidence：** source commit、candidate/result SHA、三 profile 命令/环境/退出状态、父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 所有实现/CI/E2E完成 -> docs/release diff -> clean-room -> final result；PG 先 migrate，Desktop 先备份。
- **兼容窗口：** 无；旧文档、命令和固定 SQL 全部收缩。
- **监控信号：** rehearsal phase、schema/setup/doctor/readiness、E2E/security status。
- **回滚或前向恢复：** candidate 失败不发布；清理 disposable roots，修复后重建；数据回退依靠 rehearsal 备份。
- **不可逆操作与批准点：** 只在 disposable 数据执行；任何 publish/deploy/archive 需另行用户明确授权。
- **收缩条件：** docs/repo 扫描无旧命令、固定管理员密码、租户/Redis/兼容模式；全部 AC 有 Evidence。

## 10. 验收标准

- [ ] `AC-030~039` 的最终候选和文档合同成立。
- [ ] 三 profile 从空库可复现，签名/公证正确记为 not-required。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-21.md</Path>`。
- [ ] required clean-room、commit、candidate/result、父分支包含和 Lead 审查完整。
