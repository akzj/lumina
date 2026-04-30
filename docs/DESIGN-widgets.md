# Lumina Widget System Design

> Go 提供 **Radix 风格**原生控件（`pkg/widget/`，Lua 里为 `lumina.*`）；Lua 提供可复制、可热更的展示模板 **Lux**（源码在 `lua/lux/`，构建时由 `lua/lux/embed.go` 嵌入、`pkg/lux_modules.go` 注册到 `require("lux")`）。

**与仓库同步**：内置控件列表以 [`pkg/widget/register.go`](../pkg/widget/register.go) 的 `widget.All()` 为准；Lux 子模块以 [`pkg/lux_modules.go`](../pkg/lux_modules.go) 的 `registerLuxModules` 为准；应用侧注册见 [`pkg/app.go`](../pkg/app.go) 中 `NewApp` 对 `eng.RegisterWidget` 的循环。

## 架构总览

```
┌─────────────────────────────────────────────────────┐
│  Lua 用户脚本                                        │
│  lumina.createElement + require("lux") / "theme"    │
├─────────────────────────────────────────────────────┤
│  Lua Lux（内嵌 preload，无随盘 lua/ 亦可 require）   │
│  Card, Badge, Divider, Progress, Spinner, Dialog,   │
│  Layout, CommandPalette, ListView, …                │
│  （高级多列表格设计稿：lua/lux/data_grid.md）        │
├─────────────────────────────────────────────────────┤
│  Go Radix 控件层 (pkg/widget/)                       │
│  Button Checkbox Switch Radio Label Select Dialog    │
│  Tooltip Toast Table List Pagination Menu Dropdown   │
│  Spacer …                                            │
│  Render → *Node，OnEvent → 状态 + FireOnChange/Layer │
├─────────────────────────────────────────────────────┤
│  引擎层 (pkg/render/)                                │
│  基元: box/vbox/hbox, text, input, textarea          │
│  事件: 见下文 §1；WidgetEvent 含 Layer 输出          │
│  布局: flexbox, absolute, overflow=scroll            │
│  绘制: 增量脏区、宽字符                              │
├─────────────────────────────────────────────────────┤
│  输出层 (pkg/output/)                                │
│  TUI (ANSI) / WebSocket（CLI --web）                 │
└─────────────────────────────────────────────────────┘
```

### 层次映射

| Web 世界              | Lumina 对应         | 职责                     |
|-----------------------|---------------------|--------------------------|
| 浏览器引擎（DOM+渲染） | pkg/render/ Engine  | 节点树、布局、绘制、事件   |
| HTML 原生标签          | 基元 box/text/input | 最小渲染单元，Go 硬编码   |
| Radix UI（无障碍原语） | pkg/widget/         | 交互行为、焦点、状态机     |
| 组件模板 / shadcn 类   | Lux（`require("lux")`） | 样式组合、槽位、业务拼装（**非** Go 控件替代品） |
| Tailwind / tokens      | `lumina.getTheme()`、`require("theme")` | 与 Go `widget.CurrentTheme` 对齐的色板 |

---

## Part 1: 事件系统（当前实现）

### 1.1 基元 / 描述符上的 Lua 回调

对应 `pkg/render/node.go`、`readDescriptor` 已支持的字段（节选）：

| 回调 | 说明 |
|------|------|
| `onClick` | 点击 |
| `onMouseDown` / `onMouseUp` | 按下 / 抬起（Button pressed 态等） |
| `onMouseEnter` / `onMouseLeave` | 悬停 |
| `onKeyDown` | 键盘 |
| `onChange` | `input` / `textarea` 文本变化；以及引擎在 Widget 设置 `FireOnChange` 后桥接到节点 `onChange` |
| `onScroll` | 滚轮 |
| `onFocus` / `onBlur` | 焦点进出 |
| `onSubmit` | 如 input 内 Enter 向上冒泡到带 `onSubmit` 的容器 |
| `onOutsideClick` | 点击组件区域外（下拉、菜单等关闭语义） |

