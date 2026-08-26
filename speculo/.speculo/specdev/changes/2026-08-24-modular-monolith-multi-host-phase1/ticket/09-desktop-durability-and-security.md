---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-09
title: 完成桌面数据耐久、生命周期与 loopback 安全
status: done
planning_depth: deep
planning_depth_reason: 处理本地数据升级恢复、单实例、网络攻击面和全部后台资源关闭，事故半径为用户离线数据
ready: true
risk: critical
blocked_by: [T-07, T-08]
contract_ids: [AC-009, AC-010, AC-011, AC-015, AC-017]
owner: root
expected_changes: ["<Path>go-admin-plus/.github/workflows/desktop-windows-tracer.yml</Path>", "<Path>go-admin-plus/internal/host/desktop/**</Path>", "<Path>go-admin-plus/internal/platform/desktop/**</Path>", "<Path>go-admin-plus/internal/profile/upgrade.go</Path>", "<Path>go-admin-plus/internal/profile/directory_sync_*.go</Path>", "<Path>go-admin-plus/test/desktop/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/main.go</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/main_test.go</Path>"]
writable_paths: ["<Path>go-admin-plus/.github/workflows/desktop-windows-tracer.yml</Path>", "<Path>go-admin-plus/internal/host/desktop/**</Path>", "<Path>go-admin-plus/internal/platform/desktop/**</Path>", "<Path>go-admin-plus/internal/profile/upgrade.go</Path>", "<Path>go-admin-plus/internal/profile/directory_sync_*.go</Path>", "<Path>go-admin-plus/test/desktop/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/main.go</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/main_test.go</Path>"]
read_only_paths: ["<Path>go-admin-plus/cmd/go-admin-desktop/assets.go</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/e2e_*.go</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/main_bindings.go</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/scripts/**</Path>", "<Path>go-admin-plus/internal/application/**</Path>", "<Path>go-admin-plus/internal/profile/desktop.go</Path>", "<Path>go-admin-plus/internal/profile/server.go</Path>", "<Path>go-admin-plus/internal/profile/*_test.go</Path>", "<Path>go-admin-plus/cmd/migrate/**</Path>", "<Path>go-admin-ui-plus/apps/admin/**</Path>"]
shared_paths: ["<Path>go-admin-plus/internal/host/desktop/**</Path>"]
shared_path_owners: ["<Path>go-admin-plus/internal/host/desktop/**</Path> => T-09"]
---

