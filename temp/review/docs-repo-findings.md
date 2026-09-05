# 文档、仓库资产与治理脚本 Review

审查范围：根 README/CHANGELOG/Taskfile、`docs/`、`database/`、`deploy/`、`release/` 文档，Task/CI/发行/质量脚本，Speculo 运行时状态，以及可识别的仓库冗余资产。验证命令：`node scripts/quality/docs-check.mjs`、`node scripts/quality/architecture-check.mjs`、`node scripts/quality/compatibility-zero.mjs` 均返回 PASS；这些门禁通过不代表以下问题不存在。

## Findings

### [P1] 已删除的命令仍出现在面向开发者的文档中

- **文件/行号**：`go-admin-plus/README.md:3-10`、`database/README.md:5`、`docs/repository-architecture.md:10-16`。
- **证据**：这些文档分别把 `cmd/config-check`、`cmd/migrate` 描述为正式入口，数据库文档还说二者复用迁移 registry；实际 `go-admin-plus/cmd/` 只有 `go-admin-plus` 和 `desktop-sidecar`。`go-admin-plus/internal/application/architecture_test.go:218-228` 明确把 `cmd/config-check` 与 `cmd/migrate` 定义为必须不存在的 legacy commands。
- **影响**：新开发者会按照不存在的路径调用或查找入口；架构图、命令清单和迁移说明互相矛盾。该项目声明是全新发布且不保留兼容代码，这属于发布文档错误。
- **建议**：删除上述两项旧入口，统一描述 `cmd/go-admin-plus` 的 `migrate`、`bootstrap`、`recover-admin` 子命令；更新架构树和数据库文档中的调用流，并给 `docs-check` 增加“禁止已删除路径”规则。

### [P1] 架构门禁仍允许被删除的后端命令目录重新进入仓库

- **文件/行号**：`scripts/quality/architecture-check.mjs:325-330`。
- **证据**：`allowed` 集合仍包含 `config-check`、`migrate`，而当前真实命令只有 `desktop-sidecar`、`go-admin-plus`；同仓库 Go 架构测试又在 `internal/application/architecture_test.go:218-228` 断言前两者必须不存在。
- **影响**：`task architecture:check` 会接受重新添加的旧命令目录，导致“零兼容”架构合同和质量门禁失效；现在之所以 PASS 只是因为检查逻辑把旧目录列为允许项，而不是因为合同正确。
- **建议**：把允许集合收敛为当前两个真实入口，增加回归测试验证旧目录一旦出现就失败；同步删除文档中的旧入口说明。

### [P1] Compose 文档声称要求固定镜像摘要，但检查和默认配置没有落实

- **文件/行号**：`deploy/README.md:8`、`deploy/compose/compose.yml:4,22`、`scripts/release/linux/verify-policy.mjs:8-19`。
- **证据**：部署文档要求生产使用“已验证摘要的镜像”；Compose 对 API/Web 使用 `${GO_ADMIN_*_IMAGE}: ${GO_ADMIN_VERSION:-dev}` 可变 tag，只有 PostgreSQL 镜像带 digest。验证器只断言 Compose 文本中任意位置匹配一个 `@sha256:`，没有逐个约束 Server 和 Web 镜像引用必须是 digest。
- **影响**：生产操作员按现有入口可使用 `go-admin-plus-server:dev` 或任意可变 tag，供应链完整性与可复现部署并没有被门禁保护，文档承诺与实际安全合同不一致。
- **建议**：让生产 Compose 配置要求镜像变量为完整 `image@sha256:<64 hex>`，或在部署前显式解析并拒绝非 digest 引用；验证器逐项检查 API/Web/PostgreSQL，而不是“任意一处有 digest”。文档应说明开发默认 tag 仅限本地。

### [P2] Linux 发布指南声称归档包含 Compose 配置，构建脚本实际没有打包

- **文件/行号**：`docs/release.md:19-23`、`release/README.md:3-7`；`scripts/release/linux/build-service.sh:21-31`。
- **证据**：发行指南写明 Linux service archive 包含“current Compose configuration”，`release/README.md` 也把 `release/linux/` 描述为包含 Compose configuration；`build-service.sh` 只复制两个 profile JSON、两个 systemd unit 和 `SERVER-INSTALL.md`，未复制 `deploy/compose/compose.yml`、`.env.example`、`compose/config` 或 `deploy/README.md`。`release/linux/README.md:3-5` 的归档内容清单也没有 Compose。
- **影响**：用户下载归档后按文档寻找 Compose 文件会失败，且文档没有说明 Compose 需要从源码/仓库另行取得。
- **建议**：二选一：从文档中删除“归档包含 Compose”的表述；或在构建脚本中明确打包完整 Compose 目录并为其建立校验/版本证据。建议采用前者，保持 service archive 与容器部署资产边界清晰。

### [P2] macOS 数据目录说明在发行总览与平台安装文档之间冲突