另有 **`focusable`**、**`disabled`** 控制是否参与 Tab 焦点与是否屏蔽事件。

### 1.2 Widget：`render.WidgetEvent`（`pkg/render/engine.go`）

Go Widget 的 `OnEvent` 收到的结构体除 `Type` / `Key` / `X,Y` 外，还包括：

| 字段 | 方向 | 含义 |
|------|------|------|
| `FireOnChange` | 输出 | Widget 设置后，引擎对节点调用 `onChange`（值由 Widget 决定，如 Checkbox 的 bool） |
| `WidgetX/Y/W/H` | 输入 | 引擎在调用 `OnEvent` **前**填入的控件屏幕包围盒 |
| `CreateLayer` | 输出 | 非 nil 时引擎创建叠加层（`LayerRequest`：`ID`、`Root` 子树、`Modal`） |
| `RemoveLayer` | 输出 | 非空时移除对应层 ID |

### 1.2.1 Layer 与 `style.zIndex`（叠放职责）

二者**不是**「同一机制的两个名字」：**Layer 解决跨树根的全局叠层，`zIndex` 是单棵树 `Style` 上的字段，当前语义有限。**

| 机制 | 所在位置 | 作用 |
|------|-----------|------|
| **Layer** | `pkg/render/engine.go`（`CreateLayer` / `RemoveLayer` / `BringToFront`）、`pkg/render/layer.go` | 引擎维护**多棵根节点树**组成的栈；**靠后的层整层** layout / paint 在更上，事件 **hit-test 自上而下**按层遍历（见 `events.go` 的 `hitTestLayers`）。用于整块 UI 叠在主应用之上：模态、遮罩、Go Widget 通过 `WidgetEvent.CreateLayer` 挂出的 overlay 根等。 |
| **`style.zIndex`** | `pkg/render/node.go` 的 `Style.ZIndex`，Lua `style.zIndex` | 字段**仍存在**并由描述符解析，但**不等价于 Web/CSS 的 z-index**：当前实现**未**在 `layout.go` / `painter.go` 中按数值对 flex 子节点做完整重排；同层叠放仍主要依赖**子节点顺序**以及 **`absolute` / `fixed`** 相关的绘制与脏区逻辑（如 `painter.go` 中对「靠后兄弟」与 overlay 的处理）。`docs/DEVELOPMENT.md` 中将其标为 **「有限使用」**。 |

**选用建议**：需要盖住主界面、独立一层的事件语义（例如 modal 区外点击）→ **优先用 Layer**。仅在同一容器内表达叠放、且能接受当前引擎行为时，再使用 `zIndex`，并以源码为准验证效果。

### 1.3 不支持的回调（终端 / 产品取舍）

| 事件    | 原因                                  |
|---------|---------------------------------------|
| onKeyUp | 常见终端协议无可靠 key-up            |
| onOpen / onClose | 由业务状态（Lua `useState`）表达即可 |

### 1.4 历史：分步落地说明

早期文档中的「新增事件」分步清单已合入主干；字段与分发逻辑以 **`pkg/render/node.go`**、**`descriptor.go`**、**`events.go`**、**`input.go`** 为准，不再在此重复大段伪代码。

---

## Part 2: Widget 系统（Radix 控件）

### 2.1 Widget 接口（`pkg/widget/widget.go`）

```go
type Widget struct {
    Name     string
    Render   func(props map[string]any, state any) any // *render.Node
    OnEvent  func(props map[string]any, state any, event *render.WidgetEvent) bool
    NewState func() any
}
```

- **`Render`**：返回 `*render.Node` 树（在接口里用 `any` 避免与 `render` 包循环引用）。
- **`OnEvent`**：返回 `true` 表示 Widget 消费了事件并可能改了内部状态，需要重绘。
- **`WidgetEvent`**：定义在 **`pkg/render/engine.go`**（见 §1.2），不是 `widget` 包内重复定义。

