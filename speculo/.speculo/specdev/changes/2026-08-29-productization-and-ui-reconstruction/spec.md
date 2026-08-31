---
schema_version: 3
artifact: spec
change: 2026-08-29-productization-and-ui-reconstruction
status: ready
ready_for_tickets: true
sources:
  - USER-DECISION:三轮 Grill 决策全部接受并明确进入 S-spec
  - DESIGN-TREE:<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/design-tree.json</Path>
  - ADR:<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ADR.md</Path>
  - CONTEXT:<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/CONTEXT.md</Path>
  - PLAN:<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/PLAN.md</Path>
  - CODE:contracts/openapi/product.yaml
  - CODE:go-admin-plus/internal/app/product
  - CODE:go-admin-plus-ui/packages/app-shell
  - RESEARCH:https://pkg.go.dev/vuln/GO-2026-6112
  - RESEARCH:https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
  - RESEARCH:https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
  - RESEARCH:https://router.vuejs.org/guide/
  - RESEARCH:https://docs.github.com/en/actions/tutorials/use-containerized-services/create-postgresql-service-containers
  - RESEARCH:https://v2.tauri.app/security/capabilities/
---

# Spec: Go Admin Plus 产品化与 UI 重构

- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **当前 ADR：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ADR.md</Path>`
- **当前领域上下文：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/CONTEXT.md</Path>`

## 1. 问题与目标

### 问题陈述

Go Admin Plus 已具备模块化单体、OpenAPI、SQLite/PostgreSQL 双数据库、pnpm workspace、Tauri 2 sidecar、可靠 outbox 和根级治理基础，但当前产品闭环仍存在发布阻断：空库无法建立首管理员；Session 每次请求同时滚动 CSRF，导致多标签页相互失效并让读取产生数据库写入；登录保护不能跨重启、进程和实例生效；二级 URL 与页面标签各自维护状态；账号硬删除会留下跨模块资源生命周期缺口；文件容量与磁盘水位无完整治理；日志配置未形成真实可观测链路；CI 未证明真实 PostgreSQL、浏览器与原生 Desktop 流程。

与此同时，当前前端视觉、布局和交互未达到成熟管理脚手架的可用性标准。此次变更必须在不重复搬迁目录、不复制参考项目弱安全方案的前提下，把现有工程底座补成可从空库初始化、可管理、可恢复、可观察、可验证的完整产品，并为 Web 与 Desktop 提供同一业务语义下的现代后台工作台。

### 目标用户与场景

- 自托管 Server 操作者：以 SQLite 或 PostgreSQL 部署、迁移、初始化、诊断、恢复和运行 API/worker。
- Desktop 单机用户：仅使用 SQLite，在原生首次设置流程中建立管理员并直接进入工作区。
- 系统管理员：管理账号、角色、权限、菜单、组织、岗位、参数、字典、任务、文件与代码生成能力。
- 普通业务用户：在明确的数据范围内使用被授权功能，并在多个浏览器标签中保持稳定 Session。
- 开发者与维护者：通过统一命令、生成器、双方言测试、真实 E2E 与安全门禁演进项目。

### 成功标准

- Server SQLite、Server PostgreSQL 与 Desktop SQLite 都能从空库形成唯一首管理员并完成登录，仓库、SQL、日志和命令历史中不存在固定管理员密码。
- Session 的认证读取保持只读，稳定 CSRF 不因另一个标签页请求而失效，退出、撤销、权限变化、空闲超时和绝对超时仍由服务端数据库即时判定。
- 登录限流由正式数据库持久保存，账号与来源双桶在重启、多进程和 PostgreSQL 多实例下保持一致，外部响应不泄漏账号存在性。
- 任一已编译二级路由的直接访问、刷新、前进后退、菜单、访问标签、面包屑、标题和页面内容保持一致；未授权与不存在路由分别形成稳定 403/404 行为。
- 账号删除不产生失联文件或断裂审计引用；转移与永久清理均可观察、可重试且遵守明确不可逆边界。
- 五种数据范围在 SQLite/PostgreSQL 上产生等价授权结果，并且任何前端可见性都不能替代后端授权。
- Web/Desktop 共享业务组合，在目标视口内无内容重叠、不可操作控件或无意横向滚动，并满足键盘、焦点和 reduced-motion 合同。
- PostgreSQL、浏览器、Tauri、生成漂移和漏洞扫描门禁不能通过 Skip、缺环境或未执行伪装为成功。

### 非目标

- 不再次重排根目录或重写当前模块化单体、pnpm workspace、OpenAPI 与 Tauri 2 基础架构。
- 不兼容旧命令、旧页面标签模式、旧账号硬删除接口、固定初始化 SQL 或旧 Session 滚动行为。
- 不引入租户、Redis、JWT/localStorage bearer、Casbin URL 策略、AutoMigrate 或 Java 风格 common/system 聚合层。
- 不要求个人自用发行物签名、公证、商店发布或自动更新。
- 不把参考的 Backplane、RuoYi 或 Plus UI 代码、目录、依赖栈当作兼容目标。

## 2. 解决方案与外部行为

### 解决方案摘要

保持当前后端模块边界和前端 workspace 边界，以九个相互衔接的产品垂直面完成重构：依赖安全基线；首管理员与离线恢复；Session/CSRF 与登录保护；Vue Router 单一事实源；Element Plus 设计系统与页面重做；组织数据范围；账号/文件/配额生命周期；统一 CLI、日志、doctor 与生成器；真实 PostgreSQL、浏览器和原生 Desktop 交付门禁。

UI 重构复用当前生成 transport、无头领域、Web Domain 与 Web/Desktop 产品组合。现有业务操作的路径和 DTO 不因视觉重构改名；只有 Session 生命周期、账号删除生命周期、组织数据范围和文件治理所需的新合同按本规格原子变更。

