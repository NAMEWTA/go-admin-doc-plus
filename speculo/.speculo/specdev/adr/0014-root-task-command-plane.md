# ADR-0014: 根 Taskfile 是唯一产品命令面

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-014

## 上下文

Go、pnpm、Rust、合同和发行命令如果分别由子项目、Hook 和 CI 定义，会产生不可复现的行为漂移。

## 决策

根 `<Path>Taskfile.yml</Path>` 暴露 dev/build/test/lint/generate/migrate/package/release 与治理命令。实现分别位于 `<Path>scripts/go-admin-plus/</Path>`、`<Path>scripts/go-admin-plus-ui/</Path>`、`<Path>scripts/contracts/</Path>`、`<Path>scripts/quality/</Path>`。唯一根 `<Path>.husky/</Path>` 与 CI 调用相同合同。

## 后果

包内 scripts 只服务包工具链，CI YAML 只负责编排。工具版本、pnpm resolver、失败传播和目标清单由根 contract test 锁定；本地 release 默认不发布外部制品。
