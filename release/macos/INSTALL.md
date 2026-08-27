# 在 macOS 安装 Go Admin Plus

只安装来自预期受保护 Actions run 的 Universal、Developer ID 签名且已公证候选。进入包含
`SHA256SUMS` 的目录后先验证全部文件：

```bash
shasum -a 256 -c SHA256SUMS
```

任意校验失败都必须停止。打开 DMG，将 `Go Admin Plus.app` 拖入 Applications；Gatekeeper
应直接接受，不需要移除 quarantine、关闭 Gatekeeper 或创建本地安全例外。受管理设备的组织
策略仍然有效。应用数据位于 app bundle 之外，替换应用不会删除现有 SQLite 数据和迁移备份。
