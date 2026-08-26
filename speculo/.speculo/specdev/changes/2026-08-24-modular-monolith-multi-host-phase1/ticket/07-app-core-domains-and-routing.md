---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-07
title: 抽离 App Core、Domain 并迁移稳定路由键
status: done
planning_depth: deep
planning_depth_reason: 批量迁移页面、状态和路由消费者，并 expand 公共菜单 component 到稳定 routeKey
ready: true
risk: high
blocked_by: [T-05, T-06]
contract_ids: [AC-002, AC-003, AC-006, AC-018]
owner: root
expected_changes: ["<Path>go-admin-ui-plus/packages/app-core/**</Path>", "<Path>go-admin-ui-plus/packages/ui/**</Path>", "<Path>go-admin-ui-plus/domains/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/**</Path>", "<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>", "<Path>go-admin-ui-plus/tests/e2e/mocked/support/crud.ts</Path>", "<Path>go-admin-plus/app/admin/models/sys_menu.go</Path>", "<Path>go-admin-plus/app/admin/service/sys_menu.go</Path>", "<Path>go-admin-plus/app/admin/service/sys_menu_route_key_test.go</Path>", "<Path>go-admin-plus/cmd/migrate/migration/version/**</Path>"]
writable_paths: ["<Path>go-admin-ui-plus/packages/app-core/**</Path>", "<Path>go-admin-ui-plus/packages/ui/**</Path>", "<Path>go-admin-ui-plus/domains/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/**</Path>", "<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>", "<Path>go-admin-ui-plus/tests/e2e/mocked/support/crud.ts</Path>", "<Path>go-admin-plus/app/admin/models/sys_menu.go</Path>", "<Path>go-admin-plus/app/admin/service/sys_menu.go</Path>", "<Path>go-admin-plus/app/admin/service/sys_menu_route_key_test.go</Path>", "<Path>go-admin-plus/cmd/migrate/migration/version/**</Path>"]
read_only_paths: ["<Path>go-admin-ui-plus/packages/contracts/**</Path>", "<Path>go-admin-ui-plus/packages/api-client/**</Path>", "<Path>go-admin-plus/api/openapi/**</Path>", "<Path>go-admin-ui-plus/tests/unit/**</Path>", "<Path>go-admin-ui-plus/tests/e2e/**/*.spec.ts</Path>", "<Path>go-admin-ui-plus/tests/e2e/mocked/fixtures.ts</Path>"]
shared_paths: ["<Path>go-admin-plus/cmd/migrate/migration/version/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus/cmd/migrate/migration/version/**</Path> => T-07"]
---