### 主要流程

#### 首次安装与管理员恢复

1. `go-admin-plus migrate` 在 PostgreSQL 生产部署中显式应用 forward-only migration；Server SQLite 启动自动迁移；Desktop 在可恢复备份成功后自动迁移。
2. Server 操作者在 API/worker 未运行时执行 `go-admin-plus bootstrap`，用户名可作为普通非敏感输入，新密码只能通过交互式 TTY 或权限受限 secret file 提供。
3. IAM Bootstrap 在一个事务中验证账号表为空、创建账号、授予受保护系统管理员角色并写入脱敏审计事实。两个并发调用最多一个成功；已有任何账号时永远拒绝。
4. Desktop 检测空账号状态后显示原生首次设置流程，调用同一 IAM Bootstrap 用例并建立首个 Session；成功后直接进入工作区。
5. Desktop 已创建账号但 Session 创建失败时保留已提交账号，明确显示可恢复错误并进入普通登录，不得再次展示空库 Bootstrap。
6. 当系统已有账号但管理员访问丢失时，操作者停止 API/worker 后执行 `go-admin-plus recover-admin`，显式选择既有非删除账号、提供强密码和审计原因。命令重新启用账号、授予系统管理员角色并撤销该账号全部旧 Session；没有可恢复账号时只允许从备份恢复。

#### 登录、Session 与多标签页

1. 登录请求先原子检查账号标识桶和可信来源桶，两者都允许后再完成等价 Argon2 工作；不存在账号使用 dummy hash。
2. 普通失败返回同一凭据错误；触发任一限流桶时返回稳定限流 Problem 与粗粒度 `Retry-After`，不返回剩余次数、桶类型、准确内部截止时间或账号存在性。
3. 登录成功创建不透明 Cookie Session 与该 Session 家族稳定 CSRF。`GET /iam/session/current` 返回当前身份、权限和同一 CSRF，但不得更新 Session 行或轮换 CSRF。
4. 客户端通过受 CSRF 保护的 `POST /iam/session/heartbeat` 在实际活跃时延长 idle expiry；通过 `POST /iam/session/renew` 在进入续期窗口时轮换不透明 Session token。受保护业务写请求也可延长 idle expiry。
5. 同一 Session 家族的 CSRF 在 renew 后继续有效；退出、服务端撤销、密码修改、账号禁用/删除、绝对超时会使 Session 立即失效。
6. Browser adapter 使用 `BroadcastChannel` 同步登录、renew、退出和撤销提示；频道不可作为认证或授权事实源，任一请求仍以数据库 Session 与实时权限为准。

#### 路由、菜单与工作台

1. 产品 route manifest 是前端唯一编译期路由事实，至少声明稳定 name、path、permission code、menu key、title、Lucide icon、order 与 component loader。
2. Web 使用 HTML5 history，Desktop 使用 hash history；IAM、Settings、Scheduler 等原多标签大页面拆为真实子路由。
3. 侧栏、移动抽屉、访问标签、面包屑、文档标题、刷新恢复和前进后退全部从当前 route 派生，页面不得再维护平行的默认 tab 状态。
4. 后端 capability/menu grants 只与编译期 route manifest 取交集；数据库值不得构造任意 component import 或绕过编译期可达路由。
5. 缺少权限进入稳定 403 页面；不存在或未编译路径进入稳定 404 页面；权限被撤销后下一次授权读取/导航即更新可达集合。
6. UI 采用中性浅色工作面、炭黑侧栏、克制绿色品牌强调、紧凑密度与可持久化暗色模式；不使用装饰性渐变、营销式大卡片、页面手写基础控件或重复 Web/Desktop 页面。

#### 数据范围、账号删除与文件治理

1. 账号最多关联一个可空主部门并可分配多个岗位；岗位不隐式决定主部门。角色对 Permission Code 可声明 all、self、organization、organization-tree 或 custom 数据范围。
2. 计算某 Permission Code 时只收集实际授予该权限且仍启用角色的范围并取并集；当前模型无显式 deny。organization 与 organization-tree 从账号主部门计算，custom 保存角色对应部门集合。
3. 账号从 active/disabled 进入 deletion-pending 前，系统验证删除后仍至少存在一个启用系统管理员，并要求显式文件策略。账号进入 pending 后立即禁止登录并撤销全部 Session。
4. 账号删除公共合同替换为单账号命令 `POST /iam/administration/users/{userId}/deletion`，请求必须选择 `transfer` 并给出有效目标账号，或选择 `purge` 并完成二次确认。原账号 `DELETE` 与用户 batch-delete 合同删除，不提供兼容层。
5. `GET /iam/administration/users/{userId}/deletion` 返回 queued、claimed、completed 或 failed 状态；`POST /iam/administration/users/{userId}/deletion/cancel` 只在 worker claim 前成功。
6. IAM 写出带稳定业务键和策略的 Integration Event。Files 以幂等消费者转移 owner 或物理 purge；IAM 不读取或修改 Files 私有表。消费者完成后账号才进入 deleted，并净化个人身份字段但保留稳定审计引用。
7. 上传同时受单文件、单账号字节、单账号对象数量、全局容量、磁盘最小剩余字节和比例约束。服务先原子预留配额，再 stage/publish；失败或崩溃由 reconciliation worker 回收。
8. 低磁盘水位拒绝新上传并返回稳定容量 Problem，但下载与删除继续可用；配额与磁盘检查不得泄漏其他账号的使用详情。

#### 运维、日志与交付

