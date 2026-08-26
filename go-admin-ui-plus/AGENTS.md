# AGENTS.md - Go Admin Plus Web 管理端

本文件只记录不遵守就会造成行为错误的前端约束。依赖版本和脚本以 `package.json` 为准，
标准列表页以 `domains/demo/src/pages/product/index.vue` 为准。

## Workspace 边界

- `apps/admin/`：应用入口、布局、固定路由和宿主适配。
- `domains/`：按业务领域组织页面、领域 API 与宿主注册。
- `packages/`：跨领域复用的 API client、app core、配置、契约、runtime 和 UI。
- `tests/`：单元、API 契约和 Playwright 测试。

领域代码不得反向依赖 `apps/admin`。跨领域能力进入明确的共享 package，不复制到多个领域。

## 列表页结构

业务列表页使用 `PageContainer`、`ProTable`、`useTable`、`useForm` 和 `useRemove`。
`ProTable` 负责搜索、loading、selection、sort 和分页管道，列仍使用 Element Plus 的
`el-table-column`。

- 文本列使用 `min-width`，选择列、固定控件列和操作列才使用固定 `width`。
- 操作列使用 `#actions`，每行最多两个直接操作，其余进入菜单。
- 搜索表单由 submit 处理回车，不额外绑定重复查询事件。
- 日期使用 `DateCell`，完整时间保留在 title。
- 新增是主操作；依赖选择的批量操作使用次级或 plain 样式。

组件 `name` 必须与后端菜单 `menu_name` 一致，否则 keep-alive 无法命中。

## Composables

`useTable` 与 `useForm` 返回 reactive 对象，在模板中直接访问属性，不解构也不使用
`.value`。`useRemove` 返回可解构 ref。

- 不分页集合使用 `paginated: false`。
- 默认排序同时传给 `useTable` 和 `ProTable`。
- 延迟首次加载使用 `immediate: false`。
- `resetQuery()` 通过工厂重建全部过滤条件。
- 表单通过 `:ref="form.bindFormRef"` 绑定，确保校验实例存在。
- 请求拦截器已经展示服务端错误，页面只恢复自身状态，不重复弹错。
- `useForm` 和 `useRemove` 已阻止重复提交，不另写并行确认流程。

字典统一使用 `useDict`，导出统一使用 `useExport`。导出默认只包含传入的当前行集合。

## API 与类型

领域 API 与页面放在同一 `domains/<name>/src` 边界；共享传输能力使用
`packages/api-client`。请求函数必须声明响应信封的 TypeScript 类型：

```ts
request<ApiResponse<PageResult<SysUser>>>(...)
```

响应信封为 `{ code, data, msg }`。拦截器对非 200 直接 reject，因此 resolve 分支只处理
成功数据。上传必须传 `FormData`，让浏览器生成 multipart boundary。

修改路由、DTO、Model 或 API 后运行 `pnpm check:api -- --require-models`，并提交更新后的
`packages/contracts` 生成物。

## 权限与路由

按钮权限使用 `v-permisaction="['模块:资源:操作']"`，角色权限使用 `v-permission`。权限码
必须与后端迁移种子一致。

业务路由由后端菜单动态生成；固定路由只维护登录、首页和错误页。承载子路由使用
`RouterViewKeepAlive`，确保多级菜单页面进入 keep-alive。

## Vue 与 Element Plus

新代码使用 `<script setup lang="ts">` 和 Composition API。组件声明 props、emits 和稳定的
组件名；不得修改 prop。样式穿透使用 `:deep()`。

`el-tag` 的 type 只能是 `primary`、`success`、`info`、`warning` 或 `danger`。模板中可直接
使用已全局注册的 Element Plus 图标；项目 SVG 位于 `apps/admin/src/icons/svg`，新增后运行
`pnpm svgo`。

## 外部资源与环境

默认构建不得依赖外部 CDN，确保内网和离线部署可用。Google Analytics 只有在
`VUE_APP_GA_ID` 非空时注入，仓库内所有 env 默认值必须为空。

后端地址只通过 `VUE_APP_BASE_API` 配置。不得提交 `.env.*.local`、凭据或生产内部地址。

## 验证

前端变更至少运行：

```bash
pnpm lint
pnpm type-check
pnpm test:unit
pnpm check:api -- --require-models
pnpm build:prod
```

页面交互、路由或布局变更还必须运行 `pnpm e2e`，并在桌面与常用笔记本宽度检查无重叠、
横向遮挡或动态尺寸跳动。
