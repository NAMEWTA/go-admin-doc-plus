# ADR-0011: 前端分离无头领域与 Vue 表现层

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-011

## 上下文

业务状态、Vue 页面、transport 和宿主能力混在同一目录会让双 App 复用和独立测试失真。

## 决策

`<Path>go-admin-plus-ui/packages/domains/</Path>` 保存纯 TypeScript 领域状态和用例；`<Path>go-admin-plus-ui/packages/web-domains/</Path>` 保存 Vue 页面与表现状态；`<Path>go-admin-plus-ui/packages/platform/</Path>` 定义宿主端口；`<Path>go-admin-plus-ui/packages/adapters/</Path>` 实现运行时；App Shell 负责布局、能力和路由装配。Transport DTO 必须映射为领域模型。

## 后果

Domain 不依赖 Vue/HTTP/宿主，Web Domain 不直接 fetch 或检测运行环境，App 不拥有业务实现。共享列表/表单状态机位于 UI/领域公共合同而非页面复制。
