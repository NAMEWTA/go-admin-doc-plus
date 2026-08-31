---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-07
title: 建立结构化日志与 Doctor 诊断能力
status: done
planning_depth: deep
planning_depth_reason: 日志贯穿安全边界和所有运行角色，Doctor 需要稳定诊断合同、secret 脱敏与 profile 差异
ready: true
risk: high
blocked_by: []
contract_ids: [AC-029, AC-032, AC-033]
owner: codex-root
expected_changes: ["<Path>go-admin-plus/internal/platform/logging/**</Path>", "<Path>go-admin-plus/internal/application/operations/doctor/**</Path>", "<Path>go-admin-plus/internal/platform/config/logging*</Path>", "<Path>go-admin-plus/internal/platform/config/*_test.go</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/platform/logging/**</Path>", "<Path>go-admin-plus/internal/application/operations/doctor/**</Path>", "<Path>go-admin-plus/internal/platform/config/logging*</Path>", "<Path>go-admin-plus/internal/platform/config/*_test.go</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/app/product/**</Path>", "<Path>go-admin-plus/internal/host/**</Path>", "<Path>go-admin-plus/internal/modules/audit/**</Path>", "<Path>go-admin-plus/internal/platform/config/config.go</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-07: 建立结构化日志与 Doctor 诊断能力

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/07-observability-and-doctor.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-07.md</Path>`

## 1. 战略与来源

- **目标：** 提供真实生效、可关联、可脱敏的 logger 和三 profile 可复用 Doctor 检查模型。
- **可观察产出：** log.level 过滤生效；Server JSON、Desktop 轮转、dev console 格式可验证；Doctor 项目和退出分类可机器判定。
- **来源：** `US-014`、`AC-029`、`AC-032~033`、`PLAN/P6`。
- **当前事实：** config 解析 log.level，但产品 Host 未注入 logger；健康检查存在，尚无统一 Doctor。
- **Planning Depth 原因：** 错误日志可能泄露最高敏感 secret，且诊断合同跨数据库、文件、worker 和版本。

## 2. 决策状态

### 已锁定决策

- Server 生产 JSON stdout，Desktop 轮转文件，开发 console；Audit 与运行日志分责。
- 字段至少包含 service/version/profile/trace/request/route/module/status/latency/database/error class。
- Doctor 只读检查配置、secret reference、数据库、schema、setup、文件根、磁盘、worker 与版本。

### 已采用的低影响假设

- 使用 Go 标准结构化日志接口及最小轮转 adapter；具体编码器为内部实现细节。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| logger API/sinks/redaction、Doctor checks/result model、配置类型 | health checkers、typed profile、Audit 事实 | composition/CLI wiring（T-09）、请求正文捕获、远程 telemetry |

## 4. 要构建什么

运行角色可通过类型化配置建立目标 sink 和 level，业务/host 用稳定字段记录事件而不记录正文/secret。Doctor 在不改变系统状态的前提下执行有界检查，输出每项状态、稳定错误类和总退出分类；敏感连接信息只显示类型和可用性。

## 5. 实现契约

- **入口或接缝：** Logger factory/sink、redactor、Doctor Runner/Check。
- **输入与输出：** profile/version/level/sink 和依赖检查 ports；输出结构化事件与诊断结果。
- **公共接口变化：** 内部稳定 API；CLI 暴露由 T-09。
- **不变量：** log.level 真过滤；secret/body/session/CSRF/DSN 不输出；Doctor 只读且有界。
- **状态或数据流：** event fields -> redaction -> profile sink；checks -> aggregate -> exit class。
- **错误与失败行为：** sink/轮转失败和 checker timeout 分类处理，不递归泄漏原错误数据。
- **兼容要求：** 不保留未接入的旧 log.level 语义。
- **安全与隐私要求：** secret corpus 测试必须覆盖密码、Cookie、CSRF、DSN、request body。

## 6. 执行路线

1. 建立 level、字段 schema、sink 和 secret corpus 红灯测试。
2. 实现 logger factory、profile sinks、轮转和 redaction。
3. 实现 Doctor check/result/timeout 模型及依赖 ports。
4. 增加配置边界与失败分类测试。
5. 运行日志 capture、race、secret scan 和现有 health 回归。

## 7. 路径访问契约

- **预计修改点/可写范围：** logging、doctor 与 logging config 专属文件。
- **只读上下文：** product/host/Audit/现有 config。
- **共享路径：** 无；统一接线由 T-09 独占。
- **保留或不动：** 不把 Audit 替换为日志，不记录请求/响应正文。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | logger/Doctor | sink capture 与 fake checker | 字段、level、退出分类正确 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-07.md</Path>` |
| 失败路径 | redaction/timeout | secret corpus、sink failure、checker timeout | 无泄漏，稳定失败分类 | 同上 |
| 回归 | config/health | 定向 Go tests + vet/race | 原 health/config 行为保持 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行 logging/doctor/config 单元、race、vet、secret scan。
- **E2E disposition：** not-required：本 Ticket 交付可注入内部能力；真实进程/CLI 输出由 T-09，Desktop 文件由 T-20。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source 不声明 E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** logger/Doctor API -> T-09 composition -> T-18/T-21 门禁与文档。
- **兼容窗口：** 无旧 logger 双写；接入点一次替换。
- **监控信号：** sink write failure、dropped event、Doctor 项目状态和 timeout。
- **回滚或前向恢复：** sink 故障回退最小 stderr 安全事件；配置/adapter 前向修复。
- **不可逆操作与批准点：** 无数据不可逆操作；日志路径/权限由 T-09/T-10 宿主批准。
- **收缩条件：** 未接 logger 的启动/请求/worker 关键路径扫描为零。

## 10. 验收标准

- [ ] `AC-029、AC-032~033` 的日志、脱敏和 Doctor 模型成立。
- [ ] log.level、三 sink 和 secret corpus 测试通过。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-07.md</Path>`。
- [ ] 路径、commit、集成、result 与 E2E disposition 完整。
