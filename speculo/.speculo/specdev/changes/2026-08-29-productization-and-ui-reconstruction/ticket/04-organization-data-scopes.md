---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-04
title: 建立组织归属与五种数据范围
status: ready
planning_depth: deep
planning_depth_reason: 改变授权模型、双方言 schema、跨 IAM/Organization port 和所有受保护查询的不变量
ready: true
risk: critical
blocked_by: []
contract_ids: [AC-019, AC-020, AC-021, AC-022]
owner: unassigned
expected_changes: ["<Path>go-admin-plus/internal/modules/iam/authorization/scope*</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/data_scope*</Path>", "<Path>go-admin-plus/internal/modules/organization/projection.go</Path>", "<Path>go-admin-plus/internal/modules/organization/model.go</Path>", "<Path>go-admin-plus/internal/modules/organization/repository.go</Path>", "<Path>go-admin-plus/internal/modules/organization/service.go</Path>", "<Path>go-admin-plus/internal/modules/organization/*_test.go</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0050-data-scope/**</Path>", "<Path>go-admin-plus/test/iam/authorization/**</Path>", "<Path>go-admin-plus/test/organization/**</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/modules/iam/authorization/scope*</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/data_scope*</Path>", "<Path>go-admin-plus/internal/modules/organization/projection.go</Path>", "<Path>go-admin-plus/internal/modules/organization/model.go</Path>", "<Path>go-admin-plus/internal/modules/organization/repository.go</Path>", "<Path>go-admin-plus/internal/modules/organization/service.go</Path>", "<Path>go-admin-plus/internal/modules/organization/*_test.go</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0050-data-scope/**</Path>", "<Path>go-admin-plus/test/iam/authorization/**</Path>", "<Path>go-admin-plus/test/organization/**</Path>"]
read_only_paths: ["<Path>contracts/openapi/modules/iam-administration.yaml</Path>", "<Path>contracts/openapi/modules/organization.yaml</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/service.go</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-04: 建立组织归属与五种数据范围

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/04-organization-data-scopes.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-04.md</Path>`

## 1. 战略与来源

- **目标：** 一次形成 all/self/organization/organization-tree/custom 数据范围和明确账号组织锚点。
- **可观察产出：** 同一 Permission Code 的获准角色范围并集在双方言产生等价集合；无主部门或空 custom 不会扩大权限。
- **来源：** `US-009~010`、`AC-019~022`、`ADR-004`、`ADR-010`、`ADR-016`。
- **当前事实：** Organization 已有部门/岗位与投影，但 IAM 只有 all/self 且账号没有组织关系。
- **Planning Depth 原因：** 授权错误会跨模块越权，且涉及 schema 和组织树变化。

## 2. 决策状态

### 已锁定决策

- 账号至多一个可空主部门和多个岗位；岗位不推导主部门。
- 只合并实际授予目标 Permission Code 的启用角色范围，取并集且无显式 deny。
- organization/tree 从主部门计算，custom 由角色保存部门集合；前端过滤不构成授权。

### 已采用的低影响假设

- 组织树投影继续使用现有稳定 ID/parent 边，具体 SQL 查询策略由双方言测试约束。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 0050 migration、scope resolver、组织投影/引用规则、授权 contract tests | 现有 Permission Code、role grant、Organization port | 租户、多主部门、显式 deny、公共 DTO/UI（T-08/T-14） |

## 4. 要构建什么

管理员可在模块用例层为账号设置主部门/岗位并为角色授权数据范围。业务授权解析器基于目标 Permission Code 收集有效角色，展开主部门树或 custom 集合并返回并集。组织变更后下一请求使用新结果；无主部门、空集合、禁用或无权角色不能扩大可见数据。

## 5. 实现契约

- **入口或接缝：** IAM data-scope resolver、Organization projection port、账号组织赋值 use case。
- **输入与输出：** account ID、permission code、role grants、组织投影；输出规范化 scope predicate/department set。
- **公共接口变化：** 模块合同先建立；OpenAPI/生成 transport 由 T-08。
- **不变量：** 一个主部门；岗位部门一致；无权角色不参与；空组织范围保持空。
- **状态或数据流：** role grant + account organization -> resolver -> 后端查询/命令目标约束。
- **错误与失败行为：** 无效部门、跨部门岗位、删除中引用、树环或失效 custom 产生 validation/conflict，不降级为 all。
- **兼容要求：** 现有角色迁移到不扩大权限的确定范围；无租户兼容。
- **安全与隐私要求：** scope 只由服务端计算，错误不泄露越权资源详情。

## 6. 执行路线

1. 建立五种范围、并集、空范围和组织变化红灯测试。
2. 增加 0050 双方言 migration 与确定 backfill。
3. 实现 Organization 最小投影、账号组织用例和 scope resolver。
4. 将代表性授权查询接入 resolver contract，证明直接 API 越权被拒。
5. 运行双方言等价、组织树移动和权限回归。

## 7. 路径访问契约

- **预计修改点/可写范围：** scope/data_scope 新接缝、Organization 模型/投影、0050 migration 与测试。
- **只读上下文：** canonical OpenAPI 与现有 administration service。
- **共享路径：** 无；公共合同由 T-08，UI 由 T-14。
- **保留或不动：** 不创建 tenant 字段或跨模块私表查询。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | scope resolver | 五种范围与角色并集双方言测试 | 集合等价且可解释 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-04.md</Path>` |
| 失败路径 | 组织引用/越权 | 空主部门、空 custom、禁用角色、跨部门岗位、直接 API | 空集合或稳定拒绝 | 同上 |
| 回归 | IAM/Organization | `cd go-admin-plus && go test ./internal/modules/iam/... ./internal/modules/organization/... ./test/iam/authorization/... ./test/organization/... -count=1` | 回归通过 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行单元、双方言、race、architecture checks。
- **E2E disposition：** not-required：授权集合由双方言 contract 直接证明；管理 UI 与浏览器越权流程在 T-14/T-19。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；本 Ticket 不在 source 运行 E2E。
- **Integration evidence：** commit、direct-parent/candidate/result SHA 和 Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** 0050 schema/backfill -> resolver -> T-08 公共合同 -> T-14 UI。
- **兼容窗口：** 无旧范围双写；backfill 默认不得扩大现有权限。
- **监控信号：** scope 类型、结果集合大小分类、拒绝与失效组织引用计数，不记录敏感数据内容。
- **回滚或前向恢复：** forward-only；错误范围前向修复并可临时禁用受影响角色，不回退 schema。
- **不可逆操作与批准点：** 角色范围 backfill 前由 Lead 核对样本与总数。
- **收缩条件：** all/self 旧分支和未带 scope 的授权查询扫描为零。

## 10. 验收标准

- [ ] `AC-019~022` 在双方言、组织变化和越权测试中成立。
- [ ] 无主部门/custom 空集合不扩大权限，多角色仅合并获准角色。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-04.md</Path>`。
- [ ] 路径、commit、集成、父分支 result 和 E2E disposition 完整。

