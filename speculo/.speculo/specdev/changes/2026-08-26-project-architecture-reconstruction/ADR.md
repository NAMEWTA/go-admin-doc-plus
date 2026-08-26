# 项目架构重构决策

## ADR-001: 新架构不保留旧内部模式兼容

**Status:** accepted
**Source:** LOG-001
**Supersedes:** none

### Context
项目已脱离上游并进入自主研发。若继续保留旧目录、旧内部导入、旧脚本和旧描述的 alias，新架构仍会被历史结构约束。

### Decision
目标系统不保留旧内部模式兼容层。完成切换后，旧目录名、旧内部导入、旧脚本入口和旧文档描述必须消失。

### Trade-off
该决定降低迁移后的长期复杂度，但要求一次性更新全部仓库内消费者，也使部分历史分支或外部自动化无法直接复用。

### Consequences
迁移可以使用有明确删除门禁的施工支架，但支架不能成为产品公共合同。HTTP、数据库和配置的外部边界已由 ADR-004 收紧为 Greenfield 合同。

### Verification / Migration
通过全仓引用扫描、构建、测试和发布门禁证明旧路径与入口已清零。

## ADR-002: 固定产品子目录名称

**Status:** accepted
**Source:** LOG-002
**Supersedes:** none

### Context
当前前端目录名仍携带上游命名顺序，并且根脚本、CI、Docker、桌面宿主和文档均引用物理路径。

### Decision
Go 服务端使用 <Path>go-admin-plus/</Path>，前端工作区使用 <Path>go-admin-plus-ui/</Path>。

### Trade-off
产品命名更一致，但目录迁移会触发全仓高冲突修改，不能拆成互不协调的局部改名。

### Consequences
所有仓库内引用必须在同一迁移 Gate 下更新，不提供旧路径 symlink 或 alias。

### Verification / Migration
运行全仓旧名称扫描、Go/前端构建、桌面构建和发布合同。

## ADR-003: 前端采用真实 pnpm Workspace

**Status:** accepted
**Source:** LOG-003
**Supersedes:** none

### Context
当前前端已有 apps、domains、packages 和 workspace 文件，但 domain 同时承载 transport 与 Vue 页面，子 package 也未声明真实依赖，物理拆分尚未形成可信架构边界。

### Decision
前端以 pnpm workspace 作为编译期模块化单体。每个激活 package 声明真实依赖，通过 workspace 协议使用内部包，通过公开 exports 消费，并以自动化检查拒绝循环、deep import 和禁止依赖。

### Trade-off
比单体 alias 增加 manifest 和装配成本，但换取可验证边界、独立 App 构建和共享依赖治理。

### Consequences
用户提供的 Plus UI 多应用领域化方案作为原则输入；本项目的具体 App、领域和桌面组合仍由当前设计树决定。

### Verification / Migration
以 workspace list、manifest 审查、循环依赖检查、每 App typecheck/test/build 和非法依赖 fixture 验证。

## ADR-004: 采用 Greenfield 外部合同

**Status:** accepted
**Source:** LOG-004
**Supersedes:** ADR-001 中尚未确定的外部兼容边界

### Context
本次重构目标是直接形成自主产品的现代架构。继续迁移旧 API、旧数据、旧 schema 或旧配置会把历史实现重新提升为设计约束。

### Decision
目标系统采用 Greenfield 合同。旧 HTTP API、数据库数据、schema、配置键和操作模式可以废弃；功能完整由新能力清单、用户流程和验收场景定义。

### Trade-off
获得最干净的领域和合同设计，但现有部署不能原地升级，旧数据库不能直接继续使用，外部消费者必须改用新合同。

### Consequences
数据库从新基线初始化，OpenAPI 重新设计，Web、桌面和部署只支持新合同。不设计旧数据转换器、API 兼容层或运行时双轨。

### Verification / Migration
以新环境从零安装、目标能力矩阵、Web/桌面端到端场景和发布候选验证；旧路径、旧 API 与旧 schema 扫描必须为零。

## ADR-005: 根级资产按生命周期分域

