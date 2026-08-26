---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-21
title: 原子切换、质量门禁与旧体系归零
status: ready
planning_depth: deep
planning_depth_reason: 全仓不可逆删除、目录改名、根 CI 与兼容归零是最终高风险收缩操作
ready: true
risk: critical
blocked_by: [T-18, T-19, T-20]
contract_ids: [AC-001, AC-002, AC-023, AC-028, AC-029, AC-033]
owner: unassigned
expected_changes: ["<Path>.github/workflows/ci.yml</Path>", "<Path>scripts/quality/**</Path>", "<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>go-admin-ui-plus/**</Path>", "<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/common/**</Path>", "<Path>go-admin-plus/internal/tenant/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>"]
writable_paths: ["<Path>.github/workflows/ci.yml</Path>", "<Path>.github/workflows/codeql.yml</Path>", "<Path>scripts/quality/**</Path>", "<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>go-admin-ui-plus/**</Path>", "<Path>go-admin-plus/.github/**</Path>", "<Path>go-admin-plus/.gitignore</Path>", "<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/common/**</Path>", "<Path>go-admin-plus/api/**</Path>", "<Path>go-admin-plus/internal/tenant/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-plus/config/db*</Path>", "<Path>go-admin-plus/docker-compose.yml</Path>", "<Path>go-admin-plus/scripts/**</Path>", "<Path>go-admin-plus/docs/**</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>go-admin-plus/internal/app/product/**</Path>", "<Path>go-admin-plus-ui/**</Path>", "<Path>release/**</Path>", "<Path>deploy/**</Path>"]
shared_paths: ["<Path>.github/workflows/ci.yml</Path>", "<Path>scripts/quality/**</Path>", "<Path>README.md</Path>", "<Path>docs/**</Path>"]
shared_path_owners: ["<Path>.github/workflows/ci.yml</Path> => T-21", "<Path>scripts/quality/**</Path> => T-21", "<Path>README.md</Path> => T-21", "<Path>docs/**</Path> => T-21"]
---

