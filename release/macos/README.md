# macOS Universal 发布

`identity.json` 是 bundle identity、最低系统版本、双架构、固定 Generator 工具链和供应链
证据的唯一权威。正式候选只允许 `signed-production`：Tauri host 与 Go sidecar 均为
`x86_64 + arm64`，所有嵌套 Mach-O 先签名，App 最后签名，然后创建、签名、公证并 staple
DMG。不存在 ad-hoc、自用、ARM64-only 或 Wails 发行模式。

App Resources 内含只读 tracked skeleton、离线 pnpm store、Go module cache，以及按宿主架构
选择的 Go/Node/pnpm 工具链，因此已安装应用的 Generator 不依赖用户开发环境。构建脚本只从
官方固定 URL 下载与 `identity.json` SHA-256 一致的 archives。

`.github/workflows/release-macos.yml` 绑定 exact source SHA 和受保护的
`macos-production` environment。它需要 Developer ID P12 与 App Store Connect API key
secrets，但不创建 GitHub Release、不上传远端 release asset，也不执行外部分发。Actions
artifact 只保存本次受保护 Gate 的本地候选证据。
