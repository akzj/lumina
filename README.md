# Lumina — 终端 UI 框架

> Go 渲染引擎 + Lua 声明式 UI = 高性能终端应用

Lumina 是一个 **React 风格的终端 UI 框架**。用 Lua 声明 UI 组件，Go 负责渲染、布局和事件处理。

```lua
lumina.createComponent({
    id = "hello",
    render = function(props)
        local count, setCount = lumina.useState("count", 0)
        return lumina.createElement("box", {
            onClick = function() setCount(count + 1) end,
        },
            lumina.createElement("text", {foreground = "#89B4FA"}, "Count: " .. count)
        )
    end,
})
```

## ✨ 特性

### 渲染引擎
- **持久化节点树** — 组件渲染输出直接 patch 到节点树，无 Virtual DOM 中间层
- **增量布局** — 只重算 `LayoutDirty` 子树
- **脏区绘制** — 只重绘 `PaintDirty` 节点
- **O(k) 复杂度** — k = 实际变化量，与总节点数无关

### 组件系统
- **`createComponent`** — 创建根组件
- **`defineComponent`** — 定义可复用子组件（工厂模式）
- **`createElement`** — 创建 UI 元素（JSX 等价物）
- **`useState`** — 组件状态管理（React 风格 Hook）
- **Hooks** — `useEffect`、`useRef`、`useMemo`、`useCallback`（与 `useState` 同一套调用顺序规则）
- **Radix 风格控件（Go Widget）** — 在 `pkg/widget/` 实现、以 `lumina.<Name>` 暴露给 Lua（与 Web 里 Radix 那层「无障碍原语 + 交互状态机」同角色）；基元仍是 `box` / `text` / `input` 等，Widget 在其之上组合布局与事件
- **Lua Lux** — `require("lux")` 的纯 Lua 模板（Card、Dialog 等），偏展示与业务拼装；复杂交互优先用上面的 Go Widget，再用 Lux 包一层样式即可

### 布局
- **Flexbox** — `vbox`（垂直）/ `hbox`（水平）
- **Flex 分配** — `flex` 属性按比例分配空间
- **对齐** — `justify`（主轴）/ `align`（交叉轴）
- **间距** — `padding`, `margin`, `gap`
- **定位** — `relative`, `absolute`, `fixed`
- **边框** — `single`, `double`, `rounded`
- **约束** — `minWidth`, `maxWidth`, `minHeight`, `maxHeight`

### 事件
- **鼠标** — `onClick`, `onMouseEnter`, `onMouseLeave`, `onScroll`
- **键盘** — `onKeyDown`
- **输入** — `onChange`（input/textarea 值变化）
- **冒泡** — 事件从最深层节点向上冒泡到最近的处理器

### 输入组件
- **`input`** — 单行文本输入
- **`textarea`** — 多行文本输入
- **焦点** — `Tab` 循环、点击聚焦、`autoFocus`
- **编辑** — 光标移动、Backspace、字符插入

### 运行时
- **60fps 事件循环** — 定时渲染脏组件
- **热加载** — 文件变化自动重载 Lua 脚本（`lumina --watch script.lua`）
- **定时器** — `setInterval`, `setTimeout`, `clearInterval`, `clearTimeout`
- **异步** — `lumina.spawn`、`lumina.sleep`、`lumina.readFile`、`lumina.exec`；在 spawn 协程内用 `require("async").await(...)` 等待 Future（示例见 `pkg/testdata/lua_tests/async_test.lua`）
- **开发者工具** — F12 切换，显示组件树和性能指标

---

## 🚀 快速开始

### 安装

```bash
go install github.com/akzj/lumina/cmd/lumina@latest
```

### 运行示例

```bash
# 计数器
lumina examples/counter.lua

# Todo MVC
lumina examples/todo_mvc.lua

# 压力测试（1840 个独立组件）
lumina examples/stress_test.lua

# 系统仪表盘
lumina examples/dashboard.lua

# Widget + Lux 组件展示
lumina examples/components_showcase.lua
```

### CLI 常用参数

```text
lumina [--web :8080] [--mcp :8088] [--watch] <script.lua>
```

- **`--watch`** — 监听脚本所在目录，保存后热重载
- **`--web :端口`** — WebSocket 输出到浏览器（终端里会打印本地 URL）
- **`--mcp :端口`** — 并行启动 MCP HTTP 服务（便于 IDE / 工具对接）

### 退出

`Ctrl+C` 或 `Ctrl+Q`

---

## 📖 Lua API 参考

### lumina.createComponent(config)

创建根组件并注册到渲染引擎。

```lua
lumina.createComponent({
    id = "my-app",          -- 必填，唯一标识
    name = "MyApp",         -- 可选，显示名称
    render = function(props)
        -- 返回 createElement 结果
        return lumina.createElement("box", {}, ...)
    end,
})
```

### lumina.defineComponent(name, renderFn)

定义可复用的子组件工厂。返回一个工厂表，可传给 `createElement`。

