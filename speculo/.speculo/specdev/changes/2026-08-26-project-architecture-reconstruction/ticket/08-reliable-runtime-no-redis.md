---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-08
title: 无 Redis 可靠运行时
status: review
planning_depth: deep
planning_depth_reason: 事务 Outbox、幂等、缓存正确性和多实例执行权涉及并发与可靠副作用
ready: true
risk: critical
blocked_by: [T-03, T-04]
contract_ids: [AC-018, AC-034]
owner: codex-t08-reliability
expected_changes: ["<Path>go-admin-plus/internal/platform/outbox/**</Path>", "<Path>go-admin-plus/internal/platform/coordination/**</Path>", "<Path>go-admin-plus/internal/platform/localcache/**</Path>", "<Path>go-admin-plus/internal/platform/migrations/reliable-runtime/**</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/platform/outbox/**</Path>", "<Path>go-admin-plus/internal/platform/coordination/**</Path>", "<Path>go-admin-plus/internal/platform/localcache/**</Path>", "<Path>go-admin-plus/internal/platform/migrations/reliable-runtime/**</Path>", "<Path>go-admin-plus/test/reliable-runtime/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/app/kernel/**</Path>", "<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus/internal/platform/migrations/**</Path>"]
shared_paths: ["<Path>go-admin-plus/internal/platform/outbox/**</Path>", "<Path>go-admin-plus/internal/platform/coordination/**</Path>", "<Path>go-admin-plus/internal/platform/localcache/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus/internal/platform/outbox/**</Path> => T-08", "<Path>go-admin-plus/internal/platform/coordination/**</Path> => T-08", "<Path>go-admin-plus/internal/platform/localcache/**</Path> => T-08"]
---

# Ticket T-08: 无 Redis 可靠运行时

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/08-reliable-runtime-no-redis.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-08.md</Path>`

## 1. 战略与来源

- **目标：** 用当前数据库和受限进程缓存替代 Redis 的 Session 外可靠协调能力。
- **可观察产出：** 业务事务与事件原子提交，失败可重试且消费幂等；PostgreSQL 唯一 executor 可接管，SQLite 不出现第二 executor。
- **来源：** `US-002`、`US-017`、`AC-018`、`AC-034`、`ADR-019`、`DEC-010`。
- **当前事实：** 旧 Redis/内存队列分散正确性状态，断连与多实例所有权缺少统一合同。
- **Planning Depth 原因：** 重复或丢失副作用会破坏业务和调度正确性。

## 2. 决策状态

### 已锁定决策

- 状态与 Integration Event 同事务写入 Outbox；消费者按稳定业务键幂等。
- PostgreSQL advisory lock 协调唯一 worker；SQLite 依赖 T-04 单实例。
- 本地缓存可随时清空/禁用，仅允许性能变化。

### 已采用的低影响假设

- claim 使用租约/超时语义，失败恢复为可重试状态。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| Outbox、dispatcher、幂等接缝、advisory lock、本地缓存 | T-03 lifecycle、T-04 transaction | 具体业务事件消费者、外部消息代理 |

## 4. 要构建什么

业务用例在同一事务写状态和事件；worker 取得执行权后 claim、分发并记录结果，断连立即停止副作用并允许接管，重复投递不会重复产生业务效果。

## 5. 实现契约

- **入口或接缝：** transaction event writer、outbox claimer/dispatcher、consumer idempotency、executor lease。
- **输入与输出：** 稳定事件 envelope 和业务键；delivered/retry 状态及脱敏指标。
- **公共接口变化：** 新增模块 Integration Event 与 consumer port 基座。
- **不变量：** 状态/事件原子；delivered 不重复生效；失锁立即停副作用；缓存不是事实源。
- **状态或数据流：** transaction -> pending -> claimed -> delivered；失败 claim -> retry。
- **错误与失败行为：** DB/consumer/lock 失败保留可恢复状态，不吞错、不忙循环。
- **兼容要求：** 不连接 Redis，不兼容旧队列 payload。
- **安全与隐私要求：** 事件和指标不包含 secret、密码或原始 Session。

## 6. 执行路线

1. 建立事务回滚、重复投递、断连接管和 cache-disabled 测试。
2. 实现 Outbox schema、writer、claim 与 dispatcher。
3. 实现 PostgreSQL advisory lock 和 SQLite 单实例适配。
4. 实现有界本地缓存及清空/禁用模式。
5. 运行故障注入、race 和双方言恢复矩阵。

## 7. 路径访问契约

- **预计修改点：** 后端可靠运行时平台及自有 migration。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** kernel、Database 和 migration Provider。
- **共享路径：** outbox/coordination/cache 由 T-08 唯一拥有。
- **保留或不动：** Scheduler 业务归 T-12，Audit 消费者归 T-11。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | outbox recovery suite | `task test -- reliable-runtime` | 事件提交、分发和幂等消费成立 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-08.md</Path>` |
| 失败路径 | DB/lock fault injection | 中断连接、消费者失败、worker 失锁 | 保留重试状态、停止副作用并可接管 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-08.md</Path>` |
| 回归 | cache/Redis scan | 清空禁用缓存并监测连接 | 正确性等价且无 Redis 连接尝试 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-08.md</Path>` |

- **Workspace checks：** 按 Goal Plan 在 current-workspace 或 source-worktree 运行 Go/race/unit/integration。
- **E2E disposition：** required：必须以多 PostgreSQL 进程和 SQLite 实例验证唯一 executor 与接管。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不运行 required E2E。
- **Integration evidence：** 记录 implementation/source commit、parent before、candidate/result SHA、故障 E2E 和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** Outbox/coordination schema 先于 Audit/Scheduler 消费者，T-17 后启用产品 worker。
- **兼容窗口：** 无 Redis 双写或旧 payload 兼容。
- **监控信号：** pending age、retry count、claim duration、active executor、lost lock。
- **回滚或前向恢复：** 禁用 dispatcher 保留 pending；修复后前向恢复并重放幂等消费者。
- **不可逆操作与批准点：** 清理 delivered 数据前必须有保留策略和独立批准。
- **收缩条件：** T-21 扫描 Redis 客户端、配置和连接尝试零命中。

## 10. 验收标准

- [ ] `AC-018`：PostgreSQL 唯一 executor、接管、重试和幂等成立，SQLite 无第二 executor。
- [ ] `AC-034`：缓存空/禁用时认证与业务正确性不变且无 Redis 尝试。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-08.md</Path>`。
- [ ] 修改未超出 `writable_paths`，共享路径仅由 T-08 修改。
- [ ] 形成非空 implementation/source commit，并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 状态一致且无未批准偏差。
