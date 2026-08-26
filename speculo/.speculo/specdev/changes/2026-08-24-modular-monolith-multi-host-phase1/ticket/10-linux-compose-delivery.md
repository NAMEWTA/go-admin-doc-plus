---
schema_version: 3
artifact: ticket
change: 2026-08-24-modular-monolith-multi-host-phase1
id: T-10
title: 交付 Linux AMD64 Web/API 生产 Compose
status: done
planning_depth: deep
planning_depth_reason: 重建生产容器、数据库迁移、网络、持久卷和安全配置，直接影响服务器部署与数据
ready: true
risk: high
blocked_by: [T-04, T-07]
contract_ids: [AC-012, AC-013, AC-014, AC-015, AC-016]
owner: root
expected_changes: ["<Path>deploy/compose/**</Path>", "<Path>release/linux/**</Path>", "<Path>go-admin-plus/Dockerfile</Path>", "<Path>go-admin-plus/.dockerignore</Path>", "<Path>go-admin-plus/.github/workflows/release-linux.yml</Path>", "<Path>go-admin-plus/internal/platform/files.go</Path>", "<Path>go-admin-plus/app/admin/apis/captcha.go</Path>", "<Path>go-admin-ui-plus/Dockerfile</Path>", "<Path>go-admin-ui-plus/.dockerignore</Path>", "<Path>go-admin-ui-plus/deploy/nginx/**</Path>"]
writable_paths: ["<Path>deploy/compose/**</Path>", "<Path>release/linux/**</Path>", "<Path>go-admin-plus/Dockerfile</Path>", "<Path>go-admin-plus/.dockerignore</Path>", "<Path>go-admin-plus/.github/workflows/release-linux.yml</Path>", "<Path>go-admin-plus/internal/platform/files.go</Path>", "<Path>go-admin-plus/app/admin/apis/captcha.go</Path>", "<Path>go-admin-ui-plus/Dockerfile</Path>", "<Path>go-admin-ui-plus/.dockerignore</Path>", "<Path>go-admin-ui-plus/deploy/nginx/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/host/server/**</Path>", "<Path>go-admin-plus/cmd/migrate/**</Path>", "<Path>go-admin-ui-plus/apps/admin/**</Path>", "<Path>go-admin-plus/docker-compose.yml</Path>", "<Path>go-admin-ui-plus/scripts/k8s/default.conf</Path>"]
shared_paths: ["<Path>deploy/compose/**</Path>", "<Path>release/linux/**</Path>"]
shared_path_owners: ["<Path>deploy/compose/**</Path> => T-10", "<Path>release/linux/**</Path> => T-10"]
---

