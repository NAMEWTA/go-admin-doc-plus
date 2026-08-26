---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-13
title: 汇合产品发布并收缩旧兼容契约
status: done
planning_depth: deep
planning_depth_reason: 跨三个仓库状态、三类产物、共享版本/API 契约和不可逆兼容收缩的最终 Gate
ready: true
risk: critical
blocked_by: [T-10, T-11, T-12]
contract_ids: [AC-002, AC-005, AC-006, AC-016, AC-018]
owner: root
expected_changes: ["<Path>Taskfile.yml</Path>", "<Path>release/manifest/**</Path>", "<Path>.github/workflows/product-release.yml</Path>", "<Path>go-admin-plus/.github/workflows/release-linux.yml</Path>", "<Path>go-admin-plus/.github/workflows/release-macos.yml</Path>", "<Path>go-admin-plus/.github/workflows/release-windows.yml</Path>", "<Path>go-admin-plus/cmd/api/**</Path>", "<Path>go-admin-plus/cmd/go-admin/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/**</Path>", "<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>"]
writable_paths: ["<Path>Taskfile.yml</Path>", "<Path>release/manifest/**</Path>", "<Path>.github/workflows/product-release.yml</Path>", "<Path>go-admin-plus/.github/workflows/release-linux.yml</Path>", "<Path>go-admin-plus/.github/workflows/release-macos.yml</Path>", "<Path>go-admin-plus/.github/workflows/release-windows.yml</Path>", "<Path>go-admin-plus/cmd/api/**</Path>", "<Path>go-admin-plus/cmd/go-admin/**</Path>", "<Path>go-admin-ui-plus/apps/admin/src/**</Path>", "<Path>go-admin-ui-plus/scripts/check-api-contract.mjs</Path>"]
read_only_paths: ["<Path>deploy/compose/**</Path>", "<Path>release/linux/**</Path>", "<Path>release/macos/**</Path>", "<Path>release/windows/**</Path>", "<Path>go-admin-plus/api/openapi/**</Path>", "<Path>go-admin-ui-plus/packages/**</Path>", "<Path>go-admin-ui-plus/domains/**</Path>"]
shared_paths: ["<Path>Taskfile.yml</Path>", "<Path>release/manifest/**</Path>", "<Path>.github/workflows/product-release.yml</Path>"]
shared_path_owners: ["<Path>Taskfile.yml</Path> => T-13", "<Path>release/manifest/**</Path> => T-13", "<Path>.github/workflows/product-release.yml</Path> => T-13"]
---

