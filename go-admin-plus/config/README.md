# 后端配置

`settings.yml` 是服务器 profile 的基础模板，`settings.sqlite.yml` 和
`settings.demo.yml` 提供特定运行方式的公开示例，`settings.full.yml` 展示完整字段。

本机开发使用 Git 忽略的 `settings.local.dev.yml`，生产环境使用由部署系统渲染且不提交
到仓库的配置。完整本地示例见仓库 [开发指南](../../docs/development.md)。

关键规则：

- 生产环境必须使用 `mode: prod` 并设置独立的 JWT secret。
- 数据库密码、对象存储密钥和第三方凭据不得写入已跟踪文件。
- 代码生成前端根为 `../go-admin-ui-plus/apps/admin/src`。
- SQLite 运行命令必须带 `-tags sqlite3`。
- 日志、上传、临时文件和数据库属于运行时状态，不写入源码目录。
