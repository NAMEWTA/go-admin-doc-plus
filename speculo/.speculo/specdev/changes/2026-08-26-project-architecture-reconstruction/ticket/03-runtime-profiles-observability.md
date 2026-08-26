---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-03
title: 应用内核、类型化 Profile 与可观测启动链
status: ready
planning_depth: deep
planning_depth_reason: 启动生命周期、配置 secret 边界和运行状态端点影响全部宿主
ready: true
risk: high
blocked_by: [T-02]
contract_ids: [AC-005, AC-026, AC-027]
owner: unassigned
expected_changes: ["<Path>go-admin-plus/internal/app/kernel/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>", "<Path>go-admin-plus/internal/platform/observability/**</Path>", "<Path>go-admin-plus/config/schema/**</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/app/kernel/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>", "<Path>go-admin-plus/internal/platform/observability/**</Path>", "<Path>go-admin-plus/config/schema/**</Path>", "<Path>go-admin-plus/cmd/config-check/**</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>"]
shared_paths: ["<Path>go-admin-plus/internal/app/kernel/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>", "<Path>go-admin-plus/internal/platform/observability/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus/internal/app/kernel/**</Path> => T-03", "<Path>go-admin-plus/internal/platform/config/**</Path> => T-03", "<Path>go-admin-plus/internal/platform/observability/**</Path> => T-03"]
---

# Ticket T-03: 应用内核、类型化 Profile 与可观测启动链

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/03-runtime-profiles-observability.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-03.md</Path>`

## 1. 战略与来源

- **目标：** 建立 Server PostgreSQL、Server SQLite、Desktop SQLite 的不可变配置和一致应用生命周期。
- **可观察产出：** 无效配置在监听/开窗前失败；live、ready、metrics、capabilities 和 status 语义可区分。
- **来源：** `US-002`、`US-003`、`US-014`、`AC-005`、`AC-026`、`AC-027`、`ADR-020`。
- **当前事实：** 旧全局配置和宿主状态可被运行时读取/修改，健康语义未形成统一合同。
- **Planning Depth 原因：** 配置错误或状态误报会影响数据安全、部署编排与桌面启动。

## 2. 决策状态

### 已锁定决策

- 三个 profile 使用独立最小 schema；优先级为 defaults < file < env < 非敏感 CLI。
- secret 仅来自 env 或 `_FILE`；Desktop 路径和启动材料来自 Tauri。
- 应用状态固定为 `starting -> ready -> draining -> stopped`，失败状态不得 ready。

### 已采用的低影响假设

- 配置解析使用显式 decode/validate，不支持运行时 reload。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| kernel、配置 schema、生命周期和观测端点 | Go 标准 context/http 生命周期能力 | 数据库连接、业务模块、桌面窗口 |

## 4. 要构建什么

宿主先解析并校验指定 profile，再构造不可变配置快照并推进生命周期；任何未知字段、冲突或 secret 失败都在外部入口开放前中止，运维端点只暴露脱敏状态。

## 5. 实现契约

- **入口或接缝：** config loader、application kernel、observability handler。
- **输入与输出：** profile/file/env/CLI 输入；类型化快照、生命周期事件和脱敏 HTTP 响应。
- **公共接口变化：** 新增三 profile 配置合同及职责分离的运维端点。
- **不变量：** 单进程单 profile；配置不可变；unknown field 失败；ready 仅代表可接收业务。
- **状态或数据流：** sources -> precedence merge -> decode/validate -> kernel -> lifecycle/handlers。
- **错误与失败行为：** 错误指出字段和规则，不包含值、DSN、secret 内容或完整路径。
- **兼容要求：** 不读取旧全局配置、多数据源或 tenant 字段。
- **安全与隐私要求：** dump、日志、metrics 和 status 均不得暴露 secret。

## 6. 执行路线

1. 以配置失败和生命周期状态测试建立红线。
2. 实现三 schema、precedence、secret resolver 与不可变快照。
3. 实现 kernel 状态机、drain 顺序和运维 handler。
4. 加入未知字段、冲突、脱敏和 dependency-failed 探针。
5. 运行 profile 与 host 回归。

## 7. 路径访问契约

- **预计修改点：** kernel、配置平台、观测平台和 schema。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** 根任务和公共 HTTP 合同。
- **共享路径：** kernel/config/observability 由 T-03 唯一拥有。
- **保留或不动：** 数据库实现归 T-04，产品装配归 T-17。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | profile/host suite | `task test -- runtime-profiles` | 三 profile precedence 和状态转换成立 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-03.md</Path>` |
| 失败路径 | 启动负向 suite | 注入缺失、未知、冲突和不可读 secret | 监听/开窗前失败且输出脱敏 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-03.md</Path>` |
| 回归 | observability suite | 覆盖 starting/ready/failed/draining | live/ready/status 语义区分且 no-store | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-03.md</Path>` |

- **Workspace checks：** 按 Goal Plan 在 current-workspace 或 source-worktree 运行 Go 单元、race、静态和构建检查。
- **E2E disposition：** required：必须验证真实进程在配置失败前不监听以及状态端点转换。
- **E2E owner/environment：** Lead / current-workspace（current 策略）或 parent-candidate（required 策略），禁止在 source-worktree 声明通过。
- **Integration evidence：** 记录 implementation/source commit、parent before、适用 candidate/result SHA 和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 新 kernel/config 先独立落地，T-04/T-16/T-17 依次接入宿主。
- **兼容窗口：** 无运行时双配置；旧配置只读保留到 T-21。
- **监控信号：** startup failure、ready 状态、drain 时长和脱敏测试。
- **回滚或前向恢复：** 产品装配前可回滚；装配后修复 schema/adapter 并重新验证三个 profile。
- **不可逆操作与批准点：** 无。
- **收缩条件：** T-21 扫描证明旧全局配置和 runtime reload 零引用。

## 10. 验收标准

- [ ] `AC-005`：所有无效配置在监听或开窗前失败且诊断脱敏。
- [ ] `AC-026`：运维端点在各生命周期状态下语义正确。
- [ ] `AC-027`：三 profile schema、precedence、secret 和 Desktop 来源边界可验证。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-03.md</Path>`。
- [ ] 修改未超出 `writable_paths`，形成非空 commit 并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 状态一致且无未批准偏差。
