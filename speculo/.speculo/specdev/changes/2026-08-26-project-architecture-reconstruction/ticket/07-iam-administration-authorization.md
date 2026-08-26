---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-07
title: IAM 管理与 Permission Code 授权闭环
status: in_progress
planning_depth: deep
planning_depth_reason: 权限、数据范围、管理 schema 和即时授权决定全部受保护业务行为
ready: true
risk: critical
blocked_by: [T-06]
contract_ids: [AC-011, AC-013, AC-035, AC-036]
owner: codex-t07-iam
expected_changes: ["<Path>contracts/openapi/modules/iam-administration.yaml</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/**</Path>", "<Path>go-admin-plus/internal/modules/iam/authorization/**</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/administration/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/administration/**</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/iam-administration.yaml</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/**</Path>", "<Path>go-admin-plus/internal/modules/iam/authorization/**</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0020-administration-*</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/administration/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/administration/**</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus/test/iam/authorization/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/iam/administration/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/session/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/core/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-07 under T07-D01; before T-11"]
---

# Ticket T-07: IAM 管理与 Permission Code 授权闭环

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/07-iam-administration-authorization.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-07.md</Path>`

## 1. 战略与来源

- **目标：** 交付用户、角色、菜单、API Permission Code 和数据范围的管理与后端最终授权。
- **可观察产出：** 管理员可维护授权；普通用户只看到获权能力，直接越权 API 被拒绝且无状态变化。
- **来源：** `US-006`、`AC-011`、`AC-013`、`AC-035`、`AC-036`、`ADR-018`。
- **当前事实：** 旧 Casbin/菜单可见性/tenant 数据范围分散，无法形成稳定 Permission Code 合同。
- **Planning Depth 原因：** 授权错误会造成全产品越权或拒绝服务。

## 2. 决策状态

### 已锁定决策

- 稳定 Permission Code + Role + Data Scope 替代 Casbin；前端可见性不是授权边界。
- 权限变更从数据库事实即时生效，缓存清空不影响正确性。

### 已采用的低影响假设

- 保护对象和系统角色使用稳定业务键，冲突返回公共 conflict 类别。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| IAM 管理、授权策略、菜单投影和页面 | T-06 身份/Session，T-05 UI | Organization 实现、Casbin 兼容 |

## 4. 要构建什么

管理员维护用户与角色并分配菜单、Permission Code 和数据范围；请求进入每个受保护用例前以后端授权 Port 判定，前端仅据相同 manifest 控制可见性。

## 5. 实现契约

- **入口或接缝：** IAM admin use cases、Authorization Port、manifest projection、Web management pages。
- **输入与输出：** CRUD/assignment 命令与身份上下文；返回列表/详情或稳定拒绝。
- **公共接口变化：** 新 IAM 管理 API、Permission Code 和 data-scope contract。
- **不变量：** 后端每次最终判定；权限改变即时生效；保护对象不可误删；无 tenant 维度。
- **状态或数据流：** Session identity -> role/permission/scope query -> use case -> repository transaction -> manifest refresh。
- **错误与失败行为：** 缺权、越界、引用冲突和保护对象操作失败且不改变状态。
- **兼容要求：** 不导入旧 Casbin policy、JWT claims、菜单 ID 或 tenant scope。
- **安全与隐私要求：** 列表与详情也受数据范围；重置密码不返回明文持久值。

## 6. 执行路线

1. 建立权限/数据范围矩阵和保护对象失败测试。
2. 增加 IAM administration schema、repository 和 Authorization Port。
3. 实现 API mapping、manifest projection 和管理端页面。
4. 覆盖即时权限变更、缓存禁用、直接 API 越权和重复命令。
5. 运行双方言 API 与浏览器 E2E。

## 7. 路径访问契约

- **预计修改点：** IAM administration/authorization、对应前端子域，以及经 `T07-D01` 批准的两个 IAM package manifest。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** T-06 Session、T-05 Shell/UI。
- **共享路径：** `T07-D01` 令 T-07 在 T-11 之前串行拥有 workspace lock；其他模块通过消费者 Port 和稳定 Permission Code 消费。
- **批准偏差：** `T07-D01` 仅允许两个 IAM package manifest 增加 `./administration` 公共 export、聚合 test/typecheck，并为 Web IAM 声明 `@go-admin/ui` 直接依赖；lockfile 只记录对应 importer，禁止新增外部版本、修改产品 kernel 或扩大其他依赖。
- **保留或不动：** Organization 由 T-09 实现。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | IAM API/Web suite | `task test -- iam-authorization` | CRUD、分配和获权操作即时生效 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-07.md</Path>` |
| 失败路径 | permission matrix | 直接调用缺权/越界/保护对象操作 | 后端拒绝且状态不变 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-07.md</Path>` |
| 回归 | cache/browser suite | 禁用缓存并刷新 manifest/菜单 | 结果正确且 UI/后端语义一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-07.md</Path>` |

- **Workspace checks：** 按 Goal Plan 在 current-workspace 或 source-worktree 运行 Go/TS/unit/type/lint/build。
- **E2E disposition：** required：菜单可见性、管理 CRUD 和直接越权 API 必须联合验证。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不运行 required E2E。
- **Integration evidence：** 记录 implementation/source commit、parent before、candidate/result SHA、权限 E2E 和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 在 T-06 身份基础上原子建立新授权 schema/API，业务模块随后注册 Permission Code。
- **兼容窗口：** 无 Casbin/JWT/tenant 双写或导入。
- **监控信号：** authorization deny、保护对象冲突、权限刷新与 scope query。
- **回滚或前向恢复：** 业务模块接入前可回滚；接入后用前向策略修复并强制 Session 重新授权。
- **不可逆操作与批准点：** 系统保护对象删除永不开放；schema 破坏性变更需批准。
- **收缩条件：** T-21 证明 Casbin、JWT claims 和 tenant scope 零引用。

## 10. 验收标准

- [ ] `AC-011`：Permission Code/数据范围矩阵在 UI 和后端最终判定均成立。
- [ ] `AC-013`：IAM 管理闭环、引用冲突和即时权限变更可判定。
- [ ] `AC-035/AC-036`：管理交互、菜单和路由状态符合共享合同。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-07.md</Path>`。
- [ ] 修改未超出 `writable_paths`，形成非空 commit 并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 状态一致且无未批准偏差。
