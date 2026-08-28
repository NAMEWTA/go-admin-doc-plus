# 发行指南

发行候选必须绑定同一个精确 Git SHA、产品版本、OpenAPI 摘要和最高迁移版本。平台 workflow 只在受保护环境中取得签名材料，普通 PR CI 不接触发行凭据。

| 平台 | 候选 |
|---|---|
| Linux | `linux/amd64` 与 `linux/arm64` 的 Server/Web OCI 镜像及 Compose 定义 |
| macOS | Universal Tauri 2 App 和 DMG，Developer ID 签名、公证并通过 Gatekeeper 验证 |
| Windows | x64 Tauri 2 NSIS 安装包，应用、sidecar 和安装包均带时间戳 Authenticode 签名 |

本地预检只验证来源和策略，不签名、不公证、不部署：

```bash
task release VERSION=0.1.0
```

`task package TARGET=desktop` 只生成当前宿主的本地未签名构建：macOS 为 `.app`，Windows
为 NSIS。macOS Universal App/DMG、Windows Authenticode NSIS 及其安装证据只能由受保护
workflow 生成；本地 package 不能替代发行 Gate。

平台构建、安装验证和制品收集必须由 `.github/workflows/product-release.yml` 的受保护 jobs 执行。最终产品 manifest 只有在三个平台结果、摘要、SBOM、签名状态和来源完全一致时才可形成。

平台细节见 [Linux](../release/linux/README.md)、[macOS](../release/macos/README.md) 和 [Windows](../release/windows/README.md)。