**Status:** accepted
**Source:** LOG-005
**Supersedes:** none

### Context
当前 Git 治理、Dockerfile、Compose、SQL、脚本和发行逻辑散落在根与两个子项目。全部塞入 release 又会混淆开发、部署、数据库演进和发行的生命周期。

### Decision
仓库根统一拥有 <Path>.github/</Path>、<Path>.husky/</Path>、<Path>scripts/</Path>、<Path>deploy/</Path>、<Path>release/</Path> 与 <Path>database/</Path>。子项目不保留重复治理和通用部署资产。

### Trade-off
所有权与入口清晰，但跨语言任务需要根编排，且必须逐项判断“根资产”与“App 自有构建元数据”的边界。

### Consequences
Tauri 的 <Path>src-tauri/</Path>、package manifest、Go migration 等随所属 App 或模块；根 deploy/release/scripts 只负责编排与工程生命周期。

### Verification / Migration
通过重复资产扫描、根任务 smoke test、GitHub workflow 路径检查、Husky hook 测试和部署/发行合同验证。

## ADR-006: 后端采用 Go 原生模块化单体

**Status:** accepted
**Source:** LOG-006
**Supersedes:** 当前 app/common 结构

### Context
Java 参考项目的宿主、业务模块、跨模块 API 和基础设施职责值得吸收，但 Maven 模块语法不适合逐字复制到 Go。当前 other 与 common 已形成无边界聚合和反向依赖。

### Decision
后端使用 <Path>cmd/</Path>、<Path>internal/app/</Path>、<Path>internal/modules/</Path>、<Path>internal/contracts/</Path>、<Path>internal/platform/</Path>。cmd 仅是入口，app 仅装配，modules 按业务能力内聚，contracts 只含真实跨域合同，platform 实现技术端口。

### Trade-off
依赖方向可由 Go 工具验证，但需要重新识别业务边界，不能通过批量移动旧目录完成。

### Consequences
删除 common 与 other 概念；包名保持简短单词；按复杂度需要建立内部层次，不强制每个模块拥有相同模板目录。

### Verification / Migration
增加 import graph 架构测试、禁止依赖 fixture、模块级测试与完整应用装配测试。

## ADR-007: Tauri 2 管理 Go sidecar

**Status:** accepted
**Source:** LOG-007
**Supersedes:** 当前 Wails 2 桌面宿主

### Context
当前桌面使用 Wails 2 binding。用户要求在本次无兼容重构中采用 Tauri 2；Tauri 官方支持按目标 triple 捆绑外部二进制并以 capabilities 限制宿主权限。

### Decision
桌面端仅使用 Tauri 2。Go 应用编译为 sidecar，由 Tauri 管理启动、随机 loopback 地址、一次性启动令牌、健康检查、异常退出和清理；删除 Wails 代码与 binding。

### Trade-off
获得显式权限和现代 Tauri 发行链，但引入 Rust 工具链、每目标 sidecar 构建以及跨进程监督复杂度。

### Consequences
桌面 App 自有 <Path>src-tauri/</Path> 配置与 Rust 宿主；业务能力仍在共享前端包和 Go 模块中，不在 Rust 重写。每个桌面目标必须拥有匹配的 Go sidecar。

### Verification / Migration
验证 capability 最小权限、随机端口与令牌边界、启动/退出/崩溃清理、第二次启动、离线安装和各平台原生包。

## ADR-008: Web 与桌面采用双 App 单一业务组合

**Status:** accepted
**Source:** LOG-008
**Supersedes:** 当前 apps/admin 的隐式 Web/Wails 双模式

### Context
浏览器与 Tauri 桌面在启动、权限、运行时 adapter 和打包上拥有不同合同，但其管理端功能、领域状态、页面与路由属于同一产品。

### Decision
前端建立 <Path>apps/admin-web/</Path> 与 <Path>apps/admin-desktop/</Path> 两个独立可构建 App。二者消费同一业务组合清单、领域包、页面包和 UI；desktop 只拥有 Tauri 宿主、sidecar bootstrap 与桌面 adapter。

