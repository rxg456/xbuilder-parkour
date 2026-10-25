# 三人协作与 PR 流程

## 来源与分工

本原型由三个 AI agent 辅助生成 player、world、game 及测试，
主 agent 补充入口、配置、CI 和文档。经用户确认账号持有人授权，代理使用
对应账号的 SSH 密钥推送，并按模块设置提交作者；这不表示三位成员独立编写或已完成人工审核。
成员仍需实际审核、测试和完善，个人贡献以真实修改及评审记录为准。

| 成员 | 模块 | 分支 | 人工审核重点 |
| --- | --- | --- | --- |
| rxg456 | 玩家物理与测试 | `feature/player-v2` | 跳跃手感、落地、禁止连跳、异常时间步 |
| autoops-dev | 世界、障碍与测试 | `feature/world-v2` | 障碍间距、速度递增、回收、碰撞范围 |
| jesko-road | 游戏状态、集成与文档 | `feature/game-v2` | 计分、结束、重开、集成和交付材料 |

## 安全与登录

之前暴露的私钥应撤销；用户已确认本地密钥已更换并授权使用。
私钥内容不写入仓库、文档或日志，不修改全局 Git 配置。

每位成员在自己的设备或隔离登录环境中使用本人合法凭据。
提交前查看 `git var GIT_AUTHOR_IDENT`，确认作者身份准确；GitHub 登录和 Git 作者
是两套信息，仅切换 SSH 凭据不会自动改变 commit 作者。

SSH 可用于 Git 推送，不能创建 PR。PR 需要成员网页登录或 GitHub CLI 授权，
不把 token 写进项目。当前尚未创建 PR。

## 三个 PR 的提交顺序

三个分支采用依赖链，分别只新增对应模块。创建 PR 时选用下面的 base，
不要直接把后两个分支都对比 main，否则会包含前置模块。
按玩家、世界、集成顺序审核合并；前置 PR 合并后再将后续 PR 的 base 改为 main，
并核对差异，必要时更新分支。为避免 squash/rebase 后重复提交，
前置分支建议使用保留提交的 merge 方式。不要向 main 直接推送。

| PR | 纳入文件 | 初始 base |
| --- | --- | --- |
| 玩家 | `go.mod`、`.gitignore`、`internal/player/` | `main` |
| 世界 | `internal/world/` | `feature/player-v2` |
| 集成 | `internal/game/`、`cmd/`、`spx/`、`.github/`、`README.md`、`docs/` | `feature/world-v2` |

已推送分支可通过 GitHub 网页选择 base/head 创建 PR。
若使用已登录的 GitHub CLI，可执行：

```powershell
gh pr create --base main --head feature/player-v2 --title "feat: player jumping and physics"
gh pr create --base feature/player-v2 --head feature/world-v2 --title "feat: obstacle generation and collision"
gh pr create --base feature/world-v2 --head feature/game-v2 --title "feat: game state and spx integration"
```

以上 PR 命令尚未执行。PR 说明实现目的、测试和待验证项，并链接本文件的来源说明，
请另一位成员审核后合并。后续修复也走分支和 PR，保留真实开发记录。
此次为用户确认的分支重建，旧提交保存在本地 `archive/pre-v2`；
删除旧远程分支不意味着彻底清除 GitHub 中的旧提交。

## 提交清单

- 三位成员的真实 commit、PR URL、审核与后续修改记录。
- 源代码、运行说明、架构与分工文档。
- XBuilder/spx 实际运行验证，确认目标议题和参赛组要求。
- 实际录制的 demo 视频及播放方式，当前尚未录制。
- 基线标签：当前仓库初始仅有 README，未导入第三方游戏 codebase，也未创建 `v0`。
  如果后续采用既有游戏代码，必须在修改该基线前按比赛要求打 `v0` 并注明来源和许可证。
- 按截图，截止时间为 2026-10-26 09:00；时区以主办方说明为准，截止前完成
  审核与交付，截止后不再修改仓库。
