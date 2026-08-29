# ADR-0008: Web 与 Desktop 双 App 共享单一业务组合

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-008

## 上下文

浏览器与 Tauri 在启动、凭据、能力和打包上合同不同，但属于同一管理产品。

## 决策

`<Path>go-admin-plus-ui/apps/admin-web/</Path>` 与 `<Path>go-admin-plus-ui/apps/admin-desktop/</Path>` 是独立可构建 App。二者消费同一 ProductWorkspace、路由、页面、领域包和 UI；App 只选择 Browser/Desktop adapter，不复制业务实现。

## 后果

宿主差异必须位于 Platform Port 和 adapter。Desktop 不得 deep import Web App，两个 App 的能力 manifest、授权语义和业务页面必须保持一致。
