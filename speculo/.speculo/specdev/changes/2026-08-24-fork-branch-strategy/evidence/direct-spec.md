# Evidence: Direct Spec — fork 双主线与上游同步治理

- **Change：** `2026-08-24-fork-branch-strategy`
- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-fork-branch-strategy/spec.md</Path>`
- **Goal Plan：** 不适用，用户批准 Direct Spec
- **Lead：** `root`
- **Workspace/branch：** current workspace / 父仓库及两个子仓库均为 `main`
- **实施前 checkpoint：** 父仓库 `b24d19b`；后端 `d34d30a`；前端 `5355aff`
- **最终 checkpoint：** 父仓库 `9924fb0`；后端 `01d9aa1`；前端 `1d1af04`
- **状态：** done

## 1. 实现摘要

两个 fork 均创建并推送 `main`，GitHub 默认分支和本地 `origin/HEAD` 均指向 `main`。`master` 保持与 `upstream/master` 完全一致。仓库级 Git 配置固定 `origin` 为 push 目标、pull 仅快进、fetch prune 与 rerere 开启。现有 workflow 覆盖 `main` 并支持手动触发；部署步骤继续只允许 `master` push。父仓库固定两个最终子仓库提交。

两个 `main`/`master` 禁止强推和删除。`main` 额外要求 PR、分支同步、评论解决和全部已验证 CI context；未启用 linear history，因此未来允许保留上游 merge commit。

## 2. Lead Dispatch And Candidate Return

- **Implementation owner：** Lead
- **Dispatch Packet/checkpoint：** Lead direct；未派遣 subagent
- **允许动作：** 创建/推送分支、修改 CI 触发、提交已授权治理变更、修改 fork 默认分支和保护、更新父仓库 gitlink
- **返回：** 后端 `01d9aa1`、前端 `1d1af04`、父仓库 `9924fb0`；所有提交已推送
- **Lead 独立核对：** pass；重读本地 refs、远端 refs、GitHub API、workflow checks、gitlink 与工作区状态
- **只读 Agent findings：** 无

## 3. 修改范围与路径所有权

| 路径 | 所有权 | 改动目的 |
|---|---|---|
| `<Path>go-admin-plus/.github/workflows/go.yml</Path>` | writable:Lead | `main` 构建/PR 验证和手动入口 |
| `<Path>go-admin-plus/.github/workflows/build.yml</Path>` | writable:Lead | `main` SQLite 构建验证，部署 guard 不变 |
| `<Path>go-admin-ui-plus/.github/workflows/verify.yml</Path>` | writable:Lead | `main` 静态、单测、契约、E2E 和手动入口 |
| `<Path>go-admin-ui-plus/.github/workflows/nodejs.yml</Path>` | writable:Lead | `main` 双 Node 版本构建 |
| `<Path>go-admin-ui-plus/.github/workflows/build.yml</Path>` | writable:Lead | `main` 生产构建验证，部署 guard 不变 |
| `<Path>go-admin-plus</Path>`、`<Path>go-admin-ui-plus</Path>` | shared:Lead | 父仓库精确 gitlink 更新 |

- **read-only 修改：** 无
- **未声明路径：** 无
- **生成文件/锁文件：** 无
- **保留用户改动：** 后端两个上传文件删除、`.go-version` 与 `temp/` 始终未暂存、未提交、未改写

## 4. 验收与合同映射

| Contract / Acceptance ID | 验证接缝 | 证据 | 结果 |
|---|---|---|---|
| DS-001 | Git refs/ancestry | 两个 `origin/main` 等于本地 `main`；`master...upstream/master` 均为 `0 0`；`master` 是 `main` 祖先 | pass |
| DS-002 | GitHub API | 两个 `defaultBranchRef.name` 均为 `main` | pass |
| DS-003 | 本地 Git config | `remote.pushDefault=origin`、`branch.master.pushRemote=origin`、`pull.ff=only`、`fetch.prune=true`、`rerere.enabled=true` | pass |
| DS-004 | workflow diff/checks | 5 个 workflow 覆盖 `main`；所有远端构建、静态、单测、契约和 E2E checks 通过；部署 steps skipped | pass |
| DS-005 | porcelain 状态对比 | 实施前后均为相同两个删除和两个未跟踪目标 | pass |
| DS-006 | 父仓库 gitlink/remote | `9924fb0` 固定后端 `01d9aa1`、前端 `1d1af04`，三者均已推送 | pass |

## 5. Workspace Verification

| 命令或步骤 | 运行环境 | 结果 | 摘要 |
|---|---|---|---|
| YAML parse + `git diff --check` | current-workspace | pass | 5 个 workflow 可解析，无空白错误 |
| `pnpm lint` | current-workspace | pass | 0 error、29 条既有 warning |
| `pnpm type-check` | current-workspace | pass | Vue/Node TypeScript 检查通过 |
| `pnpm test:unit` | current-workspace | pass | 27 files、220 tests 通过 |
| `go test ./...` | current-workspace | baseline-fail | `common/file_store` 依赖缺失的 `test.png` 和 OBS endpoint；其余已运行包通过，与仅 YAML diff 无因果关系 |
| 后端 GitHub `build` | GitHub Actions | pass | 普通 Go build 通过；镜像步骤按非 tag 跳过；`<Url>https://github.com/NAMEWTA/go-admin-plus/actions/runs/32687279627</Url>` |
| 后端 GitHub `Build` | GitHub Actions | pass | SQLite build 通过；部署步骤在 `main` 跳过；`<Url>https://github.com/NAMEWTA/go-admin-plus/actions/runs/32687279672</Url>` |
| 前端 GitHub `build` | GitHub Actions | pass | Node 22/24 生产构建通过；`<Url>https://github.com/NAMEWTA/go-admin-ui-plus/actions/runs/32687282698</Url>` |
| 前端 GitHub `Build CI` | GitHub Actions | pass | 生产构建通过，部署步骤在 `main` 跳过；`<Url>https://github.com/NAMEWTA/go-admin-ui-plus/actions/runs/32687282681</Url>` |
| 前端 GitHub `Verify` | GitHub Actions | pass | lint、types、220 unit、API contract 和 E2E 全部通过；`<Url>https://github.com/NAMEWTA/go-admin-ui-plus/actions/runs/32687282707</Url>` |

