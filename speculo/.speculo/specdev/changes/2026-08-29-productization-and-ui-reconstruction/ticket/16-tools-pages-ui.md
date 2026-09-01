---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-16
title: 重构 Files、Generator 与 Demo 页面
status: in_progress
planning_depth: standard
planning_depth_reason: 三个工具型 Web Domain 使用稳定 API 和共享 UI，但包含文件容量错误与生成写入反馈
ready: true
risk: high
blocked_by: [T-06, T-08, T-11, T-12]
contract_ids: [AC-014, AC-015, AC-016, AC-017, AC-018, AC-027, AC-028, AC-029, AC-038]
owner: codex-root
expected_changes: ["<Path>go-admin-plus-ui/packages/web-domains/files/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/generator/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/files/src/files*</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/src/generator*</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/product*</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/packages/web-domains/files/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/generator/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/files/src/files*</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/src/generator*</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/product*</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/packages/domains/files/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-16: 重构 Files、Generator 与 Demo 页面

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/16-tools-pages-ui.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-16.md</Path>`

## 1. 战略与来源

- **目标：** 为文件、代码生成和 Demo CRUD 提供完整、一致、可扫描的工具页面。
- **可观察产出：** 文件容量/低磁盘反馈明确且仍可下载/删除；生成 preview/write 状态清晰；Demo CRUD 功能完整。
- **来源：** `US-008`、`US-012`、`US-015`、`US-017`、frontmatter AC。
- **当前事实：** 三个 Web Domain 业务 controller 可复用，但页面依赖原生 controls 和宽泛 CSS。
- **Planning Depth 原因：** UI 模式明确，但文件容量与生成写入失败需要可信交互。

## 2. 决策状态

### 已锁定决策

- Files 低水位拒绝上传但保留下载/删除；不展示其他账号精确配额。
- Generator 继续 preview 后 write，并显示 required gate 结果。
- 三页面使用 T-11 组件、T-12 routes，不修改现有业务 API。

### 已采用的低影响假设

- 容量展示只显示当前账号可公开汇总；后端未提供的精确值不在前端推断。

### 已批准偏差

- **DEV-16-001（USER-DECISION:all-approved）：** T-16 可更新 `<Path>go-admin-plus-ui/tests/shell/visual-contract.spec.ts</Path>` 中仅与 Demo 共享 `FormDialog` 识别和 Generator preview 成功逻辑源码排版有关的既有断言；不得改变其他 visual contract、共享 UI 实现、App Shell 或其他测试所有权。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| Files/Generator/Demo controllers 与 pages、容量/preview/CRUD states | generated clients、T-11 UI、T-12 routes | 生成模板实现（T-17）、后端存储、E2E runner（T-19） |

## 4. 要构建什么

用户可在 Files 页面筛选、上传、下载和删除，并在容量拒绝时仍执行释放空间操作；Generator 以步骤化表单完成表选择、配置、preview 和 write；Demo 作为标准 CRUD 参考页。loading/empty/error/validation/confirmation 均使用共享组件。

## 5. 实现契约

- **入口或接缝：** T-12 loaders、三个 domain controllers、generated clients。
- **输入与输出：** typed query/form/file/result/Problem；稳定 pages 和反馈。
- **公共接口变化：** 无后端变化；Web Domain exports 可拆分。
- **不变量：** capacity 错误不禁用下载/删除；preview 不等于 write；Demo 遵守权限。
- **状态或数据流：** route -> query/form -> controller -> generated API -> refresh/download。
- **错误与失败行为：** capacity/content/conflict/gate/auth errors 映射准确并保留 trace。
- **兼容要求：** 删除旧手写 controls/CSS，不保留平行页面。
- **安全与隐私要求：** 文件名安全渲染；preview 只作文本，不执行生成内容。

## 6. 执行路线

1. 建立三个页面正常/失败/空/loading 的组件红灯 tests。
2. 迁移 Files 并接入容量/低磁盘动作状态。
3. 迁移 Generator preview/write wizard 与 gate feedback。
4. 迁移 Demo 标准 CRUD 参考页。
5. 删除旧 markup/style 并运行 package/workspace/视口验证。

## 7. 路径访问契约

- **预计修改点/可写范围：** 三个 domains 的非生成控制器和 Web Domains。
- **只读上下文：** generated clients、App Shell、UI package。
- **共享路径：** 无；manifest/lock/生成物不修改。
- **保留或不动：** T-17 后端 generator templates。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | component/controller | file flows、generator preview/write、Demo CRUD | 功能完整 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-16.md</Path>` |
| 失败路径 | Problem states | capacity/low disk/content/gate/conflict/403 | 正确动作仍可用且不泄漏 | 同上 |
| 回归 | packages/workspace | package tests + pnpm typecheck/lint/build | 双 App 构建通过 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 执行 package unit、typecheck/lint/build、视口检查。
- **E2E disposition：** not-required：组件/domain 完成；真实上传/生成/CRUD 由 T-19 required E2E。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；T-19 执行 E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-06/T-08/T-11/T-12 -> pages -> T-17 generator template -> T-19。
- **兼容窗口：** 无旧页面双轨。
- **监控信号：** capacity/gate/CRUD Problems、route load errors。
- **回滚或前向恢复：** UI 可回退；文件/生成写入状态仍以服务端为准。
- **不可逆操作与批准点：** 文件删除/生成 write 延用明确确认；无新增不可逆协议。
- **收缩条件：** 三页面旧 controls/style 和旧 capacity 假设扫描为零。

## 10. 验收标准

- [ ] frontmatter 所列 route/visual/capacity/regression AC 成立。
- [ ] 文件低水位仍可下载/删除，Generator/Demo 功能完整。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-16.md</Path>`。
- [ ] commit、integration/result 和 T-19 E2E 归属完整。
