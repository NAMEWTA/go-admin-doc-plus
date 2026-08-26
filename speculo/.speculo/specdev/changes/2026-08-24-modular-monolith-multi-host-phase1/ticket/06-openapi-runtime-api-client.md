---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-06
title: 建立 OpenAPI 契约、Runtime 与纯 ApiClient
status: done
planning_depth: deep
planning_depth_reason: 变更共享 HTTP wire contract 的生成方式和所有前端请求入口，并引入 Web/Desktop 运行时接缝
ready: true
risk: high
blocked_by: [T-04, T-05]
contract_ids: [AC-004, AC-005, AC-018]
owner: root
expected_changes: ["<Path>go-admin-plus/api/openapi/**</Path>", "<Path>go-admin-ui-plus/packages/contracts/**</Path>", "<Path>go-admin-ui-plus/packages/runtime/**</Path>", "<Path>go-admin-ui-plus/packages/api-client/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/utils/request.ts</Path>", "<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>"]
writable_paths: ["<Path>go-admin-plus/api/openapi/**</Path>", "<Path>go-admin-ui-plus/packages/contracts/**</Path>", "<Path>go-admin-ui-plus/packages/runtime/**</Path>", "<Path>go-admin-ui-plus/packages/api-client/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/utils/request.ts</Path>", "<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>"]
read_only_paths: ["<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/internal/host/server/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/api/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/stores/**</Path>"]
shared_paths: ["<Path>go-admin-plus/api/openapi/**</Path>", "<Path>go-admin-ui-plus/packages/contracts/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus/api/openapi/**</Path> => T-06", "<Path>go-admin-ui-plus/packages/contracts/**</Path> => T-06"]
---

# Ticket T-06: 建立 OpenAPI 契约、Runtime 与纯 ApiClient

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/06-openapi-runtime-api-client.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-06.md</Path>`

## 1. 战略与来源

- **目标：** 让后端 OpenAPI 成为 wire contract 权威，并把传输、token、Host 配置和 UI 错误处理拆到真实接缝。
- **可观察产出：** 相同 Admin build 在 WebRuntime 和测试 DesktopRuntime 下调用同一 ApiClient；生成契约零 diff，错误只被 App Shell 呈现一次。
- **来源：** `AC-004`、`AC-005`、`AC-018`、`CODE:<Path>go-admin-ui-plus/src/utils/request.ts</Path>`。
- **当前事实：** request 同时依赖 Axios、Pinia、Element Plus、location 和构建期 base URL；类型检查通过正则读取 Go struct。
- **Planning Depth 原因：** 共享 wire contract 和认证错误流影响所有消费者。

## 2. 决策状态

### 已锁定决策

- OpenAPI 描述实际 `/api/v1` envelope、认证、错误与新增 health/capabilities；生成 TypeScript 只读。
- ApiClient 只依赖 transport、RuntimeConfig 与 TokenProvider，不依赖 Vue/Pinia/Element Plus。
- WebRuntime 返回同源 base；DesktopRuntime bootstrap interface 在 T-08 实现；业务 Domain 不探测平台。
- 旧 request 作为兼容 facade 保留到消费者迁移完成。

### 已采用的低影响假设

- 保留 Axios 作为 HTTP implementation；不在本 Ticket 更换 fetch 库。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| OpenAPI、生成 types/client、runtime、transport/token/error、旧 facade | Axios、现有 envelope、JWT 存储和 workspace | Domain 页面迁移、Wails 实现、业务接口重设计 |

## 4. 要构建什么

开发者运行一条确定命令从后端 OpenAPI 生成 contracts；应用启动先解析 Runtime，再创建 ApiClient。成功请求返回 typed envelope，401/业务/网络错误转换为稳定错误对象；App Shell 可决定重登录和提示，底层不触碰 UI。Web 和模拟 Desktop 地址均通过相同 contract tests。

## 5. 实现契约

- **入口或接缝：** OpenAPI generation、`Runtime.bootstrap()`、`createApiClient(config, tokenProvider)`、兼容 request facade。
- **输入与输出：** runtime config/token/request；typed envelope 或标准 ApiError。
- **公共接口变化：** wire contract 文档化，不改变既有业务行为。
- **不变量：** 生成代码不可手改；Domain 不拼 URL；错误只在 UI owner 呈现。
- **状态或数据流：** bootstrap -> client -> token interceptor -> HTTP -> typed result/error -> App Shell。
- **错误与失败行为：** bootstrap 缺字段、schema 漂移、未知 envelope 均显式失败；不静默 fallback 构建期 URL。
- **兼容要求：** 旧 API module 经 facade 保持可用到 T-07。
- **安全与隐私要求：** token 不写入日志或 URL；Desktop launch token 只放受控 header。

## 6. 执行路线

1. 用 characterization 补足 OpenAPI DTO/错误描述并建立确定性生成命令。
2. 创建 contracts 与漂移检查，替代正则检查但暂不删除旧脚本。
3. 建立 Runtime、TokenProvider、Transport 和 ApiError 接口及测试 adapter。
4. 实现纯 ApiClient，并让旧 request facade 委托它。
5. 验证 Web/模拟 Desktop、401、业务错误、网络错误和生成零 diff。

## 7. 路径访问契约

- **预计修改点/可写范围：** frontmatter 所列 OpenAPI、contracts、runtime、api-client 和兼容 request/check 脚本。
- **只读上下文：** 业务 DTO、ServerHost、现有 API/store 消费者。
- **共享路径：** OpenAPI 与生成 contracts 唯一 owner T-06；其他 Ticket 只消费。
- **保留或不动：** UI 页面、路由、根 workspace manifests 和业务 Go 实现。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 生成 + client contract | generate、type-check、Web/Desktop adapter tests | typed 请求通过、零 diff | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-06.md</Path>` |
| 失败路径 | ApiError/bootstrap | 401、业务、网络、缺配置用例 | 稳定错误且无 UI 副作用 | 同上 |
| 回归 | 旧 facade + API 基线 | unit/build/characterization | 旧消费者绿色 | 同上 |

- **Workspace checks：** Go OpenAPI tests；pnpm generate/type-check/unit/build。
- **E2E disposition：** required；共享 HTTP 契约跨两个仓库和 Runtime。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；Web live contract 与模拟 Desktop runtime。
- **Integration evidence：** 提交、parent/candidate/result 与包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** OpenAPI expand -> contracts -> 新 client -> 旧 facade 委托 -> T-07 消费者迁移 -> T-13 contract。
- **兼容窗口：** `request.ts` 和正则脚本保留但不得成为新代码入口。
- **监控信号：** generation diff、OpenAPI contract tests、401/error UI 计数仅在测试观察。
- **回滚或前向恢复：** facade 可暂时回指旧 implementation；公开 OpenAPI 变更必须前向兼容。
- **不可逆操作与批准点：** 删除旧 request/正则检查需 T-13 零引用和人工批准。
- **收缩条件：** 旧 imports 为零，generated diff/type/E2E 通过。

## 10. 验收标准

- [x] `AC-004`、`AC-005`、`AC-018` 通过。
- [x] ApiClient 无 Vue/Pinia/Element 依赖，生成代码零手改。
- [x] 正常、失败、回归 Evidence 完整。
- [x] 路径、提交、集成与 required E2E 合同满足。
- [x] 无未批准 wire/auth 偏差，Map/Evidence 同步。
