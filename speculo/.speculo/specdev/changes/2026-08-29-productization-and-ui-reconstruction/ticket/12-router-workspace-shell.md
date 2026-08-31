---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-12
title: 建立 Vue Router 单一事实源与工作台 Shell
status: done
planning_depth: deep
planning_depth_reason: 共享 App Shell、双宿主 history、权限路由和所有业务页面导航均受影响
ready: true
risk: high
blocked_by: [T-11]
contract_ids: [AC-014, AC-015, AC-016, AC-017, AC-018]
owner: codex-root
expected_changes: ["<Path>go-admin-plus-ui/packages/app-shell/src/**</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/main.ts</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/styles.css</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/main.ts</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/styles.css</Path>", "<Path>go-admin-plus-ui/tests/shell/app-shell.spec.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/web-runtime.spec.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/workspace-boundary.test.mjs</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/packages/app-shell/src/**</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/main.ts</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/styles.css</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/main.ts</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/styles.css</Path>", "<Path>go-admin-plus-ui/tests/shell/app-shell.spec.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/web-runtime.spec.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/workspace-boundary.test.mjs</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/packages/ui/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/packages/app-shell/src/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/packages/app-shell/src/**</Path> => T-12"]
---

# Ticket T-12: 建立 Vue Router 单一事实源与工作台 Shell

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/12-router-workspace-shell.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-12.md</Path>`

## 1. 战略与来源

- **目标：** 消除 URL、tab 和页面状态分裂，建立双 App 共用的真实路由工作台。
- **可观察产出：** 深链接、刷新、前进后退、菜单、访问标签、面包屑、标题与内容一致；403/404 稳定。
- **来源：** `US-007~008`、`AC-014~018`、`PLAN/P3~P4`。
- **当前事实：** manifest 有二级 paths，但 ProductWorkspace 只识别一级 module，领域页面另有默认 tabs。
- **Planning Depth 原因：** App Shell 是所有页面共享核心并承载权限可达性。

## 2. 决策状态

### 已锁定决策

- manifest 声明 name/path/permission/menu/title/icon/order/component loader。
- Web HTML5 history，Desktop hash history；数据库菜单只与编译期 routes 取交集。
- 侧栏、抽屉、标签、面包屑、标题和历史只从 route 派生。

### 已采用的低影响假设

- Dashboard 保持 Shell 内编译路由；领域 Tickets 替换 loader 指向的页面实现而不修改 manifest。

### 已批准执行偏差 DEV-12-001

- **触发事实：** 既有 workspace boundary 以静态 `<FilesPage :platform>` 强制旧页面组合；T-12 改为编译期 loader + `RouterView` 后，该断言与必须删除的平行 module 状态冲突。
- **批准范围：** T-12 临时拥有 `tests/shell/workspace-boundary.test.mjs`，只把 Files Platform Port 断言迁移到 route-derived component props 和编译期 manifest loader。
- **禁止扩大：** 不修改 adapters、领域页面或 package dependency allowlist，不放宽双 App Platform Port 与 Desktop command allowlist。
- **批准来源：** 用户“都批准”覆盖全部必要偏差。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| router factory、manifest、guards、Shell/layout、双 App install、403/404 | T-11 components、现有 runtime/capability | 领域页面内部 UI（T-14~16）、Session protocol（T-13） |

## 4. 要构建什么

两个 App 从同一 manifest 创建 router，只替换 history adapter。授权 manifest 与编译 routes 取交集后形成菜单和可达路由；所有 Shell 元素监听当前 route。移动端侧栏变抽屉，桌面侧栏可折叠，访问标签不会创建第二套页面状态。

## 5. 实现契约

- **入口或接缝：** createProductRouter(host, runtime)、route manifest、ProductWorkspace。
- **输入与输出：** host/history、capability grants、compiled routes；输出 router 与导航 projections。
- **公共接口变化：** App Shell API 改为 router 驱动；后端 API 无变化。
- **不变量：** 一个 route truth；数据库不能提供 component string；403 与 404 不混淆。
- **状态或数据流：** manifest + grants -> allowed routes -> router -> menu/tags/breadcrumb/title/view。
- **错误与失败行为：** loader failure 显示可恢复错误/trace；权限撤销清理不可达 tab 并导航 403。
- **兼容要求：** 删除一级 module 和页面默认 tab 平行状态，不保留同步桥。
- **安全与隐私要求：** route guard 只改善 UX，后端仍授权；URL 不包含 secret。

## 6. 执行路线

1. 建立二级 URL、history、权限交集和恶意 component string 红灯 tests。
2. 实现 route manifest/router factory 与双 history adapters。
3. 重构 ProductWorkspace 为响应式 Shell，并派生所有导航状态。
4. 接入 Web/Desktop app entry 和 403/404/error views。
5. 删除平行 module/tab 状态并运行 manifest/unit/typecheck/build。

## 7. 路径访问契约

- **预计修改点/可写范围：** App Shell 全部源文件、双 App entry/style、shell tests。
- **只读上下文：** UI components、领域页面、adapters。
- **共享路径：** App Shell 由 T-12 唯一拥有；后续页面 Ticket 不改 manifest。
- **保留或不动：** Desktop Rust/main.rs 不修改。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | router/shell unit | deep link、history、menu/tags/title tests | 状态一致 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-12.md</Path>` |
| 失败路径 | guard/loader | 403/404、撤权、未知 component、loader error | 稳定安全页面 | 同上 |
| 回归 | dual apps | pnpm test/typecheck/build | Web/Desktop manifest 相同 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 执行 shell unit、typecheck/lint/build 与 manifest contract。
- **E2E disposition：** not-required：router/Shell 在组件层完成；真实刷新/history/视口 E2E 集中由 T-19/T-20。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source 不运行 E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-11 -> router/Shell -> T-14~16 pages -> T-19/T-20 E2E。
- **兼容窗口：** 无旧 module/tab 双轨；单 commit 切换 router。
- **监控信号：** navigation failure、403/404、lazy loader error、不可达 tab cleanup。
- **回滚或前向恢复：** 无数据迁移；可回退实现 commit，优先修复 manifest。
- **不可逆操作与批准点：** 无。
- **收缩条件：** ProductWorkspace/页面中的手写 path/tab 状态扫描为零。

## 10. 验收标准

- [x] `AC-014~018` 的组件/manifest 合同成立。
- [x] Web/Desktop history 不同但业务 route/权限一致，数据库不能加载任意组件。
- [x] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-12.md</Path>`。
- [x] shared owner、commit、integration/result 和 E2E disposition 完整。
