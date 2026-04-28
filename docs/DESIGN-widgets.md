# Lumina Widget System Design

> Go 提供高性能原生组件内核（Radix），Lua 提供可复制、可定制、可热更的组件模板和设计系统（lux）。

## 架构总览

```
┌─────────────────────────────────────────────────────┐
│  Lua 用户代码 / lua/lux/ 组件模板                   │
│  Card, Badge, Alert, Tabs, Progress, Spinner, ...   │
├─────────────────────────────────────────────────────┤
│  Go Widget 层 (pkg/widget/)                          │
│  Button, Checkbox, Radio, Switch, Select,           │
│  Label, Form, Dialog, Menu, Dropdown, Tooltip, ...  │
│  每个 Widget: Render() → *Node, OnEvent(), NewState()│
├─────────────────────────────────────────────────────┤
│  引擎层 (pkg/render/)                                │
│  基元: box/vbox/hbox, text, input, textarea         │
│  事件: 鼠标/键盘/焦点/滚动/表单/覆盖层               │
│  布局: flexbox, absolute, overflow=scroll           │
│  绘制: 增量脏区绘制, CJK 宽字符                      │
├─────────────────────────────────────────────────────┤
│  输出层 (pkg/output/)                                │
│  TUI (ANSI) / WebSocket                             │
└─────────────────────────────────────────────────────┘
```

### 层次映射

| Web 世界              | Lumina 对应         | 职责                     |
|-----------------------|---------------------|--------------------------|
| 浏览器引擎（DOM+渲染） | pkg/render/ Engine  | 节点树、布局、绘制、事件   |
| HTML 原生标签          | 基元 box/text/input | 最小渲染单元，Go 硬编码   |
| Radix UI（无障碍原语） | pkg/widget/         | 交互行为、焦点、状态机     |
| lux/ui（组件模板）  | lua/lux/         | 样式组合、主题、业务定制   |
| Tailwind CSS          | lua/theme/          | 颜色、间距、边框预设      |

---

## Part 1: 事件系统

### 1.1 当前状态

已实现 6 个事件：

| 事件          | Node 字段       | 触发时机                    |
|---------------|-----------------|----------------------------|
| onClick       | OnClick         | 鼠标点击（mousedown 映射）   |
| onMouseEnter  | OnMouseEnter    | 鼠标进入节点区域             |
| onMouseLeave  | OnMouseLeave    | 鼠标离开节点区域             |
| onKeyDown     | OnKeyDown       | 键盘按下                    |
| onChange      | OnChange        | input/textarea 内容变化      |
| onScroll      | OnScroll        | 鼠标滚轮                    |

### 1.2 新增事件

#### 鼠标事件：onMouseDown + onMouseUp

**目的**：Button pressed 态、拖拽。当前 mousedown 直接映射为 onClick，丢失了 press/release 语义。

**修改点**：

1. `pkg/render/node.go` — 新增字段：
```go
OnMouseDown LuaRef
OnMouseUp   LuaRef
```

2. `pkg/render/descriptor.go` — 新增字段（同上）

3. `pkg/render/engine.go` → `readDescriptor` — 新增读取：
```go
desc.OnMouseDown = getRefField(L, absIdx, "onMouseDown")
desc.OnMouseUp   = getRefField(L, absIdx, "onMouseUp")
```

4. `pkg/render/reconciler.go` → `reconcileImpl` — 新增 updateRef

5. `pkg/render/events.go` — 新增：
```go
func (e *Engine) HandleMouseDown(x, y int)  // hit-test → 触发 onMouseDown
func (e *Engine) HandleMouseUp(x, y int)    // hit-test → 触发 onMouseUp
```

6. `pkg/app.go` → `HandleEvent` — 拆分 mousedown/mouseup：
```go
case "mousedown":
    a.engine.HandleMouseDown(e.X, e.Y)  // 触发 onMouseDown
    a.engine.HandleClick(e.X, e.Y)      // 保持 onClick 兼容
case "mouseup":
    a.engine.HandleMouseUp(e.X, e.Y)
```

