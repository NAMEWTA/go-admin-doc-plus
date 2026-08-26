# 项目架构重构设计日志

本文件按轮次记录高影响架构决策。

## LOG-001 — 2026-08-26T03:36:18Z — 自主架构与旧模式兼容
- **设计树节点：** D-001
- **轮次与依赖：** round 1 / 无
- **状态：** confirmed
- **问题：** 脱离上游后，是否继续为旧目录、旧内部入口和旧描述保留兼容模式。
- **事实与来源：** 用户明确允许没有任何兼容性保留的重构，并以功能完整作为结果要求；USER-DECISION:2026-08-26。
- **选项：** 继续兼容上游；保留旧入口渐进演进；以新架构为唯一权威。
- **推荐：** 以新架构为唯一权威，不为旧目录、内部包、脚本和文档保留永久兼容层。
- **结论：** 项目自主演进，旧目录、内部导入、脚本入口和旧描述不构成目标合同。
- **原因：** 继续兼容会让上游历史结构支配新系统边界，违背本次重构目标。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** “无需旧模式兼容”不自动回答业务 API、数据库数据和可观察行为是否可破坏，该边界由 D-004 单独决定。
- **后续：** D-004 明确功能完整的外部边界。
- **替代/被替代：** 无

## LOG-002 — 2026-08-26T03:36:18Z — 产品子目录规范名
- **设计树节点：** D-002
- **轮次与依赖：** round 1 / 无
- **状态：** confirmed
- **问题：** 前后端产品子目录如何命名。
- **事实与来源：** 用户指定 go-admin-plus 与 go-admin-plus-ui；USER-DECISION:2026-08-26。
- **选项：** 保持当前 go-admin-ui-plus；使用通用 backend/frontend；使用产品名 go-admin-plus/go-admin-plus-ui。
- **推荐：** 使用用户指定的产品名，并通过根任务隐藏日常工作目录细节。
- **结论：** Go 服务端目录固定为 <Path>go-admin-plus/</Path>，前端工作区目录固定为 <Path>go-admin-plus-ui/</Path>。
- **原因：** 名称与产品一致，且不再保留上游反序命名。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 全仓引用、CI、Docker、桌面宿主、文档和生成器必须原子更新，不保留旧路径 alias。
- **后续：** 下游 Ticket 负责全仓路径迁移与引用扫描。
- **替代/被替代：** 无

## LOG-003 — 2026-08-26T03:36:18Z — 前端 pnpm Monorepo 基线
- **设计树节点：** D-003
- **轮次与依赖：** round 1 / 无
- **状态：** confirmed
- **问题：** 前端是否采用 pnpm workspace 多包架构，并复用 Plus UI 的领域化原则。
- **事实与来源：** 用户明确要求 pnpm Monorepo 并提供完整 Plus UI 设计工件作为参考；当前前端已有 apps、domains、packages 和 pnpm-workspace，但 domain 同时包含 API 与 Vue 页面、包清单未声明实际依赖。USER-DECISION:2026-08-26；CODE:<Path>go-admin-ui-plus/pnpm-workspace.yaml</Path>；CODE:<Path>go-admin-ui-plus/domains/</Path>。
- **选项：** 回到单包；维持形式 workspace；建立具有真实 package 依赖与表现层分离的 workspace。
- **推荐：** 建立真实 workspace 依赖、公开 exports、禁止 deep import 与循环依赖，并按 App/Domain/Web Domain/Platform/Adapter 职责分层。
- **结论：** 前端目标采用 pnpm workspace；Plus UI 的编译期组合、公开入口、领域/表现分离原则是本 change 的设计输入。
- **原因：** 当前已有可利用的多包骨架，但依赖声明和层次还不足以形成可执行架构合同。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 不机械复制另一个项目的应用数量、兼容迁移策略或领域名称；这些由本项目设计树决定。
- **后续：** D-008 与 D-009 决定本项目应用拓扑和包边界。
- **替代/被替代：** 无

## Research: 全仓架构重构的一手证据
- Decision / target: 支持 D-004 至 D-007，并约束后续目录、模块和桌面设计。
- Scope / version: 2026-08-26；当前 Go 1.26、pnpm 11、Vue 3、Wails 2 实现与 Tauri 2 候选。
- Stop condition: 官方语言/工具规范、至少三个成熟开源仓库结构和用户提供的两个参考项目能够解释主要取舍。

### R-001
- Claim: GitHub 只从仓库根的 .github/workflows 发现工作流，因此单 Git 仓库的 workflow、Issue 和 PR 治理必须统一位于根 .github。
- Type: official fact
- Source: <Url>https://docs.github.com/en/actions/concepts/workflows-and-actions/workflows</Url>
- Confidence: high
- Limits: 子目录可以保存被根 workflow 调用的普通脚本，但不能拥有独立可发现的 workflow 根。
- Artifact impact: D-005、D-014

### R-002
- Claim: Husky 的 hook 属于 Git 仓库；当前端 package 不在 Git 根时，官方做法是由 package 的 prepare 切换到父目录安装指定 hook 目录，而不是把子目录当成第二个 Git 根。
- Type: official fact
- Source: <Url>https://typicode.github.io/husky/how-to.html</Url>
- Confidence: high
- Limits: hook 可调用前端命令，但唯一 hook 目录仍应由仓库根治理。
- Artifact impact: D-005、D-014

### R-003
- Claim: Go 官方针对 server project 推荐将多个命令放入 cmd，将不对外承诺的服务逻辑尽量放入 internal；Go package 名应简短、小写、单词化。
- Type: official fact
- Source: <Url>https://go.dev/doc/modules/layout</Url>；<Url>https://go.dev/doc/effective_go#package-names</Url>
- Confidence: high
- Limits: 官方不规定领域驱动目录名，具体 bounded context 仍需按本项目耦合事实决定。
- Artifact impact: D-006、D-010、D-011

### R-004
- Claim: pnpm workspace 要求根 pnpm-workspace.yaml；workspace 协议能保证本地包解析，单一 lockfile 仍只允许 package 使用其 manifest 明确声明的依赖，且可配置禁止 workspace cycle。
- Type: official fact
- Source: <Url>https://pnpm.io/workspaces</Url>
- Confidence: high
- Limits: pnpm 解决包管理，不自动证明业务领域划分正确。
- Artifact impact: D-003、D-009、D-014、D-017

### R-005
- Claim: pnpm catalog 用于在多个 package 中集中管理共享依赖版本，减少重复版本、升级面和 manifest 冲突。
- Type: official fact
- Source: <Url>https://pnpm.io/catalogs</Url>
- Confidence: high
- Limits: catalog 只集中版本，不应集中每个 package 的真实依赖声明。
- Artifact impact: D-009、D-014

