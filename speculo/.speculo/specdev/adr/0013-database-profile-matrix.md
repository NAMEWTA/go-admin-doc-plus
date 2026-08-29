# ADR-0013: Server 双数据库，Desktop 仅 SQLite

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-013

## 上下文

产品需要兼顾完整 PostgreSQL 服务部署、轻量 Server SQLite 和自包含 Desktop SQLite，而不维护旧四数据库兼容矩阵。

## 决策

正式 profile 只有 `server-postgres`、`server-sqlite`、`desktop-sqlite`。Server SQLite 覆盖完整功能，不是开发 fallback；Desktop 仅使用 SQLite。SQLite profile 单实例，PostgreSQL 可多 API 副本并由唯一协调 Worker 执行调度/outbox。

## 后果

所有模块 migration/repository 必须在 PostgreSQL 与 SQLite 验证。MySQL/SQL Server 不受支持；API 启动禁止 AutoMigrate，Desktop 在服务就绪前备份并迁移。
