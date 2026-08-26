---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-05
title: 将现有 Admin 等价迁入前端 workspace
status: done
planning_depth: deep
planning_depth_reason: 大规模移动单一前端根、构建配置和导入路径，需要 expand 期间保持 URL、页面和 CI 兼容
ready: true
risk: high
blocked_by: [T-01]
contract_ids: [AC-003, AC-018]
owner: root
expected_changes: ["<Path>go-admin-ui-plus/apps/admin/**</Path>", "<Path>go-admin-ui-plus/packages/config/**</Path>", "<Path>go-admin-ui-plus/src/**</Path>", "<Path>go-admin-ui-plus/public/**</Path>", "<Path>go-admin-ui-plus/.env*</Path>", "<Path>go-admin-ui-plus/index.html</Path>", "<Path>go-admin-ui-plus/package.json</Path>", "<Path>go-admin-ui-plus/pnpm-lock.yaml</Path>", "<Path>go-admin-ui-plus/pnpm-workspace.yaml</Path>", "<Path>go-admin-ui-plus/tsconfig*.json</Path>", "<Path>go-admin-ui-plus/jsconfig.json</Path>", "<Path>go-admin-ui-plus/vite.config.mjs</Path>", "<Path>go-admin-ui-plus/vitest.config.mjs</Path>", "<Path>go-admin-ui-plus/eslint.config.mjs</Path>", "<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>", "<Path>go-admin-ui-plus/plop-templates/**</Path>"]
writable_paths: ["<Path>go-admin-ui-plus/apps/admin/**</Path>", "<Path>go-admin-ui-plus/packages/config/**</Path>", "<Path>go-admin-ui-plus/src/**</Path>", "<Path>go-admin-ui-plus/public/**</Path>", "<Path>go-admin-ui-plus/.env*</Path>", "<Path>go-admin-ui-plus/index.html</Path>", "<Path>go-admin-ui-plus/package.json</Path>", "<Path>go-admin-ui-plus/pnpm-lock.yaml</Path>", "<Path>go-admin-ui-plus/pnpm-workspace.yaml</Path>", "<Path>go-admin-ui-plus/tsconfig*.json</Path>", "<Path>go-admin-ui-plus/jsconfig.json</Path>", "<Path>go-admin-ui-plus/vite.config.mjs</Path>", "<Path>go-admin-ui-plus/vitest.config.mjs</Path>", "<Path>go-admin-ui-plus/eslint.config.mjs</Path>", "<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>", "<Path>go-admin-ui-plus/plop-templates/**</Path>"]
read_only_paths: ["<Path>go-admin-ui-plus/tests/**</Path>", "<Path>go-admin-ui-plus/playwright.config.ts</Path>", "<Path>go-admin-ui-plus/.github/workflows/**</Path>", "<Path>go-admin-ui-plus/Dockerfile</Path>"]
shared_paths: ["<Path>go-admin-ui-plus/package.json</Path>", "<Path>go-admin-ui-plus/pnpm-lock.yaml</Path>", "<Path>go-admin-ui-plus/pnpm-workspace.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-ui-plus/package.json</Path> => T-05", "<Path>go-admin-ui-plus/pnpm-lock.yaml</Path> => T-05", "<Path>go-admin-ui-plus/pnpm-workspace.yaml</Path> => T-05"]
---

