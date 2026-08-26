---
schema_version: 3
artifact: spec
change: 2026-08-24-modular-monolith-multi-host-phase1
status: ready
ready_for_tickets: true
sources:
  - "USER-DECISION:Phase 1 仅交付单一管理前端、macOS ARM64、Windows AMD64 与 Linux Docker"
  - "USER-DECISION:采用模块化单体加多 Host 适配器并按上一轮完整规划固化 Spec"
  - "CODE:<Path>go-admin-plus/cmd/api/server.go</Path>"
  - "CODE:<Path>go-admin-ui-plus/src/utils/request.ts</Path>"
  - "RESEARCH:<Url>https://wails.io/docs/reference/cli/</Url>"
  - "RESEARCH:<Url>https://v3.wails.io/status/</Url>"
---

# Spec: Phase 1 模块化单体与多 Host 管理端

- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **当前 ADR：** 不适用；本 change 的架构约束直接由本 Spec 锁定，尚未进入永久知识提升阶段
- **当前领域上下文：** 不适用；沿用项目中的 admin、jobs、demo、other 等模块名称

## 1. 问题与目标

### 问题陈述

当前 Go 启动入口把配置、全局运行时、队列、定时任务、路由、HTTP 监听和 OS 信号组合在同一流程中，无法将同一业务核心安全嵌入桌面生命周期。当前前端是单根 Vue 应用，请求模块同时承担传输、登录态、UI 弹窗和页面跳转，API 地址依赖构建期环境，无法在同一构建产物中适配 Web 与桌面运行时。现有 Docker 交付还内置演示配置与数据库、使用特权容器，且没有 Web、迁移、数据库和就绪编排。

### 目标用户与场景

- 系统管理员在 Linux 服务器上通过 Docker Compose 部署 Web 管理端和统一后端。
- 离线用户在 Apple Silicon Mac 或 Windows x64 设备上安装桌面管理端，在完全断网环境中使用同一组管理能力。
- 开发者以一套 Go 业务模块、一份 HTTP 契约和一个 Admin 前端持续开发，不复制服务器版与桌面版业务代码。
- 发布维护者可从固定的前后端 submodule 提交生成可验证、可升级的三类交付物；Phase 1 桌面包面向自有设备，允许用户明确确认后运行的 unsigned self-use 分发。

### 成功标准

- 同一个 Go Application 可由 ServerHost 和 DesktopHost 构建、启动和有序停止，Host 之外不处理 OS 信号、监听地址或窗口生命周期。
- Web 与桌面通过同一 `/api/v1` 业务契约完成登录、菜单、权限、CRUD、上传和任务管理；不形成第二套业务 RPC。
- 单一 Admin 前端成为 workspace 中的可部署 App，并通过明确 Domain 与共享 Package 组合。
- macOS ARM64 和 Windows AMD64 安装包可在断网机器首次安装、启动、持久化数据、升级和恢复迁移失败。
- Linux AMD64 Compose 从空卷启动后提供 Web、API、迁移、PostgreSQL 和 Redis，并满足最小权限与健康检查要求。
- OpenAPI、生成的 TypeScript 类型、跨 Host contract test 和发布流水线阻止接口及交付漂移。

### 非目标

- 不交付 Mobile、Linux 桌面、macOS Intel、Windows ARM64 或 Linux ARM64 镜像。
- 不交付普通 Web 用户端、Mobile 管理端或独立 Mac 用户端等额外 App。
- 不实现桌面数据与服务器数据同步、冲突合并、账号漫游或云端备份。
- 不把模块化单体拆成微服务，不重写既有全部业务功能或 UI 视觉体系。
- 不在 Phase 1 购买或要求 Developer ID、Apple notarization、Windows Authenticode、Store 上架或企业受管设备绕过策略；这些属于未来可信公共分发增强。

## 2. 解决方案与外部行为

### 解决方案摘要

建立一个无 Host 假设的 Go Application。Application 组合显式 ModuleSet，提供 HTTP Handler 与 `Start(context.Context)`、`Stop(context.Context)` 生命周期；ServerHost 将其接入网络 HTTP、服务端数据库、Redis 和容器信号，DesktopHost 将其接入 Wails 2、loopback HTTP、SQLite、内存基础设施、单实例锁和桌面数据目录。

