---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-21
title: 原子切换、质量门禁与旧体系归零
status: in_progress
planning_depth: deep
planning_depth_reason: 全仓不可逆删除、目录改名、根 CI 与兼容归零是最终高风险收缩操作
ready: true
risk: critical
blocked_by: [T-18, T-19, T-20]
contract_ids: [AC-001, AC-002, AC-023, AC-028, AC-029, AC-033]
owner: codex-root
expected_changes: ["<Path>Taskfile.yml</Path>", "<Path>.gitignore</Path>", "<Path>.github/workflows/ci.yml</Path>", "<Path>.github/workflows/codeql.yml</Path>", "<Path>.github/workflows/product-release.yml</Path>", "<Path>.github/workflows/desktop-windows-tracer.yml</Path>", "<Path>scripts/quality/**</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>deploy/README.md</Path>", "<Path>database/README.md</Path>", "<Path>release/README.md</Path>", "<Path>release/manifest/**</Path>", "<Path>go-admin-ui-plus/**</Path>", "<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/common/**</Path>", "<Path>go-admin-plus/api/**</Path>", "<Path>go-admin-plus/template/**</Path>", "<Path>go-admin-plus/internal/profile/**</Path>", "<Path>go-admin-plus/internal/tenant/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>"]
writable_paths: ["<Path>Taskfile.yml</Path>", "<Path>.gitignore</Path>", "<Path>.github/workflows/ci.yml</Path>", "<Path>.github/workflows/codeql.yml</Path>", "<Path>.github/workflows/product-release.yml</Path>", "<Path>.github/workflows/desktop-windows-tracer.yml</Path>", "<Path>scripts/quality/**</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>scripts/go-admin-plus-ui/common.sh</Path>", "<Path>scripts/go-admin-plus-ui/build.sh</Path>", "<Path>scripts/go-admin-plus-ui/dev.sh</Path>", "<Path>scripts/go-admin-plus-ui/lint.sh</Path>", "<Path>scripts/go-admin-plus-ui/package.sh</Path>", "<Path>scripts/go-admin-plus-ui/path-contract.sh</Path>", "<Path>scripts/go-admin-plus-ui/test.sh</Path>", "<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>deploy/README.md</Path>", "<Path>database/README.md</Path>", "<Path>release/README.md</Path>", "<Path>release/manifest/**</Path>", "<Path>release/linux/Containerfile.server.dockerignore</Path>", "<Path>release/linux/Containerfile.web.dockerignore</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/.gitignore</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/binaries/.gitignore</Path>", "<Path>go-admin-ui-plus/**</Path>", "<Path>go-admin-plus/.github/**</Path>", "<Path>go-admin-plus/.gitignore</Path>", "<Path>go-admin-plus/.dockerignore</Path>", "<Path>go-admin-plus/.go-version</Path>", "<Path>go-admin-plus/AGENTS.md</Path>", "<Path>go-admin-plus/README.md</Path>", "<Path>go-admin-plus/LICENSE.md</Path>", "<Path>go-admin-plus/_config.yml</Path>", "<Path>go-admin-plus/main.go</Path>", "<Path>go-admin-plus/restart.sh</Path>", "<Path>go-admin-plus/stop.sh</Path>", "<Path>go-admin-plus/go-admin-db.db</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>", "<Path>go-admin-plus/Makefile</Path>", "<Path>go-admin-plus/Dockerfile</Path>", "<Path>go-admin-plus/docker-compose.yml</Path>", "<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/common/**</Path>", "<Path>go-admin-plus/api/**</Path>", "<Path>go-admin-plus/template/**</Path>", "<Path>go-admin-plus/static/**</Path>", "<Path>go-admin-plus/ssh/**</Path>", "<Path>go-admin-plus/scripts/**</Path>", "<Path>go-admin-plus/docs/**</Path>", "<Path>go-admin-plus/cmd/api/**</Path>", "<Path>go-admin-plus/cmd/app/**</Path>", "<Path>go-admin-plus/cmd/config/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-plus/cmd/migrate/**</Path>", "<Path>go-admin-plus/cmd/version/**</Path>", "<Path>go-admin-plus/cmd/cobra.go</Path>", "<Path>go-admin-plus/config/README.md</Path>", "<Path>go-admin-plus/config/db*</Path>", "<Path>go-admin-plus/config/pg.sql</Path>", "<Path>go-admin-plus/config/extend.go</Path>", "<Path>go-admin-plus/config/seeds.go</Path>", "<Path>go-admin-plus/config/settings*.yml</Path>", "<Path>go-admin-plus/internal/application/*.go</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus/internal/host/desktop/**</Path>", "<Path>go-admin-plus/internal/modules/default.go</Path>", "<Path>go-admin-plus/internal/modules/default_test.go</Path>", "<Path>go-admin-plus/internal/modules/jobs.go</Path>", "<Path>go-admin-plus/internal/modules/runtime_queue.go</Path>", "<Path>go-admin-plus/internal/platform/dependencies.go</Path>", "<Path>go-admin-plus/internal/platform/dependencies_test.go</Path>", "<Path>go-admin-plus/internal/platform/files.go</Path>", "<Path>go-admin-plus/internal/platform/files_test.go</Path>", "<Path>go-admin-plus/internal/platform/cache/**</Path>", "<Path>go-admin-plus/internal/platform/localcache/**</Path>", "<Path>go-admin-plus/internal/platform/observability/**</Path>", "<Path>go-admin-plus/internal/profile/**</Path>", "<Path>go-admin-plus/internal/tenant/**</Path>", "<Path>go-admin-plus/test/characterization/**</Path>", "<Path>go-admin-plus/test/desktop/**</Path>", "<Path>go-admin-plus/test/api.go.template</Path>", "<Path>go-admin-plus/test/model.go.template</Path>", "<Path>go-admin-plus/test/gen_test.go</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/app/product/**</Path>", "<Path>go-admin-plus/internal/application/health/**</Path>", "<Path>go-admin-plus/internal/modules/*/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>", "<Path>go-admin-plus/internal/platform/coordination/**</Path>", "<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus/internal/platform/desktop/**</Path>", "<Path>go-admin-plus/internal/platform/migrations/**</Path>", "<Path>go-admin-plus/internal/platform/outbox/**</Path>", "<Path>go-admin-plus/cmd/go-admin-plus/**</Path>", "<Path>go-admin-plus/cmd/desktop-sidecar/**</Path>", "<Path>go-admin-plus/cmd/config-check/**</Path>", "<Path>go-admin-plus/config/schema/**</Path>", "<Path>go-admin-plus-ui/**</Path>", "<Path>release/linux/**</Path>", "<Path>release/macos/**</Path>", "<Path>release/windows/**</Path>", "<Path>release/shared/**</Path>", "<Path>deploy/**</Path>"]
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