1. 单一 Server 二进制 `go-admin-plus` 暴露 `serve`、`worker`、`migrate`、`bootstrap`、`recover-admin`、`doctor` 和 `version`；旧 Server 小命令原子删除，不提供 alias。Desktop sidecar 继续独立。
2. PostgreSQL 生产 `serve` 只承载 API，`worker` 独立部署；Server SQLite 与开发 profile 可使用 `serve --with-worker`；Desktop 始终组合本地 API 与 worker。
3. PostgreSQL `serve/worker` 只校验 schema，过旧、过新或未知时不进入 ready 并以脱敏错误退出；Server SQLite 自动迁移；Desktop 先成功备份再自动迁移，备份失败则不修改数据库。
4. `doctor` 只读检查配置、secret reference、数据库连接、schema、Bootstrap 状态、文件根、磁盘水位、worker 协调和版本兼容性，以机器可判定退出码报告结果且不输出 secret。
5. Server 生产日志为结构化 JSON stdout，Desktop 为受控轮转文件，开发可读 console；`log.level` 必须实际过滤。运行日志与 Audit 事实分责，均禁止密码、Session token、CSRF、完整 DSN 和请求/响应正文。
6. `kin-openapi` 升级到不低于 v0.144.0 的兼容版本并通过可达漏洞检查。
7. required PostgreSQL job 提供真实健康检查 service 与一次性 DSN；required browser/native/security job 缺环境、Skip 或未执行时失败。个人发行签名与公证明确为 not-required。

### 边界、失败与稳定错误行为

- Bootstrap 对非空账号库返回稳定 conflict，不透露已有账号详情；弱密码、非 TTY 输入、secret file 权限过宽和并发失败均不得产生部分账号或明文审计。
- recover-admin 在运行角色未停、目标不存在、目标已删除、密码不合格、原因缺失或数据库无法获得排他运维锁时失败，不创建救援账号、不复活 Tombstone。
- Session `current`、heartbeat 和 renew 对无效/过期/撤销 Session 使用同一认证失败语义；CSRF 错误与权限不足可审计但不回显 token。
- 删除最后一个启用系统管理员、转移给自己、转移给删除中/已删除账号、缺少处置策略或 purge 未二次确认时，账号状态保持不变。
- purge 在 queued 阶段可取消；claim 之后取消返回 conflict。重复事件、worker 崩溃和部分物理删除进入可重试 failed/claimed 流程，不静默切换为 transfer。
- 无主部门账号从 organization、organization-tree 范围得到空集合；custom 空集合也得到空集合；all/self 不因此改变。
- PostgreSQL required test 没有 DSN、没有运行任何目标测试或出现 Skip 时 job 失败；E2E runner 不接受 opt-in 缺失即成功。
- UI 请求错误使用后端 Problem 的稳定 error code 映射为用户可行动反馈，未知错误显示通用失败并保留 trace ID，不显示服务端内部细节。

### 状态转换与不变量

```text
System setup:  empty -> bootstrapping -> initialized
                              | conflict/rollback -> empty

Account:       active <-> disabled -> deletion-pending -> deleted
                                      | cancel before claim -> disabled

Deletion:      queued -> claimed -> completed
                  | cancel          | retryable failure -> failed -> claimed

Session:       active -> renewed -> revoked/idle-expired/absolute-expired
```

- 一个数据库生命周期只允许一次 empty 到 initialized；recover-admin 永远不是 Bootstrap。
- 任何已初始化系统始终至少存在一个 enabled system administrator；不能用用户名字符串表达该不变量。
- deleted 账号不可登录、不可恢复、不可成为文件转移目标，其稳定匿名审计引用永久保留。
- 一个账号最多一个主部门；岗位可以多个，但岗位不得改变主部门或直接扩大数据范围。
- 同一 Permission Code 的有效多角色范围只取并集；没有授予该权限的角色不参与计算。
- IAM 与 Files 不共享私有表写权限；跨模块最终一致性只通过 versioned integration event/outbox 建立。
- 认证 GET 不更新 Session 或登录限流数据；CSRF 只在新 Session 家族建立时生成。
- SQLite 只允许单运行实例；PostgreSQL 可多 API、多 worker，但 scheduler/outbox 的一次执行语义继续由数据库协调。

## 3. 用户故事

- **US-001**：作为 Server 操作者，我希望从 SQLite 或 PostgreSQL 空库安全建立唯一首管理员，以便首次安装可实际登录且没有默认凭据。
- **US-002**：作为 Desktop 用户，我希望在原生首次设置中建立管理员并直接进入工作区，以便无需命令行或重复输入密码。
- **US-003**：作为灾难恢复操作者，我希望离线恢复一个既有非删除账号的管理员访问，以便不开放远程提权旁路。
- **US-004**：作为用户，我希望登录保护跨重启和实例持续生效且不泄漏账号是否存在，以便抵抗串行撞库与枚举。
- **US-005**：作为多标签页用户，我希望一个标签的读取或续期不使另一个标签 CSRF 失效，以便稳定完成并行工作。
- **US-006**：作为系统管理员，我希望退出、撤销、禁用、删除和权限变更即时由服务端生效，以便客户端状态不能绕过安全事实。
- **US-007**：作为后台用户，我希望二级 URL、菜单、访问标签、面包屑、标题和页面内容始终一致，以便刷新和浏览器历史可预测。
- **US-008**：作为 Web 或 Desktop 用户，我希望使用一致、紧凑、响应式且可访问的管理工作台，以便在不同宿主和目标尺寸高效操作。
- **US-009**：作为权限管理员，我希望以五种明确数据范围配置角色并得到可解释的多角色合成，以便按组织边界授权业务数据。
- **US-010**：作为组织管理员，我希望账号拥有一个明确主部门和多个岗位，以便组织范围与人员岗位都可稳定管理。
- **US-011**：作为账号管理员，我希望删除账号前显式决定文件转移或永久清理，并看到执行状态，以便不会形成孤儿资源或意外数据损失。
- **US-012**：作为文件使用者和操作者，我希望上传受账号、全局和磁盘水位约束且失败可恢复，以便存储不会被无界耗尽。
- **US-013**：作为部署操作者，我希望一个产品 CLI 提供明确的运行、迁移、初始化、恢复、诊断与版本入口，以便不同 profile 行为一致可发现。
- **US-014**：作为运维人员，我希望真实的结构化日志与 doctor 结果可关联、可过滤且不泄密，以便定位启动、请求、worker 与依赖故障。
- **US-015**：作为模块开发者，我希望生成器继续一次产生双方言、OpenAPI、后端与前端垂直切片并通过架构门禁，以便扩展能力不破坏边界。
- **US-016**：作为发布维护者，我希望真实 PostgreSQL、浏览器、原生 Desktop 和安全检查成为不可跳过证据，以便绿色 CI 能证明候选行为。
- **US-017**：作为现有业务用户，我希望 Audit、Organization、Settings、Scheduler、Generator、Files 与 Demo 的既有业务语义在 UI 重构后保持完整，以便产品化不会丢失功能。

