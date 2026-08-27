---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-16
title: Tauri 2 Desktop 安全宿主闭环
status: in_progress
planning_depth: deep
planning_depth_reason: 原生宿主、sidecar、Stronghold、随机 loopback、SQLite 迁移与进程监督构成关键桌面安全边界
ready: true
risk: critical
blocked_by: [T-06, T-08, T-14]
contract_ids: [AC-006, AC-007, AC-009, AC-021, AC-027, AC-036]
owner: codex-t16-desktop
expected_changes: ["<Path>go-admin-plus-ui/apps/admin-desktop/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/desktop/**</Path>", "<Path>go-admin-plus/cmd/desktop-sidecar/**</Path>", "<Path>go-admin-plus/internal/platform/desktop/**</Path>", "<Path>release/shared/sidecar/**</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/apps/admin-desktop/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/desktop/**</Path>", "<Path>go-admin-plus/cmd/desktop-sidecar/**</Path>", "<Path>go-admin-plus/internal/platform/desktop/**</Path>", "<Path>go-admin-plus/test/desktop/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>release/shared/sidecar/**</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/app/kernel/**</Path>", "<Path>go-admin-plus/internal/modules/iam/session/**</Path>", "<Path>go-admin-plus/internal/modules/demo/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/core/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/packages/adapters/desktop/**</Path>", "<Path>release/shared/sidecar/**</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/packages/adapters/desktop/**</Path> => T-16 under T16-D01; desktop runtime adapter reserved by T-05", "<Path>release/shared/sidecar/**</Path> => T-16", "<Path>go-admin-plus-ui/package.json</Path> => T-16 under T16-D02; Desktop aggregate checks only", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path> => T-16 under T16-D02; Desktop specs only", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-16 under T16-D02; existing admin-desktop and adapter-desktop importers only"]
---

