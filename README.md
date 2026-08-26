# go-admin-plus

本仓库是 go-admin 前后端 monorepo，统一管理后端、前端、开发环境、发布契约与项目文档。

## 仓库组成

| 目录 | 角色 |
| --- | --- |
| `go-admin-plus/` | Go 后端 |
| `go-admin-ui-plus/` | Vue 前端 |
| `deploy/`、`release/` | 部署与产品发布契约 |
| `speculo/` | 本地规格驱动开发工件 |

前后端源码属于同一个 Git 历史、同一个 `main` 分支和同一个远端仓库。项目独立演进，不再通过 fork 或 upstream remote 同步原项目改动。

## 本地开发方案

本项目当前采用以下组合：

| 范围 | 工具或方案 | 说明 |
| --- | --- | --- |
| Go 版本管理 | goenv | 后端目录通过 `.go-version` 固定 Go 版本 |
| Go 依赖管理 | Go Modules | `go.mod` / `go.sum` 是依赖版本的唯一来源，不需要 Maven 一类的额外工具 |
| Node.js 版本管理 | Volta | 沿用本机现有的 Volta 环境 |
| 前端包管理 | pnpm | `package.json` 固定为 `pnpm@9.15.1` |
| 数据库 | SQLite | 数据文件保存在父仓库的 `dev_store/` 中 |
| 缓存与队列 | 进程内存 | 本地开发不依赖 Redis，重启后数据会清空 |
| 后台进程 | tmux（可选） | 让前后端退出终端后继续运行 |

SQLite 本地方案不需要启动 MySQL、PostgreSQL、Redis、Docker 或其他中间件。它适合单机开发；多实例部署时应改用外部数据库、Redis 缓存和可靠队列。

## 目录结构

```text
go-admin-plus/
|-- go-admin-plus/                         # 后端源码
|   |-- .go-version                        # goenv 的项目 Go 版本
|   |-- config/settings.local.dev.yml      # 本地配置（Git 忽略）
|   |-- static/uploadfile -> ../../dev_store/backend/uploads
|   `-- temp -> ../dev_store/backend/temp
|-- go-admin-ui-plus/                      # 前端源码
|   `-- .env.development.local             # 本地后端地址（Git 忽略）
|-- dev_store/                             # 本地持久化根目录（Git 忽略）
|   |-- backend/
|   |   |-- db/go-admin.db                 # SQLite 数据库
|   |   |-- logs/                          # 应用及控制台日志
|   |   |-- uploads/                       # 上传文件
|   |   `-- temp/                          # 后端临时文件
|   `-- frontend/
|       `-- logs/                          # 前端开发服务器日志
`-- README.md
```

`dev_store/` 是本地运行数据的统一根目录，已由父仓库 `.gitignore` 忽略。删除该目录会同时删除本地数据库、日志、上传文件和临时文件；操作前请按需备份。

## 首次初始化

以下命令均从仓库根目录 `go-admin-plus/` 开始执行。

### 1. 克隆仓库

```bash
git clone https://github.com/NAMEWTA/go-admin-plus.git
cd go-admin-plus
```

### 2. 安装 Go 环境

后端 `go.mod` 当前要求 Go `1.26.5`，本地使用 goenv 安装并固定兼容的 `1.26.7`：

```bash
brew install goenv
```

在 `~/.zshrc` 中加入：

```bash
# Go version management
eval "$(/opt/homebrew/bin/goenv init - zsh)"
export PATH="$(go env GOPATH)/bin:$PATH"
```

重新打开终端，或执行 `source ~/.zshrc`，然后安装 Go：

```bash
goenv install 1.26.7
goenv global 1.26.7
cd go-admin-plus
goenv local 1.26.7
go version
go mod download
go mod verify
cd ..
```

如果是 Intel Mac，Homebrew 的 goenv 路径通常是 `/usr/local/bin/goenv`；也可以用 `command -v goenv` 确认实际路径。

### 3. 准备前端环境

Node.js 继续由 Volta 管理。前端要求 Node.js `>= 22`，并通过 `packageManager` 固定 pnpm `9.15.1`：

```bash
volta install node@22
volta install pnpm@9.15.1
cd go-admin-ui-plus
node --version
pnpm --version
pnpm install --frozen-lockfile
cd ..
```