前端重构为 pnpm workspace：`apps/admin` 是唯一可部署 App；`domains/system|jobs|demo|tools|monitor` 对应业务能力；`packages/contracts|api-client|runtime|app-core|ui` 提供生成契约、纯传输、Host 启动配置、应用壳与共享 UI。WebRuntime 使用同源 `/api/v1`，DesktopRuntime 仅通过 Wails bootstrap 获取本次 loopback 地址和启动令牌，随后所有业务请求仍使用 HTTP。

### 主要流程

1. Web 部署者启动 Compose；一次性 Migration 成功后 API 进入 ready，Nginx 向浏览器提供 Admin 静态资源并同源代理 `/api`、`/login` 和静态文件。
2. 桌面用户启动应用；DesktopHost 获得单实例锁，解析平台应用数据目录，迁移前备份 SQLite，运行迁移，启动仅绑定 loopback 随机端口的 Application，再显示 Wails 窗口。
3. Admin 启动时由 Runtime Adapter 解析 API 位置；登录、菜单、权限、CRUD、文件和任务请求都经过同一个 ApiClient 和同一 OpenAPI 契约。
4. 关闭服务器或桌面窗口时，Host 取消根 context；任务、队列和 HTTP Server 在超时边界内停止，数据库与日志完成释放，不遗留永久 goroutine。
5. 发布流水线在原生 macOS ARM64 和 Windows AMD64 runner 上构建、测试并打包明确标记的 unsigned self-use 产物；Linux runner 构建固定 digest 的 Web/API 镜像与 Compose bundle。桌面用户先验证 SHA-256，再使用操作系统提供的单个应用确认/例外路径运行。

### 边界、失败与稳定错误行为

- Application 配置、数据库、迁移或必需 Module 启动失败时，不进入 ready；ServerHost 退出非零，DesktopHost 在不暴露敏感细节的恢复界面中给出日志位置和重试/退出路径。
- Desktop SQLite 迁移失败时保留迁移前备份，不显示业务 UI、不继续写入半迁移数据库；恢复动作不得静默删除用户数据。
- Desktop loopback 地址不可用时可重新选择 OS 分配端口；不得回退到固定公网监听地址或关闭启动令牌校验。
- 第二个桌面实例不得启动第二个 SQLite 写入者；应激活已有窗口或给出稳定提示后退出。
- 缺少 WebView2 的 Windows 离线机器由安装包内的离线依赖完成安装或给出可判定失败；不得要求联网下载。
- macOS/Windows 自用包必须明确显示未获平台发行签名、可能触发 Gatekeeper/SmartScreen，且受管设备或 Windows Smart App Control 可能禁止用户继续；不得建议全局关闭 Gatekeeper、SmartScreen、Defender 或 Smart App Control。
- 前端收到未授权、业务错误或网络错误时，传输层返回标准化错误，App Shell 负责一次性用户提示和登录态转换，Domain 不重复弹窗。
- Compose 迁移失败时 API 不启动；PostgreSQL、Redis 或文件存储不可用时 readiness 为失败，liveness 只反映进程是否存活。

### 状态转换与不变量

- Application 状态仅按 `constructed -> starting -> ready -> stopping -> stopped` 转换；启动失败直接进入 stopped/failed，不允许部分 ready。
- 同一发布版本中 Web 与 Desktop 使用相同 OpenAPI 契约和响应 envelope；Host 不得改变业务字段语义、授权规则或路由路径。
- 桌面业务数据只写平台应用数据目录；嵌入资源和安装目录视为只读。
- 版本化迁移一经发布不得修改；新变更以新迁移版本追加，并在 SQLite 与服务端目标数据库上验证。
- Domain 不等于可部署 App；一个 `apps/admin` 组合多个 Domain，Domain 不直接相互深层导入。
- Wails Binding 不承载业务 CRUD，只承载 bootstrap 和明确的原生桌面动作。

## 3. 用户故事

