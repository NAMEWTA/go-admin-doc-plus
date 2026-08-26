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
expected_changes: ["<Path>contracts/openapi/modules/demo.yaml</Path>", "<Path>go-admin-plus/internal/modules/demo/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/package.json</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/demo.yaml</Path>", "<Path>go-admin-plus/internal/modules/demo/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/package.json</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus/test/demo/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/demo/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/authorization/**</Path>", "<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/package.json</Path> => T-14 under T14-D01; Demo aggregate scripts only", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path> => T-14 under T14-D01; Demo specs only", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-14 under T14-D01 after T-11 result; Demo importers only"]
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
- **安全与隐私要求：** 后端最终授权，错误不泄露 SQL/stack。

## 6. 执行路线

1. 建立双方言 CRUD、负向合同和页面行为测试。
2. 实现 fragment、migration、repository、用例和 mapping。
3. 实现 headless domain、Vue 页面和标准交互。
4. 加入重启持久化、缓存禁用和架构边界检查。
5. 运行 PostgreSQL/Server SQLite Web E2E。

## 7. 路径访问契约

- **预计修改点：** Demo 独占路径，以及经 `T14-D01` 批准的两个 Demo package manifest、根聚合脚本、测试 include 与两个 Demo lock importer。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** IAM、Database 和共享 UI。
- **共享路径：** 根 package 只增加 Demo 聚合脚本，shell Vitest config 只纳入 Demo specs；lockfile 与 T-11 串行，必须等待 T-11 result 并 rebase 后才可只更新两个 Demo importer。
- **批准偏差：** `T14-D01` 允许两个 Demo package manifest 补齐 canonical API client、`@go-admin/ui`、Vue 直接依赖、公开 export 与标准 test/typecheck 入口，并把 Demo checks 接入根 `pnpm verify`。只可复用既有 catalog/lock 版本，禁止外部版本、其他 importer、共享 UI 或产品 composition 漂移。
- **保留或不动：** Desktop App 归 T-16，Generator 归 T-15。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | API/Web CRUD | `task test -- demo` | 双方言 CRUD、分页和重启一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |
| 失败路径 | negative contract | 校验、缺权、not-found、conflict | 稳定错误且状态不变 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |
| 回归 | architecture/cache suite | 禁用缓存并扫描 mapping/边界 | 正确性不变且无跨层泄露 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-14.md</Path>` |

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
