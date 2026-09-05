# 后端 Review 子报告

审查范围：`go-admin-plus/`、`contracts/openapi/`、`database/`、`deploy/`。本报告只记录有代码证据的问题；未发现的问题不会以猜测形式列出。

## Findings

### P1：管理员删除的旧入口绕过账户生命周期和文件清理工作流

- 文件：`go-admin-plus/internal/modules/iam/administration/service.go:300-380`
- 证据：`DeleteUser`/`DeleteUsers` 直接执行 `DELETE FROM iam_accounts`。当前正式生命周期实现位于 `deletion.go`，通过 `iam_account_deletions`、`files` 消费者和 outbox 事件完成 transfer/purge；这两个入口不创建 deletion 记录、不发事件、不调用文件清理。
- 影响：调用这些 service 方法会留下 `files_objects`、容量计数和本地对象文件；同时绕过 `deletion-pending/deleted`、恢复阻断和最后管理员保护。即使当前 HTTP 路由未暴露它们，生产代码仍保留可调用的破坏性旧 API，任何适配器/未来调用都可能造成不可恢复的数据与配额泄漏。
- 建议：删除这两个旧方法及其测试/调用面；所有删除统一调用 `StartDeletion`，并由生命周期 worker 完成。若短期必须保留，至少改为返回明确的 `ErrConflict`，不能执行物理删除。

### P1：文件容量检查默认使用虚拟容量，实际磁盘耗尽不会被预防

- 文件：`go-admin-plus/internal/modules/files/quota.go:86-92`，调用点 `go-admin-plus/internal/modules/files/service.go:78-87`
- 证据：`NewService` 在未提供探针时安装 `configuredCapacityProbe`；该实现固定返回 `TotalBytes = MaximumTotalBytes + MinimumAvailableBytes`、`AvailableBytes = TotalBytes`，与宿主文件系统真实可用空间无关。注释也明确这是 “compatibility default until T-09”。
- 影响：上传可持续预留并写入最多 10 GiB（默认策略），即使数据目录只有很少剩余空间；系统会在写盘阶段失败，造成大量临时文件/恢复任务、请求级 DoS 和运维不可预测性。容量策略表面存在但不能履行“最低可用空间”约束。
- 建议：发布前实现并强制注入基于存储根目录的真实 `Statfs`/平台探针；探针不可用时拒绝上传并报告 `ErrDiskCapacity`，不要安装虚拟兼容实现。移除 T-09 兼容注释及默认实现。

### P2：数据范围迁移创建了运行时完全未使用的第二份事实源

- 文件：`go-admin-plus/internal/modules/iam/migrations/0050-data-scope/{sqlite,postgres}/7110000000000_iam_data_scope.sql:2-8`
- 证据：迁移创建并填充 `iam_role_data_scopes`。运行时授权查询 `go-admin-plus/internal/modules/iam/authorization/service.go:97-124`、角色 CRUD `go-admin-plus/internal/modules/iam/administration/service.go:431-481` 全部读写 `iam_roles.data_scope`，全仓库除该迁移及其测试外没有任何 `iam_role_data_scopes` 引用。
- 影响：数据库同时保留两个看似权威的 scope 表，后续手工修复/运维查询可能修改未生效的表；迁移和代码契约不一致，增加审计和数据范围回归风险，并违反当前“单一所有权/无旧兼容结构”要求。
- 建议：确认 `iam_roles.data_scope` 为当前唯一模型后，移除该迁移及 provider/test（新发布无需历史兼容）；或反向完成迁移，把所有查询/写入切换到新表并删除旧列，不能两者并存。

### P2：审计当前存在两套 payload 协议，生产写入绕过正式 TopicSchema

- 文件：`go-admin-plus/internal/modules/audit/service.go:391-411`
- 证据：`audit/service.go:391-411` 的 `present` 仍接受 `action`、`reason` 字段，并在特定 `resource:iam_bootstrap`/`resource:iam_recovery` 记录上推断 `SourceServer`；代码注释标注为 `legacy offline bootstrap/recovery records`。同时 `app/product/operations.go:121-132` 的当前 bootstrap/recovery 生产者仍直接写入这些字段，而 `audit.TopicSchemas()` 只声明 `source`，TransactionalConsumers 只处理正式 outbox payload。
- 影响：同一张 `audit_facts` 表同时接受正式合同和旁路格式，合同 lint 无法覆盖旁路写入；读取端的宽松兼容掩盖了生产者没有遵循统一事件协议的问题，也使后续审计字段校验、重放和迁移行为不一致。
- 建议：把 bootstrap/recovery 事件改为当前统一的 outbox/topic schema（必要字段在正式 schema 中声明），再删除 `action`/`reason` fallback 及其兼容测试；不能继续依赖读取端兼容来维持旁路写入。

## 复用性/架构观察（非阻断）

- 各模块 HTTP 层均重复实现 cookie 提取、CSRF/认证错误映射和 problem response 组装（例如 files/demo/scheduler）。共享适配器已有部分基础能力，但模块仍保留近乎同构代码，后续错误码/安全头修复容易漏改。建议在 transport 边界抽取一个只负责响应映射的内部适配器，保持模块不互相依赖。
- `contracts/openapi` 合同 lint、后端生成代码一致性检查、架构检查和兼容性门禁均通过；未发现 OpenAPI 中可直接利用的未授权 mutation 路由。

## 验证记录

- `go test ./internal/modules/... ./internal/platform/...`：通过。
- `go test ./test/iam/... ./test/files/... ./test/scheduler/...`：通过。
- `go vet ./...`：通过。
- `task contract:lint`、`task architecture:check`、`task compatibility:zero`、`task docs:check`：通过（门禁未覆盖上述运行时语义问题）。
