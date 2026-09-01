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

### 已批准执行偏差 DEV-20-012

- **触发事实：** DEV-20-011 attempt 1 `33520584846`（job `99898523529`，probe `b714d6a`）在真实 macOS 15.7.7 arm64 上通过 prebuild，随后在既有 30 秒 Demo page wait 后精确返回 `login-demo-workspace`。该后缀只观察到全局“账户菜单”；该控件存在于所有 authenticated product routes，不能证明既有 accessibility `click` 是否已激活导航，也不能排除 Demo route 已进入但 projection 仍 busy。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs`、`run.mjs` 与 `run.test.mjs`；三个既有 Demo 导航动作必须定位精确、enabled 的 `AXButton`，聚焦后执行标准 `AXPress` action。失败分类只可在既有固定枚举前检查同一按钮的 `AXARIACurrent=page` 与页面 `AXBusy` 固定属性，由 self-test 锁定名称、action 和顺序。形成新 source/candidate 后，只可使用 DEV-20-011 剩余 2 次有序 hosted attempt，不新增配额。
- **禁止扩大：** 不修改产品、API、capability、migration、fixture、workflow、30 秒上限、retry/sleep/skip/allow-failure；不输出任意 UI；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-013

- **触发事实：** DEV-20-011 attempt 2 `33524019742`（job `99910130934`，probe `1ae5049`）在真实 macOS 15.7.7 arm64 上通过 prebuild，随后仍返回 `login-demo-workspace`。DEV-20-012 在所有旧 suffix 之前检查 `AXARIACurrent=page`；该结果证明 exact Demo `AXButton` 的标准 `AXPress` 返回成功但 30 秒后 WebKit/Vue route 仍未激活。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs` 与 `run.test.mjs`；同一个 exact、enabled AXButton 在 frontmost process 聚焦后必须用 Enter `key code 36` 激活，复用已在登录 keyboard submit 中通过的机制。三个调用点、current/busy classifier、固定名称和所有 timeout 不变；self-test 必须拒绝该 helper 中的 AXPress、click 与 delay。形成新 source/candidate 后，只可使用 DEV-20-011 最后 1 次有序 hosted attempt，不新增配额。
- **禁止扩大：** 不修改产品、API、capability、migration、fixture、workflow、timeout/retry/sleep/skip/allow-failure；不输出任意 UI；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-014

- **触发事实：** DEV-20-011 第三次且最后一次 attempt `33526583893`（job `99918873988`，probe `915cf58`）在真实 macOS 15.7.7 arm64 上通过 prebuild，随后在既有 90 秒普通首次设置工作区等待后返回无后缀 `first-setup-workspace`。该 candidate 只改变更晚发生的 Demo 导航 helper 与 self-test，普通首次设置提交未变且已在多个更早 attempt 中通过；现有 phase 无法区分仍在 setup、进入 recovery、落到 login、runtime unavailable/loading 或 unknown。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/run.mjs` 与 `run.test.mjs`，在普通首次设置工作区等待失败后按固定字符串枚举 setup/recovery/login/unavailable/loading/unknown，并由 self-test 锁定分类集合与调用顺序。形成新 source/candidate 后，Lead 可更新同一 workflow-only probe descendant、push 同一具名 probe branch，并最多 dispatch 3 次逐个归因的 `macos-15` attempt。
- **禁止扩大：** 不输出任意 UI，不修改产品、API、capability、migration、fixture、workflow、90 秒上限、retry/sleep/skip/allow-failure；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。每个前一红灯必须先归因并修正，hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-015

- **触发事实：** DEV-20-014 attempt 1 `33528845263`（job `99926524466`，probe `20c6b63`）在真实 macOS 15.7.7 arm64 上通过 prebuild 与 ordinary first setup，证明前一无后缀红灯为非确定性结果，随后再次精确返回 `login-demo-workspace`。exact enabled Demo AXButton helper 聚焦后立即发送 Enter，而已在同一 hosted 路径通过的登录/表单 submit helper 在聚焦后固定等待 `0.2s` 再发送同一 `key code 36`。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs` 与 `run.test.mjs`，在 `pressButtonScript` 设置同一 exact enabled button focused 后、`key code 36` 前插入且只插入 `delay 0.2`；self-test 必须锁定顺序并拒绝其他 delay。形成新 source/candidate 后沿用 DEV-20-014 剩余 2 次有序 hosted attempt，不新增配额。
- **禁止扩大：** 不修改产品、API、capability、migration、fixture、workflow、调用点、30/90 秒上限、retry/skip/allow-failure 或其他 sleep/delay；不输出任意 UI；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-016

