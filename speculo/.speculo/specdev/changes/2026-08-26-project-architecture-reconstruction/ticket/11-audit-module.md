---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-11
title: Audit 登录与操作审计垂直切片
status: in_progress
planning_depth: deep
planning_depth_reason: 审计跨认证和业务事件且包含隐私、保留与清理安全要求
ready: true
risk: high
blocked_by: [T-06, T-08]
contract_ids: [AC-016, AC-035]
owner: codex-t11-audit
expected_changes: ["<Path>contracts/openapi/modules/audit.yaml</Path>", "<Path>go-admin-plus/internal/modules/audit/**</Path>", "<Path>go-admin-plus-ui/packages/domains/audit/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/audit/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/audit/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/audit/package.json</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/audit.yaml</Path>", "<Path>go-admin-plus/internal/modules/audit/**</Path>", "<Path>go-admin-plus-ui/packages/domains/audit/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/audit/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/audit/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/audit/package.json</Path>", "<Path>go-admin-plus/test/audit/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/audit/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/session/**</Path>", "<Path>go-admin-plus/internal/platform/outbox/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-11: Audit 登录与操作审计垂直切片

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/11-audit-module.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-11.md</Path>`

## 1. 战略与来源

- **目标：** 交付登录与操作审计的可靠记录、筛选、详情和授权清理。
- **可观察产出：** 审计员能查询真实活动，记录不含密码、原始 Session、secret 或完整敏感请求体。
- **来源：** `US-009`、`AC-016`、`AC-035`、`ADR-009`、`ADR-012`、`ADR-019`。
- **当前事实：** 旧日志模型与中间件耦合，敏感请求体和 tenant 字段可能进入持久化。
- **Planning Depth 原因：** 跨模块可靠性与隐私泄露风险并存。

## 2. 决策状态

### 已锁定决策

- Audit 消费稳定 Integration Event，不读取其他模块私有表；登录关键事实使用受控同步接缝。
- 清理是受权业务命令，不提供任意表清空。

### 已采用的低影响假设

- 审计 payload 使用字段 allowlist 和长度上限。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 登录/操作记录、消费、查询、清理和页面 | IAM 事件、Outbox、共享 UI | 全量请求体、外部 SIEM |

## 4. 要构建什么

登录与受审计业务操作产生稳定事件，Audit 以幂等键持久化脱敏事实；授权审计员可筛选/查看并按产品策略清理，失败事件可重试而不重复记录。

## 5. 实现契约

- **入口或接缝：** Audit event consumer、repository、query/cleanup API、Web pages。
- **输入与输出：** allowlisted audit envelope；返回分页事实或稳定错误。
- **公共接口变化：** 新 Audit fragment、event schema 和 Permission Code。
- **不变量：** 幂等记录；不可包含敏感字段；清理受权限和策略限制；无 tenant。
- **状态或数据流：** IAM/business event -> outbox -> Audit consumer -> repository -> query/UI。
- **错误与失败行为：** 消费失败可重试，脱敏失败拒绝持久化，越权清理无状态变化。
- **兼容要求：** 不迁移旧日志表或 payload。
- **安全与隐私要求：** 明确 allowlist、截断、哈希/脱敏和保留策略测试。

## 6. 执行路线

1. 建立敏感值、幂等、重试和清理权限测试。
2. 实现 migration、event consumer、repository 和 redactor。
3. 实现查询/清理合同及 Web 页面。
4. 接入登录探针和标准业务事件 fixture。
5. 运行双方言、故障恢复和敏感扫描。

## 7. 路径访问契约

- **预计修改点：** 当前阶段为 Audit 独占路径和经 `T11-D01` 批准的两个 Audit package manifest；T-07 result 后由 Lead 激活 `T11-D02`，再加入 Session 登录审计接缝与既有集成测试。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** 当前阶段全部 IAM Session、Outbox、共享 UI。
- **共享路径：** 当前 frontmatter 无共享可写路径；workspace lock 和两个待开放 Session 路径都先归 T-07。T-11 可并行编写独占源码和 manifest，但必须等待 T-07 result、Lead 激活第二阶段 frontmatter 并 rebase 后，才可写 Audit importer 与 Session 登录审计接缝。
- **批准偏差：** `T11-D01` 当前阶段仅允许两个 Audit package manifest 补齐 public export、canonical API client、`@go-admin/ui`、Vue 直接依赖与标准 test/typecheck 入口。T-07 result 进入父分支后，Lead 必须先更新本 Ticket 并 rebase，才能开放只更新两个 Audit importer 的第二阶段；其他漂移必须再次停止。
- **批准偏差：** `T11-D02` 已批准但当前尚未激活；T-07 result 后由 Lead 把精确 Session service/test 路径加入 frontmatter 并 rebase，才允许增加模块无关的登录事实 Port。既有 Session 测试必须证明成功登录与 Session 创建同事务、失败登录同步记录、审计失败不产生已签发 Session、attempt ID 不含用户名且密码/token/CSRF 零泄露。Audit 通过结构化适配器消费该 Port；禁止 Session 导入 Audit、读取 Audit 私表、修改 Session schema/HTTP/Cookie/policy，或把未来业务模块事件硬编码为 Demo 私有 topic。
- **保留或不动：** 事件生产者实现和全局产品注册。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | audit integration/UI | `task test -- audit` | 登录/操作事实可筛选查看清理 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-11.md</Path>` |
| 失败路径 | redaction/retry suite | 注入敏感字段、重复事件、消费者失败和越权清理 | 无泄露、幂等且可恢复 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-11.md</Path>` |
| 回归 | secret scan/dialects | 双方言执行并扫描记录/响应 | 行为等价且敏感值零命中 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-11.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 非 E2E 检查。
- **E2E disposition：** required：登录/业务事件到审计 UI 的跨进程链路必须验证。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不声明通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、E2E 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** Audit consumer 在 Outbox 后落地，T-17 启用完整生产者集合。
- **兼容窗口：** 无旧日志导入。
- **监控信号：** consumer lag/retry、redaction reject、cleanup 和查询错误。
- **回滚或前向恢复：** 可停 consumer 保留 Outbox；修复后幂等重放。
- **不可逆操作与批准点：** 审计清理必须二次确认、权限和保留策略同时满足。
- **收缩条件：** T-21 证明旧日志模型、tenant 与敏感 payload 零引用。

## 10. 验收标准

- [ ] `AC-016`：审计事实完整、幂等可恢复，查询/清理受权且敏感值零泄露。
- [ ] `AC-035`：审计列表/筛选/清理交互符合共享合同。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-11.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] E2E disposition 已执行且 shared path 无越权写入。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
