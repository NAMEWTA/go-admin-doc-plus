---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-10
title: 实现 Desktop 原生首次设置与 Session 转换
status: ready
planning_depth: deep
planning_depth_reason: 跨 Tauri/Rust/Go/WebView 安全边界处理最高权限凭据、首次 Session 和不可重复 Bootstrap
ready: true
risk: critical
blocked_by: [T-02, T-09, T-13]
contract_ids: [AC-004, AC-005, AC-031]
owner: unassigned
expected_changes: ["<Path>go-admin-plus/internal/host/desktop/setup*</Path>", "<Path>go-admin-plus/internal/app/product/desktop_setup*</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/first_setup*</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/vault.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/first-setup/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/host/desktop/setup*</Path>", "<Path>go-admin-plus/internal/app/product/desktop_setup*</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/first_setup*</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/vault.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/first-setup/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/bootstrap/**</Path>", "<Path>go-admin-plus/internal/host/desktop/host.go</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path> => T-10"]
---

# Ticket T-10: 实现 Desktop 原生首次设置与 Session 转换

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/10-desktop-first-setup.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-10.md</Path>`

## 1. 战略与来源

- **目标：** 让 Desktop SQLite 空库在原生受控流程建立管理员和首个 Session，避免 CLI 与 Web 初始化后门。
- **可观察产出：** 首次设置成功直接进入工作区；Session 建立失败保留账号并转普通登录；重启不重复 Bootstrap。
- **来源：** `US-002`、`AC-004~005`、`AC-031`、`ADR-001`、`ADR-008`、`LOG-020`。
- **当前事实：** Desktop 已有 Tauri 2 vault/control boundary；`<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>` 有用户未提交修改，必须逐行融合。
- **Planning Depth 原因：** 原生/WebView secret 隔离和一次性初始化失败处理不可通过普通页面模拟。

## 2. 决策状态

### 已锁定决策

- Desktop 仅 SQLite，备份成功后自动迁移，sidecar 一体运行 API/worker。
- Bootstrap 和首 Session 由原生宿主编排；密码和原始 Session 不进入 WebView 持久状态。
- 账号已提交但 Session/工作区失败时不回滚，切到普通登录。

### 已采用的低影响假设

- 首次设置 UI 使用 T-11 token/组件可逐步消费；原生命令只暴露最小状态与提交动作。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| setup detection、Tauri command、Go adapter、vault Session、恢复 UI | T-02 Bootstrap、Desktop loopback/vault、T-09 runtime | Web Bootstrap、远程 API、完整 native E2E（T-20） |

## 4. 要构建什么

Desktop 启动并完成备份/migration 后，原生宿主判断是否为空库。为空时只显示首次设置入口，受控收集凭据并调用共享 Bootstrap；成功后建立并保存 Session，再通知 WebView进入工作区。部分成功显示明确可恢复状态并切换登录，任何重启都根据数据库事实而不是客户端标记判断 setup 状态。

## 5. 实现契约

- **入口或接缝：** Tauri first-setup commands、Go Desktop private adapter、vault。
- **输入与输出：** 用户名/密码只跨原生 command；WebView 仅得 setup state、非敏感错误和认证完成信号。
- **公共接口变化：** 新增最小 Tauri capability，不新增公开 HTTP 初始化端点。
- **不变量：** 数据库非空永不再 Bootstrap；原始 Session/密码不进 localStorage、DOM、日志。
- **状态或数据流：** empty -> native form -> Bootstrap commit -> Session/vault -> workspace；Session failure -> login。
- **错误与失败行为：** backup/migration/Bootstrap/Session/workspace 分层反馈；已提交账号不回滚。
- **兼容要求：** 无旧 Desktop setup 模式。
- **安全与隐私要求：** capability allowlist 最小化，secret 缓冲及时清理，现有 main.rs 修改保留融合。

## 6. 执行路线

1. 记录 main.rs 用户 diff，建立 setup/部分成功/重启红灯 host 与 Rust tests。
2. 实现 Go setup adapter 和原生 setup-state/submit command。
3. 接入 Bootstrap、Session 创建和 vault，限制 capability。
4. 实现首次设置/错误/登录转换 UI。
5. 运行 Go/Rust/TS、secret scan、production capability 与构建检查。

## 7. 路径访问契约

- **预计修改点/可写范围：** Desktop setup 专属 Go/Rust/UI 文件和 main.rs/vault 指定文件。
- **只读上下文：** T-02 use case、host、共享 UI、T-20 E2E。
- **共享路径：** main.rs 仅 T-10 可写；必须融合而非覆盖用户现有修改。
- **保留或不动：** Desktop E2E runner、签名/公证配置不在本 Ticket。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | Go host/Rust command | 空库 setup + Session/vault integration tests | 直接进入工作区且 secret 隔离 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-10.md</Path>` |
| 失败路径 | 部分成功/重启 | 注入 Session/workspace failure、重启 | 保留账号，进入登录，不重复 setup | 同上 |
| 回归 | Desktop build | Go Desktop tests、cargo test/clippy、Tauri no-bundle build | 宿主与能力合同通过 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行 Go/Rust/TS、clippy、build、secret/capability checks。
- **E2E disposition：** not-required：本 Ticket 先完成编码和 native integration；真实 macOS 首次设置 E2E 由 T-20 单一 Gate 集中执行。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；T-20 在 parent-candidate/current-workspace 执行 native E2E。
- **Integration evidence：** 用户 diff 基线、commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-02/T-09/T-13 -> setup adapter -> Tauri capability/UI -> T-20 native Gate。
- **兼容窗口：** 无公开 setup 双轨；数据库事实是唯一状态。
- **监控信号：** 本地脱敏 setup stage/result、vault failure、backup/migration status。
- **回滚或前向恢复：** Bootstrap 前可安全退出；提交后通过普通登录前向恢复；数据库迁移失败恢复备份。
- **不可逆操作与批准点：** 首管理员提交需用户明确 submit；main.rs 修改前需保留并核对现有 diff。
- **收缩条件：** 固定凭据、Web setup API、客户端 setup flag 和 secret 持久化扫描为零。

## 10. 验收标准

- [ ] `AC-004~005、AC-031` 的 Desktop 集成行为成立。
- [ ] main.rs 现有用户修改被完整融合，WebView 无密码/原始 Session。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-10.md</Path>`。
- [ ] shared owner、commit、integration/result 和 T-20 E2E 归属完整。
