---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-11
title: 建立 Element Plus 管理后台设计系统
status: done
planning_depth: standard
planning_depth_reason: 跨 workspace 共享依赖、主题 token 和基础组件，但不改变后端公共合同或持久数据
ready: true
risk: high
blocked_by: []
contract_ids: [AC-017, AC-018]
owner: codex-root
expected_changes: ["<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/apps/admin-web/package.json</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/list-form.spec.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/visual-contract.spec.ts</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/apps/admin-web/package.json</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/list-form.spec.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/visual-contract.spec.ts</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/packages/app-shell/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/**</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/package.json</Path> => T-11", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-11", "<Path>go-admin-plus-ui/packages/ui/**</Path> => T-11"]
---

# Ticket T-11: 建立 Element Plus 管理后台设计系统

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/11-ui-design-system.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-11.md</Path>`

## 1. 战略与来源

- **目标：** 用 Element Plus、Sass token、Lucide 和稳定共享组件替换全局宽泛 CSS 与页面手写基础控件。
- **可观察产出：** light/dark、紧凑密度和核心组件状态可独立渲染/测试，依赖锁可复现。
- **来源：** `US-008`、`AC-017~018`、`LOG-024`、`PLAN/P4`。
- **当前事实：** `<Path>go-admin-plus-ui/packages/ui/src/admin-theme.css</Path>` 使用宽泛后代选择器模拟控件，workspace 尚未正式依赖 Element Plus/Vue Router/Lucide。
- **Planning Depth 原因：** 共享 UI/lockfile 会被全部页面消费，错误会放大到双 App，但无数据迁移。

## 2. 决策状态

### 已锁定决策

- 中性浅色工作面、炭黑侧栏、克制绿色强调、紧凑密度、持久暗色；不使用渐变或单色铺满。
- 基础组件至少含 AppPage、QueryBar、TableToolbar、DataTable、FormDialog、StatusTag、Pagination、EmptyState、FormGrid。
- Element Plus 控件和 Lucide 图标是表现层标准；颜色不是唯一状态表达。

### 已采用的低影响假设

- 依赖使用当前 pnpm catalog/lock 规则固定版本；组件默认中文文案由产品层提供。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 依赖、Sass/token、共享组件、主题持久化 primitive、组件测试 | Vue 3 workspace、现有 list/mutation helpers | App Shell（T-12）、领域页面迁移（T-14~16）、营销页 |

## 4. 要构建什么

页面作者只通过语义 token 和共享组件组合管理页面。组件提供正常、loading、empty、error、disabled、validation、danger 等完整状态，图标按钮有 accessible name/tooltip。主题切换只改变 token，不要求领域页面硬编码颜色或复制 CSS。

## 5. 实现契约

- **入口或接缝：** `@go-admin-plus/ui` exports、主题 stylesheet/plugin。
- **输入与输出：** typed props/events/slots；输出稳定 DOM/ARIA 和 tokenized visual state。
- **公共接口变化：** 新增前端共享组件 API；不改后端 API。
- **不变量：** 卡片圆角不超过 8px；letter spacing 为 0；固定控件尺寸不随内容抖动。
- **状态或数据流：** theme preference -> root attribute -> tokens -> Element Plus variables/components。
- **错误与失败行为：** 长文本换行、不覆盖相邻元素；缺失 slot 显示明确 empty 而非空白。
- **兼容要求：** 不保留旧全局控件模拟 CSS，消费者按后续 Tickets 迁移。
- **安全与隐私要求：** 不渲染 raw HTML；对话框/focus trap 使用成熟组件行为。

## 6. 执行路线

1. 固定 token、组件状态、键盘/focus 和长文本视觉 contract tests。
2. 安装并锁定 Element Plus、Lucide、Sass 与 Vue Router 依赖，建立样式入口。
3. 实现语义 token、light/dark/compact 与共享组件。
4. 删除宽泛控件模拟规则，保留最小 reset。
5. 运行 UI unit、typecheck、lint、build 和组件截图检查。

## 7. 路径访问契约

- **预计修改点/可写范围：** workspace/lock/package manifests、`packages/ui` 与两份 shell visual tests。
- **只读上下文：** App Shell、apps、领域页面。
- **共享路径：** root manifest、lockfile、UI package 仅 T-11 可写。
- **保留或不动：** 不在领域页面中提前迁移业务交互。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | component/unit | theme 与核心组件状态测试 | props/events/ARIA/token 正确 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-11.md</Path>` |
| 失败路径 | visual contract | 长文本、empty/error/disabled、reduced motion | 无溢出重叠且可操作 | 同上 |
| 回归 | workspace | `pnpm --dir go-admin-plus-ui lint && pnpm --dir go-admin-plus-ui typecheck && pnpm --dir go-admin-plus-ui test && pnpm --dir go-admin-plus-ui build` | workspace 通过 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 执行 pnpm install frozen、unit/typecheck/lint/build。
- **E2E disposition：** not-required：本 Ticket 交付组件库；目标视口真实页面 E2E 统一由 T-19/T-20。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source 不声明 E2E。
- **Integration evidence：** lock diff、commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** 共享依赖/token/components -> T-12 Shell -> T-14~16 pages。
- **兼容窗口：** 临时只允许未迁移页面继续旧 markup，不允许新增旧样式；最终 T-16 后收缩。
- **监控信号：** build size、Vue warnings、visual contract failures。
- **回滚或前向恢复：** 组件 API 前向修复；依赖问题可回退 lock/implementation commit。
- **不可逆操作与批准点：** 无。
- **收缩条件：** 宽泛全局表单/按钮/表格模拟选择器和页面手写基础控件扫描为零。

## 10. 验收标准

- [x] `AC-017~018` 的共享视觉/主题基础可独立验证。
- [x] lockfile frozen，核心组件状态、键盘/focus/reduced-motion tests 通过。
- [x] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-11.md</Path>`。
- [x] shared owner、commit、integration/result 和 E2E disposition 完整。
