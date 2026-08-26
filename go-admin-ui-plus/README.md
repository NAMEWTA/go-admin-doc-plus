# Go Admin Plus Web 管理端

本目录是 Vue 3、Element Plus、TypeScript 和 pnpm 组成的前端 workspace。

## Workspace

| 路径 | 职责 |
| --- | --- |
| `apps/admin/` | Web 应用入口、布局和固定路由 |
| `domains/` | demo、jobs、monitor、system、tools 领域页面 |
| `packages/` | API client、app core、配置、契约、runtime 和 UI 公共包 |
| `tests/` | 单元、契约和 Playwright 测试 |

业务列表页使用 `PageContainer`、`ProTable` 和 composables。标准页面位于
`domains/demo/src/pages/product/index.vue`。

## 常用命令

```bash
corepack enable
corepack install --global pnpm@11.1.3
pnpm install --frozen-lockfile
pnpm dev --host 0.0.0.0
pnpm test:ci
pnpm e2e
pnpm build:prod
```

本地后端地址写入 Git 忽略的 `.env.development.local`：

```dotenv
VUE_APP_BASE_API = 'http://localhost:8000'
```

## 开发入口

- [仓库开发指南](../docs/development.md)
- [仓库架构](../docs/repository-architecture.md)
- [前端强约束](AGENTS.md)
- API 契约检查：`scripts/check-api-contract.mjs`

项目许可证见仓库根 [LICENSE](../LICENSE)。