## 4. 验收合同

| ID | 前置条件 | 动作或事件 | 可观察结果 | 验证接缝 |
|---|---|---|---|---|
| AC-001 | SQLite 或 PostgreSQL 已迁移且账号为空 | 通过 TTY/secret file 执行 Server Bootstrap | 创建一个启用账号并授予受保护系统管理员角色，写入脱敏审计，随后可登录 | 双方言 CLI 集成测试，覆盖 US-001 |
| AC-002 | 两个操作者对同一空库并发 Bootstrap | 两个事务同时提交 | 恰好一个成功，另一个返回 conflict，库中只有一个首管理员 | PostgreSQL 并发与 SQLite 串行化测试，覆盖 US-001 |
| AC-003 | 数据库已有任意账号或输入弱密码/不安全 secret file | 执行 Bootstrap | 请求被拒绝且无部分账号、授权或明文凭据残留 | CLI contract 与数据库断言，覆盖 US-001 |
| AC-004 | Desktop SQLite 账号为空 | 完成原生首次设置 | 同一用例创建首管理员和首个 Session，工作区直接打开，WebView 持久存储无密码/原始 Session | Tauri native smoke，覆盖 US-002 |
| AC-005 | Desktop 已提交首管理员 | Session 创建失败或工作区加载失败 | 管理员不回滚，显示可恢复错误并进入普通登录，重启不再展示 Bootstrap | Desktop host 集成与 native smoke，覆盖 US-002 |
| AC-006 | 已有一个非删除账号且 API/worker 已停止 | recover-admin 获得运维锁并提交强密码、原因 | 账号重新启用、获得系统管理员角色、旧 Session 全撤销并产生脱敏审计 | 双方言 CLI 集成，覆盖 US-003 |
| AC-007 | API/worker 运行、目标为不存在/deleted 账号或无符合目标 | 执行 recover-admin | 稳定失败；不创建账号、不复活 Tombstone；无可恢复账号提示从备份恢复 | 进程与数据库 contract，覆盖 US-003 |
| AC-008 | 连续错误登录或多实例并发错误登录 | 账号桶或来源桶达到阈值 | 所有实例一致返回限流 Problem 与粗粒度 Retry-After，重启后限制仍存在 | 双方言时间/并发集成，覆盖 US-004 |
| AC-009 | 一个存在账号和一个不存在账号使用等价输入 | 进行未触发限流的错误登录 | HTTP status、Problem code、公开正文与密码工作类别等价，均不显示剩余次数 | transport contract 与时序预算测试，覆盖 US-004 |
| AC-010 | 有效 Session 已建立 | 重复调用 `GET /iam/session/current` 和其他认证 GET | 返回身份/权限/稳定 CSRF，Session 行的 token hash、last seen、idle expiry、CSRF hash 均不变 | Session repository 集成，覆盖 US-005 |
| AC-011 | 两个标签共享同一 Session 家族 | A heartbeat/renew，B 随后执行受保护写请求 | B 的既有 CSRF 仍有效，不因 A 的请求得到偶发 403；两个标签最终同步认证状态 | 真实浏览器双页 E2E，覆盖 US-005 |
| AC-012 | Session 进入续期窗口或用户仍活跃 | 调用 heartbeat/renew 或业务写请求 | idle expiry 按合同延长；renew 轮换不透明 token 但保持 CSRF；绝对 expiry 不延长 | API contract 与竞态集成，覆盖 US-005 |
| AC-013 | Session 有效 | 退出、管理员撤销、密码修改、账号禁用/删除或权限撤销 | 下一请求按对应认证/授权结果失败，客户端频道只负责提示和状态同步 | 后端集成与浏览器 E2E，覆盖 US-006 |
| AC-014 | 已编译且已授权的二级 route | 直接访问、刷新、前进或后退 | route、页面、菜单选中、访问标签、面包屑和标题指向同一资源 | Router/component 与浏览器 E2E，覆盖 US-007 |
| AC-015 | route 不存在或 capability 不包含所需权限 | 直接导航或权限变化后再次导航 | 分别显示稳定 404 或 403；数据库菜单不能加载任意 component | manifest contract 与浏览器 E2E，覆盖 US-007 |
| AC-016 | Web HTML5 history 与 Desktop hash history | 运行相同 route manifest 合同 | 两个宿主提供相同可达业务页面和权限语义，URL 编码符合各自 history | 双 App manifest test，覆盖 US-007、US-008 |
| AC-017 | 目标视口 1440x900、1280x800、390x844 与 Desktop 目标窗口 | 使用导航、查询、表格、分页、表单和对话框 | 无重叠、截断关键操作或无意横向破版；键盘焦点可见，reduced-motion 生效 | Playwright 截图/交互与 Tauri smoke，覆盖 US-008 |
| AC-018 | 用户切换并持久化 dark mode | 重启 Web/Desktop | 使用 token 驱动的暗色主题恢复；页面不以硬编码颜色破坏可读性 | UI token/unit 与视觉检查，覆盖 US-008 |
| AC-019 | 角色对同一 Permission Code 配置五种范围之一 | 用户查询或修改受范围保护资源 | 后端只允许范围内资源；SQLite/PostgreSQL 结果集合等价 | 授权双方言 contract，覆盖 US-009 |
| AC-020 | 多个启用角色中仅部分角色授予目标 Permission Code | 计算有效范围 | 只合并授予权限角色的范围并取并集；无显式 deny；组织变化在下一请求生效 | IAM scope unit/integration，覆盖 US-009 |
| AC-021 | 账号无主部门或 custom 集合为空 | 使用 organization/organization-tree/custom 范围 | 可见集合为空而不是退化为 all/self | 双方言授权测试，覆盖 US-009、US-010 |
| AC-022 | 账号设置主部门和多个岗位 | 变更、删除或移动部门/岗位 | 一个主部门不变量保持；跨部门岗位按合同拒绝；岗位不隐式修改主部门 | Organization/IAM 集成，覆盖 US-010 |
| AC-023 | 删除账号不会破坏最后启用系统管理员不变量 | 提交 transfer 策略与有效目标 | 账号立即不可登录并进入 pending，Files 幂等转移 owner，完成后账号净化且审计引用稳定 | Outbox 双方言集成，覆盖 US-011 |
| AC-024 | 删除目标为最后启用系统管理员，或策略缺失/目标无效 | 提交删除命令 | 返回稳定 validation/conflict，账号和 Session 状态不变 | IAM application contract，覆盖 US-011 |
| AC-025 | 管理员二次确认 purge 且 worker 尚未 claim | 查看状态或取消 | 状态为 queued，取消成功后账号恢复 disabled 且不删除文件 | 账号/Files 状态机集成，覆盖 US-011 |
| AC-026 | purge 已被 worker claim | 尝试取消、重复消费或经历崩溃恢复 | 取消 conflict；消费幂等可重试；完成后物理内容与 metadata 删除且账号净化 | Files/outbox 故障注入，覆盖 US-011 |
| AC-027 | 上传将超过任一账号/全局配额 | 并发上传 | 只有可被原子预留的对象成功，其余返回稳定容量 Problem，不超卖计数或字节 | 双方言配额并发测试，覆盖 US-012 |
| AC-028 | stage/publish 之间失败或进程崩溃 | reconciliation 扫描超时预留 | 预留被安全回收或已发布对象恢复 ready，不产生重复内容或永久占用 | Files 故障注入，覆盖 US-012 |
| AC-029 | 可用磁盘低于字节或比例水位 | 上传、下载和删除 | 上传拒绝；下载与删除继续成功；doctor 与日志显示脱敏水位状态 | 存储 adapter 集成，覆盖 US-012、US-014 |
| AC-030 | PostgreSQL 生产 profile | 分别运行 migrate、serve、worker | migrate 独占执行；serve 仅 API；worker 独立；schema 过旧/过新/未知均不 ready 并退出 | CLI/process integration，覆盖 US-013 |
| AC-031 | Server SQLite 或开发 profile、Desktop SQLite | 启动运行角色 | Server 可 `serve --with-worker` 且单实例；Desktop 备份成功后自动迁移并组合 API/worker | Host lifecycle 集成，覆盖 US-013 |
| AC-032 | 任意支持 profile | 执行 doctor/version 或配置 log.level | doctor 以非零退出码标识失败项，version 可机器读取，日志级别真实过滤且不泄漏 secret | CLI snapshot、日志 capture 与 secret scan，覆盖 US-013、US-014 |
| AC-033 | 请求、worker 或依赖发生已分类事件 | 检查运行日志与 Audit | 运行日志含 service/version/profile/trace/request/route/module/status/latency/database/error class；Audit 记录业务事实；两者均无请求正文和 secret | 日志 schema contract，覆盖 US-014 |
| AC-034 | Generator 接收合法单表 CRUD draft | preview 后 write 并在隔离工作区验证 | 一次生成双方言 migration、OpenAPI、Go 垂直模块、domain/web-domain、权限注册与测试，并通过架构/构建门禁 | Generator isolation integration，覆盖 US-015 |
| AC-035 | CI required PostgreSQL job | 启动固定版本 PostgreSQL service 并注入一次性 DSN | 目标双方言集成真实执行；缺 DSN、Skip、零目标测试或 service 不健康使 job 失败 | CI contract 与 PostgreSQL job，覆盖 US-016 |
| AC-036 | 功能编码、静态检查、单元/集成和构建已完成 | 运行真实浏览器与 macOS Tauri 候选验证 | Bootstrap、登录、路由、权限、CRUD、双标签 Session、撤销、文件生命周期、首次设置、重启与持久化通过 | Playwright 与 native smoke evidence，覆盖 US-016 |
| AC-037 | 当前依赖和发行候选 | 执行 govulncheck、pnpm audit、Rust advisory/deny、secret scan、SBOM 与生成漂移 | GO-2026-6112 不可达且 kin-openapi 不低于修复版本；required scan 无未处置可达漏洞或漂移 | 安全/供应链 CI，覆盖 US-016 |
| AC-038 | UI 重构后的 Web/Desktop | 对 Audit、Organization、Settings、Scheduler、Generator、Files、Demo 运行代表性流程 | 既有业务操作和 Problem 语义保持完整，视觉替换未移除功能 | API contract、组件与浏览器流程，覆盖 US-017 |
| AC-039 | 从空库启动三个 profile 的发布候选 | 按当前文档完成迁移、初始化、登录、核心管理与重启 | SQLite Server、PostgreSQL Server、Desktop SQLite 均可复现；命令和文档无固定密码或旧入口 | clean-room release rehearsal，覆盖 US-001、US-002、US-013、US-016 |

