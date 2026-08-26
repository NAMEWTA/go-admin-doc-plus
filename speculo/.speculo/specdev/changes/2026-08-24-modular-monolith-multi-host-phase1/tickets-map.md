---
schema_version: 3
artifact: tickets-map
change: 2026-08-24-modular-monolith-multi-host-phase1
status: completed
---

# Tickets Map: Phase 1 模块化单体与多 Host 管理端

- **Map：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/tickets-map.md</Path>`
- **Spec：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/spec.md</Path>`
- **Ticket 目录：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/</Path>`
- **Evidence 目录：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/evidence/</Path>`
- **可选 Goal Plan：** `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/goal-plan.md</Path>`

## 1. 目标与拆分策略

本 Map 覆盖 `US-001` 至 `US-009` 和 `AC-001` 至 `AC-018`。T-01 先冻结行为基线；T-02/T-03 以 prefactor 形成 Application 与基础设施接缝；随后 Server、Frontend 和 Desktop 形成可运行曳光弹；平台发布在共同桌面能力之上分流；T-13 只在所有消费者迁移后收缩兼容层并执行整体验收。公共接口、路由键和生成契约采用 expand-migrate-contract，避免宽重构期间出现红色中间态。

## 2. 执行清单

| ID | Ticket | 可观察产出 | Blocked By | Depth | Risk | Ready | Owner | Contract IDs | Wave/Gate | Status |
|---|---|---|---|---|---|---|---|---|---|---|
| T-01 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/01-freeze-behavior-baseline.md</Path>` | 可重复的 API、UI、数据与启动基线 | — | standard | medium | yes | root | AC-002, AC-003 | W0/G0 | done |
| T-02 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/02-application-lifecycle-kernel.md</Path>` | Host 无关 Application 可显式装配和关闭 | T-01 | deep | high | yes | root | AC-001, AC-002, AC-011 | W1/G1 | done |
| T-03 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/03-platform-adapters-and-migrations.md</Path>` | Server/Desktop 基础设施 profile 通过同一合同 | T-02 | deep | high | yes | root | AC-010, AC-011, AC-015, AC-017 | W2/G2 | done |
| T-04 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/04-server-host-and-health.md</Path>` | 新 ServerHost 保持 API 并提供健康语义 | T-02, T-03 | deep | high | yes | root | AC-001, AC-002, AC-013 | W3/G3 | done |
| T-05 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/05-admin-workspace-shell.md</Path>` | Admin 在 workspace 中等价构建和导航 | T-01 | deep | high | yes | root | AC-003, AC-018 | W1/G1 | done |
| T-06 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/06-openapi-runtime-api-client.md</Path>` | 同一生成契约和纯传输层服务双 Runtime | T-04, T-05 | deep | high | yes | root | AC-004, AC-005, AC-018 | W4/G4 | done |
| T-07 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/07-app-core-domains-and-routing.md</Path>` | Admin 由 Domain 组合并兼容 routeKey 菜单 | T-05, T-06 | deep | high | yes | root | AC-002, AC-003, AC-006, AC-018 | W5/G5 | done |
| T-08 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/08-desktop-host-tracer.md</Path>` | Wails 桌面壳离线完成登录和代表性 CRUD | T-03, T-06 | deep | high | yes | root | AC-004, AC-007, AC-008, AC-017 | W5/G5 | done |
| T-09 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/09-desktop-durability-and-security.md</Path>` | 桌面升级、单实例、loopback 和文件安全可恢复 | T-07, T-08 | deep | critical | yes | root | AC-009, AC-010, AC-011, AC-015, AC-017 | W6/G6 | done |
| T-10 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/10-linux-compose-delivery.md</Path>` | 空卷 Compose 安全部署 Web/API | T-04, T-07 | deep | high | yes | root | AC-012, AC-013, AC-014, AC-015, AC-016 | W6/G6 | done |
| T-11 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/11-macos-arm64-release.md</Path>` | 可校验自授权的离线 ARM64 DMG | T-09 | deep | high | yes | root | AC-007, AC-010, AC-016, AC-017 | W7/G7 | done |
| T-12 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/12-windows-amd64-release.md</Path>` | 可校验且含 WebView2 的自用离线 x64 NSIS | T-09 | deep | high | yes | root | AC-008, AC-010, AC-016, AC-017 | W7/G7 | done |
| T-13 | `<Path>{roots.state}/specdev/changes/2026-08-24-modular-monolith-multi-host-phase1/ticket/13-product-release-and-contract.md</Path>` | 三类产物同版本，兼容层可证收缩 | T-10, T-11, T-12 | deep | critical | yes | root | AC-002, AC-005, AC-006, AC-016, AC-018 | W8/G8 | done |

## 3. 依赖 DAG

```text
T-01 [BASELINE]
  ├─→ T-02 [APPLICATION PREFACTOR] ─→ T-03 [ADAPTER EXPAND] ─┬─→ T-04 [SERVER]
  │                                                          └─→ T-08 [DESKTOP TRACER]
  └─→ T-05 [WORKSPACE EXPAND] ───────────────┬─→ T-06 [CONTRACT/RUNTIME]
                                              │       ├─→ T-07 [DOMAIN MIGRATE]
                                              │       └─→ T-08
                                              └───────────────┘
T-07 + T-08 ─→ T-09 [DESKTOP HARDEN]
T-04 + T-07 ─→ T-10 [LINUX DELIVERY]
T-09 ─┬─→ T-11 [MAC RELEASE]
      └─→ T-12 [WINDOWS RELEASE]
T-10 + T-11 + T-12 ─→ T-13 [CONTRACT + RELEASE GATE]
```

