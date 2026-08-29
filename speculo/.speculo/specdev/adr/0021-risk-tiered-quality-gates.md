# ADR-0021: 质量门禁按风险分层

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-022；用户个人自用决定（2026-08-29）

## 上下文

格式、架构、双数据库、浏览器、原生宿主和公开发行风险的最早可信反馈位置不同。

## 决策

Local/Hook 执行快速静态、生成漂移、secret 和零兼容检查；PR/集成执行 Go/pnpm/Rust、架构、OpenAPI、双方言、认证授权、可靠运行时、Web 与适用 native tracer；Protected Release 在需要公开分发时执行目标平台构建、安装、签名/公证、SBOM/provenance 和 trust。个人自用完成不要求签名、公证或受保护安装，但不得把未执行保护动作记为 passed。

## 后果

根 Taskfile 是本地可复现合同，CI 只编排。Required gate 失败阻断对应阶段；flaky test 不得靠静默 retry 变绿。任何临时豁免必须有明确 owner、范围、理由和恢复条件。
