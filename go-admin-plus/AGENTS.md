# AGENTS.md - Go Admin Plus 后端

本文件只记录不遵守就会造成行为错误的后端约束。版本以 `go.mod` 为准，命令以可执行配置
为准，标准单表 CRUD 以 `app/demo/` 为准。

## 运行架构

后端通过 `internal/application` 组装应用生命周期，通过 `internal/modules.Default()` 显式
注册 `runtime-queue`、`admin`、`demo`、`jobs` 和 `other` 模块。新增顶层模块必须实现
`application.Module`，并在默认模块集合中登记顺序、生命周期和测试。

业务请求遵循以下分层：

```text
Router -> API -> Service -> Model
```

API 不直接操作 ORM，Service 不引用 `gin.Context`。单表 CRUD 使用通用 Action 时可以省略
API 和 Service，但 Model、DTO、Router 的职责仍须保持分离。

## 单表 CRUD

`common/actions` 提供列表、详情、新增、修改和删除 Action。完整实现位于
`app/demo/`，包含并发安全、DTO 绑定、数据权限和迁移种子测试。

使用通用 Action 必须满足：

- Model 实现 `models.ActiveRecord`，包括 `Generate`、`GetId`、`TableName`。
- `Generate()` 返回副本；返回共享实例会在并发请求之间串数据。
- 列表 DTO 实现 `dto.Index`，写操作 DTO 实现 `dto.Control`。
- 详情和删除 DTO 内嵌 `dto.ObjectById`，复用 URI 与批量 ID 绑定。
- 列表和详情路由使用 `actions.PermissionAction()` 注入 DataScope。

只有跨表事务、外部调用或复杂业务校验才编写专用 API 与 Service。

## API 与 Service

API 结构体嵌入 `api.Api`。完成 `MakeContext`、`MakeOrm`、`Bind`、`MakeService` 链后必须
检查 `Errors`。响应统一使用 `OK`、`PageOK`、`Error`，不直接调用 `c.JSON`。

Service 结构体嵌入 `service.Service`，只使用请求上下文注入的 `e.Orm`。涉及数据权限的
查询必须组合 `actions.Permission(tableName, permission)`。错误向上返回，日志使用 Service
logger，不使用 `panic`。

## DTO 与 Model

搜索条件通过 `search` tag 声明，支持 `exact`、`iexact`、`contains`、`icontains`、`gt`、
`gte`、`lt`、`lte`、`order` 和 `left`。未声明的字段不得进入动态条件。

参与数据权限的表必须包含创建人字段，通常通过 `models.ControlBy` 提供。Model 必须显式
实现 `TableName()`；不要依赖 GORM 自动推导。

## 路由与权限

模块内路由通过 `init()` 加入对应 router slice，模块由 `internal/modules.Default()` 纳入
应用。认证路由至少包含 JWT 与 Casbin 中间件；需要 DataScope 的读取路由还要包含
`PermissionAction`。

权限标识格式为 `模块:资源:操作`，必须与前端 `v-permisaction`、`sys_menu`、`sys_api`、
`sys_menu_api_rule` 和 Casbin 规则一致。种子写法以
`cmd/migrate/migration/version/1786700001000_demo_menu.go` 为准。

## 数据库迁移

仓库交付的全部迁移位于 `cmd/migrate/migration/version/`。文件名前 13 位是版本号；已经
执行的版本不可修改，只能增加更高版本的新迁移。迁移必须可重复检查，必要时使用幂等
upsert，并提供针对目标数据库行为的测试。

`version-local/` 只用于 Git 忽略的本机实验，任何产品变更都不得放入该目录。

## OpenAPI

公开 Handler 维护完整 Swagger 注解。修改路由、DTO 或响应后运行：

```bash
go generate ./...
node api/openapi/generate.mjs
```

提交 Swagger 2 生成物、规范化 OpenAPI 3 工件和相应前端契约。不得手工制造与运行路由
不一致的 API 文档。

## 配置与运行

- SQLite 命令必须带 `-tags sqlite3`。
- 生产配置必须使用 `mode: prod`、独立 JWT secret 和外部化凭据。
- 代码生成的前端根是 `../go-admin-ui-plus/apps/admin/src`。
- 不使用全局 DB 变量；多数据源隔离依赖请求上下文中的 ORM。
- 不提交真实凭据、本地数据库、日志、上传文件和临时文件。

## 验证

后端变更至少运行 `go test ./...`。涉及 SQLite、桌面宿主、迁移或 OpenAPI 时，还要运行
对应带 tag 测试、契约生成和目标平台验证。
