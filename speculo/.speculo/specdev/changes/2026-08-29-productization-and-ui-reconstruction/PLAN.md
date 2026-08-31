# Go Admin Plus 产品化与 UI 重构完整计划

## 工件定位

本文件完整保存 2026-08-29 形成的设计计划，供后续 Grill、Spec、Tickets 与 Goal Plan 继续细化。它不是实现授权，也不替代 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/design-tree.json</Path>`、`<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ADR.md</Path>`、`<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>` 或未来的 Ticket 状态。

## 结论

本次不再进行第二轮目录搬迁。当前模块化单体、OpenAPI、双方言迁移、pnpm workspace、Tauri 2 sidecar 和根治理已经是正确基础。下一轮工作聚焦产品闭环、安全、UI/UX、运维与交付可信度。

源码与工程规范已经达到严肃产品工程水平，但全新安装、跨标签会话、登录防护、二级路由、真实 PostgreSQL/浏览器证据、可达依赖漏洞、文件生命周期和结构化日志仍阻止公开生产候选。

## 已确认问题

1. 空库迁移后没有账号，管理员 API 又要求现有管理员授权，导致首次安装无法进入系统。
2. manifest 声明二级 URL，但 ProductWorkspace 与领域页面各自维护一级模块和默认标签，URL、标题、页面和历史状态脱节。
3. 所有认证请求刷新 Session 并覆盖 CSRF hash，跨浏览器标签互相使对方失效，认证读取成为数据库写入热点。
4. Argon2 工作预算只限制单进程并发，没有账号、来源、时间窗和多实例持久登录限流。
5. CI 后端任务没有真实 PostgreSQL service/DSN，浏览器和 native runner 可以因缺少 opt-in 或环境变量而成功退出。
6. `<Path>go-admin-plus/go.mod</Path>` 的 kin-openapi v0.142.0 受 GO-2026-6112 影响，请求验证链路可触发 nil pointer panic；修复下限为 v0.144.0。
7. 账号硬删除与 Files owner 生命周期分离，且只有单文件 10 MiB 上限，没有账号、对象数量、全局容量和磁盘水位治理。
8. 配置解析了 log.level，但产品 Host 没有注入 logger，启动、请求、worker 和依赖失败缺少统一结构化日志。

## 融合原则

| 来源 | 吸收 | 明确拒绝 |
|---|---|---|
| Go Admin Plus | 模块化单体、端口边界、OpenAPI、Goose、Bun、Outbox、不透明 Session、Web/Desktop 共用业务 | 每请求刷新 CSRF、无首管理员、手写路由、零散原生 UI |
| Backplane | 单一运维入口、登录锁定、结构化日志字段、操作审计维度、动态菜单易用性 | admin/admin123、JWT/localStorage、AutoMigrate、Casbin 路径授权、Redis 正确性依赖 |
| RuoYi / Plus UI | 系统管理能力模型、菜单权限、数据范围、代码生成器、成熟后台布局、Element Plus 样式体系 | Java 模块机械翻译、庞大 common 聚合、租户体系 |

参考代码只是设计输入，不构成依赖或兼容目标。外部安全和框架基准来自：

- RESEARCH:<Url>https://pkg.go.dev/vuln/GO-2026-6112</Url>
- RESEARCH:<Url>https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html</Url>
- RESEARCH:<Url>https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html</Url>
- RESEARCH:<Url>https://router.vuejs.org/guide/</Url>
- RESEARCH:<Url>https://docs.github.com/en/actions/tutorials/use-containerized-services/create-postgresql-service-containers</Url>
- RESEARCH:<Url>https://v2.tauri.app/security/capabilities/</Url>

## 冻结架构

```text
go-admin-plus/
  cmd/go-admin-plus/       serve | worker | migrate | bootstrap | doctor
  cmd/desktop-sidecar/     Tauri 专用本地宿主
  internal/app/            产品组合、命令编排
  internal/application/    跨模块工作流
  internal/modules/        iam、organization、settings、audit、
                           scheduler、generator、files、demo
  internal/platform/       database、config、outbox 等技术能力

go-admin-plus-ui/
  apps/admin-web/
  apps/admin-desktop/
  packages/app-shell/      Vue Router、布局、导航、访问标签
  packages/ui/             Element Plus、主题、稳定共享组件
  packages/domains/        无头业务状态
  packages/web-domains/    按路由拆分的 Vue 页面
  packages/adapters/       browser、desktop