### 2.2 已注册的 Go Radix 控件（`widget.All()` → `lumina.<Name>`）

下列控件在 [`register.go`](../pkg/widget/register.go) 中导出，并由 `NewApp` 逐个 `RegisterWidget`。**Lua 工厂名与 `Widget.Name` 一致**（如 `lumina.Button`）。

| `lumina.*` | 典型状态 / 职责 |
|------------|-----------------|
| `Button` | `Hovered` / `Pressed`；`variant`、点击 |
| `Checkbox` | `Checked` + `Hovered`；`checked` 受控、`onChange(bool)` |
| `Switch` | 开关态 |
| `Radio` | 单选组 |
| `Label` | 文案 + 关联聚焦 |
| `Select` | `Open` / `Selected` / `Highlighted`；下拉与键盘导航 |
| `Dialog` | `Open`；`open`/`title`/`message` 等 props |
| `Tooltip` | 悬停提示 |
| `Toast` | 轻提示 |
| `Table` | 表格数据展示 |
| `List` | 列表 |
| `Pagination` | 分页控件 |
| `Menu` | 菜单 |
| `Dropdown` | 下拉菜单 |
| `Spacer` | 布局占位 |

> **未实现**：文档早期草稿中的 **`Form`**、**`Popover`** 等 **不在** 当前 `widget.All()`；表单提交可组合 `input` + 容器 `onSubmit`，或由业务 Lua 收集字段。

### 2.3 Lux（Lua 模板，内嵌 preload）

源码目录为 **`lua/lux/`**；构建时由 **`lua/lux/embed.go`** 嵌入，`pkg/lux_modules.go` 的 **`registerLuxModules`** 写入 `package.preload`（与二进制同发，无需随盘携带 `lua/`）。

当前 umbrella **`require("lux")`** 暴露的子模块包括（与 `lua/lux/init.lua` 一致）：

| 模块 | 说明 |
|------|------|
| `lux.Card` | 圆角边框容器 + 可选 `title` + `props.children` |
| `lux.Badge` | 彩色标签 |
| `lux.Divider` | 水平分割线字符 |
| `lux.Progress` | 文本进度条 |
| `lux.Spinner` | `useEffect` + `setInterval` 帧动画 |
| `lux.Dialog` | 槽位 API（`Dialog.Title` / `Content` / `Actions`）的对话框模板（与 Go `lumina.Dialog` 不同层） |
| `lux.Layout` | 布局辅助 |
| `lux.CommandPalette` | 命令面板模板 |

主题：`require("theme")` 与 `lumina.getTheme()` / Go `widget.CurrentTheme` 对齐。

### 2.4 注册到引擎（实际代码路径）

```go
// pkg/app.go（节选）
eng := render.NewEngine(L, w, h)
for _, wgt := range widget.All() {
    eng.RegisterWidget(wgt)
}
eng.RegisterLuaAPI()
```

`RegisterWidget` 在 **`pkg/render/engine.go`**：把每个 `Widget.Name` 注册为 Lua 全局表 `lumina` 上的可调用工厂（与 `defineComponent` 工厂同一套 `createElement` 路径）。

### 2.5 Lua 使用方式

```lua
-- Go Radix 控件
lumina.createElement(lumina.Button, {
    label = "Submit",
    variant = "primary",
    onClick = function() print("clicked") end,
})

lumina.createElement(lumina.Checkbox, {
    checked = true,
    label = "Remember me",
    onChange = function(checked) print(checked) end,
})

lumina.createElement(lumina.Select, {
    options = {
        {label = "Red", value = "red"},
        {label = "Blue", value = "blue"},
    },
    value = "red",
    onChange = function(value) print(value) end,
})

-- Lux：子节点写在第 3 参数起，引擎会注入为 props.children（数组）
local lux = require("lux")

lumina.createElement(lux.Card, {
    title = "User Info",
},
    lumina.createElement(lumina.Label, {text = "Name:"}),
    lumina.createElement("input", {placeholder = "Enter name"}),
)
```

