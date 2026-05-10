# Lumina — 终端 UI 框架

> Go 渲染引擎 + Lua 声明式 UI = 高性能终端应用

Lumina 是一个 **React 风格的终端 UI 框架**。用 Lua 声明 UI 组件，Go 负责渲染、布局和事件处理。专为 AI Agent 调试设计了 MCP DevTools，是生成式 TUI 应用的最佳底座。

```lua
local app = lumina.app({
    id = "hello",
    store = { count = 0 },
    render = function()
        local count = lumina.useStore("count")
        return lumina.createElement("box", {
            onClick = function() lumina.store.set("count", count + 1) end,
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
- **分层渲染** — 支持 main 层 + 多个 overlay 层（窗口、对话框等）
- **2D 滚动支持** — 水平/垂直滚动，自动滚动条

### 组件系统
- **`lumina.app`** — 应用入口（全局配置、状态、路由、快捷键）
- **`defineComponent`** — 定义可复用子组件（工厂模式）
- **`createElement`** — 创建 UI 元素（JSX 等价物）
- **Hooks** — 19个 React 风格 Hook
  - 基础：`useState`、`useEffect`、`useRef`、`useMemo`、`useCallback`
  - 框架：`useStore`、`useRoute`、`useTheme`、`useAnimation`
- **Lux 纯 Lua 组件库** — 30+ shadcn 风格组件，嵌入二进制，无需外置依赖
  - 基础：Button、Card、Badge、Divider、Progress、Spinner、Alert、Accordion、Breadcrumb
  - 输入：TextInput、Checkbox、Radio、Switch、Form、Command Palette
  - 布局：Layout、Slot、Window、SplitPane、ScrollView、VList、DataGrid、Pagination、Tabs、Tree
  - 反馈：Dialog、Toast
  - 高级：WM(窗口管理)、DataGrid、Atlantis 主题

### 布局系统
- **Flexbox** — `vbox`（垂直）/ `hbox`（水平）/ `grid`（网格）
- **Flex 分配** — `flexGrow`/`flexShrink`/`flexBasis` 属性
- **对齐** — `justifyContent`（主轴）/ `alignItems`（交叉轴）/ `alignSelf`
- **间距** — `padding`, `margin`, `gap`，支持简写和单边设置
- **定位** — `static`, `relative`, `absolute`, `fixed`
- **边框** — `single`, `double`, `rounded`，支持自定义颜色
- **约束** — `minWidth`, `maxWidth`, `minHeight`, `maxHeight`
- **单位支持** — 无单位（字符）、px、%、vw、vh
- **溢出处理** — `hidden`/`scroll`/`visible`，自动滚动条

### 事件系统
- **鼠标事件** — `onClick`, `onMouseDown`, `onMouseUp`, `onMouseEnter`, `onMouseLeave`, `onScroll`, `onScrollH`
- **键盘事件** — `onKeyDown`, `onKeyUp`, `onKeyPress`
- **输入事件** — `onChange`（input/textarea 值变化）
- **自定义事件** — `emit`/`on`/`off` 事件总线
- **命中测试** — 多层重叠节点的事件冒泡
- **焦点管理** — Tab/Shift+Tab 循环、点击聚焦、`autoFocus`、全局快捷键
- **文本编辑** — 支持删除、方向键、选择、IME 兼容

### 运行时
- **60fps 事件循环** — 定时渲染脏组件
- **热加载** — 文件变化自动重载 Lua 脚本（`lumina --watch script.lua`）
- **定时器** — `setInterval`, `setTimeout`, `clearInterval`, `clearTimeout`
- **异步支持** — 协程调度、异步IO
- **动画系统** — `useAnimation` Hook，支持缓动函数、循环动画
- **全局状态** — `lumina.store` 全局状态管理，`useStore` 响应式订阅
- **路由系统** — `lumina.router` 路由管理，`useRoute` 响应式订阅
- **开发者工具** — F12 切换，包含 Elements 面板（组件树检查）、性能面板（渲染指标）、实时修改状态

### MCP DevTools（AI 专属）
Lumina 内置 AI 友好的调试协议：
- **inspect** — 读取组件树、组件详情、计算样式
- **simulate** — 模拟点击、按键、滚动等用户操作
- **console** — 日志收集、错误栈输出
- **patch** — 热修复组件代码，无需重启
- **diff** — 帧对比，输出变化区域
- **profile** — 性能分析，渲染耗时统计

### Web 运行时
- WebSocket 服务器
- xterm.js 前端
- 多会话管理
- 浏览器中直接访问 Lumina 应用

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

# 表单演示
lumina examples/form_demo.lua

# 系统仪表盘
lumina examples/dashboard.lua

# 文件浏览器
lumina examples/file_browser.lua

# 组件展示
lumina examples/components_showcase.lua

# 2D 滚动演示
lumina examples/scroll_2d.lua

# 真实案例
lumina examples/ai-agent/main.lua    # AI Agent 界面
lumina examples/kanban/main.lua     # 看板应用
lumina examples/api-client/main.lua # API 客户端
lumina examples/markdown-viewer/main.lua # Markdown 阅读器
```

