---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-17
title: 产品组合与三 Profile 全链路汇合
status: in_progress
planning_depth: deep
planning_depth_reason: 八模块、公共合同、三数据库 Profile 和双 App 在共享产品注册点汇合
ready: true
risk: critical
blocked_by: [T-09, T-10, T-11, T-12, T-13, T-15, T-16]
contract_ids: [AC-003, AC-004, AC-011, AC-021, AC-022, AC-024, AC-025, AC-028, AC-034, AC-035, AC-036]
owner: codex-root
expected_changes: ["<Path>contracts/openapi/product.yaml</Path>", "<Path>go-admin-plus/internal/app/product/**</Path>", "<Path>go-admin-plus/cmd/go-admin-plus/**</Path>", "<Path>go-admin-plus/internal/modules/generator/migrations/0010-config/sqlite/6400000000000_generator_configs.sql</Path>", "<Path>go-admin-plus/internal/modules/generator/migrations/0010-config/postgres/6400000000000_generator_configs.sql</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus/test/product/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/product/**</Path>"]
writable_paths: ["<Path>go-admin-plus-ui/packages/platform/src/index.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/web-runtime.spec.ts</Path>", "<Path>contracts/openapi/product.yaml</Path>", "<Path>go-admin-plus/internal/app/product/**</Path>", "<Path>go-admin-plus/cmd/go-admin-plus/**</Path>", "<Path>go-admin-plus/cmd/desktop-sidecar/runtime.go</Path>", "<Path>go-admin-plus/cmd/desktop-sidecar/runtime_test.go</Path>", "<Path>go-admin-plus/internal/contracts/product-generated/**</Path>", "<Path>go-admin-plus/internal/modules/generator/migrations/0010-config/sqlite/6400000000000_generator_configs.sql</Path>", "<Path>go-admin-plus/internal/modules/generator/migrations/0010-config/postgres/6400000000000_generator_configs.sql</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/index.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/index.spec.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/desktop/**</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-web/package.json</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/package.json</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/proxy.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/product_contract.rs</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/workspace-boundary.test.mjs</Path>", "<Path>go-admin-plus-ui/packages/api-client/src/product-generated/**</Path>", "<Path>go-admin-plus/test/product/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/product/**</Path>"]
read_only_paths: ["<Path>contracts/openapi/modules/**</Path>", "<Path>go-admin-plus/internal/modules/**</Path>", "<Path>go-admin-plus-ui/packages/domains/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/**</Path>", "<Path>go-admin-plus-ui/apps/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/packages/platform/src/index.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/web-runtime.spec.ts</Path>", "<Path>contracts/openapi/product.yaml</Path>", "<Path>go-admin-plus/internal/app/product/**</Path>", "<Path>go-admin-plus/internal/modules/generator/migrations/0010-config/sqlite/6400000000000_generator_configs.sql</Path>", "<Path>go-admin-plus/internal/modules/generator/migrations/0010-config/postgres/6400000000000_generator_configs.sql</Path>", "<Path>go-admin-plus/cmd/desktop-sidecar/runtime.go</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/product/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/index.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/desktop/**</Path>", "<Path>go-admin-plus-ui/apps/admin-web/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src/App.vue</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/proxy.rs</Path>", "<Path>go-admin-plus-ui/package.json</Path>", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path>", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path>", "<Path>go-admin-plus-ui/tests/shell/workspace-boundary.test.mjs</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/packages/platform/src/index.ts</Path> => T-17 under T17-D04; PermissionCode type only", "<Path>go-admin-plus-ui/tests/shell/web-runtime.spec.ts</Path> => T-17 under T17-D04; canonical Permission fixtures only", "<Path>contracts/openapi/product.yaml</Path> => T-17", "<Path>go-admin-plus/internal/app/product/**</Path> => T-17", "<Path>go-admin-plus/internal/modules/generator/migrations/0010-config/sqlite/6400000000000_generator_configs.sql</Path> => T-17 under T17-D02; remove Down section only", "<Path>go-admin-plus/internal/modules/generator/migrations/0010-config/postgres/6400000000000_generator_configs.sql</Path> => T-17 under T17-D02; remove Down section only", "<Path>go-admin-plus/cmd/desktop-sidecar/runtime.go</Path> => T-17 under T17-D03; consume product runtime", "<Path>go-admin-plus-ui/packages/app-shell/src/product/**</Path> => T-17", "<Path>go-admin-plus-ui/packages/app-shell/package.json</Path> => T-17 under T17-D01/D03", "<Path>go-admin-plus-ui/packages/adapters/browser/src/index.ts</Path> => T-17 under T17-D03", "<Path>go-admin-plus-ui/packages/adapters/desktop/**</Path> => T-17 under T17-D03", "<Path>go-admin-plus-ui/apps/admin-web/src/App.vue</Path> => T-17 under T17-D03", "<Path>go-admin-plus-ui/apps/admin-desktop/src/App.vue</Path> => T-17 under T17-D03", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path> => T-17 under T17-D03", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/proxy.rs</Path> => T-17 under T17-D03", "<Path>go-admin-plus-ui/package.json</Path> => T-17 under T17-D03 exact product checks only", "<Path>go-admin-plus-ui/pnpm-lock.yaml</Path> => T-17 under T17-D03 exact touched importers only", "<Path>go-admin-plus-ui/tests/shell/vitest.config.ts</Path> => T-17 under T17-D03 exact product specs only", "<Path>go-admin-plus-ui/tests/shell/workspace-boundary.test.mjs</Path> => T-17 under T17-D03 exact product dependency/export assertions only"]
---

