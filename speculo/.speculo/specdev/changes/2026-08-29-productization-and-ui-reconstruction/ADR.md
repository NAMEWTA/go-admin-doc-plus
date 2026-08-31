# 产品化与 UI 重构架构决策

当前 change 继承 `<Path>{roots.state}/specdev/adr/</Path>` 中已经提升的永久架构决策。以下条目只记录本 change 新增并已由用户确认的高影响取舍。

## ADR-001: 首管理员使用宿主适配的一次性 Bootstrap

**Status:** accepted
**Source:** LOG-007
**Supersedes:** none

### Context
空库没有账号，而管理 API 要求既有管理员授权。自动建立默认弱管理员会形成公开凭据，全部要求 CLI 又会破坏 Desktop 首次使用体验。

### Decision
IAM 提供只在账号为空时成功的一次性 Bootstrap 用例。Server 通过离线产品 CLI 调用；Desktop 通过 Tauri 首次设置流程调用同一用例。固定密码 SQL和未认证 Web 初始化 API 不存在。

### Trade-off
需要维护 CLI 与 Desktop 两个宿主 adapter，但身份创建规则仍只有一份，并避免公开初始化入口。

### Consequences
凭据不能进入 argv、日志、审计 payload 或源码。并发 Bootstrap 只有一个成功，结果必须产生脱敏审计事实。

### Verification / Migration
删除固定凭据草案；在 SQLite/PostgreSQL 验证空库、重复、并发、已有账号和敏感值泄漏场景。

## ADR-002: 账号删除采用 Tombstone 与跨模块生命周期

**Status:** accepted
**Source:** LOG-008
**Supersedes:** none

### Context
IAM 直接硬删除账号会让 Files owner 和 Audit actor 失去可追踪身份，跨模块数据库级联又破坏模块所有权。

### Decision
删除账号先进入不可登录 Tombstone 并撤销全部 Session，保留稳定审计引用。Files 等消费者通过 Integration Event 完成转移或清理，随后才允许净化身份信息。

### Trade-off
删除成为可观察异步流程，UI 与运维需要展示 pending/failed 状态，但不产生孤儿资源或跨模块事务。

### Consequences
IAM 不直接查询或修改 Files 表。消费者必须幂等，失败可重试，系统管理员最小存续不变量必须在进入 Tombstone 前验证。

### Verification / Migration
双方言增加生命周期字段和事件状态；覆盖撤销、重试、崩溃恢复、审计引用和最终净化。

## ADR-003: 登录限流使用数据库持久双桶

**Status:** accepted
**Source:** LOG-009
**Supersedes:** none

### Context
进程内 Argon2 工作预算不能阻止串行、多进程或多实例撞库；Redis 又不属于产品运行矩阵。

### Decision
登录前由当前正式数据库原子维护账号桶和来源桶，两个维度都必须通过。阈值、观察窗和退避时长只能在产品规定安全范围内配置，保护不能关闭。

### Trade-off
登录路径增加数据库写入和双方言并发实现，但限流可跨重启、多进程和 PostgreSQL 多实例保持一致。

### Consequences
不存在账号执行等价密码工作；审计使用可聚合且脱敏的目标与来源维度；过期桶由有界 worker 清理。

### Verification / Migration
覆盖双方言原子竞争、时间窗、退避、重启、多实例、代理来源和账号枚举时序。

## ADR-004: IAM 提供五种组织数据范围

**Status:** accepted
**Source:** LOG-010
**Supersedes:** none

### Context
all/self 无法表达成熟后台脚手架常见的本组织、组织树和自定义组织授权，后补会再次改变角色合同。

### Decision
数据范围稳定枚举为 all、self、organization、organization-tree 和 custom。Organization 通过消费者 Port 为 IAM 提供最小组织投影，无租户语义。

### Trade-off
授权查询、角色编辑和测试矩阵显著扩大，但一次形成完整数据权限能力，避免临时 SQL 过滤和二次契约重构。

### Consequences
前端可见性不构成授权；每个后端用例必须把 Permission Code 与有效数据范围同时应用到查询或命令目标。

### Verification / Migration
双方言、组织树变化、自定义集合、越权直接 API 和多角色组合必须有合同测试。

## ADR-005: Server 运维能力收敛为单一产品 CLI

**Status:** accepted
**Source:** LOG-011
**Supersedes:** none

### Context
多个 Server 小二进制让配置、版本、错误输出和运维发现方式漂移，但 Desktop sidecar 有独立 Tauri 安全边界。

### Decision
单一 go-admin-plus 二进制提供 serve、worker、migrate、bootstrap、doctor 和 version 子命令。Desktop sidecar 保持独立；根 Taskfile 继续拥有跨语言产品命令面。

### Trade-off
需要统一命令解析和内部 command application 层，但发行、文档和运维入口更清晰。

### Consequences
旧 Server 小命令在原子切换后删除，不提供 alias。所有子命令复用类型化配置、脱敏错误和统一 logger。

