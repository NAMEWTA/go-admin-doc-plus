# ADR-0006: 后端采用 Go 原生模块化单体

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-006

## 上下文

Go package 和 `internal` 可见性比复制 Java/Maven 顶级模块更适合表达当前单体的职责与依赖方向。

## 决策

`<Path>go-admin-plus/cmd/</Path>` 只处理进程入口和 launch material；`<Path>go-admin-plus/internal/app/</Path>` 拥有产品组合和跨模块 adapter；`<Path>go-admin-plus/internal/application/</Path>` 拥有应用级能力；`<Path>go-admin-plus/internal/host/</Path>` 拥有进程/HTTP/资源生命周期；`<Path>go-admin-plus/internal/modules/</Path>` 按业务能力内聚；`<Path>go-admin-plus/internal/contracts/</Path>` 只保存真实共享值；`<Path>go-admin-plus/internal/platform/</Path>` 实现技术端口。

## 后果

禁止 `common`/`other` 杂物箱、command 拥有运行时组合、模块间 production import 和 service locator。Go AST、command boundary 与 table-owner 测试持续验证该结构。
