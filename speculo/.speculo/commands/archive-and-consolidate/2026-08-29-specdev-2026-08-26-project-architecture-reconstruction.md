# Archive And Consolidate Execution Report

- **mode:** `executed`
- **scope:** `archive-single`
- **workflow:** `specdev`
- **change:** `2026-08-26-project-architecture-reconstruction`
- **generated_at:** `2026-08-29T12:20:00+08:00`
- **executed_at:** `2026-08-29T12:22:07+08:00`
- **user_confirmation:** `confirmed`（用户于 2026-08-29 明确回复“执行”）
- **final_result:** `verified`
- **completion commit:** `9cab3aa`
- **remote reconcile:** `main` 与 `origin/main` 一致，ahead/behind 为 `0/0`

> 本报告先以 dry-run 固化执行集合，随后依据用户对该固定计划的明确确认完成归档、知识提升、清理和验证。

## 1. Path Context

| Root | Resolved Path |
|---|---|
| workflow | `<Path>{roots.workflows}/specdev/</Path>` |
| state | `<Path>{roots.state}/specdev/</Path>` |
| changes | `<Path>{roots.state}/specdev/changes/</Path>` |
| archive | `<Path>{roots.state}/specdev/archive/</Path>` |
| commands report | `<Path>{roots.state}/commands/archive-and-consolidate/2026-08-29-specdev-2026-08-26-project-architecture-reconstruction.md</Path>` |
| permanent ADR | `<Path>{roots.state}/specdev/adr/</Path>` |
| permanent context | `<Path>{roots.state}/specdev/context/</Path>` |
| permanent research | `<Path>{roots.state}/specdev/research/</Path>` |

所有解析路径均位于项目根和对应 state namespace 内；没有符号链接逃逸。

## 2. Archive Plan

### Preflight

| Check | Result |
|---|---|
| change 名称符合日期 kebab 规则 | pass |
| change `.status.json` 可解析且为 `completed` | pass |
| 21/21 Ticket 为 `done` | pass |
| 21/21 worktree 为 `removed`，source/result 已进入 `main` | pass |
| blocker / 未批准 deviation / active candidate | 0 / 0 / 0 |
| `--stage complete` | `0 error / 0 warning` |
| Git root / local heads | 仅根 `main` / 仅 `main` |
| remote reconcile | `main@9cab3aa` 已推送，ahead/behind `0/0` |
| external reconcile | `not-applicable`；change 没有远程 Issue 来源，也没有 triage 文件 |
| source | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/</Path>` 存在 |
| target | `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/</Path>` 不存在 |
| global status | change 唯一位于 `active`，不在 `archived` |

### Move

| # | Source | Target | Action | Risk | Status |
|---|---|---|---|---|---|
| 1 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/</Path>` | `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/</Path>` | 原子移动完整 change | medium；路径移动会使旧 active locator 失效，但 Git 可恢复 | ready |

### State Changes

- 从 `<Path>{roots.state}/specdev/status.json</Path>` 的 `active` 删除该 change，并去重追加到 `archived`。
- 归档后的 `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/.status.json</Path>` 更新为 `change_status: archived`、`current_work: null`、`archived: true`，并写入规范项目相对 `archive_path`。
- 将 `specdev/archive-and-consolidate` 去重加入归档 change 的 `works_run`。
- 保留 `<Path>{roots.state}/specdev/changes/</Path>` 目录，不删除 namespace。

## 3. Knowledge Stores

| Store | Current State | Planned State |
|---|---|---|
| `<Path>{roots.state}/specdev/adr/</Path>` | 仅 `.gitkeep` | 创建 21 份 Accepted ADR |
| `<Path>{roots.state}/specdev/context/</Path>` | 仅 `.gitkeep` | 创建 1 份 61-term 当前架构词汇表 |
| `<Path>{roots.state}/specdev/research/</Path>` | 仅 `.gitkeep` | 保持不变；调研证据随 change 归档 |

## 4. ADR Graduation Plan

序号从空 store 的 `0001` 开始。每份永久 ADR 使用当前实现作为事实裁决，并标注来源 change `2026-08-26-project-architecture-reconstruction` 与日期 `2026-08-29`。

