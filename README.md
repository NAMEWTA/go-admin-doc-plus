# Go Admin Plus

Go Admin Plus 是一个单仓库管理系统产品，包含 Go Server、Vue Web App 和 Tauri 2 Desktop App。

## 仓库结构

| 路径 | 职责 |
|---|---|
| `go-admin-plus/` | Go Server、Desktop sidecar、业务模块和数据库迁移 |
| `go-admin-plus-ui/` | pnpm workspace、Web App、Tauri 2 Desktop App 和共享领域包 |
| `scripts/` | 根任务调用的开发、质量、合同和发行脚本 |
| `deploy/` | Linux 容器部署定义 |
| `release/` | 三平台打包、签名、验证和制品策略 |
| `database/` | 数据库支持和迁移约束 |
| `docs/` | 当前架构、开发和发行文档 |

## 开发启动

安装 Go 1.26.5、Node.js 22+、pnpm 11.1.3，或使用当前 Node 安装提供的 Corepack。Workspace 的 `packageManager` 固定实际 pnpm 版本；根任务在 pnpm shim 对 Node 子进程不可见时，会在已忽略的 `.artifacts/tool-shims/` 中生成 Corepack shim。Desktop 还需要 Rust stable 与当前平台的 Tauri 2 系统依赖。

```bash
# 终端 1：Server SQLite
task dev TARGET=server PROFILE=server-sqlite

# 终端 2：Web
task dev TARGET=web

# Desktop 使用本地 SQLite sidecar，无需先启动 Server
task dev TARGET=desktop
```

PostgreSQL Server 使用 `GO_ADMIN_DATABASE_DSN` 或 `GO_ADMIN_DATABASE_DSN_FILE`：

```bash
GO_ADMIN_DATABASE_DSN_FILE=/absolute/path/to/dsn task dev TARGET=server PROFILE=server-postgres
```

## 根命令

`Taskfile.yml` 是产品命令的唯一入口。常用门禁为：

```bash
task test
task lint
task contract:lint
task generate:check
task governance:check
task architecture:check
task compatibility:zero
task docs:check

# 构建 Server、Web 和当前受支持宿主的 sidecar/Tauri 可执行文件
task build TARGET=all PROFILE=server-sqlite

# 生成当前宿主的本地制品
task package TARGET=server PROFILE=server-sqlite
task package TARGET=web
task package TARGET=desktop
```

详细说明见 [开发指南](docs/development.md)、[仓库架构](docs/repository-architecture.md) 和 [发行指南](docs/release.md)。