### CLI 常用参数
```text
lumina [--web :8080] [--mcp :8088] [--watch] <script.lua>
```
- **`--watch`** — 监听脚本所在目录，保存后热重载
- **`--web :端口`** — WebSocket 输出到浏览器（终端里会打印本地 URL）
- **`--mcp :端口`** — 并行启动 MCP HTTP 服务（便于 IDE / AI Agent 对接）

### 退出
`Ctrl+C` 或 `Ctrl+Q`

---

## 📖 Lua API 参考

### lumina.app(config)
创建并启动应用，是 V2 的标准入口。
```lua
lumina.app({
    id = "my-app",          -- 必填，唯一标识
    name = "MyApp",         -- 可选，显示名称
    store = { count = 0 },  -- 初始全局状态
    routes = { "/", "/settings/:id" }, -- 路由表
    keys = {                -- 全局快捷键
        ["Ctrl+C"] = function() lumina.quit() end,
        ["F12"] = function() -- 自定义快捷键 end
    },
    render = function()
        -- 返回 createElement 结果
        return lumina.createElement("box", {}, ...)
    end,
})
```

### lumina.defineComponent(config)
定义可复用的子组件工厂。返回一个工厂表，可传给 `createElement`。
```lua
local Button = lumina.defineComponent({
    name = "Button",
    init = function(props)
        return { hovered = false }
    end,
    render = function(instance)
        return lumina.createElement("box", {
            style = {background = instance.hovered and "#313244" or "#1E1E2E"},
            onMouseEnter = function() instance.setState({hovered = true}) end,
            onMouseLeave = function() instance.setState({hovered = false}) end,
            onClick = instance.props.onClick,
        },
            lumina.createElement("text", {foreground = "#89B4FA"}, instance.props.label)
        )
    end
})

-- 使用子组件
lumina.createElement(Button, {key = "btn1", label = "Click me", onClick = handler})
```

### lumina.createElement(type, props, ...children)
创建 UI 元素描述。
```lua
-- 基本元素
lumina.createElement("box", {style = {background = "#1E1E2E"}},\n    lumina.createElement("text", {foreground = "#CDD6F4"}, "Hello")
)

-- 子组件（工厂来自 defineComponent）
lumina.createElement(MyComponent, {key = "unique-key", someProp = "value"})

-- 子组件 + 子节点：第 3 个参数起的子节点会进入 props.children（数组），
-- 便于在 defineComponent 里用 table.unpack(props.children or {}) 组合布局。
lumina.createElement(MyComponent, {title = "Panel"},\n    lumina.createElement("text", {}, "Line A"),\n    lumina.createElement("text", {}, "Line B")
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
| `"fragment"` | 透明容器，不占空间 |
| `"component"` | 组件占位符 |

### Lux 组件库（`require("lux")`）
30+ 纯 Lua 实现的 shadcn 风格组件，嵌入二进制，无需外置依赖：
```lua
local lux = require("lux")

-- 按钮
lumina.createElement(lux.Button, {
    label = "OK",
    variant = "primary",
    onClick = function() end,
})

-- 复选框
lumina.createElement(lux.Checkbox, {
    checked = true,
    label = "Remember me",
    onChange = function(checked) end,
})

-- 卡片
lumina.createElement(lux.Card, {title = "Hello"},
    lumina.createElement("text", {}, "Content")
)

-- 下拉选择
lumina.createElement(lux.Select, {
    options = {"A", "B", "C"},
    value = "A",
    onChange = function(value) end,
})
```
完整组件列表和用法见 [docs/COMPONENTS.md](docs/COMPONENTS.md)。

### 状态管理
```lua
-- 读取全局状态并订阅变化
local count = lumina.useStore("count")

-- 修改全局状态
lumina.store.set("count", count + 1)

-- 批量修改
lumina.store.batch({
    count = 1,
    user = {name = "Alice"}
})
```

### 路由系统
```lua
-- 获取当前路由信息并订阅变化
local route = lumina.useRoute()
-- route = {path = "/settings/123", params = {id = "123"}}

-- 导航
lumina.router.navigate("/settings/456")

-- 返回上一页
lumina.router.back()
```

### 定时器
```lua
local id = lumina.setInterval(function()
    -- 每 1000ms 执行
end, 1000)

