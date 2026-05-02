# Lumina Widget System Design

> **当前架构：Lua-First（零 Go Widget）**。Go 提供布局/绘制/事件/焦点等**机制**，Lua（lux）提供所有 UI 组件。`pkg/widget/` 包已删除，`widget.All()` 返回 nil 的过渡期已结束。

## 架构总览

```
┌─────────────────────────────────────────────────────┐
│  Lua 用户脚本                                        │
│  lumina.createElement + require("lux") / "theme"    │
├─────────────────────────────────────────────────────┤
│  Lua Lux（内嵌 preload，无随盘 lua/ 亦可 require）   │
│  Button, Checkbox, Radio, Switch, Dialog, Toast,    │
│  List, Pagination, Card, Badge, Divider, Progress,  │
│  Form, Tree, Spinner, Layout, CommandPalette, …     │
│  （高级多列表格设计稿：lua/lux/data_grid.md）        │
├─────────────────────────────────────────────────────┤
│  引擎层 (pkg/render/)                                │
│  基元: box/vbox/hbox, text, input, textarea          │
│  事件: onClick, onKeyDown, onScroll, onMouse* …     │
│  布局: flexbox, absolute, overflow=scroll            │
│  绘制: 增量脏区、宽字符                              │
│  主题: Theme struct, ThemeToMap, SetThemeByName      │
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
| Radix UI / shadcn     | Lux（`require("lux")`） | 交互行为、样式、槽位、业务拼装 |
| Tailwind / tokens     | `lumina.getTheme()`、`require("theme")` | 与 `render.CurrentTheme` 对齐的色板 |

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

### 1.2 Layer 与 `style.zIndex`（叠放职责）

二者**不是**「同一机制的两个名字」：**Layer 解决跨树根的全局叠层，`zIndex` 是单棵树 `Style` 上的字段，当前语义有限。**

| 机制 | 所在位置 | 作用 |
|------|-----------|------|
| **Layer** | `pkg/render/engine.go`（`CreateLayer` / `RemoveLayer` / `BringToFront`）、`pkg/render/layer.go` | 引擎维护**多棵根节点树**组成的栈；**靠后的层整层** layout / paint 在更上，事件 **hit-test 自上而下**按层遍历（见 `events.go` 的 `hitTestLayers`）。用于整块 UI 叠在主应用之上：模态、遮罩、overlay 根等。 |
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

## Part 2: 组件系统（Lua-First）

### 2.1 所有 UI 组件均为 Lua Lux

Go 不再提供任何 UI widget。所有组件在 `lua/lux/` 中以纯 Lua 实现，使用 `lumina.createElement("vbox"/"hbox"/"text"/...)` 构建。

当前 umbrella `require("lux")` 暴露的子模块包括（与 `lua/lux/init.lua` 一致）：

| 模块 | 说明 |
|------|------|
| `lux.Button` | 按钮（hover/press 态、variant） |
| `lux.Checkbox` | 复选框 |
| `lux.Radio` | 单选按钮 |
| `lux.Switch` | 开关 |
| `lux.Dialog` | 模态对话框（槽位 API） |
| `lux.Toast` | 轻提示 |
| `lux.List` | 可滚动列表 |
| `lux.Pagination` | 分页 |
| `lux.Card` | 圆角边框容器 + 可选 `title` |
| `lux.Badge` | 彩色标签 |
| `lux.Divider` | 水平分割线 |
| `lux.Progress` | 进度条 |
| `lux.Spinner` | 帧动画加载指示器 |
| `lux.Form` | 表单布局 |
| `lux.Tree` | 树形视图 |
| `lux.Layout` | 布局辅助 |
| `lux.CommandPalette` | 命令面板模板 |

主题：`require("theme")` 与 `lumina.getTheme()` / `render.CurrentTheme` 对齐。

### 2.2 主题系统

Theme 定义在 `pkg/render/theme.go`：

```go
type Theme struct {
    Base, Surface0, Surface1, Surface2 string
    Text, Muted                        string
    Primary, PrimaryDark, Hover, Pressed string
    Success, Warning, Error            string
}
```

内置主题：`DefaultTheme`（Catppuccin Mocha）、`LatteTheme`、`NordTheme`、`DraculaTheme`。

Lua API：`lumina.getTheme()` → 色板表；`lumina.setTheme(name)` 切换内置主题或传入自定义表。

### 2.3 Lua 使用方式

```lua
-- Lux 组件
local lux = require("lux")

lumina.createElement(lux.Button, {
    label = "Submit",
    variant = "primary",
    onClick = function() print("clicked") end,
})

lumina.createElement(lux.Checkbox, {
    checked = true,
    label = "Remember me",
    onChange = function(checked) print(checked) end,
})

-- Lux Card：子节点写在第 3 参数起
lumina.createElement(lux.Card, {
    title = "User Info",
},
    lumina.createElement("text", {}, "Name:"),
    lumina.createElement("input", {placeholder = "Enter name"}),
)
```

### 2.4 已删除的 Go Widget（迁移对照）

| 旧 API | 新 API |
|--------|--------|
| `lumina.Button` | `require("lux.button")` |
| `lumina.Checkbox` | `require("lux.checkbox")` |
| `lumina.Switch` | `require("lux.switch")` |
| `lumina.Radio` | `require("lux.radio")` |
| `lumina.Dialog` | `require("lux.dialog")` |
| `lumina.Toast` | `require("lux.toast")` |
| `lumina.List` | `require("lux.list")` |
| `lumina.Pagination` | `require("lux.pagination")` |
| `lumina.Label` | `lumina.createElement("text", {...}, "content")` |
| `lumina.Spacer` | `lumina.createElement("box", {style = {flex = 1}})` |
| `lumina.Select` / `lumina.Dropdown` | 用 vbox/text + state/layer 组合 |
| `lumina.Table` | 用 hbox/vbox/text 行组合 |
| `lumina.ScrollView` | 容器上设置 `overflow = "scroll"` |

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

| 阶段 | 规划要点 | 当前状态 |
|------|-----------|----------|
| Phase 0 | 引擎事件：`mousedown`/`mouseup`、`focus`/`blur`、`onSubmit`、`onOutsideClick`、`Focusable`/`Disabled` | **已合入** `pkg/render/`（见 §1） |
| Phase 1-3 | Go Radix 控件（Button, Checkbox, Switch, Radio, Select, Label, Dialog, Toast, Table, List, Pagination, Menu, Dropdown, Spacer） | **已删除** — 所有控件迁移至 Lua lux |
| Phase 4 | Lua Lux + 主题 | **已完成**；Lux 源码在 **`lua/lux/`**（`embed.go` 嵌入），`pkg/lux_modules.go` 负责 preload 注册 |
| Phase 5 | Go Widget 基础设施清理 | **已完成** — `pkg/widget/` 包已删除，`WidgetEvent`、`WidgetDef`、`widgets` map、`widgetStates`、`capturedComp`、`RegisterWidget` 等全部移除 |
| Phase 6 | 主题系统迁移 | **已完成** — Theme 定义移至 `pkg/render/theme.go` |

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
