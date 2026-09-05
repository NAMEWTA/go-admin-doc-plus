# 前端 Review 发现

范围：`go-admin-plus-ui/` 的 Web/Desktop App、app-shell、runtime adapters、headless domains、web-domains、测试与前端文档。检查日期：2026-09-05。

## 发现

### [P1] Desktop 首次设置预填固定管理员密码

- 文件：[FirstSetupGate.vue](/D:/Document/code/go-admin-plus/go-admin-plus-ui/apps/admin-desktop/src/first-setup/FirstSetupGate.vue:12-16)
- 证据：首次设置表单把用户名预填为 `admin`，并把密码与确认密码都预填为公开常量 `1234567890`。该值满足模板的 `minlength="10"`，用户不编辑密码即可提交创建管理员。
- 影响：任何能拿到安装包、截图或操作说明的人都知道新安装实例的初始凭据；用户误点“创建并进入工作区”会直接生成可猜测的管理员密码。Desktop 首次安装是高权限入口，这属于可被直接利用的账户接管风险。
- 建议：密码字段默认置空并要求用户主动输入；提交前拒绝常见/默认密码；最好由本地安全随机生成一次性初始化口令并强制首次登录修改，且不要在源码、测试或文档中保留可用默认凭据。

### [P1] 登录切换期间运行时授权投影没有版本校验，旧请求可覆盖新用户

- 文件：[ProductWorkspace.vue](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue:159-179、181-213)
- 证据：`loadRuntime()` 同时请求 identity/navigation，但没有请求序列、AbortController 或当前 session 操作版本；`authenticated()` 在每次登录后直接启动新的 `loadRuntime()`。会话订阅只清空当前投影，没有使已经在途的 `loadRuntime()` 失效。
- 复现路径：账号 A 登录后触发 `loadRuntime(A)`，立刻注销并登录账号 B，触发 `loadRuntime(B)`；若 A 的 identity/navigation 响应晚于 B 完成，A 的 `permissions`、`dataScope`、`navigationPaths` 会覆盖 B 的投影。之后页面可能显示错误菜单/操作，并用错误能力集发起业务请求（后端仍会拒绝越权请求，但用户界面和错误处理已处于错误账号上下文）。
- 建议：为 runtime 加单调 sequence 与 session generation；启动新 restore/login/logout 时取消或失效旧请求，只有当前 generation 才能写入权限、数据范围和导航状态。为该竞态补充 delayed-response 测试。

### [P2] 自制管理弹窗没有完整焦点约束

- 文件：[AdministrationPage.vue](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/web-domains/iam/src/administration/AdministrationPage.vue:163-190)、[AuditPage.vue](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/web-domains/audit/src/AuditPage.vue:140-146)
- 证据：多个 `role="dialog" aria-modal="true"` 使用普通 `div` backdrop，自行处理点击和 Escape；没有 focus trap、初始焦点恢复（审计详情除外）或打开时锁定背景。IAM 编辑/删除/菜单/角色弹窗中的 Tab 可以离开弹窗并落到背景导航，屏幕阅读器会遇到 `aria-modal=true` 与可访问背景同时存在的矛盾。
- 建议：统一复用 `ui` 的 `FormDialog`/对话框基础组件，或实现共享 dialog composable，保证 `aria-labelledby`、初始焦点、Tab 循环、Escape、关闭后恢复触发按钮以及忙碌状态下的关闭策略；补充键盘和屏幕阅读器测试。

### [P2] 各业务 Web client 重复实现 CSRF、错误分类和串行队列，行为已出现不一致

- 文件：[web-administration-client.ts](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/web-domains/iam/src/administration/web-administration-client.ts:5-39)、[web-scheduler-client.ts](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/web-domains/scheduler/src/web-scheduler-client.ts:5-36)、[web-demo-client.ts](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/web-domains/demo/src/web-demo-client.ts:7-37)、[web-files-client.ts](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/web-domains/files/src/web-files-client.ts:20-57)、[web-audit-client.ts](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/web-domains/audit/src/web-audit-client.ts:11-48)
- 证据：每个 client 都维护独立 `csrf`、请求尾链和可变的响应错误状态；Demo/Files 会校验 CSRF header 格式，Administration/Scheduler/Audit 直接接受任意非空 `X-CSRF-Token`。不同 client 对 401/403/CSRF_REJECTED、traceId 和响应头的处理也不一致。当前 app-shell 会同时创建全部 client，因此同一页面存在多套 transport policy。
- 影响：后端合同或 token 轮换策略调整时容易只修复部分模块；异常请求可能被错误映射成 unavailable/forbidden，且非法 token 会被本地保存直到下一次失败，造成难以诊断的会话失效。重复队列也增加了竞态和维护成本。
- 建议：在 `adapters/browser` 或 `api-client` 提供唯一 session-aware transport（统一 CSRF 格式、凭据、问题响应和 traceId），业务 client 只负责 OpenAPI 调用与领域解包；删除各模块的复制实现并用跨模块合同测试覆盖。

## 复用、架构与清理观察

- 领域包与宿主适配器的依赖方向总体符合仓库架构，生成 OpenAPI 文件也保持在 generated 目录；未发现需要保留的旧 package scope 或旧路由别名。
- `ProductWorkspace.vue` 直接实例化 IAM、Audit、Scheduler、Demo、Files 的全部 client/controller。这样符合当前 shell composition 约定，但会让首次加载装配所有业务 transport；应在后续拆分时保持组合根集中，避免页面继续复制 controller 创建逻辑。
- IAM/Audit 的手写弹窗、表格工具栏和错误显示存在重复 markup；可逐步收敛到 `@go-admin-plus/ui/components`，以减少焦点、ARIA 和 busy 状态分叉。
- 前端目录下仅发现 Desktop native E2E 说明和 sidecar staging 说明；未发现明显的历史业务文档需要删除。`tests/e2e/desktop/README.md` 明确说明 native E2E 需要 macOS/显式 opt-in，不应把未执行当作发布通过。

## 验证记录

- 通过：`pnpm lint`
- 通过：`pnpm typecheck`
- 通过：`pnpm check:workspace`
- 通过：`pnpm build`（Web 构建、Desktop 生产资产校验均通过）
- `pnpm test`：并发默认配置下 29 个测试文件、190 个测试通过，但 Vitest worker 在 `tests/shell/list-form.spec.ts` 启动时超时，命令以非零退出；使用 `pnpm exec vitest run --config tests/shell/vitest.config.ts --maxWorkers=1` 后 30 个测试文件、213 个测试全部通过，说明当前失败更接近 worker 并发/启动稳定性问题。CI 建议限制 worker 并发或增加启动超时，并保留一次完整并发重跑。