# Ticket T-07: 抽离 App Core、Domain 并迁移稳定路由键

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/07-app-core-domains-and-routing.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-07.md</Path>`

## 1. 战略与来源

- **目标：** 让 `apps/admin` 只装配 App Core、UI 和业务 Domain，并用白名单 routeKey 逐步替代后端物理 Vue component 路径。
- **可观察产出：** 登录、菜单、权限和全部现有管理页面保持可用；Domain 可单独类型检查，未知路由键安全失败。
- **来源：** `AC-002`、`AC-003`、`AC-006`、`AC-018`。
- **当前事实：** store 通过 `import.meta.glob('../views/**/*.vue')` 解析后端 component；页面/API/store 以全局 `@` 深层互相引用。
- **Planning Depth 原因：** 宽消费者迁移、菜单数据兼容和共享路由安全。

## 2. 决策状态

### 已锁定决策

- App Core 拥有 session、permission、menu、error presentation 和 Domain registry；UI package 不包含业务状态。
- Domain 为 system/jobs/demo/tools/monitor，公开 barrel interface，禁止 Domain 到 Domain 深层 import。
- routeKey 是稳定白名单标识；兼容解析器在 routeKey 缺失时读取旧 component，但未知值绝不任意动态导入。
- 迁移顺序先 demo tracer，再 system/jobs/tools/monitor 批次，每批保持 workspace 绿色。

### 已采用的低影响假设

- dashboard、login、profile 和错误页作为 App Core/App Shell 能力，不单独创建业务 Domain。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| App Core、UI、Domain registry/迁移、routeKey expand、菜单 migration | 现有页面、组件、ApiClient、generated contracts | 新视觉体系、新业务页面、删除兼容层 |

## 4. 要构建什么

Admin 启动后由 App Core 初始化 session 和菜单，将 routeKey 映射到显式注册的 Domain 页面。旧 component 数据在兼容期继续工作；未知键进入稳定不可用页并记录可诊断信息。每个 Domain 通过自身 facade 调用 ApiClient，用户看到的现有页面和权限保持不变。

## 5. 实现契约

- **入口或接缝：** App Core bootstrap、Domain registry/barrel、route resolver、菜单 DTO expand。
- **输入与输出：** session/menu/capabilities；输出 Vue routes、Domain 页面或安全错误 route。
- **公共接口变化：** 菜单响应可新增 routeKey，旧 component 保留。
- **不变量：** 无任意路径 import；Domain 不操作平台；UI 不访问业务 API；权限先于页面加载。
- **状态或数据流：** Runtime -> App Core -> auth/menu -> resolver -> Domain route -> ApiClient。
- **错误与失败行为：** 未知/重复 routeKey、Domain 初始化失败和无权限均稳定隔离，不白屏。
- **兼容要求：** 页面 URL、菜单层级、权限与 T-01 E2E 保持。
- **安全与隐私要求：** routeKey 白名单；错误日志不包含 token。

## 6. 执行路线

1. 建立 App Core、UI 与 Domain registry interface，并测试依赖规则。
2. 用 demo Domain 做端到端迁移，保留旧 consumer facade。
3. expand 后端 menu routeKey 和兼容 resolver，验证旧/新/未知值。
4. 按 system、jobs、tools、monitor 分批迁移页面/API/composable，每批运行门禁。
5. 扫描跨 Domain 深层 import 和旧物理路径消费者，保留待 T-13 收缩清单。

## 7. 路径访问契约

- **预计修改点/可写范围：** App Core/UI/Domains/Admin src、routeKey menu 版本迁移，以及菜单树序列化中保留 routeKey 的最小 service 接缝与测试。实现审查发现该树由手工字段复制构造，因此仅修改模型无法满足既定响应契约。另允许唯一 mocked E2E response helper 补齐 T-06 已锁定的 `msg/data` envelope；测试用例、断言与业务 fixture 保持只读。T-08 回归门发现 API contract checker 仍只扫描迁移前的 Admin views，允许纠正其 page roots 与 Options-API 路径，使 Domain pages 恢复自动覆盖。
- **只读上下文：** generated contracts、ApiClient、OpenAPI 和测试。
- **共享路径：** 新 migration version 目录由 T-07 唯一拥有；历史 migration 不动。
- **保留或不动：** 根 manifests、桌面、Docker、现有 UI 视觉和业务权限语义。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | App Core/route resolver/Playwright | unit + 全页面导航 | 旧新菜单均可用 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-07.md</Path>` |
| 失败路径 | 未知/重复键/无权限 | resolver 与 E2E 用例 | 安全不可用页、无任意导入 | 同上 |
| 回归 | workspace 全门禁 | lint/type/unit/build/e2e | 核心 UI/权限不回归 | 同上 |

- **Workspace checks：** package dependency scan、pnpm 全门禁、后端 menu tests。
- **E2E disposition：** required；全部 Admin 页面和菜单行为迁移。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；mocked/live 导航与权限矩阵。
- **Integration evidence：** 提交、parent/candidate/result 与包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** registry expand -> demo -> routeKey expand -> Domain batches -> observe；T-13 contract。
- **兼容窗口：** component 与 routeKey 双读，后端优先发 routeKey；旧字段不在本 Ticket 删除。
- **监控信号：** unknown routeKey 日志、E2E 页面覆盖、import 扫描。
- **回滚或前向恢复：** 单 Domain 可回指兼容 facade；已写 routeKey 数据仍保留旧 component 以支持回滚。
- **不可逆操作与批准点：** 删除 component 字段/解析器需 T-13 零调用证据与批准。
- **收缩条件：** 所有菜单有有效 routeKey、旧 component 消费为零、完整导航 E2E 通过。

## 10. 验收标准

- [x] `AC-002`、`AC-003`、`AC-006`、`AC-018` 通过。
- [x] Domain 依赖与 routeKey 安全不变量可自动检查。
- [x] 迁移批次、失败和回归 Evidence 完整。
- [x] 路径、提交、集成和 required E2E 合同满足。
- [x] 未删除兼容字段，无未批准偏差。