### Trade-off
需要维护两个 App 入口和构建门禁，但避免大量宿主条件分支，也不会复制业务页面。

### Consequences
每个 App 拥有自己的 Vite/runtime 配置；共享路由与能力通过编译期 manifest 组合；宿主特性通过 capability 和 platform port 暴露。

### Verification / Migration
分别运行两个 App 的 dev/typecheck/test/build，并比较业务 manifest、路由能力与关键 E2E 场景；禁止 desktop deep import admin-web 源码。

## ADR-009: 后端采用八个业务能力模块

**Status:** accepted
**Source:** LOG-009
**Supersedes:** 当前 admin、other、jobs、demo 职责地图

### Context
当前 admin 混合多个变化原因，other 是文件、生成器与运维的杂项聚合并反向依赖 admin。仅重命名无法修复边界。

### Decision
<Path>internal/modules/</Path> 下建立 iam、organization、settings、audit、scheduler、generator、files、demo。health、metrics 与 server status 由应用运维层和 <Path>internal/platform/observability/</Path> 提供，不建立 operations 业务模块。

### Trade-off
模块数量增加，部分现有 CRUD 会拆分迁移；换取独立变化、清晰数据所有权和可验证依赖方向。

### Consequences
每个模块拥有其用例与持久化实现；跨模块协作必须通过后续合同决定；common、other、笼统 system/admin 均不是目标边界。

### Verification / Migration
以能力清单映射现有用户流程，增加 import 架构测试、模块装配测试和禁止 other/common/system 聚合的路径扫描。

## ADR-010: 固定首期部署与发行矩阵

**Status:** accepted
**Source:** LOG-010
**Supersedes:** 当前 Linux AMD64、macOS ARM64 Wails、Windows AMD64 Wails 矩阵

### Context
服务端需要覆盖主流 Linux 宿主，桌面端需要替换为 Tauri 2。工具链支持的全部目标不应自动成为产品承诺。

### Decision
首期发布 Linux amd64/arm64 OCI manifest 与 Compose、macOS Universal DMG、Windows x64 NSIS，不发布 Linux 桌面包。deploy 管运行定义；release 管版本候选、签名、公证、SBOM 与 provenance；Tauri 元数据随 admin-desktop App。

### Trade-off
双架构 OCI 与 Universal sidecar 提高构建复杂度，但扩大服务端和 macOS 覆盖；暂缓 Linux 桌面以控制原生发行矩阵。

### Consequences
CI 必须原生或可靠交叉构建每个 Go sidecar/OCI 目标，macOS sidecar 合并为 Universal，所有平台产物关联同一版本与源码 provenance。

### Verification / Migration
验证 OCI manifest 两架构启动、macOS Universal 架构与签名/公证、Windows x64 离线安装，以及 SBOM/provenance/校验和的产品级收敛。

## ADR-011: 前端分离无头领域与 Vue 表现层

**Status:** accepted
**Source:** LOG-011
**Supersedes:** 当前 domains 内 API/page 混合结构

### Context
现有 workspace 的 domain 同时包含 transport 与 Vue 页面，且跨包依赖通过相对路径进入 src，物理目录没有形成可执行边界。