版本检查应在 `go-admin-ui-plus/` 内执行，以便 Volta 读取当前项目的工具版本约束。

### 4. 创建持久化目录

```bash
mkdir -p \
  dev_store/backend/db \
  dev_store/backend/logs \
  dev_store/backend/temp \
  dev_store/frontend/logs
```

后端源码目前将上传目录固定为 `static/uploadfile/`，临时目录固定为 `temp/`。首次迁移到统一存储目录时执行：

```bash
if [ -d go-admin-plus/static/uploadfile ] && [ ! -L go-admin-plus/static/uploadfile ]; then
  mv go-admin-plus/static/uploadfile dev_store/backend/uploads
else
  mkdir -p dev_store/backend/uploads
fi

[ -L go-admin-plus/static/uploadfile ] || \
  ln -s ../../dev_store/backend/uploads go-admin-plus/static/uploadfile

[ -e go-admin-plus/temp ] || \
  ln -s ../dev_store/backend/temp go-admin-plus/temp
```

这两个软链接属于本地运行环境，不应作为业务代码提交。仓库内原有的示例上传文件会被移动到 `dev_store/backend/uploads/`，因此 `git status` 可能将这些已跟踪示例显示为删除；这是本地目录迁移的结果，不要将其作为业务改动提交。

### 5. 创建本地配置

新建 `go-admin-plus/config/settings.local.dev.yml`：

```yaml
settings:
  application:
    mode: dev
    host: 0.0.0.0
    name: go-admin-local
    port: 8000
    readtimeout: 3000
    writetimeout: 2000
    enabledp: false
  logger:
    path: ../dev_store/backend/logs
    stdout: ''
    level: trace
    enableddb: false
  jwt:
    secret: go-admin
    timeout: 3600
  database:
    driver: sqlite3
    source: ../dev_store/backend/db/go-admin.db
  gen:
    dbname: go-admin
    frontpath: ../go-admin-ui-plus/src
  cache:
    memory: ''
  queue:
    memory:
      poolSize: 100
```

新建 `go-admin-ui-plus/.env.development.local`：

```dotenv
# Local backend started from go-admin-plus.
VUE_APP_BASE_API = 'http://localhost:8000'
```

两个文件都已被各自仓库忽略，适合保存本机配置。示例中的 JWT secret 仅用于本地开发，不能直接用于生产环境。

### 6. 初始化 SQLite

项目自带一个基础 SQLite 数据库。首次初始化时将它复制到持久化目录，然后执行迁移：

```bash
test -f dev_store/backend/db/go-admin.db || \
  cp go-admin-plus/go-admin-db.db dev_store/backend/db/go-admin.db

cd go-admin-plus
go run -tags sqlite3 . migrate -c config/settings.local.dev.yml
cd ..
```

所有 SQLite 命令都必须带 `-tags sqlite3`。遗漏该构建标签会编译到不含 SQLite 驱动的实现，并在启动时失败。

## 启动项目

### 前台启动

打开两个终端，分别运行后端和前端。

终端一：

```bash
cd go-admin-plus
go run -tags sqlite3 . server -c config/settings.local.dev.yml
```

终端二：

```bash
cd go-admin-ui-plus
pnpm dev --host 0.0.0.0
```

配置文件中的相对路径以子项目目录为基准，因此后端命令必须在 `go-admin-plus/` 中执行。

### 使用 tmux 后台启动

需要让服务在终端关闭后继续运行时，可使用两个固定的 tmux session：

```bash
tmux new-session -d -s go-admin-backend -c "$PWD/go-admin-plus" \
  "zsh -lic 'exec go run -tags sqlite3 . server -c config/settings.local.dev.yml'"

tmux new-session -d -s go-admin-frontend -c "$PWD/go-admin-ui-plus" \
  "zsh -lic 'exec pnpm dev --host 0.0.0.0'"

tmux pipe-pane -o -t go-admin-backend \
  "cat >> '$PWD/dev_store/backend/logs/backend-console.log'"

tmux pipe-pane -o -t go-admin-frontend \
  "cat >> '$PWD/dev_store/frontend/logs/vite.log'"
```

