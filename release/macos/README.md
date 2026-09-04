# macOS Universal 发布

`identity.json` 是 bundle identity、最低系统版本、双架构和供应链证据的唯一权威。
正式候选只允许 `signed-production`：Tauri host 与 Go sidecar 均为
`x86_64 + arm64`，所有嵌套 Mach-O 先签名，App 最后签名，然后创建、签名、公证并 staple
DMG。不存在 ad-hoc、自用或 ARM64-only 发行模式。

Desktop 运行时只包含签名的 Tauri host、Go sidecar 和本地产品资源，不依赖用户开发环境或远程
数据库。构建脚本仍通过固定来源和摘要校验生成 sidecar。

`.github/workflows/release-macos.yml` 绑定 exact source SHA 和受保护的
`macos-production` environment。它需要 Developer ID P12 与 App Store Connect API key
secrets，但不创建 GitHub Release、不上传远端 release asset，也不执行外部分发。Actions
artifact 只保存本次受保护 Gate 的本地候选证据。