lumina.clearInterval(id)  -- 取消
```

### lumina.quit()
退出应用。

---

## 🎨 样式系统
样式可以通过 `style` 子表传入：
```lua
lumina.createElement("box", {
    style = {
        width = "100%",
        height = "100%",
        background = "#1E1E2E",
        border = "single",
        borderColor = "#89B4FA",
        padding = 1,
        gap = 1
    },
})
```

### 尺寸属性
| 属性 | 说明 |
|------|------|
| `width`, `height` | 固定尺寸，支持数字、`%`、`vw`、`vh` |
| `minWidth`, `maxWidth` | 宽度约束 |
| `minHeight`, `maxHeight` | 高度约束 |
| `flexGrow` | 弹性增长因子 |
| `flexShrink` | 弹性收缩因子 |
| `flexBasis` | 弹性基准尺寸 |

### 间距属性
| 属性 | 说明 |
|------|------|
| `padding` | 四边内边距（简写） |
| `paddingTop/Bottom/Left/Right` | 单边内边距（覆盖简写） |
| `margin` | 四边外边距（简写） |
| `marginTop/Bottom/Left/Right` | 单边外边距（覆盖简写） |
| `gap` | 子元素间距 |

### 对齐属性
| 属性 | 值 | 说明 |
|------|-----|------|
| `justifyContent` | `"start"`, `"center"`, `"end"`, `"space-between"`, `"space-around"`, `"space-evenly"` | 主轴对齐 |
| `alignItems` | `"stretch"`, `"start"`, `"center"`, `"end"` | 交叉轴对齐 |
| `alignSelf` | 同 alignItems | 单个子元素交叉轴对齐 |
| `flexDirection` | `"row"`, `"column"`, `"row-reverse"`, `"column-reverse"` | 主轴方向 |
| `flexWrap` | `"nowrap"`, `"wrap"`, `"wrap-reverse"` | 换行 |

### 视觉属性
| 属性 | 说明 |
|------|------|
| `foreground` / `fg` | 前景色（如 `"#89B4FA"` 或颜色名） |
| `background` / `bg` | 背景色 |
| `bold` | 粗体 |
| `dim` | 暗淡 |
| `underline` | 下划线 |
| `border` | 边框样式: `"single"`, `"double"`, `"rounded"` |
| `borderColor` | 边框颜色 |

### 位置属性
| 属性 | 说明 |
|------|------|
| `position` | `"static"`, `"relative"`, `"absolute"`, `"fixed"` |
| `top`, `left`, `right`, `bottom` | 偏移量 |
| `zIndex` | 层叠顺序 |

### 溢出属性
| 属性 | 说明 |
|------|------|
| `overflow` | `"hidden"`, `"scroll"`, `"visible"` |
| `scrollX`/`scrollY` | 滚动偏移量 |

完整布局属性参考见 [docs/LAYOUT_DESIGN.md](docs/LAYOUT_DESIGN.md)。

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
  ↓ lumina.app / defineComponent / createElement / hooks
Render Engine (Go)
  ↓ renderInOrder()     — 调用脏组件的 Lua renderFn
  ↓ readDescriptor()    — Lua 表 → Descriptor
  ↓ Reconcile()         — Descriptor vs Node 树，就地 patch
  ↓ graftChildComponents() — 嫁接子组件到父树
  ↓ LayoutIncremental() — 只重算脏子树
  ↓ PaintDirty()        — 只重绘脏节点到 CellBuffer
  ↓ ToBuffer()          — CellBuffer → Buffer
Output Adapter
  ↓ WriteDirty(buf, dirtyRects) — 只输出变化区域
终端 / WebSocket（--web） / MCP 服务（--mcp）
```
详细架构设计见 [DESIGN.md](DESIGN.md)。

---

## 🔧 开发指南
### 运行测试
```bash
# 全部测试（751+ 测试用例）
go test ./pkg/...

# 渲染引擎测试
go test ./pkg/render/...

# 集成测试
go test ./pkg/ -run TestE2E

# 压力测试 benchmark
go test ./pkg/ -bench BenchmarkStress -benchtime 5s

# Lua 测试框架
go test ./pkg/ -run TestLuaTestFramework
```

### 项目结构
```
cmd/
  lumina/           — CLI 入口
  lumina-server/    — MCP 服务器入口
pkg/                  — 核心框架（package v2）
  render/             — 渲染引擎（Engine, Node, Reconciler, Layout, Painter）
  buffer/             — Buffer 类型
  output/             — 输出适配器（ANSI, TestAdapter, WebSocket）
  event/              — 事件系统、命中测试、焦点管理
  perf/               — 性能追踪
  devtools/           — 开发者工具
  animation/          — 动画系统
  router/             — 路由管理
  hotreload/          — 热加载
  store/              — 全局状态管理
  mcp/                — MCP 协议实现
  testdata/lua_tests/ — Lua 侧单元测试脚本
lua/                  — Lux 组件库 / theme 源码（内嵌到二进制）
examples/             — 示例应用
docs/                 — 文档
```

---

## ⚠️ 已知问题

### macOS Terminal.app 中文输入显示错位

在 macOS 自带的 Terminal.app 中使用中文输入法（IME）时，**选词过程中**界面可能出现短暂的边框错位。选词确认后界面会自动恢复正常。

**原因**：Terminal.app 在 IME 预编辑阶段使用**插入模式**在光标位置绘制拼音文本，将 TUI 已渲染的内容（如边框）向右推移，导致视觉错位。这是终端模拟器的行为，非 Lumina 框架 bug。

**推荐方案**：使用 **iTerm2** 替代 Terminal.app。iTerm2 的 IME 使用覆盖模式，不会影响 TUI 布局。

| 终端 | 中文输入 | 说明 |
|------|---------|------|
| **iTerm2** ✅ | 正常 | IME 覆盖模式，不影响 TUI 渲染 |
| **Terminal.app** ⚠️ | 选词时短暂错位 | IME 插入模式，推移 TUI 内容 |

---

## 📄 许可证
MIT
