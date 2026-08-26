---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-11
title: 交付可校验自授权的 macOS ARM64 离线 DMG
status: done
planning_depth: deep
planning_depth_reason: 原生 CGO 构建、稳定应用身份、Gatekeeper 自授权、安装升级和用户数据恢复属于高风险发布链
ready: true
risk: high
blocked_by: [T-09]
contract_ids: [AC-007, AC-010, AC-016, AC-017]
owner: root
expected_changes: ["<Path>release/macos/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/build/darwin/**</Path>", "<Path>go-admin-plus/.github/workflows/release-macos.yml</Path>"]
writable_paths: ["<Path>release/macos/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/build/darwin/**</Path>", "<Path>go-admin-plus/.github/workflows/release-macos.yml</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/host/desktop/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-ui-plus/apps/admin/**</Path>"]
shared_paths: ["<Path>release/macos/**</Path>"]
shared_path_owners: ["<Path>release/macos/**</Path> => T-11"]
---

# Ticket T-11: 交付可校验自授权的 macOS ARM64 离线 DMG

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/11-macos-arm64-release.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-11.md</Path>`

## 1. 战略与来源

- **目标：** 在原生 Apple Silicon runner 构建可校验、ad-hoc 完整性签名但不具备 Developer ID/notarization 信任、可由设备 owner 对单个 App 授权运行的离线 DMG。
- **可观察产出：** 自有 Apple Silicon Mac 验证 SHA-256 后安装，通过系统“仍要打开”或只移除该 App quarantine 的 fallback 启动，登录/CRUD/上传/重启可用；旧 fixture 升级保留数据。
- **来源：** `AC-007`、`AC-010`、`AC-016`、`AC-017`、`USER-DECISION:2026-08-25 Phase 1 发布 unsigned self-use 包`、`RESEARCH:<Url>https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unknown-developer-mh40616/mac</Url>`。
- **当前事实：** T-09 提供安全桌面核心；最终 root `bf45f45` / backend `365ea9c` 和原生 run `32868421377` 已交付明确的 self-use App/DMG、安装授权说明、SBOM/checksum/provenance，并通过双次 tracer 与 scoped install gate。
- **Planning Depth 原因：** 平台发布身份影响升级路径；用户主动绕过平台信誉检查必须先做来源校验、限制授权范围并保留数据验证。

## 2. 决策状态

### 已锁定决策

- 只发布 `darwin/arm64` `.app` + `.dmg`；Phase 1 正式 artifact class 为 `unsigned-self-use`，使用 ad-hoc code signature 做 bundle 完整性校验，不宣称 Developer ID、notarization、staple 或 Gatekeeper trust。
- 产品 identity 从版本化 release manifest 读取；self-use 首次发布后 bundle ID、产品名和应用数据目录保持稳定，变更必须有迁移 Spec。
- 用户先核对发布的 SHA-256；首选首次启动失败后使用 macOS“系统设置 -> 隐私与安全性 -> 仍要打开”对单个 App 建立例外。fallback 只允许 `xattr -dr com.apple.quarantine "/Applications/Go Admin Plus.app"`，禁止 `spctl --master-disable` 或全局降低安全设置。
- 首个版本化 identity 为 `com.namewta.go-admin-plus`；`signed-production` 流程保留为未来可选 hardening，在 `identity_status=approved` 和凭据存在前继续 fail closed，但不阻塞本 Ticket。

### 已采用的低影响假设

- 最低运行系统遵循 Wails 2 Apple Silicon 支持线 macOS 11；更高最低版本必须由运行测试证据驱动。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| Info.plist/entitlements/icon、arm64 build、ad-hoc codesign、DMG、checksum/SBOM、自授权安装/升级 E2E | T-09 Desktop、Admin dist | Developer ID/notary/staple、App Store、Intel/universal、自动更新、受管设备策略绕过 |

## 4. 要构建什么

发布维护者以版本 tag 在 macOS runner 构建嵌入 Admin 和 SQLite 的 ARM64 App，应用 ad-hoc signature 并创建 DMG。artifact 同时提供 checksum、SBOM、provenance、trust-state 和安装说明。自有设备先验 SHA-256，再通过单应用例外安装；断网验证数据保留和核心流程。

## 5. 实现契约

- **入口或接缝：** release workflow/script、Wails darwin build metadata、原生安装 E2E。
- **输入与输出：** 固定源码版本与 manifest；输出明确标记 `unsigned-self-use` 的 DMG、app、checksum、SBOM、provenance、安装说明和验证记录。
- **公共接口变化：** 无业务接口；建立稳定 bundle/version identity。
- **不变量：** ARM64、无外部运行资源、ad-hoc signature 完整覆盖、用户数据不在 app bundle、trust state 不虚报。
- **状态或数据流：** build -> ad-hoc sign -> verify -> dmg -> checksum/SBOM -> user checksum verify -> per-app authorization -> offline install/upgrade -> publish metadata。
- **错误与失败行为：** 架构、ad-hoc 完整性、checksum、DMG、安装或 E2E 任一失败阻止发布；来源 checksum 不符时禁止提供绕过命令。
- **兼容要求：** 旧 AppData identity 和 SQLite fixture 升级。
- **安全与隐私要求：** 不需要签名密钥；日志脱敏；文档明确未经过 Apple malware/notary review，授权只能作用于已校验的单个 App。

## 6. 执行路线

1. 固定 manifest、Info.plist/entitlements 与 self-use trust state。
2. 在原生 runner 构建 ARM64 app 并验证架构和嵌入资源。
3. 应用并验证 ad-hoc signature，创建 DMG/checksum/SBOM/provenance。
4. 打包风险说明、SHA-256 验证、macOS Open Anyway 与单 App quarantine fallback；自动验证 fallback 不改全局策略。
5. 运行全新/升级/断网/迁移失败恢复 E2E，成功后形成 self-use artifact。

## 7. 路径访问契约

- **预计修改点/可写范围：** macOS release/build metadata 和专用 CI。实现预检确认 Wails 项目根位于 `cmd/go-admin-desktop`，其实际 metadata seam 是 `cmd/go-admin-desktop/build/darwin/**`，不是仓库根 `build/darwin/**`；因此在修改前将路径合同修正到实际 seam，不扩大到 DesktopHost/cmd 业务代码。
- **只读上下文：** DesktopHost/cmd/Admin。
- **共享路径：** macOS release 唯一 owner T-11；产品总 manifest 由 T-13 消费/汇合。
- **保留或不动：** 业务代码、Windows/Linux、真实证书和用户数据。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 原生 self-use 安装 | SHA-256、ad-hoc codesign、单 App 授权、离线 install/login/CRUD/restart | trust state 准确且数据保留 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-11.md</Path>` |
| 失败路径 | checksum/完整性/升级恢复 | checksum 错、ad-hoc signature 破坏、迁移失败 fixture | 发布阻断、备份可恢复且不提示绕过 | 同上 |
| 回归 | Web/API/Desktop suites | 原生 build + 受影响全套 | 无平台外回归 | 同上 |

- **Workspace checks：** native build、Go/pnpm tests、manifest schema、ad-hoc integrity 与 scoped quarantine 测试。
- **E2E disposition：** required；原生 DMG/self-use 授权、断网登录/CRUD/restart 和升级。
- **E2E owner/environment：** Lead / `current-workspace`；macOS ARM64 本机与原生 GitHub runner，全新临时用户数据目录与升级 fixture。
- **Integration evidence：** 提交、result、artifact SHA、trust state、checksum/SBOM、安装说明和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** build/ad-hoc sign/checksum 先于安装；升级先备份 DB。
- **兼容窗口：** 保持 bundle id/AppData/schema；旧 self-use 版本用于升级 fixture。
- **监控信号：** CI architecture/ad-hoc integrity/checksum/DMG、migration 和 E2E 状态。
- **回滚或前向恢复：** 安装包可回滚仅限 schema 兼容；数据使用备份或前向修复。
- **不可逆操作与批准点：** self-use 分发策略已由用户于 2026-08-25 批准；任何不可逆 migration 仍需人工批准；本 Ticket 不自动发布外部渠道。
- **收缩条件：** 未来启用 Developer ID/notarization 时保留 self-use trust-state 兼容并以新 release 偏差升级门禁。

## 10. 验收标准

- [x] `AC-007`、`AC-010`、`AC-016`、`AC-017` 通过。
- [x] SHA-256、ad-hoc codesign、scoped user authorization、断网全新/升级 Evidence 完整。
- [x] 路径、提交、集成和 required E2E 合同满足。
- [x] artifact 明确标识 `unsigned-self-use`，无全局安全关闭建议，无未批准外部 publish。
