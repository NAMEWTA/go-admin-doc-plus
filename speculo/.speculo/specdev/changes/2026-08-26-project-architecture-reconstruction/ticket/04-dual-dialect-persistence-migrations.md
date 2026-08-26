---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-04
title: 双方言持久化与模块迁移基座
status: done
planning_depth: deep
planning_depth_reason: 数据库双方言、schema 迁移、单实例约束和恢复直接影响数据完整性
ready: true
risk: high
blocked_by: [T-03]
contract_ids: [AC-003, AC-004, AC-024, AC-034]
owner: codex-t04-database
expected_changes: ["<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus/internal/platform/migrations/**</Path>", "<Path>database/**</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus/internal/platform/migrations/**</Path>", "<Path>go-admin-plus/internal/platform/cache/**</Path>", "<Path>database/**</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>go-admin-plus/internal/app/kernel/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>", "<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>"]
shared_paths: ["<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus/internal/platform/migrations/**</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>"]
shared_path_owners: ["<Path>go-admin-plus/internal/platform/database/**</Path> => T-04", "<Path>go-admin-plus/internal/platform/migrations/**</Path> => T-04", "<Path>go-admin-plus/go.mod</Path> => T-04", "<Path>go-admin-plus/go.sum</Path> => T-04"]
---

# Ticket T-04: 双方言持久化与模块迁移基座

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/04-dual-dialect-persistence-migrations.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-04.md</Path>`

## 1. 战略与来源

- **目标：** 为模块提供 Bun + `database/sql` 私有 Repository 与 Goose 双方言只前进迁移基座。
- **可观察产出：** PostgreSQL、Server SQLite、Desktop SQLite 可从空库和上一新架构 fixture 确定性迁移；Server SQLite 拒绝第二实例。
- **来源：** `US-002`、`US-003`、`AC-003`、`AC-004`、`AC-024`、`AC-034`、`ADR-013`、`ADR-016`。
- **当前事实：** 旧代码混合多数据库、全局模型、AutoMigrate 倾向和 tenant 字段。
- **Planning Depth 原因：** schema 与迁移错误可能导致不可恢复数据损坏。

## 2. 决策状态

### 已锁定决策

- Server 支持 PostgreSQL/SQLite，Desktop 仅 SQLite；SQLite 使用固定无 CGo driver 且单实例。
- 模块拥有 record/repository/migration，禁止跨模块表查询和 AutoMigrate。

### 已采用的低影响假设

- 双方言差异封装在明确 dialect adapter，不暴露到领域用例。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| Database port、Bun adapter、Goose 组合器、锁与测试 fixture | T-03 profile/lifecycle | 业务表、旧数据迁移、Redis 协调 |

## 4. 要构建什么

宿主根据已验证 profile 打开唯一 Database，组合模块迁移并在 ready 前执行；空库、重复运行和上一版本升级均可判定，失败绝不以部分 schema 提供服务。

## 5. 实现契约

- **入口或接缝：** database factory、dialect、module migration provider、migration runner、single-instance lock。
- **输入与输出：** 类型化数据库配置与模块 migration 集；返回 Database/transaction 边界或脱敏启动错误。
- **公共接口变化：** 新增模块迁移 Provider 和 transaction port。
- **不变量：** 一个进程一个 Database；模块不访问他表；无 AutoMigrate；缓存为空不影响正确性。
- **状态或数据流：** profile -> driver -> lock -> migration compose/run -> repository readiness。
- **错误与失败行为：** 迁移失败不 ready；SQLite 第二实例拒绝；错误不回显 DSN/SQL 数据。
- **兼容要求：** 不支持 MySQL、SQL Server、旧多数据源或旧生产数据。
- **安全与隐私要求：** DSN/凭据脱敏；测试 fixture 不含真实 secret。

## 6. 执行路线

1. 建立空库、重复迁移、失败和第二实例测试。
2. 固定数据库依赖，建立 factory、dialect 和 transaction port。
3. 实现 Goose Provider 组合与模块迁移注册合同。
4. 加入双方言 fixture、SQLite 锁和 cache-disabled 探针。
5. 在真实 PostgreSQL 与两个 SQLite profile 运行矩阵。

## 7. 路径访问契约

- **预计修改点：** 数据库/迁移平台、根数据库资产和 Go 依赖清单。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** T-03 kernel/config。
- **共享路径：** Database、migration API 和 Go 依赖由 T-04 唯一拥有；模块只实现 Provider。
- **保留或不动：** 各业务 migration 由对应模块 Ticket 拥有。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 双方言矩阵 | `task migrate:test -- all-profiles` | 空库/升级/重复运行成功且 schema 等价 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-04.md</Path>` |
| 失败路径 | migration/lock suite | 注入中途失败并启动第二 SQLite 实例 | 不 ready、原状态可恢复且第二实例拒绝 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-04.md</Path>` |
| 回归 | cache-disabled repository probe | 清空/禁用缓存重复执行探针 | 仅性能变化，结果一致且无 Redis 连接 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-04.md</Path>` |

- **Workspace checks：** 按 Goal Plan 在 current-workspace 或 source-worktree 运行 Go、race、migration fixture 和静态检查。
- **E2E disposition：** required：必须连接真实 PostgreSQL 和 SQLite 文件验证迁移与锁。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate，禁止在 source-worktree 声明 required E2E 通过。
- **Integration evidence：** 记录 implementation/source commit、parent before、适用 candidate/result SHA、数据库 E2E 和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 平台基座先落地，模块按 DAG 增加 Provider，T-17 组合全量序列。
- **兼容窗口：** 只支持空库与上一新架构 fixture，不读取旧产品数据。
- **监控信号：** migration version、ready、锁冲突、失败阶段和双方言矩阵。
- **回滚或前向恢复：** migration 只前进；失败保留原库/备份并用修复 migration 前向恢复。
- **不可逆操作与批准点：** 破坏性 schema 变更需独立批准；本 Ticket 不含此类变更。
- **收缩条件：** T-21 扫描 AutoMigrate、旧 driver、多数据源和 tenant schema 零命中。

## 10. 验收标准

- [x] `AC-003/AC-004`：三个正式 profile 可迁移，Server SQLite 第二实例被拒绝。
- [x] `AC-024`：双方言空库/升级/幂等矩阵成立且无 AutoMigrate。
- [x] `AC-034`：缓存禁用不改变持久化和业务探针正确性且无 Redis 尝试。
- [x] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-04.md</Path>`。
- [x] 修改未超出 `writable_paths`，形成非空 commit 并记录 integration result SHA。
- [x] Ticket、Map 和 Evidence 状态一致且无未批准偏差。
