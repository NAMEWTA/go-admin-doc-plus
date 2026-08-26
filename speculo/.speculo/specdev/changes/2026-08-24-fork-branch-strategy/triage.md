---
schema_version: 1
artifact: triage
change: 2026-08-24-fork-branch-strategy
mode: intake
source: <Path>{roots.state}/specdev/changes/2026-08-24-fork-branch-strategy/source.md</Path>
classification: operations
risk: medium
route: specdev/implement
ready_for_implementation: true
external_action: not-applicable
updated_at: 2026-08-24T11:35:00+08:00
---

# Triage: 为两个 fork 建立上游同步与二次开发主线

## 当前判定

- **影响：** 改变两个 fork 的长期分支职责、远端默认分支和本地 Git 行为；不改变业务代码。
- **紧急度：** normal
- **当前证据：** 两个仓库的 `master` 均与官方 `upstream/master` 一致；两个 fork 的远端默认分支当前均为 `master`；后端工作区存在用户未提交改动；两个目录均为父仓库 submodule。
- **相关代码/工件：** `<Path>.gitmodules</Path>`、两个子仓库的 Git refs 与本地配置

## 未知项

- **可发现事实：** GitHub 分支写入权限和默认分支切换结果可通过 GitHub CLI 核验。
- **需要用户决定：** 无；用户已批准已给出的双主线方案实施。
- **低影响实现细节：** 安全配置按仓库级配置写入；不自动提交后端既有未提交内容。

## 路由

- **下一 Work：** `<Path>{roots.workflows}/specdev/I-implement/I-implement.md</Path>`
- **理由：** Direct Spec 已冻结分支职责、验收合同、外部写入授权和验证方式，不需要额外 Ticket 拆分。

## 外部动作

- **远程目标：** `<Url>https://github.com/NAMEWTA/go-admin-plus</Url>`、`<Url>https://github.com/NAMEWTA/go-admin-ui-plus</Url>`
- **关闭能力：** not-applicable
- **当前状态：** not-applicable
- **授权记录：** 用户于当前对话明确要求按已确认方案实施，涵盖分支创建、推送和默认分支调整。
- **尝试与结果：** 尚未执行。

外部动作只投影最终完成，不替代本地状态、Ticket、Map 或 Evidence。
