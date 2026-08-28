# 仓库架构

## 后端

`go-admin-plus/` 采用组合根、应用层、模块和平台能力分离的结构：

```text
cmd/
  go-admin-plus/       Server 进程入口
  desktop-sidecar/     Tauri 管理的本地进程入口
  config-check/        配置预检入口
  migrate/             向前迁移入口
internal/
  app/product/         唯一产品组合根
  application/         跨模块用例与应用协议
  contracts/           生成的传输合同
  host/                Server 宿主
  modules/             iam、audit、organization、settings、generator、scheduler、files、demo
  platform/            config、database、migrations、coordination、outbox 等技术能力
```

模块拥有自己的领域、存储、传输、权限声明和迁移。模块之间通过端口或产品组合根协作，不直接导入其他模块的私有实现。Server 支持 SQLite 与 PostgreSQL；Desktop profile 只映射 SQLite。

Go module 固定为 `github.com/NAMEWTA/go-admin-plus/go-admin-plus`，所有内部 import 使用该仓库路径，不保留重构前的短 module path。

## 前端

`go-admin-plus-ui/` 是 pnpm workspace：

```text
apps/
  admin-web/           浏览器 App
  admin-desktop/       Tauri 2 App 与 Rust host
packages/
  app-shell/           产品装配与导航壳
  platform/            平台无关类型和端口
  api-client/          OpenAPI 生成客户端
  ui/                  共享 UI 与交互状态机
  adapters/            browser、desktop 运行时适配
  domains/             领域逻辑包
  web-domains/         领域 Vue 页面包
```

依赖方向为 App -> app-shell/adapters -> domain ports；领域逻辑不依赖具体宿主。Web 与 Desktop 复用领域功能，但分别选择浏览器 HTTP 和 Tauri IPC 适配器。

## 根治理

GitHub Actions、Hook、忽略规则、Taskfile、部署、数据库和发行资产只归仓库根管理。`scripts/quality/architecture-check.mjs` 校验目标目录与层级，`scripts/quality/compatibility-zero.mjs` 阻止已移除体系重新进入活动命令、CI 和文档。
