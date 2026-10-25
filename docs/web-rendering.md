# XBuilder 原生精灵画面设计

## 范围与结论

目标是在 XBuilder 网页项目中用原生 Sprite 代替 `debugDrawRect`，不导入
`github.com/rxg456/xbuilder-parkour/internal/*`，不依赖本机 `replace` 或绝对路径。
建议由 autoops-dev 检查角色与地图；本文不是任何成员的已完成评审记录。
以下保留早期基于本地 spx v2.0.4 的设计研究，不能直接视为当前实现。
后续独立网页项目使用 spx v3.3.0、松鼠动画和绘图地面及障碍，
实际采用 `480×360` 视口，逻辑坐标按半比例显示，玩家位置为
`(-150, feet/2-104)`，地面顶边为 `-120`。下文的 `960×540` 布局、
克隆障碍和 UI 坐标是早期方案，不是当前配置。
用户已反馈视口修正版可以游玩；完整画面和碰撞验收仍需记录实际结果。

## 已证实的 API 与来源

下表路径相对于本机 `D:\dev\go\pkg\mod\github.com\goplus\spx\v2@v2.0.4\`。
“已证实”仅指读取到示例或实现；不表示以下组合方案已在网页执行。

| 能力 | 确凿示例 / 实现 |
| --- | --- |
| Sprite 资源结构 | `tutorial/03-Clone/assets/sprites/Calf/index.json:2-18`：`costumes` 中有 `name/path/x/y`，精灵有 `costumeIndex/size/visible/x/y`；造型与精灵的 `x/y` 不能混淆 |
| 事件与造型切换 | `tutorial/01-Weather/Kai.spx:1-20`：`onStart => {...}`、`onMsg "6", => {...}`、`broadcast "1"`、`setCostume Next` |
| 创建及初始化克隆 | `tutorial/04-Bullet/MyAircraft.spx:4`：`Bullet.clone`；`Bullet.spx:1-8`：`onCloned => {...}`、`setXYpos MyAircraft.xpos, MyAircraft.ypos+5`、`show`、`destroy` |
| 每帧与秒制移动 | `tutorial/05-Animation/SmallEnemy.spx:14-17`：`waitNextFrame`、`changeYpos -48 * deltaTime`，出底边后 `destroy` |
| 按 ID 回收、消息屏障 | `tutorial/03-Clone/Calf.spx:9-20`：克隆保存 `id = gid`，消息中按 ID `destroy`；`Arrow.spx:2`：`broadcastAndWait "undo"` |
| 背景与图层 | `tutorial/01-Weather/assets/index.json:2-16`：`backdrops`、`backdropIndex`、`stretchMode`、`zorder`；不能把 `stretchMode` 当作精灵宽高拉伸 API |
| 变量监视器 | `tutorial/03-Clone/assets/index.json:8-30`：`type: "monitor"`、`target: ""`、`val: "getVar:gid"`、`label/mode/x/y/visible` |
| 分辨率与比例 | `tutorial/01-Weather/assets/sprites/Kai/index.json:2-26`：造型 `bitmapResolution: 2`、精灵 `size: 0.95`；`component_transform.go:225-237`：`SetSize` 设置统一 `Scale` |
| 锚点与显示尺寸 | `costume.go:96-117,160-163`：造型 `x/y` 存为中心，尺寸除以 `bitmapResolution`；`sprite_render_util.go:28-36`：中心、比例参与渲染偏移 |
| 克隆字段与隐藏 | `sprite_clone.go` 的 `cloneSprite`：复制用户字段，引用字段并非深拷贝；`sprite_render.go` 的 `Hide/Show` 是本地 API，网页脚本调用仍待验 |
| UI 坐标 | `internal/ui/ui_monitor.go:67-69` 调用 `ViewToUI`；`internal/ui/util.go:54-58`：Y 取反并加视口半尺寸 |

现有 `spx/main.spx` 的 `onKey KeySpace`、`onKey KeyR` 是项目源码证据，
不是教程或网页运行证据。只复用这些小型语法模式，不在本轮生成完整新脚本。

## 网页项目划分

在网页编辑器建立舞台脚本及 `Player`、`Obstacle`、`GroundA/GroundB` 精灵，
由编辑器维护脚本、造型和资源索引；不要假定仓库 ZIP 可直接导入。
计划对应 `main.spx`、各精灵 `.spx`、`assets/index.json`、
`assets/sprites/<名称>/index.json` 与真实造型文件，命名关联以编辑器导出验证。

- 舞台是唯一逻辑写入者：维护状态、脚底 Y、速度、障碍记录、距离、分数及最高分。
  将现有 Go 算法迁成平台支持的 XGo 字段与计算，不调用本地 Go 包。
- Player 只读取舞台快照、设置位置、切换跑步/腾空造型，不另起一套重力循环。
- Obstacle 的隐藏原件只负责生成；每个克隆保存稳定 ID，读取对应记录并显示。
  ID 不采用会因回收而移动的数组下标。克隆不生成下一代、不单独更新计分。
- GroundA/GroundB 只负责视觉滚动；背景用静态 backdrop，UI 用舞台 monitor。

单一帧入口取 `deltaTime`，忽略非法/非正值，夹紧到 0.05 秒，再分成不超过
1/120 秒的小步。更新前检测一次；每步先更新玩家和世界，再严格 AABB 检测，
碰撞即终止剩余步骤；
整帧完成后同步精灵。显示消息处理器必须短且无 `wait`，避免阻塞帧屏障。
网页端共享字段、记录集合和消息顺序需先用一个玩家、一个障碍验证。

## 坐标与造型契约

舞台 960×540，中心为原点、Y 向上，地面 Y=-170。逻辑玩家脚底为 `pY`，
左边 X=160，宽 40、高 64；障碍 `X` 是左边，脚底为逻辑 Y=0。
所有造型采用中心锚点、无额外 pivot、不旋转，按下面位置显示：

| 对象 | 舞台中心 |
| --- | --- |
| 玩家 | `x = 160 + 40/2 - 480 = -300`；`y = pY + 64/2 - 170 = pY - 138` |
| 障碍 | `x = X + W/2 - 480`；`y = H/2 - 170` |
| 地面图块 | 高度 T 时 `y = -170 - T/2`，顶边必须落在 -170 |

以独立 PNG 为首轮尺寸基准。原图像素 Pw×Ph、分辨率 r、精灵倍率 s，
未旋转显示尺寸为 `Pw/r*s × Ph/r*s`；中心锚点填写 `Pw/2, Ph/2`。
优先使用 r=1、s=1、整数尺寸，避开尺寸查询的整数除法与平台重采样差异。
例如最终玩家造型画布 40×64、锚点 20/32，或 80×128、r=2、锚点 40/64。
跑步和跳跃帧须统一画布、锚点、脚底与主体包围盒，不能照搬 Kai 的不同帧中心。
透明留白会使可见主体小于碰撞盒；美术须提供尺寸表及叠加碰撞盒的审核截图。

**障碍尺寸的限制与选择：**现有世界独立随机 W、H，均为连续的 30 至 60。
已查到的 `size/setSize` 只能等比缩放，不能用一个方形造型忠实实现任意 W/H。
首轮建议制作 W、H 各取 `{30,40,50,60}` 的 16 个真实障碍造型，
共用尺寸表选择造型并设置逻辑 W/H，s=1；造型选择语法先在平台验证。
这是连续随机尺寸改为离散尺寸的玩法变更，须 autoops-dev 批准，不能声称完全等价。
若必须保持原尺寸分布，则先验证网页是否支持独立宽高变换及锚点同步；
未找到确凿 API 前不编造 `setWidth/setHeight`，也不让图片与碰撞各用不同尺寸。

严格碰撞仍在逻辑坐标中计算：
`160 < X+W && 200 > X && pY < H && pY+64 > 0`。
仅接触边缘不碰撞。不要用 `touching` 或 `onTouchStart` 替换此规则；
教程虽有接触事件，`sprite_query.go` 委托引擎查询，不证明其等价于本项目 AABB。

## 跳跃、障碍与重开

- 状态保持 Ready / Running / GameOver。首次空格仅开始，Running 且落地才起跳：
  `vy=600`；每小步 `pY += vy*dt - 0.5*1800*dt*dt`，`vy -= 1800*dt`。
  下降到 `pY<=0` 时归零并落地，空中拒绝再跳；结束后画面与逻辑冻结。
- 世界速度从 300 以每秒 10 增至 550；按现有世界算法积分位移。
  首障碍左边 X=960；下一左边为前障碍右边加
  `max(speed*1.5, 550*(2*600/1800+0.35))+40`，到 X<=960 时生成。
- 舞台串行发布生成记录；仅原件响应生成消息并 `clone`，克隆忽略该消息；
  `onCloned` 立即保存
  ID 与本局编号、完成造型和位置初始化后 `show`。初始化不等待下一帧，
  不通过多个克隆竞争修改共享临时字段；该协议的网页调度必须实测。
- 当逻辑 `X+W<=0` 时移除记录并让对应克隆 `destroy`，而非 `touching(Edge)`：
  在右侧刚出生就接触边缘不能算回收。首轮采用销毁回收，不预设对象池。
- R 先用短消息屏障清掉本局克隆，再重置记录、玩家、地面偏移、距离及分数，
  回到 Ready，保留最高分。本局编号防止迟到初始化串入新局，原件始终隐藏。
  随机序列算法与网页支持待验；不能承诺与 Go seed=42 的序列一致。

## 背景与得分 UI

背景使用经授权的最终场景素材或编辑器手绘造型，按 960×540 制作，
避免依赖 `stretchMode` 导致地平线比例变化；本轮不生成图片或占位资源。
地面两块各宽 960，相距 960，中心 X 为 `-offset`、`960-offset`，
offset 为非负距离模 960。相同采样、边缘无透明缝；地面上缘固定 -170。
建议层次为背景、地面、障碍、玩家、UI，具体 zorder 与克隆层级在网页验收。

舞台公开 `score/best/status`，监视器按 `target: ""`、`getVar:<字段名>` 绑定。
分数为非负 `distance/10` 向下取整，结束时更新 best，重开不清 best；
最高分只保留当前运行会话，不承诺刷新后持久化。
按本地中心坐标先试 Score=(-450,240)、Best=(-250,240)、Controls=(-450,205)，
不要把现有配置的 (20,20) 当成左上角像素。网页 UI 的锚点、字体与缩放待调。
`monitor.go:37,115-139` 的本地普通刷新间隔为 0.2 秒，不要求 UI 与物理同帧刷新。

## 下一轮验收

由 autoops-dev 人工核对，未完成前不得标记“网页可运行”：

- 验证目标平台版本、资源绑定、按键、帧 API、造型选择与 monitor 变量访问；
  最小网页项目不依赖本地 Go 包，正常显示一个玩家和一个障碍。
- 叠加检查玩家 40×64、中心 X=-300、脚底 -170；各障碍 W/H、
  左边 `X-480` 与顶边一致，动画换帧不漂移，缩放窗口不改变世界碰撞范围。
- 验证边缘仅接触不结束、实际重叠结束、跳过与下降撞击、禁止连跳；
  最高速、低帧率和失焦恢复后仍无明显穿透，小步不等于连续碰撞保证。
- 验证首障碍不会出生即销毁、移出左侧后回收、反复 R 无残留克隆、
  Ready/GameOver 冻结、得分和最高分正确、地面循环无缝、UI 不被遮挡。
- 保存实际网页运行截图/视频、尺寸表及审核记录；本文无渲染验证证据。
