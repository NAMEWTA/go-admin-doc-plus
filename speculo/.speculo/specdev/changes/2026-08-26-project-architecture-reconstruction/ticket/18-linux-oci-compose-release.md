---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-18
title: Linux OCI 与 Compose 发行切片
status: in_progress
planning_depth: deep
planning_depth_reason: 双架构生产镜像、迁移、secret reference、持久化和供应链证据影响正式部署
ready: true
risk: high
blocked_by: [T-17]
contract_ids: [AC-030, AC-033]
owner: codex-root
expected_changes: ["<Path>deploy/compose/**</Path>", "<Path>release/linux/**</Path>", "<Path>scripts/release/linux/**</Path>", "<Path>.github/workflows/release-linux.yml</Path>"]
writable_paths: ["<Path>deploy/compose/**</Path>", "<Path>release/linux/**</Path>", "<Path>scripts/release/linux/**</Path>", "<Path>.github/workflows/release-linux.yml</Path>"]
read_only_paths: ["<Path>Taskfile.yml</Path>", "<Path>go-admin-plus/cmd/go-admin-plus/**</Path>", "<Path>contracts/openapi/product.yaml</Path>", "<Path>release/shared/sidecar/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-18: Linux OCI 与 Compose 发行切片

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/18-linux-oci-compose-release.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-18.md</Path>`

## 1. 战略与来源

- **目标：** 从同一源码身份生成 Linux amd64/arm64 OCI，并用 Compose 验证正式 Server 运行合同。
- **可观察产出：** 两架构镜像迁移、健康、持久化和重启通过，Compose 无 Redis，并关联 checksum/SBOM/provenance。
- **来源：** `US-016`、`AC-030`、`AC-033`、`ADR-010`、`ADR-022`。
- **当前事实：** 现有 release/deploy 仍有旧路径与产品身份，缺少完整供应链 Gate。
- **Planning Depth 原因：** 生产镜像与迁移失败会影响部署和供应链可信度。

## 2. 决策状态

### 已锁定决策

- Linux 仅 Server OCI/Compose，不发布 Linux Desktop；支持 amd64/arm64。
- Compose 使用 secret reference，包含 PostgreSQL 与单实例 SQLite 场景且不包含 Redis。

### 已采用的低影响假设

- 多架构 manifest 只在两架构 smoke 均通过后生成。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| OCI、Compose、smoke、SBOM/provenance、专属 workflow | T-17 产品 binary/matrix | Kubernetes、Linux Desktop、发布到远端 registry |

## 4. 要构建什么

发行工程师从根任务构建相同源码/版本的双架构镜像，在隔离 Compose 中以 secret reference 迁移并启动，验证 ready、持久化、重启和 worker 语义后产生供应链证据。

## 5. 实现契约

- **入口或接缝：** root package/release 委派、Docker buildx、Compose smoke、artifact manifest。
- **输入与输出：** 源码/version/profile secret refs；OCI、checksum、SBOM、provenance 和 smoke Evidence。
- **公共接口变化：** 新 Linux 发行合同。
- **不变量：** 两架构同源码；非 root 运行；无内嵌 secret/Redis；Gate 失败不可 publishable。
- **状态或数据流：** source -> build -> scan/SBOM -> compose migrate/start -> smoke/restart -> provenance。
- **错误与失败行为：** 任一架构、迁移、健康或证据失败阻断候选。
- **兼容要求：** 不保留旧镜像名称/入口或 Redis Compose。
- **安全与隐私要求：** 最小镜像、固定 base digest、secret reference 和制品扫描。

## 6. 执行路线

1. 固定镜像身份、架构和 Compose policy 测试。
2. 实现双架构构建、最小 runtime 和两个 Server profile Compose。
3. 实现迁移/健康/持久化/重启 smoke。
4. 生成 checksum、SBOM、provenance 并接入专属 protected workflow。
5. 在原生容器 runner 完成发行候选验证。

## 7. 路径访问契约

- **预计修改点：** Linux deploy/release/script/workflow 专属路径。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** 根任务和 T-17 产品输出。
- **共享路径：** 无；根 CI 归 T-21。
- **保留或不动：** macOS/Windows 发行资产。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | native/container gate | `task package -- linux && task release:verify -- linux` | 两架构 smoke 与供应链证据通过 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-18.md</Path>` |
| 失败路径 | policy/failure fixture | 缺 secret、迁移失败、单架构失败或 Redis 服务 | 候选阻断且不标记 publishable | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-18.md</Path>` |
| 回归 | compose restart | PostgreSQL/SQLite 重启并检查持久化/ready | 数据保留且运行语义正确 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-18.md</Path>` |

- **Workspace checks：** current-workspace 或 source-worktree 运行静态/build/compose 配置检查。
- **E2E disposition：** deferred：双架构 OCI 与 Compose 运行场景保留到全部 Ticket 实现集成后的统一系统 E2E；本 Ticket 完成本地构建、配置、policy 和供应链非 E2E Gate。
- **E2E owner/environment：** Lead / 最终系统候选的 Linux/OCI 环境；逐 Ticket source-worktree 与 parent-candidate 不运行或声明 E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、本地非 E2E Gate、统一平台 E2E 引用与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** T-17 完整产品后构建候选，先 migration smoke 后启动与重启。
- **兼容窗口：** 无旧镜像/配置兼容。
- **监控信号：** build digest、scan、migration、ready、restart、SBOM/provenance。
- **回滚或前向恢复：** 未发布候选直接废弃；已部署只回滚到同新架构兼容版本或前向修复。
- **不可逆操作与批准点：** 远端 push/release 不在本 Ticket 授权内。
- **收缩条件：** T-21 汇总 Gate 并证明旧 Docker/Redis 资产零引用。

## 10. 验收标准

- [ ] `AC-030`：Linux 双架构 OCI/Compose、迁移、健康、持久化和供应链证据通过。
- [ ] `AC-033`：Linux required Gate 失败会阻断候选。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-18.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] 未执行远端发布，非 E2E 实现 Gate 已执行且平台场景已登记到最终统一矩阵。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
