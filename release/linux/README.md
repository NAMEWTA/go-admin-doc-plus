# Linux AMD64 Compose 发布

发布包只向宿主机暴露 Web 服务。API、迁移、PostgreSQL 和 Redis 只存在于 Compose
网络，默认绑定 `127.0.0.1:8080`；远程访问必须在前面部署 TLS 反向代理。

从源码构建并启动：

```bash
cp deploy/compose/.env.example deploy/compose/.env
./release/linux/prepare-config.sh
docker compose --env-file deploy/compose/.env \
  -f deploy/compose/compose.yml \
  -f deploy/compose/compose.build.yml \
  up -d --build --wait
```

一次性 `migrate` 是第一个修改数据的服务，迁移失败会阻止 API 启动。PostgreSQL、Redis
和文件数据由 named volume 持久化。运行时密钥保存在 Git 忽略的
`deploy/compose/runtime/`；`prepare-config.sh` 只创建缺失值，不覆盖已有非空配置。

发布部署使用 manifest 中的不可变镜像 digest，省略 `compose.build.yml`。升级前备份三个
named volume 和 runtime 目录；回滚时恢复这些数据并使用上一版 digest。

验证和停止：

```bash
./release/linux/verify-compose.sh
docker compose --env-file deploy/compose/.env \
  -f deploy/compose/compose.yml \
  -f deploy/compose/compose.build.yml down
```

除非明确要永久删除全部 volume 数据且已有验证通过的备份，否则不要对 `down` 使用 `-v`。
