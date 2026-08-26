# 在 macOS 安装 Go Admin Plus

当前 ARM64 包用于自主使用，只有 ad-hoc 完整性签名，没有 Apple Developer ID 签名，也未
经过 Apple notarization。macOS 无法验证发布者或证明 Apple 已检查该应用。

只使用来自预期 Actions run 的工件。进入包含 `SHA256SUMS` 的目录后先验证全部文件：

```bash
shasum -a 256 -c SHA256SUMS
```

任意校验失败都必须停止。打开 DMG，将 `Go Admin Plus.app` 拖入 Applications 并尝试启动。
若系统阻止，在“系统设置 > 隐私与安全性”中仅为 Go Admin Plus 选择“仍要打开”。

只有独立核对全部校验和后，才可用以下命令移除这个应用的 quarantine：

```bash
xattr -dr com.apple.quarantine "/Applications/Go Admin Plus.app"
```

不得全局关闭 Gatekeeper。受管理设备可能禁止本地例外；当前包不会绕过组织策略。应用数据
位于 app bundle 之外，替换应用不会删除现有数据和迁移备份。
