# 部署

`deploy/compose/compose.yml` 提供两个互斥的 Server profile：

- `sqlite`：Server 与持久化 SQLite volume。
- `postgres`：Server、PostgreSQL 17 与独立产品数据 volume。

两者都包含只读 Web 容器、健康检查、最小 capability、内部后端网络和显式 secret 文件。生产部署必须使用已验证摘要的镜像，并在启动前通过 `scripts/release/linux/verify-policy.mjs`。

```bash
scripts/release/linux/prepare-secrets.sh
scripts/release/linux/compose.sh sqlite config
scripts/release/linux/compose.sh postgres config
```

运行时 secret 位于 `deploy/compose/runtime/secrets/`，不得提交到 Git。具体镜像构建和验收见 [Linux 发行说明](../release/linux/README.md)。

PostgreSQL profile 由一次性 `migrate-postgres` 服务独占执行迁移；`api-postgres` 和
`worker-postgres` 都以 `condition: service_completed_successfully` 等待它。API/worker 只检查
schema 与 readiness，过旧、过新、未知版本或迁移失败都会阻止启动。SQLite profile 由单实例
宿主在持有文件锁时向前迁移，不得与 PostgreSQL profile 在同一 Compose project 并行启动。

升级前备份所选数据库 volume 与对应 product-data volume，并记录当前镜像摘要和迁移版本。
验证 `config` 后先观察迁移服务退出码，再观察 API readiness，最后运行：

```bash
scripts/release/linux/verify-compose.sh postgres
# SQLite 部署使用：scripts/release/linux/verify-compose.sh sqlite
```

schema mismatch 时保持 API/worker 停止，保留 volume 与日志；恢复完整备份或提供修正后的前向
迁移再重试。普通 `down` 不删除 volume，任何显式 volume 删除都是需单独批准的数据销毁操作。
