# xbuilder-parkour

三人协作的 2D 横版跑酷原型。玩家横坐标固定，障碍向左移动，
空格跳跃，碰撞结束，可重开并保留当前进程内最高分。

## 当前状态

- 已实现玩家、世界、游戏状态机及测试。
- Go 核心可直接运行，无外部依赖；没有借用已有游戏 codebase。
- `spx/` 是基于 spx v2.0.4 / XGo v1.7.1 的原生接入源码，使用几何线框绘制。
  它依赖本地 Go 核心，不是已验证可上传 XBuilder 的独立 ZIP。
- 尚未完成 XBuilder 云端导入、实际引擎画面验证、美术音效和 demo 视频。
- 首轮三个功能 PR 已合并；下一轮网页适配正在开发，人工审核记录仍需维护。

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

XBuilder 网页版是否允许本地 Go 模块及自定义依赖，仍需要目标环境验证。
不能只粘贴 `main.spx` 就认为接入完成；若平台不支持该依赖结构，需要将核心
移植为平台支持的 XGo 类文件，并重新验证，不应改用普通网页游戏冒充 XBuilder 作品。

## 分工与提交

| 成员 | 模块 | 建议分支 |
| --- | --- | --- |
| rxg456 | 玩家物理 | `feature/player-v2` |
| autoops-dev | 地图与障碍 | `feature/world-v2` |
| jesko-road | 游戏状态、集成与文档 | `feature/game-v2` |

来源说明、文件分组、PR 顺序和交付要求见 [团队流程](docs/team-workflow.md)；
模块接口、坐标和状态机见 [架构说明](docs/architecture.md)。

## 网页适配

`web-native/` 基于用户从 XBuilder 导出的 `parkour.xbp` 素材构建，
不导入本仓库 Go 模块，目标为网页端 spx v3.3.0。
保留草地背景与松鼠动画，舞台统一更新跳跃、障碍、碰撞和计分。
网页画面采用原始背景的 `480×360` 视口，逻辑坐标以半比例显示，
避免网页保持默认视口时把角色和计分界面裁到画面外。
初版采用两个循环回收的方形障碍，不与 Go 核心的随机序列完全等价。
网页实际编译、运行及导入往返仍需要验证，不能将源码准备视为完成验收。

下一轮分工：rxg456 审核原生脚本与项目格式，autoops-dev 审核角色与地图画面，
jesko-road 负责集成验收与交付说明。画面契约见 [渲染设计](docs/web-rendering.md)。