- **US-001**：作为系统管理员，我希望通过一个 Compose bundle 部署 Web 管理端和统一后端，以便在单台 Linux 服务器上可靠运行和升级。
- **US-002**：作为离线桌面用户，我希望在 Mac ARM64 或 Windows x64 安装后无需网络即可使用管理能力，以便在隔离环境工作。
- **US-003**：作为管理员，我希望 Web 与桌面的登录、菜单、权限和业务操作保持一致，以便不学习两套产品行为。
- **US-004**：作为桌面用户，我希望数据在关闭和升级后保留，迁移失败时可恢复，以免升级破坏本地数据。
- **US-005**：作为开发者，我希望业务核心与 Host 生命周期分离，以便新增运行形态时不复制或分叉业务模块。
- **US-006**：作为前端开发者，我希望 Admin App 通过 Domain 与共享 Package 组合，以便业务模块可定位、可测试并避免平台判断散落。
- **US-007**：作为接口维护者，我希望 OpenAPI 自动生成前端契约并在 CI 检查，以便阻止 Go DTO 与 TypeScript 漂移。
- **US-008**：作为早期自用发布维护者，我希望获得可校验、明确标识信任状态的 Mac、Windows 和 Linux 产物，以便不购买代码签名证书也能在自有设备重复构建、安装和升级。
- **US-009**：作为运维人员，我希望能区分进程存活、依赖就绪和迁移失败，以便自动化部署与故障定位。

## 4. 验收合同

| ID | 前置条件 | 动作或事件 | 可观察结果 | 验证接缝 |
|---|---|---|---|---|
| AC-001 | Go Application 使用测试配置和显式 ModuleSet | 分别由测试 ServerHost 与 DesktopHost 启停 | 两者暴露同一 Handler 行为并完成有序停止，无 Host 逻辑进入业务 Module | Go 生命周期与 handler 集成测试 |
| AC-002 | 记录改造前 `/api/v1` 基线 | 对登录、菜单、权限、CRUD、上传和任务执行兼容请求 | 路径、方法、响应 envelope 与授权结果保持兼容；批准的新增 capability/health 接口除外 | API contract/characterization test |
| AC-003 | 全新安装前端依赖 | 构建 `apps/admin` | 生成可部署 Web 静态资源，现有核心管理页面与动态菜单可访问 | pnpm build、Playwright mocked/live |
| AC-004 | 同一份 Admin 构建产物分别运行于 WebRuntime 和 DesktopRuntime | 启动应用并发起业务请求 | Web 使用同源路径，Desktop 使用 bootstrap 返回的 loopback 地址；Domain 无平台分支 | Runtime 单元测试与双 Host E2E |
| AC-005 | 后端 OpenAPI 已生成 | 运行契约生成与漂移检查 | `packages/contracts` 可重复生成且工作区无未提交差异；ApiClient 类型检查通过 | OpenAPI generation diff、type-check |
| AC-006 | 后端菜单同时存在旧 component 数据和新 routeKey 数据 | Admin 加载动态路由 | 兼容期两种数据均能解析到白名单页面；未知键稳定进入不可用/404 行为而不执行任意导入 | 路由单元测试与 Playwright |
| AC-007 | Apple Silicon Mac 完全断网且无已有应用数据 | 安装并启动发布产物 | 应用启动、登录并完成代表性 CRUD，重启后数据仍存在 | 原生 macOS 离线 E2E |
| AC-008 | Windows 10/11 x64 完全断网且 WebView2 缺失或版本不满足 | 运行离线安装包并启动应用 | 安装包提供所需运行时，应用完成代表性登录和 CRUD | 原生 Windows 离线 E2E |
| AC-009 | DesktopHost 已启动 | 从本机和局域网探测监听端口，并发送缺令牌/错误 Origin 请求 | 只存在 loopback 随机监听；未授权请求被拒绝；局域网不可达 | socket 探测与安全集成测试 |
| AC-010 | 已有旧版本 SQLite 数据 | 启动新版本并触发迁移成功或注入迁移失败 | 成功时数据保留且 schema 升级；失败时业务 UI 不启动且迁移前备份可用于恢复 | SQLite fixture upgrade/rollback test |
| AC-011 | DesktopHost 运行任务和队列 | 关闭窗口或再次启动程序 | 后台工作收到取消并停止；第二实例不创建第二写入者 | 生命周期、goroutine 与单实例测试 |
| AC-012 | 空 Docker 主机和空持久卷 | 启动生产 Compose | Migration 先成功，随后 PostgreSQL、Redis、API、Web healthy，浏览器可同源登录 | Compose smoke/E2E |
| AC-013 | API 依赖正常、异常或迁移未完成 | 请求 live、ready 与 capabilities | live 与 ready 语义可区分；capabilities 正确反映 Host 能力且不泄露敏感配置 | HTTP 集成测试 |
| AC-014 | 检查生产容器定义与运行实例 | 执行容器安全检查 | 无 privileged，应用非 root，API 默认不发布宿主端口，配置/secret 不烘焙进镜像 | Compose config 与运行时检查 |
| AC-015 | Desktop 与 Server 分别上传、读取文件 | 通过相同文件接口操作 | Desktop 写入应用数据目录；Server 写入持久卷或配置的对象存储；HTTP 行为一致 | FileStore contract test 与 Host 集成测试 |
| AC-016 | 创建版本标签 | 执行三平台发布流水线 | 产出 macOS ARM64 self-use DMG、Windows x64 self-use NSIS、Linux AMD64 Web/API 镜像和 Compose bundle，并附版本、checksum、SBOM、源码 provenance 与明确 trust state；桌面包不得伪称 Developer ID/Authenticode signed | 原生 CI release verification |
| AC-017 | 桌面设备在运行期间无网络 | 执行登录、菜单、CRUD、上传和本地任务场景 | 除明确由用户配置的 HTTP Job 外，核心流程不访问公网或 CDN | 网络隔离 E2E 与请求观察 |
| AC-018 | 所有迁移 Ticket 完成 | 扫描旧全局注册、旧请求入口和物理 component 路径消费者 | 兼容层只在批准窗口内存在；达到零调用与回归通过后可收缩 | 调用点扫描、构建与回归套件 |