**Lua API**：
```lua
lumina.createElement("box", {
    onMouseDown = function(e) ... end,
    onMouseUp   = function(e) ... end,
})
```

#### 焦点事件：onFocus + onBlur

**目的**：input 获焦高亮、Select 获焦准备键盘交互、失焦验证。

**修改点**：

1. `pkg/render/node.go` — 新增字段：
```go
OnFocus LuaRef
OnBlur  LuaRef
```

2. `pkg/render/node.go` — 新增 Focusable 字段：
```go
Focusable bool  // 替代 hardcoded type == "input" || type == "textarea"
Disabled  bool  // 禁用状态（事件屏蔽 + 视觉）
```

3. `pkg/render/input.go` — 修改 `collectFocusable`：
```go
func collectFocusable(node *Node) []*Node {
    if node.Focusable && !node.Disabled {
        result = append(result, node)
    }
    // ...
}
```

4. `pkg/render/input.go` — 焦点切换时触发事件：
```go
func (e *Engine) setFocus(newNode *Node) {
    old := e.focusedNode
    if old == newNode { return }

    // Blur old
    if old != nil && !old.Removed && old.OnBlur != 0 {
        e.callLuaRefSimple(old.OnBlur)
    }

    e.focusedNode = newNode

    // Focus new
    if newNode != nil && newNode.OnFocus != 0 {
        e.callLuaRefSimple(newNode.OnFocus)
    }

    // Paint dirty
    if old != nil && !old.Removed { old.PaintDirty = true }
    if newNode != nil { newNode.PaintDirty = true }
    e.needsRender = true
}
```

5. `FocusNext()`, `HandleClick` 中的焦点切换改为调用 `setFocus()`

6. `readDescriptor` 读取 focusable / disabled / onFocus / onBlur

**Lua API**：
```lua
lumina.createElement("box", {
    focusable = true,
    onFocus = function() ... end,
    onBlur  = function() ... end,
})
```

#### 表单事件：onSubmit

**目的**：Form 提交。

**修改点**：

1. `pkg/render/node.go` — 新增：`OnSubmit LuaRef`
2. `pkg/render/descriptor.go` — 同上
3. `readDescriptor` / `reconciler` — 同上
4. `input.go` — Enter 键在 input 上时，向上冒泡查找 OnSubmit 处理器

**Lua API**：
```lua
lumina.createElement("box", {
    onSubmit = function(values) ... end,
},
    lumina.createElement("input", {id = "name"}),
    lumina.createElement("input", {id = "email"}),
)
```

#### 覆盖层事件：onOutsideClick

**目的**：Select/Menu/Dialog 的「点击外部关闭」。

**设计**：作为 Overlay 系统的一部分实现。

1. `pkg/render/node.go` — 新增：`OnOutsideClick LuaRef`
2. `pkg/render/engine.go` — 维护 overlay 栈
3. HandleClick 时：如果有 overlay 激活，检查点击是否在 overlay 节点树内。不在则触发 onOutsideClick。

**Lua API**：
```lua
lumina.createElement("box", {
    style = {position = "absolute", zIndex = 100},
    onOutsideClick = function() setOpen(false) end,
})
```

### 1.3 不支持的事件

| 事件    | 原因                                  |
|---------|---------------------------------------|
| onKeyUp | 终端协议（VT100）无按键释放事件        |
| onOpen  | 业务语义，Lua 自己管理 open 状态       |
| onClose | 业务语义，Lua 自己管理 open 状态       |

### 1.4 完整事件清单（实施后）

```
鼠标:    onClick, onMouseDown, onMouseUp, onMouseEnter, onMouseLeave
键盘:    onKeyDown
焦点:    onFocus, onBlur
输入:    onChange
滚动:    onScroll
表单:    onSubmit
覆盖层:  onOutsideClick
```

