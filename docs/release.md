# 发行指南

发行候选必须绑定同一个精确 Git SHA、产品版本、OpenAPI 摘要和最高迁移版本。平台 workflow 只在受保护环境中取得签名材料，普通 PR CI 不接触发行凭据。

| 平台 | 候选 |
|---|---|
| Linux | `linux/amd64` 与 `linux/arm64` 的 Server/Web OCI 镜像及 Compose 定义 |
| macOS | Universal Tauri 2 App 和 DMG，Developer ID 签名、公证并通过 Gatekeeper 验证 |
| Windows | x64 Tauri 2 NSIS 安装包，应用、sidecar 和安装包均带时间戳 Authenticode 签名 |

本地预检只验证来源和策略，不签名、不公证、不部署：

```bash
task release VERSION=0.0.1
task release:verify
```

`task package TARGET=desktop` 只生成当前宿主的本地未签名构建：macOS 为 `.app`，Windows
为 NSIS。macOS Universal App/DMG、Windows Authenticode NSIS 及其安装证据只能由受保护
workflow 生成；本地 package 不能替代发行 Gate。

平台构建、安装验证和制品收集必须由 `.github/workflows/product-release.yml` 的受保护 jobs 执行。最终产品 manifest 只有在三个平台结果、摘要、SBOM、签名状态和来源完全一致时才可形成。

平台细节见 [Linux](../release/linux/README.md)、[macOS](../release/macos/README.md) 和 [Windows](../release/windows/README.md)。

## Three-profile clean-room

最终本地候选必须从 disposable 空状态依次验证 `server-sqlite`、`server-postgres` 与
`desktop-sqlite`。Server 两个 profile 都按 migrate、bootstrap、Doctor、API/worker、Web
登录、核心管理和重启顺序执行；PostgreSQL 的 migrate 独占且先于 API/worker。Desktop 通过
原生首次设置、SQLite/Stronghold 重启、登录、权限、CRUD 与清理。任何缺环境、skip、not-run、
旧命令、固定凭据或缺少精确 marker 都使候选失败。

本地个人使用的 clean-room 验证不发布制品，因此签名和公证记录为 `not-required`，不能写成
passed。受保护的正式平台 workflow 仍要求 macOS Developer ID/公证和 Windows Authenticode；
这些发布动作不属于本地候选，也不会由 `task release:verify` 触发。

最终证据同时复用 required PostgreSQL、安全供应链、真实 Web 与真实 macOS native Gate，并在
当前 candidate 重新运行治理、架构、零兼容、文档、生成、Go、pnpm、Rust 和 release policy
检查。`task release VERSION=0.0.1` 仅执行本地 preflight，不 push、publish、deploy 或迁移生产。
