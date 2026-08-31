---
schema_version: 3
artifact: ticket
change: 2026-08-29-productization-and-ui-reconstruction
id: T-01
title: 修复已知依赖漏洞并建立安全基线
status: in_progress
planning_depth: standard
planning_depth_reason: 请求校验链路存在可达漏洞，虽修改集中于 Go 依赖图，但需要兼容、生成和回归证明
ready: true
risk: high
blocked_by: []
contract_ids: [AC-037]
owner: codex-root
expected_changes: ["<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>"]
writable_paths: ["<Path>go-admin-plus/go.mod</Path>", "<Path>go-admin-plus/go.sum</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>scripts/contracts/**</Path>", "<Path>.github/workflows/ci.yml</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-01: 修复已知依赖漏洞并建立安全基线

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/ticket/01-dependency-security-baseline.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-01.md</Path>`

## 1. 战略与来源

- **目标：** 消除请求校验链路的 GO-2026-6112 可达漏洞，给后续公共合同变更建立可信依赖基线。
- **可观察产出：** govulncheck 不再报告该可达漏洞，OpenAPI 请求校验、生成漂移和后端测试保持通过。
- **来源：** `AC-037`、`PLAN/P0`、`RESEARCH:<Url>https://pkg.go.dev/vuln/GO-2026-6112</Url>`。
- **当前事实：** `<Path>go-admin-plus/go.mod</Path>` 使用 `kin-openapi v0.142.0`，修复下限为 v0.144.0。
- **Planning Depth 原因：** 依赖升级局部，但位于所有 HTTP 请求合同的安全关键路径。

## 2. 决策状态

### 已锁定决策

- 升级到不低于 v0.144.0 的当前兼容版本，不通过禁用校验规避漏洞。
- 本 Ticket 不改 OpenAPI 业务合同；持续 CI 扫描由 T-18 接管。

### 已采用的低影响假设

- 优先选择满足修复条件的最小兼容版本；若传递依赖解析需要更高 patch/minor，记录解析理由并保持 API 兼容。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| Go 依赖升级、漏洞和合同回归 | 现有 nethttp middleware 与生成器 | CI job、业务 API、handler 重构 |

## 4. 要构建什么

维护者更新并整理 Go 依赖图后，现有请求校验仍按原合同工作；安全扫描无法再从产品调用链到达 GO-2026-6112。升级导致的不兼容必须在本 Ticket 内适配，不能通过删除验证、吞错或跳过测试制造绿色。

## 5. 实现契约

- **入口或接缝：** Go module graph、OpenAPI middleware、govulncheck。
- **输入与输出：** 锁定依赖输入；输出可复现的 `go.mod/go.sum`。
- **公共接口变化：** 无。
- **不变量：** OpenAPI 校验继续启用；生成物无漂移。
- **状态或数据流：** 不适用：无持久状态。
- **错误与失败行为：** 依赖不兼容或仍可达时 Ticket 失败。
- **兼容要求：** 仅保持当前代码合同，不保留漏洞版本兼容。
- **安全与隐私要求：** 扫描输出不得包含配置 secret。

## 6. 执行路线

1. 固定升级前漏洞与请求校验测试基线。
2. 更新 kin-openapi 及必要传递依赖并整理 module graph。
3. 运行合同、生成漂移、Go 测试和 govulncheck。
4. 核对 diff 仅包含依赖图及必要兼容调整。

## 7. 路径访问契约

- **预计修改点/可写范围：** 仅 `<Path>go-admin-plus/go.mod</Path>`、`<Path>go-admin-plus/go.sum</Path>`。
- **只读上下文：** `<Path>go-admin-plus/internal/contracts/**</Path>`、`<Path>scripts/contracts/**</Path>`、`<Path>.github/workflows/ci.yml</Path>`。
- **共享路径：** 无。
- **保留或不动：** 所有业务合同与用户未提交文件。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | 依赖与后端 | `cd go-admin-plus && go test ./... -count=1` | 全部通过 | `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-01.md</Path>` |
| 失败路径 | 可达漏洞 | `cd go-admin-plus && govulncheck ./...` | 不报告 GO-2026-6112 可达 | 同上 |
| 回归 | 合同生成 | `task contract:lint && task generate:check` | 合同与生成物 clean | 同上 |

- **Workspace checks：** Goal Plan 选择的 current-workspace 或 source-worktree 执行非 E2E 检查。
- **E2E disposition：** not-required：依赖升级的外部风险由请求合同和漏洞扫描直接证明；真实系统 E2E 由 T-19/T-20。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；本 Ticket 不运行 E2E。
- **Integration evidence：** implementation/source commit、direct-parent 或 candidate 结果、父分支 result SHA 与 Lead Evidence。

## 9. 发布、迁移与恢复

- **迁移顺序/兼容窗口：** 不适用：无数据与 wire 迁移。
- **监控信号：** govulncheck 与请求校验回归。
- **回滚或前向恢复：** 优先修复不兼容；回退仅可回到另一个无漏洞版本。
- **不可逆操作与批准点：** 无。
- **收缩条件：** module graph 不再解析到受影响版本。

## 10. 验收标准

- [ ] `AC-037`：GO-2026-6112 不再可达且修复版本下限满足。
- [ ] Go、合同和生成漂移验证通过并写入 `<Path>{roots.state}/specdev/changes/2026-08-29-productization-and-ui-reconstruction/evidence/T-01.md</Path>`。
- [ ] 修改未超出 `writable_paths`，形成非空实现 commit，direct-parent/candidate 验证和父分支 result 已记录。
- [ ] E2E disposition、双轴审查和未批准偏差检查完成。
