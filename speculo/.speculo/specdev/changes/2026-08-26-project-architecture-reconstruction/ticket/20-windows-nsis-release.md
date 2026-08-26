---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-20
title: Windows x64 NSIS 发行切片
status: ready
planning_depth: deep
planning_depth_reason: Windows 原生签名、安装、首次启动、重启和卸载边界需要受保护发行 Gate
ready: true
risk: critical
blocked_by: [T-17]
contract_ids: [AC-032, AC-033]
owner: unassigned
expected_changes: ["<Path>release/windows/**</Path>", "<Path>scripts/release/windows/**</Path>", "<Path>.github/workflows/release-windows.yml</Path>"]
writable_paths: ["<Path>release/windows/**</Path>", "<Path>scripts/release/windows/**</Path>", "<Path>.github/workflows/release-windows.yml</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/**</Path>", "<Path>release/shared/sidecar/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-20: Windows x64 NSIS 发行切片

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/20-windows-nsis-release.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-20.md</Path>`

## 1. 战略与来源

- **目标：** 生成、签名并原生安装验证 Windows x64 NSIS。
- **可观察产出：** 安装器身份正确，首次启动、Demo CRUD、重启和卸载边界通过并关联供应链证据。
- **来源：** `US-016`、`AC-032`、`AC-033`、`ADR-010`、`ADR-022`。
- **当前事实：** 旧 Windows release 指向 Wails/旧身份，缺少安装、卸载和 provenance Gate。
- **Planning Depth 原因：** 原生安装、签名、本地数据保留和卸载错误影响用户系统与供应链。

## 2. 决策状态

### 已锁定决策

- 仅 Windows x64 NSIS；生产制品在受保护 runner 签名。
- 卸载默认不静默删除用户业务数据，清理边界必须明确验证。

### 已采用的低影响假设

- 原生 smoke 使用隔离用户 profile 并检查残留白名单。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| NSIS、签名、安装/重启/卸载 smoke、SBOM/provenance、workflow | T-16 Desktop、T-17 产品 | MSI、ARM64、Store、远端发布 |

## 4. 要构建什么

发行工程师在受保护 Windows runner 构建并签名 NSIS，原生安装后完成首次启动、登录、业务与重启，随后验证卸载仅删除产品文件并遵守用户数据边界。

## 5. 实现契约

- **入口或接缝：** root package/release、Tauri NSIS、签名、install/uninstall tracer。
- **输入与输出：** 源码/version/受保护凭据；NSIS、checksum、SBOM、provenance 和原生证据。
- **公共接口变化：** 新 Windows 发行合同。
- **不变量：** x64 identity 稳定；App/sidecar 签名有效；卸载边界明确；凭据不泄露。
- **状态或数据流：** build -> sign -> NSIS -> install -> first-run/restart -> uninstall boundary -> evidence。
- **错误与失败行为：** 签名、安装、启动、持久化或卸载越界失败均阻断。
- **兼容要求：** 不保留 Wails identity、旧安装器或旧注册表合同。
- **安全与隐私要求：** 最小权限安装、签名验证、路径引用安全和凭据隔离。

## 6. 执行路线

1. 固定 identity、x64、安装/卸载和证据 policy。
2. 实现 Desktop/sidecar 构建、签名和 NSIS 配置。
3. 生成 checksum、SBOM、provenance。
4. 实现隔离安装、首次启动、重启和卸载 tracer。
5. 在受保护 runner 完成 Gate。

## 7. 路径访问契约

- **预计修改点：** Windows 专属 release/script/workflow。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** Tauri App 和 sidecar layout。
- **共享路径：** 无；根 CI 归 T-21。
- **保留或不动：** macOS/Linux 资产。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | protected native gate | `task package -- windows && task release:verify -- windows` | 签名 NSIS 安装/启动/重启/卸载通过 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-20.md</Path>` |
| 失败路径 | policy fixture | 缺签名/asset、安装失败或卸载越界 | 阻断且用户数据不被误删 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-20.md</Path>` |
| 回归 | install/restart/uninstall | CRUD、重启、卸载并检查残留边界 | 数据持久且清理符合合同 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-20.md</Path>` |

- **Workspace checks：** current-workspace 或 source-worktree 运行 Rust/Go/TS/build 和 policy 检查。
- **E2E disposition：** required：受保护 Windows 原生安装、运行、重启和卸载必须实际执行。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不声明通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、原生 Gate 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-17 后构建，签名后安装/重启/卸载验证。
- **兼容窗口：** 只支持上一新架构 Desktop 数据 fixture。
- **监控信号：** signature、install、first-run/restart、uninstall boundary、SBOM/provenance。
- **回滚或前向恢复：** 未发布候选废弃；迁移失败保留备份并由修复版本恢复。
- **不可逆操作与批准点：** 生产凭据和远端发布需额外授权；用户数据删除不属于默认卸载。
- **收缩条件：** T-21 汇总 Gate 并证明旧 Windows/Wails 资产零引用。

## 10. 验收标准

- [ ] `AC-032`：签名 NSIS、安装、首次启动、重启和卸载边界证据完整。
- [ ] `AC-033`：Windows required Gate 失败阻断候选。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-20.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] 未执行远端发布，E2E disposition 已执行。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
