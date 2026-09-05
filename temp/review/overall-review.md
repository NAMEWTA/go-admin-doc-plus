# Go Admin Plus 全面 Review 报告

审查日期：2026-09-05  
审查范围：Go 后端、OpenAPI、数据库/部署/发行脚本、Vue Web、Tauri Desktop、测试、根治理脚本、当前产品文档及 Speculo 状态。  
审查依据：`AGENTS.md`、Speculo `specdev` 工作流入口、`backend-development` 与 `frontend-development` 约束，以及当前代码和门禁实际结果。

## 结论

当前项目的目录分层和大部分调用方向已经接近目标架构，但不具备直接发布条件。至少有四项 P1 运行时问题，分别涉及默认管理员凭据、登录身份竞态、用户删除数据一致性和文件容量安全；另有三项 P1 文档/治理/供应链合同会让新安装或生产部署走到错误路径。现有 `architecture:check`、`compatibility:zero`、`docs:check` 全部通过，但它们没有覆盖这些运行时语义和发布资产一致性问题。

## 必须在发布前修复

### P1. Desktop 首次设置预填公开管理员密码

位置：[FirstSetupGate.vue](/D:/Document/code/go-admin-plus/go-admin-plus-ui/apps/admin-desktop/src/first-setup/FirstSetupGate.vue:12)

`password` 和 `confirmation` 初始值都是 `1234567890`，且满足最小长度，用户不修改即可创建管理员。该值在源码和构建产物中可见，任何拿到安装包或截图的人都知道新安装实例的初始凭据。

影响是首次设置阶段直接暴露高权限账户接管窗口。密码字段必须为空并要求用户主动设置；同时应拒绝常见默认密码，或者生成一次性随机初始化凭据并强制首次登录修改。测试、文档和示例中也不能保留可用默认密码。

### P1. 登录切换竞态可污染新用户的权限投影

位置：[ProductWorkspace.vue](/D:/Document/code/go-admin-plus/go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue:159)

`loadRuntime()` 并行流程没有请求序列、会话 generation 或取消机制。用户 A 的 `loadIdentity/loadNavigation` 在注销并登录用户 B 后仍可能返回，并覆盖 B 的 `permissions`、`dataScope` 和导航集合。后端仍会重新授权，因此主要影响是错误菜单、错误页面和错误请求上下文，但在高权限切换场景会造成严重的界面误导和操作风险。

每次 restore/login/logout 都应递增 session generation 或取消旧请求；只有仍属于当前 generation 的响应才能写入运行时投影。必须增加延迟响应的竞态回归测试。

### P1. IAM 删除接口绕过正式账户生命周期

位置：[administration/service.go](/D:/Document/code/go-admin-plus/go-admin-plus/internal/modules/iam/administration/service.go:300)

`DeleteUser` 和 `DeleteUsers` 直接执行 `DELETE FROM iam_accounts`。正式删除流程在 `deletion.go` 中通过 deletion 记录、outbox 事件和 Files 消费者处理文件迁移/清理；这两个 service 入口不创建生命周期记录、不发事件，也不执行文件对象和容量计数清理。

调用后会遗留 `files_objects`、容量计数和本地对象文件，并绕过删除恢复、最后管理员保护等状态约束。当前 HTTP 路由未直接暴露不改变风险，因为它们仍是生产可调用的破坏性 API。新发布应删除这两个旧方法及其测试/调用面，所有删除统一进入 `StartDeletion`；短期不能删除时也必须改为明确返回冲突错误，不能继续物理删除。

### P1. 文件容量探针使用虚拟容量，无法防止真实磁盘耗尽

位置：[quota.go](/D:/Document/code/go-admin-plus/go-admin-plus/internal/modules/files/quota.go:86)、[service.go](/D:/Document/code/go-admin-plus/go-admin-plus/internal/modules/files/service.go:44)

未注入探针时，`NewService` 安装 `configuredCapacityProbe`，固定返回“总容量 = 最大配额 + 最低剩余空间，且全部可用”，完全不读取存储根目录的真实剩余空间。代码注释也明确它是等待 T-09 的兼容临时实现。

