---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-20
title: 建立 macOS Desktop 原生端到端门禁
status: ready
planning_depth: deep
planning_depth_reason: required native E2E 跨 Tauri、Go sidecar、SQLite、Keychain、原生窗口和进程清理安全边界
ready: true
risk: high
blocked_by: [T-10, T-14, T-15, T-16, T-19]
contract_ids: [AC-004, AC-005, AC-016, AC-017, AC-018, AC-031, AC-036]
owner: unassigned
expected_changes: ["<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>scripts/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-build.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-production.mjs</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>", "<Path>scripts/e2e/desktop/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-build.mjs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-production.mjs</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/**</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/**</Path>", "<Path>release/shared/sidecar/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/tests/e2e/desktop/**</Path> => T-20"]
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

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| native runner、first setup/restart/window/accessibility/secret/process checks | T-10 product、现有 sidecar build、production verifiers | 修改产品 main.rs、签名、公证、Windows native E2E |

## 4. 要构建什么

Lead 在 macOS parent-candidate/current-workspace 构建 production-like Desktop，使用临时 data/log root 和测试 Keychain 启动真实窗口。自动化完成首次设置、工作区导航、退出/登录、重启和持久化检查，同时验证 loopback、vault、capability、窗口尺寸、键盘可达和 sidecar 清理。

## 5. 实现契约

- **入口或接缝：** required Desktop native E2E command、Tauri binary、macOS accessibility/Keychain/process tools。
- **输入与输出：** clean temp roots/test keychain；输出脱敏 phase results 和 artifact verification。
- **公共接口变化：** 无产品变化；native Gate 强化。
- **不变量：** production binary 无 test control；Keychain/data/process 全清理；不修改真实用户数据。
- **状态或数据流：** build -> launch empty -> setup/session -> workspace -> restart -> verify -> cleanup。
- **错误与失败行为：** opt-in 缺失、非 macOS、window timeout、secret leak、sidecar leak 均 required fail。
- **兼容要求：** 删除 Skip 成功语义，不要求签名/公证。
- **安全与隐私要求：** diagnostics 不含 password/token/control nonce/paths；测试 Keychain 独立并删除。

## 6. 执行路线

1. 强化 runner self-tests 与缺环境/Skip/cleanup 反向验证。
2. 加入 empty DB first setup、部分成功恢复和工作区场景。
3. 加入重启/SQLite/vault/route/目标窗口/accessibility 场景。
4. 验证 production asset、capability、loopback 和进程/Keychain 清理。
5. 在 macOS candidate 运行完整 native required Gate 并核对 Evidence。

## 7. 路径访问契约

- **预计修改点/可写范围：** Desktop E2E、专用 scripts 和 production verifiers。
- **只读上下文：** main.rs、Tauri product code、WebView、sidecar build。
- **共享路径：** Desktop E2E tree 由 T-20 唯一拥有；main.rs 只读且归 T-10。
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

