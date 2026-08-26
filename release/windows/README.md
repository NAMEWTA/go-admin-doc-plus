# Windows AMD64 自主使用发布

`identity.json` 是 Windows 产品名、公司名、可执行文件、安装器和 WebView2 payload 身份的
权威来源。当前只允许 `unsigned-self-use` 发布类型，不得把应用或 NSIS 安装器描述为受
Microsoft 信任的工件。

安装器嵌入完整的 Microsoft WebView2 Evergreen Standalone x64 payload。workflow 在打包前
验证固定字节数、SHA-256 和 Microsoft Authenticode 签名；Wails 应用使用 `error` WebView2
策略，不回退到在线安装器。

用户级安装器在安装、升级和卸载时保留 `%LOCALAPPDATA%\go-admin-plus`。运行工件前必须
阅读 `INSTALL.md`。workflow 不创建 GitHub Release，也不向外部渠道发布。
