# ADR-0003: pnpm Workspace 是前端模块边界

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-003

## 上下文

目录拆分只有在 package 声明真实依赖、公开 exports 且可独立验证时才构成架构边界。

## 决策

前端使用单一 pnpm workspace。每个激活 package 必须声明真实依赖并只通过公开 exports 消费内部包；禁止循环、跨包 `src` deep import、未声明依赖和 App 反向承载业务实现。当前门禁动态验证全部 workspace package 的 typecheck，以及拥有 spec 的 package test。

## 后果

新增 package 必须同时加入 workspace、manifest、lockfile 和独立验证合同。根递归命令不能用 `--if-present` 掩盖缺失的 package contract。