- **触发事实：** DEV-20-014 attempt 2 `33530698516`（job `99932815718`，probe `f062e11`）在真实 macOS 15.7.7 arm64 上通过 prebuild，随后仍精确返回 `login-demo-workspace`。同一个 exact enabled Demo AXButton 的 standard AXPress、element click、focus + Enter 以及与已通过 native submit 对齐的 `delay 0.2` focus synchronization 均未激活 WebKit/Vue route。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs` 与 `run.test.mjs`；`pressButtonScript` 继续定位同一 exact enabled AXButton，从其 AX position/size 计算中心点，并在既有 frontmost/focus synchronization 后只发送一次 System Events `click at`。self-test 必须锁定 position/size/center/single-click 顺序并拒绝 AXPress、`click currentElement` 与 `key code`。形成新 source/candidate 后使用 DEV-20-014 最后 1 次有序 hosted attempt，不新增配额。
- **禁止扩大：** 不修改产品、API、capability、migration、fixture、workflow、调用点、30/90 秒上限、retry/skip/allow-failure，不输出任意 UI 或坐标，不增加其他 click/sleep/delay；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。

### 已批准执行偏差 DEV-20-017

- **触发事实：** DEV-20-014 第三次且最后一次 attempt `33532649355`（job `99939316440`，probe `84e97c9`）在真实 macOS 15.7.7 arm64 上通过 prebuild，随后仍精确返回 `login-demo-workspace`。`ProductWorkspace` 只有一份导航 DOM，但 sidebar navigation 为 `height: calc(100vh - 64px); overflow: auto`；native window 为 `960x640`，每个 route/group 至少 50px，order 600 的 Demo 位于多个早期模块之后。AX tree 因而能找到 exact enabled 的屏外 Demo button，但 AX/keyboard/mouse 都无法激活其裁剪外位置。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs` 与 `run.test.mjs`；`pressButtonScript` 在同一 exact enabled AXButton 上先且只执行一次 `AXScrollToVisible`，用既有 `delay 0.2` 同步后再执行既有单次 center-point click。self-test 必须锁定 action/order、唯一 delay 与唯一 click。形成新 source/candidate 后，Lead 可更新同一 workflow-only probe descendant、push 同一具名 probe branch，并最多 dispatch 3 次逐个归因的 `macos-15` attempt。
- **禁止扩大：** 不修改产品、API、capability、migration、fixture、workflow、调用点、30/90 秒上限、retry/skip/allow-failure，不输出任意 UI 或坐标，不增加其他 click/action/sleep/delay；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `4f840d347c6b3210ef39370d5f59234421daf797` 已按边界实现；candidate `9b47a86ee8008e170f0446bc0c94eba0de78370f`（tree `f98ff829785d3eee292e993f912cb90f2907d30f`）包含治理父 `ffd15bd` 与 source，Node 48/48、Vitest 41/41 files / 256/256 tests、lint、Speculo 0/0、diff-check 与 clean tree 已通过。hosted attempt 1 为 Actions `33535248320`（job `99947849962`，probe `acd5a91`）：prebuild 通过，required gate 从 17:08:11Z 运行至 17:16:27Z 后仍返回 `login-demo-workspace`；`AXScrollToVisible` 未抛错但后续 center click 未激活 route，无 native pass marker。归因并修正前不使用剩余 2 次 attempt。

### 已批准执行偏差 DEV-20-018

