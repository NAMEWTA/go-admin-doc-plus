---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-19
title: 建立真实 Web 浏览器端到端门禁
status: in_progress
planning_depth: deep
planning_depth_reason: required E2E 跨真实后端、双数据库、浏览器多标签、权限、路由和文件生命周期，是最终 Web 行为证据
ready: true
risk: high
blocked_by: [T-13, T-14, T-15, T-16, T-18]
contract_ids: [AC-011, AC-013, AC-014, AC-015, AC-016, AC-017, AC-018, AC-036, AC-038]
owner: codex-root
expected_changes: ["<Path>go-admin-plus-ui/tests/e2e/**</Path>", "<Path>scripts/e2e/web/**</Path>", "<Path>go-admin-plus/test/e2e/**</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/tests/e2e/**</Path>", "<Path>scripts/e2e/web/**</Path>", "<Path>go-admin-plus/test/e2e/**</Path>"]
read_only_paths: ["<Path>go-admin-plus-ui/apps/admin-web/**</Path>", "<Path>go-admin-plus-ui/packages/**</Path>", "<Path>go-admin-plus/internal/**</Path>", "<Path>.github/workflows/ci.yml</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/tests/e2e/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/tests/e2e/**</Path> => T-19"]
---

# Ticket T-19: 建立真实 Web 浏览器端到端门禁

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/19-web-browser-e2e.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-19.md</Path>`

## 1. 战略与来源

- **目标：** 在全部编码、静态检查、单元/集成和构建完成后，用真实浏览器证明 Web 产品闭环。
- **可观察产出：** SQLite/PostgreSQL 后端与真实 Web 覆盖 Bootstrap/login/routes/permissions/CRUD/双标签 Session/files；required 环境缺失失败。
- **来源：** `US-005~008`、`US-016~017`、frontmatter AC、`LOG-006`、`PLAN/P7`。
- **当前事实：** 已有多个 CDP browser harness，但部分需 opt-in、覆盖分散，且未验证统一产品 Shell 与双标签 Session。
- **Planning Depth 原因：** 这是跨边界外部行为的发布证据，不能由组件或 harness compile 替代。

## 2. 决策状态

### 已锁定决策

- 真实后端、真实浏览器；核心流程覆盖 Server SQLite/PostgreSQL。
- 覆盖 Bootstrap、登录、深链接/history、权限、CRUD、双标签 CSRF/renew、撤销、文件生命周期。
- required 缺 Chromium/DSN/环境或未执行必须失败。

### 已采用的低影响假设

- 复用现有 CDP runner 基础并抽取统一生命周期，避免引入第二套浏览器框架。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 统一 Web E2E runner/fixtures、双页场景、视口/截图、required failure | T-18 PG service、现有 CDP helpers、完成后的产品 | 修改产品代码、Desktop native（T-20）、签名 |

## 4. 要构建什么

Lead 在 parent-candidate/current-workspace 启动 disposable SQLite/PG 产品、完成安全 Bootstrap，再启动真实 Web/Chromium。场景从用户入口执行关键模块流程，打开两个标签验证稳定 CSRF/renew/撤销，并检查 route/视觉/键盘/错误。runner 总是清理子进程、临时 DB、浏览器 profile 和 secret。

## 5. 实现契约

- **入口或接缝：** required Web E2E command、Chromium CDP、真实产品 CLI。
- **输入与输出：** Chromium executable、disposable DSN、temp roots；输出场景级 pass/fail、trace/截图（脱敏）。
- **公共接口变化：** 无产品变化；测试 Gate 新增。
- **不变量：** E2E 只在 integrated candidate；缺环境失败；fixture secret 不进 artifact/log。
- **状态或数据流：** build -> migrate/bootstrap -> serve/web/browser -> scenarios -> cleanup/evidence。
- **错误与失败行为：** 子进程提前退出、timeout、Skip、cleanup leak、console error 均失败并分类。
- **兼容要求：** 删除 opt-in 缺失即成功和分散未汇总的 required 语义。
- **安全与隐私要求：** DSN/password/session/redacted paths 不写 screenshot/trace；环境 allowlist。

## 6. 执行路线

1. 统一现有 runner 生命周期/diagnostics，并建立缺环境/Skip/泄漏反向测试。
2. 建立 SQLite/PostgreSQL product setup 与 Bootstrap fixture。
3. 实现登录、route/history/权限/CRUD/模块代表流程。
4. 实现双标签 heartbeat/renew/CSRF/revoke 和文件 lifecycle 场景。
5. 执行目标视口、键盘/reduced-motion、cleanup 与完整 required Gate。

## 7. 路径访问契约

- **预计修改点/可写范围：** Web E2E tests/scripts 和专用后端 harness。
- **只读上下文：** 产品应用/包、CI workflow。
- **共享路径：** Web E2E tree 由 T-19 唯一拥有；T-20 仅写 Desktop 子树前需在 Goal Plan 明确例外或使用独立路径。
- **保留或不动：** 不为测试增加生产后门/测试 route，不修改业务断言使失败变绿。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | real browser/products | required Web E2E SQLite+PG | 所列用户流程通过 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-19.md</Path>` |
| 失败路径 | reverse runner | 缺 Chromium/DSN、host crash、timeout、leak | Gate 非零且清理完整 | 同上 |
| 回归 | candidate | `task test && task lint && task build TARGET=all PROFILE=server-sqlite` 后再 E2E | 非 E2E 与 E2E 都通过 | 同上 |

- **Workspace checks：** source-worktree 仅运行 runner unit/typecheck；不接受 source E2E pass。
- **E2E disposition：** required：真实 Web、双方言、双标签、route/权限/files 是本 Ticket 唯一交付。
- **E2E owner/environment：** Lead / parent-candidate（required policy）或 current-workspace（current policy）；场景覆盖 AC-011、013~018、036、038。
- **Integration evidence：** source commit、parent before、candidate/result SHA、E2E logs/traces、父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 所有产品/UI tickets + T-18 -> candidate -> required Web E2E -> parent result。
- **兼容窗口：** 无旧 runner 成功语义；required 命令原子接管。
- **监控信号：** scenario duration/status、browser console、child cleanup、executed profile count。
- **回滚或前向恢复：** 测试失败不推进父分支；修复产品或 runner 后重建 fresh candidate。
- **不可逆操作与批准点：** 只使用 disposable DB/files；禁止连接非 disposable DSN，执行前由 Lead 验证标记。
- **收缩条件：** required runner 无 Skip/opt-in exit 0，全部指定场景有非空执行证据。

## 10. 验收标准

- [ ] `AC-011、AC-013~018、AC-036、AC-038` 在真实浏览器/候选成立。
- [ ] SQLite/PostgreSQL、双标签、路由、权限、CRUD、文件场景实际执行且无资源泄漏。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-19.md</Path>`。
- [ ] required E2E、candidate/result、父分支包含和 Lead 双轴审查完整。
