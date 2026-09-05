# Go Admin Plus 0.0.2

Go Admin Plus 0.0.2 是一个单仓库管理系统产品，包含 Go Server、Vue Web App 和 Tauri 2 Desktop App。

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
| `.agents/skills/` | 项目级后端、前端和垂直切片开发规范 |

## 开发启动

安装 Go 1.26.5 或更高版本、Go Task 3.48.0、Node.js 22 或更高版本和 pnpm 11.1.3；CI 当前使用 Node.js 22.22.3。pnpm 也可由当前 Node 安装提供的 Corepack 启动，Workspace 的 `packageManager` 固定实际版本。Desktop 还需要 Rust 1.88 或更高版本和当前平台的 Tauri 2 系统依赖，CI 当前使用 Rust 1.96.0。完整安装与 PATH 说明见[开发指南](docs/development.md)。

新 Server SQLite 必须先迁移，再用权限受限的密码文件离线创建首个系统管理员。下面的
`$SECRET_FILE` 只保存文件路径；密码内容不得放入 argv、环境变量、日志或仓库：

```bash
task migrate PROFILE=server-sqlite

cd go-admin-plus
go run ./cmd/go-admin-plus bootstrap --profile server-sqlite \
  --sqlite-path ../.data/server/go-admin-plus.sqlite3 --data-root ../.data/server \
  --username first.admin --display-name "First Administrator" \
  --email first.admin@example.test --secret-file "$SECRET_FILE"
cd ..

task doctor PROFILE=server-sqlite
task dev TARGET=server PROFILE=server-sqlite
```

另开终端运行 `task dev TARGET=web`，在 Web 登录后完成 IAM、审计、调度、文件和 Demo
管理。Server PostgreSQL 同样先运行
`GO_ADMIN_DATABASE_DSN_FILE="$DSN_FILE" task migrate PROFILE=server-postgres`，再用统一 CLI 的
`bootstrap --profile server-postgres --secret-file "$SECRET_FILE"` 初始化；API 与 worker
不会隐式迁移 PostgreSQL。完整命令见[开发指南](docs/development.md)。

Desktop 使用 `desktop-sqlite` sidecar，不连接 Server。`task dev TARGET=desktop` 后在原生
首次设置页创建管理员；宿主在备份成功后迁移 SQLite，并在重启后恢复本地 Session。

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