```lua
local Button = lumina.defineComponent("Button", function(props)
    local hovered, setHovered = lumina.useState("h", false)
    return lumina.createElement("box", {
        style = {background = hovered and "#313244" or "#1E1E2E"},
        onMouseEnter = function() setHovered(true) end,
        onMouseLeave = function() setHovered(false) end,
        onClick = props.onClick,
    },
        lumina.createElement("text", {foreground = "#89B4FA"}, props.label)
    )
end)

-- 使用子组件
lumina.createElement(Button, {key = "btn1", label = "Click me", onClick = handler})
```

### lumina.createElement(type, props, ...children)

创建 UI 元素描述。

```lua
-- 基本元素
lumina.createElement("box", {style = {background = "#1E1E2E"}},
    lumina.createElement("text", {foreground = "#CDD6F4"}, "Hello")
)

-- 子组件（工厂来自 defineComponent）
lumina.createElement(MyComponent, {key = "unique-key", someProp = "value"})

-- 子组件 + 子节点：第 3 个参数起的子节点会进入 props.children（数组），
-- 便于在 defineComponent 里用 table.unpack(props.children or {}) 组合布局。
lumina.createElement(MyComponent, {title = "Panel"},
    lumina.createElement("text", {}, "Line A"),
    lumina.createElement("text", {}, "Line B")
)
```

**元素类型**:

| 类型 | 说明 |
|------|------|
| `"box"` | 通用容器（默认垂直堆叠） |
| `"vbox"` | 垂直容器 |
| `"hbox"` | 水平容器 |
| `"text"` | 文本节点 |
| `"input"` | 单行文本输入 |
| `"textarea"` | 多行文本输入 |

### Radix 风格控件：Go Widget（`lumina.*`）

应用启动时（`pkg/app.go`）会把 `pkg/widget` 里内置控件全部注册进引擎，因此在 Lua 里与基元一样通过 `createElement` 使用，工厂表挂在全局 `lumina` 上（名称与 Go 侧 `Widget.Name` 一致）。

**当前内置控件（Lua 工厂名）**

| `lumina.*` | 说明 |
|------------|------|
| `Button` | 按钮（variant、hover/pressed 等） |
| `Checkbox` | 勾选框，支持 `checked` + `onChange(bool)` |
| `Switch` | 开关 |
| `Radio` | 单选 |
| `Label` | 标签（可与输入控件关联） |
| `Select` | 下拉选择，`options` + `value` + `onChange(string)` |
| `Dialog` | 对话框容器 |
| `Tooltip` | 提示 |
| `Toast` | 轻提示 |
| `Table` | 表格 |
| `List` | 列表 |
| `Pagination` | 分页 |
| `Menu` | 菜单 |
| `Dropdown` | 下拉菜单 |
| `Spacer` | 占位间距 |

示例：

```lua
lumina.createElement(lumina.Button, {
    label = "OK",
    variant = "primary",
    onClick = function() end,
})

lumina.createElement(lumina.Checkbox, {
    checked = true,
    label = "Remember me",
    onChange = function(checked) end,
})
```

实现细节、事件与无障碍相关约定见 [docs/DESIGN-widgets.md](docs/DESIGN-widgets.md)；源码目录为 [`pkg/widget/`](pkg/widget/)。

### Lua Lux 组件库（`require("lux")`）

与上一节的 **Go Radix 风格控件** 区分：`lux` 是 **Lua 侧可热更的 UI 模板**（由 `pkg/lux_modules.go` 通过 `lua/lux/embed.go` 将 `lua/lux/*.lua` 打进二进制并注册到 `package.preload`，运行时 `require` 即可），不负责底层焦点/键盘路由等——那些由引擎 + Go Widget 处理。

典型用法：

```lua
local lux = require("lux")

lumina.createElement(lux.Card, {title = "Hello"},
    lumina.createElement("text", {}, "Content")
)
```

### lumina.useState(key, defaultValue)

在当前组件中声明一个状态变量。返回 `(currentValue, setterFn)`。

```lua
local count, setCount = lumina.useState("count", 0)
-- 更新状态（触发组件重新渲染）
setCount(count + 1)
```

> **注意**: `key` 在组件内必须唯一。相同 key 的多次调用返回同一个状态。

### lumina.quit()

退出应用。

### lumina.setInterval(fn, ms) / lumina.setTimeout(fn, ms)

设置定时器，返回 timer ID。

```lua
local id = lumina.setInterval(function()
    -- 每 1000ms 执行
end, 1000)

lumina.clearInterval(id)  -- 取消
```

---

## 🎨 样式系统

样式可以通过 `style` 子表或直接作为 props 传入：

```lua
-- 方式 1: style 子表
lumina.createElement("box", {
    style = {width = 40, height = 10, background = "#1E1E2E"},
})

-- 方式 2: 顶层属性（style 子表优先级更高）
lumina.createElement("text", {
    foreground = "#89B4FA",
    bold = true,
})
```

### 尺寸

