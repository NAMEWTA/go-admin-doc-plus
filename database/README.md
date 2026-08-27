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
