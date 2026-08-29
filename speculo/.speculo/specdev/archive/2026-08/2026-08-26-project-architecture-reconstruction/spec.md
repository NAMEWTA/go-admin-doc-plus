---
schema_version: 3
artifact: spec
change: 2026-08-26-project-architecture-reconstruction
status: ready
ready_for_tickets: true
sources:
  - "USER-DECISION:2026-08-26 完全 Greenfield 重构且只要求目标功能完整"
  - "DESIGN-TREE:<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/design-tree.json</Path>"
  - "ADR:<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ADR.md</Path>"
  - "CONTEXT:<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/CONTEXT.md</Path>"
  - "CODE:<Path>go-admin-plus/app/</Path>"
  - "CODE:<Path>go-admin-ui-plus/domains/</Path>"
  - "CODE:<Path>go-admin-plus/internal/host/desktop/runtime_integration_test.go</Path>"
  - "CODE:<Path>go-admin-ui-plus/tests/e2e/mocked/</Path>"
---

# Spec: Go Admin Plus 自主产品架构重构

- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **当前 ADR：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ADR.md</Path>`
- **当前领域上下文：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/CONTEXT.md</Path>`

## 1. 问题与目标

### 问题陈述

Go Admin Plus 已经成为自主维护产品，但仓库仍同时存在上游遗留命名、子项目级 Git/发布资产、混合职责目录、全局运行时状态、旧 Wails 桌面宿主和多套相互漂移的合同。当前后端把认证、系统管理、组织、审计和其他工具混在 `admin/common/other` 等边界中；前端所谓 domain 同时拥有 API 与 Vue 页面；Server 与 Desktop 依赖不同的缓存/队列/数据库组合；JWT、Casbin、Redis、租户、多数据库和旧配置进一步扩大部署与安全表面。

这种状态导致以下用户可观察问题：

- 贡献者无法从仓库根获得唯一、可复现的开发、验证、迁移和发行入口；
- Web 与 Desktop 的能力、会话和运行方式可能分叉，桌面端不能形成清晰的 Tauri 2 产品合同；
- PostgreSQL、Server SQLite 与 Desktop SQLite 没有统一的功能和迁移验收矩阵；
- 权限、会话撤销、异步可靠性和调度所有权分散在 JWT、Casbin、Redis、内存队列及租户 runtime 中；
- 旧 API、旧 schema、旧配置与当前文档互相牵制，无法表达自主产品的当前真相；
- 编译成功不足以证明模块边界、跨端功能、数据库、桌面安装包和发行供应链完整。

本 change 采用 Greenfield 功能合同。当前路由、页面、测试和种子只用于发现目标业务能力，不是旧 API、旧数据或旧操作方式的兼容基线。

### 目标用户与场景

- **系统管理员：** 登录管理端，管理用户、角色、权限、菜单、API 权限、数据范围、部门、岗位、参数和字典。
- **审计与运营人员：** 查询登录/操作审计，管理调度任务并查看执行结果，查看系统健康和运行状态。
- **业务开发者：** 通过数据库元数据配置、预览并生成符合新架构的业务模块，通过 Demo CRUD 理解标准能力闭环。
- **Web 使用者：** 通过浏览器连接 Server PostgreSQL 或 Server SQLite 部署，使用完整管理能力。
- **Desktop 使用者：** 在 macOS 或 Windows 上使用本地 SQLite、离线可运行且与 Web 共享业务组合的桌面应用。
- **贡献者与发行工程师：** 从仓库根使用统一任务完成开发、验证、迁移、打包和受保护发行。
- **平台运维人员：** 运行 Linux amd64/arm64 OCI 与 Compose，检查健康、就绪、指标、日志和 worker 状态。

### 成功标准

- 仓库只保留一个 Git 产品边界，根级治理资产按生命周期归属，后端和前端目录固定为 `<Path>go-admin-plus/</Path>` 与 `<Path>go-admin-plus-ui/</Path>`。
- Web 与 Desktop 独立构建但消费同一共享业务组合；除明确的宿主能力差异外，两端菜单、页面、权限和业务行为一致。
- IAM、Organization、Settings、Audit、Scheduler、Generator、Files、Demo 八个业务模块具备本 Spec 定义的外部能力。
- Server PostgreSQL、Server SQLite、Desktop SQLite 均可从空库迁移并完成适用的核心业务场景；SQLite profile 明确限制单实例。
- Web 使用安全 Cookie 会话，Desktop 通过 Tauri 2 宿主安全存储和 transport proxy 使用会话；撤销、超时、轮换与权限变更可被立即验证。
- Redis、JWT、refresh token、Casbin、租户、Wails、MySQL、SQL Server、旧多数据源和旧兼容结构在目标产品中零保留。
- OpenAPI 3.1 是 Go HTTP transport 与 TypeScript transport 的唯一事实源，生成物和实现不漂移。
- 根 Taskfile 提供产品级开发、构建、测试、检查、生成、迁移、打包和发行入口，Hook 与 CI 复用相同任务。
- Linux 双架构 OCI/Compose、macOS Universal DMG、Windows x64 NSIS 均产生原生安装/启动证据、SBOM 与 provenance。
- 本地、PR 和受保护发行门禁能够阻止架构、功能、数据库、安全、桌面或供应链不符合合同的变更。

### 非目标

- 不迁移、读取或兼容旧生产数据、旧 API、旧 schema、旧配置键、旧命令或旧内部 import。
- 不提供租户能力、公共自助注册、MySQL、SQL Server、Redis、JWT、Casbin、Wails 或 Linux 桌面安装包。
- 不在本 Spec 中承诺云对象存储厂商、外部身份提供商、移动端、自动生产部署或自动对外发布。
- 不把 Java/Maven 的多模块目录机械翻译到 Go，也不创建 `common` 杂物箱或跨模块数据库 join。
- 不规定逐文件施工步骤、私有函数结构、具体错误码枚举或未经测量的性能 SLA。

## 2. 解决方案与外部行为

### 解决方案摘要