| 属性 | 说明 |
|------|------|
| `width`, `height` | 固定尺寸（0 = 自动） |
| `minWidth`, `maxWidth` | 宽度约束 |
| `minHeight`, `maxHeight` | 高度约束 |
| `flex` | Flex 增长因子（按比例分配剩余空间） |

### 间距

| 属性 | 说明 |
|------|------|
| `padding` | 四边内边距（简写） |
| `paddingTop/Bottom/Left/Right` | 单边内边距（覆盖简写） |
| `margin` | 四边外边距（简写） |
| `marginTop/Bottom/Left/Right` | 单边外边距（覆盖简写） |
| `gap` | 子元素间距 |

### 对齐

| 属性 | 值 | 说明 |
|------|-----|------|
| `justify` | `"start"`, `"center"`, `"end"`, `"space-between"`, `"space-around"` | 主轴对齐 |
| `align` | `"stretch"`, `"start"`, `"center"`, `"end"` | 交叉轴对齐 |

### 视觉

| 属性 | 说明 |
|------|------|
| `foreground` / `fg` | 前景色（如 `"#89B4FA"`） |
| `background` / `bg` | 背景色 |
| `bold` | 粗体 |
| `dim` | 暗淡 |
| `underline` | 下划线 |
| `border` | 边框样式: `"single"`, `"double"`, `"rounded"` |

### 定位

| 属性 | 说明 |
|------|------|
| `position` | `"relative"`, `"absolute"`, `"fixed"` |
| `top`, `left`, `right`, `bottom` | 偏移量 |
| `zIndex` | 层叠顺序 |

### 溢出

| 属性 | 说明 |
|------|------|
| `overflow` | `"hidden"`, `"scroll"` |
| `scrollY` | 垂直滚动偏移量（配合 `overflow: "scroll"`） |

---

## 🎯 事件系统

```lua
lumina.createElement("box", {
    onClick = function(e)
        -- e.x, e.y: 鼠标位置
    end,
    onMouseEnter = function(e) ... end,
    onMouseLeave = function(e) ... end,
    onKeyDown = function(e)
        -- e.key: 按键名（"a", "Enter", "ArrowUp", ...）
    end,
    onScroll = function(e)
        -- e.delta: 滚动方向（-1=上, 1=下）
        -- e.key: "up" 或 "down"
    end,
    onChange = function(value)
        -- input/textarea 值变化时触发
    end,
})
```

事件从最深层节点向上**冒泡**，直到找到对应的处理器。

---

## 🏗️ 架构概览

```
Lua 用户代码（含 require("lux") / require("theme")）
  ↓ createComponent / defineComponent / createElement / hooks
Engine (Go) + Widget（pkg/widget）
  ↓ renderInOrder()     — 调用脏组件的 Lua renderFn
  ↓ readDescriptor()    — Lua 表 → Descriptor
  ↓ Reconcile()         — Descriptor vs Node 树，就地 patch
  ↓ graftChildComponents() — 嫁接子组件到父树
  ↓ LayoutIncremental() — 只重算脏子树
  ↓ PaintDirty()        — 只重绘脏节点到 CellBuffer
  ↓ ToBuffer()          — CellBuffer → Buffer
Output Adapter
  ↓ WriteDirty(buf, dirtyRects) — 只输出变化区域
终端 / WebSocket（--web）
```

详细架构设计见 [DESIGN.md](DESIGN.md)。

---

## 🔧 开发指南

### 运行测试

```bash
# 全部测试
go test ./pkg/...

# 渲染引擎测试
go test ./pkg/render/...

# 集成测试
go test ./pkg/ -run TestE2E

# 压力测试 benchmark
go test ./pkg/ -bench BenchmarkStress -benchtime 5s

# Lua 测试框架（testdata/lua_tests 下 *_test.lua，含子目录）
go test ./pkg/ -run TestLuaTestFramework
```

### 项目结构

```
cmd/lumina/           — CLI 入口
pkg/                  — 核心框架（package v2）
  render/             — 渲染引擎（Engine, Node, Reconciler, Layout, Painter）
  widget/             — Go 内置 Widget（Button、Checkbox、Select…）
  buffer/             — Buffer 类型
  output/             — 输出适配器（ANSI, TestAdapter）
  event/              — 事件类型
  perf/               — 性能追踪
  devtools/           — 开发者工具
  animation/          — 动画系统
  router/             — 路由
  hotreload/          — 热加载
  store/              — 状态管理
  testdata/lua_tests/ — Lua 侧单元测试脚本
lua/                  — Lux / theme 源码（由 pkg 内嵌到运行时 require）
examples/             — 示例 Lua 应用
docs/                 — 文档
```

### 添加新的元素类型

1. 在 `render/node.go` 中定义类型字符串
2. 在 `render/layout.go` 的 `computeFlex()` 中添加布局分支
3. 在 `render/painter.go` 的 `paintNode()` 中添加绘制分支
4. 在 `render/engine.go` 的 `readDescriptor()` 中读取特有属性
5. 写测试，运行 `go test ./pkg/render/...`

---

## 📄 许可证

MIT
