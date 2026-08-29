# ADR-0004: 外部合同采用 Greenfield 基线

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-004

## 上下文

旧 API、数据、schema、配置和操作模式并非自主产品的长期约束。

## 决策

产品功能完整性由当前能力、OpenAPI、迁移和验收场景定义。旧 HTTP API、旧数据库数据/schema、旧配置键和旧操作模式不提供迁移器、双写、兼容层或运行时双轨。

## 后果

部署从当前迁移基线初始化，消费者直接使用新合同。破坏当前合同仍需新的 Spec/ADR，但无需为上游历史行为保留兼容。
