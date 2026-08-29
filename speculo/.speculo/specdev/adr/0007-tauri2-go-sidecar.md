# ADR-0007: Tauri 2 管理 Go Sidecar

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-007

## 上下文

桌面端需要复用 Go 产品能力，同时以显式原生权限、凭据和进程生命周期替代 Wails binding。

## 决策

Tauri 2 是唯一桌面宿主。Go 应用作为按 target triple 打包的 sidecar，由 Rust host 管理实例锁、随机 loopback、一次性启动控制、健康/退出、迁移恢复和清理。Session 保存在 Stronghold，由受控 transport proxy 注入；WebView JavaScript 不读取凭据或原生路径。

## 后果

桌面构建必须先准备匹配 sidecar，再构建 `custom-protocol` host。Tauri capability/command 使用精确 allowlist；production 资产必须拒绝 native E2E control marker。
