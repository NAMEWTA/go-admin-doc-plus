# 开发指南

所有命令默认从仓库根目录开始。本地开发使用 SQLite 与进程内缓存/队列，不要求 Docker、
MySQL、PostgreSQL 或 Redis。

## 环境

- Go：`.go-version` 固定 `1.26.7`，`go.mod` 声明模块最低版本 `1.26.5`
- Node.js：`>=22`，建议与发布流水线一致使用 `22.22.3`
- pnpm：由 `package.json` 固定为 `11.1.3`

```bash
cd go-admin-plus
go mod download
go mod verify
cd ../go-admin-ui-plus
corepack enable
corepack install --global pnpm@11.1.3
pnpm install --frozen-lockfile
```

## 本地数据目录

`dev_store/` 是 Git 忽略的本地持久化根：

```text
dev_store/
|-- backend/
|   |-- db/go-admin.db
|   |-- logs/
|   |-- uploads/
|   `-- temp/
`-- frontend/logs/
```

创建目录和后端软链接：

```bash
mkdir -p dev_store/backend/{db,logs,uploads,temp} dev_store/frontend/logs
test -e go-admin-plus/static/uploadfile || \
  ln -s ../../dev_store/backend/uploads go-admin-plus/static/uploadfile
test -e go-admin-plus/temp || \
  ln -s ../dev_store/backend/temp go-admin-plus/temp
```

## 本地配置

创建 Git 忽略的 `go-admin-plus/config/settings.local.dev.yml`：

```yaml
settings:
  application:
    mode: dev
    host: 0.0.0.0
    name: go-admin-plus-local
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
    secret: local-development-only
    timeout: 3600
  database:
    driver: sqlite3
    source: ../dev_store/backend/db/go-admin.db
  gen:
    dbname: go-admin-plus
    frontpath: ../go-admin-ui-plus/apps/admin/src
  cache:
    memory: ''
  queue:
    memory:
      poolSize: 100
```

创建 `go-admin-ui-plus/.env.development.local`：

```dotenv
VUE_APP_BASE_API = 'http://localhost:8000'
```

两个文件都只能保存本机开发值，不能提交生产凭据。

## 初始化与启动

```bash
test -f dev_store/backend/db/go-admin.db || \
  cp go-admin-plus/go-admin-db.db dev_store/backend/db/go-admin.db
cd go-admin-plus
go run -tags sqlite3 . migrate -c config/settings.local.dev.yml
go run -tags sqlite3 . server -c config/settings.local.dev.yml
```

另开终端启动前端：

```bash
cd go-admin-ui-plus
pnpm dev --host 0.0.0.0
```

SQLite 命令必须带 `-tags sqlite3`。配置中的相对路径以 `go-admin-plus/` 为基准，因此
后端命令必须在该目录执行。

## 提交前验证

```bash
(cd go-admin-plus && go test ./...)
(cd go-admin-ui-plus && pnpm test:ci && pnpm e2e && pnpm build:prod)
node release/manifest/scan-compatibility.mjs
node --test release/manifest/product-release.test.mjs
```

依赖发生变化时才运行 `go mod tidy`。已经执行过的数据库迁移不得修改，应增加新迁移。
