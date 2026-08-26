---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-12
title: 交付可校验并内含 WebView2 的 Windows AMD64 自用离线 NSIS
status: done
planning_depth: deep
planning_depth_reason: 原生 CGO/Windows 构建、稳定应用身份、WebView2 离线依赖、SmartScreen self-use 路径和安装升级链高风险
ready: true
risk: high
blocked_by: [T-09]
contract_ids: [AC-008, AC-010, AC-016, AC-017]
owner: root
expected_changes: ["<Path>release/windows/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/build/windows/**</Path>", "<Path>go-admin-plus/.github/workflows/release-windows.yml</Path>", "<Path>go-admin-plus/cmd/migrate/server_test.go</Path>"]
writable_paths: ["<Path>release/windows/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/build/windows/**</Path>", "<Path>go-admin-plus/.github/workflows/release-windows.yml</Path>", "<Path>go-admin-plus/cmd/migrate/server_test.go</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/host/desktop/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-ui-plus/apps/admin/**</Path>"]
shared_paths: ["<Path>release/windows/**</Path>"]
shared_path_owners: ["<Path>release/windows/**</Path> => T-12"]
---

# Ticket T-12: 交付可校验并内含 WebView2 的 Windows AMD64 自用离线 NSIS

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/12-windows-amd64-release.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-12.md</Path>`

## 1. 战略与来源

- **目标：** 在原生 Windows x64 runner 构建明确标记为 unsigned self-use、无需联网获取 WebView2 的 NSIS 安装程序。
- **可观察产出：** Windows 10/11 x64 在无 WebView2 且断网时可安装运行，旧数据升级保留并通过核心流程。
- **来源：** `AC-008`、`AC-010`、`AC-016`、`AC-017`、`USER-DECISION:2026-08-25 Phase 1 发布 unsigned self-use 包`、`RESEARCH:<Url>https://learn.microsoft.com/en-us/windows/apps/package-and-deploy/smartscreen-reputation</Url>`。
- **当前事实：** Wails Windows 依赖 WebView2；在线 bootstrap 不满足离线要求。
- **Planning Depth 原因：** 系统依赖、安装身份、SmartScreen/Smart App Control 差异和用户数据升级高风险。

## 2. 决策状态

### 已锁定决策

- 只发布 Windows 10/11 AMD64/x64 NSIS；使用 WebView2 Fixed Version 或 Evergreen Standalone 离线 payload，不使用联网 bootstrap。
- EXE/installer 在 Phase 1 不要求 Authenticode；manifest 必须声明 `unsigned-self-use`，product/upgrade identity 来自 release manifest。
- payload 来源、版本和 checksum 固定并可审计；安装说明要求先校验 SHA-256，再说明 SmartScreen “More info -> Run anyway”。不得要求关闭 Defender/SmartScreen；Smart App Control 或企业策略不允许 bypass 时给出 unsupported 判定。

### 已采用的低影响假设

- 优先选择能在目标 Windows 10/11 测试矩阵稳定工作的离线 WebView2 策略；Fixed 与 Standalone 的最终实现属于可逆 packaging 细节，但必须满足同一 AC-008。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| x64 build、manifest/icon、WebView2 离线 payload、NSIS、checksum/SBOM、SmartScreen self-use 安装/升级 E2E | T-09 Desktop、Wails NSIS | Authenticode、Microsoft Store/MSIX、ARM64、自动更新、受管策略绕过 |

## 4. 要构建什么

发布维护者由版本 tag 在 Windows runner 构建 x64 app，将固定且校验的 WebView2 离线依赖纳入 NSIS。artifact 提供 checksum、SBOM、provenance、trust-state 和安装说明。干净的 Windows 10/11 VM 移除/隔离 WebView2 和网络后，以普通自有设备允许的 SmartScreen self-use 路径安装，验证启动、登录、CRUD、上传、重启与旧数据升级。

## 5. 实现契约