## 5. 范围

### IN

- **IN-001**：修复 GO-2026-6112 对应的可达依赖版本并建立持续漏洞门禁。
- **IN-002**：IAM 一次性 Bootstrap、Desktop 首次设置、离线 recover-admin 与最后启用系统管理员不变量。
- **IN-003**：数据库持久账号/来源双桶登录限流、dummy Argon2、可信来源解析与脱敏反馈。
- **IN-004**：Session 家族稳定 CSRF、只读认证 GET、heartbeat、renew、多标签客户端同步与即时服务端撤销。
- **IN-005**：Vue Router 单一事实源、真实子路由、菜单/capability 交集、访问标签/面包屑/标题/history/403/404。
- **IN-006**：Element Plus、Sass token、Lucide 与共享管理组件驱动的 Web/Desktop 全页面 UI/UX 重构。
- **IN-007**：all/self/organization/organization-tree/custom 数据范围、角色范围并集、一个可空主部门和多个岗位。
- **IN-008**：账号 Tombstone、显式 transfer/purge、取消边界、Outbox 消费与最终身份净化。
- **IN-009**：文件账号/全局配额、磁盘水位、预留、stage/publish 与 reconciliation。
- **IN-010**：统一产品 CLI、profile 运行拓扑、profile migration 策略、doctor 与结构化日志。
- **IN-011**：Generator 全垂直切片能力保持并适配新合同与设计系统。
- **IN-012**：真实 PostgreSQL、浏览器、Tauri native、安全供应链、文档与 clean-room 发布验证。

