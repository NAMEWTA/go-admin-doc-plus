---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-05
title: pnpm Workspace、App Shell 与交互基座
status: in_progress
planning_depth: deep
planning_depth_reason: 新前端工作区、共享核心路径和双 App 组合边界影响全部前端切片
ready: true
risk: high
blocked_by: [T-02]
contract_ids: [AC-035, AC-036]
owner: codex-t05-frontend
expected_changes: ["<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/apps/admin-web/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/**</Path>", "<Path>go-admin-plus-ui/packages/platform/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/.npmrc</Path>", "<Path>go-admin-plus-ui/apps/admin-web/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/package.json</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/desktop/package.json</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/core/**</Path>", "<Path>go-admin-plus-ui/packages/platform/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>", "<Path>go-admin-plus-ui/packages/domains/*/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/*/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/**</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>go-admin-plus-ui/packages/api-client/**</Path>", "<Path>go-admin-ui-plus/**</Path>", "<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/desktop/package.json</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/core/**</Path>", "<Path>go-admin-plus-ui/packages/platform/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/package.json</Path> => T-05", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path> => T-05", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-05", "<Path>go-admin-plus-ui/packages/adapters/browser/**</Path> => T-05", "<Path>go-admin-plus-ui/packages/adapters/desktop/package.json</Path> => T-05", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path> => T-05", "<Path>go-admin-plus-ui/packages/app-shell/src/core/**</Path> => T-05", "<Path>go-admin-plus-ui/packages/platform/**</Path> => T-05", "<Path>go-admin-plus-ui/packages/ui/**</Path> => T-05"]
---

# Ticket T-05: pnpm Workspace、App Shell 与交互基座

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/05-frontend-workspace-shell.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-05.md</Path>`

## 1. 战略与来源

- **目标：** 建立真实 pnpm Workspace、无头领域/表现层分离、双 App Shell 和标准管理交互。
- **可观察产出：** Admin Web 可加载稳定登录/未知/未授权状态；共享 list/form 支持搜索、分页、校验和确认语义。
- **来源：** `US-001`、`US-018`、`US-019`、`AC-035`、`AC-036`、`ADR-003`、`ADR-008`、`ADR-011`。
- **当前事实：** 旧前端名称、App 内业务实现和 domain 中 API/Vue 混合，锁文件会成为并行冲突点。
- **Planning Depth 原因：** Workspace/lock/App Shell 为全部前端消费者共享核心。

## 2. 决策状态

### 已锁定决策

- `apps/admin-web` 与 `apps/admin-desktop` 共享 domains、web-domains、app-shell、platform、ui。
- T-05 预建所有包 manifest 并唯一拥有 lockfile；后续模块只写自身 `src`。

### 已采用的低影响假设

- 包之间只经公开 exports 导入，禁止 `src` deep import。

### 已批准 Ticket 偏差

- `T05-D01`：T-05 的目标和 `expected_changes` 要求 app-shell 是具有公开 exports 的真实 workspace package，但初始 `writable_paths` 漏列其 manifest。implementation owner 提交偏差后，Lead 于 2026-08-26 批准只新增 `<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>`；app-shell 的其他非 `src/core` 路径仍不授权。
- `T05-D02`：规范轴检查发现 ADR-011 要求 Runtime Adapter 位于 `<Path>packages/adapters/{browser,desktop}</Path>`，App 只选择 adapter，但初始 Ticket 漏列 adapters。Lead 于 2026-08-26 批准新增 `<Path>go-admin-plus-ui/packages/adapters/browser/**</Path>` 与仅 `<Path>go-admin-plus-ui/packages/adapters/desktop/package.json</Path>`；browser 实现从 Admin Web 移出，desktop adapter 源码仍归 T-16。
- `T05-D03`：T02-D04 把已规划但此前缺失的 `<Path>packages/api-client/</Path>` 加入新 Workspace 后，边界测试因把当时发现的 package 数量硬编码为 `23` 而失败。Lead/Ticket owner 于 2026-08-26 重新打开 T-05，批准只修改 `<Path>go-admin-plus-ui/tests/shell/workspace-boundary.test.mjs</Path>`：固定必需 package 名称集合，同时继续逐一检查全部发现的未来 package，不再用脆弱总数阻止合法扩展。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| Workspace、Web shell、平台 Port、UI/list/form 基座 | 旧页面仅作能力与交互参考 | 业务页面、产品 manifest、Tauri 宿主 |

## 4. 要构建什么

前端贡献者可在固定包边界内实现业务，App 只负责宿主启动和组合；用户获得稳定导航状态以及所有管理列表一致的搜索、表单、删除确认和刷新行为。

## 5. 实现契约

- **入口或接缝：** pnpm filters、Admin Web bootstrap、App Shell state、Platform Port、共享 list/form。
- **输入与输出：** runtime adapter、identity/manifest 状态和列表 schema 输入；稳定页面状态与命令输出。
- **公共接口变化：** 建立共享包公开 exports、宿主能力和管理交互合同。
- **不变量：** App 不含业务规则；headless domain 不依赖 Vue；重复确认不重复写入；禁止 deep import/cycle。
- **状态或数据流：** App -> runtime adapter -> shell -> manifest/domain -> web-domain/UI。
- **错误与失败行为：** 未认证、未授权、未知路由和 adapter 失败显示稳定状态，不白屏。
- **兼容要求：** 不兼容旧 `go-admin-ui-plus` imports、路由或 store。
- **安全与隐私要求：** 平台 Port 不向页面暴露 secret/session 原值。

## 6. 执行路线

1. 固定 Workspace cycle、exports、shell 状态和 list/form 行为测试。
2. 创建 workspace、统一 catalog/lock 和全部包 skeleton。
3. 实现 platform、app-shell core、ui 与 Admin Web bootstrap。
4. 建立搜索/分页/校验/确认/刷新组件接缝。
5. 运行 unit、type、build、navigation 和边界回归。

## 7. 路径访问契约

- **预计修改点：** 新前端根、Admin Web 和共享 packages。
- **可写范围：** 仅 frontmatter `writable_paths`；产品聚合目录除外。
- **只读上下文：** 旧前端和 T-02 API client。
- **共享路径：** workspace/lock/app-shell core/platform/ui 由 T-05 唯一拥有。
- **保留或不动：** 产品 manifest 归 T-17，Desktop 源码归 T-16。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | shell/component suite | `task test -- frontend-shell` | Web shell 和标准 list/form 行为成立 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-05.md</Path>` |
| 失败路径 | navigation/adapter fixture | 模拟未认证、未授权、未知路由和 adapter 失败 | 显示稳定状态且不发送越权命令 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-05.md</Path>` |
| 回归 | workspace boundary | `task lint typecheck build -- frontend` | 无 cycle/deep import，包与 Web 可构建 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-05.md</Path>` |

- **Workspace checks：** 按 Goal Plan 在 current-workspace 或 source-worktree 运行 unit/type/lint/build。
- **E2E disposition：** required：需在真实浏览器验证 shell 导航和重复交互保护。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate，禁止在 source-worktree 声明通过。
- **Integration evidence：** 记录 implementation/source commit、parent before、适用 candidate/result SHA 和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 新 Workspace 扩展落地，模块逐个填充，T-17 组合，T-21 删除旧前端。
- **兼容窗口：** 施工期目录并存但不提供产品双轨。
- **监控信号：** lock drift、cycle/deep-import、bundle build 和浏览器错误。
- **回滚或前向恢复：** 产品切换前可回滚；之后修复共享包并全量重建。
- **不可逆操作与批准点：** 本 Ticket 不删除旧前端。
- **收缩条件：** T-21 证明旧 workspace/import/router 零引用。

## 10. 验收标准

- [ ] `AC-035`：共享 list/form 的正常、校验失败、确认和刷新合同可判定。
- [ ] `AC-036`：App Shell 对登录、授权和未知路由呈现稳定状态。
- [ ] T05-D03 修正与验证记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-05.md</Path>`。
- [ ] 修改未超出 `writable_paths`，共享路径仅由 T-05 修改。
- [ ] 形成新的非空 source checkpoint，并记录新的 integration result SHA。
- [ ] Ticket、Map 和 Evidence 状态一致且无未批准偏差。