产品形成单一 Monorepo：根目录拥有 Git、任务、合同、部署、发行和数据库工程治理；Go 服务端采用模块化单体；前端采用 pnpm workspace；桌面由 Tauri 2 管理 Go sidecar。业务能力由八个后端模块和对应的无头领域/Web Domain 共同表达，Web 与 Desktop App 只组合共享能力并选择运行 adapter。

系统以模块化 OpenAPI 3.1 定义 HTTP 合同，以模块 Migration 定义 PostgreSQL/SQLite schema，以不可变类型化配置选择 `server-postgres`、`server-sqlite` 或 `desktop-sqlite`。会话、任务定义和可靠事件均持久化在当前 profile 的数据库中；缓存只是可禁用的有界本地优化；可靠异步通过 Transactional Outbox 和协调 Worker 完成。

施工在新结构中按目标垂直切片推进，所有目标能力和发行证据完成后执行一次原子切换，同时删除旧结构。开发期间不得形成可发布的新旧双轨。

### 主要流程

#### 2.1 开发者启动与根任务

1. 贡献者在仓库根安装已固定版本的 Go、Node.js、pnpm、Rust/Tauri 及适用平台工具。
2. 根 `dev` 任务暴露明确的运行目标，至少能够选择 Server PostgreSQL、Server SQLite、Admin Web 和 Admin Desktop 开发形态。
3. 根任务负责调用后端、前端、合同、桌面、部署与发行脚本；子项目原生命令可以存在，但不得成为 CI 或文档的产品级唯一入口。
4. 启动输出只显示非敏感的 profile、监听地址、版本和健康入口，不打印数据库凭据、Session 或 secret reference 内容。

#### 2.2 Server 启动

1. 操作者选择 `server-postgres` 或 `server-sqlite`，加载对应最小配置 schema。
2. 启动器按 `defaults < config file < environment < explicit non-secret CLI flags` 合成不可变配置快照，并读取环境变量或 `_FILE` secret reference。
3. 配置和依赖校验通过后，统一 migration runner 从空库或新架构上一版本迁移到当前版本。
4. 应用装配八个业务模块、HTTP adapter、observability 和适用 worker，再进入 ready 状态。
5. PostgreSQL 部署允许 API 多副本；独立 Worker 通过固定 advisory lock 产生唯一 scheduler/outbox executor。Server SQLite 只允许单实例。

#### 2.3 Web 登录与管理

1. 未认证用户只能进入登录流程；提交有效凭据后 IAM 创建高熵不透明 Session，数据库只保存 token hash。
2. 浏览器仅通过 `__Host-*`、`Secure`、`HttpOnly`、`SameSite`、`Path=/` 且无 `Domain` 的 Cookie 携带 Session；状态改变请求必须通过 CSRF 校验。
3. App Shell 根据当前用户、角色、稳定 Permission Code 和能力 manifest 生成菜单、路由与操作权限。
4. 用户在授权范围内使用 IAM、Organization、Settings、Audit、Scheduler、Generator、Files 和 Demo 页面；后端对每个受保护用例独立执行认证、权限与数据范围校验。
5. 用户可查看和更新个人资料、头像与密码并主动退出；管理员可撤销 Session 或更改权限，后续请求立即按新状态判定。

#### 2.4 Desktop 启动与离线使用

1. Tauri 2 宿主解析本机数据/日志目录并取得单实例锁；第二个实例不得同时写入同一 SQLite 数据库。
2. 宿主在需要升级时先建立可恢复备份，再运行 Desktop SQLite migration；备份或迁移失败时不打开主窗口。
3. 宿主生成随机 loopback 端口和一次性启动材料，启动并监督按目标 triple 捆绑的 Go sidecar；sidecar 不监听 LAN 地址。
4. sidecar ready 后才显示窗口。Tauri transport proxy 从 Stronghold 读取 Session 并注入请求，WebView JavaScript 不得获得原始 Session 或启动令牌。
5. Desktop 使用与 Admin Web 相同的业务组合、页面和 Permission Code，并通过 desktop adapter 提供文件选择、下载、剪贴板、通知等宿主能力。
6. 应用重启后保留 SQLite 业务数据、文件元数据和有效会话状态；退出时先停止接收请求、排空受控工作并关闭 sidecar/数据库资源。

#### 2.5 业务能力

- **IAM：** 登录/退出、Session 生命周期、个人资料与密码；用户、角色、菜单、API Permission Code、角色权限和数据范围管理；授权管理员可启停用户/角色和重置密码。
- **Organization：** 部门树、岗位及其组织关系的查询、创建、编辑和删除；IAM 通过消费者 Port 使用必要组织信息。
- **Settings：** 应用参数、界面设置和字典类型/字典数据的查询与维护；运行 Profile secret 不属于业务设置。
- **Audit：** 记录登录和受审计业务操作，允许授权人员筛选、查看和按产品策略清理记录；审计数据不得包含原始 Session、密码或 secret。
- **Scheduler：** 维护已注册任务类型的任务定义，启停调度并查看执行记录；不允许通过配置执行任意未注册代码。
- **Generator：** 读取当前数据库的授权元数据，导入/编辑表与列配置，预览并生成符合新 OpenAPI、Go 模块和 pnpm 前端边界的单表 CRUD 代码；生成结果必须通过目标门禁。
- **Files：** 经授权上传、持久化文件元数据、受控读取/下载并阻止目录穿越；首期保证本地文件存储适配器可用于三个 profile，其他存储厂商不是本 Spec 承诺。
- **Demo：** 提供标准产品记录的查询、创建、编辑和删除闭环，作为双方言、Web/Desktop 与权限行为的参考业务场景。
- **Observability：** 提供职责分离的 liveness、readiness、metrics、runtime capabilities 和 server status；失败响应对依赖与 secret 细节脱敏。

#### 2.6 合同、数据与可靠异步

1. `<Path>contracts/openapi/</Path>` 中的模块化 OpenAPI 3.1 描述认证、分页、校验、错误类别、幂等语义和所有公共操作。
2. 合同生成 Go strict server interface/transport types 与 TypeScript transport types/client；两端通过显式 mapping 转换领域模型。
3. 每个模块拥有自己的 schema、Repository、Persistence Record 和 Migration；模块不得查询其他模块私有表。
4. 业务状态与 Integration Event 在同一事务写入；Worker 认领 Outbox、执行可重试分发，Consumer 根据稳定业务键保持幂等。
5. 进程内缓存可以清空、禁用或随进程丢失，不改变认证、授权、调度或业务正确性。

