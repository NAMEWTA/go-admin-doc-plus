# 仓库架构

Go Admin Plus 是单一 Git 仓库、单一 `main` 分支和单一产品版本的模块化单体。根提交 SHA
同时标识后端、前端、桌面宿主和部署契约，不存在子模块提交漂移。

## 运行形态

```text
Vue 3 Admin Web
       |
       | HTTP / OpenAPI contract
       v
Go application kernel
       |
       +-- Server host  -> PostgreSQL/MySQL/SQLite + Redis/Memory
       +-- Desktop host -> SQLite + Memory + embedded Web assets

Linux delivery   -> Nginx + API + migration + PostgreSQL + Redis
macOS / Windows  -> Wails native application
```

后端业务模块位于 `go-admin-plus/app/`，共享运行时装配位于 `internal/`。Web 管理端以
`go-admin-ui-plus/apps/admin` 为宿主，领域页面位于 `domains/`，跨领域能力位于
`packages/`。

## 契约边界

- `go-admin-plus/api/openapi/openapi.json` 是前后端 HTTP 契约的规范化工件。
- `go-admin-ui-plus/packages/contracts` 保存由 OpenAPI 生成的前端契约。
- `go-admin-ui-plus/scripts/check-api-contract.mjs` 校验页面、fixture、DTO 和 Go Model。
- 数据库结构只通过 `cmd/migrate/migration/version/` 中的新迁移演进。
- `release/manifest/` 将各平台候选工件绑定到同一根提交和产品版本。

## 配置与状态

提交到 Git 的文件只能保存可公开的默认值和模板。开发数据库、上传文件、日志、临时文件、
渲染后的生产配置及密钥均属于运行时状态，分别由 `dev_store/`、Compose runtime 目录或
桌面应用数据目录持有。

Speculo 的历史 change 是设计和实现证据；当前项目使用说明只存在于根 README、`docs/`、
组件 README 和各级 `AGENTS.md`。
