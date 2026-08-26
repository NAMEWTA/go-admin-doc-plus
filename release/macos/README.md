# macOS ARM64 发布

`identity.json` 是产品名、bundle ID、最低系统版本、应用数据目录和工件名的权威来源。
`self_use_release_status=approved` 允许生成 `unsigned-self-use` 工件；
`identity_status=candidate` 表示正式签名身份尚未批准。

`release-macos.yml` 固定根提交，构建带 tracer 的桌面应用并在同一数据目录运行两次，再构建
生产应用。自主使用模式执行 ad-hoc 签名，验证限定范围的 quarantine 流程，并把应用、
安装说明、来源记录、校验和与 SPDX SBOM 打包到 DMG。

未来正式签名需要以下 secrets：

- `APPLE_DEVELOPER_ID_P12_BASE64`
- `APPLE_DEVELOPER_ID_P12_PASSWORD`
- `APPLE_SIGNING_IDENTITY`
- `APPLE_NOTARY_KEY_P8_BASE64`
- `APPLE_NOTARY_KEY_ID`
- `APPLE_NOTARY_ISSUER_ID`

运行工件前必须阅读 `INSTALL.md`。workflow 不创建 GitHub Release，也不执行外部分发。