### R-006
- Claim: Vue Vben Admin 采用 apps、packages、internal tooling 和 root scripts 的大型 pnpm/Turbo Monorepo；Grafana 按 API、domain services、infra、frontend features 和 shared packages 分责；n8n 在根分离 .github、docker、packages、scripts；PocketBase 让可执行迁移靠近应用。
- Type: code fact
- Source: <Url>https://github.com/vbenjs/vue-vben-admin</Url>；<Url>https://github.com/grafana/grafana</Url>；<Url>https://github.com/n8n-io/n8n</Url>；<Url>https://github.com/pocketbase/pocketbase</Url>
- Confidence: high
- Limits: 热门仓库证明的是可行组织方式，不是可复制模板；本项目规模较小，不应照搬其全部工具复杂度。
- Artifact impact: D-005、D-006、D-009、D-012、D-013、D-014

### R-007
- Claim: 用户提供的 Java 参考把可执行宿主与认证放在 admin，将业务能力放在 modules，将跨模块服务/DTO/事件放在 api，将基础设施按能力拆在 common；其核心价值是依赖职责，而不是 Maven 目录名本身。
- Type: code fact
- Source: USER-REFERENCE:ruoyi-vue-plus-namewta
- Confidence: high
- Limits: Java Maven 多模块与 Go package/internal 可见性不同，逐字翻译会制造 common 杂物箱和无必要的构建模块。
- Artifact impact: D-006、D-010、D-011

### R-008
- Claim: Tauri 2 可以把任意语言编译的 API server 作为带 target triple 的 sidecar，并通过 capabilities 精细限制前端对 sidecar/命令的权限；Wails 2 则原生把 Go 与 WebView 生命周期和 bindings 放在同一应用中。
- Type: official fact
- Source: <Url>https://v2.tauri.app/develop/sidecar/</Url>；<Url>https://v2.tauri.app/security/capabilities/</Url>；<Url>https://wails.io/docs/guides/application-development/</Url>
- Confidence: high
- Limits: 这些事实不替代产品取舍；Tauri 会引入 Rust host、sidecar 进程监督和跨平台打包矩阵。
- Artifact impact: D-007、D-008、D-013、D-017

### Conflicts and Unknowns
- Docker、SQL 与发行资产是否全部进入 release 的冲突已由 D-005 解决：根级统一治理，但按 deploy、release、database 与应用内 migration 的生命周期分域。
- “功能完整”与外部兼容的歧义已由 D-004 解决：旧 API、旧数据、旧 schema 和旧配置不保留，只按新目标能力验收。
- 桌面宿主歧义已由 D-007 解决：Tauri 2 + Go sidecar 直接替换 Wails 2。

### Recommendation
- 采用单仓根治理，但使用 .github、.husky、scripts、deploy、release 和应用内 migrations 的生命周期边界。
- 后端采用 Go 原生 cmd/internal/modules/contracts/platform 语法，吸收 Java 参考的职责而不复制 Maven 形态。
- 前端在现有 workspace 基础上建立真实依赖声明，并拆分无头领域与 Vue 表现层。
- 桌面直接使用 Tauri 2 + Go sidecar 替换 Wails，不维护双宿主。

## LOG-004 — 2026-08-26T03:57:08Z — Greenfield 功能合同
- **设计树节点：** D-004
- **轮次与依赖：** round 2 / D-001
- **状态：** confirmed
- **问题：** 旧 HTTP API、数据库数据、schema、配置键和操作模式是否仍属于必须保留的产品合同。
- **事实与来源：** 用户选择完全重构，并明确旧 API、旧数据和旧模式均可丢弃，只要求最终功能完整；USER-DECISION:2026-08-26。
- **选项：** 保留全部外部合同；重设计合同但迁移旧数据；完全 Greenfield。
- **推荐：** 原推荐保留目标业务行为并迁移数据；用户选择更彻底的 Greenfield 边界。
- **结论：** 旧 HTTP API、旧数据库数据、旧 schema、旧配置键和旧操作模式不构成目标合同；新系统以确认后的目标能力和新数据基线为准。
- **原因：** 用户希望本次直接达到现代目标状态，不让历史接口与数据结构约束设计。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** “功能完整”必须由目标能力清单和验收场景证明，不能仅以旧接口数量或旧页面可打开作为替代指标。
- **后续：** D-012 设计全新 schema 基线，D-015 设计新 API 合同，D-016 禁止运行时双轨。
- **替代/被替代：** 收紧 LOG-001 中待定的外部兼容边界。

## LOG-005 — 2026-08-26T03:57:08Z — 根级资产按生命周期治理
- **设计树节点：** D-005
- **轮次与依赖：** round 2 / D-001、D-002
- **状态：** confirmed
- **问题：** 根级治理资产是全部进入 release，还是按生命周期分域。
- **事实与来源：** GitHub Actions 只从根 <Path>.github/workflows/</Path> 发现 workflow；Husky hook 属于 Git 仓库；当前部署、发行与子项目 Docker/SQL/脚本重复散落。USER-DECISION:2026-08-26；R-001；R-002；CODE:<Path>deploy/</Path>；CODE:<Path>release/</Path>。
- **选项：** 全部集中到 release；按子项目保留；根级统一且按生命周期分域。
- **推荐：** 根级统一且分为 .github、.husky、scripts、deploy、release、database。
- **结论：** 采用生命周期分域；子项目不再拥有仓库治理、独立发布 workflow、通用部署定义或重复 Git hook。
- **原因：** 集中所有权同时保留开发、部署、发行和数据库工程之间真实不同的变更节奏与职责。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** App 自有的构建元数据仍随 App，例如 Tauri 的 <Path>src-tauri/</Path> 和 Go 领域迁移；根目录只拥有跨产品编排与工程资产。
- **后续：** D-013 决定 deploy/release 明细，D-014 决定 scripts/Taskfile/Husky 唯一入口，D-012 决定 database 与应用迁移边界。
- **替代/被替代：** 无

## LOG-006 — 2026-08-26T03:57:08Z — Go 原生模块化单体
- **设计树节点：** D-006
- **轮次与依赖：** round 2 / D-001、D-002
- **状态：** confirmed
- **问题：** 后端逐字采用 Java Maven 顶级模块，还是翻译为 Go 原生结构。
- **事实与来源：** Go 官方推荐 server 项目以 <Path>cmd/</Path> 承载命令、<Path>internal/</Path> 隐藏实现；当前 <Path>app/other/</Path> 反向导入 admin，<Path>common/</Path> 同时容纳 middleware、ORM、storage、response 等多类职责。USER-DECISION:2026-08-26；R-003；R-007；CODE:<Path>go-admin-plus/app/</Path>；CODE:<Path>go-admin-plus/common/</Path>。
- **选项：** 照搬 go-admin/go-common/go-app/go-api；保留当前 app/common；采用 cmd/internal/modules/contracts/platform。
- **推荐：** 采用 Go 原生模块化单体并消除 common 杂物箱。
- **结论：** 后端目标语法为 <Path>cmd/</Path>、<Path>internal/app/</Path>、<Path>internal/modules/</Path>、<Path>internal/contracts/</Path>、<Path>internal/platform/</Path>；Java 参考只提供职责映射。
- **原因：** 利用 Go package 与 internal 可见性表达边界，比模拟 Maven 多模块更符合语言工具链并减少空壳层级。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 不创建新的通用 common 包；每个子包必须有单一职责和真实消费者，不强制每个领域机械拥有全部分层目录。
- **后续：** D-010 决定业务域，D-011 决定跨域合同，D-012 决定迁移所有权。
- **替代/被替代：** 无

