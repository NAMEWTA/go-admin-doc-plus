# 发行资产

本目录按平台和生命周期管理产品发行：

- `linux/`：双架构 OCI、Compose、secret 与安装验证。
- `macos/`：Universal App、Developer ID 签名、公证、DMG 和安装验证。
- `windows/`：x64 sidecar、Tauri 2 NSIS、Authenticode 签名和安装验证。
- `shared/`：Desktop sidecar 与生成器运行时的共享打包逻辑。
- `manifest/`：跨平台来源、摘要、SBOM 和签名状态聚合。

普通 CI 只运行可复现构建和策略测试。签名、公证、原生安装以及最终候选收集必须在受保护平台环境执行；任何缺失证据都使发行失败。

```bash
task release VERSION=0.1.0
```

该命令只执行本地预检，不上传制品、不触发远端 workflow。
