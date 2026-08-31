---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-02
title: 建立首管理员 Bootstrap 与离线恢复用例
status: in_progress
planning_depth: deep
planning_depth_reason: 涉及认证、双方言 schema、并发唯一性、凭据处理和灾难恢复授权边界
ready: true
risk: critical
blocked_by: []
contract_ids: [AC-001, AC-002, AC-003, AC-006, AC-007]
owner: codex-root
expected_changes: ["<Path>go-admin-plus/internal/modules/iam/bootstrap/**</Path>", "<Path>go-admin-plus/internal/modules/iam/recovery/**</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0030-bootstrap-recovery/**</Path>", "<Path>go-admin-plus/test/database/**</Path>", "<Path>database/bootstrap/**</Path>", "<Path>database/README.md</Path>"]
writable_paths: ["<Path>go-admin-plus/internal/modules/iam/bootstrap/**</Path>", "<Path>go-admin-plus/internal/modules/iam/recovery/**</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0030-bootstrap-recovery/**</Path>", "<Path>go-admin-plus/test/database/**</Path>", "<Path>database/bootstrap/**</Path>", "<Path>database/README.md</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/account/**</Path>", "<Path>go-admin-plus/internal/modules/iam/session/**</Path>", "<Path>go-admin-plus/internal/modules/iam/administration/**</Path>", "<Path>go-admin-plus/internal/platform/database/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-02: 建立首管理员 Bootstrap 与离线恢复用例

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/02-bootstrap-and-admin-recovery.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-02.md</Path>`

## 1. 战略与来源

- **目标：** 为 Server/Desktop 提供唯一共享 IAM Bootstrap 用例，并为已有系统提供严格离线 recover-admin 用例。
- **可观察产出：** SQLite/PostgreSQL 空库只能建立一次系统管理员；恢复只能修改既有非删除账号且撤销旧 Session。
- **来源：** `US-001~003`、`AC-001~003`、`AC-006~007`、`ADR-001`、`ADR-007`、`ADR-013`。
- **当前事实：** 迁移后 `accounts=0`；`<Path>database/bootstrap/**</Path>` 含固定凭据 SQL 草案且属于当前未提交工作，必须按最新 Spec 安全替换而非盲目清理。
- **Planning Depth 原因：** 身份创建和恢复具有最高权限、并发和凭据泄漏风险。

## 2. 决策状态

### 已锁定决策

- Bootstrap 仅账号为空时成功；Server CLI 与 Desktop 原生流程调用同一用例，不提供未认证 HTTP API。
- recover-admin 要求离线运维锁，只重置/启用既有非删除账号、授予 system-admin 并撤销旧 Session。
- 密码仅来自 TTY 或权限受限 secret file，不接受 argv，不写日志/审计 payload。

### 已采用的低影响假设

- 运维锁使用当前数据库/文件锁能力组合；具体锁名为内部细节，但必须跨进程排他。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 双方言用例、迁移、审计事实、运维锁、固定 SQL 移除 | Argon2 密码规则、数据库事务、Session 撤销 port | CLI 解析/产品 wiring（T-09）、Desktop UI（T-10）、远程恢复 API |

## 4. 要构建什么

受信任宿主把规范化账号标识、强密码 secret 和可选恢复原因交给 IAM。Bootstrap 在同一事务检查系统为空、创建账号、授予受保护角色并写审计；并发只有一个成功。Recovery 获得排他运维锁后验证目标未删除，再重置密码、启用账号、补齐系统管理员授权、撤销旧 Session 并写脱敏审计。任一步失败都不留下部分状态。

## 5. 实现契约

- **入口或接缝：** `Bootstrap`、`RecoverAdmin` application use case；CLI/Desktop adapter 后续注入。
- **输入与输出：** 账号标识、secret reader、恢复原因；输出非敏感账号引用和稳定结果码。
- **公共接口变化：** 本 Ticket 不暴露网络接口；公共 CLI 在 T-09。
- **不变量：** Bootstrap 一生一次；系统管理员依据角色授权而非用户名；deleted 账号不可恢复。
- **状态或数据流：** empty -> bootstrapping -> initialized；恢复只在 initialized 内修改既有账号。
- **错误与失败行为：** 非空库、弱密码、不安全文件、并发冲突、锁失败、目标无效均原子失败。
- **兼容要求：** 删除固定凭据 SQL，不保留默认 admin 或旧初始化方式。
- **安全与隐私要求：** secret 零回显；审计仅记录操作者、目标稳定引用、结果和原因分类。

## 6. 执行路线

1. 为 SQLite/PostgreSQL 空库、重复、并发和恢复失败建立红灯测试。
2. 增加双方言 migration 与模块级 Bootstrap/Recovery ports。
3. 实现事务、系统管理员不变量、运维锁和 Session 撤销接缝。
4. 删除固定凭据 SQL，重写 database 说明为 CLI/原生初始化合同。
5. 运行双方言、竞态、secret 扫描和回归验证。

## 7. 路径访问契约

- **预计修改点/可写范围：** frontmatter 所列 Bootstrap、Recovery、0030 migration、数据库测试与 bootstrap 文档路径。
- **只读上下文：** 现有 account/session/administration/database 实现。
- **共享路径：** 无；产品 migration 注册和 CLI wiring 由 T-09。
- **保留或不动：** `<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs</Path>` 现有用户修改由 T-10 独占。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | use case 双方言 | 定向 Go tests + required PG DSN | 初始化/恢复状态完整 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-02.md</Path>` |
| 失败路径 | 并发/secret/目标 | 并发 Bootstrap、运行中恢复、deleted 目标、弱 secret | 仅允许合法事务，无泄漏或部分写入 | 同上 |
| 回归 | IAM/数据库 | `cd go-admin-plus && go test ./internal/modules/iam/... ./test/database/... -count=1` | IAM 回归通过 | 同上 |

- **Workspace checks：** current-workspace 或 source-worktree 运行双方言模块测试、race、vet；真实 required PG 汇总由 T-18。
- **E2E disposition：** not-required：模块用例和数据库事务是本 Ticket 的稳定入口；CLI、Desktop 与 clean-room 外部链路分别由 T-09、T-10、T-20/T-21。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；本 Ticket 不声明 source E2E。
- **Integration evidence：** source/implementation commit、direct-parent/candidate、result SHA 与 Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序：** 0030 双方言表/约束先应用，再注册用例；T-09 才暴露宿主入口。
- **兼容窗口：** 无；固定 SQL 与新用例不得并存于候选。
- **监控信号：** Bootstrap/Recovery 脱敏审计结果、冲突与锁失败分类。
- **回滚或前向恢复：** migration forward-only；失败事务回滚；已初始化账号不得自动删除，恢复依靠备份。
- **不可逆操作与批准点：** 首管理员提交和管理员恢复需 TTY 明确确认；Deep Ticket 实施前需要 Goal Plan 用户授权。
- **收缩条件：** 仓库无固定管理员密码、默认账号和旧 bootstrap 调用点。

## 10. 验收标准

- [ ] `AC-001~003、AC-006~007` 全部按双方言和失败矩阵成立。
- [ ] 固定凭据 SQL 被安全移除，现有未提交数据库资产按 Spec 处理并有 Evidence。
- [ ] 验证记录到 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-02.md</Path>`。
- [ ] 路径所有权、实现 commit、direct-parent/candidate、E2E disposition 和父分支 result 完整。
