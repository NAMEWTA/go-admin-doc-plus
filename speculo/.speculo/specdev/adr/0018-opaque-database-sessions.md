# ADR-0018: 不透明数据库 Session 取代 JWT 与 Casbin

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-018

## 上下文

浏览器/WebView 可读 bearer token、JWT 双 token 与 URL-based Casbin policy 会形成多份身份和授权事实源。

## 决策

IAM 生成高熵 opaque token，数据库只保存 hash，并管理撤销、空闲/绝对超时和轮换。Web 使用 `__Host-*` Secure/HttpOnly/SameSite Cookie 与 CSRF；Desktop 由 Tauri Stronghold proxy 注入凭据。密码使用 Argon2id。RBAC、数据范围、API 和前端 capability 共用稳定 dot-separated permission code。

## 后果

Session 和权限变更可即时生效。日志、审计、错误、URL、Web Storage 和 WebView 状态不得出现原始 token；JWT、refresh token 和 Casbin 不得恢复。
