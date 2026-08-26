# 在 Windows 安装 Go Admin Plus

当前 Windows 10/11 x64 安装包用于自主使用，应用和安装器都没有 Authenticode 签名。
Windows 无法验证发布者，Microsoft Defender SmartScreen 可能显示保护提示。

只使用来自预期 Actions run 的工件。在包含 `SHA256SUMS` 的目录中计算安装器哈希：

```powershell
Get-FileHash -Algorithm SHA256 .\go-admin-plus-0.1.0-windows-amd64-unsigned-self-use-setup.exe
```

将完整结果与 `SHA256SUMS` 对应行比较，不一致时立即停止。独立验证后，如果 SmartScreen
提供自主使用例外，可以选择“更多信息”再选择“仍要运行”。安装器包含完整的 Microsoft
WebView2 Evergreen Standalone x64 runtime，不依赖在线 bootstrapper。

不得关闭 Microsoft Defender、SmartScreen 或 Smart App Control。受管理设备可能不允许
运行未签名安装器，当前包不会绕过该策略。升级或卸载应用时，
`%LOCALAPPDATA%\go-admin-plus` 中的数据会保留。
