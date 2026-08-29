# ADR-0010: 固定发行能力矩阵与个人自用边界

**状态：** Accepted
**日期：** 2026-08-29
**来源：** `<Path>{roots.state}/specdev/archive/2026-08/2026-08-26-project-architecture-reconstruction/ADR.md</Path>` ADR-010；用户个人自用决定（2026-08-29）

## 上下文

工具链可构建的目标不应自动成为日常交付承诺；个人自用也不需要生产签名、公证和受保护安装。

## 决策

仓库保留 Linux amd64/arm64 OCI/Compose、macOS Universal App/DMG 和 Windows x64 NSIS 的构建与 policy 能力，不发布 Linux Desktop。个人自用完成只要求当前可构建、可自行安装打开和本地产品 E2E；Developer ID、notary、Gatekeeper、Authenticode 和受保护安装不是个人自用门禁。公开分发时必须重新启用对应 protected workflow。

## 后果

跨平台 workflow、identity、SBOM/provenance 和签名接缝继续维护，但不能把未执行保护动作记为 passed。个人自用豁免不等于删除未来发行能力。