| # | Source | Target | Decision | Graduation | Current-truth Adjustment |
|---|---|---|---|---|---|
| 1 | change ADR-001 | `<Path>{roots.state}/specdev/adr/0001-zero-compatibility.md</Path>` | 新架构不保留旧内部模式兼容 | stable mechanism / must-know | 保持；零兼容扫描已实现 |
| 2 | change ADR-002 | `<Path>{roots.state}/specdev/adr/0002-product-directory-names.md</Path>` | 固定 `go-admin-plus` / `go-admin-plus-ui` | must-know | 保持当前物理目录名 |
| 3 | change ADR-003 | `<Path>{roots.state}/specdev/adr/0003-pnpm-workspace-boundaries.md</Path>` | 真实 pnpm Workspace 与可执行包边界 | stable mechanism | 使用当前 25-workspace/24 package contract |
| 4 | change ADR-004 | `<Path>{roots.state}/specdev/adr/0004-greenfield-external-contracts.md</Path>` | Greenfield API/schema/config/data 边界 | stable mechanism / must-know | 保持；不补旧数据/API 兼容 |
| 5 | change ADR-005 | `<Path>{roots.state}/specdev/adr/0005-root-asset-lifecycle-ownership.md</Path>` | 根级资产按 Git/deploy/release/database 生命周期分域 | stable mechanism | 使用当前根资产布局 |
| 6 | change ADR-006 | `<Path>{roots.state}/specdev/adr/0006-go-modular-monolith.md</Path>` | Go 原生模块化单体 | stable mechanism | 记录当前 app/application/host/modules/contracts/platform 分责 |
| 7 | change ADR-007 | `<Path>{roots.state}/specdev/adr/0007-tauri2-go-sidecar.md</Path>` | Tauri 2 管理 Go sidecar | stable mechanism / must-know | 使用当前 native host、Stronghold 与 bounded proxy 合同 |
| 8 | change ADR-008 | `<Path>{roots.state}/specdev/adr/0008-dual-app-single-business-composition.md</Path>` | Web/Desktop 双 App、单一业务组合 | stable mechanism | 使用当前 product composition 与双 adapter |
| 9 | change ADR-009 | `<Path>{roots.state}/specdev/adr/0009-business-capability-modules.md</Path>` | 八个业务能力模块 | stable mechanism | 运维现状改为 `internal/application/health` 与 `internal/host`；不提升已删除的 platform observability 设计 |
| 10 | change ADR-010 | `<Path>{roots.state}/specdev/adr/0010-release-target-matrix.md</Path>` | Linux OCI、macOS Universal、Windows x64 发行能力矩阵 | stable mechanism | 保留发行实现；个人自用不要求签名、公证或受保护安装，公开发行时再启用 protected gate |
| 11 | change ADR-011 | `<Path>{roots.state}/specdev/adr/0011-headless-domain-vue-presentation.md</Path>` | 无头 Domain 与 Vue Web Domain 分离 | stable mechanism | 使用当前 package exports/dependency rules |
| 12 | change ADR-012 | `<Path>{roots.state}/specdev/adr/0012-consumer-ports-integration-events.md</Path>` | 消费者 Port 与 Integration Event | stable mechanism / must-know | 加入当前 app-owned adapter 与 table-owner gate |
| 13 | change ADR-013 | `<Path>{roots.state}/specdev/adr/0013-database-profile-matrix.md</Path>` | Server PostgreSQL/SQLite，Desktop SQLite | stable mechanism / must-know | 保持三 profile 与 SQLite 单实例约束 |
| 14 | change ADR-014 | `<Path>{roots.state}/specdev/adr/0014-root-task-command-plane.md</Path>` | 根 Taskfile 是产品命令唯一入口 | stable mechanism | 脚本路径使用当前 `scripts/go-admin-plus`、`scripts/go-admin-plus-ui`、`scripts/contracts`、`scripts/quality` |
| 15 | change ADR-015 | `<Path>{roots.state}/specdev/adr/0015-openapi-contract-source.md</Path>` | OpenAPI 3.1 是跨端合同唯一事实源 | stable mechanism | 保持 canonical generator / strict Go / TS client 合同 |
| 16 | change ADR-016 | `<Path>{roots.state}/specdev/adr/0016-bun-goose-persistence.md</Path>` | Bun + Goose 双方言持久化 | stable mechanism | 加入 migration/table ownership 和 command/product lifecycle 现状 |
| 17 | change ADR-017 | `<Path>{roots.state}/specdev/adr/0017-no-tenancy.md</Path>` | 租户能力与数据模型零保留 | stable mechanism / must-know | 保持全仓/schema 零命中合同 |
| 18 | change ADR-018 | `<Path>{roots.state}/specdev/adr/0018-opaque-database-sessions.md</Path>` | 数据库不透明 Session 取代 JWT/Casbin | stable mechanism / must-know | 使用 Web Cookie/CSRF 与 Desktop Stronghold proxy 当前合同 |
| 19 | change ADR-019 | `<Path>{roots.state}/specdev/adr/0019-no-redis-database-coordination.md</Path>` | Redis 零保留与数据库可靠协调 | stable mechanism / must-know | 保持 Outbox、advisory lock、SQLite 单实例语义 |
| 20 | change ADR-020 | `<Path>{roots.state}/specdev/adr/0020-typed-immutable-profile-config.md</Path>` | 三 Profile 不可变类型化配置与 secret reference | stable mechanism | 使用当前 product-owned Host/operation 注入边界 |
| 21 | change ADR-022 | `<Path>{roots.state}/specdev/adr/0021-risk-tiered-quality-gates.md</Path>` | Local/PR/Protected Release 风险分层门禁 | stable mechanism | Protected Release 只约束需要公开发行的候选；个人自用完成不要求签名/公证/受保护安装 |

