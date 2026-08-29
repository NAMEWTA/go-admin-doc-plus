# ADR-0016: 持久化采用 Bun 与 Goose

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-016

## 上下文

双数据库支持需要可见 SQL、明确 schema 演进和不污染领域模型的 adapter，而不是全局 ORM runtime。

## 决策

模块 repository 使用 Bun + `database/sql`，领域模型不带数据库 tag。模块拥有双方言不可变 migration，统一 Goose Provider 按确定顺序执行。Product-owned migration operation 管理 profile、数据库与 runner 生命周期；command 只处理 launch material 和输出。

## 后果

禁止 AutoMigrate、跨模块 table/repository、根 database 中的生产 migration 和领域对象直接序列化。架构门禁从 migration 推导 table owner，并在 Generator 发布前运行相同检查。