```

后端不创建 common、utils、system 或 operations 等无边界聚合包。前端业务 API 路径和 DTO 不因 UI 重构改名；认证新增合同与生成客户端必须原子提交。

## P0：建立变更基线与解除已知漏洞

1. 使用当前 change 持久化设计树、Spec、Tickets、Goal Plan 和 Evidence。
2. 升级 kin-openapi 到 v0.144.0 或更高兼容版本。
3. 运行 Go 测试、生成漂移、OpenAPI conformance 和 govulncheck。
4. 记录当前未提交文件并保护 `<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>` 中已有用户修改。

Gate：可达 GO-2026-6112 消失，生成物 clean，未触碰用户无关修改。

## P1：首管理员 Bootstrap

1. 在 IAM 中实现唯一 Bootstrap 用例，不通过公开 HTTP 管理 API绕过授权。
2. Server 提供一次性可审计 bootstrap 子命令；凭据只接受 TTY stdin 或权限受限的 secret file，不接受 argv，不输出明文。
3. Bootstrap 在同一事务检查账号为空、创建账号、授予受保护系统管理员角色并写入审计事实；并发执行只有一个成功。
4. Desktop 空库使用原生首次设置页调用同一用例，WebView 不读取 sidecar 启动材料或原始 Session。
5. 删除当前固定密码 SQL 草案；数据库工程资产只保留 schema、非敏感参考数据和显式开发 fixture。
6. 将“用户名等于 admin”保护替换为“系统始终至少存在一个启用系统管理员”的业务不变量。

Gate：SQLite/PostgreSQL 空库均可初始化一次并登录；重复、并发、弱密码、已有账号和日志泄密测试通过。

## P2：Session、CSRF 与登录防护

1. CSRF 改为认证 Session 级稳定随机值，不再每请求覆盖；Session token 轮换沿用该 CSRF 家族。
2. 认证 GET 走只读路径；空闲续期由受 CSRF 保护的 heartbeat、业务写请求与提前 renew 负责。
3. Browser adapter 使用共享 Cookie 和 BroadcastChannel 同步续期、退出与重新登录状态；BroadcastChannel 不是后端授权正确性依赖。
4. 建立双方言持久限流表，账号桶与来源桶必须同时通过；原子更新适用于多进程/多实例。
5. 不存在账号执行等价 dummy Argon2 验证，统一响应；可信代理与来源 IP 解析必须显式配置。
6. 记录可聚合但不泄漏凭据的登录成功、失败、限流和锁定审计事实，并由 worker 有界清理过期桶。

Gate：双标签并发读写不出现相互 403；普通读取不更新 Session；重启/多实例后限流仍有效；时序不暴露账号是否存在。

## P3：统一前端路由事实源

1. 引入 Vue Router，Web 使用 HTML5 history，Desktop 使用 hash history。
2. 产品 route manifest 统一声明 name、path、permission、menu key、title、icon、order 和 component loader。
3. IAM、Settings、Scheduler 等巨型多标签页面拆成真实子路由页面；页面不再维护与 URL 平行的默认 tab。
4. 侧边栏、访问标签、面包屑、标题、刷新恢复、前进后退和 403/404 全部派生自当前 route。
5. 后端授权菜单与编译期 route manifest 做交集，数据库不能通过任意 component string 加载前端代码。

Gate：所有二级 URL 直接刷新、前进后退、页内切换、权限变化和双 App manifest 合同通过。

## P4：全新 UI/UX 与设计系统

1. 引入 Element Plus、Sass token 分层、Lucide 图标和可控组件自动注册。
2. 建立 neutral surface、品牌强调色、状态色、light/dark token、密度、间距、圆角、阴影和层级合同，避免单色主题和装饰性渐变。
3. 在 `<Path>go-admin-plus-ui/packages/ui/</Path>` 建立 AppPage、QueryBar、TableToolbar、DataTable、FormDialog、StatusTag、Pagination、EmptyState 和 FormGrid 等稳定组件。
4. 重构 App Shell：固定/折叠侧栏、移动抽屉、顶栏、访问标签、账户菜单、响应式内容区和无障碍焦点管理。
5. 重做登录页，删除终端/吉祥物式说明和过时端口文案，形成克制的产品身份与认证表单。
6. 按 Dashboard、IAM、Organization、Settings、Audit、Scheduler、Files、Generator、Demo 顺序迁移页面。
7. 删除全局宽泛选择器、页面手写按钮/输入/表格/遮罩以及无法维护的单行巨型模板。

Gate：Web/Desktop 在 1440x900、1280x800、390x844 和目标桌面窗口无重叠、横向破版或不可操作控件；键盘、焦点和 reduced-motion 合同通过。

## P5：账号、文件与容量生命周期

1. 账号删除变为明确的 active、disabled、deletion-pending、deleted 生命周期并立即撤销 Session。
2. 使用 Integration Event 通知 Files 和其他消费者，禁止 IAM 直接操作 Files 私有表。
3. 账号的审计引用保持稳定；文件必须显式转移或完成异步物理清理后才允许最终净化身份信息。
4. 增加单文件、单账号字节、单账号对象数、全局容量、磁盘最小剩余字节/比例配置。
5. 上传先原子预留配额，再 stage/publish；失败、崩溃和超时由 reconciliation worker 回收。
6. 低水位时拒绝新上传但继续允许下载和删除。

Gate：账号删除、配额竞争、崩溃恢复、低磁盘水位、孤儿文件和双方言一致性测试通过。

## P6：日志、审计、CLI 与脚手架能力

1. 注入统一结构化 logger；Server 输出 JSON stdout，Desktop 输出受控轮转文件，开发 profile 可使用可读 console。
2. 字段包含 service、version、profile、trace、request、route、module、status、latency、database 和稳定 error class；禁止密码、Session、DSN 和完整请求体。
3. 保留 Audit 事实模型，只吸收 Backplane 的操作维度，不复制请求/响应正文捕获。
4. 收敛 Server 侧 serve、worker、migrate、bootstrap、doctor、version 到单一产品 CLI；Desktop sidecar 保持独立安全边界。
5. doctor 检查配置、secret reference、数据库、迁移、Bootstrap 状态、文件根、磁盘水位、worker 和版本兼容性。
6. 完善在线 Session 查询/撤销、登录/操作审计视图、动态菜单/capability 交集、组织数据范围、参数字典、任务、文件和生成器体验。
7. Generator 一次生成双方言迁移、OpenAPI、Go 垂直模块、前端 domain/web-domain、权限注册和测试，并通过架构门禁。

Gate：log.level 生效、敏感数据扫描通过、doctor 可诊断三 profile、生成模块在隔离环境完成 lint/typecheck/test/build。

## P7：真实交付门禁

1. 后端 CI 使用固定 digest PostgreSQL service 与健康检查，显式注入 disposable DSN。
2. Required PostgreSQL 测试启用 require flag；任何 Skip 或缺失 DSN 都使 job 失败。
3. 浏览器 E2E 启动真实 Web 与真实后端，覆盖 Bootstrap、登录、路由、权限、CRUD、双标签 CSRF、Session 撤销和文件生命周期。
4. 核心用户流程至少覆盖 Server SQLite/PostgreSQL；UI 视觉与路由只运行必要浏览器矩阵。
5. macOS 原生 runner 验证 Tauri sidecar、首次设置、登录、重启、SQLite 持久化、能力 allowlist 和生产资产无测试控制标记。
6. 增加 govulncheck、生产 pnpm audit、Rust advisory/deny、secret scan、SBOM 和生成漂移门禁。

Gate：required 脚本不再缺环境即成功，真实 PG、浏览器和 Desktop 证据与具体验收合同关联。

## P8：发布候选与文档收敛

1. 从空库依次完成 SQLite Server、PostgreSQL Server、Desktop SQLite 的迁移、Bootstrap、核心管理流程和重启。
2. 更新 README、开发、数据库、部署、故障排查、安全和脚手架扩展文档，删除固定密码与旧命令描述。
3. 根 Taskfile 继续作为产品命令面，CI 只编排相同任务。
4. 个人自用不要求签名、公证或受保护安装；未执行项目明确记为 not-required/not-run，不伪装为 passed。
5. 执行治理、架构、零兼容、文档、生成、Go/pnpm/Rust 和适用 E2E 门禁后形成候选。

Gate：clean tree 候选从零可运行，三 profile 适用能力完整，文档与命令当前有效。

## 最终验收门槛

- SQLite、PostgreSQL 空库均能一次性 Bootstrap 并登录。
- 不存在固定管理员密码、自动弱管理员或通过 HTTP 绕过首次授权的后门。
- 二级 URL 刷新、前进后退、访问标签、标题、菜单和页面内容完全一致。
- 两个浏览器标签并行读写不会因另一个标签刷新 CSRF 而偶发 403。
- 登录限流在重启、多进程和 PostgreSQL 多实例后仍有效。
- Web/Desktop 的功能 manifest、权限、业务页面和错误语义一致。
- 账号删除不会产生不可追踪文件，容量与磁盘水位可观测并受控。
- 结构化日志实际接入且不泄漏 secret；审计与运行日志职责分离。
- PostgreSQL、浏览器、Tauri 和漏洞扫描不再通过缺环境或 Skip 变绿。
- govulncheck 无可达漏洞，构建、测试、生成和文档 clean。
- 不保留旧命令、旧页面模式、固定管理员密码或兼容层。

## 执行与回退边界

- 每个阶段先形成可构建、可测试的垂直落点，再进入下一阶段；不长期保留新旧双轨。
- 数据库只使用 forward-only migration。回退依靠源码/制品回退与迁移前备份，不编写破坏性 down migration。
- 公共合同、安全语义、数据生命周期或验收发生变化时必须返回 Grill/Spec 更新，不能只在 Ticket 或代码中覆盖。
- 当前工作区已有修改属于用户；实现前先建立路径清单并逐项保留，不能通过 reset/checkout 清理。
- 本计划不授权 commit、push、merge、worktree 清理、部署或发布。
