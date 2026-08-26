# 发布说明

当前产品版本由一个 monorepo commit 和一个三段式版本号共同标识。平台 workflow 构建并
验证候选工件，产品 manifest 再将三个平台结果收敛为同一份来源记录。

| 平台 | 形态 | 文档 |
| --- | --- | --- |
| Linux AMD64 | Docker Compose | `release/linux/README.md` |
| macOS ARM64 | Wails DMG，自用签名策略 | `release/macos/README.md`、`INSTALL.md` |
| Windows AMD64 | Wails NSIS，自用签名策略 | `release/windows/README.md`、`INSTALL.md` |

本地预检：

```bash
task release:preflight VERSION=0.1.0
```

平台门禁需要显式 dispatch。现有 workflow 只上传短期 Actions artifact，不创建 GitHub
Release、不推送生产镜像、不部署环境，也不改变系统级安全设置。外部发布必须通过独立
授权和发布流程。

每个平台的安装文档包含校验和、签名状态及系统安全警告。制作或分发候选工件时不得省略
这些文件，也不得把未签名产物描述为受 Apple 或 Microsoft 信任。
