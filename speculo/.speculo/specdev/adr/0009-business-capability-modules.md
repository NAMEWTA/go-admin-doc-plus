# ADR-0009: 后端采用八个业务能力模块

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-009

## 上下文

旧 `admin`、`other`、`jobs` 聚合多个变化原因，并造成跨模块反向依赖。

## 决策

`<Path>go-admin-plus/internal/modules/</Path>` 下固定 iam、organization、settings、audit、scheduler、generator、files、demo 八个业务能力模块。Health/readiness/metrics/status 由 `<Path>go-admin-plus/internal/application/health/</Path>` 提供并由 Server/Desktop host 接入；它们不是业务模块。Host 资源生命周期位于 `<Path>go-admin-plus/internal/host/</Path>`。

## 后果

每个模块拥有自己的用例、migration、persistence 和 transport。禁止恢复 `system`、`operations`、`other` 或 platform observability 杂项聚合。
