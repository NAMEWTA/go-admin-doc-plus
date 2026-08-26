---
schema_version: 3
artifact: ticket
change: 2026-08-26-project-architecture-reconstruction
id: T-06
title: 安全登录、Session 与个人账户闭环
status: ready
planning_depth: deep
planning_depth_reason: 认证、密码、Session schema、Cookie/CSRF 和审计脱敏属于关键安全边界
ready: true
risk: critical
blocked_by: [T-03, T-04, T-05]
contract_ids: [AC-008, AC-010, AC-012, AC-025, AC-036]
owner: unassigned
expected_changes: ["<Path>contracts/openapi/modules/iam-session.yaml</Path>", "<Path>go-admin-plus/internal/modules/iam/session/**</Path>", "<Path>go-admin-plus/internal/modules/iam/account/**</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/session/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/session/**</Path>"]
writable_paths: ["<Path>contracts/openapi/modules/iam-session.yaml</Path>", "<Path>go-admin-plus/internal/modules/iam/session/**</Path>", "<Path>go-admin-plus/internal/modules/iam/account/**</Path>", "<Path>go-admin-plus/internal/modules/iam/migrations/0010-session-*</Path>", "<Path>go-admin-plus-ui/packages/domains/iam/src/session/**</Path>", "<Path>go-admin-plus-ui/packages/web-domains/iam/src/session/**</Path>", "<Path>go-admin-plus/test/iam/session/**</Path>", "<Path>go-admin-plus-ui/tests/e2e/iam/session/**</Path>"]
read_only_paths: ["<Path>go-admin-plus/internal/contracts/**</Path>", "<Path>go-admin-plus/internal/platform/database/**</Path>", "<Path>go-admin-plus-ui/packages/app-shell/src/core/**</Path>", "<Path>go-admin-plus-ui/packages/api-client/**</Path>"]
shared_paths: []
shared_path_owners: []
---

# Ticket T-06: 安全登录、Session 与个人账户闭环