### 2.6 终端外观规范

```
Button:
  ┌──────────┐     ┌──────────┐      Submit
  │ Submit   │     │ Submit   │     (ghost)
  └──────────┘     └──────────┘
  (primary)        (outline)

  ┌──────────┐
  │ Submit   │  ← hovered: 背景变亮
  └──────────┘

Checkbox:
  [ ] Remember me     [x] Remember me     [-] Disabled

Radio:
  ( ) Option A        (●) Option B        (-) Disabled

Switch:
  [○  ] Off           [  ●] On            [--] Disabled

Select (collapsed):
  ┌─────────────┐
  │ Red        ▼│
  └─────────────┘

Select (expanded):
  ┌─────────────┐
  │ Red        ▲│
  ├─────────────┤
  │ Red       ✓ │
  │ Blue      ← │  ← highlighted
  │ Green       │
  └─────────────┘

Label:
  Username:          (纯文本；与 input 的关联在 Lua 里用布局 + onClick 等自行实现)

Progress:
  ████████░░░░░░░ 60%

Spinner:
  ⠋ Loading...       (每 80ms 切换: ⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏)
```

---

## 实施路线（里程碑 ↔ 当前仓库）

下表对照本文早期 **Phase** 规划与**主干现状**，便于查阅；细节以代码为准。

| 阶段 | 规划要点 | 当前状态 |
|------|-----------|----------|
| Phase 0 | 引擎事件：`mousedown`/`mouseup`、`focus`/`blur`、`onSubmit`、`onOutsideClick`、`Focusable`/`Disabled` | **已合入** `pkg/render/`（见 §1） |
| Phase 1 | `RegisterWidget` + Button | **已完成** |
| Phase 2 | Checkbox、Switch、Radio | **已完成** |
| Phase 3 | Select、Label；草稿中的 **Form** | Select、Label **已完成**；**无** `pkg/widget/form.go` |
| Phase 4 | Lua Lux + 主题 | **已完成**；Lux 源码在 **`lua/lux/`**（`embed.go` 嵌入），`pkg/lux_modules.go` 负责 preload 注册 |
| Phase 5 | Overlay、Dialog、Menu、Dropdown… | **部分完成**：上述控件已在 `widget.All()`；**Layer**（`WidgetEvent.CreateLayer` 等）与全局 overlay 策略仍在 **迭代**；**Popover**、独立 **`lua/lux/toast.lua`** 等**未**按原清单落地 |

### 后续工作分期（P0 → P1 → P2）

按 **依赖顺序** 推进（非排期承诺；**Lux DataGrid** 自有分期见 [`lua/lux/data_grid.md`](../lua/lux/data_grid.md) §9）。

**P0 — 叠层与语义收束**

- **Layer**：modal / 点击区外 / `BringToFront` / `RemoveLayer` 与焦点陷阱行为与 **§1.2.1** 文档一致；缺回归处补 **单测或 DevTools 检查清单**。
- **`style.zIndex`**：**二选一**落地——在 `pkg/render/layout.go` / `painter.go` 实现 **可文档化的绘制顺序**，或在 API 中 **正式标注 deprecated / 内部保留**，避免「字段可传却无统一定义」长期悬置。与 **Layer** 分工见 **§1.2.1**。

**P1 — 产品化控件与范式**

- **Form**：新增 `pkg/widget/form.go`，或在 README / [`docs/lux-api.md`](lux-api.md) 给出 **Lua 官方范式**（`onSubmit` + 校验），二选一写入仓库标准。
- **Popover**、**Lux Toast**（与 Go `Toast` 分工）等 Phase 5 项：按需求逐项实现，或从清单 **正式划除** 并注明替代方案。

**P2 — 引擎与复杂布局**

- 全局 **overlay** 策略、多 Layer 与（若有）动画的统一定义与文档。
- **DataGrid** 横向粘性列、更强冻结、展开行等：依赖布局/引擎能力，见 `lua/lux/data_grid.md` §3 / §9。
