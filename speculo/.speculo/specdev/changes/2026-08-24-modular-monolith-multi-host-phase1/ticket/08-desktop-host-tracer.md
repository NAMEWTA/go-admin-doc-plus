---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-08
title: 交付 Wails DesktopHost 垂直曳光弹
status: done
planning_depth: deep
planning_depth_reason: 新增跨 OS 桌面进程、CGO、WebView、loopback 与前端资源嵌入的完整运行边界
ready: true
risk: high
blocked_by: [T-03, T-06]
contract_ids: [AC-004, AC-007, AC-008, AC-017]
owner: root
expected_changes: ["<Path>go-admin-plus/.github/workflows/desktop-windows-tracer.yml</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-plus/internal/host/desktop/**</Path>", "<Path>go-admin-plus/internal/profile/desktop.go</Path>", "<Path>go-admin-plus/internal/profile/desktop_dsn_test.go</Path>", "<Path>go-admin-plus/common/actions/create.go</Path>", "<Path>go-admin-plus/common/actions/delete.go</Path>", "<Path>go-admin-plus/common/actions/index.go</Path>", "<Path>go-admin-plus/common/actions/update.go</Path>", "<Path>go-admin-plus/common/actions/view.go</Path>", "<Path>go-admin-plus/common/middleware/db.go</Path>", "<Path>go-admin-plus/common/middleware/auth.go</Path>", "<Path>go-admin-plus/common/middleware/auth_test.go</Path>", "<Path>go-admin-plus/common/middleware/header.go</Path>", "<Path>go-admin-plus/common/middleware/handler/auth.go</Path>", "<Path>go-admin-plus/cmd/migrate/migration/models/initdb.go</Path>", "<Path>go-admin-plus/config/seeds.go</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>", "<Path>go-admin-ui-plus/.gitignore</Path>", "<Path>go-admin-ui-plus/packages/runtime/src/desktop/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/utils/request.ts</Path>", "<Path>go-admin-ui-plus/apps/admin/src/utils/auth.js</Path>", "<Path>go-admin-ui-plus/tests/unit/utils/auth.spec.js</Path>", "<Path>go-admin-ui-plus/tests/unit/utils/request.spec.ts</Path>"]
writable_paths: ["<Path>go-admin-plus/.github/workflows/desktop-windows-tracer.yml</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-plus/internal/host/desktop/**</Path>", "<Path>go-admin-plus/internal/profile/desktop.go</Path>", "<Path>go-admin-plus/internal/profile/desktop_dsn_test.go</Path>", "<Path>go-admin-plus/common/actions/create.go</Path>", "<Path>go-admin-plus/common/actions/delete.go</Path>", "<Path>go-admin-plus/common/actions/index.go</Path>", "<Path>go-admin-plus/common/actions/update.go</Path>", "<Path>go-admin-plus/common/actions/view.go</Path>", "<Path>go-admin-plus/common/middleware/db.go</Path>", "<Path>go-admin-plus/common/middleware/auth.go</Path>", "<Path>go-admin-plus/common/middleware/auth_test.go</Path>", "<Path>go-admin-plus/common/middleware/header.go</Path>", "<Path>go-admin-plus/common/middleware/handler/auth.go</Path>", "<Path>go-admin-plus/cmd/migrate/migration/models/initdb.go</Path>", "<Path>go-admin-plus/config/seeds.go</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>", "<Path>go-admin-ui-plus/.gitignore</Path>", "<Path>go-admin-ui-plus/packages/runtime/src/desktop/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/utils/request.ts</Path>", "<Path>go-admin-ui-plus/apps/admin/src/utils/auth.js</Path>", "<Path>go-admin-ui-plus/tests/unit/utils/auth.spec.js</Path>", "<Path>go-admin-ui-plus/tests/unit/utils/request.spec.ts</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/application/**</Path>", "<Path>go-admin-plus/internal/platform/**</Path>", "<Path>go-admin-ui-plus/packages/api-client/**</Path>"]
shared_paths: ["<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>"]
shared_path_owners: ["<Path>go-admin-plus/go.mod</Path> => T-08", "<Path>go-admin-plus/go.sum</Path> => T-08"]
---

