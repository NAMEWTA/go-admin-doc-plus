# 产品发行 Manifest

`product-release.mjs` 将 Linux 双架构 OCI、macOS Universal 签名公证候选和 Windows x64 签名候选聚合为一个 schema v2 产品来源记录。

三个平台必须来自同一精确 source SHA 和产品版本，并由各自受保护 workflow 成功产生。聚合器校验 workflow 身份、run title、artifact 名、archive SHA-256、SBOM、签名状态和迁移/OpenAPI 来源。生成 manifest 不等于授权上传或公开发行。

```bash
node release/manifest/product-release.mjs preflight --version 0.0.1 --allow-dirty
node --test release/manifest/product-release.test.mjs
node release/manifest/scan-compatibility.mjs
```