#### 2.7 打包与发行

1. 根 `package`/`release` 任务从同一产品版本和源码身份生成候选制品。
2. Linux 生成 amd64/arm64 OCI，并通过 Compose 验证迁移、secret reference、健康和持久化。
3. macOS 原生 runner 生成包含目标架构的 Universal DMG；生产制品完成签名、公证、安装、首次启动与重启 smoke。
4. Windows 原生 runner 生成 x64 NSIS；生产制品完成签名、安装、首次启动、重启和卸载/清理边界验证。
5. 每个平台关联 checksum、SBOM、provenance 和源码/版本身份。未通过必需门禁的候选不得标记为可发行。

### 边界、失败与稳定错误行为

- 配置缺失、未知字段、无效组合或 secret 不可读时，进程在监听端口或打开桌面窗口前失败；错误指出字段路径和规则，但不包含敏感值。
- Migration 失败时应用不得进入 ready；Desktop 必须保留原数据库和备份，Server 不得以部分 schema 继续服务。
- 未认证、Session 已撤销/过期或 CSRF 校验失败时，受保护操作不执行；Web 返回登录流程，Desktop 保持凭据对 JavaScript 不可见。
- 已认证但缺少 Permission Code 或数据范围时，前端不展示相应操作，后端仍必须拒绝直接调用且不产生状态变化。
- 公共 API 对 validation、authentication、authorization、not-found、conflict 和 internal failure 使用 OpenAPI 声明的稳定错误类别；不得泄露 SQL、文件系统路径、stack、secret 或原始 Session。
- 重复提交在 UI 层不得产生并发重复命令；需要网络重试的非幂等操作必须由 OpenAPI 明确声明和验证幂等策略。
- 文件路径逃逸、符号链接逃逸、未授权下载或超出适配器约束的上传必须失败且不写入授权根之外。
- Outbox 分发失败必须保留可重试状态；Worker 丢失 PostgreSQL advisory lock 或数据库连接时必须停止 scheduler/outbox 副作用并允许其他实例接管。
- SQLite profile 检测到第二实例时必须拒绝启动，不以多写方式降级。
- sidecar 绑定非 loopback、一次性启动材料缺失、来源不可信或 assets 不完整时，Desktop 不打开主窗口。
- liveness 只表达进程存活；readiness 在应用未 ready 或必要依赖失败时返回 not-ready，响应不得包含依赖凭据。
- 任一 required PR gate 失败时禁止合并；任一 protected release gate 失败时禁止发行。Flaky retry 不得把失败静默变绿。

### 状态转换与不变量

- **应用：** `starting -> ready -> draining -> stopped`；任一启动阶段可进入 `failed`，且 failed 不得对外宣称 ready。
- **Session：** `active -> rotated | revoked | expired`；旧 token 在轮换、撤销或到期后不可再次恢复为 active。
- **Scheduler：** 任务定义只有在有效、启用且由当前 active executor 持有执行权时才能触发；停止/禁用后不得产生新的执行。
- **Outbox：** `pending -> claimed -> delivered`；失败的 claim 回到可重试状态，已 delivered 的事件不得被消费者重复产生业务效果。
- **Desktop 数据：** `locked -> backed-up-if-needed -> migrated -> ready`；任何前置失败均不得越过窗口显示边界。
- **发行候选：** `built -> verified -> publishable`；缺少原生安装证据、签名/公证要求、SBOM 或 provenance 时不得进入 publishable。
- **权限不变量：** 前端可见性不是授权边界；每个受保护后端用例以 IAM Permission Code 和数据范围为最终判定。
- **数据不变量：** 一个进程只装配一个 Database；产品 schema、接口、配置与审计中不存在 tenant 概念。
- **架构不变量：** 生成 transport 不进入领域模型；模块不共享 ORM record/repository；App 和 adapter 不拥有业务规则。

## 3. 用户故事

- **US-001**：作为贡献者，我希望从仓库根使用统一任务启动、构建、测试、生成、迁移和打包，以便本地行为与 CI 一致。
- **US-002**：作为 Server 运维人员，我希望使用 PostgreSQL profile 部署可多副本 API 和唯一协调 Worker，以便获得可扩展且无 Redis 的服务端运行形态。
- **US-003**：作为小型部署运维人员，我希望使用单实例 Server SQLite profile 运行完整管理能力，以便无需外部数据库服务。
- **US-004**：作为 Desktop 使用者，我希望安装 Tauri 2 桌面应用并离线使用本地 SQLite，以便不启动独立服务器也能完成管理工作。
- **US-005**：作为管理端用户，我希望安全登录、退出和维护个人资料/密码，以便控制自己的身份与会话。
- **US-006**：作为 IAM 管理员，我希望管理用户、角色、菜单、API 权限和数据范围，以便最小权限地授权系统能力。
- **US-007**：作为组织管理员，我希望管理部门树和岗位，以便为用户与数据范围提供清晰的组织结构。
- **US-008**：作为系统管理员，我希望管理应用参数和字典数据，以便在不修改代码的情况下维护业务参考配置。
- **US-009**：作为审计人员，我希望查询登录和操作审计且敏感值已脱敏，以便追踪安全和业务活动。
- **US-010**：作为调度运营人员，我希望维护、启停计划任务并查看执行记录，以便受控运行后台工作。
- **US-011**：作为业务开发者，我希望导入数据库表、编辑生成配置、预览并生成新架构单表 CRUD，以便快速扩展业务而不破坏边界。
- **US-012**：作为授权用户，我希望上传并受控访问文件，以便业务数据可以安全关联本地文件资产。
- **US-013**：作为开发者或验收人员，我希望使用 Demo 产品 CRUD，以便验证权限、合同、双方言和 Web/Desktop 业务闭环。
- **US-014**：作为平台运维人员，我希望区分存活、就绪、指标、能力和服务器状态，以便正确监控与编排实例。
- **US-015**：作为 API 与前端开发者，我希望从同一 OpenAPI 3.1 合同获得 Go 和 TypeScript transport，以便避免跨端漂移。
- **US-016**：作为发行工程师，我希望生成并验证 Linux、macOS 和 Windows 原生制品及供应链证据，以便只发布可安装、可追溯的产品。
- **US-017**：作为安全管理员，我希望 Session 可撤销、超时和轮换，权限变更即时生效，以便不依赖长期 JWT 或缓存状态。
- **US-018**：作为架构维护者，我希望自动阻止跨模块导入、循环依赖、deep import 和生成漂移，以便重构后的边界长期有效。
- **US-019**：作为同时使用 Web 与 Desktop 的管理员，我希望两端拥有相同的菜单、页面和业务语义，以便无需学习两套产品。

