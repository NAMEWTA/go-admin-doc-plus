# ADR-0020: 三 Profile 使用不可变类型化配置

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-020

## 上下文

全量 mutable config、业务代码读环境和原始配置打印会让宿主能力不清并泄漏 secret。

## 决策

`server-postgres`、`server-sqlite`、`desktop-sqlite` 各自拥有最小强类型 schema。启动时按 defaults、config file、environment、显式 non-secret CLI 合成并一次校验，再通过 product/host 构造函数注入不可变值。Secret 只来自环境或 `_FILE` reference；Desktop 路径、随机端口与启动控制由 Tauri host 提供。

## 后果

禁止全局 singleton/setter、运行时 reload、secret CLI flag、明文 merged config 和原始 dump。验证错误只报告字段与规则，不输出敏感值。
