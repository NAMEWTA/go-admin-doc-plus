# ADR-0012: 跨模块使用消费者 Port 与 Integration Event

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-012

## 上下文

模块直接导入其他模块 service/model/repository 会泄漏所有权，并允许跨表写入和隐式运行时依赖。

## 决策

同步协作由消费者定义最小 Port，具体映射在 `<Path>go-admin-plus/internal/app/adapters/</Path>` 统一实现并由 product composition 注入。异步协作使用不可变 Integration Event 和 Transactional Outbox。`internal/contracts` 只容纳真正共享的值语义与 capability 常量。

## 后果

禁止跨模块 ORM/transport DTO、repository、数据库 join、模块内 IAM adapter 和 service locator。Go import graph、Generator architecture gate 与 migration-derived table ownership 持续阻止回退。
