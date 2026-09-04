# 发行资产

本目录按平台和生命周期管理产品发行：

- `linux/`：双架构 OCI、Compose、secret 与安装验证。
- `macos/`：Universal App、Developer ID 签名、公证、DMG 和安装验证。
- `windows/`：x64 sidecar、Tauri 2 NSIS、Authenticode 签名和安装验证。
- `shared/`：Desktop sidecar 的共享打包逻辑。
- `manifest/`：跨平台来源、摘要、SBOM 和签名状态聚合。

普通 CI 只运行可复现构建和策略测试。签名、公证、原生安装以及最终候选收集必须在受保护平台环境执行；任何缺失证据都使发行失败。

```bash
task release VERSION=0.1.0
```

该命令只执行本地预检，不上传制品、不触发远端 workflow。

本地发布候选还必须运行 `task release:verify` 与文档规定的 three-profile clean-room。该演练只用
disposable 数据；个人自用签名和公证为 `not-required`。正式跨平台分发仍由受保护 workflow
执行签名、公证和安装验证，不能用本地未签名构建替代。
