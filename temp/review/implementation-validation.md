# 实施验证报告

## 已完成

- 后端删除用户统一走 deletion record、claim、transfer/purge、outbox 和 Files consumer 生命周期，移除 `DeleteUser`/`DeleteUsers`。
- Files 使用绑定 storage root 的真实磁盘容量探针，Unix 使用 `Statfs`，Windows 使用 `GetDiskFreeSpaceEx`；容量探针失败时拒绝 reservation。
- 审计 bootstrap/recovery 使用正式 outbox payload，移除 action/reason legacy fallback。
- 移除 `iam_role_data_scopes` 双数据库迁移、provider、测试及产品注册，`iam_roles.data_scope` 成为唯一事实源。
- Server 增加 `ReadHeaderTimeout`、`IdleTimeout`、`MaxHeaderBytes`。
- Desktop 首次设置密码默认为空，拒绝常见弱密码，提交后清空表单。
- ProductWorkspace 增加 session generation、AbortController 和 stale response 防护；Web 域客户端统一使用 session-aware transport。
- 架构命令 allowlist、docs 扫描范围、Linux Compose digest policy 和发行文档已收敛到当前契约。
- 创建实现分支 `codex/release-hardening`。

## 通过的门禁

- `go test ./...`
- `go test -tags sqlite ./...`
- `go vet ./...`
- `pnpm lint`
- `pnpm typecheck`
- `pnpm exec vitest run --config tests/shell/vitest.config.ts --maxWorkers=1`（30 files / 213 tests）
- `node --test tests/e2e/desktop/run.test.mjs`（22/22）
- `task architecture:check`
- `task compatibility:zero`
- `task docs:check`
- `task contract:lint`
- `task generate:check`
- `task governance:check`
- `task task:contract`
- `task release:verify`
- `task build TARGET=all PROFILE=server-sqlite`

## 环境限制

- Web CDP runner 已使用本机 Chrome 启动入口；当前环境没有可用 PostgreSQL disposable DSN，runner 在 PostgreSQL host 启动前退出，未宣称 SQLite/PostgreSQL 双 profile E2E 通过。
- Desktop native runner 明确要求 macOS；当前 Windows 环境只完成 native harness 静态/契约测试，未宣称原生运行通过。
- IAM/Audit 页面现有手写 dialog 标记保留，当前共享 UI FormDialog 未强制替换；Element/现有键盘语义测试仍通过，但完整焦点陷阱验收应在下一次可运行浏览器环境补齐。

## 清理

- `.data/` 和 `.artifacts/` 本地运行/构建产物已删除；发布构建期间生成的 `.artifacts/` 也已再次删除。
- `specdev-worktree/` 已确认无 Speculo 状态引用，但其中包含大量共享 `node_modules` 链接；Windows `rd`、robocopy 和 `git clean` 在链接竞态下未能完成根目录删除，目录仍需在无相关进程的维护环境中删除。
