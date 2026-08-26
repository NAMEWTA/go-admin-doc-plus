# Go Admin Plus

Go Admin Plus 是一个自主维护的前后端一体化管理平台。仓库同时包含 Go API、Vue 3
管理端、Wails 桌面应用、Linux Compose 部署和跨平台发布契约，所有组件共享同一个
Git 历史与 `main` 分支。

## 仓库结构

| 路径 | 职责 |
| --- | --- |
| `go-admin-plus/` | Go API、数据库迁移、Wails 桌面宿主 |
| `go-admin-ui-plus/` | Vue 3 + Element Plus 管理端 workspace |
| `deploy/` | Linux Compose 部署定义 |
| `release/` | Linux、macOS、Windows 发布与安装契约 |
| `.github/` | Issue、PR、CI、安全扫描和发布门禁 |
| `docs/` | 当前开发、架构与发布文档 |
| `speculo/` | 本地规格驱动开发运行时与工件 |

## 工具版本

| 工具 | 当前约束 | 权威来源 |
| --- | --- | --- |
| Go | 模块最低 `1.26.5`，本地工具链 `1.26.7` | `go.mod`、`.go-version` |
| Node.js | `>=22`，发布流水线使用 `22.22.3` | `package.json`、根级 workflow |
| pnpm | `11.1.3` | `package.json#packageManager` |

## 快速启动

首次准备环境和本地配置见 [开发指南](docs/development.md)。完成初始化后，在两个终端
分别启动服务：

```bash
cd go-admin-plus
go run -tags sqlite3 . server -c config/settings.local.dev.yml
```

```bash
cd go-admin-ui-plus
corepack enable
pnpm install --frozen-lockfile
pnpm dev --host 0.0.0.0
```

默认访问地址：

| 服务 | 地址 |
| --- | --- |
| Web 管理端 | <http://localhost:9527> |
| Swagger | <http://localhost:8000/swagger/admin/index.html> |
| 就绪检查 | <http://localhost:8000/health/ready> |

本地基础管理员账号为 `admin / 123456`。该凭据只适用于开发数据，生产部署必须修改
初始密码、JWT secret 和数据库凭据。

## 质量门禁

```bash
cd go-admin-plus
go test ./...
```

```bash
cd go-admin-ui-plus
pnpm test:ci
pnpm e2e
pnpm build:prod
```

根级 CI 在 `main` 和 Pull Request 上执行同等检查。平台发布 workflow 只生成短期候选
工件，不会自动创建 GitHub Release、部署生产环境或向外部分发。

## 文档

- [文档索引](docs/README.md)
- [开发指南](docs/development.md)
- [仓库架构](docs/repository-architecture.md)
- [发布说明](docs/release.md)
- [后端架构](go-admin-plus/docs/architecture.md)
- [后端开发约束](go-admin-plus/AGENTS.md)
- [前端开发约束](go-admin-ui-plus/AGENTS.md)

## 许可证

项目使用 MIT License，详见 [LICENSE](LICENSE)。派生来源和第三方声明见
[NOTICE.md](NOTICE.md)。