## LOG-007 — 2026-08-26T03:57:08Z — Tauri 2 桌面宿主
- **设计树节点：** D-007
- **轮次与依赖：** round 2 / D-001、D-003
- **状态：** confirmed
- **问题：** 桌面端继续使用 Wails 2，还是替换为 Tauri 2 + Go sidecar。
- **事实与来源：** Tauri 2 官方支持以 <Code>externalBin</Code> 打包任意语言 sidecar，并按目标 triple 解析平台二进制；capabilities 可限制窗口权限。当前桌面通过 Wails binding 启动同一 Go 应用。USER-DECISION:2026-08-26；R-008；<Url>https://v2.tauri.app/reference/config/</Url>。
- **选项：** 保留 Wails；双宿主过渡；直接切换 Tauri 2 + Go sidecar。
- **推荐：** 直接替换 Tauri 2，不保留 Wails。
- **结论：** Tauri 2 是唯一桌面宿主；Go 后端编译为按 target triple 命名的 sidecar，由宿主管理生命周期、loopback 地址、启动令牌和退出清理。
- **原因：** 这是用户指定的现代目标状态，且 Tauri 的权限与打包模型能把宿主能力显式化。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 禁止固定端口、令牌进入 URL/query 或浏览器存储；桌面构建必须为每个目标提供匹配 sidecar，并验证异常退出与孤儿进程清理。
- **后续：** D-008 决定桌面 App 与 Web 组合，D-013 决定平台发行矩阵，D-017 决定原生验证。
- **替代/被替代：** 替代当前 Wails 2 桌面实现。

## Research: Tauri 2 发行与多目标约束补充
- Decision / target: 支持 D-008 与 D-013 的 App 所有权和发行矩阵。
- Scope / version: Tauri 2 官方文档，检索于 2026-08-26。
- Stop condition: 官方配置、sidecar 和 CI 发行文档足以确定工程资产所有权与目标矩阵约束。

### R-009
- Claim: Tauri 配置默认位于应用自身的 <Path>src-tauri/tauri.conf.json</Path>，并支持 macOS、Windows、Linux 平台覆盖配置；因此它是 desktop App 自有构建元数据，而不是根 release 中的通用脚本。
- Type: official fact
- Source: <Url>https://v2.tauri.app/reference/config/</Url>
- Confidence: high
- Limits: 根 release 仍可编排版本、签名、SBOM 与候选收集。
- Artifact impact: D-008、D-013

### R-010
- Claim: Tauri 官方 CI 示例覆盖 Windows x64、Linux x64/Arm64、macOS x64/Arm64；sidecar 必须为每个目标 triple 提供匹配二进制，平台分发通常需要签名，macOS 站外分发还需要 notarization。
- Type: official fact
- Source: <Url>https://v2.tauri.app/distribute/pipelines/github/</Url>；<Url>https://v2.tauri.app/develop/sidecar/</Url>；<Url>https://v2.tauri.app/distribute/</Url>
- Confidence: high
- Limits: 官方支持范围不等于本产品首发支持范围；每增加一个目标都会增加 Go/Rust/Node 构建、原生验证和签名成本。
- Artifact impact: D-013、D-017

## LOG-008 — 2026-08-26T04:07:24Z — Web 与桌面双 App 拓扑
- **设计树节点：** D-008
- **轮次与依赖：** round 3 / D-003、D-007
- **状态：** confirmed
- **问题：** Web 与 Tauri 桌面是否建立独立 App，以及如何共享业务组合。
- **事实与来源：** 当前只有 <Path>apps/admin/</Path>，通过 Wails 全局 binding 在同一源码中隐式切换运行时；Tauri 配置是 desktop App 自有构建元数据。USER-DECISION:2026-08-26；R-009；CODE:<Path>go-admin-ui-plus/apps/admin/</Path>。
- **选项：** 双 App 共享业务组合；单 App 环境切换；桌面只包装 Web 构建产物。
- **推荐：** 建立 admin-web 与 admin-desktop 两个独立构建 App，复用同一组合清单和业务页面。
- **结论：** <Path>apps/admin-web/</Path> 与 <Path>apps/admin-desktop/</Path> 是两个交付 App；desktop 只增加 Tauri 宿主、sidecar bootstrap 和桌面 adapter，不复制业务页面。
- **原因：** 浏览器与桌面的启动、权限和打包合同不同，但产品能力与页面不应分叉。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 两个 App 必须各自可 dev/typecheck/test/build；共享组合清单必须能证明路由与能力一致，宿主专属功能通过 capability/port 表达。
- **后续：** D-009 决定共享包图与领域/表现分离，D-015 决定共同 API 合同。
- **替代/被替代：** 替代当前单一 admin App 的隐式 Web/Wails 双模式。

## LOG-009 — 2026-08-26T04:07:24Z — Go 业务能力地图
- **设计树节点：** D-010
- **轮次与依赖：** round 3 / D-006
- **状态：** confirmed
- **问题：** admin、other、jobs、demo 应被哪些稳定业务能力替代。
- **事实与来源：** 当前 admin 混合身份、RBAC、组织、配置、字典和审计；other 混合文件、生成器与监控且直接导入 admin；jobs 与 demo 边界相对独立。USER-DECISION:2026-08-26；CODE:<Path>go-admin-plus/app/admin/</Path>；CODE:<Path>go-admin-plus/app/other/</Path>。
- **选项：** 八个细粒度能力模块；六个粗粒度模块；仅重命名旧四模块。
- **推荐：** 使用 iam、organization、settings、audit、scheduler、generator、files、demo，并把运维端点移出业务域。
- **结论：** <Path>internal/modules/</Path> 下采用上述八个能力模块；health、metrics 与 server status 由应用运维和 <Path>internal/platform/observability/</Path> 承担。
- **原因：** 以业务变化原因划分模块，既消除 other/common，也避免 system/operations 再次成为无边界聚合。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** package 名使用简短小写单词；模块不按 ORM 表一表一包；运维端点不得依赖业务 transport 或 ORM model。
- **后续：** D-011 决定跨模块协作，D-012 决定各模块 migration，D-015 映射 HTTP 合同。
- **替代/被替代：** 替代 app/admin、app/other、app/jobs、app/demo 的职责地图。

