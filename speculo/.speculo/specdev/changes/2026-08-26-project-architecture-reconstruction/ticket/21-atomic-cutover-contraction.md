---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-21
title: 原子切换、质量门禁与旧体系归零
status: in_progress
planning_depth: deep
planning_depth_reason: 全仓不可逆删除、目录改名、根 CI 与兼容归零是最终高风险收缩操作
ready: true
risk: critical
blocked_by: [T-18, T-19, T-20]
contract_ids: [AC-001, AC-002, AC-023, AC-028, AC-029, AC-033]
owner: codex-root
expected_changes: ["<Path>Taskfile.yml</Path>", "<Path>.gitignore</Path>", "<Path>.github/workflows/ci.yml</Path>", "<Path>.github/workflows/codeql.yml</Path>", "<Path>.github/workflows/product-release.yml</Path>", "<Path>.github/workflows/desktop-windows-tracer.yml</Path>", "<Path>scripts/contracts/generate.sh</Path>", "<Path>scripts/quality/**</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>scripts/go-admin-plus-ui/common.test.mjs</Path>", "<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>deploy/README.md</Path>", "<Path>database/README.md</Path>", "<Path>release/README.md</Path>", "<Path>release/manifest/**</Path>", "<Path>go-admin-ui-plus/**</Path>", "<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/common/**</Path>", "<Path>go-admin-plus/api/**</Path>", "<Path>go-admin-plus/template/**</Path>", "<Path>go-admin-plus/internal/profile/**</Path>", "<Path>go-admin-plus/internal/tenant/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>"]
writable_paths: ["<Path>Taskfile.yml</Path>", "<Path>.gitignore</Path>", "<Path>.github/workflows/ci.yml</Path>", "<Path>.github/workflows/codeql.yml</Path>", "<Path>.github/workflows/product-release.yml</Path>", "<Path>.github/workflows/desktop-windows-tracer.yml</Path>", "<Path>scripts/contracts/generate.sh</Path>", "<Path>scripts/quality/**</Path>", "<Path>scripts/go-admin-plus/**</Path>", "<Path>scripts/go-admin-plus-ui/common.sh</Path>", "<Path>scripts/go-admin-plus-ui/common.test.mjs</Path>", "<Path>scripts/go-admin-plus-ui/build.sh</Path>", "<Path>scripts/go-admin-plus-ui/dev.sh</Path>", "<Path>scripts/go-admin-plus-ui/lint.sh</Path>", "<Path>scripts/go-admin-plus-ui/package.sh</Path>", "<Path>scripts/go-admin-plus-ui/path-contract.sh</Path>", "<Path>scripts/go-admin-plus-ui/test.sh</Path>", "<Path>README.md</Path>", "<Path>docs/**</Path>", "<Path>deploy/README.md</Path>", "<Path>database/README.md</Path>", "<Path>release/README.md</Path>", "<Path>release/manifest/**</Path>", "<Path>release/linux/Containerfile.server.dockerignore</Path>", "<Path>release/linux/Containerfile.web.dockerignore</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/.gitignore</Path>", "<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/binaries/.gitignore</Path>", "<Path>go-admin-ui-plus/**</Path>", "<Path>go-admin-plus/.github/**</Path>", "<Path>go-admin-plus/.gitignore</Path>", "<Path>go-admin-plus/.dockerignore</Path>", "<Path>go-admin-plus/.go-version</Path>", "<Path>go-admin-plus/AGENTS.md</Path>", "<Path>go-admin-plus/README.md</Path>", "<Path>go-admin-plus/LICENSE.md</Path>", "<Path>go-admin-plus/_config.yml</Path>", "<Path>go-admin-plus/main.go</Path>", "<Path>go-admin-plus/restart.sh</Path>", "<Path>go-admin-plus/stop.sh</Path>", "<Path>go-admin-plus/go-admin-db.db</Path>", "<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>", "<Path>go-admin-plus/Makefile</Path>", "<Path>go-admin-plus/Dockerfile</Path>", "<Path>go-admin-plus/docker-compose.yml</Path>", "<Path>go-admin-plus/app/**</Path>", "<Path>go-admin-plus/common/**</Path>", "<Path>go-admin-plus/api/**</Path>", "<Path>go-admin-plus/template/**</Path>", "<Path>go-admin-plus/static/**</Path>", "<Path>go-admin-plus/ssh/**</Path>", "<Path>go-admin-plus/scripts/**</Path>", "<Path>go-admin-plus/docs/**</Path>", "<Path>go-admin-plus/cmd/api/**</Path>", "<Path>go-admin-plus/cmd/app/**</Path>", "<Path>go-admin-plus/cmd/config/**</Path>", "<Path>go-admin-plus/cmd/go-admin-desktop/**</Path>", "<Path>go-admin-plus/cmd/migrate/**</Path>", "<Path>go-admin-plus/cmd/version/**</Path>", "<Path>go-admin-plus/cmd/cobra.go</Path>", "<Path>go-admin-plus/config/README.md</Path>", "<Path>go-admin-plus/config/db*</Path>", "<Path>go-admin-plus/config/pg.sql</Path>", "<Path>go-admin-plus/config/extend.go</Path>", "<Path>go-admin-plus/config/seeds.go</Path>", "<Path>go-admin-plus/config/settings*.yml</Path>", "<Path>go-admin-plus/internal/application/*.go</Path>", "<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus/internal/host/desktop/**</Path>", "<Path>go-admin-plus/internal/modules/default.go</Path>", "<Path>go-admin-plus/internal/modules/default_test.go</Path>", "<Path>go-admin-plus/internal/modules/jobs.go</Path>", "<Path>go-admin-plus/internal/modules/runtime_queue.go</Path>", "<Path>go-admin-plus/internal/platform/dependencies.go</Path>", "<Path>go-admin-plus/internal/platform/dependencies_test.go</Path>", "<Path>go-admin-plus/internal/platform/files.go</Path>", "<Path>go-admin-plus/internal/platform/files_test.go</Path>", "<Path>go-admin-plus/internal/platform/cache/**</Path>", "<Path>go-admin-plus/internal/platform/localcache/**</Path>", "<Path>go-admin-plus/internal/platform/observability/**</Path>", "<Path>go-admin-plus/internal/profile/**</Path>", "<Path>go-admin-plus/internal/tenant/**</Path>", "<Path>go-admin-plus/test/characterization/**</Path>", "<Path>go-admin-plus/test/desktop/**</Path>", "<Path>go-admin-plus/test/api.go.template</Path>", "<Path>go-admin-plus/test/model.go.template</Path>", "<Path>go-admin-plus/test/gen_test.go</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/app/product/**</Path>", "<Path>go-admin-plus/internal/application/health/**</Path>", "<Path>go-admin-plus/internal/modules/*/**</Path>", "<Path>go-admin-plus/internal/platform/config/**</Path>", "<Path>go-admin-plus/internal/platform/coordination/**</Path>", "<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus/internal/platform/desktop/**</Path>", "<Path>go-admin-plus/internal/platform/migrations/**</Path>", "<Path>go-admin-plus/internal/platform/outbox/**</Path>", "<Path>go-admin-plus/cmd/go-admin-plus/**</Path>", "<Path>go-admin-plus/cmd/desktop-sidecar/**</Path>", "<Path>go-admin-plus/cmd/config-check/**</Path>", "<Path>go-admin-plus/config/schema/**</Path>", "<Path>go-admin-plus-ui/**</Path>", "<Path>release/linux/**</Path>", "<Path>release/macos/**</Path>", "<Path>release/windows/**</Path>", "<Path>release/shared/**</Path>", "<Path>deploy/**</Path>"]
shared_paths: ["<Path>.github/workflows/ci.yml</Path>", "<Path>scripts/quality/**</Path>", "<Path>README.md</Path>", "<Path>docs/**</Path>"]
shared_path_owners: ["<Path>.github/workflows/ci.yml</Path> => T-21", "<Path>scripts/quality/**</Path> => T-21", "<Path>README.md</Path> => T-21", "<Path>docs/**</Path> => T-21"]
---

