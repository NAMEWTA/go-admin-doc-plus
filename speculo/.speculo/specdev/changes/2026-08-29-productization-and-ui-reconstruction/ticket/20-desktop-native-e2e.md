---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-20
title: 建立 macOS Desktop 原生端到端门禁
status: in_progress
planning_depth: deep
planning_depth_reason: required native E2E 跨 Tauri、Go sidecar、SQLite、Keychain、原生窗口和进程清理安全边界
ready: true
risk: high
blocked_by: [T-10, T-14, T-15, T-16, T-19]
contract_ids: [AC-004, AC-005, AC-016, AC-017, AC-018, AC-031, AC-036]
owner: codex-root
expected_changes: ["<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/web-shell/browser-driver.ts</Path>", "<Path>scripts/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-build.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-production.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/web-shell/browser-driver.ts</Path>", "<Path>scripts/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-build.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-production.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/main.ts</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/first-setup/**</Path>", "<Path>release/shared/sidecar/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/web-shell/browser-driver.ts</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path> => T-20", "<Path>go-admin-plus-ui/tests/e2e/web-shell/browser-driver.ts</Path> => T-20 via DEV-20-003", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path> => T-20 via DEV-20-002", "<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path> => T-20 via DEV-20-002", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path> => T-20 via DEV-20-002", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-20 via DEV-20-002"]
---

# Ticket T-20: 建立 macOS Desktop 原生端到端门禁

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/20-desktop-native-e2e.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-20.md</Path>`

## 1. 战略与来源

- **目标：** 在 macOS 原生候选上证明 Tauri 2、sidecar、首次设置、Session vault、SQLite 重启和完整工作台。
- **可观察产出：** 原生窗口从空库 setup 到认证工作区，重启保持数据/认证合同，capability/loopback/secret/进程边界成立。
- **来源：** `US-002`、`US-008`、`US-016`、frontmatter AC、`PLAN/P7`。
- **当前事实：** 已有 native runner 和 production verifier，但可因 opt-in 缺失 Skip，且未覆盖新首次设置。
- **Planning Depth 原因：** 原生宿主、Keychain 与进程生命周期不能由 Web 或 Rust 单元测试替代。

## 2. 决策状态

### 已锁定决策

- Desktop 仅 SQLite且 API/worker 一体；首次设置由原生宿主完成。
- 个人自用不要求签名/公证；native functional/security smoke 仍 required。
- production assets 不得含测试控制标记，sidecar 仅 loopback 且 capability allowlist 最小。

### 已采用的低影响假设

- 复用现有 AppleScript/CDP/native process harness，在 macOS runner 上使用独立测试 Keychain namespace。

### 已批准执行偏差 DEV-20-002

- **触发事实：** AC-018 要求 Web/Desktop 用户切换并持久化暗色主题，但当前共享主题控制器只存在于 UI 包单元层，双宿主未实例化且工作台无主题控件；现有 T-11/T-12/T-19 Evidence 因合同映射过宽未证明该外部行为。
- **批准范围：** T-20 临时拥有 `ProductWorkspace.vue`、App Shell package manifest、lockfile 的精确 importer 与 test-only `src/native-e2e/App.vue`；只组合既有主题控制器、Lucide 明暗切换按钮，以独立 native Tauri identifier/WebKit data store 验证重启持久化，并在结束前清除测试主题存储。
- **禁止扩大：** 不改变主题存储格式、路由、Session、领域页面、native host、其他 lock importer 或 dependency version。
- **批准来源：** 用户“都批准”及“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-003

- **触发事实：** T-20 theme composition candidate 的真实 Web Shell PostgreSQL/390x844 回归到达 `/iam/users` workspace 后，T-19 driver 在 lazy `RouterView` 完成前同步查询 `user-search`，首个真实红灯为 `user search was not available for revocation check`。
- **批准范围：** T-20 临时拥有 `tests/e2e/web-shell/browser-driver.ts`，只把该同步查询改为等待同一稳定 DOM 接缝后继续原 Session revoke 断言。
- **禁止扩大：** 不增加整 suite retry、sleep、skip 或 allow-failure，不修改产品行为、viewport、timeout 上限或其他 Web E2E 场景。
- **批准来源：** 用户“都批准”及“相关的所需要批准的外部条件都批准”。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| native runner、first setup/restart/window/accessibility/secret/process checks、DEV-20-002 主题组合 | T-10 product、现有 sidecar build、production verifiers、T-11 theme controller | 修改产品 main.rs、签名、公证、Windows native E2E |

## 4. 要构建什么

Lead 在 macOS parent-candidate/current-workspace 构建 production-like Desktop，使用临时 data/log root 和测试 Keychain 启动真实窗口。自动化完成首次设置、工作区导航、退出/登录、重启和持久化检查，同时验证 loopback、vault、capability、窗口尺寸、键盘可达和 sidecar 清理。

## 5. 实现契约

- **入口或接缝：** required Desktop native E2E command、Tauri binary、macOS accessibility/Keychain/process tools。
- **输入与输出：** clean temp roots/test keychain；输出脱敏 phase results 和 artifact verification。
- **公共接口变化：** 仅增加工作台主题切换 UI；主题存储/API 与其他产品合同不变。
- **不变量：** production binary 无 test control；Keychain/data/process 全清理；不修改真实用户数据。
- **状态或数据流：** build -> launch empty -> setup/session -> workspace -> restart -> verify -> cleanup。
- **错误与失败行为：** opt-in 缺失、非 macOS、window timeout、secret leak、sidecar leak 均 required fail。
- **兼容要求：** 删除 Skip 成功语义，不要求签名/公证。
- **安全与隐私要求：** diagnostics 不含 password/token/control nonce/paths；测试 Keychain 独立并删除。

## 6. 执行路线

1. 强化 runner self-tests 与缺环境/Skip/cleanup 反向验证。
2. 加入 empty DB first setup、部分成功恢复和工作区场景。
3. 加入重启/SQLite/vault/route/目标窗口/accessibility/暗色主题持久化场景。
4. 验证 production asset、capability、loopback 和进程/Keychain 清理。
5. 在 macOS candidate 运行完整 native required Gate 并核对 Evidence。

## 7. 路径访问契约

- **预计修改点/可写范围：** Desktop E2E、专用 scripts、production verifiers，以及 DEV-20-002 精确开放的 App Shell 主题组合路径。
- **只读上下文：** main.rs、Tauri product code、WebView、sidecar build。
- **共享路径：** Desktop E2E tree 由 T-20 唯一拥有；DEV-20-002 临时开放 test-only native App、ProductWorkspace/App Shell manifest/lock importer；main.rs 与 production Desktop entry 只读且归 T-10。
- **保留或不动：** 不增测试后门，不触碰签名/公证或真实 Keychain/data。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | native app | required macOS Desktop E2E | setup/session/workspace/restart 通过 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-20.md</Path>` |
| 失败路径 | reverse/security | opt-in missing、secret marker、sidecar/keychain leak | Gate 非零且清理 | 同上 |
| 回归 | production build | cargo test/clippy + Tauri no-bundle + verifiers | native artifact 安全可运行 | 同上 |

- **Workspace checks：** source-worktree 运行 runner unit、Rust/Go/TS 非 E2E；source native E2E pass 不作为 Gate。
- **E2E disposition：** required：真实 macOS Tauri/sidecar/Keychain/window 是本 Ticket 的唯一交付。
- **E2E owner/environment：** Lead / parent-candidate（required policy）或 current-workspace；覆盖 AC-004~005、016~018、031、036。
- **Integration evidence：** source commit、candidate/result SHA、native logs/artifact hashes、父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-10/UI/Web Gate -> production-like build -> native E2E -> parent result。
- **兼容窗口：** 无旧 native runner Skip 模式。
- **监控信号：** phase duration/status、window/sidecar count、Keychain cleanup、asset hash。
- **回滚或前向恢复：** Gate 失败不推进；清理临时数据/Keychain，修复后重建 candidate。
- **不可逆操作与批准点：** 只使用临时目录/测试 Keychain；执行前 Lead 验证 target，绝不删除真实数据。
- **收缩条件：** production assets 无 native-e2e marker，required runner 无 opt-in exit 0。

## 10. 验收标准

- [ ] frontmatter 所列 Desktop/native AC 在真实 macOS candidate 成立。
- [ ] production assets、loopback、capability、vault、SQLite 重启和清理验证通过。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-20.md</Path>`。
- [ ] required E2E、candidate/result、父分支包含和 Lead 审查完整。
