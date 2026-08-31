---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-08
title: 汇合公共 OpenAPI 合同与 HTTP Adapters
status: done
planning_depth: deep
planning_depth_reason: 原子改变共享 OpenAPI、生成 wire 类型、认证端点和破坏性账号删除公共接口
ready: true
risk: critical
blocked_by: [T-02, T-03, T-04, T-05, T-06]
contract_ids: [AC-009, AC-010, AC-011, AC-012, AC-013, AC-019, AC-020, AC-021, AC-022, AC-023, AC-024, AC-025, AC-026, AC-027, AC-028, AC-029, AC-038]
owner: codex-root
expected_changes: ["<Path>contracts/openapi/**</Path>", "<Path>go-admin-plus/internal/app/product/registry_test.go</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus/internal/modules/iam/session/http.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/service.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/transport/**</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0040-session-protection/**</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/http*</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/transport/**</Path>", "<Path>go-admin-plus/internal/modules/files/http.go</Path>", "<Path>go-admin-plus/internal/modules/files/transport/**</Path>", "<Path>go-admin-plus/test/audit/audit_test.go</Path>", "<Path>go-admin-plus/test/iam/authorization/http_test.go</Path>", "<Path>go-admin-plus/test/iam/session/session_test.go</Path>", "<Path>go-admin-plus-ui/packages/api-client/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/audit/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/files/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/scheduler/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/*/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/administration/administration-controller*</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/administration/index.ts</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/administration/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/organization/browser-driver.ts</Path>", "<Path>go-admin-plus-ui/tests/e2e/scheduler/browser-driver.ts</Path>", "<Path>go-admin-plus-ui/tests/e2e/settings/browser-driver.ts</Path>", "<Path>scripts/contracts/generated/**</Path>"]
writable_paths: ["<Path>contracts/openapi/**</Path>", "<Path>go-admin-plus/internal/app/product/registry_test.go</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus/internal/modules/iam/session/http.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/service.go</Path>", "<Path>go-admin-plus/internal/modules/iam/session/transport/**</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0040-session-protection/**</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/http*</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/transport/**</Path>", "<Path>go-admin-plus/internal/modules/files/http.go</Path>", "<Path>go-admin-plus/internal/modules/files/transport/**</Path>", "<Path>go-admin-plus/test/audit/audit_test.go</Path>", "<Path>go-admin-plus/test/iam/authorization/http_test.go</Path>", "<Path>go-admin-plus/test/iam/session/session_test.go</Path>", "<Path>go-admin-plus-ui/packages/api-client/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/audit/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/files/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/scheduler/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/src/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/*/generated/**</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/administration/administration-controller*</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/administration/index.ts</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/administration/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/organization/browser-driver.ts</Path>", "<Path>go-admin-plus-ui/tests/e2e/scheduler/browser-driver.ts</Path>", "<Path>go-admin-plus-ui/tests/e2e/settings/browser-driver.ts</Path>", "<Path>scripts/contracts/generated/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/bootstrap/**</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/deletion*</Path>", "<Path>go-admin-plus/internal/modules/files/quota*</Path>", "<Path>scripts/contracts/cli.mjs</Path>"]
shared_paths: ["<Path>contracts/openapi/**</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus-ui/packages/api-client/src/generated/**</Path>", "<Path>scripts/contracts/generated/**</Path>"]
shared_path_owners: ["<Path>contracts/openapi/**</Path> => T-08", "<Path>go-admin-plus/internal/contracts/**</Path> => T-08", "<Path>go-admin-plus-ui/packages/api-client/src/generated/**</Path> => T-08", "<Path>scripts/contracts/generated/**</Path> => T-08"]
---