### T21-D01（Lead 批准）

G7 RED 盘点证明旧目录不是全部孤立文件：未被正式入口消费的 expand 期 `internal/modules` 根兼容集合、`profile`、旧 Desktop host、root platform adapter、legacy command/template/test 仍引用待删除的 `app/common/api/tenant`；根 Task、CI、Windows tracer 与聚合 release manifest 也仍调用旧 root main、`sqlite3` tag、Wails、旧前端名和 unsigned-self-use。若只按初始路径表删除 12 项治理文件与显式旧目录，全量构建立即失败，且 AC-002/AC-029/AC-033 会形成伪通过。

因此 Lead 基于用户已批准的“零兼容完整重构”精确开放 frontmatter 新增路径：删除不在 `cmd/go-admin-plus`、`cmd/desktop-sidecar` 和 `cmd/config-check` 正式依赖闭包中的遗留 scaffolding；`go mod tidy` 删除 Wails/go-admin-core/Casbin/Redis/MySQL/SQL Server 等失去消费者的依赖；根 Task/CI/docs/quality/release manifest 只引用新入口、三种 profile、双 App 和 T18-T20 平台 policy。新 product/module/platform 实现与三个正式 command 保持只读，任何实际需要修改只读产品代码的发现必须再次登记偏差。

### T21-D02（Lead 批准）

首次删除后的 compile RED 纠正了依赖分类：`internal/application` 与 `internal/contracts` 是 `product.Build` 的直接依赖，`localcache` 仍由可靠运行时验收消费，`cache/observability` 也是已完成的新平台能力；四者全部恢复且不得以“入口暂未调用”为由收缩。治理预检同时证明 `scripts/go-admin-plus-ui/{build,dev,package}.sh` 使用不存在的 `@go-admin/desktop` filter，Linux 两份 dockerignore 与 `<Path>go-admin-plus/config/README.md</Path>` 仍引用旧前端。Lead 只追加开放这 7 个治理文件：改为 `@go-admin/admin-desktop` 和当前 Workspace/目录合同；产品源码仍保持只读。

### T21-D03（Lead 批准）

根治理收口证明前端公共脚本的 workspace 根、lint 和路径合同仍消费已删除的旧目录，而 Tauri 2 App 内两份 `.gitignore` 会继续形成嵌套 Git 治理。Lead 精确追加开放 `scripts/go-admin-plus-ui/{common,lint,path-contract}.sh` 与两份 Tauri `.gitignore`：脚本统一指向 `<Path>go-admin-plus-ui/</Path>`，忽略规则迁移到根 `<Path>.gitignore</Path>` 后删除嵌套文件。Tauri/Rust/TypeScript 产品源码与其他前端文件继续只读。

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
- **E2E disposition：** deferred/final：T-21 candidate 先完成删除、全量非 E2E Gate 与零兼容扫描；其实现 result 进入 `main` 后，Lead 创建唯一最终系统候选并统一执行三 profile、双 App、原生发行与全模块 E2E。
- **E2E owner/environment：** Lead / 最终系统候选；source-worktree 与 T-21 实现 candidate 不运行或声明 E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、删除清单、零兼容/全量非 E2E Gate、最终统一 E2E 与父分支包含关系。

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