没有 supersede 目标：永久 ADR store 目前为空。所有动作均为 `create`，风险为 low/medium；第 9、10、14、21 项包含基于最终实现和用户完成决定的当前真相收敛，需随整体计划确认。

## 5. Context Graduation Plan

| Source | Target | Action | Content | Graduation | Risk |
|---|---|---|---|---|---|
| change `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/CONTEXT.md</Path>` | `<Path>{roots.state}/specdev/context/product-architecture.md</Path>` | create | 合并全部 61 个项目专有术语，按 repository、backend、frontend、runtime/security、delivery/quality 分组 | stable mechanism / must-know | medium |

词汇表不复制历史叙事；定义按当前代码校正 app/application/host、运维端点、根脚本和个人自用发行边界。原 `_Avoid_` 语义保留；没有现有术语冲突。

## 6. Research And Ephemeral Disposition

| Source Knowledge | Disposition | Reason |
|---|---|---|
| change ADR-021：目标垂直切片与最终原子切换 | ephemeral | 施工策略已执行完毕，不是归档后的运行架构约束 |
| LOG-001 至 LOG-023 | ephemeral | 决策已由毕业 ADR 去重；轮次、选项和访谈过程留在归档 |
| R-001 至 R-020 | ephemeral | 一手来源仍随归档保留；稳定结论已进入 ADR，避免 research/ 与 ADR 重复 |
| Ticket、Evidence、失败候选、调试和性能时长 | ephemeral | 属于可审计交付历史，脱离 change 上下文会误导 |
| Goal Plan、Spec、Tickets Map、Design Tree | ephemeral | 作为完整历史随 change 原样归档，不提升为当前知识副本 |

`<Path>{roots.state}/specdev/research/</Path>` 本轮不创建新文件。

## 7. Cleanup Candidates

| # | Path | Classification | Planned Action | Rationale | Risk |
|---|---|---|---|---|---|
| 1 | `<Path>{roots.state}/specdev/adr/.gitkeep</Path>` | delete | 创建 ADR 后删除 | 目录不再为空 | low |
| 2 | `<Path>{roots.state}/specdev/context/.gitkeep</Path>` | delete | 创建 context 后删除 | 目录不再为空 | low |
| 3 | `<Path>{roots.state}/specdev/research/.gitkeep</Path>` | keep | 无动作 | research store 保持空目录 | none |

现有知识 store 没有 ADR/context 内容，因此没有 merge、rewrite、supersede、重复、孤立术语或 conflict 候选；`needs-confirmation` 冲突数为 0。