# Ticket T-08: 交付 Wails DesktopHost 垂直曳光弹

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/08-desktop-host-tracer.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-08.md</Path>`

## 1. 战略与来源

- **目标：** 以最小完整行为证明同一 Application、Admin dist 和 HTTP contract 能嵌入 Wails 2，并在 Mac ARM64/Windows x64 离线运行。
- **可观察产出：** 两平台开发产物无需服务器或网络即可启动、bootstrap、登录并完成 demo CRUD。
- **来源：** `AC-004`、`AC-007`、`AC-008`、`AC-017`、`DEC-002`、`DEC-003`。
- **当前事实：** 仓库无桌面入口；当前 Vite 8 不适合依赖 Wails 2 AssetServer Handler 承载业务 API。
- **Planning Depth 原因：** 新平台运行边界、CGO/WebView 与安全敏感的本地 HTTP。

## 2. 决策状态

### 已锁定决策

- Wails 2 只负责窗口、资源和 bootstrap/native actions；同进程 Gin 监听 `127.0.0.1:0`。
- DesktopRuntime 通过唯一 bootstrap binding 获得 base URL、launch token、版本和 capabilities；业务请求仍为 HTTP。
- Admin dist 由根构建编排复制到生成资产目录后 embed，不提交生成 dist。
- 本 Ticket 先交付开发签名的 tracer；完整安全、迁移恢复和正式打包在 T-09/T-11/T-12。

### 已采用的低影响假设

- 窗口尺寸、菜单和开发图标沿用 Wails 默认/现有 Admin 最小适配，不进行视觉设计。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| Wails entry、资源 embed、loopback listener、bootstrap、DesktopRuntime、demo CRUD | Application、Desktop profile、Admin dist、ApiClient | 正式签名安装包、迁移恢复、完整安全 hardening |

## 4. 要构建什么

开发者在原生 Mac/Windows 构建桌面应用。启动时创建临时/应用数据 SQLite、Application 与 loopback listener，Wails 加载嵌入 Admin，bootstrap 后完成登录和 demo CRUD。断网不影响静态资源或核心请求；关闭窗口触发 context cancel 和基本清理。

## 5. 实现契约

- **入口或接缝：** desktop main、Wails lifecycle、bootstrap binding、DesktopRuntime。
- **输入与输出：** 嵌入 dist、desktop profile、窗口 lifecycle；输出运行窗口和 loopback HTTP。
- **公共接口变化：** 新 bootstrap interface；业务 HTTP 不变。
- **不变量：** bind loopback/随机端口；不使用 Wails binding 做 CRUD；资源无 CDN。
- **状态或数据流：** app launch -> profile/Application/listener -> Wails -> bootstrap -> ApiClient -> Gin。
- **错误与失败行为：** 资源缺失、bind/bootstrap/Application 失败时不显示半工作 UI，返回可诊断开发错误。
- **兼容要求：** 同一 Admin build 与 `/api/v1` contract。
- **安全与隐私要求：** token 不进 URL/日志；仅 bootstrap 进程内返回。

## 6. 执行路线

1. 固定 Wails 2 版本并建立原生 doctor/build smoke。
2. 创建 DesktopHost 和生成 assets seam，启动 Application + loopback。
3. 实现最小 bootstrap 与 DesktopRuntime，接入 ApiClient。
4. 嵌入 Admin dist 并跑登录/demo CRUD 离线 tracer。
5. 验证关闭、资源缺失、bootstrap 失败和两平台构建。
6. 以手动触发、只读权限的 Windows 原生 workflow 运行同一 tagged tracer；workflow 必须固定后端 ref、接受显式前端 ref、硬超时并保存环境/stdout/stderr，不能把 workflow 文件本身当作运行证据。

## 7. 路径访问契约

- **预计修改点/可写范围：** desktop cmd/host、Go manifests、DesktopRuntime，以及 Admin request 入口唯一的 Host Runtime 选择接缝。实现预检确认该入口当前固定创建 WebRuntime；若不修改这一行，同一 Admin dist 无法消费 Wails bootstrap，桌面请求会错误落到资源 origin。WebKit 的 `wails://` origin 无法可靠使用 Cookie，故 Desktop token 由同一 auth seam 写入 localStorage；Desktop transport 使用原生 fetch 避免 XHR 在自定义 scheme 下阻塞，Web 仍保留 Axios transport。真实登录 tracer 另发现 go-admin-core 默认成功回调缺少 canonical `msg/data`，因此允许在既有 AuthInit/Logout 接缝补齐统一 Web/Desktop envelope 与最小测试；生产 ApiClient 不放宽，业务 API 不分叉。首次真实 SQLite tracer 发现初始 migration 运行时依赖 cwd 下 `config/*.sql`，故允许把同一 seed 静态资源嵌入 config 并从原 InitDb 接缝读取；另允许通用 CORS middleware 保留 Desktop gateway 已设置的精确 Origin，而非覆盖为通配符。原生 CRUD 的 race 验证发现复用的 `gin.Context` 被传给 `database/sql` 后在请求结束异步读取，故允许通用 Action/DB middleware 改传稳定的 `Request.Context()`。Wails frozen install 会在自定义 frontend 根生成标准 `package.json.md5` 缓存标记，允许前端 `.gitignore` 忽略这一个工具缓存，构建产物与 lockfile 仍不提交。当前 macOS 环境无法执行 Windows 原生 Gate，故允许新增一条专用手动 workflow 和可复用 PowerShell runner；它只提供 tracer 证据，不执行签名、发布或生产动作。Windows 原生 runner 随后发现 Desktop profile 把 `D:/...` 编码为 `file:D:/...`，modernc SQLite 将 `D:` 误判为 URI authority；因此允许在原 profile 数据库打开接缝修正 drive-letter URL，并以纯 DSN 回归测试覆盖，无需扩大 Application/platform 权限。
- **只读上下文：** Application/platform/ApiClient，以及除两个显式 writable utility 接缝外的 Admin 源码。
- **共享路径：** Go manifests 由 T-08 唯一拥有并固定 Wails 2。
- **保留或不动：** Domain 页面、ServerHost、Docker、发布凭据和用户数据。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 原生桌面 tracer | 两 OS build/run/login/CRUD | 完全离线通过 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-08.md</Path>` |
| Windows runner | 手动原生 tracer workflow | 固定 backend dispatch ref + 显式 frontend ref；构建后运行并要求唯一 PASS marker | 超时/非零退出/FAIL/marker 缺失均失败并保留诊断 | 同上 |
| 失败路径 | bind/assets/bootstrap | 注入缺资源和启动错误 | 不显示半工作 UI | 同上 |
| 回归 | Web build/API | pnpm build/e2e + Go tests | Web/Server 不回归 | 同上 |

- **Workspace checks：** 非 E2E Go/pnpm tests；各原生 runner 构建。
- **E2E disposition：** required；跨 Wails/WebView/HTTP/SQLite 完整边界。
- **E2E owner/environment：** Lead / parent-candidate 或 current-workspace；macOS ARM64 与 Windows AMD64 原生、网络隔离。
- **Integration evidence：** 提交、candidate/result、平台构建 SHA 和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** dev tracer 先行，不替代 Web；T-09 harden 后才允许正式发布。
- **兼容窗口：** bootstrap schema 在 Phase 1 内向后兼容；业务接口只走 HTTP。
- **监控信号：** 本地启动阶段、listener、bootstrap 和 shutdown 日志。
- **回滚或前向恢复：** 删除桌面入口不影响 Server/Web；已创建开发数据由用户显式清理。
- **不可逆操作与批准点：** 无正式分发；签名/发布留给平台 Ticket。
- **收缩条件：** 不适用；本 Ticket 是后续 hardening 的稳定 tracer。

## 10. 验收标准

- [x] `AC-004`、`AC-007`、`AC-008`、`AC-017` 的 tracer 范围通过。
- [x] 业务 CRUD 无 Wails binding，网络隔离下无 CDN/公网依赖。
- [x] 两平台正常/失败/回归 Evidence 完整。
- [x] 路径、提交、集成和 required 原生 E2E 合同满足。
- [x] tracer 未被标记为正式发布产物，无未批准偏差。