因此数据库预留通过并不代表写盘可用，磁盘耗尽时会在 staging 阶段失败，留下恢复/临时文件压力并形成请求级资源消耗。发布前应实现并强制注入基于存储根目录的真实平台探针；探针不可用时拒绝上传并返回 `ErrDiskCapacity`，删除该虚拟兼容实现和 T-09 注释。

### P1. 开发文档仍指向已删除命令

位置：[go-admin-plus/README.md](/D:/Document/code/go-admin-plus/go-admin-plus/README.md:3)、[database/README.md](/D:/Document/code/go-admin-plus/database/README.md:5)、[repository-architecture.md](/D:/Document/code/go-admin-plus/docs/repository-architecture.md:10)

文档仍把 `cmd/config-check`、`cmd/migrate` 描述为正式入口；实际命令目录只有 `cmd/go-admin-plus` 和 `cmd/desktop-sidecar`，迁移是统一 CLI 的 `migrate` 子命令。新开发者会按不存在的路径操作，且与项目“零兼容”决策冲突。

应统一改为 `cmd/go-admin-plus` 的 `migrate/bootstrap/recover-admin/serve/worker` 子命令，并给文档门禁加入已删除路径断言。

### P1. 架构门禁允许旧命令目录重新进入

位置：[architecture-check.mjs](/D:/Document/code/go-admin-plus/scripts/quality/architecture-check.mjs:325)

命令允许集合仍包含 `config-check` 和 `migrate`，而后端架构测试明确要求这两个目录不存在。当前门禁 PASS 只是因为旧目录被列入 allowlist；一旦目录被重新添加，门禁不会失败。

应将允许集合收敛为当前真实入口，并增加回归测试验证旧目录出现时必须失败。

### P1. Compose 生产镜像没有逐项强制 digest

位置：[compose.yml](/D:/Document/code/go-admin-plus/deploy/compose/compose.yml:3)、[verify-policy.mjs](/D:/Document/code/go-admin-plus/scripts/release/linux/verify-policy.mjs:8)

部署文档要求生产使用固定镜像摘要，但 API/Web 默认仍是可变 tag（例如 `go-admin-plus-server:dev`、`go-admin-plus-web:dev`）。发行校验器只检查 Compose 文本任意位置存在一个 `@sha256:`，没有分别约束 API、Web 和 PostgreSQL 镜像。

生产入口可以因此绕过可复现部署和供应链校验。生产 Compose 应要求 API/Web/PostgreSQL 都使用完整 `image@sha256:<64 hex>`，或在执行前拒绝非 digest 引用；校验器必须逐项检查。

## P2 问题与结构整改

1. **数据范围存在第二事实源**：迁移 `iam_role_data_scopes`，但运行时全部读写 `iam_roles.data_scope`。位置：`go-admin-plus/internal/modules/iam/migrations/0050-data-scope/`、`internal/modules/iam/authorization/service.go:97-124`。应保留一个唯一模型，删除未使用表/迁移，或完成全量切换后删除旧列。
2. **Audit 仍接受 legacy payload**：`audit/service.go:391-411` 接受 `action/reason` 并在 `Source` 为空时推断离线来源。新发布应移除 fallback，只接受当前 payload schema；旧数据要在迁移阶段显式转换或进入人工处理状态。
3. **前端五套 Web client 重复 transport policy**：IAM、Audit、Scheduler、Demo、Files 各自实现 CSRF、错误分类和请求串行队列，且 token 校验不一致。应在 browser adapter/api-client 提供统一 session-aware transport，业务 client 只做契约调用和领域解包。
4. **IAM/Audit 自制弹窗缺少完整焦点约束**：`AdministrationPage.vue` 和 `AuditPage.vue` 使用 `role="dialog" aria-modal="true"` 的普通 backdrop，没有完整 focus trap、初始焦点和关闭后的焦点恢复。应复用共享 `FormDialog` 或共享 dialog composable，并补键盘/读屏测试。
5. **Linux service archive 与文档不一致**：`docs/release.md` 和 `release/README.md` 声称归档包含 Compose，但 `build-service.sh:21-31` 只复制 profile、systemd unit 和安装说明。应删除该承诺，或实际打包并验证完整 Compose 目录。
6. **macOS 数据路径文档冲突**：总览写成安装目录下 `data/`，平台文档和实现写成 `.app/data/`。备份指引可能漏备份数据库、文件和 Stronghold。应按平台统一路径说明。
7. **Windows 开发前置缺少 POSIX shell**：`run-script.mjs` 在 Windows 依赖 `sh.exe`，但开发指南没有 Git Bash/MSYS2 或 `GO_ADMIN_POSIX_SHELL` 前置说明。应补齐环境合同，或提供 PowerShell 等价入口。
8. **服务端 HTTP resilience 配置不完整**：生产 Server 设置了 `ReadTimeout/WriteTimeout`，但没有明确 `ReadHeaderTimeout`、`IdleTimeout` 和受控的 `MaxHeaderBytes`。这不是已确认的越权漏洞，但会放大慢连接和空闲连接资源消耗，建议按部署流量补充并增加宿主测试。
9. **文档门禁覆盖不完整**：`docs-check.mjs` 未扫描根 `CHANGELOG/NOTICE/CLAUDE`、`release/**/INSTALL.md`、`release/manifest`、`release/shared/sidecar` 和 UI 子包 README。应以 `git ls-files '*.md'` 为主集合，仅显式排除 Speculo 模板/示例并记录原因。