- **触发事实：** DEV-20-017 attempt 1 的 `AXScrollToVisible` 未抛错，但 `pressButtonScript` 在 action 后继续使用滚动前 `entire contents` 扫描得到的 `currentElement` 完成 focus、position、size 与 center click；真实 route 后置条件仍返回 `login-demo-workspace`。滚动会改变 WebKit layout/AX geometry，当前脚本没有证明 click 使用了刷新后的可见元素引用。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs` 与 `run.test.mjs`；在既有唯一 `AXScrollToVisible` 后立即重新扫描一次 `entire contents of window 1`，重新取得同名 exact enabled AXButton，并只用该 refreshed element 执行既有 focus、`delay 0.2`、position/size/center 与唯一 click。self-test 必须锁定 scroll -> fresh query -> click 顺序，以及唯一 action/delay/click。形成新 source/candidate 并通过 portable checks 后，可使用 DEV-20-017 剩余 2 次中的下一次 ordered attempt。
- **禁止扩大：** 不修改产品、API、capability、migration、fixture、workflow、调用点、30/90 秒上限、retry/skip/allow-failure，不输出任意 UI 或坐标，不增加 attempt 配额、其他 click/action/delay/sleep；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `70db7a2773e336ab23c879e8c4ebb69d6f167118` 已实现 scroll 后 fresh query；candidate `2d302d215e59ad3e7f64ed62716005bb30da041d`（tree `5fde3a796d1dab6a94d8e6d720d176c4885b5ee9`）包含治理父 `b4a69f7` 与 source，Node 48/48、Vitest 41/41 files / 256/256 tests、lint、Speculo 0/0、diff-check 与 clean tree 已通过。hosted attempt 2 为 Actions `33537653264`（job `99955810105`，probe `b3244c0`）：prebuild 通过，required gate 从 17:31:54Z 运行至 17:39:53Z 后仍返回 `login-demo-workspace`；fresh query 排除了 stale reference，但 route 仍未激活，无 native pass marker。最后 1 次 attempt 在下一修正验证前不执行。

### 已批准执行偏差 DEV-20-019

- **触发事实：** DEV-20-017/018 两次 hosted attempt 中 `AXScrollToVisible` 均未抛错但 route 后置条件均为 `login-demo-workspace`；attempt 2 已在 action 后 fresh query exact enabled button，故 stale AX reference 不再能解释红灯。Apple AX 合同提供 `AXParent`、滚动容器的 `AXVerticalScrollBar`、可写 scroller `AXValue` 与 `AXMaxValue`，可以直接验证并移动真实 overflow container。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs` 与 `run.test.mjs`；移除 `AXScrollToVisible`，从同一 exact enabled Demo AXButton 沿 `AXParent` 在固定有限层数内找到 enclosing `AXScrollArea`，取得其 `AXVerticalScrollBar` 并把 value 唯一一次设为 `AXMaxValue`；随后保留现有 fresh exact-enabled-button query，只用 refreshed element 执行既有 focus、`delay 0.2`、position/size/center 与唯一 click。self-test 锁定 target -> bounded parent -> scrollbar max -> refresh -> click 顺序、唯一 value mutation/delay/click，并拒绝任何 `perform action`。新 source/candidate portable checks 通过后，可使用 DEV-20-017 最后 1 次 ordered attempt。
- **禁止扩大：** 不修改产品、API、capability、migration、fixture、workflow、调用点、30/90 秒上限、retry/skip/allow-failure，不输出任意 UI 或坐标，不增加 attempt 配额、其他 click/action/value mutation/delay/sleep；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `7c2e7c3fc0dae510c4d65fdb1e30f13c5baecc94` 已实现 bounded AXScrollArea/vertical-scrollbar maximum；candidate `1e175790a7f5ea598f0adb53b454f13f9fd7cd4e`（tree `90cc12f4d42b0247baf02809af9150d5e8ebe8e0`）包含治理父 `3ca4abf` 与 source，Node 48/48、Vitest 41/41 files / 256/256 tests、lint、Speculo 0/0、diff-check 与 clean tree 已通过。最后一次 ordered hosted attempt 为 Actions `33539914535`（job `99963330288`，probe `2f3e87d`）：macOS 15.7.7 arm64/AX enabled 上 prebuild 从 17:49:20Z 至 17:56:05Z 通过，required gate 从 17:56:05Z 至 18:05:04Z 后返回 `login-navigation`: `desktop native accessibility button unavailable`。该红灯位于滚动辅助函数内部，早于 Demo route 后置条件；无 skip、无 pass marker，DEV-20-017 的 3 次 attempt 已用尽。