### Decision
使用 <Path>packages/domains/*</Path>、<Path>packages/web-domains/*</Path>、<Path>packages/platform/</Path>、<Path>packages/adapters/{browser,desktop}/</Path>、<Path>packages/app-shell/</Path> 与 <Path>packages/ui/</Path>。Apps 只选择 shell、adapter 与能力 manifest。

### Trade-off
包数量和映射代码增加，但领域可脱离 Vue/HTTP 测试，Web/桌面宿主可以复用同一业务能力且保持独立。

### Consequences
内部 package 必须声明真实 workspace 依赖并只公开 exports；禁止循环、deep import 和 App 反向依赖。transport DTO 必须映射为领域模型。

### Verification / Migration
对 package graph、workspace manifest、exports、循环和禁止依赖建立自动检查，并分别运行无头领域测试、Web Domain 组件测试和双 App 构建。

## ADR-012: 跨模块使用消费者 Port 与 Integration Event

**Status:** accepted
**Source:** LOG-012
**Supersedes:** other/admin/common/global 隐式协作

### Context
当前跨模块调用会导入其他模块 DTO/model/service，数据库与 Casbin 通过全局 runtime 取得，依赖方向和运行关系不可见。

### Decision
同步依赖使用消费者定义的最小 port，由 <Path>internal/app/</Path> 显式注入。异步协作使用不可变 integration event。<Path>internal/contracts/</Path> 只容纳被多个模块真实共享的 ID、值语义和事件。

### Trade-off
需要 adapter、映射和显式装配，但模块可以独立变化和测试，contracts 不会成为新的 common。

### Consequences
禁止跨模块 ORM model、repository、transport DTO、数据库 join 和 service locator。跨模块工作流在应用用例中编排，模块数据由其 owner 修改。

### Verification / Migration
通过 Go import graph、禁止依赖 fixture、模块 contract test、事件 schema 测试和无全局数据库扫描验证。

## ADR-013: Server 支持 PostgreSQL/SQLite，Desktop 仅 SQLite

**Status:** accepted
**Source:** LOG-013
**Supersedes:** 旧 MySQL/PostgreSQL/SQL Server/SQLite 通用驱动合同

### Context
用户要求服务端同时覆盖完整 PostgreSQL 部署和轻量 SQLite 部署，桌面保持自包含 SQLite。旧多驱动代码并未提供等价验证。

### Decision
正式 profile 为 Server PostgreSQL、Server SQLite、Desktop SQLite。删除 MySQL 与 SQL Server。各模块拥有 Greenfield schema migration，由统一 runner 组合为全局确定、不可变、只前进序列；根 database 只保存开发 bootstrap、测试 fixture 和工程工具。

### Trade-off
比单一 PostgreSQL 增加方言和并发语义测试，但仍显著小于四数据库兼容矩阵，并支持轻量自托管和桌面场景。

### Consequences
Server SQLite 必须覆盖完整产品功能，不是 dev fallback；所有 schema/query 必须在 PostgreSQL 与 SQLite 上验证。Server 迁移由独立命令执行，Desktop 在提供服务前备份并迁移；API 启动禁止 AutoMigrate。

### Verification / Migration
对三个 profile 执行从空库迁移、幂等状态检查、模块 repository 集成测试、并发/锁测试、备份恢复和目标 E2E。

## ADR-014: 根 Taskfile 是产品命令唯一入口

**Status:** accepted
**Source:** LOG-014
**Supersedes:** 子项目 Hook 与散落的产品级脚本入口

### Context
当前已有根 Taskfile，但子项目 scripts、前端 package scripts、前端内 Husky 和 CI/发布脚本仍可能分别定义行为。

### Decision
根 Taskfile 暴露 dev/build/test/lint/generate/migrate/package/release。复杂实现位于 <Path>scripts/{backend,frontend,contracts,desktop,deploy,release}/</Path>。唯一 <Path>.husky/</Path> 位于 Git 根并运行快速 staged 检查；CI 调用相同任务执行完整门禁。

### Trade-off
根编排需要理解 Go、pnpm、Rust 与平台工具链，但用户和 CI 获得稳定统一入口。

### Consequences
包内 scripts 仅服务包工具链；文档、Hook 和 workflow 不复制命令逻辑。发行任务默认只产出本地候选，外部写入仍需授权。

### Verification / Migration
运行 Taskfile 列表与 smoke test，对本地/Hook/CI 命令做等价性检查，并扫描子项目 Husky 与重复产品入口。

## ADR-015: OpenAPI 3.1 是跨端合同唯一事实源

**Status:** accepted
**Source:** LOG-015
**Supersedes:** 当前手写 OpenAPI、Gin route inventory 与自定义字段扫描

### Context
当前 OpenAPI 不能生成并约束 Go server interface，前端脚本通过扫描旧 Go struct 猜测字段，合同漂移只能在部分测试中发现。

### Decision
<Path>contracts/openapi/</Path> 保存模块化 OpenAPI 3.1。使用固定版本生成器生成 Go strict server interface/transport types 和 TypeScript transport types/client；Go HTTP 使用 Chi adapter。两端领域模型显式映射，CI lint、bundle、generate 并拒绝 drift。

### Trade-off
修改 API 必须先修改合同并处理生成差异，但服务端、Web 与桌面在编译期共享同一协议。

### Consequences
生成代码不可手改；transport 不定义领域边界；operationId、分页、错误、认证、幂等和日期/ID 编码必须在合同中统一。

### Verification / Migration
执行 OpenAPI lint/bundle、确定性生成、dirty diff、Go strict interface 编译、TS typecheck 和运行时 request/response conformance。

## ADR-016: 持久化采用 Bun 与 Goose

**Status:** accepted
**Source:** LOG-016
**Supersedes:** GORM model、AutoMigrate 与全局 database runtime

### Context
目标需要同等支持 PostgreSQL 与 SQLite，并让领域模型和模块边界不受 ORM/global runtime 污染。

### Decision
使用 SQL-first Bun + database/sql 实现模块私有 repository。PostgreSQL 使用 Bun pgdriver，SQLite 使用固定且验证的 CGo-free driver。领域模型不含数据库 tag。Goose Provider 从 embed.FS 组合双方言的模块 migration，禁止 AutoMigrate。

### Trade-off
需要显式 persistence record 和 mapping，也要维护少量方言差异；换取可见 SQL、跨平台构建和可验证 schema 演进。

### Consequences
模块不能访问其他模块 repository/table；双方言差异封装在 persistence/migration adapter；driver 与传递依赖固定并进入 SBOM。

### Verification / Migration
执行 PostgreSQL/SQLite repository contract suite、migration from-zero/upgrade、查询计划与锁语义测试，以及 CGo-free 多目标 sidecar 构建。

## ADR-017: 租户功能与数据模型全部删除

**Status:** accepted
**Source:** LOG-017
**Supersedes:** tenant resolver、多数据库 runtime 与 tenant-aware SQL

### Context
当前即使是 Desktop 也使用 fixed tenant，租户概念传播到数据库、Casbin、cron、middleware、审计和 migration。用户明确取消该产品能力。

### Decision
不保留任何租户抽象。删除 tenant package、host/domain resolver、多数据库注册、tenant context、tenant-aware DB/Casbin/cron API 和租户配置；SQL/schema/migration/index/fixture 删除租户表、字段、约束和种子；前端、测试、文档同步删除。

### Trade-off
未来若要提供 SaaS 多租户必须作为新架构 change 重新设计，但当前所有数据访问、授权和运行 profile 显著简化。

### Consequences
进程只装配一个 Database、IAM policy/session store 与 Scheduler/worker 运行集合；不存在 fixed/default/local tenant、空 tenant_id、禁用开关、兼容视图或旧配置键。

### Verification / Migration
对源码、生成产物、OpenAPI、配置、SQL、migration、fixture、前端、测试和文档执行 tenant/租户/旧 runtime API 零命中扫描，并验证新空库 schema 不含租户结构。

## ADR-018: 不透明数据库会话取代 JWT 与 Casbin

**Status:** accepted
**Source:** LOG-018
**Supersedes:** JWT access/refresh、浏览器与 WebView 可读 bearer token、Casbin URL policy

### Context
当前 Web cookie 可由 JavaScript 读取，Desktop 把 bearer token 放入 localStorage，系统没有完整 refresh lifecycle；JWT、会话撤销和 Casbin URL policy 还会形成多份身份与授权状态。

### Decision
IAM 生成高熵 opaque session token，数据库只持久化 token hash，并管理撤销、空闲超时、绝对超时和周期轮换。Web 使用受 CSRF 防护的 <Code>__Host-*</Code> Secure/HttpOnly/SameSite cookie；Desktop 由 Tauri transport proxy 从 Stronghold 注入，JavaScript 不可读取。密码使用 Argon2id。删除 JWT、refresh token 与 Casbin，由 IAM application policy 依据稳定 permission code 实现 RBAC 与数据范围。

### Trade-off
每个认证请求需要一次数据库或有界本地缓存查找，且需要显式实现 session lifecycle 和 IAM policy；换取即时撤销、统一状态、较小客户端攻击面以及与 URL 解耦的授权模型。

### Consequences
OpenAPI 使用 cookie/session security contract；Web 和 Desktop 仅在 adapter 层区分凭据携带；权限码成为 API、IAM policy 与前端 capability 的稳定合同；日志、审计和错误禁止出现原始 token。

### Verification / Migration
建立 session create/rotate/revoke/idle/absolute timeout contract suite、Web CSRF/cookie 属性测试、Desktop WebView token 不可见 tracer、Argon2id 参数测试和 RBAC/数据范围矩阵；对 JWT、refresh、Casbin、localStorage token 与 JavaScript 可读 auth cookie 执行零命中门禁。

## ADR-019: Redis 零保留与数据库内可靠协调

**Status:** accepted
**Source:** LOG-019
**Supersedes:** Redis cache/queue/lock、memory queue 可靠交付、每副本 scheduler

### Context
当前 Server profile 强制 PostgreSQL + Redis cache/queue，而 Desktop 使用 SQLite + memory 实现。三种正式数据库 profile、桌面离线交付和租户零保留要求统一 session、任务与可靠事件语义。

### Decision
完整删除 Redis。Session、task definition 与 reliable event 存在当前数据库；性能缓存仅使用可禁用的有界进程内实现，系统正确性不得依赖缓存。可靠异步采用 transactional outbox。Server SQLite 与 Desktop SQLite 仅支持单实例；Server PostgreSQL 支持 API 多副本，并由独立 worker 使用固定 PostgreSQL advisory lock 选出唯一 scheduler/outbox executor。Compose 同步删除 Redis service、volume 与配置。

### Trade-off
数据库承担额外 session/outbox 写入与 worker polling，需要容量、索引、清理和锁监控；换取更少基础设施、三 profile 一致恢复模型和桌面端可独立运行。

### Consequences
API replica 不直接拥有调度执行权；outbox consumer 必须幂等；缓存可随进程丢失；SQLite profile 明确拒绝多副本；PostgreSQL worker 在 advisory lock 连接断开后停止执行并允许接管。

### Verification / Migration
验证业务事务与 outbox 原子性、失败重试/幂等、进程崩溃恢复、PostgreSQL leader handoff、SQLite 单实例拒绝、缓存禁用等价行为；对依赖、源码、配置、Compose、volume、脚本、测试和文档执行 Redis 零命中扫描。

## ADR-020: 三 Profile 不可变类型化配置与 Secret 边界

**Status:** accepted
**Source:** LOG-020
**Supersedes:** Viper/global mutable config、多份全量 settings YAML、原始配置打印

### Context
当前服务直接读取全局配置，诊断命令能够打印 JWT 和数据库对象，示例与部署模板同时携带 JWT、Redis、多数据库、MySQL/SQL Server 和明文连接信息。这样的配置面与三种正式 profile、桌面宿主边界及租户/Redis/JWT 零保留决策冲突。

### Decision
为 <Code>server-postgres</Code>、<Code>server-sqlite</Code>、<Code>desktop-sqlite</Code> 定义独立最小强类型 schema。启动时按 <Code>defaults &lt; config file &lt; environment &lt; explicit non-secret CLI flags</Code> 合成、一次校验并通过构造函数注入不可变快照。Secret 只能来自环境变量或 <Code>_FILE</Code> reference，禁止 secret CLI flag、明文合并文件、全局可变配置、运行时 reload 和原始配置打印。Desktop 数据/日志路径、随机端口及一次性启动材料由 Tauri 宿主提供。删除 JWT、Redis、租户、多数据源、MySQL 与 SQL Server 配置键。

### Trade-off
需要维护三个小型 schema 及显式 mapping，动态修改基础设施配置必须重启进程；换取更小攻击面、清晰宿主能力、可预测 precedence 和构造期依赖可见性。

### Consequences
业务模块不读取环境变量或配置文件；配置 loader 属于 app bootstrap；容器通过 secret file 注入，Desktop 不消费服务端 settings 文件；错误只报告字段路径和验证规则，不包含敏感值。

### Verification / Migration
为三个 profile 建立 schema/required-field/unknown-field/precedence 测试，验证 secret redaction 与 <Code>_FILE</Code> 读取、Desktop 宿主材料注入及配置对象不可变性；对全仓执行 JWT、Redis、tenant、多数据源、MySQL、SQL Server、global config 和原始 config dump 零命中检查。

## ADR-021: 目标垂直切片施工与最终原子切换

**Status:** accepted
**Source:** LOG-021
**Supersedes:** 一次性全仓修改后首次验证、长期新旧双轨与兼容 adapter

### Context
重构同时改变目录、Go 模块边界、pnpm workspace、Tauri 宿主、数据库与 migration、认证授权、API 合同及发布矩阵。最终不保留任何旧兼容，但一次性完成全部修改后才验证会使失败原因和能力缺口难以定位。

### Decision
在新结构中按合同与骨架、IAM 端到端切片、其余业务模块、共享 Web/Desktop、部署发行的顺序施工。每个目标垂直切片立即通过其单元、合同与集成门禁。全部能力与发行证据完成后执行一次原子切换，同时启用新根入口、合同和交付路径，并删除旧目录、旧 API、旧 schema、旧配置、旧命名与临时迁移代码。开发期间不维护可发布的新旧双轨，不建设数据/API 双写或兼容层。

### Trade-off
施工期仓库会短暂包含不可发布的新骨架和仍用于能力对照的旧产品，需要严格 ownership 与依赖检查；换取短反馈、可审查增量以及最终单一权威结构。

### Consequences
Ticket 按垂直能力与明确依赖排序，而不是按机械目录搬迁拆分；新代码不得依赖旧 service/repository；临时施工工具必须绑定删除 Ticket；原子切换前不得声明产品级完成或发布混合版本。

### Verification / Migration
每个切片产生 contract、unit、双方言 integration 和必要 UI 证据；原子切换门禁覆盖根任务、完整功能矩阵、发布制品以及旧目录/module path/API/schema/config/dependency/命名/临时代码的 allowlist 零命中扫描。

## ADR-022: 本地、PR 与受保护发行的风险分层门禁

**Status:** accepted
**Source:** LOG-022
**Supersedes:** 仅 lint/unit/build 验收、人工桌面与发行验收、每 PR 完整签名发行

### Context
目标架构的风险跨越依赖方向、生成合同、双方言行为、认证授权、可靠异步、Web/Desktop 宿主及三个原生发行平台。所有检查都延迟到发行会导致反馈过晚，而每个 PR 都执行签名、公证和原生安装又会放大成本与凭据暴露。

### Decision
建立三层强制门禁。Local/Hook 执行 format、静态检查、generated drift、secret 与遗留技术零命中。PR 执行 Go/pnpm build/test、Go import 方向、workspace cycle/deep-import、OpenAPI lint/bundle/generate/conformance、PostgreSQL/SQLite migration/repository、session/CSRF/RBAC/outbox、Web E2E 与 Tauri sidecar tracer。Protected Release 在原生 runner 执行 Linux amd64/arm64 OCI、macOS Universal DMG、Windows x64 NSIS 的构建、安装、smoke、签名/公证、SBOM 和 provenance。任一必需门禁失败阻止对应合并或发行。

### Trade-off
CI 矩阵、原生 runner、测试数据和制品证据需要持续维护；换取可定位的短反馈、可审计发行和对“功能完整”而非仅编译成功的自动证明。

### Consequences
根 Taskfile 是本地与 CI 的共同执行合同，CI YAML 只负责触发、矩阵和凭据边界；总体 coverage 只作趋势信号，不能替代关键路径测试；flaky test 不得靠静默 retry 变绿。豁免必须记录 owner、理由、范围、到期时间和修复 Ticket。

### Verification / Migration
为每个门禁建立明确 required job、输出证据路径、超时和失败归属；在候选发行上验证制品到源码/SBOM/provenance 的可追溯性；定期演练失败阻断、豁免到期和原生安装 smoke，确保门禁本身没有失效。