# Ticket T-16: Tauri 2 Desktop 安全宿主闭环

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/16-tauri-desktop-host.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-16.md</Path>`

## 1. 战略与来源

- **目标：** 用 Tauri 2 管理 Go sidecar 和本地 SQLite，交付安全登录、Demo CRUD、备份迁移与重启闭环。
- **可观察产出：** 首次启动在 ready 后开窗，sidecar 仅随机 loopback；Session 存 Stronghold，WebView JS 永不可见。
- **来源：** `US-004`、`US-019`、`AC-006`、`AC-007`、`AC-009`、`ADR-007`、`ADR-008`。
- **当前事实：** 现有 Desktop 是 Wails 路径，缺少 Tauri 2、Stronghold 和目标发行 triple 合同。
- **Planning Depth 原因：** 进程、凭据、数据库升级和原生宿主失败可导致本地数据或会话泄露。

## 2. 决策状态

### 已锁定决策

- Tauri 2 取得单实例锁，提供路径/一次性启动材料并监督 Go sidecar；sidecar 仅绑定随机 loopback。
- Session 由 Stronghold 保存并由 transport proxy 注入，JS/localStorage/URL/log 不可读取。
- Desktop 仅 SQLite，迁移前按策略备份，失败不打开主窗口。

### 已采用的低影响假设

- Sidecar readiness 使用一次性 nonce 绑定的本地握手和超时。

### 未决问题

无。

## 当前执行环境

- 用户最新决定把 native E2E 延后到全部 Ticket 实现集成后的统一系统 E2E。source `fdfa4afbe342edbd5b6feb00a8699780dd04caf4` 已闭合 release asset、Accessibility 查询、宿主退出竞态和窗口显示顺序问题；第 4 次 retained candidate/result `cd50bfdbd24517c8b3e8f3565d75fae8c3f6a22a` 已从头通过全部非 E2E Gate并进入 `main`。T-16 当前为 `implemented-pending-final-e2e`，可解除 T-17 实现依赖但不能标记 done。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| Tauri App、sidecar、Stronghold proxy、迁移/重启 tracer | kernel、IAM、Outbox、Demo、App Shell | Linux Desktop、远程 Server 模式、Wails 兼容 |

## 4. 要构建什么

用户启动安装后的 Desktop，Tauri 先锁实例、准备路径/备份/启动材料并监督 sidecar；sidecar 迁移 SQLite 和 ready 后才开窗，登录凭据由宿主代理，退出时 drain 并关闭资源。

## 5. 实现契约

- **入口或接缝：** Tauri lifecycle/commands、sidecar launcher/handshake、Stronghold transport proxy、Desktop adapter。
- **输入与输出：** Tauri 路径/nonce/session command；返回脱敏 runtime capability 和业务响应。
- **公共接口变化：** 新 Desktop runtime adapter 与 sidecar launch protocol。
- **不变量：** 单实例；随机 loopback；nonce 一次性；ready 前不开窗；JS 不见 token；退出按序 drain。
- **状态或数据流：** lock -> paths/backup -> launch -> migrate -> handshake/ready -> window -> proxy -> drain/stop。
- **错误与失败行为：** 非 loopback、材料缺失、握手超时、migration/asset 失败均不开窗并保留原库/备份。
- **兼容要求：** 不保留 Wails、固定端口或 Desktop PostgreSQL。
- **安全与隐私要求：** Stronghold、CSP、allowlist、sidecar origin、日志和进程参数泄露测试。

## 6. 执行路线

1. 建立首次启动、失败不开窗、JS token 不可见和重启 tracer。
2. 创建 Tauri 2 App、capabilities/CSP 和 target triple sidecar 配置。
3. 实现 Go sidecar launch protocol、SQLite backup/migration/readiness。
4. 实现 Stronghold transport proxy、Demo tracer 和有序 shutdown。
5. 已完成 macOS 开发环境的正式构建、静态安全和资源 Gate；macOS/Windows 原生行为登记到最终统一系统 E2E。

## 7. 路径访问契约

- **预计修改点：** Desktop App、经 `T16-D01` 开放的 Desktop adapter、sidecar、desktop platform、共享 sidecar packaging，以及经 `T16-D02` 精确开放的 Desktop 根 checks/importers。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** kernel、IAM、Demo 和 App Shell。
- **共享路径：** sidecar 打包布局由 T-16 唯一拥有，平台发行 Ticket 只消费；T-05 预建的 Desktop adapter manifest/source 现由 T-16 唯一实现。
- **批准偏差：** `T16-D01` 只开放 `<Path>go-admin-plus-ui/packages/adapters/desktop/**</Path>`，用于实现 ADR-011 已锁定且 T-05 明确保留给 T-16 的 runtime adapter。第一阶段根 package/Vitest/lock 继续只读，等待 T-09 result 后再由 Lead 精确开放 Desktop aggregate checks/importer；禁止修改 browser adapter、App Shell core 或产品 composition。
- **批准偏差：** `T16-D02` 在 T-09 result 后只开放根 `package.json` 的 Desktop aggregate checks、shell Vitest 的 Desktop specs，以及 lockfile 现有 `apps/admin-desktop` / `packages/adapters/desktop` 两个 importer。owner 必须先固定 stage-1 commit，再 rebase 到 amendment parent；其他 scripts、includes、importers、catalog、外部版本、共享 UI 与 composition 零漂移。
- **保留或不动：** macOS/Windows 安装器分别归 T-19/T-20。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | Desktop native tracer | `task test -- desktop-native` | 首启、登录、Demo CRUD、退出重启持久 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-16.md</Path>` |
| 失败路径 | hardening tracer | 非 loopback/nonce/asset/migration 失败与 JS 读取探针 | 不开窗、不泄露、保留数据库/备份 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-16.md</Path>` |
| 回归 | package/build suite | 构建目标 sidecar 并检查资产/triple | Tauri 正确打包且无 Wails 依赖 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-16.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 运行 Go/Rust/TS 非 E2E 检查。
- **E2E disposition：** deferred：真实 sidecar、Stronghold、迁移、窗口与重启场景保留到全部 Ticket 实现集成后的统一系统 E2E。
- **E2E owner/environment：** Lead / 最终系统候选中的可交互原生环境；逐 Ticket source-worktree 与 parent-candidate 不运行或声明 native E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、完整 Go/Rust/TS/release 非 E2E Gate、统一 native E2E 引用与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 新 Desktop 首次建库；新架构升级先备份再只前进迁移，成功才开窗。
- **兼容窗口：** 只支持上一新架构 fixture，不读取 Wails/旧数据库。
- **监控信号：** launch/handshake/migration/window/shutdown 阶段和 crash reason。
- **回滚或前向恢复：** migration 失败保留原库/备份；修复版本前向恢复，不自动覆盖原库。
- **不可逆操作与批准点：** 破坏性 migration 和备份清理需独立批准。
- **收缩条件：** T-21 证明 Wails、固定端口和 JS session 存储零命中。

## 10. 验收标准

- [ ] `AC-006/AC-007`：单实例、随机 sidecar、迁移/备份、开窗 Gate 和重启持久化成立。
- [ ] `AC-009`：Stronghold proxy 生效且 WebView/URL/log 不可读取 Session。
- [ ] `AC-021/AC-027/AC-036`：Desktop SQLite Demo、配置来源和 Shell 状态成立。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-16.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
