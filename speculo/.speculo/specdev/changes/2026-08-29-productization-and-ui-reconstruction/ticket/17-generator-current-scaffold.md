---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-17
title: 升级 Generator 为当前全垂直脚手架
status: ready
planning_depth: deep
planning_depth_reason: 模板会写入双方言 schema、公共 API、后端模块、前端 workspace 和权限注册，事故半径覆盖未来所有生成模块
ready: true
risk: high
blocked_by: [T-04, T-05, T-06, T-09, T-11, T-14, T-16]
contract_ids: [AC-034]
owner: unassigned
expected_changes: ["<Path>go-admin-plus/internal/modules/generator/templates.go</Path>", "<Path>go-admin-plus/internal/modules/generator/generator.go</Path>", "<Path>go-admin-plus/internal/modules/generator/generator_test.go</Path>", "<Path>go-admin-plus/internal/modules/generator/compile_gate.go</Path>", "<Path>go-admin-plus/internal/modules/generator/*_test.go</Path>", "<Path>scripts/quality/architecture-check.mjs</Path>", "<Path>scripts/quality/architecture-check.test.mjs</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/modules/generator/templates.go</Path>", "<Path>go-admin-plus/internal/modules/generator/generator.go</Path>", "<Path>go-admin-plus/internal/modules/generator/generator_test.go</Path>", "<Path>go-admin-plus/internal/modules/generator/compile_gate.go</Path>", "<Path>go-admin-plus/internal/modules/generator/*_test.go</Path>", "<Path>scripts/quality/architecture-check.mjs</Path>", "<Path>scripts/quality/architecture-check.test.mjs</Path>"]
read_only_paths: ["<Path>contracts/openapi/**</Path>", "<Path>go-admin-plus/internal/modules/demo/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>", "<Path>.agents/skills/new-business-module/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-17: 升级 Generator 为当前全垂直脚手架

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/17-generator-current-scaffold.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-17.md</Path>`

## 1. 战略与来源

- **目标：** 让 Generator 输出与本 change 最终架构、权限、数据范围和 UI 组件一致的单表 CRUD 垂直模块。
- **可观察产出：** 隔离工作区一次生成双方言 migration、OpenAPI、Go module、domain/web-domain、权限注册和 tests，并通过全部门禁。
- **来源：** `US-015`、`AC-034`、`PLAN/P6`。
- **当前事实：** Generator 已有 preview/write/compile gate 与大模板，但模板仍需跟随新合同、组件和架构检查更新。
- **Planning Depth 原因：** 模板是未来代码倍增器，任何坏模式都会批量扩散到全栈。

## 2. 决策状态

### 已锁定决策

- 生成输出必须同时包含 SQLite/PostgreSQL、modular OpenAPI、Go 垂直模块、headless domain、Web Domain、权限和测试。
- 不生成 tenant、Redis、JWT、common/utils 聚合或页面手写基础控件。
- preview 不写盘，write 必须经过隔离 compile gate 且输出冲突拒绝覆盖。

### 已采用的低影响假设

- Demo 模块和完成后的 T-14/T-16 页面作为当前参考形状；生成器不复制专用业务状态机。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| templates、generation model、compile gate、isolated tests、架构规则 | canonical OpenAPI generator、Demo、T-11 UI、现有 skill 只读指导 | 聚合/外部集成模块、修改 skill、自动 commit |

## 4. 要构建什么

开发者选择允许表并配置模块命名后，可以预览确定文件集合；write 在仓库边界内创建完整垂直模块并运行格式、合同、Go、pnpm 与架构检查。生成结果使用当前 list/form 组件、Permission Code 和双方言模式，不留下手工补齐的关键层。

## 5. 实现契约

- **入口或接缝：** Generator preview/write use cases、WorkspaceCompileGate。
- **输入与输出：** allowlisted table metadata、draft；输出稳定 preview manifest 或完整 files/gate result。
- **公共接口变化：** 保持现有 Generator API 语义；模板输出合同更新。
- **不变量：** 路径不能逃逸、已有文件不覆盖、双 dialect 同时生成、生成物可重复验证。
- **状态或数据流：** metadata -> normalized draft -> preview manifest -> atomic write -> isolated gates。
- **错误与失败行为：** invalid identifier/path/conflict/gate failure 稳定返回，不保留半写入目录。
- **兼容要求：** 不兼容旧模板外观；只生成当前架构。
- **安全与隐私要求：** 不读取非 allowlist schema/table，不执行数据库内容或生成 preview HTML。

## 6. 执行路线

1. 固定目标文件树、禁止模式和隔离 gate 红灯 fixture。
2. 更新生成模型和模板为当前后端/前端/权限/测试形状。
3. 强化 atomic write、冲突和失败清理。
4. 更新架构检查以识别新模板边界和禁止旧模式。
5. 在 SQLite/PostgreSQL metadata 下完成隔离 lint/typecheck/test/build。

## 7. 路径访问契约

- **预计修改点/可写范围：** Generator 核心模板/测试和架构检查。
- **只读上下文：** canonical contracts、Demo、UI、skill。
- **共享路径：** 无；不修改 OpenAPI/生成物共享路径。
- **保留或不动：** `.agents/skills`、产品业务模块和 root command plane。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | isolated generator | 双方言 preview/write fixture | 完整文件树和门禁通过 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-17.md</Path>` |
| 失败路径 | path/conflict/gate | 路径逃逸、已有文件、故意编译失败 | 原子拒绝且清理完整 | 同上 |
| 回归 | generator/architecture | 定向 Go tests + `task architecture:check` | 现有生成行为/边界通过 | 同上 |

- **Workspace checks：** current-workspace/source-worktree 运行完整隔离生成、Go/pnpm 门禁和架构检查。
- **E2E disposition：** not-required：隔离生成后的真实编译/测试是最直接系统证据；浏览器 Generator 操作由 T-19。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source 不声明浏览器 E2E。
- **Integration evidence：** generated tree manifest、commit、direct-parent/candidate/result SHA、Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** 后端/UI 模式稳定 -> templates -> isolation gate -> T-18 CI/T-19 E2E。
- **兼容窗口：** 无旧模板选择器；原子替换为当前模板。
- **监控信号：** preview/write/gate failure classes、隔离耗时与输出数量。
- **回滚或前向恢复：** 模板可回退 commit；write 失败清理，已生成用户文件不自动删除。
- **不可逆操作与批准点：** write 前仍显示 preview/目标路径并要求明确动作。
- **收缩条件：** 模板/fixtures 中旧 UI、tenant、compat 和缺失层扫描为零。

## 10. 验收标准

- [ ] `AC-034` 隔离全栈生成成立。
- [ ] 双 dialect、OpenAPI、Go、domain/Web Domain、permission、tests 均生成且门禁通过。
- [ ] Evidence 写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-17.md</Path>`。
- [ ] commit、integration/result 和 E2E disposition 完整。
