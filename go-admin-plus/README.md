# Go Admin Plus 后端

本目录包含 Go Admin Plus 的 API、业务模块、数据库迁移、共享应用内核和 Wails 桌面宿主。

## 目录

| 路径 | 职责 |
| --- | --- |
| `app/` | 管理、任务、演示及其他业务模块 |
| `common/` | Action、DTO、Model、中间件和存储公共能力 |
| `internal/` | 应用内核、运行 profile、平台与宿主适配 |
| `cmd/` | API、迁移、桌面和 CLI 入口 |
| `api/openapi/` | 规范化 OpenAPI 契约 |
| `docs/admin/` | Swagger 生成物 |
| `config/` | 可提交的配置模板与初始化 SQL |

## 常用命令

```bash
go test ./...
go build ./...
go generate ./...
```

本地 SQLite：

```bash
go run -tags sqlite3 . migrate -c config/settings.local.dev.yml
go run -tags sqlite3 . server -c config/settings.local.dev.yml
```

SQLite 构建必须带 `-tags sqlite3`。项目迁移统一提交到
`cmd/migrate/migration/version/`；已执行的迁移不可改写。

## 开发入口

- [仓库开发指南](../docs/development.md)
- [后端架构](docs/architecture.md)
- [后端强约束](AGENTS.md)
- 标准单表 CRUD：`app/demo/`

项目许可证见仓库根 [LICENSE](../LICENSE)。
