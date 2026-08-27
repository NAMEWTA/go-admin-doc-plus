# 配置

`schema/` 定义三个互斥的 JSON profile：

- `server-sqlite`：Server 监听、日志、会话和 SQLite 路径。
- `server-postgres`：Server 监听、日志和会话；连接材料只来自环境或 secret 文件。
- `desktop-sqlite`：由 Tauri host 提供绝对数据目录、日志目录、随机回环端口和一次性启动令牌。

加载顺序为内置默认值、JSON 文件、环境、允许的 CLI 覆盖。未知字段和跨 profile 字段直接失败。敏感连接材料不得写入 JSON、命令行或日志。

Server 可用 `GO_ADMIN_CONFIG_FILE` 指定非敏感 JSON 文件。SQLite 路径使用 `GO_ADMIN_SQLITE_PATH`，PostgreSQL 使用 `GO_ADMIN_DATABASE_DSN` 或 `GO_ADMIN_DATABASE_DSN_FILE`。
