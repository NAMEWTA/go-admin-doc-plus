---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-10
title: Settings 参数与字典垂直切片
status: done
planning_depth: deep
planning_depth_reason: 模块新增公共 API、双方言 schema，并必须隔离业务设置与运行 secret
ready: true
risk: high
blocked_by: [T-07]
contract_ids: [AC-015, AC-035]
owner: codex-t10-settings
expected_changes: ["<Path>contracts/openapi/modules/settings.yaml</Path>", "<Path>go-admin-plus/internal/modules/settings/**</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/settings/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/settings/package.json</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/settings.yaml</Path>", "<Path>go-admin-plus/internal/modules/settings/**</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/settings/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/settings/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/settings/package.json</Path>", "<Path>go-admin-plus/test/settings/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/settings/**</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/authorization/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/package.json</Path> => T-10 under T10-D02; Settings typecheck only", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path> => T-10 under T10-D02; Settings specs only", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-10 under T10-D02; existing domain-settings and web-domain-settings importers only"]
---

# Ticket T-10: Settings 参数与字典垂直切片

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/10-settings-module.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-10.md</Path>`

## 1. 战略与来源

- **目标：** 交付应用参数、界面设置、字典类型/数据和 option 查询闭环。
- **可观察产出：** 授权管理员可维护业务设置，消费者获得稳定 option；运行 secret 永不进入该模块。
- **来源：** `US-008`、`AC-015`、`AC-035`、`ADR-009`、`ADR-020`。
- **当前事实：** 旧 sys_config 与 runtime 配置边界模糊并带 tenant 字段。
- **Planning Depth 原因：** schema/API 和 secret 边界错误会扩大配置泄露表面。

## 2. 决策状态

### 已锁定决策

- Settings 只拥有业务可维护设置；profile、DSN、Session policy 和 secret 归 T-03。
- 字典键和值使用稳定唯一约束和确定排序。

### 已采用的低影响假设

- option 查询仅返回启用项，管理查询可查看全部状态。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 参数、界面设置、字典、options、页面 | IAM 授权、共享 UI | runtime config、secret、动态 reload |

## 4. 要构建什么

管理员通过标准列表/表单维护业务设置和字典，服务端验证键、类型、状态与引用后事务提交；业务消费者通过公开查询取得稳定且脱敏的值。

## 5. 实现契约

- **入口或接缝：** Settings API/use cases/repository、option query、Web pages。
- **输入与输出：** 参数/字典 CRUD；返回分页、详情、options 或稳定错误。
- **公共接口变化：** 新 Settings fragment 与 Permission Code。
- **不变量：** 唯一键、确定排序、业务设置不接受 secret/profile 字段、无 tenant。
- **状态或数据流：** authorized command -> validate -> transaction -> query/option projection -> UI refresh。
- **错误与失败行为：** 重复键、非法类型、引用冲突、secret-like 字段和越权无状态变化。
- **兼容要求：** 不导入旧 sys_config/sys_dict 或旧 API。
- **安全与隐私要求：** secret 关键词和值探测必须阻断并脱敏。

## 6. 执行路线

1. 固定唯一性、option、secret 拒绝和越权测试。
2. 实现 migration、repository、用例与合同 mapping。
3. 实现 headless domain、Vue 页面和共享交互。
4. 覆盖删除引用、重复提交和缓存禁用。
5. 运行双方言 API/UI/敏感值回归。

## 7. 路径访问契约

- **预计修改点：** Settings 独占合同、后端、前端、测试路径与两个 Settings package manifest。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** IAM、T-03 config 和共享 UI。
- **共享路径：** 根 workspace script、shell Vitest include 和 lockfile 仅按串行 deviation 接入。
- **批准偏差：** `T10-D01` 第一阶段只开放两个 Settings package manifest，用既有 workspace/catalog 依赖补齐公开 export 与 package-local checks；根 aggregate checks、Vitest include、lockfile、共享 UI 和 composition 保持只读，后续必须由 Lead 串行 amendment。
- **批准偏差：** `T10-D02` 在 T-15 implementation result 后只开放根 Settings 聚合 typecheck、shell Vitest 的两个 Settings include，以及 lockfile 现有 `packages/domains/settings` / `packages/web-domains/settings` importer。只允许反映两个已批准 package manifest 的既有 workspace/catalog 依赖；其他 script、include、importer、catalog、外部版本、共享 UI 和 composition 零漂移。
- **保留或不动：** runtime 配置和产品注册点。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | API/UI suite | `task test -- settings` | 参数、字典和 options 一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-10.md</Path>` |
| 失败路径 | validation/security | 提交重复键、secret 字段、引用删除和越权 | 稳定拒绝且无泄露/状态变化 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-10.md</Path>` |
| 回归 | dialect/cache E2E | 双方言并禁用缓存完成 CRUD | 行为等价且 UI 刷新正确 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-10.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 非 E2E 检查。
- **E2E disposition：** deferred：业务设置/字典跨 API/UI、双方言及 secret 拒绝场景保留到全部 Ticket 实现集成后的统一系统 E2E。
- **E2E owner/environment：** Lead / 最终系统候选；逐 Ticket source-worktree 与 parent-candidate 均不运行或声明 E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、完整非 E2E Gate、统一系统 E2E 引用与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 模块原子落地，T-17 组合。
- **兼容窗口：** 无旧设置/字典导入或双写。
- **监控信号：** validation/conflict/authorization 与 secret 拒绝计数。
- **回滚或前向恢复：** 产品接入前回滚；接入后前向 migration 修复。
- **不可逆操作与批准点：** 删除需引用检查和用户确认。
- **收缩条件：** T-21 证明旧配置表、tenant 与 runtime 混用零命中。

## 10. 验收标准

- [x] `AC-015`：业务设置、字典和 options 闭环成立且拒绝运行 secret。
- [x] `AC-035`：管理交互合同完整。
- [x] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-10.md</Path>`。
- [x] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [x] 非 E2E 实现 Gate 已执行、E2E 已登记到最终统一矩阵且 shared path 无越权写入。
- [x] Ticket、Map 和 Evidence 一致且无未批准偏差。

## 11. 当前实现结果

`main` 已包含 implementation result `030eb0b57fd47a4a5381f80a24cd3dafaeb35310`（tree `8cdd579f60f3018be864f55fb54cfa8b9fb8ab3f`）。全部非 E2E candidate Gate 已通过，T-10 进入 `implemented-pending-final-e2e`；上述验收框继续保持未勾选，直到唯一最终系统候选完成双方言 Settings API/UI、安全与产品组合验证。
