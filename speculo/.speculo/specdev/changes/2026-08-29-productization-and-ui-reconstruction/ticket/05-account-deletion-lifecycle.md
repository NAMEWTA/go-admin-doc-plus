---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-05
title: 建立账号 Tombstone 与文件处置生命周期
status: done
planning_depth: deep
planning_depth_reason: 涉及不可逆账号/文件删除、跨模块事件、双方言迁移和系统管理员存续安全不变量
ready: true
risk: critical
blocked_by: [T-02, T-04]
contract_ids: [AC-013, AC-023, AC-024, AC-025, AC-026]
owner: codex-root
expected_changes: ["<Path>go-admin-plus/internal/modules/iam/administration/deletion*</Path>", "<Path>go-admin-plus/internal/modules/files/account_lifecycle*</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0060-account-lifecycle/**</Path>", "<Path>go-admin-plus/test/iam/authorization/account_deletion*</Path>", "<Path>go-admin-plus/test/files/account_lifecycle*</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/modules/iam/administration/deletion*</Path>", "<Path>go-admin-plus/internal/modules/files/account_lifecycle*</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0060-account-lifecycle/**</Path>", "<Path>go-admin-plus/test/iam/authorization/account_deletion*</Path>", "<Path>go-admin-plus/test/files/account_lifecycle*</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/administration/service.go</Path>", "<Path>go-admin-plus/internal/modules/files/service.go</Path>", "<Path>go-admin-plus/internal/platform/outbox/**</Path>", "<Path>contracts/openapi/modules/iam-administration.yaml</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-05: 建立账号 Tombstone 与文件处置生命周期

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/05-account-deletion-lifecycle.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-05.md</Path>`

## 1. 战略与来源

- **目标：** 用可观察异步状态机替代账号硬删除，使 Files owner 与 Audit actor 始终可追踪。
- **可观察产出：** 删除必须显式 transfer/purge；账号立即失效；claim 前可取消；完成后匿名化且审计引用稳定。
- **来源：** `US-006`、`US-011`、`AC-013`、`AC-023~026`、`ADR-002`、`ADR-009`、`ADR-014`。
- **当前事实：** IAM 当前直接硬删除账号，Files owner 无跨模块生命周期；Outbox 已提供可靠事件基础。
- **Planning Depth 原因：** 永久删除和跨模块最终一致性具有不可逆数据风险。

## 2. 决策状态

### 已锁定决策

- active/disabled -> deletion-pending -> deleted；进入 pending 立即撤销 Session。
- transfer 需要有效目标；purge 二次确认且 worker claim 后不可取消/恢复。
- IAM 只发布 versioned event，不访问 Files 私表；始终保留至少一个启用系统管理员。

### 已采用的低影响假设

- 事件业务键使用 deletion ID 并包含策略版本；重试退避复用现有 outbox policy。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 0060 migration、IAM deletion use case、Files 幂等消费者、取消/净化 | Transactional Outbox、Session revoke、稳定 Audit ID | OpenAPI/UI（T-08/T-14）、隐藏回收站、批量账号删除 |

## 4. 要构建什么

管理员提交单账号删除命令时，IAM 先检查最后管理员不变量和处置参数，再使账号不可登录并写出可靠事件。Files 根据 transfer 或 purge 幂等执行并报告状态；queued 可取消，claimed 后取消冲突。只有消费者完成后 IAM 才净化个人字段并保留稳定匿名审计引用。

## 5. 实现契约

- **入口或接缝：** StartDeletion、GetDeletion、CancelDeletion；versioned integration event/consumer。
- **输入与输出：** user ID、transfer target 或 purge confirmation；输出 queued/claimed/completed/failed 与稳定 deletion ID。
- **公共接口变化：** 模块行为先建立；HTTP/OpenAPI 在 T-08。
- **不变量：** 最后启用管理员不可删除；deleted 不可恢复/登录/作为目标；IAM 不写 Files 表。
- **状态或数据流：** pending + outbox -> Files claim -> transfer/purge -> completion -> IAM anonymize。
- **错误与失败行为：** 无策略、自转移、无效目标、最后管理员、claim 后取消均稳定拒绝且不静默换策略。
- **兼容要求：** 最终删除旧硬删用例和 batch-delete，不保留兼容层。
- **安全与隐私要求：** purge 需显式权限/二次确认；匿名化字段不可由审计反推出敏感值。

## 6. 执行路线

1. 建立状态机、最后管理员、转移、purge 竞态与崩溃恢复红灯测试。
2. 增加 0060 双方言 migration 和 deletion/event 模型。
3. 实现 IAM 命令、Session revoke、审计事实与 Outbox 原子写入。
4. 实现 Files 幂等 transfer/purge consumer、claim 和完成回报。
5. 实现取消、重试、最终净化并运行故障注入回归。

## 7. 路径访问契约

- **预计修改点/可写范围：** deletion/account_lifecycle 新文件、0060 migration 与定向测试。
- **只读上下文：** 现有 services、Outbox 和 canonical OpenAPI。
- **共享路径：** 无；transport/生成物由 T-08，页面由 T-14。
- **保留或不动：** Files 现有 upload service 由 T-06；Audit 记录不物理删除。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | IAM/Files/outbox | transfer 与 purge 双方言集成 | 状态最终完成且引用正确 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-05.md</Path>` |
| 失败路径 | 竞态/崩溃 | 最后管理员、无效目标、claim 取消、重复事件、部分删除 | 稳定拒绝或幂等恢复 | 同上 |
| 回归 | IAM/Files | 定向 Go tests + outbox fault tests | 无孤儿资源，现有下载/删除正常 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行双方言、race、fault injection、architecture checks。
- **E2E disposition：** not-required：不可逆状态机由故障注入直接证明；管理交互和真实文件流程集中在 T-14/T-19。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；本 Ticket 不在 source 声明 E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** 0060 -> producer/consumer -> T-08 HTTP -> T-14 UI；部署前 worker 必须认识事件版本。
- **兼容窗口：** 内部事件可先扩展后启用生产者；旧硬删在 T-08 原子收缩。
- **监控信号：** queued/claimed/failed 年龄、重试数、孤儿 owner 查询、最后管理员拒绝。
- **回滚或前向恢复：** claim 前取消；claim 后仅前向重试/人工恢复备份，不承诺撤销物理删除。
- **不可逆操作与批准点：** purge 二次确认和 worker claim 是不可逆批准边界；Deep 实施需 Goal Plan 授权。
- **收缩条件：** 旧 delete/batch-delete 调用为零，所有 pending 可由新 worker 解释。

## 10. 验收标准

- [x] `AC-013、AC-023~026` 在双方言和故障注入中成立。
- [x] 不产生孤儿 owner，最后启用管理员不变量与 purge 不可逆边界成立。
- [x] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-05.md</Path>`。
- [x] 路径、commit、集成、父分支 result 和 E2E disposition 完整。
