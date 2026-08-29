# ADR-0019: Redis 零保留并由数据库协调可靠运行时

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-019

## 上下文

Redis 会扩大 Server/Desktop 部署、测试和正确性矩阵，而当前数据库已经是三个 profile 的共同持久化边界。

## 决策

源码、依赖、配置、Compose、volume、脚本和文档不保留 Redis。可靠异步使用 Transactional Outbox；Scheduler/outbox 由协调 Worker 执行。PostgreSQL 通过固定 advisory lock 产生唯一 executor，失锁立即停止并允许接管；SQLite profile 只运行单实例。进程内缓存必须有界、可观测、可清空且不影响正确性。

## 后果

Consumer 必须幂等，claim/retry 与业务写入必须满足事务合同。Memory queue 或 cache 不能承担可靠交付或授权事实源。