## 架构与调用流评估

目标调用流整体正确：

```text
CLI / Tauri host
  -> product.Build / BuildPrepared
  -> application module lifecycle
  -> host HTTP / desktop IPC adapter
  -> module transport
  -> service/use case
  -> repository + platform ports
  -> SQLite / PostgreSQL
```

后端组合根、模块所有权、OpenAPI 生成边界、双方言迁移和前端 `app -> adapter -> app-shell -> web-domain -> headless domain -> api-client` 方向均符合既定架构。主要偏差集中在：账户删除有绕过正式工作流的旧 service 入口；数据范围有双表事实源；前端 transport policy 没有真正共享；命令平面和文档没有同步收敛。这些问题会让调用流在边界处出现第二条路径，降低可证明性和复用性。

## 清理建议

以下内容可在确认没有待归档证据后清理：

- 删除 `DeleteUser/DeleteUsers` 及相关旧测试/调用面。
- 删除 `configuredCapacityProbe` 和所有 T-09 兼容注释。
- 删除 `audit` legacy payload fallback、对应 compatibility 测试和历史字段说明。
- 删除未使用的 `iam_role_data_scopes` 迁移/测试，或完成唯一事实源切换后删除旧列。
- 清理未被 Speculo 状态引用的 `specdev-worktree/` 临时工作树；它未被 Git 跟踪且被整体忽略，但删除前应确认没有未归档审查证据。
- 清理 `.artifacts/`、`.data/` 等本地运行产物；不要把它们当作产品源文件或发布资产。
- 删除所有旧命令、旧路径、旧兼容说明；保留必要的许可证和来源声明，不将 NOTICE 中的第三方归属误删。

本次 review 没有直接删除上述资产，以避免误删本地运行状态或 Speculo 临时证据；报告给出了明确的删除条件和路径。

## 验证证据

- 通过：`task architecture:check`
- 通过：`task compatibility:zero`
- 通过：`task docs:check`
- 通过：`task contract:lint`
- 通过：后端模块、平台、IAM/Files/Scheduler 测试和 `go vet ./...`
- 通过：前端 `pnpm lint`、`pnpm typecheck`、`pnpm check:workspace`、`pnpm build`
- `task test` 默认并发执行时，29 个文件/190 个测试通过，但 Vitest worker 启动超时导致命令非零退出；前端子审查使用 `--maxWorkers=1` 重跑为 30 个文件/213 个测试全部通过。应在 CI 固定 worker 并发或调整启动超时，避免门禁因并发不稳定而误报。
- 直接在仓库根运行 `go test ./...` 会因根目录没有 `go.mod` 失败；正确的后端命令由根 Task 脚本切换到 `go-admin-plus/` 后执行。该差异应在贡献文档中明确，避免误用根级 Go 命令。

详细分项证据见：

- [backend-findings.md](./backend-findings.md)
- [frontend-findings.md](./frontend-findings.md)
- [docs-repo-findings.md](./docs-repo-findings.md)