## 4. 验收合同

| ID | 前置条件 | 动作或事件 | 可观察结果 | 验证接缝 |
|---|---|---|---|---|
| AC-001 | 从目标分支检出仓库 | 检查根目录和子项目治理文件 | `.github`、`.husky`、`scripts`、`deploy`、`release`、`database` 由根拥有；子项目不存在第二套 Git/Hook/产品发布治理 | 根治理静态扫描 |
| AC-002 | 工具链已安装 | 从根调用 `dev/build/test/lint/generate/migrate/package/release` 产品任务 | 每个任务解析到明确目标并调用受管脚本；Hook 与 CI 调用相同根任务，不依赖私有 CI 实现 | 根 Taskfile 合同测试 |
| AC-003 | 空 PostgreSQL 和有效 `server-postgres` 配置 | 执行迁移并启动 Server | schema 从零创建，应用进入 ready，核心业务 API 可用，secret 不出现在输出 | PostgreSQL profile 集成测试 |
| AC-004 | 空 SQLite 文件和有效 `server-sqlite` 配置 | 执行迁移并启动 Server | schema 从零创建，适用业务与 PostgreSQL 语义一致，第二实例被拒绝 | Server SQLite profile 集成测试 |
| AC-005 | 配置存在缺失、未知字段、冲突或不可读 secret | 启动任一 profile | 在监听/窗口前失败；错误标识字段和规则但不回显值、DSN 或 secret path 内容 | 配置 schema/脱敏测试 |
| AC-006 | 已安装 Desktop 且无现有实例 | 首次启动应用 | Tauri 取得单实例锁，启动随机 loopback Go sidecar，迁移 SQLite，ready 后打开窗口；LAN 无法访问 sidecar | Desktop 原生 tracer |
| AC-007 | Desktop 已有上一新架构版本数据 | 升级启动、完成 CRUD、退出并重启 | 升级前备份，迁移成功后数据保留；重启迁移幂等；失败时保留原库/备份且不打开窗口 | Desktop migration/restart tracer |
| AC-008 | Web 用户凭据有效 | 登录并发起受保护请求 | 服务端只持久化 Session hash；浏览器收到符合 `__Host-*`/Secure/HttpOnly/SameSite/Path/Domain 约束的 Cookie；状态改变请求通过 CSRF | Web IAM E2E 与 Cookie 检查 |
| AC-009 | Desktop 用户凭据有效 | 登录并从页面调用受保护能力 | Session 存入 Stronghold，由 transport proxy 注入；WebView JavaScript、localStorage、URL 和日志均不可读取原值 | Desktop Session tracer |
| AC-010 | Session 处于 active | 发生轮换、主动退出、管理员撤销、空闲超时或绝对超时 | 原 token 后续请求失败且不能恢复；当前客户端进入可理解的重新登录状态 | Session lifecycle contract suite |
| AC-011 | 用户拥有部分 Permission Code 和受限数据范围 | 加载菜单并尝试允许/禁止操作及直接 API 调用 | 只展示获权能力；允许操作成功；缺权或越界调用被后端拒绝且无状态变化 | IAM 权限/数据范围矩阵 + Web E2E |
| AC-012 | 已认证普通用户 | 查看/更新资料、头像和密码并退出 | 合法变更持久化；密码按 Argon2id 保存；旧密码/会话行为符合合同；审计不含敏感值 | IAM application/API 集成测试 |
| AC-013 | IAM 管理员拥有相应 Permission Code | 查询、创建、编辑、启停、删除用户/角色并分配菜单/API/data scope | 列表、详情和状态与提交一致；保护对象和引用冲突按合同拒绝；权限变更即时生效 | IAM API + Web Domain E2E |
| AC-014 | Organization 管理员已登录 | 维护部门树、岗位和关系 | 树结构与岗位列表正确更新；非法父子关系、被引用删除或越权操作失败且不破坏数据 | Organization repository/API/UI suite |
| AC-015 | Settings 管理员已登录 | 维护参数、界面设置、字典类型和字典数据 | 查询/创建/编辑/删除和 option 查询一致；业务设置不接受运行 secret | Settings API + Web Domain E2E |
| AC-016 | 系统产生登录和受审计操作 | 审计人员筛选、查看和执行授权清理 | 登录/操作事实可查询；清理遵守权限；记录不含密码、原始 Session、secret 或完整敏感请求体 | Audit integration + redaction suite |
| AC-017 | 已注册任务类型且管理员有调度权限 | 创建/编辑/启用/停止/删除任务并查看执行记录 | 只有有效且启用的定义可执行；停止后不产生新执行；未注册执行类型被拒绝 | Scheduler API/UI + clock-controlled suite |
| AC-018 | PostgreSQL 启动多个 API/Worker，或 SQLite 启动单实例 | 产生任务和 Outbox 事件并模拟 executor 断连 | PostgreSQL 同时只有一个 active executor，断连后可接管；事件可重试且消费幂等；SQLite 不出现第二 executor | Worker leader/outbox recovery suite |
| AC-019 | Generator 可读取当前 profile 的授权数据库元数据 | 导入表、编辑列/生成配置、预览并生成单表 CRUD | 预览与生成稳定；输出符合 OpenAPI、Go 模块、pnpm 包边界并通过生成/编译/测试门禁 | Generator golden/compile E2E |
| AC-020 | 用户拥有文件权限和有效本地存储根 | 上传、读取/下载文件并尝试未授权或路径逃逸 | 元数据与内容一致；授权访问成功；越权、`..`、绝对路径和符号链接逃逸失败且不写出根目录 | Files repository/security suite |
| AC-021 | 任一正式数据库 profile ready | 对 Demo 产品执行查询、创建、编辑、删除并重启 | CRUD 结果一致且持久化；Web 与 Desktop 使用同一业务合同；双方言结果等价 | Demo API/Web/Desktop tracer |
| AC-022 | Admin Web 与 Admin Desktop 构建完成 | 以同一身份遍历能力 manifest、菜单和核心页面 | 两端业务路由、页面、Permission Code 和状态语义一致；差异仅来自声明的宿主 capability | Shared manifest contract + 双 App E2E |
| AC-023 | OpenAPI 模块合同存在 | lint、bundle、生成两端 transport 并运行实现 conformance | 合同有效，生成确定且 clean tree 无 drift；Go handler 实现 strict interface；TS client 类型通过 | OpenAPI generation/conformance suite |
| AC-024 | PostgreSQL 与 SQLite 空库及上一新架构版本 fixture | 组合并执行全部模块 Migration | 顺序确定、重复执行幂等、升级保留新架构数据；双方言 schema/Repository 行为等价；无 AutoMigrate | Goose 双方言 migration matrix |
| AC-025 | API 收到 validation/authentication/authorization/not-found/conflict/internal failure | 调用对应公共操作 | 返回 OpenAPI 声明的稳定错误类别，状态无意外变化，响应不泄露 SQL、stack、路径或 secret | API negative contract suite |
| AC-026 | 应用处于 starting/ready/dependency-failed/draining | 查询 live、ready、metrics、capabilities、server status | 各端点语义区分且 no-store/脱敏；readiness 只在可接收业务时成功 | Observability handler/host tests |
| AC-027 | 同时提供 defaults、配置文件、环境和非敏感 CLI override | 启动三个 profile 并读取有效配置 | precedence 确定；每个 profile 只接受自身字段；业务代码不能运行时读取全局配置；Desktop 路径/启动材料来自 Tauri | Config precedence/profile suite |
| AC-028 | 目标源码和生成物存在 | 运行 Go import、workspace cycle/deep-import、公开 exports 和合同边界检查 | 不存在跨模块 ORM/repository/transport DTO、service locator、前端 src deep import 或 App 业务实现 | 架构边界测试 |
| AC-029 | 原子切换候选完成 | 执行带明确 allowlist 的全仓兼容扫描 | 旧目录/模块名、旧 API/schema/config、Wails、JWT、refresh token、Casbin、Redis、tenant、MySQL、SQL Server 和临时迁移标记零命中 | 零兼容扫描 |
| AC-030 | Linux 发行候选和 secret references 准备完成 | 构建 amd64/arm64 OCI 并运行 Compose smoke | 镜像按架构启动，迁移成功，健康/持久化/重启成立，Compose 不包含 Redis，生成 checksum/SBOM/provenance | Linux 原生/容器发行门禁 |
| AC-031 | macOS 受保护 runner 与生产凭据可用 | 构建、签名、公证、安装并运行 Universal DMG | 制品包含目标架构，签名/公证有效，首次启动和持久化重启 smoke 通过，证据关联源码版本 | macOS 原生发行门禁 |
| AC-032 | Windows x64 受保护 runner 与生产凭据可用 | 构建、签名、安装、运行并验证 NSIS | 安装器/应用身份正确，首次启动和持久化重启通过，卸载边界明确，证据关联源码版本 | Windows 原生发行门禁 |
| AC-033 | 提交、PR 或发行候选触发流水线 | 执行对应风险层门禁 | Local/Hook、PR、Protected Release 各自 required job 失败会阻断对应阶段；豁免含 owner/范围/期限/Ticket 且到期恢复阻断 | CI policy contract test |
| AC-034 | 任意进程内缓存为空、被清空或禁用 | 重复执行认证、权限、查询和业务命令 | 结果正确性与启用缓存时等价，仅允许性能差异；不存在 Redis 连接尝试 | Cache-disabled integration matrix |
| AC-035 | 用户进入任一可维护的管理列表 | 执行搜索/重置、分页、适用排序、创建/编辑校验、单个或批量删除及合同声明的导出 | 查询状态确定；校验失败不发送命令；重复确认不重复写入；破坏性操作先确认；成功后列表与选择状态刷新 | Shared list/form component + Web E2E |
| AC-036 | 用户完成登录或直接访问受保护页面 | App Shell 加载身份、能力 manifest、菜单并处理授权/未知路由 | 进入可用工作台；动态导航只包含获权页面；未认证跳转登录、未授权和未知路由显示稳定状态；Web/Desktop 共享页面状态语义 | App Shell + 双 App navigation E2E |