## LOG-010 — 2026-08-26T04:07:24Z — 首期部署与发行矩阵
- **设计树节点：** D-013
- **轮次与依赖：** round 3 / D-005、D-007
- **状态：** confirmed
- **问题：** 服务端与桌面端首期支持目标，以及 deploy/release/App 元数据的所有权。
- **事实与来源：** 当前支持 Linux AMD64 Compose、macOS ARM64 Wails DMG、Windows AMD64 Wails NSIS；Tauri 官方矩阵支持 Linux x64/Arm、macOS x64/Arm 与 Windows x64，并要求每目标提供匹配 sidecar。USER-DECISION:2026-08-26；R-009；R-010；CODE:<Path>deploy/compose/</Path>；CODE:<Path>release/</Path>。
- **选项：** Linux 双架构服务端 + macOS Universal/Windows x64 桌面；再增加 Linux 桌面；维持当前单架构矩阵。
- **推荐：** 首期采用 Linux amd64/arm64 OCI 与 Compose、macOS Universal DMG、Windows x64 NSIS，不提供 Linux 桌面包。
- **结论：** <Path>deploy/</Path> 管镜像与环境运行定义；<Path>release/</Path> 管版本、签名、公证、安装器编排、SBOM、provenance 与候选收敛；<Path>apps/admin-desktop/src-tauri/</Path> 管 Tauri App 构建元数据。
- **原因：** 覆盖主流服务端架构和现有桌面平台，同时避免首期承担 Linux 桌面发行与原生验证成本。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** macOS Universal 必须包含 Universal Go sidecar；Linux OCI 必须以同一版本和 provenance 发布 amd64/arm64 manifest；无签名工件只能明确标记为本地自用候选。
- **后续：** D-014 决定本地可复现任务入口，D-017 决定多目标发布门禁。
- **替代/被替代：** 扩展当前 Linux AMD64/macOS ARM64/Windows AMD64 矩阵并替换 Wails 打包。

## Research: Round 4 代码事实
- Decision / target: 支持 D-009、D-011 与 D-012 的下一轮选择。
- Scope / version: 当前工作树，检索于 2026-08-26。
- Stop condition: 已定位前端 deep import、后端反向依赖以及数据库驱动/运行 profile 的实际边界。

### R-011
- Claim: 当前五个前端 domain 都把 api 与 Vue pages 放在同一 package，并存在至少 78 处跨 package 相对 deep import；子 package manifest 没有声明对应真实依赖。
- Type: code fact
- Source: CODE:<Path>go-admin-ui-plus/domains/</Path>；CODE:<Path>go-admin-ui-plus/packages/</Path>
- Confidence: high
- Limits: 计数用于证明边界未生效，不作为目标 package 数量依据。
- Artifact impact: D-009、D-014、D-017

### R-012
- Claim: 当前 common middleware 直接导入 admin DTO/model，other generator 直接导入 admin service/DTO，模块装配还通过 global/runtime 取得数据库与 Casbin；跨模块协作没有稳定 application contract。
- Type: code fact
- Source: CODE:<Path>go-admin-plus/common/middleware/</Path>；CODE:<Path>go-admin-plus/app/other/</Path>；CODE:<Path>go-admin-plus/internal/modules/</Path>
- Confidence: high
- Limits: 目标合同应按真实用例建立，不能把所有旧 DTO 搬入一个全局 contracts 包。
- Artifact impact: D-011、D-015、D-017

### R-013
- Claim: 旧通用配置名义支持 MySQL、PostgreSQL、SQL Server 与 SQLite；当前 server profile 和 Compose 已固定 PostgreSQL，desktop profile 固定 SQLite，但旧迁移仍包含多驱动分支和 SQL 文件。
- Type: code fact
- Source: CODE:<Path>go-admin-plus/go.mod</Path>；CODE:<Path>go-admin-plus/common/database/</Path>；CODE:<Path>go-admin-plus/internal/profile/</Path>；CODE:<Path>deploy/compose/</Path>
- Confidence: high
- Limits: 驱动依赖存在不代表每个数据库都经过同等集成验证；Greenfield 目标需要显式决定正式数据库矩阵。
- Artifact impact: D-012、D-017

