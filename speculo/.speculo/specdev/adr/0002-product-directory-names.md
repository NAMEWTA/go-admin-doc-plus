# ADR-0002: 固定产品子目录名称

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-002

## 上下文

单仓内的构建、CI、桌面宿主、部署和文档都依赖稳定的物理产品路径。

## 决策

Go 服务端固定在 `<Path>go-admin-plus/</Path>`，pnpm 前端工作区固定在 `<Path>go-admin-plus-ui/</Path>`。仓库不提供旧目录 symlink 或 alias。

## 后果

根任务隐藏日常路径细节；所有新脚本、workflow、生成器和文档必须使用这两个规范名称。
