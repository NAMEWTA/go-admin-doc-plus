# 项目架构重构

**产品 Monorepo**：包含 Go 服务端、pnpm 前端工作区、桌面宿主、部署与发行治理的单一 Git 产品仓库。
_Avoid_: 把每个子目录当成独立仓库

**Go 服务端**：位于 <Path>go-admin-plus/</Path> 的 Go 应用与命令集合。
_Avoid_: backend、旧上游仓库

**前端工作区**：位于 <Path>go-admin-plus-ui/</Path>、由 pnpm workspace 管理的 App 与共享包集合。
_Avoid_: go-admin-ui-plus、单体 UI 目录

**旧模式兼容层**：仅为旧目录、旧内部导入、旧脚本或旧文档描述继续工作而保留的 alias、转发或双结构。
_Avoid_: 将一次性迁移支架和永久兼容合同混为一谈

**Greenfield 功能合同**：以重新确认的目标能力和验收场景定义“功能完整”，旧 API、旧数据、旧 schema、旧配置键和旧操作方式均不作为兼容合同。
_Avoid_: 按旧接口数量或旧表结构推断目标功能

**根级治理资产**：由 Git 根统一拥有且影响整个产品仓库的 CI、Git hook、任务编排、部署、发行和数据库工程资产。
_Avoid_: 在前后端子项目重复 .github、.husky、release workflow、Docker 编排或通用脚本

**部署定义**：位于 <Path>deploy/</Path>，描述某个环境如何运行容器、服务、网络、密钥引用与镜像。
_Avoid_: release、应用构建元数据

**发行定义**：位于 <Path>release/</Path>，描述版本候选、平台打包、签名、公证、SBOM、provenance 与产物收敛。
_Avoid_: deploy、运行时配置、App 自有的 tauri.conf.json

**数据库工程资产**：位于 <Path>database/</Path> 的开发 bootstrap、测试 fixture、参考快照和数据库工具；不承担生产运行时迁移实现。
_Avoid_: 旧数据迁移包、业务领域的可执行 migration

**Go 应用装配**：位于 <Path>go-admin-plus/internal/app/</Path>，负责选择模块、平台适配器、HTTP 宿主和运行 profile，不承载业务规则。
_Avoid_: admin 业务域、common、全局 service locator

**Go 业务模块**：位于 <Path>go-admin-plus/internal/modules/</Path> 的稳定业务能力边界，拥有自己的用例、模型、传输适配和持久化实现。
_Avoid_: 按 CRUD 表名机械分包、other 杂项模块

**跨模块合同**：位于 <Path>go-admin-plus/internal/contracts/</Path> 的最小端口、事件与共享值语义，只为真实跨模块协作存在。
_Avoid_: 共享 ORM model、万能 DTO 或跨模块 repository

**平台适配器**：位于 <Path>go-admin-plus/internal/platform/</Path> 的数据库、缓存、对象存储、日志、时钟和进程等技术实现。
_Avoid_: 业务用例、领域规则

**桌面宿主**：位于前端 desktop App 的 Tauri 2 Rust 宿主与配置，负责窗口、权限、sidecar 生命周期和原生打包。
_Avoid_: Wails、复制业务页面、在 Rust 中重写 Go 业务

**Go sidecar**：由 Tauri 2 按目标 triple 捆绑和监督的 Go 本地服务，仅监听随机 loopback 端口并通过一次性启动令牌保护宿主会话。
_Avoid_: 固定端口、公开网络监听、把令牌放入 URL 或持久化存储

**Admin Web App**：位于 <Path>go-admin-plus-ui/apps/admin-web/</Path> 的浏览器交付入口，负责 Web runtime adapter 与共享管理端能力组合。
_Avoid_: 承载领域实现、桌面条件分支

**Admin Desktop App**：位于 <Path>go-admin-plus-ui/apps/admin-desktop/</Path> 的 Tauri 2 交付入口，复用与 Admin Web 相同的业务组合，只增加桌面宿主与 adapter。
_Avoid_: 复制页面、独立演化业务路由、Wails binding

**共享业务组合**：Web 与 Desktop 共同消费的能力 manifest、路由贡献和页面注册关系；宿主差异通过 platform port 与 capability 表达。
_Avoid_: 在两个 App 手工维护两份路由和菜单

**IAM 模块**：身份认证、会话、用户、角色、菜单权限、API 权限与数据范围的业务能力边界。
_Avoid_: 部门岗位、应用配置、审计日志

**Organization 模块**：部门、岗位及其组织关系的业务能力边界。
_Avoid_: 登录身份、RBAC 策略实现

**Settings 模块**：应用配置与字典/reference data 的业务能力边界。
_Avoid_: 进程环境变量、密钥和基础设施配置

**Audit 模块**：登录审计与业务操作审计的写入、查询和保留策略边界。
_Avoid_: 通用 logger、metrics