# Ticket T-10: 交付 Linux AMD64 Web/API 生产 Compose

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/10-linux-compose-delivery.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-10.md</Path>`

## 1. 战略与来源

- **目标：** 用生产 Compose 交付 Web、API、Migration、PostgreSQL 和 Redis，并替换特权、内置演示数据的旧容器路径。
- **可观察产出：** Linux AMD64 空主机/空卷一条命令启动后浏览器同源登录；失败迁移阻止 API ready；容器最小权限检查通过。
- **来源：** `AC-012`、`AC-013`、`AC-014`、`AC-015`、`AC-016`、`CODE:<Path>go-admin-plus/docker-compose.yml</Path>`。
- **当前事实：** 旧 Dockerfile 复制预编译 main、演示 config/DB；Compose 只有 privileged API；前端镜像未安装代理配置。
- **Planning Depth 原因：** 生产部署、数据卷、secret 和迁移顺序高风险。

## 2. 决策状态

### 已锁定决策

- Web/Nginx 是唯一默认宿主入口，API 只在 Compose 网络；代理 `/api`、`/login`、`/static` 并提供 SPA fallback。
- `migrate` 一次性服务成功后 API 才启动；PostgreSQL/Redis 用 healthcheck 和持久卷，不使用固定 sleep。
- API/Web 均多阶段构建、非 root、无 privileged；配置/secret 运行时注入，不烘焙演示数据库。
- 首期镜像平台 `linux/amd64`，版本使用不可变 tag/digest；本地开发覆盖可另行暴露端口。

### 已采用的低影响假设

- Server FileStore 默认使用命名卷；对象存储配置存在时由现有 adapter 替换，不将对象存储服务加入 Compose。

### 未决问题

无。

## 3. 范围边界

| IN | REUSE | OUT |
|---|---|---|
| API/Web 镜像、Nginx、Compose、migration、Postgres/Redis、卷、health、安全和 Linux release | ServerHost、Admin dist、MigrationRunner | TLS 证书自动签发、HA、Kubernetes、ARM 镜像 |

## 4. 要构建什么

部署者准备非敏感环境文件和 secret，启动 Compose。数据库/缓存 healthy 后 Migration 执行；成功后 API ready，Web 对外提供页面与同源 API。重启保留数据库和文件；迁移/依赖失败时 Web 可显示不可用而 API 不伪 ready。发布生成镜像 digest、Compose bundle、checksum 和 SBOM。

## 5. 实现契约

- **入口或接缝：** `docker compose`、Nginx、ServerHost health、Migration CLI。
- **输入与输出：** env/secrets/volumes；输出 healthy services、HTTP Web/API 和 OCI artifacts。
- **公共接口变化：** 外部统一 Web origin；业务 `/api/v1` 不变。
- **不变量：** API 默认无 host port；数据卷不随容器删除；无 demo DB/secret image layer。
- **状态或数据流：** postgres/redis healthy -> migrate success -> api ready -> web proxy。
- **错误与失败行为：** migration 非零阻止 API；依赖丢失 ready 失败；Nginx 不把后端错误替换为 SPA 200。
- **兼容要求：** Web 深链和现有 `/login`/`/api` 请求保持。
- **安全与隐私要求：** 非 root、非 privileged、最小网络暴露、secret 不进日志/镜像。

## 6. 执行路线

1. 建立多阶段 API/Web 镜像和确定性 Nginx 配置测试。
2. 建立 Compose services、网络、卷、health 与 secret 注入。
3. 接入 migrate gate、Server readiness 和 Web 同源代理。
4. 运行空卷启动、重启持久、迁移失败和依赖中断场景。
5. 构建 linux/amd64 artifacts，生成 checksum/SBOM/digest 并验证非 root/无 privileged。

## 7. 路径访问契约

- **预计修改点/可写范围：** root Compose/Linux release、两个 Dockerfile 和专用 CI/Nginx。Linux run `32855385927` 在容器构建前证明 `os.Root.MkdirAll` 对 symlink parent 返回 `EEXIST`，而 macOS 返回 escape 错误；因此允许仅在 `internal/platform/files.go` 将父目录创建阶段的 Linux 错误归一化为既有 `ErrInvalidFileKey`，现有 symlink escape 测试和 FileStore 公共合同不变。生产登录验证进一步确认验证码 handler 会以 info 日志记录明文答案；为满足本 Ticket 的日志隐私合同，允许仅在 `app/admin/apis/captcha.go` 删除该日志，验证码生成、存储和响应合同不变。
- **只读上下文：** ServerHost、Migration、Admin 和旧部署作对照。
- **共享路径：** Compose/Linux release 由 T-10 唯一拥有；T-13 只消费产物 manifest。
- **保留或不动：** 旧线上部署 workflow、生产 secret、真实卷和桌面代码。

## 8. 验证矩阵

| 行为或风险 | 接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 空卷 Compose/Web E2E | config/build/up/login/restart | healthy 且数据保留 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/T-10.md</Path>` |
| 失败路径 | migrate/dependency/proxy | 失败迁移、停止 DB/Redis、后端 404 | 不伪 ready/SPA 200 | 同上 |
| 回归 | image/security/API | scan、characterization、Playwright live | 最小权限且 API/UI 绿色 | 同上 |

- **Workspace checks：** Dockerfile lint/build、`docker compose config`、Go/pnpm 非 E2E 门禁。
- **E2E disposition：** required；真实容器、网络、数据和浏览器边界。
- **E2E owner/environment：** Lead / Linux parent-candidate 或 current-workspace；空卷与失败注入。
- **Integration evidence：** 提交、candidate/result、image digest、SBOM/checksum 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 构建镜像 -> 备份外部数据 -> migrate -> API/Web；旧 Compose 不自动覆盖。
- **兼容窗口：** 旧部署文件保留到新 bundle smoke 通过，T-13 决定标记废弃/收缩。
- **监控信号：** container health、migration exit、API ready、Nginx upstream errors。
- **回滚或前向恢复：** 镜像回滚必须配合 schema 兼容；数据库先备份，已不可逆迁移优先前向修复。
- **不可逆操作与批准点：** 生产迁移、删除旧卷/部署文件均需人工批准，本 Ticket 不执行生产部署。
- **收缩条件：** 新 bundle 空卷/升级通过、旧部署无消费者并获批准。

## 10. 验收标准

- [ ] `AC-012`、`AC-013`、`AC-014`、`AC-015`、`AC-016` 通过。
- [ ] 空卷、持久重启、迁移失败、依赖失败和安全 Evidence 完整。
- [ ] 路径、提交、集成和 required E2E 合同满足。
- [ ] 未部署生产或写入 secret，无未批准偏差。