T-02/T-03/T-05 是明确解除后续阻塞的 prefactor/expand；T-13 是唯一 contract 收缩与产品级汇合点。

## 4. 合同覆盖矩阵

| Contract ID | 覆盖 Ticket | 验证接缝 | 状态 | 说明 |
|---|---|---|---|---|
| AC-001 | T-02, T-04 | Application/Host 集成 | covered | 双 Host 构建与停止 |
| AC-002 | T-01, T-02, T-04, T-07, T-13 | HTTP characterization/contract | covered | 全程保持兼容 |
| AC-003 | T-01, T-05, T-07 | build 与 Playwright | covered | workspace 行为等价 |
| AC-004 | T-06, T-08 | Runtime 双适配 E2E | covered | 一份业务客户端 |
| AC-005 | T-06, T-13 | 生成零 diff 与 type-check | covered | OpenAPI 权威 |
| AC-006 | T-07, T-13 | 路由单测/E2E/扫描 | covered | expand 后收缩 |
| AC-007 | T-08, T-11 | macOS 原生离线 E2E | covered | tracer 后发布验证 |
| AC-008 | T-08, T-12 | Windows 原生离线 E2E | covered | WebView2 缺失场景 |
| AC-009 | T-09 | socket 与 Origin/token 测试 | covered | loopback 安全 |
| AC-010 | T-03, T-09, T-11, T-12 | 双方言 fixture/升级恢复 | covered | 数据完整性 |
| AC-011 | T-02, T-03, T-09 | lifecycle/single-instance | covered | 无永久 worker |
| AC-012 | T-10 | 空卷 Compose E2E | covered | Web/API/依赖编排 |
| AC-013 | T-04, T-10 | live/ready/capabilities | covered | 健康语义 |
| AC-014 | T-10 | Compose config/runtime | covered | 容器最小权限 |
| AC-015 | T-03, T-09, T-10 | FileStore contract | covered | 桌面与服务端适配 |
| AC-016 | T-10, T-11, T-12, T-13 | release verification | covered | 三类产物 |
| AC-017 | T-03, T-08, T-09, T-11, T-12 | 网络隔离 E2E | covered | 核心流程无公网依赖 |
| AC-018 | T-05, T-06, T-07, T-13 | 调用扫描与完整回归 | covered | contract 条件可证 |

## 5. 并行与路径所有权

- Goal Plan 已选择 `ticket_workspace_policy: current`，因此 implementation agent 上限收紧为 `1`；13 个 Ticket 严格串行，禁止 source/candidate worktree 和并行写入。
- Lead `root` 固定拥有 SpecDev 状态、父分支推进、Evidence 和 E2E；Ticket implementation owner 在进入 Ticket 时动态派单，但同一时刻只能有一个 writer。
- 下表只记录 DAG 在 `required` 隔离策略下原本可能存在的结构并行性，不构成本 change 的并行授权。
- 根 manifests、Go module 清单、迁移索引和发布编排继续由 frontmatter 指定的单一 Ticket owner 管理。

| Ticket A | Ticket B | Writable 交集 | 真实依赖 | 处理 |
|---|---|---|---|---|
| T-02 | T-05 | 无 | 否 | current 策略下按 T-02 -> T-05 串行 |
| T-07 | T-08 | 无 | 否 | current 策略下按 T-07 -> T-08 串行 |
| T-10 | T-11 | 无 | 否 | current 策略下按 T-10 -> T-11 串行 |
| T-11 | T-12 | 无 | 否 | current 策略下按 macOS -> Windows 串行；原生 runner 仅用于各自验证 |

## 6. Gate、Wave 与集成点

Gate 为：G0 基线可信；G1 后端/前端 prefactor 保持绿色；G2 双 profile contract；G3 ServerHost 兼容；G4 OpenAPI/Runtime；G5 Web/桌面曳光弹；G6 桌面安全与 Compose；G7 两平台 self-use 发布；G8 三类产物和兼容收缩。Goal Plan 选择 current/direct-parent，实际执行顺序固定为 `T-01 -> T-02 -> T-05 -> T-03 -> T-04 -> T-06 -> T-07 -> T-08 -> T-09 -> T-10 -> T-11 -> T-12 -> T-13`；Wave 只表达能力阶段，不表达并发。

## 7. 横切契约与风险

- `/api/v1`、授权和响应 envelope 在兼容窗口内保持稳定；新增 health/capabilities 不重定义既有业务接口。
- Desktop SQLite 与 Server PostgreSQL 共享迁移语义，发布迁移只追加；失败 fail closed，Desktop 保留备份。
- Desktop loopback 同时依赖 loopback bind、随机端口、启动令牌、严格 Origin 和单实例，任何一项不得被视为其余项的替代。
- Phase 1 桌面正式产物必须明确标记 `unsigned-self-use` 并提供 checksum/SBOM/trust state；未来签名凭据只能通过原生 CI secret 注入，启用需新 release 偏差。
- 旧 AppRouters、旧 request 入口、component 物理路径和正则契约检查由 T-13 在零调用证据后收缩。

## 8. 同步规则

- Ticket 状态变化后同步本 Map；Ticket frontmatter 是状态、依赖、深度和路径契约权威。
- Goal Plan 创建后，Wave、Gate、owner、workspace strategy 和集成顺序以 Goal Plan 为权威并投影回本 Map。
- 任何公共行为、数据、安全、迁移或发布变化返回 Spec；路径或局部契约变化按 deviation control 处理。
- 内部工件只使用完整根变量 Path，项目文件只使用项目相对 Path。