**Scheduler 模块**：计划任务定义、调度控制与执行记录的业务能力边界。
_Avoid_: 根任务运行器、CI job

**Generator 模块**：数据库元数据读取、代码预览和目标代码生成的开发者能力边界。
_Avoid_: 任意文件工具、跨模块导入 ORM model

**Files 模块**：上传、文件元数据、访问授权与存储用例的业务能力边界。
_Avoid_: 具体磁盘/S3 实现

**Observability**：应用运维端点和平台实现，包括 health、readiness、metrics 与 server status；不属于业务模块。
_Avoid_: operations 杂项业务域、审计业务数据

**首期发行矩阵**：Linux amd64/arm64 OCI 与 Compose、macOS Universal DMG、Windows x64 NSIS；不含 Linux 桌面安装包。
_Avoid_: 把工具链可构建目标误写成产品承诺

**无头领域包**：位于 <Path>go-admin-plus-ui/packages/domains/*</Path> 的纯 TypeScript 业务状态、值语义、用例和端口，不感知 Vue、Element Plus、router、HTTP 或具体宿主。
_Avoid_: API 请求函数、Vue composable、页面组件

**Web Domain**：位于 <Path>go-admin-plus-ui/packages/web-domains/*</Path> 的 Vue 页面、表现状态、composables 和路由贡献，通过公开 API 使用无头领域。
_Avoid_: 直接 fetch、跨包 src deep import、宿主检测

**Platform Port**：位于 <Path>go-admin-plus-ui/packages/platform/</Path> 的通知、存储、文件选择、下载、剪贴板和运行能力等宿主无关接口。
_Avoid_: 浏览器或 Tauri 实现

**Runtime Adapter**：位于 <Path>go-admin-plus-ui/packages/adapters/{browser,desktop}/</Path>，分别实现 Platform Port 与 transport bootstrap。
_Avoid_: 业务规则、业务页面

**App Shell**：位于 <Path>go-admin-plus-ui/packages/app-shell/</Path> 的共享布局、能力 manifest 聚合、路由装配和应用启动协议。
_Avoid_: 作为第三个交付 App、拥有领域数据访问

**消费者 Port**：由需要跨模块能力的 Go 消费模块定义的最小同步接口，由应用装配层注入提供者 adapter。
_Avoid_: 提供者导出大型 service interface、全局 service locator

**Integration Event**：由已经发生的稳定业务事实构成的不可变跨模块消息，其 schema 位于真正共享的 contracts 边界。
_Avoid_: 命令式事件名、ORM model、可变指针

**正式数据库 Profile**：Server PostgreSQL、Server SQLite 与 Desktop SQLite；三者必须通过目标功能、migration 和 repository 集成测试。
_Avoid_: MySQL、SQL Server、把 Server SQLite 当作仅开发 fallback

**模块 Migration**：由拥有 schema 的业务模块维护，并由统一 runner 按全局确定顺序组合的不可变前进迁移。
_Avoid_: API 启动 AutoMigrate、根 database 中的生产迁移、旧数据转换

**根任务入口**：由根 <Path>Taskfile.yml</Path> 暴露的产品级开发、验证、生成、迁移、打包和发行命令；Hook 与 CI 调用同一入口。
_Avoid_: 把子 package script 或 CI YAML 作为唯一行为实现

**根 Scripts**：位于 <Path>scripts/{backend,frontend,contracts,desktop,deploy,release}/</Path>、由 Taskfile 调用的复杂脚本实现。
_Avoid_: 重复任务别名、子项目产品级 scripts 目录

**API 合同**：位于 <Path>contracts/openapi/</Path> 的模块化 OpenAPI 3.1，是 Go HTTP transport 与 TypeScript transport 的唯一事实源。
_Avoid_: Go model 反射、Swagger 注释、手写双端 DTO

**生成 Transport**：由 API 合同生成且不可手改的 Go strict server interface/DTO 与 TypeScript types/client。
_Avoid_: 领域模型、业务规则、生成文件内补丁

**Persistence Record**：模块 repository adapter 私有的 Bun 数据库映射结构，只表达该模块表与查询结果。
_Avoid_: 作为领域实体、跨模块共享、直接序列化到 HTTP

**租户能力零保留**：产品不存在 tenant 概念、tenant context、host resolver、多数据库选择或 tenant SQL；单个进程只装配一个数据库和一套授权/调度实例。
_Avoid_: fixed/local/default tenant、空 tenant_id、disabled 开关、兼容视图

**不透明 Session**：由 IAM 生成的高熵随机认证凭据；客户端持有原值，数据库只保存不可逆 hash，并记录撤销、空闲超时、绝对超时与轮换状态。
_Avoid_: JWT、refresh token、可解码身份载荷、日志或数据库中的原始 token

**Web Session Cookie**：名为 <Code>__Host-*</Code>、Path=/、无 Domain、Secure、HttpOnly、SameSite 且配套 CSRF 校验的浏览器会话载体。
_Avoid_: JavaScript 可读 cookie、localStorage bearer token、URL token

**Desktop Session Proxy**：Tauri 2 宿主从 Stronghold 读取 session 并在本地 transport proxy 内注入请求，WebView JavaScript 只调用受控 transport，不能读取凭据。
_Avoid_: localStorage、前端状态中的 token、把 Stronghold 密钥回传给 JavaScript

**IAM Permission Code**：由 IAM 拥有、与 HTTP URL 解耦的稳定授权标识，供 RBAC、数据范围、API policy 和前端 capability 共同引用。
_Avoid_: Casbin policy、路由 URL 充当权限事实源、前端自行推导授权

**Transactional Outbox**：业务状态与待发布事件在同一数据库事务写入，再由可恢复 worker 认领、分发和重试的可靠异步模式。
_Avoid_: 内存队列承担可靠交付、先提交业务再写事件、无幂等 consumer

**有界进程内缓存**：容量、TTL 和指标均受控，任意时刻可清空或禁用且不改变系统正确性的本地性能优化。
_Avoid_: Redis、无限 map、缓存作为授权或业务事实源

**协调 Worker**：负责 scheduler 与 outbox 执行的独立进程角色；SQLite profile 只允许单实例，PostgreSQL 多副本通过固定 advisory lock 产生唯一 active executor。
_Avoid_: 每个 API replica 都执行调度、SQLite 多副本、锁外副作用

**Redis 零保留**：源码、依赖、配置、Compose、volume、脚本、测试和文档均不存在 Redis，产品正确性与交付矩阵不依赖外部缓存服务。
_Avoid_: optional Redis、disabled key、兼容 adapter、遗留环境变量

**运行 Profile Schema**：<Code>server-postgres</Code>、<Code>server-sqlite</Code> 与 <Code>desktop-sqlite</Code> 各自拥有最小、强类型且可独立校验的配置结构，只表达该交付形态真实支持的能力。
_Avoid_: 一个包含所有数据库/宿主字段的全量配置、Desktop 继承服务端配置、字符串模式分支

**不可变配置快照**：启动时按确定优先级加载和校验，随后仅通过构造函数传入组件且运行期间不可修改的类型化值。
_Avoid_: Viper 全局读取、global config、setter、运行时 reload、业务代码读取环境变量

**Secret Reference**：通过环境变量或对应 <Code>_FILE</Code> 路径提供的敏感值；加载器读取后只保留内存值，并在验证、日志、错误与诊断输出中统一脱敏。
_Avoid_: committed YAML secret、CLI secret flag、合并后的明文配置、配置 dump

**宿主启动材料**：由 Tauri 2 宿主在每次 Desktop 启动时生成或解析的数据目录、日志目录、随机 loopback 端口与一次性握手材料，只通过受控 sidecar 启动通道传递。
_Avoid_: 用户配置中的固定端口/启动 token、WebView 可读 secret、复用服务端配置文件

**目标垂直切片**：完全位于新架构中、从合同和应用用例贯穿 persistence/transport 到 Web 与 Desktop 消费面的最小可验证能力闭环。
_Avoid_: 只搬目录不闭合行为、调用旧 service/repository、为旧 API 写 adapter

**原子切换**：所有目标能力和交付证据完成后，在一个受控阶段同时启用新根入口、合同和发行路径，并删除全部旧结构与临时施工代码。
_Avoid_: 混合版本发布、逐步保留旧入口、可回退到旧 schema 的兼容开关

**零兼容验收**：对旧目录名、module path、API、schema/table/column、配置键、依赖、Wails、JWT、Casbin、Redis、tenant 及临时迁移标记执行明确 allowlist 的全仓零命中检查。
_Avoid_: 仅保证编译、注释/fixture/生成物遗留、没有删除期限的临时兼容代码

**风险分层门禁**：按最早且可信的反馈位置，把格式与静态风险放在本地/Hook，把架构和功能集成风险放在 PR，把原生制品、凭据和供应链风险放在受保护发行。
_Avoid_: 所有风险推迟到 release、每个 PR 执行签名发行、人工口头验收

**必需证据**：由根 Taskfile 可复现、能定位到具体合同或能力且失败会阻断对应阶段的机器可读测试、报告、制品或 tracer 结果。
_Avoid_: 只有截图、总体 coverage 数字、无法复现的 CI-only 脚本、允许失败的 required job

**限时门禁豁免**：仅在记录 owner、理由、精确范围、风险、到期时间和修复 Ticket 后临时绕过特定门禁的受审计例外，到期自动恢复阻断。
_Avoid_: 永久 allow-failure、无 owner skip、通过重复 retry 隐藏 flaky test

**原生发行证据**：在目标 OS/架构 runner 上完成制品构建、安装、启动、关键 smoke、签名/公证验证并关联 SBOM 与 provenance 的发布合同。
_Avoid_: 只证明交叉编译、未安装制品、用开发服务器代替发行包