# Ticket T-08: 汇合公共 OpenAPI 合同与 HTTP Adapters

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/08-public-contract-integration.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-08.md</Path>`

## 1. 战略与来源

- **目标：** 把已完成的安全/数据模块用例原子发布为唯一 OpenAPI 合同和生成 transport。
- **可观察产出：** heartbeat/renew、数据范围、账号 deletion/status/cancel 与容量错误可经 HTTP 调用；旧账号 DELETE/batch-delete 消失。
- **来源：** `AC-009~013`、`AC-019~029`、`AC-038`、永久 `ADR-0015`、当前 `DEC-007/010`。
- **当前事实：** canonical OpenAPI 和所有 Go/TS 生成物是高冲突共享路径，必须由单一 Ticket 所有。
- **Planning Depth 原因：** 认证与不可逆删除的 wire contract 是跨端共享核心，且零兼容收缩必须原子完成。

## 2. 决策状态

### 已锁定决策

- 增加 `/iam/session/heartbeat`、`/iam/session/renew`。
- 用单账号 deletion POST/GET/cancel 替换用户 DELETE 与 batch-delete。
- 用户/角色增加组织/范围字段，Files 增加稳定容量 Problem；视觉改造不改其他业务 API。

### 已采用的低影响假设

- operationId 使用现有模块命名惯例；Problem code 采用稳定大写领域代码且不暴露内部状态。

### 已批准执行偏差 DEV-08-001

- **等级：** ticket path ownership；用户于 `2026-09-01T01:36:14+08:00` 以“都批准”明确批准。
- **触发事实：** breaking OpenAPI 生成后，T-14-owned IAM consumer 仍调用已删除的单删与 batch-delete 路径，且旧 controller 把角色范围限制为 `all/self`；`pnpm --dir go-admin-plus-ui typecheck` 因此失败。T-08 必须先完成、T-14 又阻塞于 T-08/T-09/T-12，原路径分工形成无法关闭的 consumer pause。
- **批准范围：** T-08 临时拥有 IAM administration domain controller contract、Web Domain administration consumer 和 Organization E2E driver 的 scope type adapter，只做旧删除调用归零、生成 DTO 适配和使候选可编译；不得提前实施 T-14 的 route/page 视觉重构或改变 E2E 流程。
- **协调结果：** T-14 在 T-08 result 后基于新生命周期和五态范围继续完整交互；T-08 Gate 在 consumer scan、typecheck 和回归通过前保持关闭。

### 已批准执行偏差 DEV-08-002

- **等级：** ticket + migration/release seam；用户于 `2026-09-01T01:57:08+08:00` 以“都批准”明确批准。
- **触发事实：** T-03 只保存 `csrf_hash`，但 DEC-007/AC-010 要求只读 `current` 在零 Session 写入下返回稳定 family CSRF；plaintext 不可由 hash 恢复，且 `AuthorizeRequest` 因此向 IAM adapters 返回空 CSRF。T-08 全量 Go 回归稳定复现空响应头。
- **批准范围：** 在 0040 provider 内新增双方言前向 migration，为新 Session 保存可返回的 family CSRF，统一撤销无法恢复值的旧活动 Session；修正 Session service 投影与旧 rotate-on-read HTTP 测试。不得恢复 GET 写入、滚动 CSRF 或兼容旧活动 Session。
- **恢复与 Gate：** migration 是 forward-only；旧 Session 按既定零兼容策略重新登录。SQLite、PostgreSQL migration、只读零写入、HTTP header 与 secret scan 未通过前不得关闭 T-08。

### 已批准执行偏差 DEV-08-003

- **等级：** ticket regression ownership；用户已以“都批准”授权全部必要外部条件。
- **触发事实：** 前向 CSRF migration 使产品 migration matrix 从 16 增至 17；稳定 CSRF 合同使 Audit 真实 IAM adapter 的旧 rotate-on-read 断言失效。两处均由 `go test ./...` 与 `go test -tags sqlite ./...` 稳定复现。
- **批准范围：** T-08 临时拥有 `<Path>go-admin-plus/internal/app/product/registry_test.go</Path>` 与 `<Path>go-admin-plus/test/audit/audit_test.go</Path>`，只更新 migration 总数与稳定 CSRF 断言，不改产品 composition 或 Audit 实现。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| canonical fragments/product、生成物、IAM/Files HTTP adapters、conformance tests | T-02~06 use cases、现有生成器 | 产品 composition（T-09）、前端手写 domain/UI、兼容端点 |

## 4. 要构建什么

调用者只能从 canonical OpenAPI 发现新认证、数据范围、账号生命周期和容量行为。HTTP adapter 做输入映射、CSRF/认证、稳定 Problem 和响应 header，不复制领域决策。生成 Go/TypeScript transport 与模块 manifest 一次更新；旧账号删除调用点和路径为零。

## 5. 实现契约

- **入口或接缝：** `<Path>contracts/openapi/product.yaml</Path>` 与模块 fragments、strict handlers。
- **输入与输出：** Spec 第 7 节公开 paths/DTO/Problems；生成 transports 为唯一 wire 类型。
- **公共接口变化：** 按 Spec 全部发布，账号删除为破坏性替换。
- **不变量：** handler 只适配 use case；current GET 只读；错误不泄漏账号/容量/删除内部细节。
- **状态或数据流：** HTTP -> generated validation -> adapter -> T-02~06 use case -> generated response。
- **错误与失败行为：** validation/authentication/authorization/conflict/capacity 使用一致 Problem 和 trace。
- **兼容要求：** 零兼容；旧 DELETE/batch-delete、滚动 CSRF 描述和生成调用点全部删除。
- **安全与隐私要求：** OpenAPI 示例、日志和 responses 不含 secret；CSRF header 描述更新为家族稳定。

## 6. 执行路线

1. 建立新 paths/operationId/Problem 与旧路径缺失的 contract tests。
2. 更新 canonical fragments/product 并一次生成 Go/TS outputs。
3. 实现 strict HTTP adapters，调用既有模块 use cases。
4. 扫描并删除旧账号删除和 rotate-on-read wire 调用点。
5. 运行 lint、generate drift、handler contract、Go/TS typecheck 和回归。

## 7. 路径访问契约

- **预计修改点/可写范围：** canonical OpenAPI、全部受管生成物和指定 IAM/Files HTTP adapters。
- **只读上下文：** T-02~06 模块用例与生成器实现。
- **共享路径：** frontmatter 四组路径均由 T-08 唯一拥有；其他 Ticket 只消费。
- **保留或不动：** 不手改生成文件，不修改业务用例或 UI。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | HTTP contracts | 定向 IAM/Files transport tests | 新端点/DTO 调用模块用例 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-08.md</Path>` |
| 失败路径 | Problem/旧路径 | invalid/unauth/conflict/capacity + 旧路径扫描 | 稳定错误；旧路径不存在 | 同上 |
| 回归 | canonical/generation | `task contract:lint && task generate:check` + Go/TS tests | lint、生成和类型 clean | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行合同 lint、生成、Go transport、TS typecheck/unit、架构检查。
- **E2E disposition：** not-required：HTTP contract integration 直接证明 wire 行为；真实浏览器/native E2E 集中在 T-19/T-20。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source 不运行 E2E。
- **Integration evidence：** source/implementation commit、direct-parent/candidate/result SHA、生成 diff 与 Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-02~06 已集成 -> canonical/生成/adapter 原子切换 -> T-09 composition -> 客户端迁移。
- **兼容窗口：** 无产品兼容窗口；工作分支内只有本 Ticket 的原子 commit 可包含暂态。
- **监控信号：** 新 operation status/problem code、旧路径 404、Session write rate、deletion/capacity conflicts。
- **回滚或前向恢复：** 数据 schema 已 forward-only，合同问题优先前向修复；旧客户端不受支持。
- **不可逆操作与批准点：** 删除旧 API 前扫描调用点并由 Lead 批准；不执行实际 purge。
- **收缩条件：** 旧 OpenAPI paths、生成 methods、handlers 和 UI calls 扫描为零。

## 10. 验收标准

- [x] frontmatter 所列 AC 的公开合同和失败语义成立。
- [x] 旧账号 DELETE/batch-delete 与滚动 CSRF 描述/调用为零。
- [x] contract lint、generate check、Go/TS 验证写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-08.md</Path>`。
- [x] shared owner、commit、candidate/direct-parent、result 和 E2E disposition 完整。