### REUSE

- **REUSE-001**：复用 `<Path>go-admin-plus/internal/modules/</Path>` 的模块所有权、application/port 边界、Bun/Goose 双方言和 Transactional Outbox。
- **REUSE-002**：复用 `<Path>contracts/openapi/product.yaml</Path>` 的模块化 OpenAPI 与生成 transport，所有合同变化先改 canonical OpenAPI。
- **REUSE-003**：复用 `<Path>go-admin-plus-ui/packages/domains/</Path>`、`<Path>go-admin-plus-ui/packages/web-domains/</Path>` 和 adapter 边界，不复制 Web/Desktop 业务实现。
- **REUSE-004**：复用 `<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/</Path>` 的 Tauri 2 sidecar、loopback capability 与原生 secret 隔离；保留该路径现有用户修改。
- **REUSE-005**：复用 `<Path>Taskfile.yml</Path>` 作为跨语言产品命令面，并让 CI 编排相同命令而不是另建隐藏流程。
- **REUSE-006**：复用 Audit、Organization、Settings、Scheduler、Generator、Files、Demo 的现有业务语义与稳定错误模型。

### OUT

- **OOS-001**：租户能力永久不做；本产品仅有组织数据范围，不恢复 tenant 字段、SQL 或 UI。
- **OOS-002**：Redis、JWT/localStorage bearer、Casbin URL 策略和 AutoMigrate 不做；数据库与不透明 Session 继续是真实正确性来源。
- **OOS-003**：旧命令、旧账号 DELETE/batch-delete、旧滚动 CSRF、旧页面 tab 和固定初始化 SQL 不兼容；零兼容策略要求原子切换。
- **OOS-004**：不按 Backplane、RuoYi 或 Java 工程外观再次搬迁目录，不创建 common/utils/system/operations 无边界聚合包。
- **OOS-005**：多主部门、显式 deny、任意数据库 component string、隐藏文件回收站与远程管理员恢复 API 不做。
- **OOS-006**：签名、公证、应用商店发布和自动更新不做；个人本地可安装运行即可。
- **OOS-007**：多节点 SQLite 与 Desktop PostgreSQL 不做；Server 支持 SQLite/PostgreSQL，Desktop 仅 SQLite。

## 6. 已锁定实现约束

- **DEC-001**：保留当前模块化单体、OpenAPI、pnpm workspace、Tauri 2 和根治理，不进行第二轮目录重排。来源：`PLAN/P0`、`LOG-002`。
- **DEC-002**：Bootstrap 是共享 IAM 用例的 Server CLI/Desktop 原生 adapter，且只在账号为空时成功。来源：`ADR-001`。
- **DEC-003**：账号删除采用 Tombstone、Integration Event 与最终净化，IAM 不跨模块操作 Files 私有表。来源：`ADR-002`。
- **DEC-004**：登录限流使用当前数据库持久账号/来源双桶，不能关闭。来源：`ADR-003`。
- **DEC-005**：数据范围稳定为五种，后端同时应用 Permission Code 与 scope。来源：`ADR-004`。
- **DEC-006**：Server 运维入口收敛为单一产品 CLI，Desktop sidecar 独立。来源：`ADR-005`。
- **DEC-007**：认证 GET 只读，Session 通过 heartbeat、受保护写请求和提前 renew 续期，CSRF 在 Session 家族内稳定。来源：`ADR-006`。
- **DEC-008**：recover-admin 是停服务、持有直接数据库权限和显式原因的离线流程。来源：`ADR-007`。
- **DEC-009**：Desktop Bootstrap 后建立本地 Session；Session 阶段失败不回滚管理员。来源：`ADR-008`、`LOG-020`。
- **DEC-010**：账号删除逐次显式选择 transfer 或 purge，禁止默认策略。来源：`ADR-009`。
- **DEC-011**：同一权限的多角色有效范围取并集，不引入显式 deny。来源：`ADR-010`。
- **DEC-012**：登录限流只暴露模糊凭据错误或粗粒度 Retry-After。来源：`ADR-011`。
- **DEC-013**：PostgreSQL 生产 API/worker 分离，Server SQLite/dev 可一体，Desktop 一体。来源：`ADR-012`。
- **DEC-014**：recover-admin 只修改既有非删除账号，不能建立救援账号或复活 Tombstone。来源：`ADR-013`。
- **DEC-015**：purge 在 worker claim 前可取消，之后永久且不可恢复。来源：`ADR-014`。
- **DEC-016**：PostgreSQL 生产显式 migrate，Server SQLite 自动迁移，Desktop 成功备份后自动迁移。来源：`ADR-015`。
- **DEC-017**：账号最多一个可空主部门和多个岗位，组织类范围从主部门计算。来源：`ADR-016`。
- **DEC-018**：Web 使用 HTML5 history、Desktop 使用 hash history；route manifest 是导航唯一事实。来源：`LOG-005`、`PLAN/P3`。
- **DEC-019**：视觉基线为中性浅色工作面、炭黑侧栏、克制绿色强调和紧凑密度，并提供持久暗色模式。来源：`LOG-024`。
- **DEC-020**：代码、静态检查、单元/集成与构建先完成，真实 E2E 在候选阶段集中执行，但适用 E2E 是最终必需门禁。来源：`LOG-006`。

