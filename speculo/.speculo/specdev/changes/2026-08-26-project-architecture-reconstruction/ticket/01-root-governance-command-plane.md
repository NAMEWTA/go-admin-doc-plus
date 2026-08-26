---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-01
title: 根级产品治理与统一命令面
status: ready
planning_depth: deep
planning_depth_reason: 根级共享路径、治理资产迁移和后续全部 Ticket 的命令合同具有高事故半径
ready: true
risk: high
blocked_by: []
contract_ids: [AC-001, AC-002]
owner: unassigned
expected_changes: ["<Path>Taskfile.yml</Path>", "<Path>.gitignore</Path>", "<Path>.github/ISSUE_TEMPLATE/**</Path>", "<Path>.github/PULL_REQUEST_TEMPLATE.md</Path>", "<Path>.husky/**</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>scripts/go-admin-plus-ui/**</Path>"]
writable_paths: ["<Path>Taskfile.yml</Path>", "<Path>.gitignore</Path>", "<Path>.gitattributes</Path>", "<Path>.editorconfig</Path>", "<Path>.github/ISSUE_TEMPLATE/**</Path>", "<Path>.github/PULL_REQUEST_TEMPLATE.md</Path>", "<Path>.husky/**</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>scripts/go-admin-plus-ui/**</Path>", "<Path>deploy/README.md</Path>", "<Path>release/README.md</Path>", "<Path>database/README.md</Path>"]
read_only_paths: ["<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>", "<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ADR.md</Path>"]
shared_paths: ["<Path>Taskfile.yml</Path>", "<Path>.husky/**</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>scripts/go-admin-plus-ui/**</Path>"]
shared_path_owners: ["<Path>Taskfile.yml</Path> => T-01", "<Path>.husky/**</Path> => T-01", "<Path>scripts/go-admin-plus/**</Path> => T-01", "<Path>scripts/go-admin-plus-ui/**</Path> => T-01"]
---

# Ticket T-01: 根级产品治理与统一命令面

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/01-root-governance-command-plane.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-01.md</Path>`

## 1. 战略与来源

- **目标：** 建立唯一仓库治理边界和稳定的根命令接口，解除后续切片对散落脚本与子项目治理文件的依赖。
- **可观察产出：** 贡献者可从根调用 `dev/build/test/lint/generate/migrate/package/release`，Hook 使用同一入口。
- **来源：** `US-001`、`AC-001`、`AC-002`、`ADR-005`、`ADR-014`、`DEC-002`。
- **当前事实：** 根与两个子项目仍并存 Git、Hook、脚本和发布治理；已有用户改动必须吸收到目标结构而不是回滚。
- **Planning Depth 原因：** 根共享文件由全部后续 Ticket 消费，错误设计会造成全仓并行冲突。

## 2. 决策状态

### 已锁定决策

- 根 Taskfile 是产品命令唯一公开入口；脚本按后端和前端归档，CI 不拥有私有实现逻辑。
- 本 Ticket 只扩展目标根治理，不提前删除旧产品结构；删除由 T-21 原子完成。

### 已采用的低影响假设

- Task 名称保持稳定，具体子命令可委派到各模块自有工具。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 根治理、Task 合同、Hook 和脚本分类 | 当前可用脚本中的有效行为 | 业务代码、平台发行实现、旧目录最终删除 |

## 4. 要构建什么

贡献者从仓库根执行统一任务时获得明确目标、工具缺失提示和非零失败码；Hook 与后续 CI 调用相同任务，子项目不再定义第二套产品命令权威。

## 5. 实现契约

- **入口或接缝：** 根 Task CLI、Git Hook、治理静态扫描。
- **输入与输出：** 任务名和可选 profile 输入；输出稳定退出码及脱敏诊断。
- **公共接口变化：** 新增根产品任务合同。
- **不变量：** Task 只委派受管脚本；Hook 不复制检查逻辑；不存在嵌套 Git 产品边界。
- **状态或数据流：** CLI/Hook -> Taskfile -> 受管脚本 -> 对应工具链。
- **错误与失败行为：** 未知任务、缺少工具或子任务失败必须向上传播失败，不能静默跳过。
- **兼容要求：** 不兼容旧子项目命令；旧资产仅保留到 T-21 收缩 Gate。
- **安全与隐私要求：** 命令不得打印 secret 或读取未声明凭据文件。

## 6. 执行路线

1. 固化根目录生命周期分类和任务合同测试。
2. 建立 Taskfile、后端/前端脚本入口和 Hook 委派。
3. 迁移可复用治理内容并标记仍待 T-21 删除的旧资产。
4. 运行治理扫描、任务解析和失败传播回归。

## 7. 路径访问契约

- **预计修改点：** 与 frontmatter `expected_changes` 一致。
- **可写范围：** 仅 frontmatter `writable_paths`；不得修改产品源码。
- **只读上下文：** 当前 Spec 与 ADR。
- **共享路径：** 根 Task、Hook 与两类脚本由 T-01 唯一拥有；消费者只读。
- **保留或不动：** 用户当前产品改动和 T-21 的收缩路径。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 根任务合同 | `task governance:check task:contract` | 所有公开任务可解析且 Hook 委派成立 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-01.md</Path>` |
| 失败路径 | 失败传播测试 | 调用未知任务并模拟子任务失败 | 返回非零且不伪造成功 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-01.md</Path>` |
| 回归 | 根静态扫描 | 检查治理文件归属和脚本引用 | 新增目标结构无重复 owner | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-01.md</Path>` |

- **Workspace checks：** 按 Goal Plan 在 current-workspace 或 source-worktree 运行非 E2E 检查。
- **E2E disposition：** not-required：该 Ticket 只建立静态命令与治理合同。
- **E2E owner/environment：** Lead / current-workspace，确认无需跨进程 E2E。
- **Integration evidence：** 记录 implementation/source commit、parent before、direct-parent 或 candidate/result SHA 及父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 先扩展根入口，后续 Ticket 迁移消费者，T-21 再删除旧入口。
- **兼容窗口：** 仅施工期内部窗口，不形成产品兼容承诺。
- **监控信号：** 根任务合同、Hook 委派和重复治理扫描。
- **回滚或前向恢复：** 在消费者迁移前可回滚；迁移后通过修复根委派前向恢复。
- **不可逆操作与批准点：** 本 Ticket 不删除旧资产。
- **收缩条件：** T-21 证明所有消费者已使用根入口且旧治理扫描归零。

## 10. 验收标准

- [ ] `AC-001`：目标根治理资产具有唯一归属，子项目重复治理被明确纳入 T-21 收缩清单。
- [ ] `AC-002`：九类根产品任务可解析，Hook 复用相同入口并正确传播失败。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-01.md</Path>`。
- [ ] 修改未超出 `writable_paths`，共享路径仅由 T-01 修改。
- [ ] 形成非空 implementation/source commit，并记录 direct-parent 或 candidate/result SHA。
- [ ] 未发生未批准偏差，Ticket、Map 和 Evidence 状态一致。
