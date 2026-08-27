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