### 用户故事覆盖索引

| 用户故事 | 验收合同 |
|---|---|
| US-001 | AC-001、AC-002、AC-033 |
| US-002 | AC-003、AC-005、AC-018、AC-024、AC-026、AC-027、AC-030、AC-034 |
| US-003 | AC-004、AC-005、AC-024、AC-026、AC-027、AC-034 |
| US-004 | AC-006、AC-007、AC-009、AC-020、AC-021、AC-022、AC-027、AC-031、AC-032、AC-036 |
| US-005 | AC-008、AC-010、AC-012、AC-025、AC-036 |
| US-006 | AC-011、AC-013、AC-035、AC-036 |
| US-007 | AC-014、AC-035 |
| US-008 | AC-015、AC-035 |
| US-009 | AC-016、AC-035 |
| US-010 | AC-017、AC-018、AC-035 |
| US-011 | AC-019、AC-023、AC-028、AC-035 |
| US-012 | AC-020、AC-035 |
| US-013 | AC-021、AC-024、AC-035 |
| US-014 | AC-026、AC-030 |
| US-015 | AC-023、AC-025 |
| US-016 | AC-030、AC-031、AC-032、AC-033 |
| US-017 | AC-008、AC-009、AC-010、AC-011 |
| US-018 | AC-001、AC-002、AC-023、AC-028、AC-029、AC-033 |
| US-019 | AC-021、AC-022、AC-036 |

