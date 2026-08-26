---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-04
title: 迁移 ServerHost 并建立健康与能力接口
status: done
planning_depth: deep
planning_depth_reason: 替换生产进程入口并新增运维公共接口，影响启动、信号、TLS 和 readiness
ready: true
risk: high
blocked_by: [T-02, T-03]
contract_ids: [AC-001, AC-002, AC-013]
owner: root
expected_changes: ["<Path>go-admin-plus/cmd/go-admin/**</Path>", "<Path>go-admin-plus/internal/host/server/**</Path>", "<Path>go-admin-plus/internal/application/health/**</Path>", "<Path>go-admin-plus/Makefile</Path>"]
writable_paths: ["<Path>go-admin-plus/cmd/go-admin/**</Path>", "<Path>go-admin-plus/internal/host/server/**</Path>", "<Path>go-admin-plus/internal/application/health/**</Path>", "<Path>go-admin-plus/Makefile</Path>"]
read_only_paths: ["<Path>go-admin-plus/cmd/api/**</Path>", "<Path>go-admin-plus/internal/platform/**</Path>", "<Path>go-admin-plus/test/characterization/**</Path>"]
shared_paths: ["<Path>go-admin-plus/Makefile</Path>"]
shared_path_owners: ["<Path>go-admin-plus/Makefile</Path> => T-04"]
---

# Ticket T-04: 迁移 ServerHost 并建立健康与能力接口

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/04-server-host-and-health.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-04.md</Path>`

## 1. 战略与来源

- **目标：** 让生产 CLI 成为 Application 的 ServerHost，拥有网络 listener、TLS、signal 和超时关闭，并提供明确 live/ready/capabilities。
- **可观察产出：** `go-admin server` 保持旧 API 行为，同时依赖异常时 ready 失败、进程信号触发有序停止。
- **来源：** `AC-001`、`AC-002`、`AC-013`。
- **当前事实：** 旧 run 在 goroutine 中 ListenAndServe，并在错误时 `log.Fatal`；health 路由语义未区分 live/ready。
- **Planning Depth 原因：** 生产入口和新增公共运维接口，影响全部部署。

## 2. 决策状态

### 已锁定决策

- ServerHost 绑定配置地址，拥有 signal.NotifyContext、HTTP Server 和 shutdown timeout；Application 不感知信号。
- `/health/live` 只反映进程事件循环；`/health/ready` 检查 Application ready 与必需依赖；capabilities 只返回非敏感功能标记。
- 现有 server 命令参数继续兼容，新增 profile/health 不移除旧 flag。

### 已采用的低影响假设

- TLS 可继续由现有配置启用；Compose 默认由 Nginx 终止 TLS，不删除直连 TLS 能力。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| 新 CLI 入口、ServerHost、live/ready/capabilities、构建目标 | Application、Server profile、现有 Cobra 配置 | Docker Compose、桌面、业务路由改写 |

## 4. 要构建什么

运维人员可用原命令启动服务；Host 构建 server profile 和 Application，监听后报告版本与地址，接收终止信号时停止接收请求并关闭 Application。探针能稳定区分进程活着、依赖未就绪和完整 ready；能力接口不暴露 DSN、路径或 token。

## 5. 实现契约

- **入口或接缝：** `cmd/go-admin` server command、ServerHost Run、三个 HTTP 接口。
- **输入与输出：** 配置/flags/context；进程退出码、HTTP 状态和非敏感 JSON。
- **公共接口变化：** 新增 health/capabilities；旧 CLI/API 兼容。
- **不变量：** listener 错误返回 Host；shutdown 最多执行一次；ready 不早于迁移和 Module 启动完成。
- **状态或数据流：** CLI -> config -> profile -> Application -> listener -> signal -> shutdown。
- **错误与失败行为：** bind/依赖/启动失败退出非零；shutdown 超时返回错误但继续资源清理。
- **兼容要求：** characterization 和旧 flags 通过。
- **安全与隐私要求：** capabilities/health 不含 secret、绝对数据路径或用户数据。

## 6. 执行路线

1. 为 ServerHost signal、bind failure 和探针语义增加集成测试。
2. 建立新 cmd 与 Host，接入 T-02/T-03。
3. 实现 health registry/capabilities 并注册新增路由。
4. 将 Makefile 默认 server build 指向新入口，保留旧 CLI 兼容 shim。
5. 运行 CLI smoke、信号关闭和 API characterization。

## 7. 路径访问契约

- **预计修改点/可写范围：** 新 server cmd/host/health 与 Makefile。
- **只读上下文：** 旧 API 入口、platform 和基线测试。
- **共享路径：** Makefile 唯一 owner T-04；其他 Ticket 需要新目标时经 T-04/Goal Plan 协调。
- **保留或不动：** Dockerfile、前端、桌面和业务模型。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | CLI/HTTP | 启动、请求探针/API、发信号 | ready 正确且有序退出 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-04.md</Path>` |
| 失败路径 | bind/DB/迁移注入 | 占用端口、断开依赖 | 非零退出或 ready 失败 | 同上 |
| 回归 | characterization/build | `go test ./...`、`make build` | 旧 API/CLI 绿色 | 同上 |

- **Workspace checks：** Go test/build、静态检查和 CLI smoke。
- **E2E disposition：** required；真实进程、网络和信号边界。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；运行真实 ServerHost。
- **Integration evidence：** 记录提交、parent/candidate/result 与包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 新入口与旧 shim 并存；容器尚未切换。
- **兼容窗口：** 旧 cmd/api 到 T-13 零调用扫描后删除。
- **监控信号：** live、ready、启动阶段和退出原因。
- **回滚或前向恢复：** 可将构建目标回指兼容 shim；新增接口保留不破坏旧调用方。
- **不可逆操作与批准点：** 删除旧 CLI 入口留给 T-13。
- **收缩条件：** Makefile、Docker 和桌面均使用新入口/Application。

## 10. 验收标准

- [x] `AC-001`、`AC-002`、`AC-013` 可判定通过。
- [x] signal/bind/dependency 失败与回归证据完整。
- [x] 路径、提交、集成和 required E2E 合同满足。
- [x] health/capabilities 不泄密，Map/Evidence 同步。