### 已批准执行偏差 DEV-20-020

- **触发事实：** DEV-20-019 attempt 3 `33539914535`（job `99963330288`，probe `2f3e87d`）在真实 macOS 15.7.7 arm64 上通过 prebuild，随后于滚动辅助函数内部返回固定 `login-navigation`: `desktop native accessibility button unavailable`。既有统一错误不能识别具体 AX 子步骤；静态产品合同同时证明完整管理员导航中“用户管理”是第一个 route、“产品示例”是第十二个 route，组标题是不可聚焦的 `<p>`，二者之间恰有 11 个标准 Tab 步进。以可见 route 为起点的浏览器焦点遍历会使用 WebKit 自身的标准 scroll-into-view 行为，不依赖未暴露的 AXScrollArea/scrollbar 属性，也不需要修改既有 50px 导航视觉合同。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs`、`diagnostics.mjs` 与 `run.test.mjs`；`pressButtonScript` 删除 AX parent/scrollbar 路径，必须同时定位 exact enabled 的固定“用户管理”起点与目标“产品示例”，先聚焦可见起点，再且只执行固定 11 次 `key code 48`，沿用唯一 `delay 0.2` 后 fresh query 并要求目标为 focused，最后只用 fresh target 的 position/size 执行既有单次 center click。只可增加固定 `navigation-start`、`navigation-traversal`、`navigation-action` 安全诊断，不输出 UI、输入、路径、坐标或底层 AppleScript 文本。self-test 必须锁定完整顺序、11 次 Tab、两次 AX scan、唯一 delay/click，并拒绝 AXScrollToVisible、AXParent、AXScrollArea、scrollbar value、perform action、Enter 与额外 click。形成新 source/candidate 且 portable checks 通过后，Lead 可更新同一 workflow-only probe descendant、push 同一具名 probe branch，并最多 dispatch 3 次逐个归因的 `macos-15` attempt。
- **禁止扩大：** 不修改产品、manifest、API、capability、migration、fixture、workflow、调用点、50px 视觉合同、30/90 秒上限、retry/skip/allow-failure，不增加其他 key、click/action/delay/sleep；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。每个前一红灯必须先归因并修正，hosted failure 不得重标为 G7 通过。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `66d70030e0601f913d63b83f0205c06fc0f45299` 已实现固定可见起点、11 次 Tab、fresh focused target 与固定诊断；candidate `2582a4bc575bd941a9ecb47001ea92f8d083c8a8`（tree `7be9540499c38fa5764a13e59be96335ad3b5630`）包含治理父 `25c8ec1` 与 source 且无冲突，Vitest 41/41 files / 256/256 tests、Node 48/48、lint、diff-check 与 clean tree 已通过。hosted attempt 1 为 Actions `33542592499`（job `99972252465`，probe `c955fcb`）：prebuild 从 18:16:41Z 至 18:23:21Z 通过，required gate 从 18:23:21Z 至 18:31:57Z 后精确返回 `login-navigation`: `desktop native accessibility navigation-traversal unavailable`；起点/目标初始查询与 key 发送未抛错，但普通 Tab 后目标没有 focused。无 skip、无 pass marker，DEV-20-020 尚余 2 次 attempt。

### 已批准执行偏差 DEV-20-021

- **触发事实：** DEV-20-020 attempt 1 的固定 `navigation-traversal` 证明可见起点和目标存在、11 次普通 Tab 已发送，但目标未获得焦点。Apple 官方 Safari 键盘合同明确：普通 Tab 默认只高亮字段或弹出菜单；`Option-Tab` 高亮下一字段、弹出菜单或网页可点击项，Safari 的“Press Tab to highlight each item”设置会交换两者行为（`https://support.apple.com/guide/safari/cpsh003/mac`）。hosted runner 未开启全键盘导航与该结果一致。
- **批准范围：** T-20 只可修改 `tests/e2e/desktop/accessibility.mjs` 与 `run.test.mjs`，把 DEV-20-020 循环中的唯一 `key code 48` 改为 `key code 48 using option down`；起点、目标、固定 11 次、两次 AX scan、fresh focused 要求、唯一 `delay 0.2`、唯一 center click 与三类固定诊断全部不变。self-test 必须要求 Option 修饰并拒绝未修饰 Tab 或其他 key。新 source/candidate portable checks 通过后，只可继续使用 DEV-20-020 剩余 2 次 ordered attempt，不新增配额。
- **禁止扩大：** 不修改系统/Safari/Keyboard 设置、产品、manifest、visual contract、diagnostics、workflow、调用点、timeout/retry/skip/allow-failure，不增加其他 modifier/key/click/action/delay/sleep；不输出 UI、输入、路径、坐标或 AppleScript 文本；不发布 artifact、deploy/migrate、使用 production secrets、重写或清理远端历史。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `85d1f14c441771717c579b44d0b7c4adf7b52708` 已形成；candidate `d1b0d9c92d8e658064e7bea389bfc3d1ae8c473e`（tree `3d28be1020abf3892141ecb0e040eb8947e45133`）包含治理父 `71fadd9b8e3f00ebfbff4dac7bffb6a235a21142` 与 source，普通 merge 无冲突。Vitest 41/41 files / 256/256 tests、Node 48/48、lint、diff-check 与 clean tree 已通过；self-test 要求唯一 11 步循环使用 Option-Tab 并拒绝未修饰 Tab 或其他 key。DEV-20-020 尚余 2 次 ordered attempt，native Gate 仍为 pending。

