---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-02
title: 建立 Host 无关 Application 生命周期内核
status: done
planning_depth: deep
planning_depth_reason: 重构共享启动核心、全局模块注册和生命周期，影响全部 HTTP 行为与后台工作
ready: true
risk: high
blocked_by: [T-01]
contract_ids: [AC-001, AC-002, AC-011]
owner: root
expected_changes: ["<Path>go-admin-plus/internal/application/**</Path>", "<Path>go-admin-plus/internal/modules/**</Path>", "<Path>go-admin-plus/app/*/module.go</Path>", "<Path>go-admin-plus/cmd/api/**</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/application/**</Path>", "<Path>go-admin-plus/internal/modules/**</Path>", "<Path>go-admin-plus/app/*/module.go</Path>", "<Path>go-admin-plus/cmd/api/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/common/**</Path>", "<Path>go-admin-plus/test/characterization/**</Path>"]
shared_paths: ["<Path>go-admin-plus/cmd/api/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus/cmd/api/**</Path> => T-02"]
---

# Ticket T-02: 建立 Host 无关 Application 生命周期内核

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/02-application-lifecycle-kernel.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-02.md</Path>`

## 1. 战略与来源

- **目标：** 将现有启动流程深化为小型 Application 接口，以显式 ModuleSet 取代 `init()` 路由列表，并让 Host 之外不拥有信号或监听器。
- **可观察产出：** 测试可构建 Application、通过 Handler 调用既有 API、启动后台 Module 并用 context 有序停止。
- **来源：** `AC-001`、`AC-002`、`AC-011`、`CODE:<Path>go-admin-plus/cmd/api/server.go</Path>`。
- **当前事实：** `setup/run` 同时创建全局 Runtime、队列、任务、HTTP Server 与信号等待；AppRouters 由 package init 修改。
- **Planning Depth 原因：** 共享核心、全局状态和所有业务 Module 的事故半径高，且需要 expand-contract 兼容旧 CLI。

## 2. 决策状态

### 已锁定决策

- Application 对外只暴露 Handler、Start、Stop 和只读运行状态；依赖在 Build 时注入。
- Module 提供 ID、路由注册、迁移声明和可取消生命周期；装配顺序显式且重复 ID 启动失败。
- 旧 `go-admin server` CLI 在兼容窗口内作为 ServerHost 适配入口继续工作。

### 已采用的低影响假设

- go-admin-core 全局 Runtime 先封装在 Application 内部兼容适配器中，再由后续切片缩小；本 Ticket 不要求一次清除第三方全局状态。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| Application/ModuleSet/Lifecycle、显式模块装配、旧 CLI 兼容 | 现有 Gin、路由函数、配置、JWT/Casbin | 数据库 profile、DesktopHost、Docker、业务逻辑改写 |

## 4. 要构建什么

调用者用 Config、Dependencies 和 ModuleSet 构建 Application；构建失败返回错误而不退出进程。启动依序启动 Module，Handler 即刻可供 Host 接管；停止反序取消后台工作并在超时内返回。现有 CLI 通过临时兼容入口调用这一内核，characterization 保持绿色。

## 5. 实现契约

- **入口或接缝：** `application.Build(...)`、`Application.Handler/Start/Stop`、Module interface。
- **输入与输出：** 已解析配置、依赖和模块；返回 Application 或带上下文错误。
- **公共接口变化：** 新增进程内 Go 接口；外部 HTTP/CLI 保持兼容。
- **不变量：** 不重复启动/停止；部分启动失败时反序清理已启动 Module；业务层不处理 OS 信号。
- **状态或数据流：** Build -> constructed -> starting -> ready -> stopping -> stopped；错误不会产生部分 ready。
- **错误与失败行为：** 重复模块、依赖缺失、路由冲突或 Module 启动失败均返回错误，不 `log.Fatal`。
- **兼容要求：** `/api/v1` 与 `go-admin server -c` 基线通过。
- **安全与隐私要求：** 构建错误不打印 secret 值。

## 6. 执行路线

1. 用 T-01 接缝锁定旧入口，增加 Application 生命周期失败测试。
2. 引入 Application 和显式 ModuleSet，先用兼容适配器包住现有 Runtime。
3. 为 admin/jobs/demo/other 增加 `module.go`，保持现有路由函数实现。
4. 将旧 CLI 改为调用新内核，移除由 `init()` 修改 AppRouters 的新依赖路径。
5. 验证部分启动清理、重复启停和全部 HTTP characterization。

## 7. 路径访问契约

- **预计修改点/可写范围：** 仅 frontmatter 所列 Application、Module 装配和旧 API CLI。
- **只读上下文：** 业务实现、common 和 characterization。
- **共享路径：** `<Path>go-admin-plus/cmd/api/**</Path>` 唯一 owner 为 T-02。
- **保留或不动：** 数据模型、迁移版本、前端、Docker 和用户未提交文件。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | Application handler/lifecycle | 定向 Go tests | 可构建、请求、停止 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-02.md</Path>` |
| 失败路径 | 启动失败注入 | 重复 Module、依赖缺失、部分启动测试 | 返回错误并清理 | 同上 |
| 回归 | CLI + characterization | `go test ./...` 与 T-01 API 套件 | 既有行为绿色 | 同上 |

- **Workspace checks：** Goal Plan workspace 中运行 Go 定向/全量测试与 build。
- **E2E disposition：** required；启动核心跨配置、路由和后台生命周期。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；运行真实 CLI API smoke。
- **Integration evidence：** implementation/source commit、parent before、适用 candidate/result SHA 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 新 Application 与旧 CLI 兼容并存，再切换 CLI；不触碰数据迁移。
- **兼容窗口：** 旧 `cmd/api` 仅作为适配入口保留到 T-13 扫描确认。
- **监控信号：** 启动阶段日志、ready 状态与 shutdown 错误。
- **回滚或前向恢复：** CLI 可在兼容窗口回指旧 run；不得回滚已发布数据。
- **不可逆操作与批准点：** 删除 init/AppRouters 旧入口需留到 T-13 人工批准。
- **收缩条件：** 所有 Host 使用 Application、旧入口调用点为零且 characterization 通过。

## 10. 验收标准

- [x] `AC-001`、`AC-002`、`AC-011` 通过。
- [x] Application 无 OS signal/listen/Wails 依赖，Module 可取消。
- [x] 正常、失败、回归证据写入 `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-02.md</Path>`。
- [x] 路径、提交、direct-parent/candidate 与 required E2E 合同满足。
- [x] 无未批准契约、数据或发布偏差，Map/Evidence 状态一致。