## 5. 范围

### IN

- Go Application、Module、Lifecycle 和 ServerHost/DesktopHost 接缝。
- admin、jobs、demo、other 现有业务 Module 的显式装配与兼容迁移。
- 服务端 PostgreSQL/Redis 参考适配器、桌面 SQLite/内存适配器、文件与租户适配器。
- 单一 Admin 前端 workspace、Domain、Runtime、ApiClient、Contracts、App Core 和共享 UI。
- Wails 2 macOS ARM64 与 Windows AMD64 桌面壳、loopback 安全、单实例和离线依赖。
- Linux AMD64 API/Web 容器、Compose、迁移、健康检查、持久化和发布流水线。
- OpenAPI、characterization、contract、unit、integration、Playwright 和原生离线 E2E。

### REUSE

- 复用现有 Gin 路由、Api/Service/Model 分层、GORM 模型、不可变版本迁移、JWT/Casbin 权限和 `/api/v1` envelope。
- 复用 Vue 3、Vite 8、Element Plus、Pinia、ProTable、PageContainer、现有页面和 Playwright 测试。
- 复用现有配置可选择的服务端数据库能力；生产 Compose 的参考实现固定为 PostgreSQL，不删除其他服务端数据库适配器。
- 复用根仓库 submodule 固定提交模式，由根仓库承担跨仓库构建编排。

### OUT

- **OOS-001**：不实现任何云端/桌面双向同步或离线冲突解决；该能力需要独立数据协议 Spec。
- **OOS-002**：不创建除 `apps/admin` 外的空 App 占位目录；未来 App 在有真实交付需求时新增。
- **OOS-003**：不采用 Wails 3 Beta、Tauri、Electron 或 Wails 业务 Binding 作为第二业务接口。
- **OOS-004**：不改变现有业务角色模型、权限语义和公开 `/api/v1` 行为，除本 Spec 明确新增的运行能力接口。
- **OOS-005**：不提供自动更新器、应用商店上架、增量更新或遥测收集。
- **OOS-006**：不交付高可用、多副本 API、分布式任务调度或 Kubernetes。

## 6. 已锁定实现约束