### 已批准执行偏差 DEV-20-022

- **触发事实：** DEV-20-020 attempt 2 使用 probe `5717707`、Actions `33544806661`、job `99979569584`，在 macOS 15.7.7 arm64 与 AX enabled 环境通过 18:39:16Z~18:44:36Z prebuild，required gate 于 18:44:36Z~18:51:53Z 再次固定返回 `login-navigation`: `desktop native accessibility navigation-traversal unavailable`。这证明 Tauri WebView 的 AX focus traversal 不接受普通 Tab 或 Option-Tab；无 skip、无 pass marker，DEV-20-020 尚余最后 1 次 attempt。
- **批准范围：** T-20 可在现有 test-only `apps/admin-desktop/src/native-e2e/App.vue` 增加唯一固定可见的 `E2E open Demo` 按钮，点击时只把 hash router 目标设为 `#/demo/products`；native runner 的三处 Demo navigation 只点击该 test-only 按钮并继续验证同一真实 route guard、lazy page、权限、CRUD、Session 与重启后置条件。`verify-production.mjs` 必须把该文本加入 production byte 拒绝清单，self-test 必须锁定 exact hash、三处调用与 production App 零命中。删除已实测无效且不再调用的 sidebar AX traversal helper 与对应断言。portable candidate 全部通过后可使用 DEV-20-020 最后 1 次 ordered attempt。
- **禁止扩大：** 不修改 `ProductWorkspace`、manifest、router、产品路由/权限/视觉合同、Tauri capability/config、workflow、timeout/retry/skip/allow-failure，不增加后端 test action、公开 API、secret、任意脚本执行或生产资产中的测试控件；不发布 artifact、deploy/migrate、重写或清理远端历史。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `d2d4ed5552e244761d56c81f352729d7ed6bd7ab` 已形成；candidate `0847ff6a52762ab77dfbd9c0e2738c5e208ff143`（tree `96f8129fbd2e1d92b083eaad365cb9cf2c95f135`）包含治理父 `3056fdd33f717f7387e119f6d8d2853d16d847f6` 与 source，普通 merge 无冲突。candidate 已通过 Vitest 41/41 files / 256/256 tests、Node 48/48、Desktop runner 21/21、完整 typecheck、lint、production build/asset scan、diff-check 与 clean tree；native-e2e entry build 通过并确认测试标记存在，随后 production 重建再次通过零测试字节扫描。DEV-20-020 尚余最后 1 次 ordered attempt，native Gate 仍为 pending。

