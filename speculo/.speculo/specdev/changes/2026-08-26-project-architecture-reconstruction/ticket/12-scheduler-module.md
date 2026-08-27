---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-12
title: Scheduler 受控任务垂直切片
status: in_progress
planning_depth: deep
planning_depth_reason: 调度定义、后台副作用、唯一 executor、停止语义和执行记录具有高事故半径
ready: true
risk: critical
blocked_by: [T-07, T-08]
contract_ids: [AC-017, AC-018, AC-035]
owner: codex-t12-scheduler
expected_changes: ["<Path>contracts/openapi/modules/scheduler.yaml</Path>", "<Path>go-admin-plus/internal/modules/scheduler/**</Path>", "<Path>go-admin-plus-ui/packages/domains/scheduler/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/scheduler/**</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/scheduler.yaml</Path>", "<Path>go-admin-plus/internal/modules/scheduler/**</Path>", "<Path>go-admin-plus-ui/packages/domains/scheduler/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/scheduler/**</Path>", "<Path>go-admin-plus/test/scheduler/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/scheduler/**</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/authorization/**</Path>", "<Path>go-admin-plus/internal/platform/outbox/**</Path>", "<Path>go-admin-plus/internal/platform/coordination/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/packages/domains/scheduler/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/scheduler/package.json</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/packages/domains/scheduler/package.json</Path> => T-12 under T12-D01; package-local exports/dependencies/checks only", "<Path>go-admin-plus-ui/packages/web-domains/scheduler/package.json</Path> => T-12 under T12-D01; package-local exports/dependencies/checks only", "<Path>go-admin-plus-ui/package.json</Path> => T-12 under T12-D02; Scheduler typecheck only", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path> => T-12 under T12-D02; Scheduler specs only", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-12 under T12-D02; existing domain-scheduler and web-domain-scheduler importers only"]
---

