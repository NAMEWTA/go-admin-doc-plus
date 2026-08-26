# 产品发布契约

本目录负责把 Linux AMD64 Compose、macOS ARM64 和 Windows AMD64 三个平台候选收敛为
一个产品来源记录。根、后端和前端 provenance 字段必须指向同一个 monorepo commit。

生成的 manifest 只是候选工件，不是发布指令。策略始终记录
`external_publish_authorized: false`，workflow 只上传短期 GitHub Actions artifact。

本地检查：

```bash
node --test release/manifest/product-release.test.mjs
node release/manifest/scan-compatibility.mjs
node release/manifest/product-release.mjs preflight --version 0.1.0
```

`task release:dispatch VERSION=0.1.0` 在当前精确提交启动三个平台门禁；全部成功后，将 run
与 artifact ID 传给 `task release:collect`。这些命令不会创建 GitHub Release、部署
Compose、修改平台安全设置或签名二进制文件。