- **DEC-001**：采用“模块化单体 + 多 Host 适配器”；业务 Module 只依赖 Application 提供的接口，Host 拥有进程与窗口生命周期。来源：`USER-DECISION:模块化单体加多 Host 适配器`。
- **DEC-002**：桌面框架使用 Wails 2 稳定版；Wails 3 在本 Spec 建立时仍为 Beta。来源：`RESEARCH:<Url>https://v3.wails.io/status/</Url>`。
- **DEC-003**：桌面业务接口使用同进程 Gin loopback HTTP，不使用 Wails AssetServer 承载业务 Handler，也不以 Wails Binding 复制 CRUD。来源：`RESEARCH:<Url>https://wails.io/docs/reference/options/</Url>` 与 Vite 8 现状。
- **DEC-004**：DesktopHost 仅绑定 `127.0.0.1` 的 OS 分配端口，并要求每次启动随机令牌、严格 Origin 和单实例锁。来源：本地桌面安全约束。
- **DEC-005**：Desktop 数据使用 SQLite；Server Compose 使用 PostgreSQL 和 Redis。两者共享逻辑迁移版本，但允许方言适配实现。来源：已批准总体规划。
- **DEC-006**：前端 deployable App 与 Domain 分层；`apps/admin` 组合多个与后端 Module 对齐的 Domain，Domain 不与 App 一一等同。来源：已批准总体规划。
- **DEC-007**：OpenAPI 是 HTTP wire contract 的唯一生成源；生成代码不可手工编辑，启发式 Go struct 正则检查在兼容期后删除。来源：已批准总体规划。
- **DEC-008**：Windows 使用可离线安装的 x64 NSIS，内含 WebView2 离线运行时策略且 Phase 1 不要求 Authenticode；macOS 使用 ad-hoc code signature 保护 bundle 完整性但不具备 Developer ID/notarization 信任的 ARM64 DMG；Linux 首期发布 AMD64 OCI 镜像。桌面产物只定位为 `unsigned-self-use`，用户验证 SHA-256 后通过单应用系统例外运行，禁止把全局关闭平台安全能力写为安装前提。来源：Phase 1 平台范围、用户 2026-08-25 release 偏差批准与平台官方安全行为。
- **DEC-009**：根仓库是产品发布编排 owner，前后端继续作为独立 submodule；桌面构建把 Admin dist 复制到后端生成资产目录再嵌入，生成资产不手工提交。来源：当前仓库结构与 Go embed 限制。
- **DEC-010**：迁移使用 expand-migrate-contract；旧入口、component 路径和请求层只有在零调用证据及完整回归通过后才能删除。来源：宽重构风险控制。

## 7. 数据、接口与兼容

- **公共接口变化：** 保持现有 `/api/v1` 方法、路径、认证和 envelope；新增 `/health/live`、`/health/ready` 与 `/api/v1/runtime/capabilities`。桌面 bootstrap 不是业务接口，只返回当前进程的运行位置与原生能力。
- **数据模型与持久化：** 服务端参考数据存储为 PostgreSQL；桌面为单进程 SQLite。桌面目录包含 `db/`、`files/`、`logs/`、`backups/` 和 `temp/`，不得写入安装目录。
- **兼容要求：** 现有服务端数据库配置在 Phase 1 不主动删除；现有菜单 `component` 在 routeKey 迁移窗口内继续解析；现有前端页面 URL 和核心导航保持可用。
- **迁移要求：** 发布迁移不可修改；Desktop 自动备份后迁移并 fail closed；Server 由一次性 Migration 执行并阻塞 API ready。SQLite 与 PostgreSQL fixture 都必须验证升级路径。
- **发布或运维影响：** macOS ARM64 与 Windows x64 self-use 构建/安装验证必须分别在原生 runner 完成；Linux 镜像和 Compose 在 Linux runner 完成。Phase 1 不注入平台发行证书；未来启用 Developer ID/notarization 或 Authenticode 时必须以新 release 偏差恢复凭据、签名与平台验证门。

## 8. 非功能要求

- **NFR-001 安全与隐私：** Desktop 不开放局域网端口、不使用宽泛 CORS、不在日志打印 token/密码/签名凭据；容器非 root、非 privileged，API 默认仅在 Compose 网络可达。
- **NFR-002 性能与容量：** 不虚构绝对阈值；必须记录改造前的启动、核心 API 和前端构建基线，改造后若出现可观察回归则在发布前解释并批准。
- **NFR-003 可用性与可靠性：** Host 必须有序关闭；Desktop 升级前可恢复备份；Compose 依赖通过 ready 而非固定 sleep 编排；失败不得静默丢失数据。
- **NFR-004 可观测性与运营：** Server 提供 live/ready、结构化日志和版本信息；Desktop 提供本地日志位置、版本、Host profile 和迁移结果，但不默认上报遥测。
- **NFR-005 可维护性：** Service 不接收 `gin.Context`，API 层不直接 ORM；ApiClient 不依赖 Vue、Pinia 或 Element Plus；Domain 不深层导入其他 Domain；平台判断集中在 Runtime/Host Adapter。
- **NFR-006 可重复发布：** 所有产物绑定同一产品版本和前后端提交，提供 checksum、SBOM、trust-state 验证与构建元数据；安装包不得运行时下载核心资源。unsigned self-use 文档必须先要求校验 SHA-256，再说明单应用授权路径和未签名风险。