# Ticket T-12: Scheduler 受控任务垂直切片

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/12-scheduler-module.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-12.md</Path>`

## 1. 战略与来源

- **目标：** 交付注册任务类型、任务定义、启停、删除和执行记录闭环。
- **可观察产出：** 仅有效、启用且由 active executor 持权的任务执行；停止后不再产生新执行。
- **来源：** `US-010`、`AC-017`、`AC-018`、`AC-035`、`ADR-019`。
- **当前事实：** 旧 jobs 可能依赖 Redis/动态调用，缺少注册类型和唯一所有权合同。
- **Planning Depth 原因：** 任意代码执行、重复调度或失锁后继续副作用均为关键风险。

## 2. 决策状态

### 已锁定决策

- 任务只引用编译期注册类型和校验后的参数，不执行任意代码字符串。
- executor 权限来自 T-08；失权立即停止新触发。
- Scheduler 与 Outbox 消费 T-08 注入的同一个全局 coordination lease；T-17 只获取一次并以同一 DB、owner 和取消边界启动整个 worker group，禁止 Scheduler 再申请 namespace lease。
- task handler 的数据库副作用只能在 `Lease.WithinTx` 内发生；外部效果必须在同一事务写入 Outbox。lease loss 回滚整个事务，只允许发送脱敏 observer，不得另写“执行失败”记录。
- task registry 为编译期 typed registry；schedule 使用结构化 UTC 模型和注入时钟，不接受 cron 字符串、任意代码、动态调用目标或 API 注册 task type。

### 已采用的低影响假设

- 时钟由可注入 Clock 控制，执行记录使用稳定状态枚举。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 定义、registry、executor、记录、API/UI | IAM、Outbox、coordination | 任意脚本执行、分布式外部队列 |

## 4. 要构建什么

管理员选择已注册任务类型并配置计划；Scheduler 验证定义后持久化，active executor 按可控时钟触发并记录结果，禁用、删除或失权后不再产生新执行。

## 5. 实现契约

- **入口或接缝：** task registry、Scheduler use cases/repository/executor、API/UI。
- **输入与输出：** 类型键、计划和参数；返回定义/执行记录或稳定错误。
- **公共接口变化：** 新 Scheduler fragment、task registration 和 Permission Code。
- **不变量：** 未注册类型不执行；disabled/失权不触发；执行记录与定义关联；无 tenant。
- **状态或数据流：** admin command -> definition -> clock due -> active lease -> execution -> result event/record。
- **错误与失败行为：** 参数非法、未注册类型、失锁或执行失败被记录/拒绝且可恢复。
- **兼容要求：** 不兼容旧 job invoke target、Redis queue 或旧表达式。
- **安全与隐私要求：** 参数 schema allowlist；日志/记录不含 secret。

## 6. 执行路线

1. 建立 clock-controlled、registry、stop 和失锁测试。
2. 实现 migration、定义/记录 repository 和 task registry。
3. 实现 executor、Outbox 接缝、API mapping 和 Web 页面。
4. 覆盖重复触发、接管、未注册类型和敏感参数。
5. 将双方言、多实例与浏览器场景登记到全部 Ticket 实现集成后的统一系统 E2E；本 Ticket candidate 只运行完整非 E2E Gate。

## 7. 路径访问契约

- **预计修改点：** Scheduler 独占路径。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** IAM、Outbox 和 coordination。
- **共享路径：** 两个预建 Scheduler package manifest 由 `T12-D01` 拥有；根 workspace script、shell Vitest include 和 lockfile 仅按 `T12-D02` 串行接入。
- **批准偏差：** `T12-D01` 只开放两个预建 Scheduler package manifest，用于 package-local exports、直接依赖和检查；根 package、Vitest、lockfile、composition、Outbox 与 coordination 实现仍只读，等待 Lead 串行 amendment。
- **批准偏差：** `T12-D02` 在 T-10 implementation result 后只开放根 Scheduler 聚合 typecheck、shell Vitest 的两个 Scheduler include，以及 lockfile 现有 `packages/domains/scheduler` / `packages/web-domains/scheduler` importer。只允许反映两个已批准 package manifest 的既有 workspace/catalog 依赖；其他 script、include、importer、catalog、外部版本、composition、Outbox 和 coordination 实现零漂移。
- **保留或不动：** T-08 executor lease 实现。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | clock/API/UI suite | `task test -- scheduler` | 定义、启停、执行和记录一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-12.md</Path>` |
| 失败路径 | registry/lease suite | 未注册类型、非法参数、失锁和执行失败 | 不产生未授权副作用且状态可解释 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-12.md</Path>` |
| 回归 | multi-instance E2E | 多 worker 触发、断连接管、停止任务 | 同时唯一执行，停止后无新记录 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-12.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 非 E2E 检查。
- **E2E disposition：** deferred：多实例执行权、停止和 UI 记录保留到全部 Ticket 实现集成后的统一系统 E2E。
- **E2E owner/environment：** Lead / 最终系统候选；逐 Ticket source-worktree 与 parent-candidate 均不运行或声明 E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、完整非 E2E Gate、统一系统 E2E 引用与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-08 后落地定义/执行 schema，T-17 再启用产品 worker。
- **兼容窗口：** 无旧 job 数据或 invoke target 导入。
- **监控信号：** active executor、due lag、execution failure、lost lock 和 disabled trigger。
- **回滚或前向恢复：** 先禁用 executor 保留定义/记录；修复后前向恢复。
- **不可逆操作与批准点：** 执行类型注册必须代码审查；记录清理需保留策略。
- **收缩条件：** T-21 证明旧 jobs、Redis queue 和动态调用零引用。

## 10. 验收标准

- [ ] `AC-017`：任务管理、注册限制、停止语义和执行记录成立。
- [ ] `AC-018`：唯一 executor 与失联接管成立。
- [ ] `AC-035`：调度管理交互符合共享合同。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-12.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