## LOG-011 — 2026-08-26T04:39:36Z — 前端领域与表现层分离
- **设计树节点：** D-009
- **轮次与依赖：** round 4 / D-008
- **状态：** confirmed
- **问题：** 无头领域、Vue 页面、宿主端口、adapter、共享 shell 与 UI 如何形成真实 pnpm 包图。
- **事实与来源：** 当前五个 domain 均把 API 与 Vue pages 混在同一 package，并存在至少 78 处跨包相对 deep import。USER-DECISION:2026-08-26；R-004；R-011。
- **选项：** 无头领域与 Web Domain 分离；每 feature 混合全部层；维持当前 domains。
- **推荐：** 使用 domains/web-domains/platform/adapters/app-shell/ui 分责，并让 Apps 只做组合。
- **结论：** <Path>packages/domains/*</Path> 不依赖 Vue、Element Plus、router 或 HTTP；<Path>packages/web-domains/*</Path> 提供 Vue 页面、composables 与路由贡献；platform 定义宿主端口，browser/desktop adapter 实现端口，app-shell 共享组合，ui 提供通用表现组件。
- **原因：** 使业务规则可在两个 App 复用并独立测试，同时让宿主与表现依赖成为可执行的单向包图。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 所有内部依赖通过 workspace manifest 和公开 exports；禁止跨包 <Path>src/</Path> deep import、循环依赖以及 Apps 反向成为共享库。
- **后续：** D-014 决定包图门禁，D-015 决定 transport 类型生成。
- **替代/被替代：** 替代当前 domains 内 API/page 混合与相对路径共享。

## LOG-012 — 2026-08-26T04:39:36Z — Go 跨模块合同
- **设计树节点：** D-011
- **轮次与依赖：** round 4 / D-006、D-010
- **状态：** confirmed
- **问题：** 八个模块如何同步调用、发布事件和访问数据而不重建 common。
- **事实与来源：** 当前 middleware 导入 admin DTO/model，other generator 导入 admin service/DTO，global/runtime 暴露数据库和 Casbin。USER-DECISION:2026-08-26；R-012。
- **选项：** 消费者最小 port + integration event；提供者 application service 直连；中央 common/contracts。
- **推荐：** 消费者定义同步 port，跨模块异步使用不可变事件，app 层完成显式装配。
- **结论：** 同步 port 由消费者拥有并在 <Path>internal/app/</Path> 注入；<Path>internal/contracts/</Path> 只保存真实共享的 ID、值语义和 integration event。禁止跨模块 ORM model、repository、transport DTO、数据库 join 与全局 service locator。
- **原因：** 消费者控制最小依赖面，事件降低变化耦合，app 装配使运行时关系可见且可测试。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** contracts 不得成为 DTO 仓库；跨模块工作流由应用用例编排；事件只能携带稳定事实，不能传递可变 ORM 对象。
- **后续：** D-015 定义 HTTP 边界，D-020 定义 IAM 策略，D-017 建立 import/event 合同测试。
- **替代/被替代：** 替代当前 other/admin/common/global 的隐式反向依赖。

## LOG-013 — 2026-08-26T04:39:36Z — PostgreSQL/SQLite 正式数据库矩阵
- **设计树节点：** D-012
- **轮次与依赖：** round 4 / D-004、D-005、D-006、D-010
- **状态：** confirmed
- **问题：** 正式数据库目标、迁移序列和根 database 的所有权。
- **事实与来源：** 用户修正推荐：服务端需要同时支持 SQLite 与 PostgreSQL，桌面仅支持 SQLite；当前 profile 已分别存在 PostgreSQL server 与 SQLite desktop，但旧通用层仍携带 MySQL/SQL Server。USER-DECISION:2026-08-26；R-013。
- **选项：** Server PostgreSQL + Desktop SQLite；Server PostgreSQL/SQLite + Desktop SQLite；保留四数据库；全部 PostgreSQL。
- **推荐：** 原推荐只为 Server 保留 PostgreSQL；用户明确把 Server SQLite 提升为正式目标。
- **结论：** Server PostgreSQL、Server SQLite、Desktop SQLite 都是正式支持 profile；MySQL 与 SQL Server 被删除。各业务模块拥有从 Greenfield 基线开始的 migration，由统一 runner 组成不可变、只前进的序列；根 <Path>database/</Path> 只存开发 bootstrap、测试 fixture 与工程工具。
- **原因：** 同时满足完整服务部署与轻量单机服务部署，并让桌面维持自包含数据文件。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** Server SQLite 不是未验证的 dev fallback，必须覆盖完整功能与集成测试；所有 migration 必须同时通过 PostgreSQL/SQLite，方言差异必须显式隔离；API 进程启动不得隐式 AutoMigrate。
- **后续：** D-019 决定仓储/迁移工具链，D-021 决定租户模型，D-022 决定单节点与分布式协调。
- **替代/被替代：** 删除 MySQL、SQL Server 与旧多驱动迁移合同。

## Research: 持久化工具链补充
- Decision / target: 支持新增 D-019，并约束 PostgreSQL/SQLite 双矩阵的可实现性。
- Scope / version: 2026-08-26 官方文档与项目仓库。
- Stop condition: GORM、sqlc、Bun、Goose 与 CGo-free SQLite 的一手资料足以形成可比较选项。

### R-014
- Claim: GORM 当前提供 Generics API 并正式支持 PostgreSQL/SQLite；Bun 是基于 database/sql 的 SQL-first ORM/query builder，正式提供 PostgreSQL 与 SQLite dialect；sqlc 对 Go/PostgreSQL 为 stable，但官方支持表仍将 Go/SQLite 标为 beta。
- Type: official fact
- Source: <Url>https://gorm.io/docs/</Url>；<Url>https://github.com/uptrace/bun</Url>；<Url>https://docs.sqlc.dev/en/v1.28.0/reference/language-support.html</Url>
- Confidence: high
- Limits: 工具支持数据库不代表同一套查询天然跨方言；所有方案仍需双数据库行为测试和模块 repository 隔离。
- Artifact impact: D-019、D-017

### R-015
- Claim: Goose Provider 无全局状态、可消费 embed.FS、支持 PostgreSQL/SQLite dialect 和 Go/SQL migration；它可作为应用内统一 runner，而不需要依赖外部 goose CLI。
- Type: official fact
- Source: <Url>https://pressly.github.io/goose/documentation/provider/</Url>；<Url>https://github.com/pressly/goose</Url>
- Confidence: high
- Limits: migration SQL 仍需按方言设计；分散模块 migration 必须在构建时组合成确定的全局顺序。
- Artifact impact: D-012、D-019

### R-016
- Claim: modernc.org/sqlite 是 CGo-free 的 database/sql SQLite 驱动；官方 GORM SQLite driver 基于 CGo，同时列出了若干 pure-Go community dialector。CGo-free 驱动能显著简化 Go sidecar 的 Linux 双架构、macOS Universal 与 Windows 构建矩阵。
- Type: official/code fact
- Source: <Url>https://pkg.go.dev/modernc.org/sqlite</Url>；<Url>https://github.com/go-gorm/sqlite</Url>
- Confidence: high
- Limits: CGo-free 不是免验证；必须固定 driver/libc 版本并验证 WAL、foreign_keys、busy_timeout、备份和升级语义。
- Artifact impact: D-019、D-017

## LOG-014 — 2026-08-26T05:04:34Z — 根任务与 Hook 唯一入口
- **设计树节点：** D-014
- **轮次与依赖：** round 5 / D-005、D-006、D-009
- **状态：** confirmed
- **问题：** 本地开发、脚本、Husky 与 CI 如何共用一套产品命令。
- **事实与来源：** 当前已有根 Taskfile，但前端 package script、子项目 scripts、发布脚本和前端内 .husky 仍并存。USER-DECISION:2026-08-26；R-001；R-002；CODE:<Path>Taskfile.yml</Path>。
- **选项：** Taskfile 唯一入口；Makefile 唯一入口；子项目命令加根转发。
- **推荐：** 使用 Taskfile 编排，scripts 按生命周期分类，根 .husky 与 CI 调用相同任务。
- **结论：** 根 Taskfile 暴露 dev/build/test/lint/generate/migrate/package/release；<Path>scripts/{backend,frontend,contracts,desktop,deploy,release}/</Path> 保存复杂实现；唯一 <Path>.husky/</Path> 只运行快速 staged 门禁，CI 调用相同根任务的完整门禁。
- **原因：** 本地、Hook 与 CI 使用同一行为合同，同时避免每次 commit 运行耗时发行矩阵。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** package scripts 可作为包内原生命令，但不得成为文档或 CI 的产品级唯一入口；根任务必须可本地复现且不默认外部发布。
- **后续：** D-017 固定门禁层级与证据。
- **替代/被替代：** 替代子项目 .husky 与散落的产品级脚本入口。

## LOG-015 — 2026-08-26T05:04:34Z — OpenAPI 3.1 Contract-first
- **设计树节点：** D-015
- **轮次与依赖：** round 5 / D-004、D-006、D-009、D-011
- **状态：** confirmed
- **问题：** HTTP 合同、Go transport 与 TypeScript client 的唯一事实源。
- **事实与来源：** 当前手写 openapi.json 再通过自定义 Node 脚本扫描旧 Go model，无法从编译期约束 handler；oapi-codegen 支持 OpenAPI 3.1、strict server interface 与 Chi，openapi-typescript 支持 OpenAPI 3 类型生成。USER-DECISION:2026-08-26；CODE:<Path>go-admin-plus/api/openapi/</Path>；<Url>https://github.com/oapi-codegen/oapi-codegen/</Url>；<Url>https://github.com/openapi-ts/openapi-typescript</Url>。
- **选项：** OpenAPI contract-first；Go code-first；手工 OpenAPI + 单侧生成。
- **推荐：** OpenAPI 3.1 为根合同，生成两端 transport 代码，并显式映射领域模型。
- **结论：** <Path>contracts/openapi/</Path> 保存模块化 OpenAPI 3.1；生成 Go strict server interface/transport types 与 TypeScript transport types/client；Go HTTP 使用 Chi adapter；CI lint、bundle、generate 并拒绝 drift。
- **原因：** 同一合同在编译期约束服务实现和两个前端 App，消除自定义正则扫描与手工 DTO 漂移。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 生成代码不可手改；transport 类型不得成为领域模型；OpenAPI operationId、错误、分页、认证和幂等语义必须统一。
- **后续：** D-020 定义认证合同，D-017 定义生成与运行时 conformance 门禁。
- **替代/被替代：** 替代当前手写 openapi.json、Gin route inventory 和自定义前端字段扫描。

## LOG-016 — 2026-08-26T05:04:34Z — Bun 与 Goose 持久化工具链
- **设计树节点：** D-019
- **轮次与依赖：** round 5 / D-010、D-012
- **状态：** confirmed
- **问题：** PostgreSQL/SQLite 双矩阵的数据访问、领域隔离与 migration runner。
- **事实与来源：** Bun 正式提供 PostgreSQL/SQLite dialect 并基于 database/sql；Goose Provider 支持 embed.FS、双方言和无全局状态；CGo-free SQLite 有利于跨平台 sidecar。USER-DECISION:2026-08-26；R-014；R-015；R-016。
- **选项：** Bun + Goose；GORM Generics + Goose；sqlc 双生成 + Goose。
- **推荐：** SQL-first Bun + database/sql、模块私有 record/repository、Goose 前进迁移。
- **结论：** Repository adapter 使用 Bun；PostgreSQL 使用 Bun pgdriver，SQLite 使用固定且验证的 CGo-free database/sql driver；领域模型不带 Bun/database tag；Goose Provider 组合内嵌双方言 migration；禁止 AutoMigrate。
- **原因：** 查询保持 SQL 可见且可组合，同时避免 sqlc SQLite beta 与双方言查询包重复，也消除领域对 ORM 的耦合。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** module 只能访问自身 repository/table；方言特性必须封装并有等价行为测试；driver 及传递依赖必须固定并进入 SBOM。
- **后续：** D-017 建立双方言查询/migration/锁语义矩阵。
- **替代/被替代：** 替代 GORM model、AutoMigrate 和全局 database runtime。

## LOG-017 — 2026-08-26T05:04:34Z — 完整删除租户能力
- **设计树节点：** D-021
- **轮次与依赖：** round 5 / D-004、D-010、D-012、D-013
- **状态：** confirmed
- **问题：** Greenfield 产品是否保留任何租户或单租户兼容抽象。
- **事实与来源：** 用户明确要求取消租户功能和能力并同步修改 SQL；当前 tenant 关系进入 resolver、profile、platform dependencies、DB/Casbin/cron runtime、middleware、migration flag 与测试。USER-DECISION:2026-08-26；CODE:<Path>go-admin-plus/internal/tenant/</Path>。
- **选项：** 单实例但保留抽象；共享 schema 多租户；每租户数据库；完整删除租户能力。
- **推荐：** 原推荐单租户每部署实例并删除 resolver；用户进一步要求租户概念零保留。
- **结论：** 删除 tenant package、host/domain resolver、多数据库注册、tenant context、tenant-aware DB/Casbin/cron API 和所有租户配置；SQL/schema/migration/index/fixture 删除租户表、字段与约束；前端、测试、文档同步删除。
- **原因：** 产品没有租户能力，保留“固定 local tenant”仍会污染数据、授权、任务、缓存和接口模型。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 不保留空 tenant_id、默认 tenant、disabled flag、兼容视图或旧配置键；单进程只装配一个显式 Database、IAM policy/session store 与 Scheduler/worker 运行集合。
- **后续：** D-022 重设计无租户缓存/协调，D-023 删除配置中的租户与多数据源键，D-017 建立全仓零命中门禁。
- **替代/被替代：** 删除全部现有租户与多数据库选择能力。

## Research: 认证、租户与运行协调补充
- Decision / target: 支持 D-020、D-022 与租户零保留验收。
- Scope / version: 当前代码与 2026-08-26 官方安全资料。
- Stop condition: 已定位 token 存储、refresh 缺口、租户传播和 Redis 强依赖，并取得会话/密码/桌面密钥的一手规范。

### R-017
- Claim: 当前租户概念至少传播到 resolver、server/desktop profile、platform dependencies、数据库与 Casbin 注册、定时任务、migration CLI、中间件及审计写入；完整删除必须跨代码和 schema，而非删除一个 package。
- Type: code fact
- Source: CODE:<Path>go-admin-plus/internal/tenant/</Path>；CODE:<Path>go-admin-plus/internal/profile/</Path>；CODE:<Path>go-admin-plus/common/</Path>；CODE:<Path>go-admin-plus/app/jobs/</Path>
- Confidence: high
- Limits: 最终实现还必须通过全仓生成产物与 SQL 扫描发现新增引用。
- Artifact impact: D-021、D-022、D-023、D-017

### R-018
- Claim: 当前 Server profile 强制 PostgreSQL + Redis cache/queue，Desktop 使用 SQLite + memory cache/queue；审计写入通过 runtime queue，scheduler 按 tenant 创建 cron。租户删除后仍需决定单节点与多节点的队列、会话撤销和 scheduler ownership。
- Type: code fact
- Source: CODE:<Path>go-admin-plus/internal/profile/server.go</Path>；CODE:<Path>go-admin-plus/internal/profile/desktop.go</Path>；CODE:<Path>go-admin-plus/internal/modules/runtime_queue.go</Path>
- Confidence: high
- Limits: Redis 是否保留是产品运行 profile 决策，不能由当前实现倒推。
- Artifact impact: D-022、D-023

### R-019
- Claim: 当前 Web cookie 可被 JavaScript 读取，Desktop 把 bearer token 放入 localStorage；OWASP 明确不应在 Web Storage 保存认证 token，并推荐 HttpOnly/Secure/SameSite cookie。RFC 9700 要求公共客户端使用 sender-constrained refresh token 或 rotation；RFC 9106 要求支持 Argon2id；Tauri Stronghold 提供宿主安全 secret storage。
- Type: official/code fact
- Source: CODE:<Path>go-admin-ui-plus/apps/admin/src/utils/auth.js</Path>；<Url>https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Session_Management_Cheat_Sheet.md</Url>；<Url>https://www.rfc-editor.org/info/rfc9700/</Url>；<Url>https://datatracker.ietf.org/doc/html/rfc9106</Url>；<Url>https://v2.tauri.app/reference/javascript/stronghold/</Url>
- Confidence: high
- Limits: OAuth BCP 的 refresh 原则可用于本产品 session 设计，但本产品不是通用 OAuth authorization server；具体 cookie/domain 与桌面 vault 解锁需要按部署模型设计。
- Artifact impact: D-020、D-022、D-017

### R-020
- Claim: PostgreSQL advisory lock 提供应用定义的互斥语义，session lock 会在连接结束时由服务器清理，适合多副本中的 scheduler/outbox leader ownership；SQLite 官方定位是应用或设备的本地存储，并明确不适合由多台计算机通过网络文件系统直接并发访问同一数据库。因此 SQLite 产品 profile 应限定单实例，PostgreSQL profile 才提供多副本 worker 协调。
- Type: official fact
- Source: <Url>https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS</Url>；<Url>https://www.sqlite.org/whentouse.html</Url>；<Url>https://www.sqlite.org/wal.html</Url>
- Confidence: high
- Limits: advisory lock 只提供自愿互斥，所有 worker 必须遵守同一固定 lock key 与连接生命周期；outbox 的 claim、lease、retry 与幂等仍需由数据库事务和业务键保证。
- Artifact impact: D-022、D-017

## LOG-018 — 2026-08-26T05:16:55Z — 不透明数据库会话与 IAM 权限策略
- **设计树节点：** D-020
- **轮次与依赖：** round 6 / D-010、D-011、D-012、D-015、D-021
- **状态：** confirmed
- **问题：** Web 与 Desktop 如何统一登录、撤销、权限、数据范围和凭据存储，同时消除当前可读 token 与无 refresh 的安全缺口。
- **事实与来源：** 当前 Web cookie 可由 JavaScript 读取，Desktop 把 bearer token 存入 localStorage；OWASP 推荐 Secure/HttpOnly/SameSite cookie 且不应将认证 token 放入 Web Storage，RFC 9106 要求支持 Argon2id，Tauri Stronghold 提供宿主安全 secret storage。USER-DECISION:2026-08-26；R-019。
- **选项：** JWT access + refresh session；纯数据库不透明 session；保留当前 JWT/Casbin。
- **推荐：** 用户选择纯数据库不透明 session，并要求删除 JWT、refresh token 与 Casbin。
- **结论：** IAM 生成高熵 opaque token，数据库只持久化 token hash，并实现撤销、空闲超时、绝对超时和周期轮换；Web 通过受 CSRF 防护的 <Code>__Host-*</Code> Secure/HttpOnly/SameSite cookie 携带；Desktop 由 Tauri transport proxy 从 Stronghold 注入，JavaScript 不可读取；密码使用 Argon2id；IAM 以稳定 permission code 实现 RBAC 与数据范围，不以 URL 或 Casbin policy 为权限事实源。
- **原因：** 所有会话撤销与权限状态在当前数据库中强一致，不再承担 JWT 双 token 协议、浏览器可读 token 和 Casbin URL policy 漂移。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 日志、审计、配置和错误不得输出原始 session token；cookie 不使用 Domain 且 Path=/；状态改变请求必须验证 CSRF；权限码是 API 合同和前端 capability 的稳定标识，transport route 不能直接充当权限码。
- **后续：** D-023 定义 secret/config 输入，D-017 固定 session lifecycle、CSRF、权限矩阵与零 JWT/Casbin 扫描。
- **替代/被替代：** 替代 JWT、refresh token、bearer localStorage/js-cookie 与 Casbin 授权模型。

## LOG-019 — 2026-08-26T05:16:55Z — Redis 零保留与数据库协调
- **设计树节点：** D-022
- **轮次与依赖：** round 6 / D-010、D-012、D-013、D-021
- **状态：** confirmed
- **问题：** 三个数据库 profile 在无租户架构下如何处理缓存、session、可靠异步、scheduler ownership 与多副本协调。
- **事实与来源：** 当前 Server 强制 PostgreSQL + Redis cache/queue，Desktop 使用 SQLite + memory cache/queue；PostgreSQL advisory lock 支持应用定义的 session/transaction 互斥，SQLite 官方建议其作为应用本地存储且避免跨计算机共享数据库文件。USER-DECISION:2026-08-26；R-018；R-020。
- **选项：** 全 profile Redis；仅 PostgreSQL 多副本使用 Redis；Redis 零保留并由数据库与本地缓存承担职责。
- **推荐：** 用户选择 Redis 零保留。
- **结论：** 删除 Redis 依赖、adapter、配置、Compose service/volume 与文档；session、task definition 和可靠事件保存在当前数据库；只允许有界进程内缓存，缓存失效或丢失不得影响正确性；可靠异步采用 transactional outbox；Server SQLite 与 Desktop SQLite 限定单实例；Server PostgreSQL 可多副本，由独立 worker 使用固定 advisory lock 选出唯一 scheduler/outbox executor。
- **原因：** 数据库已经是三种 profile 的共同持久化边界，能够统一撤销、重试与恢复语义，并消除 Redis 对部署、桌面和测试矩阵的额外复杂度。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** outbox publish/claim 必须事务化、可重试且由 consumer 幂等；进程内缓存必须有界、可观测并可完全禁用；SQLite 不支持多副本；PostgreSQL worker 失锁/断连必须停止执行并可由其他实例接管。
- **后续：** D-023 删除 Redis 配置并固定 profile schema；D-017 建立 outbox recovery、leader handoff、缓存禁用与 Redis 零命中门禁。
- **替代/被替代：** 替代 Redis cache/queue/lock 以及 memory queue 作为可靠事件通道。

## LOG-020 — 2026-08-26T06:09:06Z — 不可变类型化配置与 Secret 输入边界
- **设计树节点：** D-023
- **轮次与依赖：** round 7 / D-005、D-012、D-013、D-021、D-022
- **状态：** confirmed
- **问题：** Server PostgreSQL、Server SQLite 与 Desktop SQLite 如何共享配置机制，同时避免宿主无关字段、全局状态和密钥泄露。
- **事实与来源：** 当前运行时代码直接读取全局 <Code>config.ApplicationConfig</Code> 等变量；<Path>cmd/config</Path> 可序列化并打印 JWT、数据库配置；示例 YAML 含开发 JWT secret、MySQL DSN、Redis 与多数据源字段；Desktop 的数据目录和 sidecar 启动材料应由 Tauri 宿主生命周期决定。USER-DECISION:2026-08-26；CODE:<Path>go-admin-plus/cmd/config/server.go</Path>；CODE:<Path>go-admin-plus/config/settings.yml</Path>；CODE:<Path>deploy/compose/settings.template.yml</Path>。
- **选项：** 不可变强类型配置且 secret 仅来自环境/secret file；强类型配置但允许 YAML 密钥；保留 Viper 全局配置。
- **推荐：** 选择不可变强类型配置与最小 profile schema，禁止持久配置和进程参数携带密钥。
- **结论：** 定义 <Code>server-postgres</Code>、<Code>server-sqlite</Code>、<Code>desktop-sqlite</Code> 三个最小强类型 schema；启动阶段按 <Code>defaults &lt; config file &lt; environment &lt; explicit non-secret CLI flags</Code> 合成并一次校验，再通过构造函数注入不可变值；secret 只允许环境变量或 <Code>_FILE</Code> reference，禁止 secret CLI flag、原始配置打印和运行时 reload；Desktop 数据/日志路径、随机 loopback 端口与一次性启动材料由 Tauri 宿主提供；删除 JWT、Redis、租户、多数据源、MySQL 与 SQL Server 配置键。
- **原因：** profile schema 在编译与启动边界暴露真实能力，避免 Desktop 继承服务端字段，并消除全局配置和诊断命令泄露密钥的路径。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 启动失败必须指出字段路径与规则但对值脱敏；配置对象不得导出 setter 或全局 singleton；CLI 只接受 profile/config path、bind 等非敏感显式覆盖；容器 secret file 在读取后只进入内存，不生成合并后的明文配置文件。
- **后续：** D-016 依据 profile 边界安排切换；D-017 建立 schema、precedence、redaction、零遗留键与 Desktop 宿主注入测试。
- **替代/被替代：** 替代 Viper/global mutable config、多份全量 settings YAML、<Code>cmd config</Code> 原始打印和旧 JWT/Redis/tenant/multi-database 配置。

## LOG-021 — 2026-08-26T06:19:24Z — 垂直切片施工与最终原子切换
- **设计树节点：** D-016
- **轮次与依赖：** round 8 / D-004、D-009、D-010、D-011、D-012、D-013、D-014、D-015、D-019、D-020、D-021、D-022、D-023
- **状态：** confirmed
- **问题：** 在最终零兼容的前提下，如何组织全仓施工，使每一步可验证又不形成长期双轨。
- **事实与来源：** 本次 Greenfield 重构允许删除旧 API、数据、schema、配置与行为；目标同时跨越 Go 模块、pnpm workspace、Tauri 2、双方言 migration、OpenAPI 合同和交付矩阵，一次性全仓修改会推迟关键反馈。USER-DECISION:2026-08-26；D-004；D-006 至 D-015；D-019 至 D-023。
- **选项：** 新结构内垂直切片并最终原子切换；一次性全仓修改后首次验证；长期双轨加兼容层。
- **推荐：** 选择可验证垂直切片施工，并以单一原子切换作为旧结构删除边界。
- **结论：** 依次完成合同与新骨架、IAM 端到端切片、其余业务模块、共享 Web/Desktop、部署发行；每个切片必须在新结构内闭合并立即通过对应门禁。所有目标能力完成后一次切换根任务、入口、合同与交付路径，并在同一变更中删除旧目录、旧 API、旧 schema、旧配置、旧命名和临时迁移代码。开发期产物不形成第二套可发布产品，不维护数据/API 双写或兼容 adapter。
- **原因：** 垂直切片提供短反馈和可定位证据，原子切换则保证最终仓库只有一套权威结构与合同。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 新切片只能依赖目标架构；旧代码不得反向依赖新模块；不得发布混合结构版本；临时工具必须有删除 Ticket 和零命中验收；原子切换前旧产品仅作为能力清单参考，不作为兼容基线。
- **后续：** D-017 固定切片级、集成级和原子切换级质量证据；D-018 最终复核后进入 Spec。
- **替代/被替代：** 替代大爆炸式首次验证与长期 branch-by-abstraction 双轨；不改变最终 Greenfield 零兼容目标。

## LOG-022 — 2026-08-26T06:28:14Z — 风险分层强制质量门禁
- **设计树节点：** D-017
- **轮次与依赖：** round 9 / D-009、D-011、D-012、D-013、D-015、D-016、D-019、D-020、D-021、D-022、D-023
- **状态：** confirmed
- **问题：** 如何以自动证据证明重构保持目标功能、架构边界、双方言、双前端宿主和跨平台发行完整，而不把每个 PR 都变成完整发行流水线。
- **事实与来源：** 目标矩阵包括 Go 模块边界、pnpm workspace、OpenAPI 生成、PostgreSQL/SQLite、Web/Tauri 2、Linux 双架构 OCI、macOS Universal DMG 与 Windows x64 NSIS；这些风险的反馈成本不同，需要在最早有效阶段阻断。USER-DECISION:2026-08-26；D-009；D-011 至 D-016；D-019 至 D-023。
- **选项：** 风险分层强制门禁；每个 PR 运行完整跨平台发行；仅 lint/unit/build 加人工验收。
- **推荐：** 采用本地/PR/受保护发行三层强制证据，失败阻断对应阶段。
- **结论：** 本地与 Hook 执行 format、静态检查、生成漂移、secret 和遗留技术零命中；PR 执行 Go/pnpm build/test、Go import 方向、workspace cycle/deep-import、OpenAPI lint/bundle/generate/conformance、PostgreSQL/SQLite migration/repository、session/CSRF/RBAC/outbox、Web E2E 和 Tauri sidecar tracer；受保护发行在原生 runner 执行 Linux amd64/arm64 OCI、macOS Universal DMG、Windows x64 NSIS 的构建、安装、smoke test、签名/公证、SBOM 与 provenance。任一必需门禁失败禁止对应合并或发行。
- **原因：** 风险越靠近源码越早反馈，依赖原生 OS 或发布凭据的证据留在受保护发行，同时不削弱最终交付合同。
- **影响工件：** CONTEXT / ADR / Spec / Ticket / Goal Plan
- **约束或不变量：** 根 Taskfile 是全部门禁的本地可复现入口；CI YAML 只编排；不以总体 coverage 数字替代关键路径测试；flaky test 必须隔离并绑定修复期限，不得静默 retry 成绿色；豁免必须有 owner、理由、范围、到期时间和追踪 Ticket。
- **后续：** D-018 复核完整设计共识；进入 Spec 后把每层门禁拆成可验收 Requirement 与 Scenario。
- **替代/被替代：** 替代仅编译验收、手工桌面/发行验收和每 PR 完整签名发行矩阵。

## LOG-023 — 2026-08-26T06:34:06Z — 最终设计共识与 Spec 路由
- **设计树节点：** D-018
- **轮次与依赖：** round 10 / D-016、D-017
- **状态：** confirmed
- **问题：** 目标目录、模块合同、数据库与宿主矩阵、安全模型、施工顺序、质量门禁和非目标是否已经完整并可进入正式 Spec。
- **事实与来源：** D-001 至 D-023 的全部适用分支均已有回答，所有节点依赖已关闭；用户逐项复核摘要后明确选择确认共识。USER-DECISION:2026-08-26；DESIGN-TREE:<Path>speculo/.speculo/specdev/changes/2026-08-26-project-architecture-reconstruction/design-tree.json</Path>。
- **选项：** 确认共识并路由到 Spec；指出遗漏并重新打开设计树。
- **推荐：** 关闭 Grill，将已确认设计综合为外部行为、范围与验收合同明确的正式 Spec。
- **结论：** 设计树状态设为 <Code>consensus</Code>；<Code>specdev/grill-with-docs</Code> 成功完成；下一 Work 为 <Path>speculo/workflows/specdev/S-spec/S-spec.md</Path>，本轮不自动执行下一 Work，也不授权产品实现。
- **原因：** 高影响架构、安全、数据、兼容、交付和验收分支均已显式决定，继续在 Grill 中扩展会把 Spec/Ticket 职责混入设计访谈。
- **影响工件：** Design Tree / LOG / Spec / Ticket / Goal Plan
- **约束或不变量：** 后续 Spec 必须综合全部已接受 ADR 与 CONTEXT，不得重新引入已删除的兼容、租户、Redis、JWT、Casbin、Wails、MySQL 或 SQL Server；若发现新的高影响外部决策，返回 Grill 新增节点而不是静默假设。
- **后续：** 用户激活 <Path>speculo/workflows/specdev/S-spec/S-spec.md</Path> 后编写 <Path>speculo/.speculo/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>。
- **替代/被替代：** 关闭本 change 的设计访谈阶段；不替代任何已接受架构 ADR。
