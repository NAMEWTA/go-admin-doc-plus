---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-09
title: Organization 部门与岗位垂直切片
status: in_progress
planning_depth: deep
planning_depth_reason: 模块新增公共 API、双方言 schema、组织树不变量和 IAM 消费者 Port
ready: true
risk: high
blocked_by: [T-07]
contract_ids: [AC-014, AC-035]
owner: codex-t09-organization
expected_changes: ["<Path>contracts/openapi/modules/organization.yaml</Path>", "<Path>go-admin-plus/internal/modules/organization/**</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/organization_port.go</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/organization/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/organization/package.json</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/organization.yaml</Path>", "<Path>go-admin-plus/internal/modules/organization/**</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/organization_port.go</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/organization/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/organization/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/organization/package.json</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus/test/organization/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/organization/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/authorization/**</Path>", "<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: ["<Path>go-admin-plus/internal/modules/iam/administration/organization_port.go</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus/internal/modules/iam/administration/organization_port.go</Path> => T-09 under T09-D01; consumer-defined projection Port only", "<Path>go-admin-plus-ui/package.json</Path> => T-09 under T09-D02; Organization aggregate checks only", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path> => T-09 under T09-D02; Organization specs only", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-09 under T09-D02; two Organization importers only"]
---

# Ticket T-09: Organization 部门与岗位垂直切片

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/09-organization-module.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-09.md</Path>`

## 1. 战略与来源

- **目标：** 交付部门树、岗位和组织关系的合同、持久化、授权 API 与管理页面。
- **可观察产出：** 管理员可维护合法组织结构；循环、非法父子和被引用删除被拒绝。
- **来源：** `US-007`、`AC-014`、`AC-035`、`ADR-009`、`ADR-012`。
- **当前事实：** 旧部门/岗位位于 admin 模型并夹带 tenant 与跨表访问。
- **Planning Depth 原因：** 组织树和引用数据进入 IAM 数据范围，需保护 schema 与模块边界。

## 2. 决策状态

### 已锁定决策

- Organization 独占 schema/repository；IAM 通过消费者定义 Port 读取必要组织投影。
- 不存在 tenant 根节点或 tenant 过滤。

### 已采用的低影响假设

- 部门树以稳定业务键和显式排序返回。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 部门、岗位、关系、Port、页面和迁移 | IAM 授权、共享 list/form | IAM 用户实现、跨模块 DB join |

## 4. 要构建什么

授权管理员从页面维护部门树和岗位，命令经 Organization 用例验证树与引用不变量后事务提交；IAM 只通过 Port 获取投影，不读取 Organization 私有表。

## 5. 实现契约

- **入口或接缝：** Organization API/use cases/repository、IAM consumer Port、Web domain。
- **输入与输出：** 树/岗位 CRUD 命令；返回确定排序投影或稳定 validation/conflict/authorization 错误。
- **公共接口变化：** 新 Organization OpenAPI fragment 与 Permission Code。
- **不变量：** 无环、父节点存在、保护根不可删、被引用删除失败、无跨模块表查询。
- **状态或数据流：** authorized command -> invariant check -> module transaction -> projection/event -> UI refresh。
- **错误与失败行为：** 非法父子、重复键、引用冲突或越权均无状态变化。
- **兼容要求：** 不导入旧 dept/post ID、tenant 字段或 API。
- **安全与隐私要求：** 查询和变更均执行 Permission Code 与适用数据范围。

## 6. 执行路线

1. 固定树、引用与越权失败测试。
2. 实现模块 migration、repository、用例和消费者 Port adapter。
3. 增加 OpenAPI mapping、前端 domain/web-domain 页面。
4. 覆盖重复提交、删除确认和列表刷新。
5. 运行双方言 API/UI/架构边界验证。

## 7. 路径访问契约

- **预计修改点：** Organization 独占路径、两个 Organization package manifest、经 `T09-D01` 批准的 IAM consumer Port 精确文件，以及经 `T09-D02` 批准的 Organization 根门禁与两个 lock importer。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** IAM 授权、Database、共享 UI。
- **共享路径：** `iam/administration/organization_port.go` 只定义 IAM 消费所需的最小组织投影 Port；不得导入 Organization。T-14/T-11 result 均已完成，现由 T-09 串行拥有 Organization 根 verify、Vitest include 与两个 lock importer；T-16 stage 1 不写这些路径。
- **批准偏差：** `T09-D01` 允许两个 Organization package manifest 补齐 canonical API client、`@go-admin/ui`、Vue 直接依赖、公开 export 与标准 package-local checks，并允许精确 consumer Port 文件。只可复用既有 catalog 版本；禁止修改其他 IAM 文件、其他 importer、共享 UI 或 composition。
- **批准偏差：** `T09-D02` 在 T-14 result 后只开放根 `package.json` 的 Organization 聚合 checks、shell Vitest 的 Organization specs 和 lockfile 的两个 Organization importer。必须 rebase 到最新 parent 并证明其他 scripts、includes、importers、catalog 与外部版本零漂移；禁止修改共享 list 或 IAM capability registry。
- **保留或不动：** IAM 私有 schema 和产品注册点。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | repository/API/UI | `task test -- organization` | 部门树、岗位和关系 CRUD 一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-09.md</Path>` |
| 失败路径 | invariant suite | 创建循环并删除被引用节点 | 稳定失败且数据不变 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-09.md</Path>` |
| 回归 | boundary/E2E | 检查无跨表 join 并完成浏览器 CRUD | Port 边界与交互合同成立 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-09.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 非 E2E 检查。
- **E2E disposition：** required：组织树交互、权限和引用冲突需跨 API/UI 验证。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不声明 E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、E2E 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 模块 schema/API/UI 同切片落地，T-17 再接入产品组合。
- **兼容窗口：** 无旧数据/API/tenant 兼容。
- **监控信号：** validation/conflict/authorization 和 migration 失败率。
- **回滚或前向恢复：** 产品接入前可回滚；接入后以新 migration 前向修复。
- **不可逆操作与批准点：** 数据删除保留确认与引用保护，无额外批准点。
- **收缩条件：** T-21 扫描旧 dept/post 和 tenant 引用归零。

## 10. 验收标准

- [ ] `AC-014`：合法组织 CRUD 成功，非法树、引用冲突和越权失败且不破坏数据。
- [ ] `AC-035`：搜索、表单校验、删除确认和刷新符合共享交互。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-09.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] E2E disposition 已执行且 shared path 无越权写入。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
