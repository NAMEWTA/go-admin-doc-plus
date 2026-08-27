---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-13
title: Files 本地存储安全垂直切片
status: in_progress
planning_depth: deep
planning_depth_reason: 文件写入、授权下载、路径和符号链接逃逸属于关键安全与数据完整性边界
ready: true
risk: critical
blocked_by: [T-07]
contract_ids: [AC-020, AC-035]
owner: codex-t13-files
expected_changes: ["<Path>contracts/openapi/modules/files.yaml</Path>", "<Path>go-admin-plus/internal/modules/files/**</Path>", "<Path>go-admin-plus-ui/packages/domains/files/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/files/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/files.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/files.spec.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/index.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/package.json</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/files.yaml</Path>", "<Path>go-admin-plus/internal/modules/files/**</Path>", "<Path>go-admin-plus-ui/packages/domains/files/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/files/**</Path>", "<Path>go-admin-plus/test/files/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/files/**</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/files.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/files.spec.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/index.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/package.json</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/modules/iam/authorization/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>", "<Path>go-admin-plus-ui/packages/ui/**</Path>"]
shared_paths: ["<Path>go-admin-plus-ui/packages/domains/files/package.json</Path>", "<Path>go-admin-plus-ui/packages/web-domains/files/package.json</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/src/index.ts</Path>", "<Path>go-admin-plus-ui/packages/adapters/browser/package.json</Path>"]
shared_path_owners: ["<Path>go-admin-plus-ui/packages/domains/files/package.json</Path> => T-13 under T13-D01; package-local exports/dependencies/checks only", "<Path>go-admin-plus-ui/packages/web-domains/files/package.json</Path> => T-13 under T13-D01; package-local exports/dependencies/checks only", "<Path>go-admin-plus-ui/packages/adapters/browser/src/index.ts</Path> => T-13 under T13-D01; Files transfer export only", "<Path>go-admin-plus-ui/packages/adapters/browser/package.json</Path> => T-13 under T13-D01; Files adapter dependency/check only"]
---

# Ticket T-13: Files 本地存储安全垂直切片

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/13-files-module.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-13.md</Path>`

## 1. 战略与来源

- **目标：** 交付三个 profile 通用的授权上传、元数据和受控读取/下载。
- **可观察产出：** 合法文件可持久化并下载；越权、绝对路径、`..` 和符号链接逃逸均失败且不写出根目录。
- **来源：** `US-012`、`AC-020`、`AC-035`、`ADR-009`、`ADR-013`。
- **当前事实：** 旧 static/上传实现未形成模块所有权与强路径边界。
- **Planning Depth 原因：** 路径逃逸会造成任意文件读写。

## 2. 决策状态

### 已锁定决策

- 首期仅本地存储 adapter；元数据由 Files 模块数据库拥有，内容根来自类型化配置/Tauri。
- 外部只使用 opaque file id，不接受调用方物理路径。

### 已采用的低影响假设

- 写入采用临时文件、大小/类型检查和原子落位。
- Files 构造函数接收并自行校验 canonical absolute private root；模块不读取 env/config，也不依赖旧 platform files helper。Server typed root 与 Desktop `DataDirectory/files` 的产品注入归 T-17/T-16 后续 amendment。
- 上传 handler 先完成 Session/CSRF，再通过 `MaxBytesReader` 与 `MultipartReader` 流式解析；只绕过会物化 multipart body 的通用 validator 路径，其他生成路由和错误模型保持严格合同。
- 元数据采用 `pending/ready/deleting` 状态；本地 adapter 使用 anchored/no-follow 文件操作、fsync 和原子发布，启动 reconciliation 幂等收敛中断状态。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 上传、元数据、授权下载、本地 adapter、页面 | IAM、profile config、共享 UI | S3/OSS、公共静态目录 |

## 4. 要构建什么

授权用户上传流经大小/类型/路径校验后原子写入受控根并事务保存元数据；下载以 file id 重新授权和解析，任何逃逸尝试都在访问外部路径前失败。

## 5. 实现契约

- **入口或接缝：** Files API/use cases/repository、LocalStorage Port/adapter、Web page。
- **输入与输出：** multipart/stream 与 file id；返回元数据/受控 stream 或稳定错误。
- **公共接口变化：** 新 Files fragment、storage Port 和 Permission Code。
- **不变量：** 物理路径不出边界；元数据/内容一致；无 tenant；下载每次授权。
- **状态或数据流：** upload -> validate/temp -> atomic move -> metadata transaction -> authorized download。
- **错误与失败行为：** 越权/逃逸/超限/中断清理临时文件且不写出根目录。
- **兼容要求：** 不暴露或迁移旧 static 路径/API。
- **安全与隐私要求：** canonical path、symlink、TOCTOU、MIME/size 和日志路径脱敏测试。

## 6. 执行路线

1. 建立路径/符号链接/越权/中断负向测试。
2. 实现 migration、repository、storage Port 和本地 adapter。
3. 实现上传/下载 API mapping 与前端页面。
4. 覆盖原子写、临时清理和重启持久化。
5. 运行三个 profile 文件安全 E2E。

## 7. 路径访问契约

- **预计修改点：** Files 独占路径。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** IAM、配置和共享 UI。
- **共享路径：** 两个预建 Files manifest，以及 Browser adapter 的 export/manifest，仅在 `T13-D01` 精确范围内由本 Ticket 拥有。
- **批准偏差：** `T13-D01` 只开放两个预建 Files package manifest，以及 Browser adapter 的 Files 专属实现/测试、单一 export 和 package-local dependency/check；根 package、Vitest、lockfile、共享 UI、Desktop adapter/Rust bridge、typed Server config 和 composition 继续只读，等待 Lead 串行 amendment。
- **保留或不动：** Desktop 宿主路径注入由 T-16 完成。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | API/UI/storage suite | `task test -- files` | 上传、元数据、下载和重启一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-13.md</Path>` |
| 失败路径 | security suite | 越权、`..`、绝对路径、symlink、超限和中断 | 全部失败且授权根外零写入 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-13.md</Path>` |
| 回归 | profile E2E | 三 profile 上传/重启/下载 | 内容和元数据持久且语义一致 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-13.md</Path>` |

- **Workspace checks：** Goal Plan 选定的 current-workspace 或 source-worktree 非 E2E 检查。
- **E2E disposition：** required：真实文件系统、API/UI 与重启持久化必须验证。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不声明通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、文件 E2E 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** schema/adapter/API/UI 同切片，T-16 注入 Desktop 根，T-17 组合。
- **兼容窗口：** 无旧 static 文件迁移。
- **监控信号：** upload reject、orphan temp、download deny 和 storage failure。
- **回滚或前向恢复：** 保留内容根和元数据备份，修复后对账前向恢复。
- **不可逆操作与批准点：** 内容删除需确认并先提交元数据策略；批量清理需独立批准。
- **收缩条件：** T-21 证明旧 static/upload 路径零引用。

## 10. 验收标准

- [ ] `AC-020`：合法文件闭环和全部逃逸/越权负向合同通过。
- [ ] `AC-035`：文件列表/上传/删除确认和刷新符合共享交互。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-13.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 integration result SHA。
- [ ] E2E disposition 已执行且 shared path 无越权写入。
- [ ] Ticket、Map 和 Evidence 一致且无未批准偏差。