- **文件/行号**：`docs/release.md:25-30`、`release/macos/README.md:11-13`、`release/macos/INSTALL.md:10-13`。
- **证据**：总览说 macOS 与 Windows 都把状态放在“selected installation directory/data/”；macOS 专文说路径是 `<install-directory>/Go Admin Plus.app/data`，安装文档也明确是 app bundle 内部。实现 `go-admin-plus-ui/apps/admin-desktop/src-tauri/src/main.rs:530-547` 在 macOS 以可执行文件反推 `.app` 根，再写入 `.app/data` 和 `.app/logs`。
- **影响**：用户按总览备份或替换时会漏掉 macOS app bundle 内数据，可能造成数据库、文件和 Stronghold 会话丢失；跨平台文档无法作为可靠恢复指引。
- **建议**：总览按平台拆开写：Windows 为安装目录下 `data/`、`logs/`；macOS 为 `Go Admin Plus.app/data/`、`Go Admin Plus.app/logs/`。所有安装/备份文档统一使用同一术语。

### [P2] 根 Task 在 Windows 依赖未记录的 POSIX shell

- **文件/行号**：`scripts/go-admin-plus/run-script.mjs:27-28`；`docs/development.md:8-18`。
- **证据**：Taskfile 所有公共任务通过 `run-script.mjs` 调用 `.sh`；Windows 分支硬编码使用 `sh.exe`，除非设置 `GO_ADMIN_POSIX_SHELL`。开发前置环境只列 Go/Task/Node/pnpm/Rust，没有 Git Bash、MSYS2 或其他 POSIX shell 要求。
- **影响**：在只安装文档所列工具的 Windows 环境执行 `task test/lint/build/migrate` 会报 `required POSIX shell is not installed`；“Windows 支持”与可执行开发入口不一致。
- **建议**：要么在 Windows 开发前置条件中明确安装并配置 `sh.exe`（验证路径和版本），要么为 Task 入口提供等价的 PowerShell 实现并在脚本中按平台选择；同时在 README 的安装步骤中说明该前置条件。

### [P2] 文档质量门禁没有覆盖全部当前文档

- **文件/行号**：`scripts/quality/docs-check.mjs:9-13,110-125`。
- **证据**：`documentationRoots` 只扫描根 README、`.agents/skills`、`docs`、少数 deploy/database/release README 和两个后端 README；没有扫描 `CHANGELOG.md`、`NOTICE.md`、`CLAUDE.md`、`release/manifest/README.md`、`release/shared/sidecar/README.md`、`release/*/INSTALL.md`、`go-admin-plus-ui/**/README.md`，也没有覆盖 Speculo 管理的 workflow 文档。脚本随后只对这些 roots 执行链接和禁用词检查。
- **影响**：未覆盖文档中的错误链接、旧命令、过期产品名称不会阻断 `task docs:check`；本次已发现的旧命令引用恰好落在被扫描文件内，但同类问题可以继续藏在未扫描的安装/发行文档中。
- **建议**：以 `git ls-files '*.md'` 为来源扫描全部产品文档，显式排除仅供 Speculo 模板/示例的受管资产；至少把 `release/**`, `go-admin-plus-ui/**/README.md`, 根级 CHANGELOG/NOTICE/CLAUDE 纳入链接和旧路径检查，并为忽略清单写原因。

### [P3] Speculo 状态为空但仓库保留大量未激活的临时 worktree 目录

- **文件/路径**：`speculo/.speculo/specdev/status.json`（`active: []`, `archived: []`）；`specdev-worktree/`（工作树中的未跟踪、被忽略目录）。
- **证据**：`git ls-files specdev-worktree` 返回 0，`.gitignore:24` 整体忽略该目录；实际目录下存在 `.integration/.probe` 和大量按 ticket 复制的完整仓库树。该目录不属于当前发布源，也不被 SpecDev 当前状态引用。
- **影响**：本地仓库体积、搜索范围和备份成本显著增加，容易误把历史试验工件当作当前实现；CI 虽不会提交它，但开发者的全仓扫描和安全工具可能误扫副本。
- **建议**：在发布/交付前清理 `specdev-worktree/`、`.artifacts/`、`.data/` 等本地运行产物；保留 `.gitignore` 规则和 `.gitkeep`，不要把临时 ticket 仓库作为源代码资产。若需要审计历史，将其归档到外部存储而非工作区。

## 复核结论与清理清单

1. 先修正 P1：统一 CLI/目录文档、收紧架构门禁、落实镜像 digest 校验；这三项会影响新用户正确启动、架构治理和部署供应链。
2. 再修正 P2：删掉或实现 Linux Compose 归档承诺，统一 macOS 数据路径，补充 Windows shell 前置条件，扩大 `docs-check` 覆盖面。
3. 发布前清理被忽略的 `specdev-worktree/` 与本地运行目录；它们不是 Git 追踪文件，删除不会影响产品源代码，但应先确认没有未归档的 SpecDev 证据需要保留。
4. 当前三项已有门禁均通过：`docs-check`、`architecture-check`、`compatibility-zero`。这说明现有门禁无法发现上述目录/合同不一致，修复后应新增针对性测试，而不是只依赖当前 PASS 结果。
