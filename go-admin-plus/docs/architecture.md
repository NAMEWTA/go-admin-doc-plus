# 后端架构

后端是一个可由服务器或桌面宿主启动的模块化单体。`internal/application` 管理配置、存储、
路由、模块启动和关闭；`internal/profile` 为 server 与 desktop 选择数据库、缓存和队列；
`internal/modules` 以固定顺序组装业务模块。

## 请求路径

```text
Host -> Gin Engine -> Middleware -> Module Router
                                  -> common Action -> DTO -> Model
                                  -> API -> Service -> Model
```

单表 CRUD 使用 `common/actions`。跨表事务、外部调用和复杂校验使用 API + Service。ORM
由请求上下文注入，Service 不接触 Gin context。

## 数据权限

`actions.PermissionAction()` 读取当前用户 DataScope 并写入请求上下文，查询通过
`actions.Permission(tableName, permission)` 追加 GORM Scope。参与过滤的表必须包含
`create_by`，通常通过 `models.ControlBy` 提供。

| DataScope | 行为 |
| --- | --- |
| `1` | 全部数据 |
| `2` | 当前角色关联部门 |
| `3` | 当前部门 |
| `4` | 当前部门及子部门 |
| `5` | 当前用户创建的数据 |

服务器配置 `application.enabledp` 控制 DataScope 总开关。关闭后不会追加过滤条件。

## 模块生命周期

模块实现 `ID`、`RegisterRoutes`、`Migrations`、`Start` 和 `Stop`。默认顺序由
`internal/modules/default.go` 固定并由测试锁定。异步队列在应用启动后注册处理器，在关闭
时等待退出或服从 context deadline。

## 数据库迁移

产品迁移位于 `cmd/migrate/migration/version/`，按 13 位版本号升序执行，结果记录在
`sys_migration`。已执行迁移不可改写；结构、种子或品牌默认值变化都通过新迁移演进。

服务器 profile 支持 PostgreSQL、MySQL、SQL Server 和 SQLite。桌面 profile 使用应用数据
目录内的 SQLite，并在升级前保留迁移备份。

## OpenAPI

`docs/admin/` 是由 Swagger 注解生成的 OpenAPI 2 工件，`api/openapi/openapi.json` 是经
路由校准后的 OpenAPI 3 权威契约。后端 contract test 双向核对运行路由，前端再由该工件
生成类型并校验页面字段、fixture、DTO 和 Model。

## 运行状态

日志、上传、临时文件、数据库与渲染后的密钥配置属于运行时状态。服务器部署由 Compose
volume 和 runtime secret 目录持有；桌面应用使用平台应用数据目录。源码树只保存模板、
迁移和可复现测试数据。