共 12 个事件。

---

## Part 2: Widget 系统（Radix 控件）

### 2.1 Widget 接口

```go
// pkg/widget/widget.go

package widget

import "github.com/akzj/lumina/pkg/render"

// Widget 定义一个内置控件的行为。
// 每个 Widget 是一个 Go 实现的 defineComponent：
// - 有自己的状态（每个实例独立）
// - 渲染时组合引擎基元（box/text/input）返回 *Node 树
// - 处理事件时更新内部状态
type Widget struct {
    Name     string
    Render   func(props map[string]any, state any) *render.Node
    OnEvent  func(props map[string]any, state any, event *render.WidgetEvent)
    NewState func() any
}

// WidgetEvent 是 Widget 接收的事件。
type WidgetEvent struct {
    Type string // "click", "mousedown", "mouseup", "keydown", "focus", "blur", ...
    Key  string // 键盘事件的按键
    X, Y int    // 鼠标事件的坐标
}
```

### 2.2 状态设计（每个 Widget 独立 struct）

```go
// widget/button.go
type ButtonState struct {
    Hovered bool
    Pressed bool
}

// widget/checkbox.go
type CheckboxState struct {
    Checked bool
}

// widget/radio.go
type RadioState struct {
    Checked bool
}

// widget/switch_widget.go
type SwitchState struct {
    Checked bool
}

// widget/select_widget.go
type SelectState struct {
    Open        bool
    Selected    int  // 当前选中索引
    Highlighted int  // 键盘高亮索引
}

// widget/dialog.go
type DialogState struct {
    Open bool
}
```

### 2.3 控件分类

#### Go Widget 层实现（pkg/widget/ — 有交互状态机）

| 控件     | 状态              | 关键交互                        |
|----------|-------------------|---------------------------------|
| Button   | hovered, pressed  | hover 变色, press 态, click 触发 |
| Checkbox | checked           | 点击/Space 切换, onChange        |
| Radio    | checked           | 点击/Space 选中, 组互斥, onChange |
| Switch   | checked           | 点击/Space 切换, onChange        |
| Select   | open, selected, highlighted | 展开/收起, ↑↓选择, Enter确认, Escape取消 |
| Label    | —                 | 点击聚焦关联控件                 |
| Form     | —                 | 收集子控件值, onSubmit           |

#### Lua 组件层实现（lua/lux/ — 纯组合 + 样式）

| 组件       | 组合方式                                    |
|------------|---------------------------------------------|
| Card       | box[border=rounded] + text[title] + children |
| Badge      | text[bold, colored]                          |
| Alert      | box[border] + text[icon] + text[message]     |
| Avatar     | box[border=rounded, w=3, h=1] + text[initials] |
| Divider    | text["─" × width]                           |
| Spacer     | box[flex=1]                                  |
| Tabs       | hbox[tab headers] + useState + conditional children |
| Progress   | hbox[text["████░░░░"] + text["60%"]]         |
| Spinner    | text + setInterval[cycle "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"]  |
| Toast      | overlay + setTimeout auto-dismiss            |
| Pagination | hbox[text[页码] + onClick]                   |
| ScrollView | box[overflow="scroll"] （已有基元）           |
| Table      | vbox[hbox[header cells] + vbox[row cells]]   |
| List       | vbox[children]                               |

#### 需要 Overlay 系统后实现（Go Widget）

| 控件     | 依赖                              |
|----------|-----------------------------------|
| Dialog   | overlay + focus trap + onOutsideClick |
| Tooltip  | overlay + 锚定定位 + 延迟显隐      |
| Popover  | overlay + 锚定定位 + onOutsideClick |
| Menu     | overlay + 键盘导航 + onOutsideClick |
| Dropdown | overlay + 键盘导航 + onOutsideClick |

### 2.4 Widget 注册到引擎

