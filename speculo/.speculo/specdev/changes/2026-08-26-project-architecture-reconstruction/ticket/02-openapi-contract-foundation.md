---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-02
title: OpenAPI 合同与双方生成基座
status: in_progress
planning_depth: deep
planning_depth_reason: 公共 API、错误语义和双方生成物是所有垂直切片共享的外部合同
ready: true
risk: high
blocked_by: [T-01]
contract_ids: [AC-023, AC-025]
owner: codex-root
expected_changes: ["<Path>Taskfile.yml</Path>", "<Path>contracts/openapi/openapi.yaml</Path>", "<Path>contracts/openapi/components/**</Path>", "<Path>scripts/contracts/**</Path>", "<Path>scripts/go-admin-plus/task-contract.sh</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus-ui/packages/api-client/**</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path>"]
writable_paths: ["<Path>Taskfile.yml</Path>", "<Path>contracts/openapi/openapi.yaml</Path>", "<Path>contracts/openapi/components/**</Path>", "<Path>contracts/openapi/modules/_template.yaml</Path>", "<Path>scripts/contracts/**</Path>", "<Path>scripts/go-admin-plus/task-contract.sh</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus-ui/packages/api-client/**</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path>"]
read_only_paths: ["<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>", "<Path>go-admin-plus/api/openapi/**</Path>", "<Path>go-admin-ui-plus/packages/contracts/**</Path>"]
shared_paths: ["<Path>Taskfile.yml</Path>", "<Path>contracts/openapi/openapi.yaml</Path>", "<Path>contracts/openapi/components/**</Path>", "<Path>scripts/contracts/**</Path>", "<Path>scripts/go-admin-plus/task-contract.sh</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus-ui/packages/api-client/**</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path>"]
shared_path_owners: ["<Path>Taskfile.yml</Path> => T-02 (contract:lint and generate:check only; serialized after T-01)", "<Path>contracts/openapi/openapi.yaml</Path> => T-02", "<Path>contracts/openapi/components/**</Path> => T-02", "<Path>scripts/contracts/**</Path> => T-02", "<Path>scripts/go-admin-plus/task-contract.sh</Path> => T-02 (append contract:lint and generate:check to the exact public task set only)", "<Path>go-admin-plus/go.mod</Path> => T-02 (contract generator and generated transport dependencies only)", "<Path>go-admin-plus/go.sum</Path> => T-02 (same scope)", "<Path>go-admin-plus/internal/contracts/**</Path> => T-02", "<Path>go-admin-plus-ui/packages/api-client/**</Path> => T-02", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-02 (contract tooling/client dependencies only)", "<Path>go-admin-plus-ui/pnpm-workspace.yaml</Path> => T-02 (exact @redocly/cli release-age allowlist only)"]
---

# Ticket T-02: OpenAPI 合同与双方生成基座

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/02-openapi-contract-foundation.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-02.md</Path>`

## 1. 战略与来源

- **目标：** 建立模块化 OpenAPI 3.1 唯一事实源及确定性的 Go/TypeScript transport 生成接缝。
- **可观察产出：** 合同可 lint、bundle、generate、conformance，公共失败返回稳定错误类别。
- **来源：** `US-015`、`AC-023`、`AC-025`、`ADR-015`、`DEC-007`。
- **当前事实：** 旧 Swagger、手写 Go handler DTO 与前端合同并存且会漂移。
- **Planning Depth 原因：** 该公共协议被全部模块消费，错误会扩散到双方应用。

## 2. 决策状态

### 已锁定决策

- 模块拥有自己的 OpenAPI fragment；T-02 唯一拥有公共组件、根入口和生成工具。
- 生成类型只用于 transport，领域模型通过显式 mapping 隔离。

### 已采用的低影响假设

- 生成器版本在工具配置中固定，输出按模块落入 owner 路径。

### 已批准 Ticket 偏差

- `T02-D01`：原路径合同无法实现已批准的 `task contract:lint generate:check` 和固定官方生成器。Lead/Ticket owner 于 2026-08-26 批准最小扩展：T-02 只新增这两个根任务及其精确 task-set 断言，只写入 oapi-codegen/生成 transport 所需的 Go module 记录、合同工具/客户端所需的 pnpm lock，并仅为实际 registry 当前版 `@redocly/cli@2.48.0` 增加精确 release-age allowlist；不改变 Spec、ADR、产品 API、安全策略或后续 owner 的其他职责。
- `T02-D02`：首轮双轴审查证明 fragment 模板仍指向重构前 UI 目录，且不能表达 IAM 同一 owner 下的多个 fragment。Lead/Ticket owner 于 2026-08-26 批准只在既有模板、fixture 和合同工具路径内引入独立 fragment id 与显式 owner，并把模块 TypeScript 生成目标声明为 Spec 权威的 `<Path>go-admin-plus-ui/packages/domains/{owner}/src/**</Path>`；T-02 不写未来模块源码。
- `T02-D03`：首轮双轴审查证明公开 `generate` 仍只运行施工期旧生成链，实测旧后端链还依赖未受管且缺失的 `swag` 可执行文件。Lead/Ticket owner 于 2026-08-26 批准公开 `generate` 只委派 canonical OpenAPI generator；旧脚本只作为 expand 施工 inventory 保持只读并由 T-21 物理删除，不再属于公共命令面或产品兼容合同。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 公共组件、错误信封、生成与漂移检查 | 旧合同仅用于能力盘点 | 具体业务操作和旧 API 兼容 |

## 4. 要构建什么

开发者新增模块 fragment 后，可由根任务完成校验和双方生成；无效合同、生成漂移或实现不符合 strict interface 时门禁失败，调用者只看到声明的稳定错误类别。

## 5. 实现契约

- **入口或接缝：** OpenAPI root、模块 fragment 模板、合同 CLI、Go strict interface、TS client。
- **输入与输出：** OpenAPI 3.1 YAML 输入；确定性 bundle 和模块内 transport 输出。
- **公共接口变化：** 建立全新分页、校验、认证、授权、not-found、conflict、internal 错误合同。
- **不变量：** clean tree 生成零漂移；transport 类型不进入领域层；错误不泄露内部细节。
- **状态或数据流：** fragment -> lint/bundle -> Go/TS generate -> conformance。
- **错误与失败行为：** 非法 schema、重复 operationId、未声明响应或漂移必须阻断。
- **兼容要求：** 不保留旧 Swagger 路由、字段或生成物兼容。
- **安全与隐私要求：** 公共错误禁止 SQL、stack、路径、secret 和原始 Session。

## 6. 执行路线

1. 用合同测试固定模块边界、错误类别和确定性要求。
2. 建立 OpenAPI root、公共 components 和 fragment 模板。
3. 实现固定版本的 lint、bundle、Go/TS generate 与 drift 检查。
4. 用最小探针实现 strict server 与 TS client conformance。
5. 运行负向错误与 clean-tree 回归。

## 7. 路径访问契约

- **预计修改点：** 公共合同、生成工具和双方共享 transport 包。
- **可写范围：** 仅 frontmatter `writable_paths`；不得写业务模块 fragment。
- **只读上下文：** 根任务及旧合同能力清单。
- **共享路径：** 公共合同与生成器由 T-02 唯一拥有，模块 Ticket 只消费。
- **保留或不动：** 产品聚合合同 `<Path>contracts/openapi/product.yaml</Path>` 由 T-17 拥有。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 合同生成 | `task contract:lint generate:check` | bundle 有效、双方输出确定且无 drift | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-02.md</Path>` |
| 失败路径 | 负向 fixture | 注入非法 schema、重复 operationId 和内部错误细节 | 三类问题均阻断并给出定位 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-02.md</Path>` |
| 回归 | conformance 探针 | 编译 strict Go 接口和 TS client | transport 可用且未进入领域层 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-02.md</Path>` |

- **Workspace checks：** 按 Goal Plan 在 current-workspace 或 source-worktree 运行合同、编译和静态检查。
- **E2E disposition：** not-required：本 Ticket 的跨边界风险由生成与 conformance 合同完整覆盖。
- **E2E owner/environment：** Lead / current-workspace，确认无需运行产品 E2E。
- **Integration evidence：** 记录 implementation/source commit、parent before、direct-parent 或 candidate/result SHA 及父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 公共合同先落地，模块随后各自增加 fragment，T-17 最后聚合。
- **兼容窗口：** 无产品兼容窗口；施工期旧合同保持只读直到 T-21 删除。
- **监控信号：** lint、生成漂移、strict interface 与负向响应 conformance。
- **回滚或前向恢复：** 在首个消费者前可回滚；之后通过修复生成器或 fragment 前向恢复。
- **不可逆操作与批准点：** 无删除操作。
- **收缩条件：** T-21 证明旧 Swagger 与手写共享 transport 零引用。

## 10. 验收标准

- [ ] `AC-023`：OpenAPI 3.1 可确定性生成 Go strict transport 和 TS client，clean tree 无漂移。
- [ ] `AC-025`：公共负向响应只使用已声明错误类别且不泄露内部信息。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-02.md</Path>`。
- [ ] 修改未超出 `writable_paths`，共享路径仅由 T-02 修改。
- [ ] 形成非空 implementation/source commit，并记录 direct-parent 或 candidate/result SHA。
- [ ] 未发生未批准偏差，Ticket、Map 和 Evidence 状态一致。
