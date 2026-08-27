---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-15
title: Generator 新架构代码生成闭环
status: in_progress
planning_depth: deep
planning_depth_reason: 生成器读取数据库元数据并产出公共合同、后端模块和前端包，影响共享架构边界
ready: true
risk: high
blocked_by: [T-14]
contract_ids: [AC-019, AC-023, AC-028, AC-035]
owner: codex-root
expected_changes: ["<Path>contracts/openapi/modules/generator.yaml</Path>", "<Path>go-admin-plus/internal/modules/generator/**</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/generator/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/generator/package.json</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/generator.yaml</Path>", "<Path>go-admin-plus/internal/modules/generator/**</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/src/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/generator/src/**</Path>", "<Path>go-admin-plus-ui/packages/domains/generator/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/generator/package.json</Path>", "<Path>go-admin-plus/test/generator/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/generator/**</Path>"]
read_only_paths: ["<Path>contracts/openapi/modules/demo.yaml</Path>", "<Path>go-admin-plus/internal/modules/demo/**</Path>", "<Path>go-admin-plus-ui/packages/domains/demo/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/demo/**</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-15: Generator 新架构代码生成闭环

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/15-generator-module.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-15.md</Path>`

## 1. 战略与来源

- **目标：** 交付授权元数据导入、列配置、预览和标准单表 CRUD 代码生成。
- **可观察产出：** 生成结果符合 OpenAPI、Go 模块和 pnpm 包边界，并可直接通过生成/编译/测试门禁。
- **来源：** `US-011`、`AC-019`、`AC-023`、`AC-028`、`AC-035`、`ADR-006`、`ADR-011`。
- **当前事实：** 旧生成模板绑定旧目录、DTO、前端页面和数据库模式。
- **Planning Depth 原因：** 生成器能大规模复制错误架构，且元数据读取具有安全风险。

## 2. 决策状态

### 已锁定决策

- Demo 是生成目标参考；输出包含模块 fragment、私有 migration/repository、显式 mapping 和前端双包。
- 仅读取当前 profile 的授权 schema，不读取系统 schema 或任意连接。

### 已采用的低影响假设

- 预览与落盘共享同一规范化模型，golden 输出确定。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 元数据、配置、预览、模板、落盘和编译验证 | Demo 目标模式、合同工具 | 多表聚合、运行任意模板代码、自动产品注册 |

## 4. 要构建什么

授权开发者选择允许的表，编辑列和生成配置，预览稳定 diff 后生成到隔离输出；生成物经合同、格式、编译、测试和架构检查全部通过后才报告成功。

## 5. 实现契约

- **入口或接缝：** metadata Port、Generator use cases/templates、preview/write API、Web wizard。
- **输入与输出：** 授权表/列配置；返回规范化预览和受控输出清单。
- **公共接口变化：** 新 Generator fragment、template contract 和 Permission Code。
- **不变量：** 预览=落盘模型；输出确定；禁止路径逃逸/覆盖未授权文件；生成物无 deep import/跨模块 record。
- **状态或数据流：** authorized metadata -> normalized model -> preview -> confirmed write -> gates。
- **错误与失败行为：** 未授权表、非法标识符、冲突路径或 gate 失败不产生部分成功输出。
- **兼容要求：** 不生成旧目录/API/ORM/tenant 模式。
- **安全与隐私要求：** metadata allowlist、输出根 canonicalize、模板无任意代码执行。

## 6. 执行路线

1. 以 Demo 建立双方言 metadata/golden/compile 失败测试。
2. 实现 metadata Port、配置模型和稳定 preview。
3. 重写 OpenAPI/Go/pnpm 模板与安全落盘事务。
4. 实现 Web wizard 和生成后门禁。
5. 在本 Ticket candidate 运行 golden、clean-tree、真实生成/compile 和路径逃逸非浏览器 Gate，并把 UI 场景登记到最终统一系统 E2E。

## 7. 路径访问契约

- **预计修改点：** Generator 独占路径、模板和经 `T15-D01` 精确开放的两个 Generator package manifest。
- **可写范围：** 仅 frontmatter `writable_paths`；测试生成输出使用临时目录。
- **只读上下文：** Demo 四层参考实现。
- **共享路径：** 无；公共生成器只读调用 T-02。
- **批准偏差：** `T15-D01` 只开放两个 Generator package manifest，用既有 workspace/catalog 依赖补齐 canonical API client、`@go-admin/ui`、Vue、公开 exports 与 package-local test/typecheck。第一阶段根 package/Vitest/lock 继续只读并等待 T-09 result；禁止新增外部版本、修改其他 importer、共享 UI、公共合同工具或产品 composition。
- **保留或不动：** 产品聚合和真实模块路径。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | golden/compile E2E | `task test -- generator` | 双方言预览稳定，生成物通过全部门禁 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-15.md</Path>` |
| 失败路径 | security/atomic output | 未授权表、非法名、逃逸和 gate 失败 | 拒绝且无部分/越界写入 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-15.md</Path>` |
| 回归 | architecture scan | 对生成 fixture 运行合同/Go/TS/边界检查 | clean tree 且无旧模式 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-15.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 运行 golden/unit/type/build。
- **E2E disposition：** deferred：真实临时输出生成、编译和测试仍属于本 Ticket 非 E2E Gate；Web wizard 与跨产品场景保留到全部 Ticket 实现集成后的统一系统 E2E。
- **E2E owner/environment：** Lead / 最终系统候选；逐 Ticket source-worktree 与 parent-candidate 不运行浏览器 E2E。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、真实生成/编译非 E2E Gate、统一系统 E2E 引用与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** Demo 稳定后落地 Generator，T-17 组合但不自动生成产品代码。
- **兼容窗口：** 无旧模板或生成配置兼容。
- **监控信号：** preview/write/gate failure、路径拒绝和模板 drift。
- **回滚或前向恢复：** 输出先写临时目录；失败删除临时输出，已确认输出通过版本控制恢复。
- **不可逆操作与批准点：** 覆盖既有目标一律禁止；真实落盘需用户确认。
- **收缩条件：** T-21 证明旧模板、旧路径和旧生成 API 零引用。

## 10. 验收标准

- [ ] `AC-019`：导入、配置、预览、生成与编译闭环成立。
- [ ] `AC-023/AC-028`：生成物合同确定且通过架构边界检查。
- [ ] `AC-035`：向导校验、确认和状态刷新符合共享交互。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-15.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