### Verification / Migration
建立命令 contract test、错误/secret 脱敏、版本输出和根 Task 映射验证。

## ADR-006: 认证读取只读且 Session 显式续期

**Status:** accepted
**Source:** LOG-012
**Supersedes:** none

### Context
每个认证请求更新 last_seen、idle expiry 和 CSRF 会造成数据库写热点、锁竞争和跨浏览器标签互相失效。

### Decision
普通认证 GET 只读取并验证数据库 Session。Session 活跃期只由受 CSRF 保护的 heartbeat、业务写请求和提前 renew 更新；CSRF 在同一认证 Session 家族内保持稳定。

### Trade-off
客户端需要 heartbeat/renew 协议和跨标签协调，但读取路径稳定、CSRF 不再滚动失效，续期行为可以独立观察和测试。

### Consequences
Cookie Session 仍由数据库实时验证；BroadcastChannel 只同步客户端状态，不承担授权正确性。退出、撤销和绝对超时不因 heartbeat 延后。

### Verification / Migration
覆盖双标签并发、GET 零写入、heartbeat/renew 竞争、轮换、退出、撤销、空闲/绝对超时和旧 CSRF 拒绝。

## ADR-007: 管理员恢复是严格离线运维流程

**Status:** accepted
**Source:** LOG-013
**Supersedes:** none

### Context
一次性 Bootstrap 不能解决系统已有账号但全部管理员不可用的灾难恢复；重开 Bootstrap 或提供远程恢复接口会形成提权旁路。

### Decision
产品 CLI 提供 recover-admin，但只允许在 API/worker 停止、操作者持有直接数据库访问权、通过 TTY/secret file 提供新强密码并填写审计原因时执行。Bootstrap 仍只适用于空库。

### Trade-off
恢复需要停机和数据库权限，不如远程操作方便，但其授权边界与数据库灾难恢复一致且不暴露网络攻击面。

### Consequences
恢复结果必须撤销相关旧 Session 并写入脱敏审计事实；运行进程或并发恢复必须被拒绝。

### Verification / Migration
覆盖运行中拒绝、数据库锁、弱密码、审计原因、Session 撤销、敏感值脱敏和双方言。

## ADR-008: Desktop Bootstrap 成功后建立本地 Session

**Status:** accepted
**Source:** LOG-014
**Supersedes:** none

### Context
Desktop 首次设置已经验证并提交强密码，再返回登录页重复输入没有增加实质安全边界。

### Decision
Desktop 在同一受控本地首次设置流程中创建首管理员和首个 Session，成功后直接进入工作区。Session 由 Tauri 安全宿主保存，密码和原始 Session 不进入 WebView 持久状态。

### Trade-off
首次设置需要跨 Bootstrap 与 Session 两个用例编排，但减少重复操作并保持凭据隔离。

### Consequences
UI 必须区分账号创建失败、Session 创建失败和工作区启动失败，不得把部分成功伪装成空库。

### Verification / Migration
验证首次设置、Session 保存、WebView secret 扫描、崩溃恢复和重启后的正常认证。

## ADR-009: 账号删除必须显式选择文件处置

**Status:** accepted
**Source:** LOG-015
**Supersedes:** none

### Context
自动转移会改变数据所有权，自动删除会造成不可恢复损失，只阻止删除又无法形成完整账号生命周期。

### Decision
删除拥有文件的账号必须显式选择 transfer 或 purge。transfer 指定有效目标账号；purge 通过 Transactional Outbox 驱动幂等物理删除。未选择或参数无效时不进入 deletion-pending。

### Trade-off
删除交互多一步并需要展示异步进度，但每次所有权或数据损失都是明确决定。

### Consequences
删除命令、事件与消费者必须携带稳定策略和业务键；失败保持可重试，不得静默切换策略。

### Verification / Migration
覆盖无策略、无效目标、自转移、目标删除中、purge 重试、崩溃恢复和最终净化。

## ADR-010: 同一权限的多角色数据范围取并集

**Status:** accepted
**Source:** LOG-016
**Supersedes:** none

### Context
多个角色可能同时授予同一操作但具有不同数据范围；若合成语义不固定，各模块会产生不同可见集合。

### Decision
只收集实际授予目标 Permission Code 的角色，并把其有效数据范围取并集。当前 RBAC 不支持显式 deny。

### Trade-off
并集比交集更宽，但与角色权限累加语义一致；需要确保无权角色不会参与范围扩张。

### Consequences
scope 计算必须集中、可解释并由后端应用；角色与组织变更后下一请求即时使用新范围。

### Verification / Migration
覆盖不同范围顺序、重复角色、无权角色、禁用角色、组织变化和双方言查询等价性。

## ADR-011: 登录限流只暴露模糊重试反馈

**Status:** accepted
**Source:** LOG-017
**Supersedes:** none

### Context
显示剩余次数或账号锁定状态会帮助攻击者确认账号存在；完全不给重试时间又会伤害合法用户体验。

