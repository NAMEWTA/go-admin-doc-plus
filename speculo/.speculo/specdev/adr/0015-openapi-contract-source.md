# ADR-0015: OpenAPI 3.1 是跨端合同唯一事实源

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-015

## 上下文

手写双端 DTO、路由 inventory 或从 Go struct 猜测前端字段无法可靠阻止合同漂移。

## 决策

`<Path>contracts/openapi/</Path>` 中的模块化 OpenAPI 3.1 是 HTTP transport 的唯一事实源。固定生成器产生 Go strict interface/transport type 与 TypeScript type/client；领域模型在两端显式映射。生成代码不可手改。

## 后果

API 修改先修改合同，并通过 lint、bundle、conformance、确定性 generate-check、Go/TS 编译和运行时负向测试。Transport 不成为领域边界。