- **Ticket 文件：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/ticket/06-iam-session-account.md</Path>`
- **总体 Map：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/tickets-map.md</Path>`
- **上游 Spec：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/spec.md</Path>`
- **完成 Evidence：** `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-06.md</Path>`

## 1. 战略与来源

- **目标：** 交付 Web 登录、退出、Session 生命周期、个人资料和密码完整切片。
- **可观察产出：** 用户通过安全 Cookie 登录并操作账户；轮换、撤销和超时后的旧 token 永久失效。
- **来源：** `US-005`、`US-017`、`AC-008`、`AC-010`、`AC-012`、`ADR-018`。
- **当前事实：** 旧 JWT/refresh token/Casbin 不能满足即时撤销与 Desktop 安全代理目标。
- **Planning Depth 原因：** 认证失败会直接导致账户接管或会话泄露。

## 2. 决策状态

### 已锁定决策

- 使用高熵不透明 Session，只持久化 hash；支持 idle/absolute timeout、rotation 和 revoke。
- Web 使用 `__Host-*` Secure HttpOnly SameSite Cookie 与 CSRF；密码使用 Argon2id。

### 已采用的低影响假设

- 超时与轮换参数来自 T-03 类型化配置，不作为业务设置。

### 未决问题

无。

## 3. 范围边界

| IN（本 Ticket 构建） | REUSE（复用且不改变契约） | OUT（明确不做） |
|---|---|---|
| 登录/退出、Session、资料、头像元数据、密码和 Web 页面 | T-02/03/04/05 接缝 | 角色管理、Desktop Stronghold、旧 token 兼容 |

## 4. 要构建什么

用户提交凭据后服务端校验 Argon2id、创建 hash Session 并设置安全 Cookie；受保护请求同时验证 Session 与 CSRF，资料/密码变更被持久化且敏感信息不进入日志或审计。

## 5. 实现契约

- **入口或接缝：** IAM application、Session repository、OpenAPI handler、Web login/account domain。
- **输入与输出：** 凭据、Cookie/CSRF、账户命令；返回身份状态和稳定错误，不返回原始 Session。
- **公共接口变化：** 新登录/退出/current-session/profile/password API。
- **不变量：** 仅 hash 落库；旧 token 轮换/撤销/过期后不可恢复；密码不明文持久化。
- **状态或数据流：** credentials -> account verify -> Session transaction -> Cookie -> protected use case。
- **错误与失败行为：** 失败统一且不枚举账户；CSRF/过期请求无状态变化并进入重新登录状态。
- **兼容要求：** 不接受 JWT、refresh token 或旧密码 hash 迁移。
- **安全与隐私要求：** 防 fixation/timing 泄露；日志、URL、响应和审计不含 token/password。

## 6. 执行路线

1. 建立 Session 生命周期、Cookie/CSRF 和密码负向测试。
2. 增加 IAM session/account schema、repository 和 Argon2id policy。
3. 实现用例、OpenAPI mapping 和 Web 登录/账户页面。
4. 覆盖 rotation/revoke/idle/absolute timeout 与敏感值扫描。
5. 运行 API、浏览器和双方言回归。

## 7. 路径访问契约

- **预计修改点：** IAM session/account、对应合同和前端包。
- **可写范围：** 仅 frontmatter `writable_paths`。
- **只读上下文：** 公共 transport、Database 与 App Shell。
- **共享路径：** 无；只消费 T-02/T-04/T-05 接缝。
- **保留或不动：** IAM administration 归 T-07。

## 8. 验证矩阵

| 行为或风险 | 验证接缝 | 命令或步骤 | 预期结果 | Evidence |
|---|---|---|---|---|
| 正常路径 | IAM API/Web suite | `task test -- iam-session` | 登录、资料、密码、退出和持久化成立 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-06.md</Path>` |
| 失败路径 | security contract | 测试错误密码、CSRF、撤销、idle/absolute timeout | 请求拒绝、无状态变化、无敏感泄露 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-06.md</Path>` |
| 回归 | 双方言/browser E2E | PostgreSQL/SQLite 登录并轮换 | 行为等价且旧 token 永久失败 | `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-06.md</Path>` |

- **Workspace checks：** 按 Goal Plan 在 current-workspace 或 source-worktree 运行 Go/TS/unit/type/lint/build。
- **E2E disposition：** required：Cookie 属性、CSRF 和浏览器重新登录状态必须在集成环境验证。
- **E2E owner/environment：** Lead / current-workspace 或 parent-candidate；source-worktree 不运行 required E2E。
- **Integration evidence：** 记录 implementation/source commit、parent before、candidate/result SHA、E2E 和父分支包含关系。

## 9. 发布、迁移与恢复

- **迁移顺序：** 新 Session schema 与 API 同 Ticket 原子落地，随后 T-07/T-16 消费。
- **兼容窗口：** 无旧 token 或账号数据兼容。
- **监控信号：** 登录失败、revoke/rotation、CSRF 拒绝、Session 清理和敏感值扫描。
- **回滚或前向恢复：** 切换前可回滚；切换后通过前向 migration 和强制 revoke 恢复。
- **不可逆操作与批准点：** 密码策略/schema 变更在 migration Gate 审核。
- **收缩条件：** T-21 扫描 JWT、refresh、Casbin 和旧认证路由零命中。

## 10. 验收标准

- [ ] `AC-008/AC-010`：安全 Cookie+CSRF 与 Session 全生命周期合同通过。
- [ ] `AC-012`：资料/密码/退出持久化，Argon2id 和敏感值保护成立。
- [ ] `AC-025/AC-036`：认证失败类别和 Shell 重新登录状态稳定。
- [ ] 验证矩阵记录到 `<Path>{roots.state}/specdev/changes/2026-08-26-project-architecture-reconstruction/evidence/T-06.md</Path>`。
- [ ] 修改未超出 `writable_paths`，形成非空 commit 并记录 integration result SHA。
- [ ] Ticket、Map 和 Evidence 状态一致且无未批准偏差。