## 5. 范围

### IN

- 单 Git 产品 Monorepo 与根级 `.github`、`.husky`、Taskfile、scripts、contracts、deploy、release、database 治理。
- `<Path>go-admin-plus/</Path>` 中的 Go 模块化单体、Server/Worker/迁移等可执行入口和平台适配器。
- `<Path>go-admin-plus-ui/</Path>` 中的 pnpm workspace、Admin Web、Admin Desktop、Tauri 2 宿主、共享业务组合、无头领域、Web Domain、Platform Port、Runtime Adapter、App Shell 和 UI。
- IAM、Organization、Settings、Audit、Scheduler、Generator、Files、Demo 以及非业务 Observability。
- Server PostgreSQL、Server SQLite、Desktop SQLite 三个正式 profile 和从空库开始的新架构 Migration。
- 不透明 Session、Web Session Cookie、Desktop Session Proxy、Argon2id、Permission Code RBAC 与数据范围、CSRF 和安全审计。
- Transactional Outbox、有界进程内缓存、PostgreSQL advisory lock Worker 协调、SQLite 单实例。
- OpenAPI 3.1 contract-first 和两端 transport 生成/漂移门禁。
- Linux amd64/arm64 OCI+Compose、macOS Universal DMG、Windows x64 NSIS 及签名、公证、SBOM、provenance。
- 新结构垂直切片、最终原子切换和旧技术/结构零兼容验收。

### REUSE

- 复用当前已确认的用户可观察能力清单和测试场景，例如 `<Path>go-admin-plus/app/</Path>` 的管理/任务/生成/文件/Demo 路由与 `<Path>go-admin-ui-plus/domains/</Path>` 的页面，但不复用其旧 API 或架构边界作为合同。
- 复用 `<Path>go-admin-plus/api/openapi/contract_test.go</Path>`、`<Path>go-admin-plus/cmd/migrate/</Path>`、`<Path>go-admin-plus/internal/application/</Path>`、`<Path>go-admin-plus/internal/host/desktop/</Path>` 中有价值的测试接缝思路；目标实现必须改写为新合同和 Tauri 2 宿主。
- 复用 `<Path>go-admin-ui-plus/tests/e2e/mocked/</Path>` 中列表、表单、确认、权限、菜单、调度、生成器等行为场景作为验收输入，不保留旧 token envelope、旧路由或旧 selector 作为公共合同。
- 复用现有 Linux/Windows/macOS 发行验证的可证明思路和供应链工具，但产物路径、Wails metadata 和单架构 macOS 实现不得进入目标合同。
- 用户提供的 Plus UI 架构文档和 Java 后端目录仅作为设计参考；产品不对它们形成运行依赖。

### OUT

- **OOS-001**：旧数据迁移、旧 API/schema/config/CLI/内部 import 兼容；本 change 是 Greenfield 原子切换。
- **OOS-002**：租户、默认租户、空 `tenant_id`、租户 host resolver、每租户数据库或未来 SaaS 多租户预留。
- **OOS-003**：Redis、JWT、refresh token、Casbin、Wails 及其 disabled/optional adapter 或兼容配置。
- **OOS-004**：MySQL、SQL Server、跨数据库通用 ORM abstraction 和旧 AutoMigrate。
- **OOS-005**：Linux 桌面安装包、移动端、浏览器扩展和第三个管理端 App。
- **OOS-006**：无需管理员批准的公共自助注册、OAuth/OIDC/SAML 身份提供商和社交登录。
- **OOS-007**：指定 OSS/OBS/Kodo/S3 厂商兼容；本期合同只要求安全的本地存储 adapter 和可替换 Port。
- **OOS-008**：任意脚本/命令执行型 Scheduler；任务只能引用产品注册的安全执行类型。
- **OOS-009**：自动部署生产、自动创建公开 Release 或自动向外部分发；这些动作仍需要独立授权。
- **OOS-010**：旧前端页面逐像素复刻、旧 URL/菜单 key 保持和旧桌面数据自动导入。

## 6. 已锁定实现约束

- **DEC-001**：最终产品不保留旧内部模式或外部合同兼容，以目标能力和验收合同定义完整性。来源：`ADR-001`、`ADR-004`、`ADR-021`。
- **DEC-002**：后端目录为 `<Path>go-admin-plus/</Path>`，前端工作区为 `<Path>go-admin-plus-ui/</Path>`，根资产按生命周期分域且根 Taskfile 是唯一产品命令入口。来源：`ADR-002`、`ADR-005`、`ADR-014`。
- **DEC-003**：后端使用 Go 原生模块化单体；`internal/app` 装配，八个 `internal/modules` 拥有业务，跨模块只使用消费者 Port、Integration Event 和真正共享值语义。来源：`ADR-006`、`ADR-009`、`ADR-012`。
- **DEC-004**：前端使用真实 pnpm workspace；Admin Web 与 Admin Desktop 共享无头领域、Web Domain、App Shell 和 UI，宿主差异仅由 Platform Port/Runtime Adapter 表达。来源：`ADR-003`、`ADR-008`、`ADR-011`。
- **DEC-005**：Desktop 使用 Tauri 2 管理 Go sidecar；随机 loopback、一次性启动材料、进程监督和目标 triple 打包是安全边界。来源：`ADR-007`。
- **DEC-006**：正式数据库 profile 固定为 Server PostgreSQL、Server SQLite、Desktop SQLite；持久化使用 Bun + `database/sql` 私有 record/repository，迁移使用 Goose Provider 组合模块拥有的双方言只前进序列，禁止 AutoMigrate。来源：`ADR-013`、`ADR-016`。
- **DEC-007**：模块化 OpenAPI 3.1 是跨端唯一 API 合同，生成 Go strict server transport 与 TypeScript transport，领域模型显式映射。来源：`ADR-015`。
- **DEC-008**：产品和 SQL 完整删除 tenant，不保留单租户兼容抽象。来源：`ADR-017`。
- **DEC-009**：认证使用数据库不透明 Session；Web 使用安全 Cookie+CSRF，Desktop 使用 Stronghold+transport proxy；密码使用 Argon2id；Permission Code RBAC/数据范围替代 JWT/Casbin。来源：`ADR-018`。
- **DEC-010**：完整删除 Redis；Session/任务/可靠事件归当前数据库，可靠异步使用 Transactional Outbox，本地缓存可禁用，PostgreSQL advisory lock 协调唯一 Worker，SQLite 仅单实例。来源：`ADR-019`。
- **DEC-011**：三个 profile 使用独立最小强类型 schema 和不可变配置快照；secret 只来自环境或 `_FILE`，Desktop 启动材料来自 Tauri；禁止全局配置、原始 dump 和运行时 reload。来源：`ADR-020`。
- **DEC-012**：发行矩阵固定为 Linux amd64/arm64 OCI+Compose、macOS Universal DMG、Windows x64 NSIS，不提供 Linux Desktop；deploy 管运行，release 管制品、签名、公证、SBOM 与 provenance。来源：`ADR-010`。
- **DEC-013**：施工采用新结构垂直切片和最终原子切换，不发布混合结构或维护数据/API 双轨。来源：`ADR-021`。
- **DEC-014**：质量证据按 Local/Hook、PR、Protected Release 分层；失败阻断对应阶段，豁免必须限时且可审计。来源：`ADR-022`。

