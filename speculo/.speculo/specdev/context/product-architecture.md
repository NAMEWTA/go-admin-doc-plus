# Go Admin Plus 产品架构词汇表

**状态：** Current
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/CONTEXT.md</Path>`，并由最终实现与 G8 Evidence 校正

## Repository 与治理

**产品 Monorepo**：包含 Go 服务端、pnpm 前端工作区、Tauri 2 桌面宿主、部署与发行治理的单一 Git 产品仓库。
_Avoid_：把每个产品子目录当成独立仓库

**Go 服务端**：位于 `<Path>go-admin-plus/</Path>` 的 Go 应用、命令和测试集合。
_Avoid_：backend、旧上游仓库

**前端工作区**：位于 `<Path>go-admin-plus-ui/</Path>`、由 pnpm workspace 管理的 App 与共享 package 集合。
_Avoid_：go-admin-ui-plus、单体 UI 目录

**旧模式兼容层**：仅为旧目录、module/package、内部导入、脚本或描述继续工作而保留的 alias、转发或双结构。
_Avoid_：把一次性迁移支架当成永久产品合同

**Greenfield 功能合同**：以当前目标能力、OpenAPI、schema 和验收场景定义功能完整，旧 API、数据、schema、配置和操作方式不构成兼容合同。
_Avoid_：按旧接口数量或旧表结构推断目标功能

**根级治理资产**：由 Git 根统一拥有且影响整个产品的 CI、Git Hook、Task、部署、发行和数据库工程资产。
_Avoid_：在前后端子项目重复 `.github`、`.husky`、release workflow 或通用编排

**部署定义**：位于 `<Path>deploy/</Path>`，描述环境如何运行服务、容器、网络、配置和 secret reference。
_Avoid_：release、App 构建元数据

**发行定义**：位于 `<Path>release/</Path>`，描述平台候选、身份、打包、签名接缝、SBOM/provenance 和产物收敛。
_Avoid_：deploy、运行时配置、App 自有 `tauri.conf.json`

**数据库工程资产**：位于 `<Path>database/</Path>` 的开发 bootstrap、测试 fixture、参考快照和工具，不承担生产 migration。
_Avoid_：旧数据转换包、业务模块可执行 migration

## Backend

**Go 应用装配**：位于 `<Path>go-admin-plus/internal/app/</Path>`，拥有 product composition 和跨模块 adapter，不承载业务规则或 Host 资源生命周期。
_Avoid_：admin 业务域、common、全局 service locator

**Go 业务模块**：位于 `<Path>go-admin-plus/internal/modules/</Path>` 的稳定业务能力边界，拥有自己的用例、migration、persistence 和 transport。
_Avoid_：按 CRUD 表机械分包、other 杂项模块

**跨模块合同**：位于 `<Path>go-admin-plus/internal/contracts/</Path>` 的最小共享值语义、capability 和事件合同。
_Avoid_：共享 ORM model、万能 DTO、跨模块 repository

**平台适配器**：位于 `<Path>go-admin-plus/internal/platform/</Path>` 的 database、config、coordination、outbox、cache 和 Desktop 技术实现。
_Avoid_：业务用例、领域规则、运维 endpoint

**桌面宿主**：Admin Desktop 的 Tauri 2 Rust host 与配置，负责窗口、capability、sidecar、Stronghold、原生命令和打包。
_Avoid_：Wails、复制业务页面、在 Rust 中重写 Go 业务

**Go sidecar**：由 Tauri 2 按目标 triple 打包和监督的 Go 本地服务，仅监听随机 loopback 并受宿主启动控制保护。
_Avoid_：固定端口、公开监听、URL token、孤儿进程

**IAM 模块**：身份认证、Session、用户、角色、菜单/API 权限与数据范围的业务边界。
_Avoid_：部门岗位、应用设置、审计数据

**Organization 模块**：部门、岗位和组织关系的业务边界。
_Avoid_：登录身份、RBAC 实现

**Settings 模块**：应用参数、字典和 reference data 的业务边界。
_Avoid_：进程环境、secret、基础设施配置

**Audit 模块**：登录审计、业务操作审计及其查询和保留策略。
_Avoid_：通用 logger、metrics

**Scheduler 模块**：任务定义、调度控制、执行注册和记录。
_Avoid_：根 Task runner、CI job

**Generator 模块**：数据库 metadata 导入、配置、预览和目标代码生成的开发者能力。
_Avoid_：任意文件工具、跨模块 ORM 导入

**Files 模块**：上传、文件 metadata、访问授权和存储用例。
_Avoid_：把本地磁盘或云存储实现当成领域合同

**Observability**：由 `<Path>go-admin-plus/internal/application/health/</Path>` 提供并由 `<Path>go-admin-plus/internal/host/</Path>` 接入的 health、readiness、metrics 和 status 运维能力。
_Avoid_：operations 业务模块、已删除的 platform observability package、Audit 业务数据

**消费者 Port**：由需要同步跨模块能力的消费方定义的最小 Go interface，由 app-owned adapter 实现并在 product composition 注入。
_Avoid_：提供方大型 service interface、模块内 concrete adapter、service locator

**Integration Event**：已发生的不可变业务事实，通过稳定 schema 和 Transactional Outbox 跨模块分发。
_Avoid_：命令式事件、ORM model、可变指针

**正式数据库 Profile**：`server-postgres`、`server-sqlite`、`desktop-sqlite` 三种受支持运行形态。
_Avoid_：MySQL、SQL Server、把 Server SQLite 当成开发 fallback

**模块 Migration**：由 schema owner 模块维护、由统一 Goose runner 按确定顺序组合的不可变前进迁移。
_Avoid_：API AutoMigrate、根 database 生产迁移、旧数据转换

**Persistence Record**：模块 repository adapter 私有的 Bun 数据库映射结构。
_Avoid_：领域实体、跨模块共享、直接 HTTP 序列化

## Frontend 与 App

**Admin Web App**：位于 `<Path>go-admin-plus-ui/apps/admin-web/</Path>` 的浏览器交付入口，选择 Browser adapter 并装配共享产品。
_Avoid_：承载领域实现、Desktop 条件分支

**Admin Desktop App**：位于 `<Path>go-admin-plus-ui/apps/admin-desktop/</Path>` 的 Tauri 2 交付入口，选择 Desktop adapter 并复用同一产品组合。
_Avoid_：复制页面、独立演化业务路由、Wails binding

**共享业务组合**：Web/Desktop 共用的 ProductWorkspace、能力 manifest、路由、页面和注册关系。
_Avoid_：两个 App 手工维护两份业务清单

**无头领域包**：位于 `<Path>go-admin-plus-ui/packages/domains/</Path>` 的纯 TypeScript 状态、值语义、用例和端口。
_Avoid_：Vue、HTTP、router、宿主检测

**Web Domain**：位于 `<Path>go-admin-plus-ui/packages/web-domains/</Path>` 的 Vue 页面、表现状态、composable 和路由贡献。
_Avoid_：直接 fetch、跨 package deep import、宿主分支

**Platform Port**：位于 `<Path>go-admin-plus-ui/packages/platform/</Path>` 的文件选择/保存、通知、剪贴板和运行能力等宿主无关接口。
_Avoid_：浏览器或 Tauri 具体实现

**Runtime Adapter**：位于 `<Path>go-admin-plus-ui/packages/adapters/</Path>` 的 Browser/Desktop transport、Session 和 Platform Port 实现。
_Avoid_：业务规则、业务页面

**App Shell**：位于 `<Path>go-admin-plus-ui/packages/app-shell/</Path>` 的共享布局、身份/能力加载、路由和产品启动协议。
_Avoid_：第三个交付 App、领域数据访问

## Contract 与命令

**根任务入口**：由根 `<Path>Taskfile.yml</Path>` 暴露的开发、验证、生成、迁移、打包和发行命令。
_Avoid_：把 package script 或 CI YAML 当成唯一产品命令实现

**根 Scripts**：位于 `<Path>scripts/go-admin-plus/</Path>`、`<Path>scripts/go-admin-plus-ui/</Path>`、`<Path>scripts/contracts/</Path>`、`<Path>scripts/quality/</Path>`，由 Taskfile 调用的实现。
_Avoid_：重复任务别名、子项目产品级 scripts 根

**API 合同**：位于 `<Path>contracts/openapi/</Path>` 的模块化 OpenAPI 3.1，是 Go/TypeScript transport 的唯一事实源。
_Avoid_：Go model 反射、Swagger 注释、手写双端 DTO

**生成 Transport**：由 API 合同确定性生成且不可手改的 Go strict interface/DTO 与 TypeScript type/client。
_Avoid_：领域模型、业务规则、生成文件补丁

## Security、配置与可靠运行时

**租户能力零保留**：产品不存在 tenant 概念、context、resolver、多数据库选择或 tenant SQL；单个进程只装配一套数据库与授权/Worker。
_Avoid_：fixed/local/default tenant、空 `tenant_id`、disabled flag、兼容视图

**不透明 Session**：IAM 生成的高熵随机凭据；客户端持有原值，数据库只保存 hash 和撤销/超时/轮换状态。
_Avoid_：JWT、refresh token、可解码身份载荷、原始 token 落盘

**Web Session Cookie**：`__Host-*`、Path=/、无 Domain、Secure、HttpOnly、SameSite 且配套 CSRF 的浏览器会话载体。
_Avoid_：JavaScript 可读 cookie、localStorage bearer、URL token

**Desktop Session Proxy**：Tauri host 从 Stronghold 读取 Session 并在受控本地 transport 中注入，WebView 只能调用 bounded fetch bridge。
_Avoid_：localStorage、前端 token 状态、Stronghold secret 回传 JavaScript

**IAM Permission Code**：与 URL 解耦的稳定 dot-separated `module.resource.action` 授权标识，供 IAM policy、API 和前端 capability 共用。
_Avoid_：Casbin policy、URL 充当权限源、默认放大 data scope

**Transactional Outbox**：业务状态与待发布事件在同一数据库事务写入，再由可恢复 Worker claim、dispatch 和 retry。
_Avoid_：内存队列承担可靠交付、业务提交后再写事件、非幂等 consumer

**有界进程内缓存**：容量、TTL 和指标受控，可随时清空或禁用且不改变正确性的本地性能优化。
_Avoid_：Redis、无限 map、缓存作为授权或业务事实源

**协调 Worker**：负责 Scheduler 与 Outbox 的独立进程角色；SQLite 单实例，PostgreSQL 通过固定 advisory lock 产生唯一 active executor。
_Avoid_：每个 API replica 都执行任务、SQLite 多副本、失锁后继续副作用

**Redis 零保留**：源码、依赖、配置、Compose、volume、脚本、测试和文档都不依赖 Redis。
_Avoid_：optional Redis、disabled key、兼容 adapter、遗留环境变量

**运行 Profile Schema**：三个正式 profile 各自拥有最小强类型配置，只表达该运行形态真实支持的字段。
_Avoid_：包含所有数据库/宿主字段的全量 config、字符串模式分支

**不可变配置快照**：启动时按确定 precedence 加载和校验，再通过 product/host 构造函数注入且运行期间不可修改的类型化值。
_Avoid_：Viper global、setter、reload、业务代码读取环境变量

**Secret Reference**：通过环境变量或 `_FILE` 路径提供、加载后只驻留内存并统一脱敏的敏感值。
_Avoid_：提交 YAML secret、CLI secret flag、明文 merged config、config dump

**宿主启动材料**：Tauri host 每次启动提供的数据/日志目录、随机 loopback 和一次性控制材料。
_Avoid_：用户配置中的固定端口/token、WebView 可读 secret、复用 Server config

## 交付与质量

**首期发行矩阵**：仓库保留 Linux amd64/arm64 OCI/Compose、macOS Universal App/DMG、Windows x64 NSIS 的实现；个人自用不要求生产签名、公证或受保护安装。
_Avoid_：把工具链全部目标当成承诺、把个人自用豁免当成已通过签名

**目标垂直切片**：完全位于当前架构中，从合同、用例、persistence/transport 到 Web/Desktop 的最小可验证能力。
_Avoid_：只搬目录、调用旧 service/repository、为旧 API 写 adapter

**原子切换**：目标能力和证据完成后一次启用新入口并删除旧结构的重构方法；当前重构已执行完成。
_Avoid_：长期混合版本、旧 schema fallback、无删除门禁的施工代码

**零兼容验收**：对旧路径、module/package、API/schema/config、依赖和已删除技术执行明确 allowlist 的全仓零命中门禁。
_Avoid_：只保证编译、忽略注释/fixture/生成物、永久临时代码

**风险分层门禁**：把静态风险放在 Local/Hook，架构和功能集成放在 PR/候选，公开发行的原生/凭据风险放在 Protected Release。
_Avoid_：所有风险推迟到 release、个人自用也强制生产签名、未执行保护动作记为 pass

**必需证据**：由根任务可复现、关联具体合同且失败会阻断对应阶段的测试、报告、制品或 tracer。
_Avoid_：只有截图、总体 coverage、不可复现 CI-only 脚本、允许失败的 required job

**限时门禁豁免**：有 owner、理由、精确范围、风险和恢复条件的临时例外；产品范围调整应记录为明确决策而非伪装成临时豁免。
_Avoid_：永久 allow-failure、无 owner skip、重复 retry 隐藏 flaky test

**原生发行证据**：公开分发时在目标 OS/架构完成构建、安装、启动、smoke、签名/公证验证并关联供应链信息的合同。
_Avoid_：只证明交叉编译、未安装制品、用开发服务器代替发行包、个人自用结果冒充公开发行证据