# Ticket T-09: 完成桌面数据耐久、生命周期与 loopback 安全

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/09-desktop-durability-and-security.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-09.md</Path>`

## 1. 战略与来源

- **目标：** 将桌面 tracer 提升为可发布核心：平台数据目录、备份迁移、单实例、严格 loopback 授权、文件沙箱和后台有序停止。
- **可观察产出：** 旧数据升级/失败恢复可判定，局域网不可访问，缺 token/错误 Origin 被拒绝，双开不产生第二写入者。
- **来源：** `AC-009`、`AC-010`、`AC-011`、`AC-015`、`AC-017`。
- **当前事实：** T-08 只保证 tracer；正式数据与安全门必须在平台打包前完成。
- **Planning Depth 原因：** 数据完整性与本机攻击面为 critical。

## 2. 决策状态

### 已锁定决策

- macOS 使用 Application Support，Windows 使用 LocalAppData；内部固定 `db/files/logs/backups/temp`。
- 单实例锁先于数据库打开；第二实例激活已有窗口或稳定退出。
- 请求同时校验 loopback、每启动随机 token 与精确 Origin；CORS 不允许 wildcard。
- 备份成功才迁移，失败 fail closed；恢复不自动覆盖现有数据库。
- jobs/queue/http/database/logging 全部挂根 context，并有关闭顺序。

### 已采用的低影响假设

- 备份保留策略采用版本/时间命名且至少保留最近成功迁移前备份；具体数量可配置并由磁盘测试验证。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| AppData、backup/migrate、single instance、token/origin、FileStore 沙箱、shutdown/recovery UI contract | T-03 adapters、T-08 Host、Wails single instance | 自动更新、云备份、正式签名打包 |

## 4. 要构建什么

用户启动桌面应用时，Host 在打开 DB 前获得锁，创建受控数据目录，备份并迁移旧数据库，之后才显示 UI。所有 HTTP 请求只能来自当前窗口 bootstrap 会话；本机其他进程的无令牌请求和局域网请求失败。关闭和异常启动不损坏数据，日志给出恢复所需版本和备份位置。

## 5. 实现契约

- **入口或接缝：** DesktopHost startup/shutdown、AppData resolver、backup/migration coordinator、loopback middleware。
- **输入与输出：** 平台 OS/user、旧数据、请求 token/origin；输出 ready app、备份或安全拒绝。
- **公共接口变化：** bootstrap 增加稳定安全字段；业务 API 不变。
- **不变量：** 单 writer；安装目录只读；文件路径不越界；UI 不早于迁移成功。
- **状态或数据流：** lock -> dirs -> backup -> migrate -> listener/token -> UI -> cancel -> reverse close。
- **错误与失败行为：** 锁/磁盘/备份/迁移失败均不打开业务 UI；安全失败返回稳定 401/403 且不泄 token。
- **兼容要求：** T-08 tracer 与 T-07 全 Admin 流程通过。
- **安全与隐私要求：** token 只驻内存和受控 header；日志脱敏；数据目录权限遵循平台默认私有范围。

## 6. 执行路线

1. 为数据目录、双开、token/origin、shutdown 和恢复建立失败优先测试。
2. 实现单实例与平台目录，接入 FileStore 沙箱。
3. 接入备份迁移 coordinator 和恢复状态，不显示半迁移 UI。
4. 强化 loopback middleware 与 DesktopRuntime header/origin。
5. 完成 worker/listener/DB 关闭顺序和重复启停测试。
6. 在两平台网络隔离环境运行升级、失败恢复、双开和安全探测。

## 7. 路径访问契约

- **预计修改点/可写范围：** desktop host/platform/tests；实现预检确认 Go 的 `os.UserConfigDir` 在 Windows 指向 Roaming AppData，与 Ticket 锁定的 LocalAppData 不一致，因此允许桌面 main 及其测试仅把数据根解析委托给新的 platform/desktop resolver，Wails window/assets/tracer 仍保持只读。T-09 的 Windows 原生 OS lock 无法由 macOS cross-build 证明，因此允许在 T-08 已有的手动 Windows tracer 中增加 native hardening tests；固定 SHA、最小权限、构建和 GUI marker 合同不变。Windows run `32851710313` 进一步证明 POSIX directory `fsync` 在 Windows 返回 `Access is denied`，因此允许仅将 backup directory sync 拆成 Unix/Windows 平台实现，并把 profile tests 加入 native step；逐文件 `Sync+Close`、备份发布顺序和 migration 行为不变。
- **只读上下文：** 除上述入口和 directory sync seam 外的 desktop cmd、Application、profile layout/server/tests、migration registry 和 Admin。
- **共享路径：** DesktopHost 由 T-09 唯一 hardening owner；平台 release Ticket 只消费。
- **保留或不动：** 用户真实数据、ServerHost、前端 Domain、签名配置。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 原生升级/重启 | fixture upgrade、CRUD、restart | 数据/文件保留 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-09.md</Path>` |
| 失败路径 | backup/migrate/security/lock | 注入磁盘/迁移错误、双开、socket probe | fail closed、不可达、可恢复 | 同上 |
| 回归 | 全 Admin desktop/web | 原生 E2E + Go/pnpm suites | 无 Host 回归 | 同上 |

- **Workspace checks：** Go race/生命周期/迁移 tests、pnpm tests；原生 runner 非 E2E build。
- **E2E disposition：** required；数据、安全、WebView 和 OS 生命周期。
- **E2E owner/environment：** Lead / parent-candidate 或 current-workspace；两平台、断网、局域网探测和升级 fixture。
- **Integration evidence：** 提交、candidate/result、平台/网络环境和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 单实例 -> 备份 -> migration -> listener/UI；旧数据 fixture 覆盖每个已发布基线。
- **兼容窗口：** bootstrap 旧客户端只在同版本嵌入资源中存在；不支持跨版本外部浏览器调用 Desktop API。
- **监控信号：** 本地 migration/backup/result、security rejection、shutdown duration/error。
- **回滚或前向恢复：** 迁移前备份 + 显式恢复；禁止静默自动覆盖，已成功写入新 schema 后优先前向修复。
- **不可逆操作与批准点：** 删除备份、改变 AppData 身份或发布不可逆 migration 需单独人工批准。
- **收缩条件：** 不适用；安全与恢复能力长期保留。

## 10. 验收标准

- [x] `AC-009`、`AC-010`、`AC-011`、`AC-015`、`AC-017` 通过。
- [x] 双开、迁移失败、局域网和无 token/错误 Origin 均有原生 Evidence。
- [x] 路径、提交、集成和 required E2E 合同满足。
- [x] 未触碰真实用户数据或降低安全断言，无未批准偏差。