## 7. 数据、接口与兼容

- **公共接口变化：** 全部 HTTP 接口允许重新设计。新公共合同只存在于 `<Path>contracts/openapi/</Path>`；生成的 Go/TypeScript transport 不可手改。错误类别、分页、认证、幂等和 Permission Code 引用必须在合同中统一。旧 `/api/v1` 路径、响应 envelope 和 URL 权限策略不保证保留。
- **数据模型与持久化：** 从 Greenfield schema 开始。IAM 持久化用户/角色/权限/Session hash，Organization 持久化部门/岗位，Settings 持久化参数/字典，Audit 持久化审计事实，Scheduler 持久化任务/执行记录，Generator 持久化生成配置，Files 持久化文件元数据，Demo 持久化产品示例；各模块拥有自己的表和 Migration。共享可靠事件使用 Outbox。所有 schema 不含 tenant 结构。
- **兼容要求：** 无旧兼容要求。旧 API、旧数据、旧配置、旧命令、旧目录、Wails/JWT/Casbin/Redis/tenant/MySQL/SQL Server 均是删除对象，不提供 alias、view、disabled flag 或转换层。
- **迁移要求：** 不提供旧系统数据迁移。目标 migration 必须支持新架构从空库安装，以及已经发布的新架构版本之间的只前进升级；SQLite Desktop 升级前必须备份并在失败时保持原数据可恢复。
- **发布或运维影响：** Server PostgreSQL 可多副本 API+唯一 Worker；Server SQLite 和 Desktop SQLite 单实例。Compose 只包含产品及 PostgreSQL（选用时）等必要服务，不包含 Redis。生产 Desktop 制品要求目标平台签名/公证；所有制品要求 checksum、SBOM、provenance 和版本身份。

## 8. 非功能要求

- **NFR-001 安全与隐私：** 原始 Session、密码、CSRF secret、数据库凭据、启动令牌和 secret file 内容不得进入日志、错误、审计、URL、Web Storage、生成工件或配置 dump。认证/授权在后端用例层 fail closed；文件系统、sidecar origin/loopback、Cookie 和 Desktop vault 必须通过负向测试。
- **NFR-002 性能与容量：** 核心正确性不得依赖缓存。缓存必须有界、可观测、可清空和可禁用。PostgreSQL profile 通过 API 多副本扩展，SQLite profile 明确单实例。本 Spec 不虚构吞吐/延迟 SLA；Tickets 对新增热路径记录基线、查询计划或可重复测量并防止明显回归。
- **NFR-003 可用性与可靠性：** 应用启动、排空、停止和失败清理必须有界；migration、outbox、scheduler ownership 和 Desktop backup/restart 必须可恢复；Consumer 必须幂等；丢失缓存或 Worker 连接不得破坏业务事实。
- **NFR-004 可观测性与运营：** live、ready、metrics、runtime capabilities、server status、结构化日志和 worker/outbox 指标职责分离；观测输出脱敏并能够定位 profile、版本、迁移和 executor 状态，但不暴露依赖详情。
- **NFR-005 可移植性：** 相同业务合同在 PostgreSQL/SQLite 和 Web/Desktop 上成立；方言和宿主差异封装在 adapter，并通过矩阵测试证明。SQLite driver 必须固定版本、CGo-free 且进入 SBOM。
- **NFR-006 可维护性：** Go import 方向、模块私有持久化、前端公开 exports、cycle/deep-import、OpenAPI drift 和旧技术零命中均为机器门禁；不得用总体 coverage 数字替代关键路径合同测试。
- **NFR-007 供应链与发行：** 依赖和 Actions 固定可审计版本；发行制品具备 checksum、SBOM、provenance、源码 SHA、产品版本和平台身份；生产签名凭据只在受保护原生 runner 使用。
- **NFR-008 交互一致性：** Web 与 Desktop 的共享业务页面、菜单、权限、确认、validation 和错误恢复语义一致；宿主特有能力通过 capability 显式呈现，不以隐藏条件分支复制页面。

## 9. 验证策略

