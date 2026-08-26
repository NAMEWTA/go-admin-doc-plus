---
schema_version: 3
artifact: spec
change: 2026-08-24-fork-branch-strategy
status: ready
ready_for_tickets: false
sources:
  - USER-DECISION:采用 master 上游镜像与 main 二次增强默认主线并立即实施
---

# Spec: 建立 fork 双主线与可追溯上游同步模型

- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-fork-branch-strategy/spec.md</Path>`
- **当前 ADR：** 不适用，本 change 不新增长期架构 ADR
- **当前领域上下文：** 不适用，本 change 不新增领域术语

## 1. 问题与目标

### 问题陈述

两个 fork 当前只有与官方相同的 `master`。若直接在其上开发，官方更新与自有提交会混合，难以判断分支祖先、限制快进同步并持续解决相同冲突。

### 目标用户与场景

仓库维护者需要在后端和前端 fork 上持续开发自有增强，同时定期接收官方 `upstream/master`，并在 Git 提交图中保留明确的同步 merge 节点。

### 成功标准

两个 fork 都以 `main` 为默认二次增强主线；`master` 保持官方镜像职责；本地配置优先向 `origin` 推送且阻止隐式非快进 pull；CI 验证覆盖新的默认分支；父仓库固定可复现的子仓库提交。

### 非目标

不提交或丢弃后端现有用户改动；不改变生产部署来源；不自动合并未来上游更新；不新增永久 `develop` 分支；不删除 fork 已有分支。

## 2. 解决方案与外部行为

### 解决方案摘要

保留本地和远端 `master` 作为 `upstream/master` 的快进镜像，从当前官方提交创建并推送 `main`，将 fork 默认分支切换到 `main`。日常功能从 `main` 分出；未来上游同步先快进 `master`，再以 `--no-ff` merge 到 `main`。

### 主要流程

后端先在保留未提交工作区内容的前提下切换到新 `main`；前端抓取上游 refs 后创建 `main`。两个 `main` 增加最小 CI 触发配置并推送，然后切换 GitHub 默认分支。父仓库记录两个子仓库新的 `main` 提交。

### 边界、失败与稳定错误行为

任一远端写入、默认分支切换或核验失败时停止后续依赖动作并保留当前分支，不强推、不重置。GitHub 分支保护若无法安全识别 CI context，则只启用禁止强推/删除的基础保护，不猜测 required status 名称。

### 状态转换与不变量

`master` 不包含自有提交；`main` 必须包含创建时的 `master`；后端实施前已有工作区改动在实施后保持相同；父仓库 gitlink 只记录已经推送到对应 fork 的提交；部署 workflow 不因本 change 改为从 `main` 部署。

## 3. 用户故事

- **US-001**：作为 fork 维护者，我希望官方镜像与二次开发主线分离，以便持续同步上游而不污染官方镜像。
- **US-002**：作为开发者，我希望 `main` 上的 PR 有现有 CI 验证，以便默认分支切换后不失去质量门禁。
- **US-003**：作为父仓库使用者，我希望 submodule 固定已推送的前后端提交，以便复现配套版本。

## 4. 验收合同

| ID | 前置条件 | 动作或事件 | 可观察结果 | 验证接缝 |
|---|---|---|---|---|
| DS-001 | 两个 fork 的 `master` 与各自上游一致 | 创建并推送 `main` | `origin/main` 存在且包含对应 `master` | `git ls-remote` 与 `git merge-base --is-ancestor` |
| DS-002 | `origin/main` 已存在 | 修改 fork 默认分支 | GitHub API 返回默认分支 `main` | `gh repo view --json defaultBranchRef` |
| DS-003 | 本地仓库已配置双远端 | 应用安全配置 | push 默认指向 `origin`，pull 只允许快进，fetch prune 与 rerere 开启 | `git config --get` |
| DS-004 | `main` 成为开发主线 | 提交最小 CI 触发调整 | 非部署构建/校验 workflow 覆盖 `main`，部署来源保持不变 | workflow diff 与 YAML 静态检查 |
| DS-005 | 后端存在用户未提交内容 | 切换分支并提交 CI 文件 | 原有删除、新增文件仍保持未提交且内容状态不变 | 实施前后 `git status --porcelain=v2` 对比 |
| DS-006 | 子仓库新提交已推送 | 更新父仓库 gitlinks | 父仓库记录两个可从远端到达的提交 | `git submodule status`、`git ls-remote`、父仓库提交 |

## 5. 范围

### IN

两个子仓库的 `main` 分支创建与推送、默认分支切换、仓库级 Git 安全配置、最小 CI 触发更新、基础分支保护、父仓库 gitlink 更新与核验。

### REUSE

复用现有 `origin`/`upstream` remote、现有 CI job 与父仓库 submodule 固定提交机制。

### OUT

- **OOS-001**：不提交后端现有 `.go-version`、`temp` 或上传文件删除；所有权和意图尚未由本 change 决定。
- **OOS-002**：不修改部署 secrets、部署目标或生产部署分支。
- **OOS-003**：不创建 `develop`、release 或未来 feature 分支。
- **OOS-004**：不执行下一次尚不存在的上游 merge。

## 6. 已锁定实现约束

- **DEC-001**：`master` 是纯上游镜像，更新只允许 `--ff-only`。来源：`USER-DECISION:已批准方案`。
- **DEC-002**：`main` 是两个 fork 的默认二次增强主线，上游同步必须保留 merge ancestry。来源：`USER-DECISION:已批准方案`。
- **DEC-003**：不将 submodule 改为隐式跟踪分支，父仓库继续固定精确 gitlink。来源：`CODE:<Path>.gitmodules</Path>` 与已批准方案。

## 7. 数据、接口与兼容

- **公共接口变化：** Git 默认分支由 `master` 变为 `main`；业务 API 无变化。
- **数据模型与持久化：** 无业务数据变化；新增 Git refs、仓库配置和父仓库 gitlink 提交。
- **兼容要求：** 已有 `master` 保留，既有 clone 和固定 submodule SHA 继续可用。
- **迁移要求：** 开发者后续应 fetch `origin/main` 并以其作为功能分支基线。
- **发布或运维影响：** 默认 PR 目标改为 `main`；生产部署触发保持现状。

## 8. 非功能要求

- **NFR-001 安全与隐私：** 不输出或持久化 GitHub token；不配置向 `upstream` 的默认 push。
- **NFR-002 性能与容量：** 不适用；仅 Git 元数据操作。
- **NFR-003 可用性与可靠性：** 禁止强推、删除和隐式非快进 pull；失败时不得重写历史。
- **NFR-004 可观测性与运营：** merge ancestry、远端 HEAD、CI workflow 和父仓库 gitlink 均可通过稳定命令核验。

## 9. 验证策略

| 接缝 | 层级 | 覆盖合同 | 现有先例或命令 | Evidence 类型 |
|---|---|---|---|---|
| 本地 Git refs/config | 仓库集成 | DS-001、DS-003、DS-005 | `git branch -vv --all`、`git config --get`、`git status --porcelain=v2` | 命令输出摘要 |
| GitHub repository API | 远端集成 | DS-001、DS-002 | `git ls-remote`、`gh repo view` | 远端只读核验 |
| Workflow 配置 | 静态/CI | DS-004 | diff 检查与 GitHub Actions run | 配置 diff 和运行状态 |
| Submodule gitlink | 组合仓库 | DS-006 | `git diff --submodule=log`、`git submodule status` | 父提交与远端可达性 |

## 10. 风险、假设与未决问题

### 风险

默认分支切换会改变新 PR 的默认 base；CI required status 在首次 `main` 运行前可能尚未注册；后端脏工作区限制切回 `master`，因此实施期间保留在 `main`。

### 已采用的低影响假设

两个 personal fork 允许账户所有者修改默认分支和基础保护；若高级保护不可用，不将其作为分支创建的阻塞条件，并明确报告。

### 未决问题

无。
