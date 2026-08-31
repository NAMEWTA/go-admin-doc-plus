---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-15
title: 重构 Audit、Settings 与 Scheduler 页面
status: in_progress
planning_depth: standard
planning_depth_reason: 三个独立 Web Domain 使用同一成熟页面模式，跨多文件但不改变公共 API 或数据模型
ready: true
risk: medium
blocked_by: [T-11, T-12]
contract_ids: [AC-014, AC-015, AC-016, AC-017, AC-018, AC-038]
owner: codex-root
expected_changes: ["<Path>go-admin-plus-ui/packages/web-domains/audit/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/settings/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/scheduler/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/audit/src/audit*</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/src/settings*</Path>", "<Path>go-admin-plus-ui/packages/domains/scheduler/src/scheduler*</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/packages/web-domains/audit/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/settings/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/scheduler/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/audit/src/audit*</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/src/settings*</Path>", "<Path>go-admin-plus-ui/packages/domains/scheduler/src/scheduler*</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/packages/domains/audit/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/scheduler/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-15: 重构 Audit、Settings 与 Scheduler 页面

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/15-operations-pages-ui.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-15.md</Path>`

## 1. 战略与来源

- **目标：** 将审计、参数/字典和调度定义/执行记录迁移为一致的高密度工作台页面。
- **可观察产出：** 二级 routes 对应独立页面，查询/表格/详情/编辑/状态操作完整，刷新与历史稳定。
- **来源：** `US-007~008`、`US-017`、frontmatter AC、`PLAN/P4`。
- **当前事实：** 三个 Web Domain 使用大页面/tabs 与原生表单，业务 controller 和 generated client 可复用。
- **Planning Depth 原因：** 跨三个相似领域，模式明确且无公共合同迁移。

## 2. 决策状态

### 已锁定决策

- Settings 参数/字典、Scheduler definitions/executions 使用真实 routes。
- Audit 保持事实只读/清理授权，运行日志不混入 Audit 页面。
- 全部使用 T-11 组件并保持当前 API/Problem 语义。

### 已采用的低影响假设

- 三个领域在一个 Ticket 内按相同 list/detail/form pattern 迁移，可在单上下文完成。

### 已批准偏差

- **DEV-15-001（USER-DECISION:all-approved）：** T-15 可写入 Settings 与 Scheduler Web Domain 的 `package.json` 及 `pnpm-lock.yaml` 对应 importer，仅为 route-derived 页面声明现有 catalog `vue-router` 直接运行时依赖。App Shell、route manifest、共享 UI、依赖版本和其他 lock importer 继续只读。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| controllers 的 UI 状态适配、独立页面、对话框、错误/空/loading | generated clients、T-11 UI、T-12 routes | 后端/API、App Shell、真实 E2E（T-19） |

## 4. 要构建什么

用户从独立 route 扫描审计、维护参数/字典、管理任务定义并查看 executions。页面提供紧凑查询区、稳定表格尺寸、明确状态、分页和危险操作确认；后端 Problem 映射为可行动反馈并保留 trace ID。

## 5. 实现契约

- **入口或接缝：** T-12 loaders、现有 controllers/generated clients。
- **输入与输出：** typed query/form/result；route pages 和 mutation feedback。
- **公共接口变化：** 无后端变化；Web Domain exports 可按真实页面拆分。
- **不变量：** route 是 tab 唯一事实；Audit 与运行日志语义分离。
- **状态或数据流：** route/query -> controller -> table/detail/form -> refresh。
- **错误与失败行为：** validation/conflict/auth/unknown 使用一致 UI，不吞 trace。
- **兼容要求：** 删除旧 tabs 和手写控件样式。
- **安全与隐私要求：** Audit 详情不渲染 raw HTML；敏感 setting 继续按后端脱敏。

## 6. 执行路线

1. 建立三个领域 route/page 状态组件测试。
2. 按 T-11 模式迁移 Audit 列表/详情/清理。
3. 拆分 Settings 参数与字典页面。
4. 拆分 Scheduler definitions 与 executions 页面。
5. 删除旧 tabs/markup 并运行 package/workspace/视口验证。

## 7. 路径访问契约

- **预计修改点/可写范围：** Audit/Settings/Scheduler domain 非生成文件与 Web Domains。
- **只读上下文：** generated、App Shell、UI package。
- **共享路径：** 无；manifest 不修改。
- **保留或不动：** 后端、其他领域、root lockfile。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | component/controller | 代表性查询/CRUD/status tests | 三领域功能完整 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-15.md</Path>` |
| 失败路径 | Problem states | validation/conflict/403/empty/error | 可行动且不泄漏 | 同上 |
| 回归 | packages/workspace | package tests + pnpm typecheck/lint/build | 无功能/类型回归 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 执行 package unit、typecheck/lint/build、组件视口检查。
- **E2E disposition：** not-required：组件行为先完成；真实 route/API 代表流程由 T-19 required E2E。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；T-19 执行 E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-11/T-12 -> 三领域页面 -> T-19。
- **兼容窗口：** 无 tabs 双轨。
- **监控信号：** route load/mutation Problem 和 controller state errors。
- **回滚或前向恢复：** 无数据迁移，可回退 UI commit。
- **不可逆操作与批准点：** Audit cleanup/任务删除沿用后端确认与权限；无新不可逆点。
- **收缩条件：** 三个旧大页面 tabs 和原生控件扫描为零。

## 10. 验收标准

- [ ] `AC-014~018、AC-038` 的 operations 页面合同成立。
- [ ] 三领域既有业务功能、错误和权限语义完整。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-15.md</Path>`。
- [ ] commit、integration/result 和 T-19 E2E 归属完整。
