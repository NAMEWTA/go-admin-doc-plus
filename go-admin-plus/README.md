# Go 后端

本模块提供四个正式命令：

| 命令 | 用途 |
|---|---|
| `cmd/go-admin-plus` | SQLite 或 PostgreSQL Server |
| `cmd/desktop-sidecar` | Tauri 2 管理的本地 SQLite 服务 |
| `cmd/config-check` | profile JSON 与运行材料预检 |
| `cmd/migrate` | SQLite 或 PostgreSQL 向前迁移 |

从仓库根使用 Taskfile：

```bash
task dev TARGET=server PROFILE=server-sqlite
task migrate PROFILE=server-sqlite
task test
task lint
```

业务模块位于 `internal/modules/`，产品唯一组合根位于 `internal/app/product/`。配置格式见 [配置说明](config/README.md)，数据库策略见 [数据库文档](../database/README.md)。
