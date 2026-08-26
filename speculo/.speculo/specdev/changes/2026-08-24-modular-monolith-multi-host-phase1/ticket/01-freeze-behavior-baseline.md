---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-01
title: 冻结 API、UI、数据与启动行为基线
status: done
planning_depth: standard
planning_depth_reason: 跨前后端建立后续宽重构所需的公共行为接缝，但不改变公共接口或数据
ready: true
risk: medium
blocked_by: []
contract_ids: [AC-002, AC-003]
owner: root
expected_changes: ["<Path>go-admin-plus/test/characterization/**</Path>", "<Path>go-admin-plus/common/file_store/*_test.go</Path>", "<Path>go-admin-plus/common/file_store/testdata/**</Path>", "<Path>go-admin-ui-plus/tests/e2e/**</Path>", "<Path>go-admin-ui-plus/tests/fixtures/**</Path>"]
writable_paths: ["<Path>go-admin-plus/test/characterization/**</Path>", "<Path>go-admin-plus/common/file_store/*_test.go</Path>", "<Path>go-admin-plus/common/file_store/testdata/**</Path>", "<Path>go-admin-ui-plus/tests/e2e/**</Path>", "<Path>go-admin-ui-plus/tests/fixtures/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/cmd/api/**</Path>", "<Path>go-admin-ui-plus/src/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-01: 冻结 API、UI、数据与启动行为基线

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/01-freeze-behavior-baseline.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-01.md</Path>`

## 1. 战略与来源

- **目标：** 在结构迁移前建立登录、菜单、权限、代表性 CRUD、上传、任务、启动和关闭的 characterization，使后续 Ticket 能区分兼容回归与预期新增行为。
- **可观察产出：** 一组在当前实现上可重复通过、在关键 HTTP/UI 行为漂移时失败的测试与固定 fixture。
- **来源：** `AC-002`、`AC-003`、`CODE:<Path>go-admin-plus/cmd/api/server.go</Path>`、`CODE:<Path>go-admin-ui-plus/tests/e2e/</Path>`。
- **当前事实：** 前端已有 mocked Playwright 和少量 API 正则检查，后端缺少覆盖完整启动/业务面的 characterization；当前 `go test ./...` 还会因 KODO 缺少 `test.png`、OBS 缺少真实 endpoint 并触发 nil client panic 而失败。
- **Planning Depth 原因：** 跨两个子仓库和多条稳定接缝，但只增加测试，不改变生产合同。

## 2. 决策状态

### 已锁定决策

- 基线断言外部 HTTP、页面和持久数据，不固定私有函数或全量易变快照。
- 代表性 CRUD 使用 demo product；权限覆盖 admin 与受限角色；上传使用小型本地 fixture。
- KODO、OBS、OSS 默认测试必须使用本地 fake/contract fixture；真实云服务验证只能作为显式 opt-in integration suite，缺少凭据时不得 panic，也不得让默认套件依赖公网。

### 已采用的低影响假设

- 测试数据使用独立临时数据库和临时文件目录，不读取或修改仓库中的用户数据库。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| HTTP characterization、Playwright 核心流程、升级前 fixture、启动/关闭观察 | 现有路由、mock server、Playwright 配置与测试命令 | Application 重构、OpenAPI 生成、桌面或 Compose 实现 |

## 4. 要构建什么

测试执行者可从空临时状态启动当前 API 和 Admin，完成登录、菜单加载、授权拒绝、demo CRUD、上传和任务查询，并得到稳定的路径、方法、envelope、页面 URL 与持久化结果。错误凭据、无权限和后端不可达必须有明确观察；测试结束后进程和临时数据被清理。

## 5. 实现契约

- **入口或接缝：** Go `httptest`/进程测试、Playwright mocked/live project。
- **输入与输出：** 固定 fixture 请求与用户操作；输出为 HTTP envelope、页面状态、数据库/文件结果和进程退出。
- **公共接口变化：** 无。
- **不变量：** 不接触用户数据库；不把日志时间、随机 ID 等易变值固化为合同。
- **状态或数据流：** fixture -> 临时运行实例 -> API/UI 动作 -> 可观察断言 -> cleanup。
- **错误与失败行为：** 无法隔离数据或启动依赖时测试明确失败，不静默 skip。
- **兼容要求：** 当前实现应通过；后续仅由 Spec 批准的新增行为可更新预期。
- **安全与隐私要求：** fixture 不包含真实凭据或用户数据。

## 6. 执行路线

1. 盘点现有稳定命令和 fixture，先将 KODO、OBS、OSS 默认测试改为 hermetic contract，建立隔离临时数据启动方式。
2. 先增加 HTTP 正常/失败 characterization，再补 Admin mocked/live 核心流程。
3. 固定升级前 SQLite fixture 的生成来源与校验摘要。
4. 验证受控行为变化会使对应测试失败，再恢复基线。
5. 运行前后端定向套件并证明当前对象存储基线失败已经消除，且未以 skip 或删除断言伪造绿色。

## 7. 路径访问契约

- **预计修改点：** 与 frontmatter `expected_changes` 一致。
- **可写范围：** 仅 characterization、对象存储测试及其 testdata、E2E 和 fixture 测试目录。
- **只读上下文：** 当前后端业务、启动入口与前端源码。
- **共享路径：** 无。
- **保留或不动：** 产品代码、配置、数据库、现有用户改动和部署 workflow。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | HTTP + Playwright | 定向 Go test、`pnpm e2e` | 核心流程通过 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-01.md</Path>` |
| 失败路径 | 认证/授权/不可达场景 | 运行错误凭据、受限角色和网络失败用例 | 稳定拒绝且无重复错误 | 同上 |
| 回归 | 全量现有测试 | Speculo config 的 test/typecheck/lint | 不新增失败 | 同上 |

- **Workspace checks：** 按 Goal Plan 选择的 workspace 运行定向 Go/Playwright、type-check 和 lint。
- **E2E disposition：** required；本 Ticket 定义后续行为基线。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；运行 mocked 与可用的 live 核心场景。
- **Integration evidence：** 记录 implementation/source commit、parent before、适用 candidate/result SHA 和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 不适用；仅测试资产。
- **兼容窗口：** 基线在整个 change 中保持，批准的合同演进必须同步更新。
- **监控信号：** CI characterization 与 E2E 状态。
- **回滚或前向恢复：** 回滚新增测试文件即可；不得删除测试来适配实现回归。
- **不可逆操作与批准点：** 无。
- **收缩条件：** 不适用；基线测试作为长期回归资产保留。

## 10. 验收标准

- [x] `AC-002`、`AC-003` 的主要正常和失败行为均有可重复基线。
- [x] 测试数据与用户工作区隔离，cleanup 可判定。
- [x] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-01.md</Path>`。
- [x] 修改未超出 `writable_paths`。
- [x] 按 Goal Plan 形成非空 implementation/source commit，并通过 direct-parent 或 candidate 验证。
- [x] required E2E 由 Lead 在规定环境执行。
- [x] Ticket、Map 和 Evidence 状态一致，无未批准偏差。
