---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-14
title: 重构 IAM 与 Organization 管理体验
status: done
planning_depth: deep
planning_depth_reason: 页面承载认证、授权范围和不可逆账号删除交互，需跨 domain/Web Domain 且保持后端安全语义
ready: true
risk: high
blocked_by: [T-04, T-05, T-08, T-11, T-12, T-13]
contract_ids: [AC-014, AC-015, AC-016, AC-017, AC-018, AC-019, AC-020, AC-021, AC-022, AC-023, AC-024, AC-025, AC-026, AC-038]
owner: codex-root
expected_changes: ["<Path>go-admin-plus-ui/packages/domains/iam/src/administration/administration-controller*</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/administration/index.ts</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/organization*</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/index.ts</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/administration/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/session/LoginPage.vue</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/session/AccountPage.vue</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/session/GopherMark.vue</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/organization/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/organization/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/packages/domains/iam/src/administration/administration-controller*</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/administration/index.ts</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/organization*</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/index.ts</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/administration/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/session/LoginPage.vue</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/session/AccountPage.vue</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/session/GopherMark.vue</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/organization/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/organization/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/packages/domains/iam/src/administration/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-14: 重构 IAM 与 Organization 管理体验

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/14-iam-organization-ui.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-14.md</Path>`

## 1. 战略与来源

- **目标：** 交付成熟的登录、账号、角色、菜单、权限、部门、岗位、数据范围和账号删除管理流程。
- **可观察产出：** 每个二级路由是独立页面；危险删除显式展示 transfer/purge/claim 边界；数据范围可解释编辑。
- **来源：** `US-007~011`、frontmatter AC、`PLAN/P3~P5`。
- **当前事实：** IAM/Organization 大页面以 tabs 和原生控件聚合，URL 与默认页脱节。
- **Planning Depth 原因：** UI 必须准确表达授权和不可逆状态，不能靠前端隐藏替代后端约束。

## 2. 决策状态

### 已锁定决策

- users/roles/menus/departments/positions 等使用真实 routes，不在页面保留平行 tabs。
- role scope 明确五种范围和 custom 部门；账号一个主部门/多个岗位。
- 删除账号逐次选择 transfer/purge，purge 二次确认，claim 后 UI 不承诺取消。

### 已采用的低影响假设

- 列表查询与表单验证复用 T-11 管理组件；页面拆分文件名按现有 Web Domain 惯例。

### 已批准偏差

- **DEV-14-001（USER-DECISION:all-approved）：** T-14 可写入两个目标 Web Domain 的 `package.json` 和 `pnpm-lock.yaml` 中对应 importer，只为 route-derived 页面声明现有 catalog `vue-router` 直接运行时依赖。App Shell、路由定义、生成客户端、共享 UI、依赖版本与其他 lock importer 继续只读。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| IAM/Organization domain state、页面/对话框/状态、登录/账户视觉 | generated clients、T-11 UI、T-12 routes、T-13 Session | App Shell、后端授权、E2E runner（T-19） |

## 4. 要构建什么

用户从真实路由进入各管理页，使用统一查询、表格、表单和状态反馈完成 CRUD/授权。数据范围编辑显示语义与组织集合；账号删除对话框先读取状态和文件处置需求，明确不可逆边界并轮询/刷新状态。所有后端 Problem 转为可行动反馈并保留 trace ID。

## 5. 实现契约

- **入口或接缝：** T-12 route loaders、IAM/Organization controllers、T-08 generated clients。
- **输入与输出：** typed DTO/Problem、form model；输出 route page、mutation result 和可访问反馈。
- **公共接口变化：** 前端页面/domain API；不改后端 wire。
- **不变量：** 页面可见性不等于授权；危险动作总显式确认；URL 与当前子功能一致。
- **状态或数据流：** route -> controller query -> table/form -> generated mutation -> invalidate/refresh。
- **错误与失败行为：** 403/404/conflict/validation/claim boundary 映射稳定，不展示内部 detail。
- **兼容要求：** 删除旧 AdministrationPage/OrganizationPage tab 状态和手写控件模式。
- **安全与隐私要求：** 密码字段不回显/持久化；权限撤销后页面清理敏感数据。

## 6. 执行路线

1. 建立每个 route、表单状态、scope 与 deletion 状态机组件红灯 tests。
2. 扩展 headless controllers 适配生成 DTO，不复制业务规则。
3. 按 login/account/users/roles/menus/departments/positions 拆分页面。
4. 实现 scope editor、主部门/岗位和 transfer/purge 对话框。
5. 删除旧 tabs/控件 CSS，运行 package/workspace 验证和目标视口组件截图。

## 7. 路径访问契约

- **预计修改点/可写范围：** 指定 domain controllers 与 IAM/Organization Web Domains。
- **只读上下文：** generated clients、App Shell、UI package。
- **共享路径：** 无；manifest 由 T-12，生成物由 T-08。
- **保留或不动：** 不修改 Tauri、后端或其他领域页面。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | controller/component | CRUD、scope、组织、deletion states | typed 流程完整 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-14.md</Path>` |
| 失败路径 | Problem/UI state | 403/404/validation/conflict/claim after cancel | 稳定反馈且不误导 | 同上 |
| 回归 | packages/workspace | IAM/Org tests + pnpm typecheck/lint/build | 双 App 可构建 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行 package tests、typecheck/lint/build、组件视口检查。
- **E2E disposition：** not-required：本 Ticket 完成组件/domain；真实 API、route、删除和权限 E2E 由 T-19。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；T-19 执行 required Web E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-08/T-11~13 -> domain/pages -> T-19 E2E。
- **兼容窗口：** 无旧 tabs 双轨；每个 package commit 内完整替换。
- **监控信号：** Problem code、mutation failure、deletion status age、route load failure。
- **回滚或前向恢复：** UI 可回退实现 commit；已提交 purge 仍遵守后端状态，不由 UI 回滚。
- **不可逆操作与批准点：** purge submit 二次确认；implementation 前 Deep 批准。
- **收缩条件：** 旧 tabs、硬编码默认页、旧 delete/batch-delete client 调用为零。

## 10. 验收标准

- [x] frontmatter 所列 IAM/Organization/route/visual AC 在组件合同中成立。
- [x] deletion、scope、权限错误不会在 UI 中失真或绕过后端。
- [x] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-14.md</Path>`。
- [x] commit、integration/result 和 T-19 E2E 归属完整。
