---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-19
title: macOS Universal DMG 发行切片
status: in_progress
planning_depth: deep
planning_depth_reason: Universal 原生制品、签名、公证、安装和本地数据重启需要受保护发行 Gate
ready: true
risk: critical
blocked_by: [T-17]
contract_ids: [AC-031, AC-033]
owner: codex-root
expected_changes: ["<Path>release/macos/**</Path>", "<Path>scripts/release/macos/**</Path>", "<Path>.github/workflows/release-macos.yml</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus/internal/modules/generator/compile_gate.go</Path>", "<Path>go-admin-plus/internal/modules/generator/compile_gate_test.go</Path>"]
writable_paths: ["<Path>release/macos/**</Path>", "<Path>scripts/release/macos/**</Path>", "<Path>.github/workflows/release-macos.yml</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus/internal/modules/generator/compile_gate.go</Path>", "<Path>go-admin-plus/internal/modules/generator/compile_gate_test.go</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/**</Path>", "<Path>release/shared/sidecar/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus/internal/modules/generator/compile_gate.go</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path> => T-19 under T19-D01; select the packaged Generator repository/toolchain without changing product transport or lifecycle", "<Path>go-admin-plus/internal/modules/generator/compile_gate.go</Path> => T-19 under T19-D01; preserve explicit offline Go policy in the child-process environment"]
---

# Ticket T-19: macOS Universal DMG 发行切片

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/19-macos-universal-release.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-19.md</Path>`

## 1. 战略与来源

- **目标：** 生成、签名、公证并原生安装验证 macOS Universal DMG。
- **可观察产出：** DMG 包含目标架构，首次启动、登录、Demo CRUD 与重启持久化通过并关联源码身份。
- **来源：** `US-016`、`AC-031`、`AC-033`、`ADR-010`、`ADR-022`。
- **当前事实：** 旧 macOS release 仍指向 Wails/非目标架构并缺供应链闭环。
- **Planning Depth 原因：** 签名、公证、原生安装与凭据使用均为关键发行风险。

## 2. 决策状态

### 已锁定决策

- 仅受保护 macOS runner 使用生产签名/公证凭据；制品为 Universal DMG。
- 未签名本地构建只用于开发，不可标记 publishable。
- `T19-D01`：正式 App 必须携带 Generator 的只读 tracked skeleton、离线依赖与双架构 Go/Node/pnpm 工具链；Tauri 只向 sidecar 传递精确白名单环境并把其工作目录固定到发布骨架。Generator 子进程保留并强制显式离线 Go policy。开发态继续使用当前源码根和本机工具链，但不继承无关环境变量。

### 已采用的低影响假设

- 安装 smoke 使用隔离用户数据目录并记录清理边界。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| DMG、签名、公证、安装 smoke、SBOM/provenance、workflow | T-16 Desktop、T-17 产品 | App Store、远端发布、Linux Desktop |

## 4. 要构建什么

发行工程师在受保护 runner 构建 Universal App/sidecar，签名并公证 DMG；原生安装后完成首次启动、业务和重启 smoke，全部证据齐全才成为 publishable。

## 5. 实现契约

- **入口或接缝：** root package/release 委派、Tauri bundle、codesign/notary、installer smoke。
- **输入与输出：** 源码/version/受保护凭据；DMG、checksum、SBOM、provenance、签名/公证证据。
- **公共接口变化：** 新 macOS 发行合同。
- **不变量：** Universal 架构齐全；App/sidecar 嵌套签名有效；凭据不进入日志/制品。
- **状态或数据流：** build -> sign -> package -> notarize/staple -> install -> first-run/restart -> evidence。
- **错误与失败行为：** 任一签名、公证、安装或 smoke 失败阻断 publishable。
- **兼容要求：** 不保留 Wails bundle/identity 或旧安装路径。
- **安全与隐私要求：** hardened runtime、最小 entitlement、凭据 secret store、制品扫描。

## 6. 执行路线

1. 固定 bundle identity、架构、entitlement 和证据 policy。
2. 实现 Universal Desktop/sidecar 构建和签名顺序。
3. 实现 DMG、公证、staple、checksum/SBOM/provenance。
4. 实现隔离安装、首次启动和重启 smoke。
5. 在受保护 runner 完成 Gate。

## 7. 路径访问契约

- **预计修改点：** macOS 专属 release/script/workflow，以及 `T19-D01` 精确开放的 Tauri sidecar 启动配置。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** Tauri App 和 sidecar layout。
- **共享路径：** Tauri `main.rs` 仅由 T-19 修改 Generator 资源/工具链选择，Generator compile gate 仅收紧离线子进程环境；根 CI 归 T-21。
- **保留或不动：** Windows/Linux 资产。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | protected native gate | `task package -- macos && task release:verify -- macos` | Universal DMG 签名/公证/安装/smoke 通过 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-19.md</Path>` |
| 失败路径 | policy fixture | 缺架构、签名、公证、asset 或凭据泄露 | 阻断且不产生 publishable | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-19.md</Path>` |
| 回归 | install/restart tracer | 安装、CRUD、退出重启 | 数据保留且 identity/路径正确 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-19.md</Path>` |

- **Workspace checks：** current-workspace 或 source-worktree 运行 Rust/Go/TS/build 和 policy 检查。
- **E2E disposition：** deferred：macOS 原生安装、启动和重启场景保留到全部 Ticket 实现集成后的统一系统 E2E；正式签名、公证和凭据使用仍需独立授权。
- **E2E owner/environment：** Lead / 最终系统候选的受保护 macOS 环境；逐 Ticket source-worktree 与 parent-candidate 不运行或声明 E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、本地非 E2E Gate、统一平台 E2E 引用与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-17 后构建，签名/公证后安装验证。
- **兼容窗口：** 只支持上一新架构 Desktop 数据 fixture。
- **监控信号：** architectures、codesign、notary、install、first-run/restart、SBOM/provenance。
- **回滚或前向恢复：** 未发布候选废弃；迁移失败保留本地备份并由修复版本恢复。
- **不可逆操作与批准点：** 生产凭据使用和远端发布需受保护环境及额外授权。
- **收缩条件：** T-21 汇总 Gate 并证明旧 macOS/Wails 资产零引用。

## 10. 验收标准

- [ ] `AC-031`：Universal DMG 签名、公证、安装、首次启动和重启证据完整。
- [ ] `AC-033`：macOS required Gate 失败阻断候选。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-19.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] 未执行远端发布或正式签名，非 E2E 实现 Gate 已执行且平台场景已登记到最终统一矩阵。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