## 8. Confirmed Execution Order

以下顺序已获得用户明确批准并执行：

1. 重检 source/target、completed 状态、Git clean、远端一致性和知识 store drift；任一变化即停止。
2. 创建 `<Path>{roots.state}/specdev/archive/2026-08/</Path>` 并原子移动完整 change。
3. 更新全局 status 与归档 change status/works_run。
4. 创建 21 份 ADR 与 1 份 context 词汇表，删除两份已批准 `.gitkeep`。
5. 重读 source、target、全局状态、归档状态和所有知识文件；确认 active/archived 无重叠。
6. 对归档目标运行 SpecDev `--stage complete`，并运行 SpecDev package self-check、docs/governance/architecture/compatibility 门禁。
7. 将执行结果和验证补遗追加到本报告，把 `mode` 更新为 `executed`、`user_confirmation` 更新为 `confirmed`。
8. 提交 `docs(specdev): archive architecture reconstruction` 并推送 `origin/main`；推送失败不回滚本地归档，但必须记录 reconcile 状态。

## 9. Summary

| Item | Count |
|---|---:|
| changes to archive | 1 |
| ADR files to create | 21 |
| context files to create | 1 |
| context terms to consolidate | 61 |
| research files to create | 0 |
| cleanup deletes | 2 |
| cleanup keeps | 1 |
| conflict-specific confirmations | 0 |
| blocked items | 0 |

**Dry-run verdict:** `ready-for-confirmation`。该 verdict 是执行前固定计划的历史记录，随后已按用户确认执行。

## 10. Execution And Verification

### Applied Result

| Item | Result |
|---|---|
| confirmation | `confirmed`；用户明确回复“执行” |
| source | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/</Path>` 已不存在 |
| archive target | `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/</Path>` 完整存在 |
| archive payload | 21 个 Ticket、22 份 Evidence，以及 Goal Plan、Spec、ADR、Context、Log、Design Tree、Tickets Map 和状态文件 |
| global status | `active: []`；`archived` 唯一包含本 change；两者无重叠 |
| archived status | `change_status: archived`、`current_work: null`、`archived: true`、`archive_path` 与目标一致 |
| ADR graduation | 创建 21 份 Accepted ADR；施工期 ADR-021 仅随 change 归档 |
| context graduation | 创建 1 份当前架构词汇表，共 61 个术语 |
| research graduation | 0；调研过程随 change 归档，永久 research store 保持空目录 |
| cleanup | 删除 ADR/context 两个失效 `.gitkeep`；保留 research `.gitkeep` |

### Validation Evidence

| Check | Result |
|---|---|
| execution drift preflight at `cb6be69` | pass；source/target、状态、知识 store 和 `origin/main` 均与 dry-run 一致 |
| pre-move SpecDev `--stage complete` | pass；`0 error / 0 warning` |
| post-move structural reread | pass；source absent、target complete、状态一致、21 Ticket、22 Evidence、21 ADR、61 terms |
| SpecDev package self-check | pass；`0 error / 0 warning` |
| docs check | `DOCS_CHECK_PASS` |
| governance check | `GOVERNANCE_CHECK_PASS` |
| architecture check | `ARCHITECTURE_CHECK_PASS` |
| zero compatibility check | `COMPATIBILITY_ZERO_PASS` |
| `git diff --check` | pass |

归档后再次调用 `--stage complete` 会得到预期的不适用结果：现有校验器将该 stage 硬编码为仅接受 `change_status=completed`，同时在尚未提交归档移动时将仓库判为 dirty；因此它不能作为 `change_status=archived` 的后置验证器。完成门已在原子移动前以 clean worktree 通过，移动后的真实性由归档结构、全局/局部状态、数量、JSON、路径和包级自检重读覆盖。

### Git Reconcile

归档内容和本执行报告已由提交 `9280e36`（`docs(specdev): archive architecture reconstruction`）推送至 `origin/main`，远端更新范围为 `cb6be69..9280e36`。该同步结果于 2026-08-29 验证成功；本段 reconcile 补遗以独立文档提交继续同步，最终提交号以用户交付记录为准。
