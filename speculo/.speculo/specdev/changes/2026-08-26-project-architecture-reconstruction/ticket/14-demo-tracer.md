---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-14
title: Demo 双方言业务曳光弹
status: in_progress
planning_depth: deep
planning_depth_reason: 参考模块同时验证公共 API、双方言 schema、权限、Web 交互与重启持久化
ready: true
risk: medium
blocked_by: [T-04, T-05, T-07]
contract_ids: [AC-021, AC-024, AC-035]
owner: codex-t14-demo
expected_changes: ["<Path>contracts/openapi/modules/demo.yaml</Path>", "<Path>go-admin-plus/internal/modules/demo/**</Path>", "<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry.go</Path>", "<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry_test.go</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/package.json</Path>", "<Path>go-admin-plus-ui/packages/ui/src/list.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/list-form.spec.ts</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/demo.yaml</Path>", "<Path>go-admin-plus/internal/modules/demo/**</Path>", "<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry.go</Path>", "<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry_test.go</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/package.json</Path>", "<Path>go-admin-plus-ui/packages/ui/src/list.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/list-form.spec.ts</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus/test/demo/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/demo/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/authorization/service.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/**</Path>", "<Path>go-admin-plus/internal/platform/database/**</Path>"]
shared_paths: ["<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry.go</Path>", "<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry_test.go</Path>", "<Path>go-admin-plus-ui/packages/ui/src/list.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/list-form.spec.ts</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry.go</Path> => T-14 under T14-D02/T14-D03; IAM-owned module capability registration only", "<Path>go-admin-plus/internal/modules/iam/authorization/capability_registry_test.go</Path> => T-14 under T14-D02/T14-D03; dual-dialect registry contract only", "<Path>go-admin-plus-ui/packages/ui/src/list.ts</Path> => T-14 under T14-D04; request normalization/validation and success-atomic state only", "<Path>go-admin-plus-ui/tests/shell/list-form.spec.ts</Path> => T-14 under T14-D04; shared list regression only", "<Path>go-admin-plus-ui/package.json</Path> => T-14 under T14-D01; Demo aggregate scripts only", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path> => T-14 under T14-D01; Demo specs only", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-14 under T14-D01 after T-11 result; two Demo importers only"]
---