- **失败后修复与重跑：** 首次 Ruby YAML 探针因旧版 Ruby 不支持 `aliases:` 参数，在暂存前退出；改用兼容调用后通过。fork workflow 未注册时，Actions 权限端点先禁用再启用，随后注册成功并通过真实 push 验证。
- **未运行检查：** 无与本次 Git/CI 配置变更相关的必需检查。
- **E2E：** GitHub Actions 由 Lead 持续观察至终态，9 分 6 秒通过。

## 6. 双轴审查

### 标准轴

- **固定输入：** 后端 `d34d30a..01d9aa1`、前端 `5355aff..1d1af04`、父仓库 `b24d19b..9924fb0`
- **结果：** pass
- **Findings 与修正：** 默认分支切换会使原有只监听 `master` 的 CI 失效；已补齐 `main` 触发并验证。部署 guard 保持 `refs/heads/master`，避免无意生产动作。未发现 secrets、业务代码、依赖或 lockfile 变化。

### 规范轴

- **固定输入与来源：** 当前 Spec、Source、Triage 与用户实施授权
- **结果：** pass
- **Findings 与修正：** Direct Spec 全部 DS 合同有远端和本地证据；`master` 无自有提交；merge commit 能力未被 linear history 保护破坏；父仓库继续使用精确 gitlink。

## 7. Integration Verification

| 项目 | 结果 |
|---|---|
| Parent before SHA | `b24d19b` |
| Implementation/source SHA | 后端 `01d9aa1`；前端 `1d1af04` |
| Candidate branch/workspace | not-applicable；Direct Spec current workspace |
| Method/conflicts | 子仓库直接提交；父仓库 gitlink commit；无冲突 |
| Integration checks | GitHub Actions 五个 workflow 全绿；本地 refs/config/gitlink/API 重读通过 |
| E2E disposition | required：默认前端主线切换后需确认现有 E2E 门禁可运行 |
| E2E result | passed；GitHub Actions `e2e` 9 分 6 秒通过 |
| Parent result/re-read | `9924fb0`；HEAD 等于 `origin/main`，gitlinks 等于两个已推送最终 SHA |

## 8. 偏差与决策

- **偏差：** local；fork 初始未注册 Actions workflow，采用官方仓库 Actions 权限端点禁用/重新启用完成注册，并增加可逆的 `workflow_dispatch` 手动入口
- **记录：** 本 Evidence
- **批准来源及影响：** 属于已授权 CI 门禁实施范围；不改变业务、部署、数据或分支职责

## 9. 残余风险与交付定位

- **残余风险/已知限制：** 后端完整本地测试存在上游既有对象存储环境失败；后端用户未提交文件仍需用户自行判定是否属于首个二开提交。父仓库的 SpecDev 本地状态尚未提交，按本地权威保留。
- **后续 Ticket：** 无
- **监控或回滚触发：** required context 名称只有在 workflow/job 重命名时才需同步更新保护规则；未来上游同步仍需人工解决实际冲突并验证
- **最终后端 commit：** `01d9aa1`
- **最终前端 commit：** `1d1af04`
- **父仓库 result：** `9924fb0`
- **Source workspace：** current workspace
- **Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-24-fork-branch-strategy/evidence/direct-spec.md</Path>`