## 9. 验证策略

| 接缝 | 层级 | 覆盖合同 | 现有先例或命令 | Evidence 类型 |
|---|---|---|---|---|
| Go Application/Module/Lifecycle | 单元与集成 | AC-001、AC-011、AC-013 | `cd go-admin-plus && go test ./...` | 测试输出、goroutine/关闭断言 |
| `/api/v1` characterization 与 OpenAPI | 公共 HTTP contract | AC-002、AC-005、AC-013、AC-015 | Go HTTP tests、生成 diff、前端 type-check | contract 报告与生成零 diff |
| Admin workspace 与 Domain | 静态、单元、组件 | AC-003、AC-004、AC-005、AC-006、AC-018 | `pnpm lint`、`pnpm type-check`、`pnpm test:unit`、`pnpm build:prod` | 命令和包级测试摘要 |
| Playwright Admin 流程 | Web E2E | AC-002、AC-003、AC-006、AC-012 | `cd go-admin-ui-plus && pnpm e2e`、live project | trace、截图或运行摘要 |
| SQLite 升级 fixture | 数据迁移集成 | AC-010、AC-011、AC-015 | 原生 Go/desktop migration tests | 迁移前后 schema/数据/备份摘要 |
| Docker Compose | 系统 E2E | AC-012、AC-013、AC-014、AC-016 | `docker compose config`、空卷 smoke、浏览器 live E2E | 容器状态、HTTP 与安全检查 |
| macOS ARM64 原生安装 | 发布 E2E | AC-007、AC-009、AC-010、AC-016、AC-017 | 原生 runner 离线安装、ad-hoc 完整性核验、SHA-256 与单应用授权路径 | 安装日志、`codesign`、quarantine/授权和场景结果 |
| Windows AMD64 原生安装 | 发布 E2E | AC-008、AC-009、AC-010、AC-016、AC-017 | 原生 runner 离线安装、WebView2、SHA-256 与 SmartScreen self-use 路径 | 安装日志、trust state 和场景结果 |

## 10. 风险、假设与未决问题

### 风险

- go-admin-core 全局 Runtime 可能使 Application 去全局化成为最大 prefactor；必须先建立 characterization，再逐步封装，禁止一次性替换全部业务层。
- SQLite 驱动依赖 CGO，平台原生工具链是发布前置条件；unsigned self-use 包会触发平台信誉/来源警告，在受管策略或 Windows Smart App Control 下可能无法运行。
- 菜单 component 到 routeKey、手写类型到 OpenAPI 生成类型都属于 expand-contract，过早收缩会破坏现有页面。
- Desktop loopback 增加本机进程攻击面；随机端口本身不是认证，必须同时校验启动令牌和 Origin。
- 当前开发与部署 workflow 混合，发布改造不得意外触发既有线上演示部署。

### 已采用的低影响假设

- 产品显示名、bundle identifier、Windows product/upgrade identity 等由版本化 release manifest 提供并在 self-use 发布后保持稳定；未来首次可信签名发布可以增加 publisher/signing identity，但改变现有 bundle/product/AppData identity 仍需迁移 Spec。
- Desktop 定时任务仅在应用运行期间执行；错过的任务默认不补跑，只有具体 Job 明确声明幂等与补跑策略时才允许追赶。
- 服务端文件存储的 Compose 默认实现使用持久卷；对象存储是同一 FileStore 接口的可选配置，不是 Phase 1 必需服务。
- Linux 入口由 Web/Nginx 暴露 80/443，API 不默认映射宿主端口；开发覆盖文件可以显式暴露调试端口。

以上假设分别通过 release manifest 检查、任务生命周期测试、FileStore contract test 和 Compose config 检查验证；不成立时按 SpecDev 偏差流程返回拥有该决策的工件。

### 未决问题

无。