# Ticket T-21: 原子切换、质量门禁与旧体系归零

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/21-atomic-cutover-contraction.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>`

## 1. 战略与来源

- **目标：** 在新产品与三平台候选全部验证后，一次性删除旧结构并启用最终根 CI/文档/归零门禁。
- **可观察产出：** 仓库只保留目标目录与治理；旧 API/schema/config/Wails/JWT/Casbin/Redis/tenant/旧数据库支持和临时标记零命中。
- **来源：** `US-001`、`US-016`、`US-018`、`AC-001`、`AC-029`、`AC-033`、`ADR-001`、`ADR-021`、`ADR-022`。
- **当前事实：** 用户已有部分根迁移改动，旧前后端与目标结构仍并存；不可回滚这些用户改动。
- **Planning Depth 原因：** 删除和命名切换不可逆且影响全仓构建、文档、发行与开发入口。

## 2. 决策状态

### 已锁定决策

- 只在 T-17 至 T-20 Gate 全通过后执行 contract；不发布混合结构或旧新双轨。
- 后端保留 `<Path>go-admin-plus/</Path>`，前端最终仅 `<Path>go-admin-plus-ui/</Path>`。
- Local/Hook、PR、Protected Release 复用根任务；豁免必须包含 owner、范围、原因、期限和 Ticket。

### T21-D01（Lead 批准）

G7 RED 盘点证明旧目录不是全部孤立文件：未被正式入口消费的 expand 期 `internal/modules` 根兼容集合、`profile`、旧 Desktop host、root platform adapter、legacy command/template/test 仍引用待删除的 `app/common/api/tenant`；根 Task、CI、Windows tracer 与聚合 release manifest 也仍调用旧 root main、`sqlite3` tag、Wails、旧前端名和 unsigned-self-use。若只按初始路径表删除 12 项治理文件与显式旧目录，全量构建立即失败，且 AC-002/AC-029/AC-033 会形成伪通过。

因此 Lead 基于用户已批准的“零兼容完整重构”精确开放 frontmatter 新增路径：删除不在 `cmd/go-admin-plus`、`cmd/desktop-sidecar` 和 `cmd/config-check` 正式依赖闭包中的遗留 scaffolding；`go mod tidy` 删除 Wails/go-admin-core/Casbin/Redis/MySQL/SQL Server 等失去消费者的依赖；根 Task/CI/docs/quality/release manifest 只引用新入口、三种 profile、双 App 和 T18-T20 平台 policy。新 product/module/platform 实现与三个正式 command 保持只读，任何实际需要修改只读产品代码的发现必须再次登记偏差。

### T21-D02（Lead 批准）

首次删除后的 compile RED 纠正了依赖分类：`internal/application` 与 `internal/contracts` 是 `product.Build` 的直接依赖，`localcache` 仍由可靠运行时验收消费，`cache/observability` 也是已完成的新平台能力；四者全部恢复且不得以“入口暂未调用”为由收缩。治理预检同时证明 `scripts/go-admin-plus-ui/{build,dev,package}.sh` 使用不存在的 `@go-admin/desktop` filter，Linux 两份 dockerignore 与 `<Path>go-admin-plus/config/README.md</Path>` 仍引用旧前端。Lead 只追加开放这 7 个治理文件：改为 `@go-admin/admin-desktop` 和当前 Workspace/目录合同；产品源码仍保持只读。

### T21-D03（Lead 批准）

根治理收口证明前端公共脚本的 workspace 根、lint 和路径合同仍消费已删除的旧目录，而 Tauri 2 App 内两份 `.gitignore` 会继续形成嵌套 Git 治理。Lead 精确追加开放 `scripts/go-admin-plus-ui/{common,lint,path-contract}.sh` 与两份 Tauri `.gitignore`：脚本统一指向 `<Path>go-admin-plus-ui/</Path>`，忽略规则迁移到根 `<Path>.gitignore</Path>` 后删除嵌套文件。Tauri/Rust/TypeScript 产品源码与其他前端文件继续只读。

### T21-D04（Lead 批准）

最终文档审计证明两份项目 Agent Skills 仍指导未来开发使用已删除架构，且零兼容与文档检查未覆盖 `<Path>.agents/skills/**</Path>`。Lead 精确开放该 Skill 目录与既有 T-21 quality/docs 接缝：以当前实现为权威重写业务模块和列表页工作流，并将 Skill 纳入永久扫描；不修改产品源码或恢复旧能力。

### T21-D05（Lead 批准）

全仓治理审计证明 T-18 冻结后的 `<Path>deploy/compose/runtime/.gitignore</Path>` 与 `<Path>deploy/compose/runtime/secrets/.gitignore</Path>` 仍形成第二层 Git 规则，而原治理扫描只覆盖前后端目录。Lead 在 T-18 implementation result 已进入 `main` 后串行接管这两个精确文件：规则迁移到根 `<Path>.gitignore</Path>` 后删除嵌套文件，并扩展既有 `<Path>scripts/go-admin-plus/governance-check.sh</Path>` 到整个产品仓库；T-18 其他 Compose、发行和运行时合同保持只读。

### T21-D06（Lead 批准）

AC-029 完成审计证明 `compatibility-zero` 只扫描部分治理/文档目录，且没有检测 JWT、refresh token、AutoMigrate 与旧 ORM；因此新增产品源码可以恢复已删除架构而 Gate 仍为绿色。Lead 在 T-21 已有 `<Path>scripts/quality/**</Path>` 所有权内把扫描扩展到全仓文本源码与锁文件，只剪枝 Speculo 历史、Git 数据、依赖、构建/测试产物和二进制；现有许可按“精确文件 + 禁止项类别”限定为许可证说明、安全拒绝、负向测试和发行策略断言，并增加 Go/Rust 失败探针。

### T21-D16（Lead 批准）

最终冷启动审计证明当前文档只声明语言和包管理器，没有声明根命令唯一依赖 Go Task，也没有区分 Rust 最低版本与 CI 验证版本；任意 Task、Node 或 Rust 基线漂移都可能令本地成功无法由 CI 复现。Lead 基于用户已批准的完整零兼容重构，把根 Task 精确固定为 `3.48.0`，记录 Go/Node/pnpm/Rust 的 manifest minimum 与 CI baseline，并由 architecture gate 同时约束文档和 CI；不引入旧版本兼容矩阵或额外工具管理器配置。

### T21-D17（Lead 批准）

真实原始 Volta 环境执行完整 `task test` 时，后端 Generator 的 Node 子进程因 PATH 中没有 pnpm 以 `spawnSync pnpm ENOENT` 失败，证明 D15 只闭合了前端 shell 命令面；同时从仓库根直接执行未带版本的 Corepack 会选择 `10.34.5`，不能满足 Workspace 固定的 `11.1.3`。Lead 精确开放 `<Path>scripts/contracts/generate.sh</Path>` 与 resolver 回归文件：前后端脚本共用唯一 `pnpm@11.1.3` resolver，后端 dev/test 和合同 wrapper 在进入 Generator 前准备可继承的标准 shim，后端 CI 安装 Node、pnpm 与 frozen frontend workspace；所有失败和退出码继续向根 Task 传播，不保留漂移版本 fallback。

### T21-D18（Lead 批准）

最终 AC-011/AC-022 非 E2E 审计证明 Web `/api/runtime/identity` 丢失 IAM manifest 的 `dataScope`，共享 `RuntimeIdentity` 没有该字段，而 Shell 会把缺失值默认为 `all`；因此 `self` 账号在 Web 中会错误显示只允许全部数据范围的控制项。真实完整 Go 并发测试又两次证明 Generator 直接启动 Volta Node shim 会以 `Resource temporarily unavailable (os error 35)` 失败，而同一生成器隔离运行 339 秒通过。Lead 基于用户已批准的零兼容完整重构和先完成编码/构建要求，精确开放 `<Path>go-admin-plus/internal/app/product/runtime.go</Path>`、`<Path>go-admin-plus/internal/app/product/runtime_test.go</Path>`、`<Path>go-admin-plus/internal/modules/generator/compile_gate.go</Path>`、`<Path>go-admin-plus/internal/modules/generator/generator_test.go</Path>`、`<Path>go-admin-plus-ui/packages/platform/src/index.ts</Path>`、`<Path>go-admin-plus-ui/packages/adapters/browser/src/index.ts</Path>`、`<Path>go-admin-plus-ui/packages/adapters/browser/src/index.spec.ts</Path>`、`<Path>go-admin-plus-ui/packages/adapters/desktop/src/index.ts</Path>`、`<Path>go-admin-plus-ui/packages/app-shell/src/product/ProductWorkspace.vue</Path>`、`<Path>go-admin-plus-ui/tests/shell/app-shell.spec.ts</Path>`、`<Path>go-admin-plus-ui/tests/shell/web-runtime.spec.ts</Path>` 与 `<Path>go-admin-plus-ui/tests/shell/workspace-boundary.test.mjs</Path>`：Web 与 Desktop 统一使用必填 `self|all` 数据范围且不提供兼容默认值；Generator 在 Volta 下解析已安装的真实 Node runtime，并让临时 Workspace 的 PATH 继承该实体目录。公共模块 OpenAPI、数据库 schema、UI/CSS、已删能力和受保护发行行为保持不变。

### T21-D19（Lead 批准）

最终 Desktop production artifact 反向审计发现，原 `<Path>go-admin-plus-ui/apps/admin-desktop/src/App.vue</Path>` 以环境分支包裹原生测试控件；Vite 虽能摇除 JavaScript 分支，却仍把 scoped `.native-e2e` CSS 写入 production bundle。既有 native restore 只搜索 `E2E scope self` 一项文本，无法证明 production WebView 资源不存在测试路由或其他控件。Lead 精确开放 Desktop App/入口/Vite/package 与原生 runner 静态测试路径：production App 只组合 `ProductWorkspace`；所有测试 UI 移入显式 `native-e2e` mode 的独立入口；默认 production build 完成后逐字节扫描全部产物，原生 runner 恢复 production 制品时复用同一 verifier。产品 UI/CSS、运行时行为、数据库、公共合同和受保护发行保持不变。

### T21-D20（Lead 批准）

对 D19 门禁做反向完备性审计后确认，WebView verifier 只列出五项选定 marker，漏掉其余边界/授权文案、fixture、旧环境开关和 native feature 身份；sidecar/host restore 仍另行只搜索 `/__desktop/test-control`。Lead 把 `<Path>go-admin-plus-ui/apps/admin-desktop/scripts/verify-production.mjs</Path>` 定为唯一 production marker 合同：目录资源与普通二进制都使用同一不可变集合，覆盖 route、mode/feature/env、全部 `E2E ` 文案和 `E2E-` fixture；原生恢复删除自定义单路由扫描并复用普通文件 verifier。测试逐项注入全部 marker，且静态拒绝 runner 恢复两份实现。产品代码、UI/CSS、数据库与公共合同不变。

### T21-D21（Lead 批准）

D20 的二进制扫描只由暂停的 native runner 在恢复 production 制品时调用，常规 Desktop build/package、CI 和受保护 macOS/Windows 发行仍可在不扫描最终 sidecar/host 的情况下成功；同时四字节 `E2E ` / `E2E-` 通配 marker 对已签名或压缩二进制存在随机碰撞风险。Lead 新增唯一跨平台 build verifier，复用现有 target-triple/sidecar 命名合同解析精确制品路径，并在本地 build/package、Desktop CI 及两平台受保护发行链路自动执行；marker 改为覆盖全部当前控件的精确长字符串，负向 architecture/release policy 和逐 marker 功能测试阻止接线或覆盖回退。产品代码、UI/CSS、数据库、公共合同、签名授权与暂停的 E2E 均不变。

### T21-D22（Lead 批准）

最终 Spec-to-source adapter 审计证明 `<Path>go-admin-plus-ui/packages/platform/src/index.ts</Path>` 的 `PlatformPort` 只有声明且全仓零实现/零消费，Browser/Desktop Runtime Adapter 均未提供宿主能力；Desktop Files 仍直接使用浏览器 DOM 文件输入和下载链接。这与 CONTEXT 的双 adapter 实现合同及 Spec 的 Desktop 文件选择/下载、剪贴板、通知要求冲突。Lead 精确开放现有 platform、双 adapter、双 App 产品接线、共享 Files 页、Tauri host/capability/dependency、对应 package/lock/test 与 architecture gate 路径：以一个显式 Port 注入双 App，Web 保持现有文件输入外观，Desktop 文件选择/保存只通过有界自定义 Rust command 且不向 WebView 暴露路径或通用文件系统权限；剪贴板仅写、通知仅发送。业务 API、数据库、已删能力、可见产品样式、受保护发行和暂停 E2E 不变。

实现/result 为 `aa502221356f7bbd9c3d7619ea10272f4a5971cf`（tree `bb167a772d714b427f1503e6cf69a579a1097bb5`）；完整 RED/GREEN、安全边界、构建和未执行范围记录于 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` 第 23 节。T-21 继续保持 `implemented-pending-final-e2e`。

### T21-D23（Lead 批准）

D22 后继续反查 Desktop Files 二进制数据路径，确认 Files strict OpenAPI handler 对成功下载固定输出 `application/octet-stream`，而 Tauri Rust proxy 只接受 PNG/JPEG/PDF/TXT 业务 MIME，导致所有 Desktop 下载在进入 Blob/原生保存前确定性失败。Lead 只开放 Desktop proxy/adapter 及其聚焦测试与对应 Evidence：统一使用唯一 `application/octet-stream` envelope，WebView 边界再次校验 exact keys、canonical base64 和 10 MiB 上限；不新增 MIME 兼容分支，不修改 Files API/OpenAPI、数据库、页面标记、UI/CSS、受保护发行或暂停 E2E。

授权 checkpoint 为 `63bb941e904a0c14a659f1af2f4c3afbde91bcc0`；implementation/result 为 `0554d320b74c3e3be7f2c8db8c52e38af81b42b1`（tree `108d70c73529f18f9bbb116aa526aed43a331e59`）。Rust proxy 现在只把成功 Files content 响应编码为固定 `application/octet-stream` envelope，Desktop fetch bridge 只在该路由解码 exact-key、canonical、最大 10 MiB 的 base64；错误响应继续走 JSON。完整证据见 T-21 Evidence 第 24 节；T-21 继续保持 `implemented-pending-final-e2e`。

### T21-D24（Lead 批准）

最终 Desktop source graph 审计确认 `<Path>go-admin-plus-ui/apps/admin-desktop/src-tauri/src/demo_contract.rs</Path>` 在 T-17 切换到全产品 `<Path>product_contract.rs</Path>` 后仍保留 525 行旧 Demo-only validator/test，但 `main.rs` 不再声明该 module，Cargo 永远不编译它；该死源码还单独保留 `time` 与 `uuid` 两个 Desktop 直接依赖。Lead 只开放删除该文件、移除两项无用直接依赖及 lock 投影、并把精确旧路径加入 compatibility-zero 负向门禁；不修改现行 product contract、运行行为、API、数据库、UI/CSS、受保护发行或暂停 E2E。

授权 checkpoint 为 `5f9878fa5de308d36f9118630c959a4ecd820ca4`；implementation/result 为 `a1f61f66b05a6beb4f68455f03f952909dfda269`（tree `cf38d7be04ec335d6f87936592ad6ef927805558`）。死源码已删除，Desktop 一级依赖树不再包含 `time`/`uuid`，两者仍按 Tauri 传递图保留在 lock 中；compatibility-zero 会精确拒绝旧路径恢复。完整证据见 T-21 Evidence 第 25 节；T-21 状态不变。

### T21-D25（Lead 批准）

最终前端依赖图审计确认共享 `<Path>go-admin-plus-ui/packages/app-shell</Path>` 直接导入并依赖 Browser adapter 中的 Files HTTP client，因此 Desktop App 经共享产品组合间接消费 Browser runtime adapter；这违反双 App 各自选择 adapter、共享 Shell 只组合 Web Domain 的既定边界。Lead 精确开放 Browser adapter、Files Web Domain、app-shell 产品接线、对应 package/lock、根 typecheck 与 workspace boundary test：把现有 fetch-compatible Files client 及聚焦测试按原行为迁入 Files Web Domain，删除 Shell→Browser adapter 和 Browser adapter→Files Domain 两条反向边，并把 Browser adapter 统一到标准 package test/typecheck。API、请求语义、数据库、页面、UI/CSS、受保护发行和暂停 E2E 不变。

授权 checkpoint 为 `763f7c78f278737f4d6350c0ba96aba7a5e00579`；implementation/result 为 `d236c21368f0126bc6ec3a36ef03aec3e619969e`（tree `9b61f312f159574e4553d83af694274d38c312e7`）。Files client 与测试已迁入 Files Web Domain，Shell 和 Browser adapter 的两条反向依赖均删除，静态 Files browser driver 同步消费新公共导出；完整证据见 T-21 Evidence 第 26 节，T-21 状态不变。

### T21-D26（Lead 批准）

D25 后继续审计 Desktop adapter graph，确认 `createDesktopDemoClient` 与其 Demo status mapper 只有公开导出和自测、全仓零消费者；生产和 native-e2e 产品组合都已统一以 `createDesktopFetch` 注入 Files/Demo 等 Web Domain clients，该死实现单独保留 Desktop adapter→Demo Domain 依赖。Lead 只开放 Desktop adapter 源码/测试、manifest/lock 与 workspace boundary：删除未消费 Demo client/test 和依赖边，并新增 Desktop adapter workspace dependency allowlist。通用 Desktop transport、Session/Stronghold、native-e2e control、产品/API/数据库/UI/CSS、受保护发行和暂停 E2E 均不变。

授权 checkpoint 为 `8abc93446b6aefebce2f74be54a811b42ac56238`；implementation/result 为 `124dd32344f000fab655783782d6e54c5f4fbecb`（tree `3f3cc2a823966f9cf2402e87e6d2c04f91e8132b`）。未消费 Demo client/status mapper 及其自测已删除，Desktop adapter manifest/lock 不再依赖 Demo Domain；通用 transport 自测与 native-e2e control 保留。完整证据见 T-21 Evidence 第 27 节，T-21 状态不变。

### T21-D27（Lead 批准）

最终 25-workspace 验证集合审计确认 `adapter-desktop`、`platform` 与 `ui` 含生产 TypeScript 却没有 package-local typecheck，`adapter-desktop` 还拥有 package spec 但没有 test script；根递归检查因此只执行“已经声明”的 package，依赖 App/Web Domain 间接导入不能证明 package 可独立验证。Lead 只开放三个 package manifest/tsconfig、精确 lock projection 与 architecture gate/tests：动态要求每个非根 package 声明 typecheck、拥有 spec 的 package 声明 test，并补齐三个 strict package contract；没有 spec 的双 App/platform/ui 不增加空 test。产品源码、API、数据库、页面与 UI/CSS、受保护发行和暂停 E2E 不变。

授权 checkpoint 为 `c16a2e3b90cde3b21ad592763801c90abb0d9048`；implementation/result 为 `b20e56c771d08d47da25c68dde251d1a38c07ae5`（tree `c1ae58477ebc0065f8c5577a55f536e3eb8d13a8`）。三个 package 已拥有 strict typecheck，Desktop adapter 同时拥有标准 test；architecture gate 动态拒绝缺失 contract。完整证据见 T-21 Evidence 第 28 节，T-21 状态不变。

### T21-D28（Lead 批准）

最终按 ADR-006/ADR-012 从 Go production import graph 反向审计，确认现有 architecture gate 只验证目录形态，没有实现决策要求的 Go import graph 与禁止依赖 fixture；Audit、Demo、Files、Generator、Organization、Scheduler、Settings 的生产包仍直接导入 IAM authorization/session/administration。部分消费者虽已声明最小 port，但具体 `IAM*Adapter`、IAM 返回类型和默认 `NewService` 构造仍位于消费者模块，`internal/app/product` 只调用这些构造函数，实际依赖方向没有反转；Generator 模板还会为新模块继续生成同类跨模块导入。

Lead 以 `main@f76649efa829a209b91bb9729ce5c90d4dec0371` 为审计 base，基于用户已批准的零兼容完整重构精确开放 `<Path>go-admin-plus/internal/application/architecture_test.go</Path>`、`<Path>go-admin-plus/internal/contracts/capabilities/**</Path>`、`<Path>go-admin-plus/internal/app/product/**</Path>`、`<Path>go-admin-plus/internal/modules/{audit,demo,files,generator,organization,scheduler,settings}/**</Path>`、`<Path>go-admin-plus/internal/modules/iam/authorization/**</Path>` 与对应 SpecDev Evidence：用 Go AST 对非测试生产文件建立模块边界门禁和失败 fixture；只把被多个模块真实共享的 capability 定义值移入 contracts；由各消费者保留最小同步 port，并由 product composition root 实现 Session、Authorization、Organization Projection 与 Login Fact 映射；同步 Generator 模板，删除所有消费者模块内旧 `NewIAM*` adapter 和隐式 IAM 默认构造，不提供兼容转发。IAM 内部子包协作保持 IAM 模块内部实现细节；HTTP/API/OpenAPI、数据库 schema/migration、业务行为、可见 UI/CSS、受保护发行和暂停的 E2E 不变。

#### T21-D28-A01（Lead 批准）

构造函数收缩后的全仓调用点审计证明 `<Path>go-admin-plus/test/{files,organization,settings,scheduler}/**</Path>` 等受跟踪双数据库/浏览器 harness 也是替代组合根，仍需编译并装配同一 IAM provider；若 adapter 只作为 product 私有类型，这些非 E2E 测试无法编译，在各 harness 复制映射又会形成多份安全边界。Lead 精确追加开放 `<Path>go-admin-plus/internal/app/adapters/**</Path>` 与受影响的 `<Path>go-admin-plus/test/**</Path>` 调用点：把 IAM 具体映射集中为 app-owned、按消费者接口返回的类型化 adapter 集合，正式 product 与测试组合根共同复用；`internal/app/product` 仍是正式产品唯一组合点，测试目录不得成为生产依赖。该修正不恢复模块内构造、不新增共享 service locator，也不运行被用户暂停的 E2E。

D28 授权 checkpoint 为 `cb3a428d3805723d58f3fe40022d6aaa78fbf516`，D28-A01 授权 checkpoint 为 `fe4345ff00b6065426b7473cc85eb962c3dc057c`；implementation/result 为 `c648a32ad5e20efe2059fe7993af3fa751ce5887`（tree `bb8320f5179ea9c4a04d7762986ac06f961cffa9`）。生产跨顶层模块 import edge、旧 `NewIAM*` adapter 和隐式 IAM 默认构造均已归零；普通/SQLite 全量 Go、race、vet、Generator 真实隔离编译、根 architecture/compatibility-zero 与 Server/Web/Tauri 2 全目标 production build 通过。完整证据见 T-21 Evidence 第 29 节；T-21 状态不变。

### T21-D29（Lead 批准）

D28 后继续按 ADR-012/ADR-016 审计数据库边界，确认当前八模块生产 SQL 未发现跨模块表访问，但现有 architecture gate、quality scripts 与测试都没有从模块 migration 推导表所有权，也无法拒绝模块新增对其他模块表的 SQL 字面量。一次人工零命中不能证明“模块不能访问其他模块 repository/table”长期成立。Lead 以 `main@5e76dee3a8c1f600d0b32fc865ed11a896b61b06` 为审计 base，精确开放 `<Path>go-admin-plus/internal/application/architecture_test.go</Path>` 与对应 SpecDev 状态/Evidence：从双方言 module migration 提取 `CREATE TABLE` 所有权，扫描非测试生产 Go 字符串并拒绝跨顶层模块表名，增加 own-table/foreign-table/`_test.go` 负向 fixture；重复方言声明必须归属同一模块，动态 Generator metadata 与 platform-owned reliable runtime 不改变。产品实现、数据库 schema/migration、API/OpenAPI、前端、页面 template/style、UI/CSS、发行和暂停的 E2E 均不修改。

#### T21-D29-A01（Lead 批准）

表 owner gate GREEN 后反查 T-15 “生成物通过架构边界检查”接缝，确认 `<Path>go-admin-plus/internal/modules/generator/compile_gate.go</Path>` 只运行新模块自身的 Go test/build 与前端/合同检查，不执行包含 import/table owner 规则的 `<Path>go-admin-plus/internal/application</Path>` 测试；因此 Generator 仍可能在架构门禁未运行时报告写入成功。Lead 以 `main@cba2cec93eaab9a4ae5e7c8bd1c827283d4c784f` 为审计 base，追加开放 Generator compile gate 与既有 test：把固定命令清单提取为可测试函数，要求生成后先运行 application architecture package，再运行目标模块 test/build；任何失败继续拒绝写入且不产生部分成功。生成模板/内容、自动产品注册 OUT 边界、API、schema/migration、前端与 UI/CSS 不变，不运行暂停 E2E。

D29 授权 checkpoint 为 `cba2cec93eaab9a4ae5e7c8bd1c827283d4c784f`，D29-A01 授权 checkpoint 为 `15c58d8cc9293190bae819a6d4f13b22f832b257`；implementation/result 为 `e6c300c1fdd891602ddf2b4303596b1917989e4e`（tree `f9c8d3cab05d0a440db1a47a7f47643408e7d254`）。双方言 migration owner、静态 SQL table-position 边界和 Generator application architecture 接线已建立；普通/SQLite 全量 Go、race、vet、真实 Generator 隔离编译、根 architecture/compatibility-zero 与 Server/Web/Tauri 2 全目标 production build 通过。完整证据见 T-21 Evidence 第 30 节；T-21 状态不变。

### T21-D30（Lead 批准）

D29 后按 AC-026 与 ADR-006 逐层核对正式 runtime 路由和 Go import graph，确认 Server 实际只接入 `<Code>/health/live</Code>`、`<Code>/health/ready</Code>` 与 `<Code>/api/v1/runtime/capabilities</Code>`；规范要求的 `<Code>/metrics</Code>` 和 `<Code>/api/v1/runtime/status</Code>` 只存在于零生产引用的 `<Path>internal/platform/observability</Path>` 重复实现。该死包还直接导入 `<Path>internal/app/kernel</Path>`，形成 platform -> app 反向依赖；`app/kernel` 本身拥有 Desktop host resource lifecycle，也不属于 app “只装配”职责。Desktop sidecar 仅暴露专用 readiness nonce，没有把 product readiness checker 接入规范运维合同。

Lead 以 `main@638e2a0a43cbc4fc8994ddbce9b6d37eaffe83c4` 为审计 base，精确开放 `<Path>go-admin-plus/internal/application/{health,architecture_test.go}</Path>`、`<Path>go-admin-plus/internal/host/{server,lifecycle}</Path>`、删除 `<Path>go-admin-plus/internal/{app/kernel,platform/observability}</Path>`、受影响 `<Path>go-admin-plus/cmd/{go-admin-plus,desktop-sidecar}</Path>` 调用点及对应 SpecDev 状态/Evidence：把五个运维端点收敛到唯一 application health handler，以单次 application snapshot 和有界 readiness checker 计算 redacted starting/ready/dependency-failed/draining 状态与 metrics；profile/database capabilities 显式区分三正式 profile；Server 与 Desktop 共用同一 handler，Desktop 仍要求 loopback control token。Host resource lifecycle 机械迁至 host/lifecycle，不保留旧包转发；Go AST 负向 fixture 拒绝 contracts/application/platform 的反向层依赖并拒绝 app 下恢复非 composition package。产品模块、业务 OpenAPI、schema/migration、前端页面 template/style、UI/CSS、发行和暂停 E2E 不变。

D30 授权 checkpoint 为 `de98a1f58f532812d0f4bc997eb16d04ccf9d15f`；implementation/result 为 `6c01fd983e2f26b4fb01c71be8ffcd7a11fdcb81`（tree `364840919749cb3ab736a7e18c05b6c63486ae7f`）。唯一 application health handler 已覆盖五端点和三 profile，Server/Desktop 复用同一 readiness/status 合同；旧 platform observability 删除，host lifecycle 离开 app，低层反向依赖与旧路径恢复均由 Go AST gate 拒绝。普通/SQLite 全量 Go、race、vet、真实 Generator、根静态门禁和 Server/Web/Tauri 2 全目标 production build 通过。完整证据见 T-21 Evidence 第 31 节；T-21 状态不变。

### T21-D31（Lead 批准）

D30 后继续按 ADR-006 审计正式命令入口，确认 `<Path>go-admin-plus/cmd/desktop-sidecar/runtime.go</Path>` 仍直接拥有实例锁、SQLite 备份/恢复、listener/HTTP server、product Application 生命周期与 shutdown 编排；带 `desktop_native_e2e` 标签的生产源文件也位于 command package，并让无标签 runtime 字段直接暴露 IAM Session 具体类型。`cmd` 因而不是“仅入口”，且现有 backend layer gate 不扫描 command-owned runtime，无法阻止该结构恢复。

Lead 以 `main@b70a406b6ce19ad377d779d6e91109e6eec24031` 为审计 base，基于用户已批准的零兼容完整重构，精确开放 `<Path>go-admin-plus/internal/host/desktop/**</Path>`、`<Path>go-admin-plus/internal/app/product/desktop*.go</Path>`、`<Path>go-admin-plus/internal/application/architecture_test.go</Path>`、`<Path>go-admin-plus/cmd/desktop-sidecar/**</Path>` 及对应 SpecDev 状态/Evidence：由 host/desktop 拥有 Desktop 资源与 HTTP 生命周期，通过 application-facing Builder 端口接收产品运行时；app/product 作为高层 composition 实现 Builder，并把 native 测试控制隔离到显式 build tag；command 只保留 launch material、signal/parent-pipe、listening status 与 Host 调用。删除旧 command runtime/test-control 文件，不保留转发或别名；新增负向 fixture 拒绝 Desktop command 恢复非入口生产文件。API、schema/migration、产品行为、页面 template/style、UI/CSS、发行链路和暂停的 E2E 均不变。

### 已采用的低影响假设

- 零兼容扫描使用明确 allowlist，仅允许 SpecDev 历史工件和必要否定性测试文本。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 旧路径删除、最终 CI、质量扫描、当前文档、产品命名切换 | T-01 至 T-20 已验证目标产品 | 旧数据迁移、远端发布、历史工件改写 |

## 4. 要构建什么

Lead 先冻结已通过的产品/发行证据，再删除旧目录和所有兼容调用点，更新根 CI 与当前文档；全量合同、架构、三 profile、双 App、三平台和零兼容扫描全部通过后才接受原子切换结果。

## 5. 实现契约

- **入口或接缝：** contraction inventory、zero-compat scanner、root CI policy、全量 root tasks、文档链接检查。
- **输入与输出：** 已验证 result SHA 与 allowlist；输出唯一目标仓库、完整 Gate 和删除清单 Evidence。
- **公共接口变化：** 移除全部旧外部/内部合同，只保留 Spec 定义的新产品。
- **不变量：** 不删新结构/用户无关改动/SpecDev 历史；任何 required Gate 失败阻断；无过期豁免。
- **状态或数据流：** freeze evidence -> inventory -> delete/rename -> docs/CI -> zero scan -> full matrix -> result SHA。
- **错误与失败行为：** 未知引用、扫描命中、平台 Gate 缺失或文档漂移立即停止，不形成部分切换结果。
- **兼容要求：** 明确为零兼容；无 shim、双写、legacy alias 或临时 migration 标记。
- **安全与隐私要求：** 删除前确认目标；CI/文档/扫描输出不含 secret；保留供应链证据。

## 6. 执行路线

1. 锁定 T-17 至 T-20 result SHA、制品证据和精确旧路径 inventory。
2. 先运行预期会命中的零兼容扫描并确认 allowlist。
3. 原子删除旧前端、旧后端层、旧宿主/配置/数据库/治理资产并完成命名切换。
4. 更新根 CI、质量脚本和当前产品文档，禁止过期豁免。
5. 运行全量 root tasks、三 profile、双 App、三平台证据检查和最终零扫描。

## 7. 路径访问契约

- **预计修改点：** frontmatter 列出的最终 CI/文档与精确旧路径。
- **可写范围：** 仅 `writable_paths`；删除前逐项解析，不使用宽泛工作区删除命令。
- **只读上下文：** 新产品、发行和根任务全部只读验证。
- **共享路径：** 根 CI、quality scripts、README/docs 由 T-21 唯一拥有。
- **保留或不动：** `<Path>speculo/**</Path>`、`.git`、新产品模块和用户无关改动。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | final product gate | `task test lint generate migrate package release:verify` | 根任务、产品和发行证据全部通过 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` |
| 失败路径 | contraction/policy fixture | 保留一项旧引用、过期豁免或缺平台证据 | CI/扫描阻断且不接受 result | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` |
| 回归 | zero scan/full matrix | `task governance:check architecture:check compatibility:zero` | allowlist 外旧模式零命中，三 profile/双 App 仍通过 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>` |

- **Workspace checks：** Goal Plan 在 current-workspace 或 source-worktree 运行非 E2E 全量门禁；删除前后都记录状态。
- **E2E disposition：** deferred/final：T-21 candidate 先完成删除、全量非 E2E Gate 与零兼容扫描；其实现 result 进入 `main` 后，Lead 创建唯一最终系统候选并统一执行三 profile、双 App、原生发行与全模块 E2E。
- **E2E owner/environment：** Lead / 最终系统候选；source-worktree 与 T-21 实现 candidate 不运行或声明 E2E 通过。
- **Integration evidence：** implementation/source commit、parent before、candidate/result SHA、删除清单、零兼容/全量非 E2E Gate、最终统一 E2E 与父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** expand 已由 T-01 至 T-20 完成；本 Ticket 仅在所有 Gate 后执行 contract。
- **兼容窗口：** 无产品兼容窗口，切换结果不可包含旧新双轨。
- **监控信号：** zero-scan hits、root Gate、profile/App/release result 和豁免到期。
- **回滚或前向恢复：** 删除 commit 在发布前可整体回滚；发布后只允许基于新架构前向修复，不恢复旧体系。
- **不可逆操作与批准点：** 用户本次批准覆盖规划发布，不自动授权实现删除；I-implement 仍须取得 implementation/integration 授权并在删除前确认精确 inventory。
- **收缩条件：** allowlist 外旧目录/名称/API/schema/config/Wails/JWT/refresh/Casbin/Redis/tenant/MySQL/SQL Server/临时标记全部零命中。

## 10. 验收标准

- [ ] `AC-001/AC-002`：唯一根治理、目标目录和根任务合同成立。
- [ ] `AC-023/AC-028`：合同生成、边界和 clean-tree 检查通过。
- [ ] `AC-029`：明确 allowlist 外全部旧模式零命中。
- [ ] `AC-033`：Local/PR/Protected Release 分层门禁失败均阻断且豁免可审计。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-21.md</Path>`。
- [ ] 修改未越界，形成非空 commit 并记录 final result SHA；Ticket、Map、Evidence 一致。