# Ticket T-05: 将现有 Admin 等价迁入前端 workspace

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/05-admin-workspace-shell.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-05.md</Path>`

## 1. 战略与来源

- **目标：** 在不改变用户行为的前提下建立 pnpm workspace，并把现有应用作为唯一 `apps/admin` 运行。
- **可观察产出：** 根命令仍可开发、测试和构建 Admin，登录、动态菜单、页面 URL 与核心 UI 保持基线。
- **来源：** `AC-003`、`AC-018`、`CODE:<Path>go-admin-ui-plus/package.json</Path>`。
- **当前事实：** 所有代码位于根 `src`，Vite/TS/测试脚本假设单包路径。
- **Planning Depth 原因：** 宽机械移动影响所有导入、工具和 CI；必须 expand-migrate 后才能删除旧路径。

## 2. 决策状态

### 已锁定决策

- 只创建真实 `apps/admin`；不创建未来 App 空壳。
- workspace 包范围预留 `apps/*`、`domains/*`、`packages/*`；本 Ticket 只迁移壳和共享配置。
- 根脚本继续提供 `dev/build:prod/lint/type-check/test:unit/e2e` 兼容命令并委托 Admin。
- 执行期引用图证明根 `src` 只读合同与“迁入且零旧运行依赖”冲突；路径 owner 收窄修订为源码/静态入口机械移动及其实际工具消费者，tests/Playwright/CI/Dockerfile 继续只读。

### 已采用的低影响假设

- 在 T-07 完成前，现有源码可整体位于 `apps/admin/src`，暂不按 Domain 拆分。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| workspace、Admin app、共享工具配置、根命令兼容、路径迁移 | 现有 Vue/Vite/Element/Pinia 页面和测试 | ApiClient 解耦、OpenAPI、Domain 拆分、视觉重设计 |

## 4. 要构建什么

开发者从仓库根安装依赖并运行原命令，实际构建 `apps/admin`。浏览器看到与迁移前相同的登录和管理页面，动态路由、图标、静态资源和别名正常；失败命令不被根脚本吞掉。

## 5. 实现契约

- **入口或接缝：** pnpm workspace 根脚本、Admin Vite/TS 配置、现有 Playwright webServer。
- **输入与输出：** 根命令；输出相同 dist、测试状态和开发服务器行为。
- **公共接口变化：** 无用户 HTTP/UI 路径变化；开发目录变化。
- **不变量：** 单锁文件；离线资源不改 CDN；根命令退出码透传。
- **状态或数据流：** root command -> workspace filter -> apps/admin -> dist/tests。
- **错误与失败行为：** 缺包、别名或资源时构建/测试失败，不回退旧根 src。
- **兼容要求：** 现有命令和 T-01 UI 基线绿色。
- **安全与隐私要求：** 环境文件不复制 secret；生产产物不新增遥测。

## 6. 执行路线

1. 调整 workspace/manifests 并预声明后续共享 package 所需工具依赖。
2. 移动应用与配置到 `apps/admin`，建立共享 config package。
3. 修复别名、静态资源、测试和生成目录定位。
4. 保留根命令兼容并移除对根 `src` 的运行依赖。
5. 运行 build、lint、type-check、unit 和 Playwright 基线。

## 7. 路径访问契约

- **预计修改点/可写范围：** Admin、config 和根 workspace manifests。
- **只读上下文：** 旧 src、tests 和 CI；移动时保留现有用户内容。
- **共享路径：** 根 manifest/lock/workspace 由 T-05 唯一拥有，后续 Ticket 使用 package-local manifest，新增根依赖需 deviation。
- **保留或不动：** UI 行为、后端、现有未相关资源和部署 secrets。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 根 workspace 命令 | install/build/type/lint/unit/e2e | Admin 等价通过 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-05.md</Path>` |
| 失败路径 | 脚本退出透传 | 受控错误 package/alias 验证后恢复 | 根命令非零 | 同上 |
| 回归 | T-01 Playwright | mocked/live 核心导航 | URL/页面无回归 | 同上 |

- **Workspace checks：** `pnpm install --frozen-lockfile` 后执行 config 中全部前端门禁。
- **E2E disposition：** required；目录迁移影响实际页面与资源边界。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；运行 Admin Playwright。
- **Integration evidence：** 提交、parent/candidate/result 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** workspace expand -> App 移动 -> 根命令切换 -> 零引用后移除旧根入口。
- **兼容窗口：** 根命令持续兼容；旧根 `src` 不允许与新 App 双写。
- **监控信号：** CI build/type/lint/unit/e2e、dist 路径和 bundle 可加载性。
- **回滚或前向恢复：** 在未有后续 consumer 前可整体回滚目录迁移；集成后以前向修复为主。
- **不可逆操作与批准点：** 删除旧根入口前需要零引用扫描和 Lead E2E。
- **收缩条件：** 工具、测试、CI 和源码导入不再引用根 `src`。

## 10. 验收标准

- [x] `AC-003`、`AC-018` 对应 workspace 行为通过。
- [x] 根命令兼容且只有一个锁文件和一个 Admin App。
- [x] 正常、失败、回归 Evidence 完整。
- [x] 路径、提交、集成与 required E2E 合同满足。
- [x] 无视觉/业务范围扩张，Map/Evidence 同步。