# Ticket T-13: 汇合产品发布并收缩旧兼容契约

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/13-product-release-and-contract.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-13.md</Path>`

## 1. 战略与来源

- **目标：** 让根仓库以固定 submodule 提交和同一版本汇合 Linux/Mac/Windows 产物，并在零调用证据后删除批准的旧入口。
- **可观察产出：** 一个 release manifest 可追溯全部源码、OpenAPI hash、镜像/安装包 digest、checksum/SBOM/signature；旧 AppRouters/request/component/正则契约入口不再被消费。
- **来源：** `AC-002`、`AC-005`、`AC-006`、`AC-016`、`AC-018`、`DEC-009`、`DEC-010`。
- **当前事实：** 父仓库只固定 submodule，没有统一产品构建；宽迁移 Ticket 有意保留兼容层。
- **Planning Depth 原因：** 最终跨平台发布 Gate 与 contract 删除不可逆且事故半径最大。

## 2. 决策状态

### 已锁定决策

- 根 manifest 记录 product version、前后端 SHA、OpenAPI hash、migration max version、bundle/publisher identity 和所有 artifact digest。
- 根 Task 只编排子仓库确定命令，不复制业务源码；desktop assets 是生成中间物。
- 只删除扫描为零且完整 contract/E2E 证明不再需要的兼容层；未满足条件的条目保留并记录，不以时间强制删除。
- 正式外部分发、生产部署和 source cleanup 仍需独立人工授权。

### 已采用的低影响假设

- 产品版本采用现有 tag 语义并由 release manifest 统一；具体版本号由发布时输入，不在代码中重复维护。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| root build/release orchestration、manifest、三产物汇合、contract scan/removal、全量 Gate | T-10/11/12 artifacts、OpenAPI/contracts、全部 E2E | 实际生产部署、商店发布、自动更新、归档 change |

## 4. 要构建什么

发布维护者选择固定父仓库提交并运行产品 release。编排验证 submodule clean/reachable，构建 Admin 一次并供 Web/Desktop 使用，校验 OpenAPI/hash/version，调用平台流水线并汇集不可变 metadata。随后扫描旧接口：只有 AppRouters、旧 request facade、component 物理解析和正则 contract 检查的消费者为零且所有 Gate 绿色时才收缩，最后重跑三 Host 合同与离线 E2E。

## 5. 实现契约

- **入口或接缝：** root Task、product release workflow、release manifest schema、contract scanners。
- **输入与输出：** 固定 parent/submodule refs、version、平台 artifact metadata；输出统一 manifest 和验证结论。
- **公共接口变化：** 删除内部兼容入口；外部 `/api/v1` 和页面行为不变。
- **不变量：** 同版本同 OpenAPI；manifest digest 指向真实 artifact；contract 删除前零消费者。
- **状态或数据流：** pin refs -> verify/generate -> platform builds -> collect/verify -> scan contract -> full suite -> release candidate。
- **错误与失败行为：** dirty/unreachable refs、hash/digest/version 不一致、任一平台失败或旧消费者非零均阻止 Ready；不部分发布。
- **兼容要求：** 三 Host characterization、旧数据升级、routeKey 和所有页面通过。
- **安全与隐私要求：** manifest 不含 secret；签名状态可验证但凭据不可见。

## 6. 执行路线

1. 定义 release manifest schema、root Task 和固定 submodule preflight。
2. 编排单次 Admin/OpenAPI 生成与 Linux/Mac/Windows 输入，汇集 digest/checksum/SBOM/signature。
3. 对旧 AppRouters/request/component/regex contract 运行调用点和运行时观察扫描。
4. 只收缩满足零调用条件的兼容代码；未满足项记录残余窗口。
5. 运行 OpenAPI 零 diff、三 Host contract、Compose、Mac/Windows 离线全量 Gate。
6. 生成 release candidate 报告；等待外部分发/生产部署人工批准。

## 7. 路径访问契约

- **预计修改点/可写范围：** 根编排/manifest/workflow、三个平台 workflow 的可查询 provenance/version 元数据接缝与明确旧兼容入口。实现预检确认 GitHub artifact API 不暴露 `workflow_dispatch` 输入，因此产品 Gate 必须从经过 GitHub 认证的 run display title 独立核验 root/frontend/version；该局部路径修正不改变平台构建、签名或分发策略。
- **只读上下文：** 三平台 release、Compose、OpenAPI/packages/domains。
- **共享路径：** 根 Task/manifest/product workflow 唯一 owner T-13；子平台构建与验证逻辑保持不变，仅 provenance/version 元数据接缝由 T-13 收敛。
- **保留或不动：** 未达零调用的兼容层、业务逻辑、历史 migration、真实凭据和生产环境。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | product release candidate | root Task + manifest verify + 三平台 Gate | 同版本/contract/digest 全绿 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-13.md</Path>` |
| 失败路径 | 漂移/旧调用/平台失败 | 注入 SHA/hash/version mismatch、保留 consumer | 阻止收缩和发布 | 同上 |
| 回归 | 全量三 Host | Go/pnpm/Compose/macOS/Windows suites | AC-001..018 最终覆盖 | 同上 |

- **Workspace checks：** root preflight、所有非 E2E config/test/type/lint/build/generation/scans。
- **E2E disposition：** required；最终跨仓库、容器、WebView、安装和数据升级 Gate。
- **E2E owner/environment：** Lead / parent-candidate；Linux Compose、macOS ARM64、Windows AMD64 原生环境。
- **Integration evidence：** implementation/source commit、parent before、candidate/result、submodule SHA、artifact digest 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 所有 expand/migrate Ticket 完成 -> 产物汇合 -> 零调用扫描 -> contract 删除 -> 全量验证。
- **兼容窗口：** 未达收缩条件的旧入口继续保留并有 owner；不得伪装 change 已完全 contract。
- **监控信号：** contract scans、OpenAPI diff、artifact manifest、三平台 E2E 和签名状态。
- **回滚或前向恢复：** contract 删除可在同候选恢复；已发布数据 schema 遵循平台 Ticket 的备份/前向恢复。
- **不可逆操作与批准点：** 删除公共兼容、签名外部分发、生产 Compose 部署、source cleanup 均需人工批准；当前 planning 不授权。
- **收缩条件：** 旧调用点与运行观察为零、所有相关 AC 通过、release manifest 完整且 Lead 批准。

## 10. 验收标准

- [x] `AC-002`、`AC-005`、`AC-006`、`AC-016`、`AC-018` 通过，且 Map 中全部 AC 有最终 Evidence。
- [x] 三类产物绑定同一 version/submodule/OpenAPI，并有 digest/checksum/SBOM/signature。
- [x] 只收缩零调用兼容层；残余项明确记录。
- [x] 路径、提交、集成和 required 三平台 E2E 合同满足。
- [x] 未执行外部分发、生产部署或 cleanup，无未批准偏差。