### 已批准执行偏差 DEV-20-023

- **触发事实：** DEV-20-020 attempt 3 使用 probe `0d30fd4`、Actions `33547554824`、job `99988721801`，在 macOS 15.7.7 arm64 与 AX enabled 环境通过 19:07:29Z~19:13:18Z prebuild，required gate 于 19:13:18Z~19:21:48Z 返回新的首个红灯 `theme-dark-toggle`: `desktop native accessibility button unavailable`。它已越过 login navigation 与 Demo page/boundary，直接证明 DEV-20-022 的固定 Demo 控件有效；DEV-20-020 的 3 次 attempt 已用尽，无 skip、无 pass marker。
- **批准范围：** T-20 可在同一 test-only native entry 增加唯一固定可见 `E2E use dark theme` 按钮；其 handler 只可 exact 查询产品 `button[aria-label="使用深色主题"]` 并调用 DOM `click()`，缺失时写固定 `E2E control failed: theme-dark`。runner 只把 `theme-dark-toggle` 的 AX 目标改为该 test-only 按钮，随后仍必须观察产品 aria-label `当前使用深色主题`、重启持久化与最终存储清理。production byte scanner 必须拒绝新按钮/失败标记，self-test 锁定 exact selector/click、产品 handler、runner 调用与 production App 零命中。新 source/candidate 通过 portable checks 后最多 dispatch 3 次逐项归因的同一 workflow-only probe。
- **禁止扩大：** 不直接写主题 localStorage/data-theme，不修改主题 controller、`ProductWorkspace`、router、产品视觉/交互、capability/config、workflow、timeout/retry/skip/allow-failure，不增加后端 action、公开 API、secret 或生产资产中的测试控件；不发布 artifact、deploy/migrate、重写或清理远端历史。每次首个真实红灯必须先记录与精确修正。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `1c285f8ccc241034f1b1f1e914ca0ac54bb84499` 已形成；candidate `be1f74a92ffd92121e314df27a86971c38f9bee6`（tree `8056abbcff1989ed3605fe28267e70ed9a319881`）包含治理父 `64c6154b74e949e9caa9ce2b7b48e2fb1f379a18` 与 source，普通 merge 无冲突。candidate 已通过 Vitest 41/41 files / 256/256 tests、Node 48/48、Desktop runner 21/21、lint、production build/asset scan、diff-check 与 clean tree；source 完整 typecheck 与 native-e2e build 通过，两项新标记存在，production 重建再次通过零测试字节扫描。DEV-20-023 尚余 3 次 ordered attempt，native Gate 仍为 pending。

### 已批准执行偏差 DEV-20-024