`pipe-pane` 将两个 session 后续产生的控制台输出追加到 `dev_store/`，应用自身的文件日志仍由 `settings.local.dev.yml` 写入后端日志目录。

查看运行窗口：

```bash
tmux attach -t go-admin-backend
tmux attach -t go-admin-frontend
```

从 attach 界面离开但不中止进程：按 `Ctrl+B`，松开后按 `D`。

停止服务：

```bash
tmux kill-session -t go-admin-backend
tmux kill-session -t go-admin-frontend
```

## 访问地址

| 功能 | 地址 |
| --- | --- |
| 前端管理台 | <http://localhost:9527> |
| Swagger API 文档 | <http://localhost:8000/swagger/admin/index.html> |
| Prometheus Metrics | <http://localhost:8000/api/v1/metrics> |
| 验证码接口 | <http://localhost:8000/api/v1/captcha> |

本地默认管理员账号：

```text
用户名：admin
密码：123456
```

快速检查服务：

```bash
curl -fsS -o /dev/null -w 'frontend: %{http_code}\n' http://localhost:9527/
curl -fsS -o /dev/null -w 'swagger: %{http_code}\n' http://localhost:8000/swagger/admin/index.html
lsof -nP -iTCP:8000 -sTCP:LISTEN
lsof -nP -iTCP:9527 -sTCP:LISTEN
```

## 数据持久化边界

| 数据 | 位置 | 重启后保留 |
| --- | --- | --- |
| 业务与系统数据 | `dev_store/backend/db/go-admin.db` | 是 |
| 后端日志 | `dev_store/backend/logs/` | 是 |
| 上传文件 | `dev_store/backend/uploads/` | 是 |
| 后端临时文件 | `dev_store/backend/temp/` | 是，允许按需清理 |
| 前端开发服务器日志（tmux 模式） | `dev_store/frontend/logs/` | 是，允许按需清理 |
| 缓存、验证码状态 | 后端进程内存 | 否 |
| 内存队列中的待处理任务 | 后端进程内存 | 否 |

备份本地开发数据时，至少备份 `dev_store/backend/db/` 和 `dev_store/backend/uploads/`。复制 SQLite 文件前建议先停止后端，避免得到写入中的数据库快照。

## 常用开发命令

后端：

```bash
cd go-admin-plus
go mod tidy
go build -tags sqlite3 ./...
go run -tags sqlite3 . migrate -c config/settings.local.dev.yml
go run -tags sqlite3 . server -c config/settings.local.dev.yml
```

前端：

```bash
cd go-admin-ui-plus
pnpm install --frozen-lockfile
pnpm dev --host 0.0.0.0
pnpm lint
pnpm type-check
pnpm test:unit
pnpm build:prod
```

依赖变更后才需要运行 `go mod tidy`；日常启动不需要重复执行。已经执行过的数据库迁移不要修改，应新增迁移文件继续演进。

## 常见问题

### 后端启动时出现 SQLite 驱动相关 panic

确认命令包含 `-tags sqlite3`：

```bash
go run -tags sqlite3 . server -c config/settings.local.dev.yml
```

### 前端能打开，但接口请求失败

依次确认：

1. 后端正在监听 `8000` 端口。
2. `go-admin-ui-plus/.env.development.local` 中的地址是 `http://localhost:8000`。
3. 修改 `.env.development.local` 后已经重启 Vite。

### 端口被占用

```bash
lsof -nP -iTCP:8000 -sTCP:LISTEN
lsof -nP -iTCP:9527 -sTCP:LISTEN
```

先停止已有服务，再重新启动。后端端口在 `settings.local.dev.yml` 中调整；前端可通过 `pnpm dev --host 0.0.0.0 --port 9528` 临时改用其他端口。

### 重启后验证码、缓存或队列内容消失

这是当前内存模式的预期行为。SQLite 只负责数据库持久化，不会持久化缓存和队列；需要多实例或持久队列时再接入 Redis 等外部中间件。

## 提交前后端改动

前后端现在由同一个仓库统一提交：

```bash
git add go-admin-plus go-admin-ui-plus
git commit
```

本地的 `dev_store/`、`settings.local.dev.yml` 和 `.env.development.local` 不参与提交。