## 7. 数据、接口与兼容

- **公共接口变化：** 在 canonical OpenAPI 增加 `POST /iam/session/heartbeat`、`POST /iam/session/renew`；以 `POST/GET /iam/administration/users/{userId}/deletion` 和 `POST /iam/administration/users/{userId}/deletion/cancel` 替换账号单删与 batch-delete；用户/角色合同增加主部门、岗位和五种数据范围；文件 Problem 增加稳定配额/磁盘容量错误。`GET /iam/session/current` 保持路径但改为只读且返回稳定 CSRF。其余现有业务路径和 DTO 不因 UI 重构改名。
- **数据模型与持久化：** 双方言增加 Bootstrap/系统初始化约束、登录限流桶、账号生命周期与匿名审计引用、账号主部门/岗位关联、角色数据范围/custom 部门集合、账号删除工作流、versioned outbox payload、文件配额预留与 reconciliation 状态。Session 模型支持家族稳定 CSRF、idle/absolute expiry 与显式 renew。所有表继续由所属模块管理。
- **兼容要求：** 无旧行为兼容要求。旧 Server 小二进制、固定密码 SQL、账号 DELETE/batch-delete、每请求 Session/CSRF 滚动、页面平行 tab 状态和旧 UI 原子移除，不保留 alias、双写或 feature flag。
- **迁移要求：** 仅 forward-only migration。已有账号迁移为明确 active/disabled 状态并保留 ID；现有角色得到不扩大权限的确定默认范围；现有 Session 在切换时统一撤销，用户重新登录；现有文件用量通过受控 backfill 建立配额基线。PostgreSQL 迁移前由部署流程备份，Desktop 由宿主强制备份，回退依靠备份与旧制品而非 destructive down migration。
- **发布或运维影响：** PostgreSQL 部署新增显式 `migrate` 步骤并分离 API/worker；Server SQLite/dev 可一体；Desktop 仍自包含。operator 文档必须覆盖 Bootstrap、recover-admin、doctor、备份/恢复、配额和低磁盘处理。required CI 需要真实 PostgreSQL service、浏览器和 macOS native runner。签名、公证记为 not-required。

## 8. 非功能要求

- **NFR-001 安全与隐私：** 密码仅经 TTY 或受限 secret file 输入，不进入 argv、日志、审计、前端持久化或源码；不透明 Session、CSRF 与 DSN 均按 secret 脱敏。登录保护不可关闭，不存在账号执行等价密码工作。所有授权由后端 Permission Code、数据库 Session 与有效 scope 判定。`kin-openapi` 至少使用 GO-2026-6112 修复版本。
- **NFR-002 性能与容量：** 认证 GET 必须零 Session 写入；heartbeat/renew 有客户端协调和服务端幂等/限频，避免每标签写热点。列表继续分页且 page size 有界。上传配额预留原子且不超卖，reconciliation 每批有界，低磁盘时优先保留读取和释放空间能力。登录桶清理同样有界。
- **NFR-003 可用性与可靠性：** Bootstrap、recover-admin、Session renew、账号 deletion、outbox 消费、配额预留与 migration 对重复、并发、崩溃有确定结果。PostgreSQL schema 不匹配不得 ready；Desktop 备份失败不得迁移。UI 在 1440x900、1280x800、390x844 与目标 Desktop 窗口保持可操作，关键流程支持键盘、可见焦点和 reduced-motion。
- **NFR-004 可观测性与运营：** 运行日志至少具有 service、version、profile、trace、request、route、module、status、latency、database 与稳定 error class；Server 生产 JSON stdout，Desktop 轮转文件，开发 console。doctor 输出可机器判定项目和退出码。Audit 只记录业务事实，不复制完整请求/响应正文。
- **NFR-005 可维护性：** canonical OpenAPI、route manifest、Sass token、共享 UI 组件和模块 port 分别保持单一事实源；生成物不得手改。新增能力通过架构测试禁止跨模块私表、无边界 common 包、Web/Desktop 页面复制和数据库任意组件加载。
- **NFR-006 可移植性：** Server SQLite/PostgreSQL 行为合同等价；Desktop 仅 SQLite。根命令在项目支持的开发平台使用相同 Taskfile 入口，平台差异封装在现有脚本与宿主 adapter。
- **NFR-007 供应链与发布：** Go、pnpm、Rust 生产依赖和 actions 继续固定或锁定；生成漂移、可达漏洞、secret、SBOM 与真实执行状态可审计。缺少 required 环境不能得到绿色结果。

## 9. 验证策略

