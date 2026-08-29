# ADR-0005: 根级资产按生命周期统一所有权

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-005

## 上下文

Git 治理、任务、部署、发行和数据库工程资产影响整个产品，但拥有不同生命周期。

## 决策

Git 根统一拥有 `<Path>.github/</Path>`、`<Path>.husky/</Path>`、`<Path>scripts/</Path>`、`<Path>deploy/</Path>`、`<Path>release/</Path>` 与 `<Path>database/</Path>`。子项目不保留重复 workflow、Hook 或通用部署/发行资产。App manifest、Tauri `<Path>src-tauri/</Path>` 和模块 migration 仍由所属 App/模块拥有。

## 后果

部署定义不混入发行逻辑，发行定义不承载运行时配置，根 database 不取代模块生产 migration。治理扫描拒绝嵌套 `.github`、`.husky` 和 `.gitignore` 所有权回退。
