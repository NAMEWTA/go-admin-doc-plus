# 开发指南

## 前置环境

- Go 1.26.5
- Node.js 22 或更高版本
- pnpm 11.1.3，或当前 Node 安装提供的 Corepack；实际版本由 Workspace `packageManager` 固定
- Desktop：Rust stable、Cargo 和 Tauri 2 当前平台系统依赖
- PostgreSQL profile：可访问的 PostgreSQL 实例

首次安装前端依赖：

```bash
pnpm --dir go-admin-plus-ui install --frozen-lockfile

# 仅安装了 Corepack 时使用等价命令
corepack pnpm --dir go-admin-plus-ui install --frozen-lockfile
```

根 Task 优先使用 PATH 中的 pnpm。Node 工具管理器未把 pnpm shim 传给子进程时，命令面会在根 `.artifacts/tool-shims/` 中生成并使用 Corepack shim；该目录不进入源码或发行制品。

## Server 与 Web

SQLite 是本地默认 profile，数据库写入根目录 `.data/server/`：

```bash
task dev TARGET=server PROFILE=server-sqlite
task dev TARGET=web
```

Web 开发服务器默认连接 `127.0.0.1:8080` 的 Server。PostgreSQL 的连接材料只能通过环境或只读 secret 文件传入：

```bash
GO_ADMIN_DATABASE_DSN='postgres://user:password@127.0.0.1:5432/go_admin_plus?sslmode=disable' \
  task dev TARGET=server PROFILE=server-postgres
```

非敏感配置可通过 `GO_ADMIN_CONFIG_FILE` 指向符合 `go-admin-plus/config/schema/` 的 JSON 文件。

## Desktop

Desktop 是 Tauri 2 App。开发命令先为当前平台构建 Go sidecar，再启动 Tauri；sidecar 只使用 App 数据目录中的 SQLite，不连接 Server PostgreSQL。

```bash
task dev TARGET=desktop
```

## 数据库迁移

迁移只向前执行，由模块分别拥有并由产品组合层统一排序：

```bash
task migrate PROFILE=server-sqlite
GO_ADMIN_DATABASE_DSN_FILE=/absolute/path/to/dsn task migrate PROFILE=server-postgres
```

## 构建与本地打包

`build` 验证可编译目标。`TARGET=desktop` 会先构建当前宿主的 Go sidecar，再以
`custom-protocol` release 模式编译 Tauri 2 宿主，但不生成 bundle。Desktop 仅支持 macOS
与 Windows，因此包含 Desktop 的 `TARGET=all` 也只支持这两个宿主。`package` 生成当前宿主
可用的本地制品：

```bash
task build TARGET=all PROFILE=server-sqlite
task package TARGET=server PROFILE=server-sqlite
task package TARGET=web
task package TARGET=desktop
```

Server 和 Web 制品写入根 `.artifacts/packages/`。Desktop 使用 Tauri 2 的目标目录：macOS
生成 `.app`，Windows 生成 NSIS；不支持 Linux Desktop。本地 Desktop package 只用于构建验证，
不构成已签名发行候选。

## 验证

```bash
task test
task lint
task contract:lint
task generate:check
task governance:check
task architecture:check
task compatibility:zero
task docs:check
```

前端单独验证可在 `go-admin-plus-ui/` 运行 `pnpm lint`、`pnpm typecheck`、`pnpm test`、`pnpm check:workspace` 和 `pnpm build`。