| 接缝 | 层级 | 覆盖合同 | 现有先例或命令 | Evidence 类型 |
|---|---|---|---|---|
| Spec/架构/零兼容 | 静态治理 | AC-034、AC-039 | `task governance:check`、`task architecture:check`、`task compatibility:zero`、`task docs:check` | 命令记录与 diff |
| OpenAPI 与生成 transport | 合同/生成 | AC-009 至 AC-015、AC-023 至 AC-026、AC-038 | `task contract:lint`、`task generate:check` | lint、contract 与 drift 结果 |
| IAM/Session/CLI SQLite | 单元与集成 | AC-001 至 AC-013、AC-019 至 AC-025、AC-030 至 AC-033 | `task test`；先例 `<Path>go-admin-plus/internal/modules/iam/</Path>`、`<Path>go-admin-plus/internal/host/server/host_test.go</Path>` | Go test 与数据库断言 |
| PostgreSQL 双方言与并发 | 真实服务集成 | AC-001、AC-002、AC-006、AC-008、AC-019 至 AC-023、AC-027、AC-030、AC-035 | required PostgreSQL CI service；先例 `<Path>go-admin-plus/internal/platform/migrations/postgres_integration_test.go</Path>` | 非 Skip test report 与 service log |
| Files/outbox/配额 | 故障注入集成 | AC-023 至 AC-029 | 先例 `<Path>go-admin-plus/internal/modules/files/service_test.go</Path>`、`<Path>go-admin-plus/internal/modules/files/storage_test.go</Path>` | 状态、磁盘与数据库 evidence |
| Vue route/manifest/domain | 单元与组件 | AC-014 至 AC-018、AC-038 | `pnpm --dir go-admin-plus-ui test`、`typecheck`、`build`；先例 `<Path>go-admin-plus-ui/tests/</Path>` | Vitest、typecheck 与 build report |
| 真实 Web 浏览器 | 系统 E2E | AC-011、AC-013 至 AC-018、AC-036、AC-038、AC-039 | 真实后端加 `<Path>go-admin-plus-ui/tests/e2e/</Path>` runner | Playwright trace、截图与结果 |
| Tauri 原生宿主 | native integration/smoke | AC-004、AC-005、AC-016 至 AC-018、AC-031、AC-036、AC-039 | `cargo test`、`cargo clippy`、Tauri build；先例 `<Path>go-admin-plus/internal/host/desktop/host_integration_test.go</Path>` | macOS native log 与 artifact check |
| CLI/进程/日志/doctor | 进程 contract | AC-006、AC-007、AC-029 至 AC-033 | Go command tests、`task build TARGET=server`、进程 shutdown/readiness probes | exit code、stdout/stderr 与 redaction scan |
| Generator 隔离生成 | 系统集成 | AC-034 | 先例 `<Path>go-admin-plus/internal/modules/generator/generator_test.go</Path>` | 隔离工作区 lint/test/build |
| 安全与供应链 | 静态/动态扫描 | AC-003、AC-004、AC-007、AC-009、AC-032、AC-033、AC-037 | govulncheck、pnpm audit、Rust advisory/deny、secret scan、SBOM | 扫描报告与处置记录 |
| Clean-room 候选 | 发布演练 | AC-035 至 AC-039 | `task release:verify`、`task release` 及三 profile 空库演练 | Evidence 索引与环境清单 |

验证顺序遵守已确认决策：每个垂直切片先完成实现、格式化、静态检查、单元/集成测试与构建；全部编码和构建完成后再集中执行真实 PostgreSQL、浏览器和 native Desktop 候选矩阵。该顺序不降低最终门禁，任何适用 E2E 未通过都不能形成发布候选。

## 10. 风险、假设与未决问题

### 风险

- **RISK-001**：Session 协议与账号删除合同属于安全相关破坏性变更。缓解：canonical OpenAPI、生成客户端、双方言 migration、后端与双 App 在同一垂直切片原子切换，不保留双轨。
- **RISK-002**：PostgreSQL 双桶、配额与 Outbox 的并发语义可能与 SQLite 串行行为不同。缓解：真实 PostgreSQL service、竞争测试、故障注入和 required non-Skip 检查。
- **RISK-003**：UI 全量重构容易产生功能遗漏。缓解：保留业务 domain/transport，按 route manifest 建立功能清单，AC-038 对每个现有模块执行代表性流程。
- **RISK-004**：账号 purge 不可逆。缓解：显式策略、二次确认、queued 状态可取消、claim 后明确不可取消、Audit 和备份文档。
- **RISK-005**：现有用户对 Desktop Tauri 主机文件有未提交修改。缓解：实施前记录路径级基线，所有 Tickets 把该文件视为共享/受保护路径，不使用 reset 或 checkout 清理。
- **RISK-006**：全矩阵耗时较高。缓解：先分层执行快速门禁，真实 E2E 在候选阶段集中运行；required gate 仍不允许 Skip。

### 已采用的低影响假设

- **ASM-001**：`Retry-After` 的具体分桶粒度、阈值和时间窗由安全边界内的类型化配置确定；Tickets 可选择保守默认值，但不能改变“不可关闭、双桶、粗粒度反馈”合同。验证：配置边界与时间控制测试。
- **ASM-002**：目标 Desktop 窗口最小尺寸沿用 Tauri 当前产品配置；实现时由原生配置读取，不在 Spec 固定像素。验证：读取配置并执行最小窗口 smoke。
- **ASM-003**：现有业务 DTO 仅在本规格明确的数据范围、生命周期和容量字段处演进；纯视觉改造不修改后端业务语义。验证：OpenAPI diff 分类与 AC-038。
- **ASM-004**：Desktop 本地首次设置 adapter 可复用当前 sidecar control boundary，不新增公开网络初始化端点。验证：capability allowlist 与 loopback/native contract。

### 未决问题

无。
