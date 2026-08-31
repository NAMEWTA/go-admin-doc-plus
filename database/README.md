# 数据库

Go Admin Plus 的 Server 支持 SQLite 与 PostgreSQL，Desktop 只支持 SQLite。

迁移 SQL 由各业务模块拥有，`internal/app/product` 将 provider 组合成一个全局唯一、只向前的版本序列。`cmd/migrate` 与两个运行时入口复用同一迁移 runner。

- SQLite 使用单进程文件锁、单连接和 WAL；适合 Desktop 与单实例 Server。
- PostgreSQL 使用连接池和迁移会话锁；适合多实例 Server。
- 迁移必须同时提供两种方言版本，并保持相同业务语义。
- 不提供运行时回滚命令；失败后修复输入或提交新的向前迁移。

```bash
task migrate PROFILE=server-sqlite
GO_ADMIN_DATABASE_DSN_FILE=/absolute/path/to/dsn task migrate PROFILE=server-postgres
```

## 系统管理员初始化

数据库迁移只建立 schema、受保护的 `system-admin` 角色和能力，不隐式创建可登录账号。首次部署必须在迁移完成后执行对应方言的 bootstrap SQL：

```bash
# SQLite；执行时必须停止使用该文件的 Server 或 Desktop sidecar
sqlite3 /absolute/path/to/go-admin-plus.db < database/bootstrap/sqlite/001-system-admin.sql

# PostgreSQL
psql "$GO_ADMIN_DATABASE_DSN" -v ON_ERROR_STOP=1 \
  -f database/bootstrap/postgres/001-system-admin.sql
```

首次登录凭据：

- 账号：`admin`
- 密码：`administrator password`

脚本可重复执行：已有 `admin` 时不会覆盖密码、资料或会话，只确保该账号绑定 `system-admin`。首次登录后应立即修改密码。SQLite bootstrap 使用事务和外键约束；若 migration 尚未完成或受保护角色不存在，执行会失败且不会留下半初始化数据。
