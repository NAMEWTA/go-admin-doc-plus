---
schema_version: 1
artifact: source
change: 2026-08-24-fork-branch-strategy
source_type: conversation
canonical_locator: null
captured_at: 2026-08-24T11:28:57+08:00
content_sha256: 6b5304976f46c011f1022bde6b1b5095168aebe2e234ae50d8e51244a0a83b44
remote_state: not-applicable
close_capability: not-applicable
---

# Source: 为两个 fork 建立可持续的上游同步分支模型

## Capture Metadata

- **Capture method:** conversation
- **Author:** user
- **Created / updated:** 2026-08-24 / 2026-08-24
- **Labels or classification supplied by source:** Git 分支治理、上游同步、二次开发
- **Attachments:** none
- **Redactions:** 机器绝对路径按持久化约定省略，仓库以项目相对名称表示

## Original Content

用户需要基于 `go-admin-plus` 和 `go-admin-ui-plus` 进行二次开发。fork 后要持续同步官方上游更新，将更新 merge 回二次增强开发分支，在保留可查看的 Git merge 图关系的同时继续自己的开发。用户要求设计各仓库分支、fork 默认主分支和上游 merge 方式。

在收到“双主线：`master` 保持上游镜像、`main` 作为二次增强默认主线”的完整方案后，用户确认：“非常棒，按你的建议，进行实施。”

## Source Comments

无。
