# Go Admin Plus 产品化重构上下文

**产品化补强**：保留当前已经验证的模块化单体、pnpm workspace、OpenAPI 与 Tauri 2 架构，补齐首次初始化、安全、交互、运维和交付闭环。
_Avoid_：第二轮目录搬迁、按参考项目外观重写

**互补融合**：以 Go Admin Plus 为技术与安全底座，吸收 Backplane 的运维易用性以及 RuoYi/Plus UI 的脚手架和后台交互成熟度。
_Avoid_：复制参考项目依赖栈、默认凭据或弱安全模式

**表现层重构**：保持当前业务 API、生成 transport、无头领域与 Web/Desktop 共享业务组合，只替换路由、布局、组件、主题和页面交互实现。
_Avoid_：以 UI 重构为由改写业务合同、复制双 App 页面

**真实交付门禁**：在功能编码与构建完成后，以真实 PostgreSQL、真实浏览器和适用原生 Desktop 流程证明候选行为；required gate 缺少环境时必须失败。
_Avoid_：缺环境即成功、只编译不运行、把未执行记为通过

**完整计划快照**：当前 change 中保存本轮审查结论、目标结构、分阶段路线、验收门槛和回退边界的可读设计输入。
_Avoid_：实现授权、Ticket 状态权威、Evidence

**首次管理员 Bootstrap**：只在账号为空时执行的一次性 IAM 初始化用例；Server 通过离线 CLI 调用，Desktop 通过首次设置流程调用。
_Avoid_：默认管理员、固定密码 SQL、未认证 Web 初始化 API

**账号 Tombstone**：账号删除后的不可登录稳定身份记录，用于保持审计引用并等待跨模块资源完成转移或清理。
_Avoid_：直接硬删除账号、跨模块同步级联删除

**持久登录限流**：由当前正式数据库保存、账号桶与来源桶共同判定且不能关闭的登录防暴力破解状态。
_Avoid_：单进程并发预算、只按 IP、Redis 限流事实源

**组织数据范围**：授权查询可使用 all、self、organization、organization-tree 或 custom 组织集合约束可见业务数据。
_Avoid_：租户范围、只在前端过滤

**产品 CLI**：单一 go-admin-plus Server 二进制的 serve、worker、migrate、bootstrap、doctor 和 version 子命令集合。
_Avoid_：根 Taskfile、Desktop sidecar、多个 Server 运维二进制

**Session 活跃续期**：通过显式 heartbeat、受保护业务写请求或提前 renew 更新 Session 空闲期限；普通认证 GET 不产生续期写入。
_Avoid_：每个认证请求刷新、每请求轮换 CSRF

**管理员离线恢复**：所有管理员不可用时，在停止 API/worker 且持有直接数据库访问权的前提下，通过 recover-admin 子命令重置管理访问的受审计运维流程。
_Avoid_：重新执行 Bootstrap、远程恢复 API

**显式文件处置**：删除拥有文件的账号前必须选择 transfer 或 purge；前者指定目标账号，后者通过可靠异步流程删除内容。
_Avoid_：默认系统 owner、隐式级联删除

**角色范围并集**：用户通过多个角色获得同一 Permission Code 时，只把这些实际授权角色的数据范围合并为最终可见集合。
_Avoid_：所有角色无条件参与、前端合并范围、隐式 deny

**模糊限流反馈**：普通登录失败统一返回凭据错误；触发持久限流后只暴露粗粒度可重试时间。
_Avoid_：剩余次数、精确账号桶状态、账号存在性提示

**Profile 进程拓扑**：PostgreSQL 生产 API 与 worker 分进程，Server SQLite 和开发形态可一体运行，Desktop sidecar 始终拥有本地 API 与 worker。
_Avoid_：每个 PostgreSQL API replica 执行 worker、SQLite 多实例

**可恢复管理员账号**：未删除的既有账号，可由严格离线 recover-admin 流程重新启用、重置密码、授予系统管理员角色并撤销旧 Session。
_Avoid_：救援账号、复活 Tombstone、重复 Bootstrap

**Purge 取消边界**：永久文件删除在 worker claim 前可以取消；开始物理删除后不可恢复。
_Avoid_：隐藏回收站、已开始删除仍承诺恢复

**Profile Migration 策略**：PostgreSQL 生产显式迁移且运行角色只校验 schema；Server SQLite 自动迁移；Desktop 在备份成功后自动迁移。
_Avoid_：PostgreSQL 多副本竞相迁移、Desktop 无备份迁移

**账号主部门**：账号至多拥有一个主部门并可分配多个岗位；组织类数据范围从主部门及部门树计算。
_Avoid_：多主部门、由岗位隐式推导主部门

**默认后台视觉**：中性浅色工作面、炭黑侧栏、克制绿色品牌强调和紧凑信息密度，并提供可持久化暗色模式。
_Avoid_：装饰性渐变、单色铺满、默认低密度营销式布局
