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
expected_changes: ["<Path>go-admin-plus/test/desktop/fixture/main.go</Path>", "<Path>go-admin-plus/test/desktop/fixture/main_test.go</Path>", "<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/web-shell/browser-driver.ts</Path>", "<Path>scripts/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-build.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-production.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/first-setup/FirstSetupGate.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>go-admin-plus/test/desktop/fixture/main.go</Path>", "<Path>go-admin-plus/test/desktop/fixture/main_test.go</Path>", "<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/web-shell/browser-driver.ts</Path>", "<Path>scripts/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-build.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-production.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/first-setup/FirstSetupGate.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/Cargo.toml</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/Cargo.lock</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/build.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/capabilities/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/tauri.conf.json</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/main.ts</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/first-setup/client.ts</Path>", "<Path>release/shared/sidecar/**</Path>"]
shared_paths: ["<Path>go-admin-plus/test/desktop/fixture/main.go</Path>", "<Path>go-admin-plus/test/desktop/fixture/main_test.go</Path>", "<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/web-shell/browser-driver.ts</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/first-setup/FirstSetupGate.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus/test/desktop/fixture/main.go</Path> => T-20 via DEV-20-009 for current-minus-audit migration baseline", "<Path>go-admin-plus/test/desktop/fixture/main_test.go</Path> => T-20 via DEV-20-009", "<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path> => T-20", "<Path>go-admin-plus-ui/tests/e2e/web-shell/browser-driver.ts</Path> => T-20 via DEV-20-003", "<Path>go-admin-plus-ui/apps/admin-desktop/src/native-e2e/App.vue</Path> => T-20 via DEV-20-002", "<Path>go-admin-plus-ui/apps/admin-desktop/src/first-setup/FirstSetupGate.vue</Path> => T-20 via DEV-20-008 for recovery-only session clear", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path> => T-20 via DEV-20-006 for native-e2e context injection only", "<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path> => T-20 via DEV-20-002", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path> => T-20 via DEV-20-002", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-20 via DEV-20-002"]
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

### 已批准执行偏差 DEV-20-004

- **触发事实：** AC-018 明确要求 Web/Desktop 用户切换并在重启后恢复暗色主题；DEV-20-002 已补产品组合和 native restart 场景，但 T-19 Web Evidence 只执行 viewport/reduced-motion，未在真实浏览器切换或 reload 主题。
- **批准范围：** T-20 在同一 `tests/e2e/web-shell/browser-driver.ts` 中点击共享“使用深色主题”控件，验证根 theme token，并在既有 deep-link reload 后再次验证暗色偏好与选中控件。
- **禁止扩大：** 不修改测试后门、产品存储 API、浏览器 profile、viewport、retry/skip 或其他业务断言；浏览器 user-data 仍使用 runner 的 disposable root 并清理。
- **批准来源：** 用户“都批准”及“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-005

- **触发事实：** 当前 Windows host 无 macOS VM 或可用 SSH macOS 会话，仓库无 self-hosted runner；GitHub-hosted macOS 对 Accessibility 的支持存在官方已知限制，但它是当前唯一仍可实测的真实 macOS 通道。
- **批准范围：** Lead 可从规范 candidate `2bce131` 创建一个只修改 `.github/workflows/ci.yml` 的隔离 probe descendant，增加 `macos-15` 完整 native E2E job；可 push 一个具名临时 probe branch，并在每次首个真实红灯都被精确归因和修正的前提下手动 dispatch 最多 3 次，观察真实 AX、Keychain、窗口、sidecar 与精确 marker。
- **禁止扩大：** probe 不进入 source 或规范 candidate，不修改产品/测试代码，不使用 production environment 或签名/公证 secrets，不发布 artifact，不 deploy/migrate，不重写或清理远端历史，不做无归因 retry；hosted runner 失败不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-006

- **触发事实：** DEV-20-005 第三次且最后一次 probe 在真实 macOS 15.7.7 arm64 上通过 sidecar/UI prebuild 后，于 `tauri::generate_context!()` 编译阶段报 `expected [u8; 16], found Vec<u8>`；锁定的 `tauri-utils 2.9.3` 及当前上游 `dev` 均把 `WindowConfig.data_store_identifier: Option<[u8; 16]>` 错误交给 `opt_vec_lit`，配置 JSON 无法规避该类型错误。
- **批准范围：** T-20 临时拥有 `src-tauri/src/main.rs`，仅可在 `native-e2e` feature 下构造可变 Tauri Context，并把固定测试 WebKit data store identifier 以 Rust `[u8; 16]` 注入唯一窗口；runner 从 `TAURI_CONFIG` 删除该字段并由 self-test 精确验证 Rust 注入、测试配置不含字段且生产上下文路径不变。修正形成新的非空 source checkpoint 和规范 candidate；Lead 可更新同一 workflow-only probe descendant、push 同一具名 probe branch，并最多再 dispatch 3 次逐个归因的 `macos-15` attempt。
- **禁止扩大：** 不修改 production Tauri 配置、Cargo 依赖/锁、capability、窗口合同、主题存储、Keychain namespace 或产品运行行为；生产 build 不得包含测试 data store identifier；不得 retry/skip/allow-failure、发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-007

- **触发事实：** DEV-20-006 第三次且最后一次 probe `33504927079` 使用带固定枚举恢复态诊断的 candidate；日志仍为无后缀 `first-setup-recovery-state`。按 runner 控制流，这反证 90 秒“管理员已创建”等待已通过，失败只可能在随后恢复 snapshot、点击“进入登录”或等待登录页三步之一；这三步共用旧 phase，受控命令错误还会删除原始 AppleScript 内容，现有安全日志无法继续细分。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/run.mjs` 与 `run.test.mjs`，在既有 restore、click、login poll 之前分别设置固定 phase，并由 self-test 锁定顺序；不得读取或输出任意 UI 内容、输入、路径或秘密。形成新的非空 source checkpoint 和规范 candidate 后，Lead 可更新同一 workflow-only probe descendant、push 同一具名 probe branch，并最多再 dispatch 3 次逐个归因的 `macos-15` attempt。
- **禁止扩大：** 不修改 FirstSetupGate、产品行为、Stronghold/Keychain、Tauri Context/config、workflow job、30/90 秒上限、retry/skip/allow-failure；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。每次前一红灯必须先归因并修正，hosted 失败不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-008

- **触发事实：** DEV-20-007 attempt 2 `33508881242` 在真实 macOS 15.7.7 arm64 上通过恢复态、Stronghold snapshot restore 和“进入登录”点击，随后固定页面分类精确返回 `first-setup-recovery-login-workspace`。这证明 `FirstSetupGate` 的恢复按钮只挂载工作台而未清除当前进程中已建立但未持久化的 Session，违反 AC-005“显示可恢复错误并进入普通登录”；放宽 runner 接受工作区不是合法修正。
- **批准范围：** T-20 临时拥有 `apps/admin-desktop/src/first-setup/FirstSetupGate.vue`，只可在 recovery 按钮路径复用既有 `createDesktopSession().logout()`，成功清除本地/内存 Session 后再挂载登录壳；普通首次设置 complete 路径仍直接进入工作区。清理失败必须留在 recovery 状态并提供固定错误与重试。可在既有 Desktop runner self-test 中锁定该产品接缝。形成新的非空 source checkpoint 和规范 candidate 后，Lead 可更新同一 workflow-only probe descendant、push 同一具名 probe branch，并最多再 dispatch 3 次逐个归因的 `macos-15` attempt。
- **禁止扩大：** 不新增/修改 native command、capability、公开 API、依赖/lock、Session 存储格式、超时/retry/skip/allow-failure或其他 first setup/工作台行为；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。DEV-20-007 剩余 attempt 不用于包含产品修正的 candidate。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-009

- **触发事实：** DEV-20-008 attempt 1 `33511581184` 已越过恢复登录、恢复重启、普通首次设置和普通重启，随后在 `login-window` 启动 `previous` SQLite fixture 时精确失败为 `desktop sidecar stopped during startup`。本机对同一 fixture 分步执行数据库打开与 `BuildDesktop`，确定性复现 `product migration failed`：fixture 设计上只应缺少最后的 `8100000000000_audit.sql`，但其 provider 列表还漏掉 `6210/6220/6221/7110/7120/7510` 等当前前置迁移，形成高版本已应用而低版本缺失的非前向历史。
- **批准范围：** T-20 临时拥有 `go-admin-plus/test/desktop/fixture/main.go` 与 `main_test.go`；只可补齐 audit 之前的当前产品迁移 providers，并以结构化 compose 比较锁定 previous baseline 与 `product.NewMigrationRunner()` 恰好只差 `8100000000000_audit.sql`。`migration-failure` 模式必须继续用 `audit_facts` 冲突证明迁移失败和原库恢复，`previous` 模式必须允许产品只前向应用这一项。修正形成新 source/candidate 后，继续使用 DEV-20-008 剩余 2 次有序 hosted attempt，不新增配额。
- **禁止扩大：** 不修改任何产品 migration SQL/provider、公开 API、生产数据库策略、sidecar 诊断、安全日志、runner timeout/retry/skip、workflow 或其他 fixture；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-010

- **触发事实：** DEV-20-008 attempt 2 `33515309949`（job `99880740594`）在真实 macOS 15.7.7 arm64 上通过 prebuild，修正后的 `previous` fixture sidecar 不再提前退出，但 90 秒内未出现登录页。fixture 现在应用 `6210000000000_iam_bootstrap_recovery.sql` 并创建一个账号和角色，却未写 `iam_bootstrap_state`；产品 `desktopSetup.state` 对 `accounts=1, markers=0` 必然返回 inconsistent，因此页面不会进入普通登录。
- **批准范围：** T-20 仍只可修改 `go-admin-plus/test/desktop/fixture/main.go` 与 `main_test.go`；将固定的 `marker=1`、`account-desktop-e2e` bootstrap marker 与账号、system-admin 角色在同一 fixture transaction 原子写入，并用测试锁定 marker/account/role 一致性。修正形成新 source/candidate 后，只可使用 DEV-20-008 最后 1 次有序 hosted attempt，不新增配额。
- **禁止扩大：** 不修改产品 bootstrap/setup、migration SQL/provider、公开 API、runner diagnostics/timeout/retry/skip、workflow、fixture credential 或其他 fixture；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-011

- **触发事实：** DEV-20-008 第三次且最后一次 attempt `33518029405`（job `99889910987`）在真实 macOS 15.7.7 arm64 上通过 fixture 启动、登录、认证工作区和窗口验证，随后点击既有“产品示例”导航后在 `login-demo` 超时。runner 当前只保留通用 phase，无法区分 route chunk 加载失败、产品服务 unavailable/no-projection、403、404、runtime unavailable、loading、login、workspace 或 unknown。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/run.mjs` 与 `run.test.mjs`，在三个既有 Demo page poll 失败后按固定字符串枚举上述状态，并由 self-test 锁定分类集合和调用顺序；不得输出任意 UI 内容。形成新 source/candidate 后，Lead 可更新同一 workflow-only probe descendant、push 同一具名 probe branch，并最多 dispatch 3 次逐个归因的 `macos-15` attempt。
- **禁止扩大：** 不修改产品、API、capability、migration、fixture、workflow、30 秒上限、retry/sleep/skip/allow-failure；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。每个前一红灯必须先归因并修正，hosted 失败不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| native runner、first setup/restart/window/accessibility/secret/process checks、DEV-20-002 主题组合、DEV-20-006 feature-only Context 注入、DEV-20-007/011 固定 phase 归因、DEV-20-008 recovery-only Session 清理、DEV-20-009 migration baseline、DEV-20-010 fixture bootstrap marker | T-10 product、现有 sidecar build、production verifiers、T-11 theme controller、既有 `desktop_logout` 与产品 migration composition | 修改 production migration/bootstrap/setup/Context/配置、签名、公证、Windows native E2E |

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
- **只读上下文：** DEV-20-006 条件编译接缝之外的 Tauri product code、WebView、sidecar build。
- **共享路径：** Desktop E2E tree 由 T-20 唯一拥有；DEV-20-002 临时开放 test-only native App、ProductWorkspace/App Shell manifest/lock importer；DEV-20-006 只开放 main.rs 的 feature-only Context 注入，production Desktop entry 仍只读且归 T-10。
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
