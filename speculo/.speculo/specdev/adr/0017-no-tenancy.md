# ADR-0017: 租户能力与数据模型零保留

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-017

## 上下文

当前产品没有租户能力；保留 fixed/default tenant 仍会污染数据、授权、调度、缓存和接口模型。

## 决策

产品不存在 tenant package、resolver、context、多数据库选择、tenant-aware API 或租户配置。SQL/schema/migration/index/fixture 不含租户表、字段、约束或种子；每个进程只装配一个显式 Database 和一套 IAM/Worker 运行集合。

## 后果

不得引入空 `tenant_id`、默认租户、disabled flag、兼容视图或旧配置键。未来 SaaS 多租户必须作为新的架构 change 重新设计。
