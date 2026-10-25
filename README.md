# xbuilder-parkour

AI 辅助开发的 2D 横版跑酷原型。玩家横坐标固定，障碍向左移动，
空格跳跃，碰撞结束，可重开并保留当前进程内最高分。

## 当前状态

- 已实现玩家、世界、游戏状态机及测试。
- Go 核心可直接运行，无外部依赖；没有借用已有游戏 codebase。
- `spx/` 是基于 spx v2.0.4 / XGo v1.7.1 的原生接入源码，使用几何线框绘制。
  它依赖本地 Go 核心，不是已验证可上传 XBuilder 的独立 ZIP。
- 用户已反馈视口修正后的 XBuilder 网页项目可以游玩；完整验收和 demo 视频尚未完成。
- 核心与网页适配代码已合并到主干；各成员的实际开发与评审记录需维护。

## 运行网页版游戏

网页版是当前可交互的演示入口，不需要安装 Go、XGo 或本地引擎。

1. 从本仓库下载 [`parkour-web.xbp`](parkour-web.xbp)，不要下载 GitHub 的源码 ZIP 代替项目包。
2. 使用桌面版 Chrome 或 Edge 打开 [XBuilder](https://xbuilder.com/)，登录账号。
3. 通过项目菜单的 `Open project...` 打开下载的 `.xbp` 文件，按页面提示进入编辑器。
4. 点击编辑器中的运行按钮，点击游戏区域使其获得焦点。
5. 按空格开始，再按空格跳跃躲避障碍；空中不能再次跳跃。
6. 碰撞结束后按 `R` 重置，再按空格开始新一局。

游戏显示松鼠、草地、地面、障碍，以及 Score、Best、Controls 监视器。
最高分只在当前运行会话内保留。项目包与此前的
`parkour-web-viewport-fix.xbp` 内容一致，仅文件名简化。

导入注意事项、点击操作、故障处理与验收步骤见 [网页版运行说明](docs/web-running.md)。
当前网页模块规格、数据流及分工见 [架构说明](docs/architecture.md#网页版架构)。

## 立即运行核心演示

环境：Go 1.22 或更高版本。在仓库根目录执行：

```powershell
go test -cover ./...
go vet ./...
go run ./cmd/simulate -seconds 30
go run ./cmd/simulate -seconds 30 -autojump=false
```

第一条演示用自动跳跃控制器模拟游玩，第二条不跳跃，用于复现碰撞结束。
它们是终端逻辑演示，不是可交互画面，也不能替代比赛要求的游戏 demo 视频。
默认 seed 为 42，可用 `-seed 123` 换一组障碍。

本机验证：核心测试和 vet 通过。player、world、game 语句覆盖率分别为
95.2%、97.9%、100%；演示命令本身没有单元测试。
固定 seed 42 自动跳跃 30 秒仍为 running，得分 1337；
不跳跃约 2.44 秒进入 game_over，得分 76。
本机 `go test -race ./...` 因 CGO 未启用而无法执行，CI 配置在 Linux 上运行它，
远程 CI 尚未触发。

## spx 原生入口

`spx/main.spx` 将按键转为 game API，并把玩家、障碍映射到舞台中心坐标。
舞台尺寸为 960×540，空格开始或跳跃，R 重置到待开始状态。
状态、分数、最高分通过舞台 monitor 展示。

已使用 XGo v1.7.1 执行 `xgo go -t -tags pure_engine .`，语法与类型检查通过。
不带纯引擎标签的检查因本机原生绑定未准备而失败，尚未验证窗口绘制。

本地运行前需要完整 spx 主机运行环境，不能只安装一个 Go 包：

- Go 1.25+、XGo v1.7.1、spx v2.0.4。
- 按 [spx 官方说明](https://pkg.go.dev/github.com/goplus/spx/v2#readme-how-to-build)
  准备引擎主机资源；Windows 还需要对应构建工具和 mingw-w64。
- 保留仓库结构：`spx/go.mod` 通过
  `replace github.com/rxg456/xbuilder-parkour => ..` 引入核心。
- 在 `spx` 目录运行 `xgo run .`。本文不声称该命令已成功打开游戏窗口。

`spx/` 是本地 v2 接入实验，不是网页版运行入口。网页项目使用下文的
`web-native/` 独立 v3 源码和 `.xbp` 项目包，不导入本地 Go 核心，
也不能只粘贴某一个 `main.spx` 文件代替完整项目包。

## 分工与提交

| 成员 | 模块 | 建议分支 |
| --- | --- | --- |
| rxg456 | 玩家物理 | `feature/player-v2` |
| autoops-dev | 地图与障碍 | `feature/world-v2` |
| jesko-road | 游戏状态、集成与文档 | `feature/game-v2` |

AI 使用说明、建议分工、PR 流程和交付要求见 [团队流程](docs/team-workflow.md)；
模块接口、坐标和状态机见 [架构说明](docs/architecture.md)。

## 网页适配

`web-native/` 基于用户从 XBuilder 导出的 `parkour.xbp` 素材构建，
不导入本仓库 Go 模块，目标为网页端 spx v3.3.0。
保留草地背景与松鼠动画，舞台统一更新跳跃、障碍、碰撞和计分。
网页画面采用原始背景的 `480×360` 视口，逻辑坐标以半比例显示，
避免网页保持默认视口时把角色和计分界面裁到画面外。
初版采用两个循环回收的方形障碍，不与 Go 核心的随机序列完全等价。
用户已确认视口修正版可以游玩；导出后重新导入、边界碰撞和低帧率等完整验收仍待完成。

建议分工：rxg456 检查原生脚本与项目格式，autoops-dev 检查角色与地图画面，
jesko-road 负责集成验收与交付说明。各成员应实际参与使用 AI、修改、测试和评审，
以上安排不是已完成协作的证明。画面设计参考见 [渲染设计](docs/web-rendering.md)。

## Demo 视频

实际录屏由团队后续上传至 GitHub，并在此补充播放或下载链接。
当前尚未提供视频，终端模拟输出不能替代真实游戏录屏。
