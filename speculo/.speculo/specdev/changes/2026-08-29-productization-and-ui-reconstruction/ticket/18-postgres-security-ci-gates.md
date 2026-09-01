---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-18
title: 建立真实 PostgreSQL 与安全供应链 CI 门禁
status: in_progress
planning_depth: deep
planning_depth_reason: required CI 决定发布证据真实性，跨数据库服务、并发测试、漏洞/secret/SBOM 且缺环境必须失败
ready: true
risk: high
blocked_by: [T-01, T-09, T-17]
contract_ids: [AC-002, AC-006, AC-008, AC-019, AC-023, AC-027, AC-030, AC-035, AC-037]
owner: codex-root
expected_changes: ["<Path>.github/workflows/ci.yml</Path>", "<Path>scripts/ci/**</Path>", "<Path>scripts/security/**</Path>", "<Path>go-admin-plus/test/postgres/**</Path>", "<Path>go-admin-plus/test/files/dialect_contract_test.go</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/Cargo.lock</Path>", "<Path>scripts/quality/architecture-check.mjs</Path>", "<Path>scripts/quality/architecture-check.test.mjs</Path>"]
writable_paths: ["<Path>.github/workflows/ci.yml</Path>", "<Path>scripts/ci/**</Path>", "<Path>scripts/security/**</Path>", "<Path>go-admin-plus/test/postgres/**</Path>", "<Path>go-admin-plus/test/files/dialect_contract_test.go</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/Cargo.lock</Path>", "<Path>scripts/quality/architecture-check.mjs</Path>", "<Path>scripts/quality/architecture-check.test.mjs</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
# Approved deviation DEV-18-001 (USER-DECISION:all-approved): T-18 may update
# go-admin-plus/test/files/dialect_contract_test.go only to compose the current
# 0020-capacity migration provider in the existing dual-dialect Files contract
# fixture. Product service code, migration contents, assertions, fixtures, and
# required/non-skip semantics remain unchanged.
# Approved deviation DEV-18-002 (USER-DECISION:all-approved): T-18 may update
# go-admin-plus-ui/apps/admin-desktop/src-tauri/Cargo.lock only to resolve plist
# from 1.8.0 to 1.10.0 and remove vulnerable quick-xml 0.38.4. Cargo.toml,
# direct dependency pins, and unrelated lockfile resolutions remain unchanged.
shared_paths: ["<Path>.github/workflows/ci.yml</Path>", "<Path>scripts/ci/**</Path>", "<Path>scripts/security/**</Path>"]
shared_path_owners: ["<Path>.github/workflows/ci.yml</Path> => T-18", "<Path>scripts/ci/**</Path> => T-18", "<Path>scripts/security/**</Path> => T-18"]
---

# Ticket T-18: 建立真实 PostgreSQL 与安全供应链 CI 门禁

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/18-postgres-security-ci-gates.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-18.md</Path>`

## 1. 战略与来源

- **目标：** 让绿色 CI 真实证明 PostgreSQL 并发/迁移和供应链安全，不允许 Skip 或缺环境伪装通过。
- **可观察产出：** 健康 PostgreSQL service 注入一次性 DSN并执行 required tests；漏洞、依赖、secret、SBOM 与生成漂移成为明确门禁。
- **来源：** `US-016`、frontmatter AC、`PLAN/P7`。
- **当前事实：** backend CI 没有 PostgreSQL service/DSN，若干 integration/E2E runner 会 Skip；kin-openapi 漏洞由 T-01 修复。
- **Planning Depth 原因：** 这是发布证据的可信根，错误配置会让未测试行为进入候选。

## 2. 决策状态

### 已锁定决策

- required PostgreSQL job 使用固定版本/digest service、health check 和 disposable DSN。
- 缺 DSN、service 不健康、零目标测试或 Skip 必须失败。
- govulncheck、pnpm production audit、Rust advisory/deny、secret scan、SBOM、generate drift 纳入 Gate。

### 已采用的低影响假设

- 使用 GitHub Actions 原生 PostgreSQL service；具体安全扫描工具版本固定在 workflow/lock 中。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| required PG job、non-skip harness、security/SBOM jobs、CI architecture tests | root Taskfile、模块 PG tests、locked dependencies | 浏览器/native E2E（T-19/T-20）、发布/签名 |

## 4. 要构建什么

CI 启动隔离 PostgreSQL，等待健康后创建/传入 disposable DSN，并使用 require flag 执行 migration、Bootstrap、限流、scope、deletion、quota 和 runtime tests。每个 required runner输出机器可判定执行计数，Skip/零测试为失败。安全 jobs 对生产依赖和候选源码生成可审计报告。

## 5. 实现契约

- **入口或接缝：** `.github/workflows/ci.yml` jobs、CI wrapper、required test harness。
- **输入与输出：** ephemeral service/DSN/locked tools；输出非 Skip reports、SBOM 和 job status。
- **公共接口变化：** 无产品 API；CI contract 强化。
- **不变量：** required 缺环境失败；DSN 不输出；测试数据库一次性且不共享状态。
- **状态或数据流：** service health -> DSN -> migrate/tests -> count/assert -> reports/artifacts。
- **错误与失败行为：** health timeout、Skip、零目标、scan finding、drift 均非零退出且分类。
- **兼容要求：** 删除 opt-in 缺失即成功语义。
- **安全与隐私要求：** credentials 为 CI secret/ephemeral，不写日志/artifact；SBOM 无 secret。

## 6. 执行路线

1. 建立 workflow/static contract，证明缺 service/DSN/require 会失败。
2. 加入固定 PostgreSQL service、health 和 disposable DSN boundary。
3. 汇总 required PG suites 并强制执行/Skip 计数。
4. 加入 Go/pnpm/Rust/secret/SBOM/generate security jobs。
5. 运行 workflow policy tests、可用本地扫描及一次 CI 候选验证。

## 7. 路径访问契约

- **预计修改点/可写范围：** CI workflow、CI/security scripts、PG aggregator tests、architecture assertions。
- **只读上下文：** Taskfile、产品 scripts 与 lockfiles。
- **共享路径：** CI/security 路径仅 T-18 可写。
- **保留或不动：** 不修改产品代码来迎合测试；不把 DSN 写入 repo。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | required PG CI | service healthy + full PG suite | 目标计数非零且全部通过 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-18.md</Path>` |
| 失败路径 | reverse gate | 缺 DSN/require、模拟 Skip、scan finding | job 非零且分类准确 | 同上 |
| 回归 | quality/security | task contracts + scans + generate check | 现有质量矩阵通过 | 同上 |

- **Workspace checks：** source-worktree/current-workspace 运行 workflow policy、CI wrappers、可用安全扫描；不把外部 provider 自报当最终 Evidence。
- **E2E disposition：** not-required：本 Ticket 是数据库/供应链 required Gate，不是浏览器或 native E2E；真实 E2E 由 T-19/T-20。
- **E2E owner/environment：** Lead / parent-candidate 或 current-workspace；Lead 核对实际 CI run，不接受 source-only 声明。
- **Integration evidence：** commit、candidate/direct-parent/result SHA、CI run locator、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-01/T-09/T-17 -> CI contract -> required run -> T-19/T-20/T-21。
- **兼容窗口：** 无可选 Skip 模式；required jobs 原子启用。
- **监控信号：** executed/skipped counts、service health、scan findings、SBOM/generate artifact status。
- **回滚或前向恢复：** CI infra 故障分类为环境失败但仍红；修复 workflow 后重跑，不 waive 产品门禁。
- **不可逆操作与批准点：** 无生产动作；新增 required status 前由 Lead 确认命令可重复。
- **收缩条件：** required runner 中 opt-in skip/缺环境 exit 0 扫描为零。

## 10. 验收标准

- [ ] frontmatter 所列 PG/security AC 由真实 service 和非 Skip reports 覆盖。
- [ ] 缺 DSN、Skip、零目标测试与 scan finding 的反向验证能使 Gate 失败。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-18.md</Path>`。
- [ ] shared owner、commit、CI integration/result 与 E2E disposition 完整。
