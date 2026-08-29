# ADR-0001: 新架构不保留旧模式兼容

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-001

## 上下文

项目已脱离上游并自主研发。保留旧目录、内部导入、脚本入口、依赖或文档别名会让历史结构继续约束当前产品。

## 决策

当前架构是唯一权威。旧目录、module/package scope、内部导入、脚本入口、API/schema/config、Wails、JWT、Casbin、Redis、tenant、MySQL、SQL Server 和施工支架不得通过 alias、转发或双轨继续存在。全仓 `compatibility:zero` 门禁锁定该边界。

## 后果

历史消费者必须迁移到当前合同，不能依赖兼容 shim。未来恢复任何已删除能力必须通过新的架构 change，而不是在现有路径加入隐式 fallback。
