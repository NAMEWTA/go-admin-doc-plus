---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-06
title: 建立文件配额与磁盘容量治理
status: done
planning_depth: deep
planning_depth_reason: 涉及双方言配额并发、文件系统崩溃一致性、容量拒绝和持久数据迁移
ready: true
risk: high
blocked_by: []
contract_ids: [AC-027, AC-028, AC-029]
owner: codex-root
expected_changes: ["<Path>go-admin-plus/internal/modules/files/service.go</Path>", "<Path>go-admin-plus/internal/modules/files/service_test.go</Path>", "<Path>go-admin-plus/internal/modules/files/repository.go</Path>", "<Path>go-admin-plus/internal/modules/files/repository_test.go</Path>", "<Path>go-admin-plus/internal/modules/files/storage.go</Path>", "<Path>go-admin-plus/internal/modules/files/storage_test.go</Path>", "<Path>go-admin-plus/internal/modules/files/quota*</Path>", "<Path>go-admin-plus/internal/modules/files/reconcile*</Path>", "<Path>go-admin-plus/internal/modules/files/migrations/0020-capacity/**</Path>", "<Path>go-admin-plus/test/files/quota*</Path>", "<Path>go-admin-plus/test/files/capacity*</Path>", "<Path>go-admin-plus/test/files/reconcile*</Path>", "<Path>go-admin-plus/test/files/dialect_capacity*</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/modules/files/service.go</Path>", "<Path>go-admin-plus/internal/modules/files/service_test.go</Path>", "<Path>go-admin-plus/internal/modules/files/repository.go</Path>", "<Path>go-admin-plus/internal/modules/files/repository_test.go</Path>", "<Path>go-admin-plus/internal/modules/files/storage.go</Path>", "<Path>go-admin-plus/internal/modules/files/storage_test.go</Path>", "<Path>go-admin-plus/internal/modules/files/quota*</Path>", "<Path>go-admin-plus/internal/modules/files/reconcile*</Path>", "<Path>go-admin-plus/internal/modules/files/migrations/0020-capacity/**</Path>", "<Path>go-admin-plus/test/files/quota*</Path>", "<Path>go-admin-plus/test/files/capacity*</Path>", "<Path>go-admin-plus/test/files/reconcile*</Path>", "<Path>go-admin-plus/test/files/dialect_capacity*</Path>"]
read_only_paths: ["<Path>contracts/openapi/modules/files.yaml</Path>", "<Path>go-admin-plus/internal/modules/files/http.go</Path>", "<Path>go-admin-plus/internal/modules/files/transport/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-06: 建立文件配额与磁盘容量治理

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/06-file-capacity-governance.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-06.md</Path>`

## 1. 战略与来源

- **目标：** 防止账号、对象和磁盘无界增长，并让上传在数据库/文件系统故障间可恢复。
- **可观察产出：** 超额上传稳定拒绝且不超卖；崩溃预留可回收；低磁盘只阻止上传，下载/删除继续。
- **来源：** `US-012`、`AC-027~029`、`PLAN/P5`。
- **当前事实：** Files 仅有单文件 10 MiB 限制；上传已有 stage/publish/reconcile 基础但无账号/全局预留。
- **Planning Depth 原因：** 数据库并发与物理文件状态不一致可能造成容量泄漏或数据丢失。

## 2. 决策状态

### 已锁定决策

- 同时约束单文件、单账号字节/对象数、全局容量、磁盘最小剩余字节/比例。
- 上传先原子 reserve，再 stage/publish；reconciliation 有界回收。
- 低水位拒绝上传但允许下载和删除。

### 已采用的低影响假设

- 配额默认值和 reconciliation 批量大小进入受验证 profile 配置；具体数值可调但不能设为无界。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 0020 migration、quota reserve、磁盘 probe、reconciliation、容量错误 | 现有 local storage、stage/publish、授权 scope | 云对象存储、回收站、HTTP DTO/UI（T-08/T-16） |

## 4. 要构建什么

上传前，Files 根据声明大小、账号当前用量、全局用量和实时磁盘水位执行原子预留；成功后才写 stage 并 publish。失败/崩溃保留可识别状态，由有界 reconciliation 完成或回收。任何容量拒绝只暴露稳定错误，不泄漏其他账号精确用量。

## 5. 实现契约

- **入口或接缝：** Upload、Delete、Reconcile、CapacityProbe。
- **输入与输出：** owner、声明/实际字节、对象状态、磁盘可用量；输出 ready metadata 或容量 Problem 类别。
- **公共接口变化：** 模块错误先固定；OpenAPI 在 T-08。
- **不变量：** 预留不超卖；published 与计费一致；删除释放；低水位不阻断读取/删除。
- **状态或数据流：** reserve -> staged -> ready；失败/超时 -> reconcile -> ready/released。
- **错误与失败行为：** 配额、磁盘、大小漂移、I/O、重复 reconcile 均幂等且可分类。
- **兼容要求：** 现有文件通过 backfill 建立用量，不保留无配额模式。
- **安全与隐私要求：** 错误和日志不包含其他账号用量、路径或原始文件内容。

## 6. 执行路线

1. 建立并发超卖、崩溃点、低磁盘和回收红灯测试。
2. 增加 0020 migration、用量 backfill 和 quota repository。
3. 把 reserve/stage/publish/release 接入 Upload/Delete。
4. 实现磁盘双阈值、reconciliation 与稳定错误。
5. 运行双方言、文件系统故障和容量回归。

## 7. 路径访问契约

- **预计修改点/可写范围：** Files service/repository/storage 及其直接测试、quota/reconcile、0020 migration 和容量专项跨数据库测试。
- **只读上下文：** canonical Files OpenAPI、HTTP transport、typed config。
- **共享路径：** 无；配置 wiring 由 T-09，公共合同由 T-08。
- **保留或不动：** T-05 的 account lifecycle consumer 文件不修改。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | Files service | 上传/下载/删除双方言测试 | 用量与对象状态一致 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-06.md</Path>` |
| 失败路径 | 竞态/I/O | 并发超额、stage/publish 崩溃、低磁盘 | 不超卖、可恢复、读取/删除可用 | 同上 |
| 回归 | Files | `cd go-admin-plus && go test ./internal/modules/files/... ./test/files/... -race -count=1` | 文件回归通过 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行双方言、race、故障注入与存储权限检查。
- **E2E disposition：** not-required：容量正确性由服务/磁盘故障测试证明；用户页面和真实上传链路由 T-16/T-19。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source 不运行 E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** 0020 schema/backfill -> quota service -> T-09 config -> T-08 wire -> T-16 UI。
- **兼容窗口：** 无无界模式；backfill 完成前上传保持关闭。
- **监控信号：** reserved/ready 字节和对象数、reconcile age、容量拒绝、磁盘水位。
- **回滚或前向恢复：** forward-only；预留异常由 reconcile/管理修复，磁盘紧急时继续允许删除。
- **不可逆操作与批准点：** backfill 与清理现有异常对象前由 Lead 核对汇总；不自动删除用户数据。
- **收缩条件：** 所有现有 ready 对象计入基线，旧无预留上传路径为零。

## 10. 验收标准

- [ ] `AC-027~029` 在双方言、并发和磁盘故障测试中成立。
- [ ] 低水位仍可下载/删除，日志不泄漏路径或其他账号用量。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-06.md</Path>`。
- [ ] 路径、commit、集成、result 和 E2E disposition 完整。