### Decision
普通失败统一返回凭据错误。触发限流时使用稳定限流错误和粗粒度 Retry-After，不显示剩余次数、具体桶、精确内部截止时间或账号存在性。

### Trade-off
合法用户无法看到精确计数，但可知道大致等待时间；审计仍保留服务端聚合信息。

### Consequences
页面、OpenAPI、HTTP header 和日志必须使用相同脱敏语义；不存在账号与错误密码走相同外部路径。

### Verification / Migration
覆盖账号存在/不存在、账号桶/来源桶、边界时间、Retry-After 分桶和响应等价性。

## ADR-012: 运行角色按数据库 Profile 组合

**Status:** accepted
**Source:** LOG-018
**Supersedes:** none

### Context
PostgreSQL 需要独立扩展 API 并保持唯一 worker，SQLite/Desktop 则需要单实例和轻量启动体验。

### Decision
PostgreSQL 生产的 serve 只运行 API，worker 作为独立子命令部署。Server SQLite 和开发环境允许 serve --with-worker 一体运行；Desktop sidecar 始终一体运行本地 API 与 worker。

### Trade-off
Profile 启动矩阵更明确但测试组合增加；换取生产扩展性和轻量环境易用性。

### Consequences
readiness 必须反映当前角色而非假设所有组件同进程；PostgreSQL worker 继续通过数据库协调唯一 active executor。

### Verification / Migration
覆盖 PostgreSQL 多 API/多 worker 接管、SQLite 单实例、一体启动、角色 shutdown 和 readiness。

## ADR-013: 管理员恢复只能修改既有非删除账号

**Status:** accepted
**Source:** LOG-019
**Supersedes:** none

### Context
离线恢复若能创建新账号或复活已删除账号，会绕过正常账号创建、删除和文件生命周期。

### Decision
recover-admin 必须显式选择一个既有非删除账号，只能重置其密码、重新启用、授予系统管理员角色并撤销旧 Session。没有可恢复账号时只能恢复数据库备份。

### Trade-off
极端情况下恢复步骤更重，但不会创造旁路身份或破坏 Tombstone 不变量。

### Consequences
恢复命令需要锁定目标账号和管理员存续状态；操作结果必须可审计且不能回显凭据。

### Verification / Migration
覆盖不存在、disabled、普通、管理员、Tombstone 账号及并发恢复和 Session 撤销。

## ADR-014: 文件 Purge 开始执行后不可恢复

**Status:** accepted
**Source:** LOG-021
**Supersedes:** none

### Context
隐藏回收站会改变配额、磁盘治理和用户对 purge 的理解，永久逻辑删除又无法真正释放内容。

### Decision
purge 经二次确认进入可靠异步流程，在 worker claim 前允许取消；开始物理删除后不可取消或恢复，不提供隐藏回收站。

### Trade-off
错误确认可能造成永久损失，因此必须用明确交互、权限和审计换取简单且真实的删除语义。

### Consequences
删除状态机必须暴露 queued、claimed、completed、failed；claim 是取消能力的稳定边界。

### Verification / Migration
覆盖确认、取消竞态、claim 后冲突、重复消费、部分失败、磁盘内容和 metadata 最终一致性。

## ADR-015: Migration 执行策略按数据库 Profile 区分

**Status:** accepted
**Source:** LOG-022
**Supersedes:** none

### Context
PostgreSQL 生产可多副本运行，服务启动自动迁移会产生部署竞态；SQLite/Desktop 则需要自包含启动与本地恢复体验。

### Decision
PostgreSQL 生产必须先显式运行 migrate，serve/worker 只校验 schema 并在版本不匹配时拒绝 ready。Server SQLite 启动时自动迁移；Desktop 仅在备份成功后自动迁移。

### Trade-off
PostgreSQL 部署增加一个显式步骤，但获得确定迁移所有权；SQLite 与 Desktop 继续保留单实例便利性。

### Consequences
doctor、部署定义和文档必须显示 schema 状态；进程不能以部分或未知 schema 对外服务。

### Verification / Migration
覆盖过旧、当前、过新、迁移失败、多副本启动、SQLite 幂等和 Desktop 备份恢复。

## ADR-016: 账号拥有一个可空主部门和多个岗位

**Status:** accepted
**Source:** LOG-023
**Supersedes:** none

### Context
五种数据范围需要稳定账号组织锚点；多部门成员关系会让 organization 语义和角色范围合成显著复杂化，只从岗位推导又使岗位变更隐式改变数据权限。

### Decision
账号至多拥有一个主部门并可分配多个岗位。organization 与 organization-tree 从主部门计算；custom 由角色保存部门集合；岗位不决定主部门。

### Trade-off
无法直接表达一个账号平等隶属多个部门，但模型与管理脚手架认知一致，授权解释清晰。

### Consequences
无主部门账号不能从 organization 类范围获得数据；岗位分配必须经过部门一致性校验；部门删除需要处理账号和岗位引用。

### Verification / Migration
覆盖主部门为空/变更/删除、岗位跨部门、自定义集合、部门树移动和多角色并集。