```go
// pkg/widget/register.go

func RegisterAll(e *render.Engine) {
    e.RegisterWidget(Button)
    e.RegisterWidget(Checkbox)
    e.RegisterWidget(Radio)
    e.RegisterWidget(Switch)
    e.RegisterWidget(Select)
    e.RegisterWidget(Label)
    e.RegisterWidget(Form)
}
```

```go
// pkg/render/engine.go — 新增

func (e *Engine) RegisterWidget(w *widget.Widget) {
    // 将 Widget 注册为 Lua 可调用的 defineComponent
    // lumina.Button, lumina.Checkbox, ... 自动可用
}
```

### 2.5 Lua 使用方式

```lua
-- 使用 Go 内置 Widget
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

-- 使用 Lua lux 组件
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
  Username:          (点击聚焦关联 input)

Progress:
  ████████░░░░░░░ 60%

Spinner:
  ⠋ Loading...       (每 80ms 切换: ⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏)
```

---

## 实施路线

### Phase 0: 引擎事件扩展

**目标**：为 Widget 系统提供底层事件支持。

**修改文件**：
- `pkg/render/node.go` — 新增字段: Focusable, Disabled, OnMouseDown, OnMouseUp, OnFocus, OnBlur, OnSubmit
- `pkg/render/descriptor.go` — 同步新增字段
- `pkg/render/engine.go` — readDescriptor 读取新字段, readStyleFields 读取 focusable/disabled
- `pkg/render/reconciler.go` — updateRef 新事件
- `pkg/render/events.go` — HandleMouseDown, HandleMouseUp, hasHandler 扩展
- `pkg/render/input.go` — setFocus() 提取, collectFocusable 改用 Focusable 字段, input 兼容设置 Focusable=true
- `pkg/app.go` — HandleEvent 拆分 mousedown/mouseup

**Input CJK 修复**：
- `pkg/render/input.go` — 将 `len(key)==1 && key[0]>=0x20 && key[0]<=0x7E` 改为 `utf8.RuneCountInString(key)==1 && unicode.IsPrint(firstRune)`

### Phase 1: Widget 基础设施 + Button

**目标**：建立 Widget 注册机制，用 Button 验证整个流程。

**新增文件**：
- `pkg/widget/widget.go` — Widget 接口定义
- `pkg/widget/button.go` — Button 实现
- `pkg/widget/register.go` — RegisterAll

**修改文件**：
- `pkg/render/engine.go` — RegisterWidget 方法

### Phase 2: Checkbox + Switch + Radio

**新增文件**：
- `pkg/widget/checkbox.go`
- `pkg/widget/switch_widget.go`（避免 Go 关键字）
- `pkg/widget/radio.go`

### Phase 3: Select + Label + Form

**新增文件**：
- `pkg/widget/select_widget.go`（避免 Go 关键字）
- `pkg/widget/label.go`
- `pkg/widget/form.go`

### Phase 4: Lua lux 组件 + 主题

**新增文件**：
- `lua/theme/init.lua`
- `lua/theme/catppuccin.lua`
- `lua/lux/init.lua`
- `lua/lux/card.lua`
- `lua/lux/badge.lua`
- `lua/lux/alert.lua`
- `lua/lux/divider.lua`
- `lua/lux/tabs.lua`
- `lua/lux/progress.lua`
- `lua/lux/spinner.lua`

### Phase 5: Overlay 系统 + 覆盖层控件

**修改文件**：
- `pkg/render/engine.go` — overlay 栈管理
- `pkg/render/painter.go` — zIndex 排序绘制
- `pkg/render/events.go` — onOutsideClick 分发

**新增文件**：
- `pkg/widget/dialog.go`
- `pkg/widget/tooltip.go`
- `pkg/widget/popover.go`
- `pkg/widget/menu.go`
- `pkg/widget/dropdown.go`
- `lua/lux/toast.lua`
