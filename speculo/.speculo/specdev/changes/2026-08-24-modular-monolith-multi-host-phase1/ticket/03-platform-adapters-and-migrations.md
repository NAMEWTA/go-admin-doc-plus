---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-03
title: 建立 Server/Desktop 基础设施适配与可恢复迁移
status: done
planning_depth: deep
planning_depth_reason: 引入双数据库、文件、租户、队列和任务生命周期接口，并承担数据迁移完整性
ready: true
risk: high
blocked_by: [T-02]
contract_ids: [AC-010, AC-011, AC-015, AC-017]
owner: root
expected_changes: ["<Path>go-admin-plus/internal/platform/**</Path>", "<Path>go-admin-plus/internal/profile/**</Path>", "<Path>go-admin-plus/internal/tenant/**</Path>", "<Path>go-admin-plus/app/jobs/**</Path>", "<Path>go-admin-plus/cmd/migrate/**</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/platform/**</Path>", "<Path>go-admin-plus/internal/profile/**</Path>", "<Path>go-admin-plus/internal/tenant/**</Path>", "<Path>go-admin-plus/app/jobs/**</Path>", "<Path>go-admin-plus/cmd/migrate/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/application/**</Path>", "<Path>go-admin-plus/common/database/**</Path>", "<Path>go-admin-plus/common/storage/**</Path>", "<Path>go-admin-plus/config/**</Path>"]
shared_paths: ["<Path>go-admin-plus/cmd/migrate/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus/cmd/migrate/**</Path> => T-03"]
---

# Ticket T-03: 建立 Server/Desktop 基础设施适配与可恢复迁移

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/03-platform-adapters-and-migrations.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-03.md</Path>`

## 1. 战略与来源

- **目标：** 在 Application 的真实变化接缝处提供 Server 与 Desktop 两套数据库、缓存、队列、文件、租户和任务适配器，并统一迁移语义。
- **可观察产出：** 相同 FileStore/Queue/Tenant/Migration contract 在 PostgreSQL/服务端与 SQLite/桌面 fixture 上通过；任务可由 context 停止。
- **来源：** `AC-010`、`AC-011`、`AC-015`、`AC-017`、`CODE:<Path>go-admin-plus/app/jobs/jobbase.go</Path>`。
- **当前事实：** SQLite 由 CGO build tag 提供；jobs setup 永久阻塞；队列和 DB 通过 sdk.Runtime 全局获取。
- **Planning Depth 原因：** 数据、迁移和后台生命周期为高事故半径 Deep 工作。

## 2. 决策状态

### 已锁定决策

- Server profile 参考 PostgreSQL、Redis、持久卷 FileStore 和 ServerTenantResolver；Desktop profile 使用 SQLite、内存 Queue/Cache、AppData FileStore 和固定 `local` 租户。
- SQLite 启用 foreign keys、busy timeout 和 WAL；只允许单应用进程写入。
- 迁移版本不可修改；方言差异由迁移实现内部适配，语义和版本序列共用。
- Desktop 错过的 cron 默认不补跑；HTTP Job 只有用户配置后才允许访问网络。

### 已采用的低影响假设

- 服务端现有其他数据库 adapter 暂时保留为兼容路径，但本 Ticket 只保证 Compose 参考 PostgreSQL 和 Desktop SQLite 的新 contract。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| Profile、adapter interface、双 DB fixture、迁移 runner、可取消 jobs/queue | GORM 模型、现有迁移版本、现有 storage/queue 实现 | Host 监听、桌面窗口、Docker、云同步 |

## 4. 要构建什么

Application 根据 profile 接收已构建基础设施。业务调用相同深层接口完成数据库、文件、队列和租户操作；Desktop fixture 的 SQLite 升级先备份、成功后提交 schema，失败时保留原数据和备份。停止 context 后 cron 与队列退出，不再永久 `select {}`。

## 5. 实现契约

- **入口或接缝：** Profile builder、MigrationRunner、FileStore/Queue/Cache/TenantResolver、jobs Module lifecycle。
- **输入与输出：** profile 配置和平台数据根；输出 Dependencies 或可诊断错误。
- **公共接口变化：** 无业务 HTTP 变化。
- **不变量：** 数据目录隔离；迁移追加；Desktop 单租户；关闭后不接受新任务。
- **状态或数据流：** profile config -> adapters -> migration -> Application dependencies -> cancel/close。
- **错误与失败行为：** 依赖或迁移失败阻止 ready；备份失败阻止 Desktop 迁移；停止错误聚合返回。
- **兼容要求：** 现有 server 配置可通过兼容 builder 继续解析。
- **安全与隐私要求：** 文件路径防止越出数据根；日志不包含 DSN 密码或 token。

## 6. 执行路线

1. 以运行中的内存/临时 adapter contract tests 定义接口和失败行为。
2. 建立 server/desktop profile builder，先包裹现有实现。
3. 改造 queue/jobs 为 context 生命周期并验证停止。
4. 实现 SQLite 选项、平台数据根 FileStore 和固定租户。
5. 建立 PostgreSQL/SQLite migration fixture，验证成功、备份失败和迁移失败。
6. 运行全量 API characterization，确认 adapter 变化不可见。

## 7. 路径访问契约

- **预计修改点/可写范围：** frontmatter 所列 platform/profile/tenant/jobs/migrate。
- **只读上下文：** Application、现有 database/storage/config。
- **共享路径：** 迁移入口由 T-03 唯一拥有；已发布 version 文件只读，不得重写。
- **保留或不动：** 前端、Host、Docker、真实用户数据库和上传目录。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | Adapter/Migration contract | Go contract tests + 双 fixture | 两 profile 语义一致 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-03.md</Path>` |
| 失败路径 | 迁移/备份/取消注入 | 注入错误与 context cancel | fail closed、可恢复、worker 退出 | 同上 |
| 回归 | API characterization | `go test ./...` | 无业务行为回归 | 同上 |

- **Workspace checks：** 原生 CGO 环境运行 SQLite tests；PostgreSQL 使用隔离容器；执行 Go 全量测试/build。
- **E2E disposition：** required；跨数据库、文件和后台生命周期。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；执行双 fixture upgrade 与真实关闭。
- **Integration evidence：** implementation/source commit、parent before、candidate/result、父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 先加接口和兼容 adapter，再切 Application profile，最后迁移 jobs；数据先备份后迁移。
- **兼容窗口：** 旧 Runtime 获取路径保留到所有 Module 消费新 Dependencies，T-13 扫描后收缩。
- **监控信号：** migration version/result、adapter init、queue/jobs start/stop。
- **回滚或前向恢复：** 未执行迁移可回滚代码；已迁移 Desktop 使用备份恢复，Server 使用数据库备份/前向迁移。
- **不可逆操作与批准点：** 对已发布 migration 的任何修改禁止；生产迁移运行需发布 Gate 批准。
- **收缩条件：** sdk.Runtime 基础设施直接调用为零、双 profile tests 与 characterization 通过。

## 10. 验收标准

- [ ] `AC-010`、`AC-011`、`AC-015`、`AC-017` 通过。
- [ ] SQLite/PostgreSQL、成功/失败/恢复证据完整。
- [ ] worker 可取消，无永久阻塞。
- [ ] Evidence、路径、提交、集成和 required E2E 合同满足。
- [ ] 未修改历史迁移或真实用户数据，无未批准偏差。