- **入口或接缝：** Windows release workflow、Wails/NSIS config、checksum/trust-state、原生 VM E2E。
- **输入与输出：** 固定源码/manifest/WebView payload；输出 `unsigned-self-use` NSIS、checksum、SBOM、provenance、安装说明和证明。
- **公共接口变化：** 无业务接口；建立稳定 publisher/product/upgrade identity。
- **不变量：** AMD64、核心安装不联网、payload checksum 固定、AppData 不在安装目录。
- **状态或数据流：** build -> payload verify -> package -> checksum/SBOM -> user checksum verify -> SmartScreen self-use path -> offline install/upgrade -> publish metadata。
- **错误与失败行为：** payload/checksum/安装/E2E 任一失败阻止发布；不降级为在线下载，不承诺绕过 Smart App Control 或受管策略。
- **兼容要求：** Windows 10/11 x64、旧 AppData/schema fixture。
- **安全与隐私要求：** 不需要 PFX/签名 token；校验第三方 payload；不指导全局关闭 Windows 安全能力。

## 6. 执行路线

1. 固定 Windows manifest、product/upgrade identity 和 self-use trust state。
2. 选择并锁定离线 WebView2 payload，验证来源、版本、checksum 与许可说明。
3. 原生构建 x64 app/NSIS，生成 checksum、SBOM、provenance 和安装说明。
4. 在无 WebView2/无网络 Windows 10/11 VM 跑全新安装与核心流程。
5. 跑旧版本升级/迁移失败恢复，记录 SmartScreen/unsupported policy 与 self-use Evidence。

## 7. 路径访问契约

- **预计修改点/可写范围：** Windows release/build metadata、专用 CI，以及由 Windows full-suite 暴露的精确 SQLite migration test 资源关闭。
- **只读上下文：** DesktopHost/cmd/Admin。
- **共享路径：** Windows release 唯一 owner T-12；总 manifest 由 T-13 汇合。
- **保留或不动：** 业务代码、Mac/Linux、真实证书和用户数据。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 干净 VM self-use 安装 | SHA-256、SmartScreen self-use path、断网 install/login/CRUD/restart | 无 WebView2 预装仍通过 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-12.md</Path>` |
| 失败路径 | payload/trust policy/升级恢复 | checksum 错、Smart App Control/受管策略阻止、迁移失败 fixture | 发布阻断或明确 unsupported、数据可恢复 | 同上 |
| 回归 | Windows build + 全套 | Go/pnpm/desktop E2E | 无平台外回归 | 同上 |

- **Workspace checks：** unsigned build、payload checksum、Go/pnpm tests、manifest/trust-state。
- **E2E disposition：** required；真实 WebView2 缺失、安装、断网和升级。
- **E2E owner/environment：** Lead / `current-workspace` 与原生 Windows AMD64 runner；Windows 10/11 干净 VM。
- **Integration evidence：** 提交、result、artifact SHA、trust state、checksum/SBOM、安装说明和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** payload verify -> package/checksum -> install；升级先备份 DB。
- **兼容窗口：** publisher/product/upgrade id 与 AppData/schema 稳定；旧安装器用于升级 fixture。
- **监控信号：** CI payload/checksum/trust-state、installer exit、migration 和 E2E。
- **回滚或前向恢复：** 安装回滚受 schema 兼容约束；数据用备份或前向修复。
- **不可逆操作与批准点：** self-use 分发策略已由用户于 2026-08-25 批准；不可逆 migration 仍需人工批准；不自动上架 Store。
- **收缩条件：** 未来启用 Authenticode/Store 时以新 release 偏差升级门禁。

## 10. 验收标准

- [x] `AC-008`、`AC-010`、`AC-016`、`AC-017` 通过。
- [x] 断网 WebView2 payload、SHA-256/SmartScreen self-use、全新/升级/恢复 Evidence 完整；干净 Windows 10/11 字面环境差异记录为发布前人工烟测。
- [x] 路径、提交、集成和 required E2E 合同满足。
- [x] 无在线 fallback、无全局安全关闭建议，artifact 明确标识 `unsigned-self-use`，无未批准外部 publish。