# Ticket T-14: Demo 双方言业务曳光弹

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/14-demo-tracer.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>`

## 1. 战略与来源

- **目标：** 建立可复制的新架构标准 CRUD，从合同穿过模块、双方言持久化到 Web 页面。
- **可观察产出：** 授权用户在 PostgreSQL/SQLite 完成产品记录 CRUD，重启后数据保留且行为等价。
- **来源：** `US-013`、`AC-021`、`AC-024`、`AC-035`、`ADR-006`、`ADR-011`。
- **当前事实：** 旧 demo 使用旧目录/DTO/ORM 模式，不能作为新模块范例。
- **Planning Depth 原因：** 该参考切片将成为 Generator 和 Desktop 的标准验收目标。

## 2. 决策状态

### 已锁定决策

- Demo 完整遵守私有 record/repository、显式 mapping、Permission Code 和无头/Vue 分层。
- 该模块不包含特殊兼容逻辑或 tenant 字段。
- Permission Code 与导航菜单由业务模块声明、IAM-owned Module Capability Registry 校验并写入 IAM 表；Demo migration 不得跨模块写 IAM 表。

### 已采用的低影响假设

- 产品记录包含足够字段覆盖校验、分页、排序和冲突，不模拟复杂领域。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 标准 CRUD、migration、API、domain、页面与 tracer | T-02/04/05/07 基座 | 复杂跨表业务、旧 Demo 兼容 |

## 4. 要构建什么

用户搜索、分页、创建、编辑和删除 Demo 产品记录，所有命令经后端权限与领域校验后持久化；重启和切换双方言不改变合同语义。

## 5. 实现契约

- **入口或接缝：** Demo API/use cases/repository、headless domain、Vue page。
- **输入与输出：** 分页查询与 CRUD 命令；返回稳定 DTO/错误。
- **公共接口变化：** 新 Demo fragment 和 Permission Code。
- **不变量：** transport/domain/record 显式 mapping；双方言等价；无跨模块表查询；无 tenant。
- **状态或数据流：** UI -> TS client -> strict handler -> use case -> repository transaction -> refreshed list。
- **错误与失败行为：** validation/not-found/conflict/authorization 无意外状态变化。
- **兼容要求：** 不兼容旧 Demo API/schema/ID。
- **安全与隐私要求：** 后端通过真实 IAM Session 与最终权限决策授权；数据范围仅接受 `self|all` 闭集，未知值 fail closed；错误不泄露 SQL/stack。

## 6. 执行路线

1. 建立双方言 CRUD、负向合同和页面行为测试。
2. 实现 IAM-owned Permission registry、Demo IAM adapters、fragment、migration、repository、用例和 mapping。
3. 实现 headless domain、Vue 页面和标准交互，闭合写入成功但投影刷新失败后的修复状态。
4. 加入真实 IAM Session/CSRF/权限撤销/数据范围、重启持久化、缓存禁用和架构边界检查。
5. 运行 PostgreSQL/Server SQLite Web E2E。

## 7. 路径访问契约

- **预计修改点：** Demo 独占路径，经 `T14-D01` 批准的两个 Demo package manifest、根聚合脚本、测试 include 与两个 Demo lock importer，经 `T14-D02/T14-D03` 精确批准的 IAM Module Capability Registry，以及经 `T14-D04` 精确批准的共享列表状态接缝与既有回归文件。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** IAM、Database 和共享 UI。
- **共享路径：** 根 package 只增加 Demo 聚合脚本，shell Vitest config 只纳入 Demo specs；T-11 result 已完成且本 amendment 已激活，source rebase 到最新 parent 后只可更新两个 Demo importer。
- **批准偏差：** `T14-D01` 允许两个 Demo package manifest 补齐 canonical API client、`@go-admin/ui`、Vue 直接依赖、公开 export 与标准 test/typecheck 入口，并把 Demo checks 接入根 `pnpm verify`。只可复用既有 catalog/lock 版本，禁止外部版本、其他 importer、共享 UI 或产品 composition 漂移。
- **批准偏差：** `T14-D02` 允许 IAM authorization 增加通用、双方言、事务化、幂等且 fail-closed 的 Permission registry；`T14-D03` 依据后续业务模块审查把同一 seam 提升并重命名为 Module Capability Registry，在同一事务内拥有 `iam_permissions`、`iam_menus` 及 protected system-admin 的权限/菜单初始授权。Demo 只声明和调用，不得跨模块写 IAM 表。禁止修改 IAM schema、Session、Administration、产品 composition 或其他业务模块。
- **批准偏差：** `T14-D04` 只开放共享 `<Path>go-admin-plus-ui/packages/ui/src/list.ts</Path>` 与既有 `<Path>go-admin-plus-ui/tests/shell/list-form.spec.ts</Path>`：列表请求可同步规范化/校验，filters/page/pageSize/sort/rows 仅在最新请求成功后原子提交；本地校验失败必须取消旧请求且保留最后成功投影。Demo 只能配置并包装该接缝，不得复制列表状态机；禁止修改其他共享 UI 文件或改变公开列表操作集合。
- **保留或不动：** Desktop App 归 T-16，Generator 归 T-15。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | API/Web CRUD | `task test -- demo` | 双方言 CRUD、分页和重启一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |
| 失败路径 | negative contract | 校验、权限撤销、未知数据范围、not-found、stale revision、批量回滚、写后刷新失败 | 稳定错误且状态不变；修复投影不重复 mutation | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |
| 回归 | architecture/cache suite | 禁用缓存并扫描 mapping/边界 | 正确性不变且无跨层泄露 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |
| 身份与权限 | real IAM browser/API harness | canonical Cookie、43 字符 CSRF、权限注册/撤销、self/all 与 Session revoke | Web 与 API 均由真实 IAM 最终授权且拒绝后状态不变 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 非 E2E 检查。
- **E2E disposition：** required：真实 Web、API、双方言和重启路径必须验证。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不声明通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、E2E 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 模块切片独立落地，T-15/T-16 消费，T-17 组合。
- **兼容窗口：** 无旧 Demo 数据/API 兼容。
- **监控信号：** CRUD error、migration、conflict 和双方言 conformance。
- **回滚或前向恢复：** 产品接入前回滚；接入后前向 migration 修复。
- **不可逆操作与批准点：** 删除仅影响 Demo 数据并保持 UI 确认。
- **收缩条件：** T-21 证明旧 Demo 目录/schema/API 零引用。

## 10. 验收标准

- [ ] `AC-021`：Demo 在正式 profile 上完成 CRUD 与重启持久化。
- [ ] `AC-024`：Demo migration 双方言确定、幂等且行为等价。
- [ ] `AC-035`：标准列表/表单/删除交互完整。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