# Ticket T-21: 原子切换、质量门禁与旧体系归零

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/21-atomic-cutover-contraction.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>`

## 1. 战略与来源

- **目标：** 在新产品与三平台候选全部验证后，一次性删除旧结构并启用最终根 CI/文档/归零门禁。
- **可观察产出：** 仓库只保留目标目录与治理；旧 API/schema/config/Wails/JWT/Casbin/Redis/tenant/旧数据库支持和临时标记零命中。
- **来源：** `US-001`、`US-016`、`US-018`、`AC-001`、`AC-029`、`AC-033`、`ADR-001`、`ADR-021`、`ADR-022`。
- **当前事实：** 用户已有部分根迁移改动，旧前后端与目标结构仍并存；不可回滚这些用户改动。
- **Planning Depth 原因：** 删除和命名切换不可逆且影响全仓构建、文档、发行与开发入口。

## 2. 决策状态

### 已锁定决策

- 只在 T-17 至 T-20 Gate 全通过后执行 contract；不发布混合结构或旧新双轨。
- 后端保留 `<Path>go-admin-plus/</Path>`，前端最终仅 `<Path>go-admin-plus-ui/</Path>`。
- Local/Hook、PR、Protected Release 复用根任务；豁免必须包含 owner、范围、原因、期限和 Ticket。

### 已采用的低影响假设

- 零兼容扫描使用明确 allowlist，仅允许 SpecDev 历史工件和必要否定性测试文本。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 旧路径删除、最终 CI、质量扫描、当前文档、产品命名切换 | T-01 至 T-20 已验证目标产品 | 旧数据迁移、远端发布、历史工件改写 |

## 4. 要构建什么

Lead 先冻结已通过的产品/发行证据，再删除旧目录和所有兼容调用点，更新根 CI 与当前文档；全量合同、架构、三 profile、双 App、三平台和零兼容扫描全部通过后才接受原子切换结果。

## 5. 实现契约

- **入口或接缝：** contraction inventory、zero-compat scanner、root CI policy、全量 root tasks、文档链接检查。
- **输入与输出：** 已验证 result SHA 与 allowlist；输出唯一目标仓库、完整 Gate 和删除清单 Evidence。
- **公共接口变化：** 移除全部旧外部/内部合同，只保留 Spec 定义的新产品。
- **不变量：** 不删新结构/用户无关改动/SpecDev 历史；任何 required Gate 失败阻断；无过期豁免。
- **状态或数据流：** freeze evidence -> inventory -> delete/rename -> docs/CI -> zero scan -> full matrix -> result SHA。
- **错误与失败行为：** 未知引用、扫描命中、平台 Gate 缺失或文档漂移立即停止，不形成部分切换结果。
- **兼容要求：** 明确为零兼容；无 shim、双写、legacy alias 或临时 migration 标记。
- **安全与隐私要求：** 删除前确认目标；CI/文档/扫描输出不含 secret；保留供应链证据。

## 6. 执行路线

1. 锁定 T-17 至 T-20 result SHA、制品证据和精确旧路径 inventory。
2. 先运行预期会命中的零兼容扫描并确认 allowlist。
3. 原子删除旧前端、旧后端层、旧宿主/配置/数据库/治理资产并完成命名切换。
4. 更新根 CI、质量脚本和当前产品文档，禁止过期豁免。
5. 运行全量 root tasks、三 profile、双 App、三平台证据检查和最终零扫描。

## 7. 路径访问契约

- **预计修改点：** frontmatter 列出的最终 CI/文档与精确旧路径。
- **可写范围：** 仅 `writable_paths`；删除前逐项解析，不使用宽泛工作区删除命令。
- **只读上下文：** 新产品、发行和根任务全部只读验证。
- **共享路径：** 根 CI、quality scripts、README/docs 由 T-21 唯一拥有。
- **保留或不动：** `<Path>speculo/**</Path>`、`.git`、新产品模块和用户无关改动。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | final product gate | `task test lint generate migrate package release:verify` | 根任务、产品和发行证据全部通过 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` |
| 失败路径 | contraction/policy fixture | 保留一项旧引用、过期豁免或缺平台证据 | CI/扫描阻断且不接受 result | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` |
| 回归 | zero scan/full matrix | `task governance:check architecture:check compatibility:zero` | allowlist 外旧模式零命中，三 profile/双 App 仍通过 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` |

- **Workspace checks：** Goal Plan 在 current-workspace 或 source-worktree 运行非 E2E 全量门禁；删除前后都记录状态。
- **E2E disposition：** required：最终三 profile、双 App 和原生发行证据必须在集成结果上复核。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不运行或声明 required E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、删除清单、全量 E2E 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** expand 已由 T-01 至 T-20 完成；本 Ticket 仅在所有 Gate 后执行 contract。
- **兼容窗口：** 无产品兼容窗口，切换结果不可包含旧新双轨。
- **监控信号：** zero-scan hits、root Gate、profile/App/release result 和豁免到期。
- **回滚或前向恢复：** 删除 commit 在发布前可整体回滚；发布后只允许基于新架构前向修复，不恢复旧体系。
- **不可逆操作与批准点：** 用户本次批准覆盖规划发布，不自动授权实现删除；I-implement 仍须取得 implementation/integration 授权并在删除前确认精确 inventory。
- **收缩条件：** allowlist 外旧目录/名称/API/schema/config/Wails/JWT/refresh/Casbin/Redis/tenant/MySQL/SQL Server/临时标记全部零命中。

## 10. 验收标准

- [ ] `AC-001/AC-002`：唯一根治理、目标目录和根任务合同成立。
- [ ] `AC-023/AC-028`：合同生成、边界和 clean-tree 检查通过。
- [ ] `AC-029`：明确 allowlist 外全部旧模式零命中。
- [ ] `AC-033`：Local/PR/Protected Release 分层门禁失败均阻断且豁免可审计。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 final result SHA；Ticket、Map、Evidence 一致。