# Ticket T-17: 产品组合与三 Profile 全链路汇合

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/17-product-composition-matrix.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-17.md</Path>`

## 1. 战略与来源

- **目标：** 在唯一组合点装配八模块、模块合同、迁移、Permission manifest、Server 和双 App。
- **可观察产出：** PostgreSQL/Server SQLite/Desktop SQLite 运行适用完整能力，Web/Desktop 菜单、页面、权限和业务语义一致。
- **来源：** `US-002`、`US-003`、`US-004`、`US-019`、`AC-003`、`AC-004`、`AC-022`、`ADR-008`、`ADR-021`。
- **当前事实：** 各 Ticket 为避免并行冲突未修改中央注册点，必须在专属汇合点生成和验证完整产品。
- **Planning Depth 原因：** 这是所有公共合同、数据库和宿主的关键集成 Gate。

## 2. 决策状态

### 已锁定决策

- product composition 是唯一模块注册点；模块不自行修改 App/manifest/root bundle。
- 双 App 消费同一业务 manifest，差异只由声明的 host capability 决定。
- Server PostgreSQL 可多 API 副本但唯一 worker；两个 SQLite profile 单实例。

### 已采用的低影响假设

- 聚合生成物按稳定模块顺序产生并由 clean-tree 检查保护。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 产品注册、聚合合同/manifest、Server executable、全矩阵 E2E | T-01 至 T-16 已完成能力 | 平台安装器、旧结构删除、模块内部改造 |

## 4. 要构建什么

系统从唯一产品组合装配所有模块、迁移和路由；同一身份分别进入 Web/Desktop 时获得相同获权业务 manifest，并在三个正式 profile 完成核心管理与 Demo tracer。

## 5. 实现契约

- **入口或接缝：** product registry、module providers、OpenAPI product bundle、App manifest、Server command。
- **输入与输出：** profile + providers；返回完整 HTTP product、worker 和双 App capability graph。
- **公共接口变化：** 发布完整产品 OpenAPI 和稳定 capability/permission manifest。
- **不变量：** 单一组合点；无循环/locator/deep import；双方言/双 App 语义等价；缓存不影响正确性。
- **状态或数据流：** config/database -> migrations -> module providers -> routes/workers -> manifest -> Web/Desktop。
- **错误与失败行为：** 任一 provider/migration/contract 失败不 ready；缺权/内部错误使用稳定类别。
- **兼容要求：** 不装配旧模块/API/config/tenant/Redis/Wails。
- **安全与隐私要求：** 全权限矩阵、secret scan、错误脱敏和 Desktop Session 边界回归。

## 6. 执行路线

1. 建立产品 registry、manifest 和全 profile tracer 的失败基线。
2. 聚合模块 OpenAPI、迁移、providers、routes、workers 和 Permission Code。
3. 生成产品 Go/TS transport 并装配 Server/Web/Desktop manifest。
4. 运行模块边界、负向错误、缓存禁用和双 App 一致性测试。
5. 形成三 profile/双 App 的完整可构建发行输入，并把全链路场景登记到全部 Ticket 实现集成后的统一系统 E2E。

## 7. 路径访问契约

- **预计修改点：** 专属 product 聚合路径和测试。
- **可写范围：** 仅 frontmatter `writable_paths`；模块目录全部只读。
- **只读上下文：** 所有模块合同/实现与双 App。
- **共享路径：** product OpenAPI、后端 composition、前端 manifest 由 T-17 唯一拥有。
- **批准偏差：** `T17-D01` 开放 app-shell product export/checks；`T17-D02` 只删除 Generator 双方言迁移 `Down` 段；stage-1 checkpoint `92a0ff8` 后，`T17-D03` 精确开放产品 UI、双 App、Browser parser、Desktop adapter/Rust/sidecar 与必需的 manifest/Vitest/boundary/lock importer；`T17-D04` 只把 platform Permission 类型和对应旧 Web fixture 切换为点分格式，不保留冒号兼容。各 domain/web-domain 业务实现、模块生成合同、无关根脚本/importer 继续只读。
- **保留或不动：** 发行平台归 T-18/T-19/T-20，旧路径归 T-21。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | product profile matrix | `task test -- product-matrix` | 三 profile 核心业务 ready 且双 App manifest 一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-17.md</Path>` |
| 失败路径 | negative/authorization matrix | 缺 provider、migration、权限、依赖和内部错误 | 不 ready或稳定拒绝，无泄露/状态变化 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-17.md</Path>` |
| 回归 | architecture/contract/cache | 生成检查、边界扫描、禁用缓存、双 App E2E | 无 drift/越界，语义等价 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-17.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 运行全量非 E2E 门禁。
- **E2E disposition：** deferred：三 profile、Web/Desktop、权限和可靠运行时场景保留到全部 Ticket 实现集成后的统一系统 E2E。
- **E2E owner/environment：** Lead / 最终系统候选；逐 Ticket source-worktree 与 parent-candidate 均不运行或声明 E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、完整 composition 非 E2E Gate、统一系统 E2E 引用和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 所有模块完成后聚合；全矩阵通过后才允许平台打包。
- **兼容窗口：** 不发布混合产品，不提供旧新 API/schema 双轨。
- **监控信号：** profile readiness、module provider、manifest drift、contract drift 和 E2E 结果。
- **回滚或前向恢复：** 发行前可回滚组合 commit；失败模块保持未发布并前向修复。
- **不可逆操作与批准点：** 无旧结构删除；进入 T-21 前须三个发行候选验证完成。
- **收缩条件：** 全部模块、三 profile 和双 App Gate 通过，旧路径消费者可由 T-21 删除。

## 10. 验收标准

- [ ] `AC-003/004/024/034`：三 profile 迁移、业务、缓存禁用和双方言矩阵通过。
- [ ] `AC-011/021/022/035/036`：权限、Demo、双 App manifest、交互和导航一致。
- [ ] `AC-025/028`：负向错误脱敏且架构边界无违规。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-17.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