- **触发事实：** DEV-20-023 attempt 1 使用 probe `a356e77f9a810794806a8a208302c35fd7a4a1eb`、Actions `33549802770`、job `99996186434`，在 macOS 15.7.7 arm64 与 AX enabled 环境通过 19:30:16Z~19:34:19Z prebuild，required gate 于 19:34:19Z~19:38:48Z 返回 `first-setup-workspace`，无 skip、无 pass marker。runner 只会在 90 秒 workspace poll 失败时把该 phase 改为固定后缀；无后缀结果因此证明 `账户菜单` 已出现，首红实际来自紧随其后的 30 秒 authenticated boundary poll 沿用了旧 phase。它也证明 DEV-20-023 已越过先前的 theme target，但本轮在更早的已知非确定性 first-setup 路径退出。
- **批准范围：** T-20 只可在普通 first-setup workspace poll 成功后、现有 `pollBoundary` 调用前设置固定 `first-setup-boundary` phase，并在既有 Desktop runner self-test 中锁定 `workspace poll -> boundary phase -> pollBoundary -> stop` 的顺序。形成并验证新 source/candidate 后，继续使用 DEV-20-023 剩余 2 次 ordered attempt。
- **禁止扩大：** 不修改 `pollBoundary`、30/90 秒 timeout、retry/skip/allow-failure、产品 boundary 逻辑、first setup、Session、主题、router、capability/config 或 workflow；不输出任意 DOM、secret、路径或凭据，不发布 artifact、deploy/migrate、重写或清理远端历史。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `6999c642ebea740438233499d4b09df9ef26a2c1` 已形成；candidate `7c446eb31729b7e83906a2fb25d22b5636f4fccf`（tree `641b756fdf4671838b62d7621a7e89aafdfcb1ad`）包含治理父 `5fa466a272fa8fa6ba6d1d85507989cd731cb48f` 与 source，普通 merge 无冲突。candidate 已通过 Vitest 41/41 files / 256/256 tests、Node 48/48、Desktop runner 21/21、完整 typecheck、lint、native-e2e 资产标记检查、production build/asset scan、diff-check 与 clean tree。DEV-20-023 尚余 2 次 ordered attempt，native Gate 仍为 pending。

### 已批准执行偏差 DEV-20-025

- **触发事实：** DEV-20-023 attempt 2 使用 probe `7435ec4cfc0196be7cc45e8cb6bcb6f067dbcb38`、Actions `33551508113`、job `100001827771`，在 macOS 15.7.7 arm64 与 AX enabled 环境通过 19:48:00Z~19:54:13Z prebuild，required gate 于 19:54:13Z~20:03:00Z 返回新的首红 `permission-authorization`，无 skip、无 pass marker。它已越过 first-setup boundary、登录导航、Demo、主题切换与 scope self/all，证明 DEV-20-023/024 有效；当前单一 phase 同时覆盖关闭权限控件、403 control 后置条件、无权页、恢复权限控件和页面恢复，不能继续归因。
- **批准范围：** T-20 只可把现有 permission sequence 拆成五个固定 phase：`permission-disable-control`、`permission-denied-boundary`、`permission-hidden`、`permission-enable-control`、`permission-restored`；各 phase 必须紧邻原有 click/poll 调用，并由 Desktop runner self-test 锁定完整顺序。形成并验证新 source/candidate 后，继续使用 DEV-20-023 最后 1 次 ordered attempt。
- **禁止扩大：** 不修改任何 click/poll target、control action、产品权限逻辑、timeout/retry/skip/allow-failure、test-only UI、后端、Session、主题、router、capability/config 或 workflow；不输出任意 DOM、secret、路径或凭据，不发布 artifact、deploy/migrate、重写或清理远端历史。
- **批准来源：** 用户“都批准”及当前目标“相关的所需要批准的外部条件都批准”。
- **执行状态：** source `9562317868000618ce85f87f8f01a09764863fcb` 已形成；candidate `cc018bcba33a1a7f11ccf606b5367405d70f1236`（tree `5302b7eacf769eddc62b896268d5c0bd6af97252`）包含治理父 `b131dfe441d52281f0b3312c044f3e905ccb3079` 与 source，普通 merge 无冲突。candidate 已通过 Vitest 41/41 files / 256/256 tests、Node 49/49、Desktop runner 22/22、完整 typecheck、lint、native-e2e 资产标记检查、production build/asset scan、diff-check 与 clean tree。DEV-20-023 尚余 1 次 ordered attempt，native Gate 仍为 pending。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| native runner、first setup/restart/window/accessibility/secret/process checks、DEV-20-002 主题组合、DEV-20-006 feature-only Context 注入、DEV-20-007/011/012/013 固定 phase 与导航归因、DEV-20-008 recovery-only Session 清理、DEV-20-009 migration baseline、DEV-20-010 fixture bootstrap marker | T-10 product、现有 sidecar build、production verifiers、T-11 theme controller、既有 `desktop_logout` 与产品 migration composition | 修改 production migration/bootstrap/setup/Context/配置、签名、公证、Windows native E2E |

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