| 接缝 | 层级 | 覆盖合同 | 现有先例或命令 | Evidence 类型 |
|---|---|---|---|---|
| 根治理/Taskfile/兼容扫描 | 静态与命令合同 | AC-001、AC-002、AC-029、AC-033 | `<Path>Taskfile.yml</Path>`；`<Path>release/manifest/scan-compatibility.mjs</Path>` | 命令日志、扫描报告、required job |
| Go 模块与应用生命周期 | architecture/unit/integration | AC-028、AC-033 | `<Path>go-admin-plus/internal/application/architecture_test.go</Path>`；`go test ./...` | 测试结果、import graph |
| OpenAPI lint/generate/conformance | 合同与编译 | AC-023、AC-025 | `<Path>go-admin-plus/api/openapi/contract_test.go</Path>`；`<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>` | bundle、生成 diff、编译/测试结果 |
| PostgreSQL/SQLite Repository 与 Migration | 双方言集成 | AC-003、AC-004、AC-018、AC-021、AC-024、AC-034 | `<Path>go-admin-plus/cmd/migrate/postgres_integration_test.go</Path>`；`<Path>go-admin-plus/cmd/migrate/migration/runner_test.go</Path>` | from-zero/upgrade 日志、schema diff、查询结果 |
| 配置与 secret 边界 | unit/integration/negative | AC-005、AC-027 | `<Path>go-admin-plus/internal/profile/server_test.go</Path>`；`<Path>go-admin-plus/internal/profile/desktop_test.go</Path>` | precedence matrix、redaction output |
| IAM Session/CSRF/Permission/Data Scope | application/API/security | AC-008 至 AC-013、AC-025 | `<Path>go-admin-plus/common/middleware/auth_test.go</Path>`；`<Path>go-admin-ui-plus/tests/unit/utils/auth.spec.js</Path>` | lifecycle matrix、负向响应、hash/cookie 属性 |
| 各业务模块 API/Repository | application/integration | AC-013 至 AC-021 | `<Path>go-admin-plus/app/</Path>` 的现有 module tests | 双方言用例结果、事件/审计记录 |
| Web Domain 与 App Shell | component/Playwright E2E | AC-008、AC-011 至 AC-017、AC-019 至 AC-022、AC-025、AC-035、AC-036 | `<Path>go-admin-ui-plus/tests/e2e/mocked/</Path>`；`pnpm test:ci`；`pnpm e2e` | Playwright trace、截图、请求合同 |
| Desktop Tauri/sidecar | native integration/E2E | AC-006、AC-007、AC-009、AC-018、AC-020 至 AC-022 | `<Path>go-admin-plus/test/desktop/native_hardening_test.go</Path>`；现有 native tracer workflow | 原生进程日志、端口探针、重启/持久化证据 |
| Observability/host lifecycle | handler/process integration | AC-003 至 AC-007、AC-026 | `<Path>go-admin-plus/internal/application/health/handler_test.go</Path>`；`<Path>go-admin-plus/internal/host/server/host_test.go</Path>` | 状态响应、shutdown/cleanup 结果 |
| Generator output gate | golden + generated compile/E2E | AC-019、AC-023、AC-028 | `<Path>go-admin-plus/test/gen_test.go</Path>`；`<Path>go-admin-ui-plus/tests/e2e/mocked/gen.spec.ts</Path>` | golden diff、生成仓编译/测试结果 |
| Files security | repository/security/native | AC-020、AC-025 | `<Path>go-admin-plus/internal/platform/files_test.go</Path>`；`<Path>go-admin-plus/test/desktop/native_hardening_test.go</Path>` | traversal/symlink/authorization 负向结果 |
| Linux OCI/Compose | 原生容器发行 | AC-003、AC-023、AC-024、AC-029、AC-030、AC-033 | `<Path>deploy/compose/</Path>`；`<Path>release/linux/verify-compose.sh</Path>` | 镜像 manifest、Compose smoke、SBOM/provenance |
| macOS Universal DMG | 受保护原生发行 | AC-006、AC-007、AC-009、AC-021、AC-022、AC-029、AC-031、AC-033 | `<Path>release/macos/</Path>`；现有 macOS release workflow | arch/sign/notary/install/tracer 报告 |
| Windows x64 NSIS | 受保护原生发行 | AC-006、AC-007、AC-009、AC-021、AC-022、AC-029、AC-032、AC-033 | `<Path>release/windows/</Path>`；现有 Windows release/tracer workflow | signature/install/restart/uninstall/tracer 报告 |

## 10. 风险、假设与未决问题

### 风险

- **全仓原子切换风险：** 目录、合同、schema、认证、桌面和发行同时改变，遗漏引用可能制造混合产品。通过目标垂直切片、AC-029 零兼容扫描和最终发行矩阵控制。
- **双方言语义漂移：** PostgreSQL/SQLite 的类型、锁、分页和约束差异可能使功能不一致。通过模块 Repository contract suite 与 AC-024 双方言矩阵控制。
- **数据库承担协调压力：** Session、Outbox、Scheduler 都进入数据库，错误索引或轮询策略可能放大负载。通过有界 polling、索引/查询计划证据、幂等和 PostgreSQL leader 测试控制。
- **Tauri sidecar 安全与生命周期：** loopback、origin、启动材料、Stronghold、备份与窗口时序任一错误都可能暴露凭据或数据。通过 AC-006/007/009 原生 tracer 和负向探针控制。
- **生成器放大架构错误：** 错误模板会批量生成不合规模块。生成输出必须在隔离 fixture 中执行完整合同、编译和边界门禁。
- **原生发行成本：** Universal macOS、Windows 签名/安装和双架构 OCI 依赖原生 runner 与受保护凭据。缺少必需证据时保持不可发行，不以交叉编译替代。
- **旧技术零命中误报/漏报：** 普通词汇与真实遗留可能混杂。扫描使用显式 allowlist、路径/依赖/schema 多层规则，并对关键门禁执行受控反向验证。

### 已采用的低影响假设

- 当前路由、页面和 E2E 只用于确认业务能力与交互风险；具体新 URL、payload、表名和组件结构由后续 OpenAPI/模块 Ticket 在本 Spec 约束内决定。验证方式：AC-023 与各模块合同套件。
- 首期 Files 至少提供三个 profile 可用的安全本地存储 adapter；增加云存储实现不改变 Files Port 和公共授权语义。验证方式：AC-020。
- 不设未经业务测量的数值性能 SLA；每个可能影响热路径的 Ticket 必须记录可重复基线和结果。验证方式：NFR-002 对应 Ticket Evidence。
- 开发候选可以在本地使用非生产签名方式，但只有满足 AC-031/032 的生产候选才能标记 publishable。验证方式：原生发行 manifest。
- 管理员创建用户是首期账户供应方式，不开放公共自助注册。验证方式：OpenAPI 和 Web 路由清单不包含未授权注册入口。

### 未决问题

无。
