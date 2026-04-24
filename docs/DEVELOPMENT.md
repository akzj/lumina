# Lumina 开发指南 v0.1

> **版本**: v0.1 (Phase 1 待实现)  
> **状态**: 开发中  
> **日期**: 2024-04-24  
> **前提**: 设计文档 `DESIGN.md`

---

## 目录

1. [项目概述](#1-项目概述)
2. [快速开始](#2-快速开始)
3. [核心概念](#3-核心概念)
4. [组件开发](#4-组件开发)
5. [Hooks 参考](#5-hooks-参考)
6. [样式系统](#6-样式系统)
7. [事件处理](#7-事件处理)
8. [主题系统](#8-主题系统)
9. [MCP 协议](#9-mcp-协议)
10. [性能优化](#10-性能优化)
11. [测试指南](#11-测试指南)
12. [常见问题](#12-常见问题)

---

## 1. 项目概述

### 1.1 什么是 Lumina

Lumina 是一个面向 AI Agent 的 TUI（终端用户界面）框架，目标是成为"终端版 React"。

```
Web React                      →    Lumina (Terminal React)
──────────────────────────────────────────────────────
DOM                            →    VirtualTerminal
JavaScript                     →    Lua
React Hooks                    →    Lumina Hooks
React State                    →    Component State
Virtual DOM diffing           →    Delta rendering
CSS                            →    CSS DSL
HTTP/WebSocket                 →    MCP Protocol
Browser                        →    Terminal
```

### 1.2 核心特性

| 特性 | 说明 |
|------|------|
| Go + Lua | Go 处理核心运行时，Lua 处理组件逻辑 |
| React 风格 | Hooks、状态管理、组件化 |
| MCP 协议 | AI Agent 可通过 MCP 与 Lumina 通信 |
| 60Hz 渲染 | 流畅的终端渲染体验 |
| 主题支持 | 支持亮色/暗色主题切换 |
| 热重载 | 开发时无需重启应用 |

### 1.3 适用场景

- AI Agent 的调试界面
- 终端工具的现代化 UI
- 需要 AI 可解析输出的交互式应用
- 需要快速迭代的 TUI 应用

---

## 2. 快速开始

### 2.1 安装

```bash
# 克隆项目
git clone https://github.com/your-org/lumina.git
cd lumina

# 安装依赖
go mod download

# 运行示例
go run ./cmd/lumina
```

### 2.2 第一个组件

创建 `main.lua`：

```lua
-- 导入 Lumina
local lumina = require("lumina")

-- 定义计数器组件
local Counter = lumina.defineComponent({
    name = "Counter",
    
    init = function(props)
        local count, setCount = lumina.useState(props.initial or 0)
        return { count = count, setCount = setCount }
    end,
    
    render = function(instance)
        return {
            type = "vbox",
            children = {
                {
                    type = "text",
                    content = "Count: " .. tostring(instance.count)
                },
                {
                    type = "button",
                    label = "+1",
                    onClick = function()
                        instance.setCount(instance.count + 1)
                    end
                },
                {
                    type = "button",
                    label = "-1",
                    onClick = function()
                        instance.setCount(instance.count - 1)
                    end
                }
            }
        }
    end
})

-- 渲染应用
lumina.render(Counter, { initial = 0 })
```

### 2.3 运行

```bash
lumina run main.lua
# 或
go run ./cmd/lumina ./examples/counter.lua
```

### 2.4 输出模式

```bash
# ANSI 模式（默认，人类可读）
lumina run main.lua

# JSON 模式（AI 可解析）
LUMINA_OUTPUT=json lumina run main.lua
```

---

## 3. 核心概念

### 3.1 组件模型

Lumina 使用 **Go Userdata + Lua Metatable** 的混合架构：

```
┌─────────────────────────────────────┐
│  Go 侧 (Userdata)                   │
│  ├─ Props (传入属性)                │
│  ├─ State (状态)                    │
│  ├─ ID (唯一标识)                   │
│  └─ Methods (方法)                  │
├─────────────────────────────────────┤
│  Lua 侧 (Metatable)                 │
│  ├─ init() 初始化                  │
│  ├─ render() 渲染                  │
│  └─ 自定义方法                      │
└─────────────────────────────────────┘
```

### 3.2 生命周期

```
mount (挂载) → render (渲染) → update (更新) → unmount (卸载)
```

| 阶段 | 说明 | Lumina 实现 |
|------|------|-------------|
| mount | 组件实例化 | `init()` 被调用 |
| render | 组件输出 UI | `render()` 返回 VDOM |
| update | 状态变化触发重渲染 | `setState()` 触发 |
| unmount | 组件销毁 | `__gc` 或显式清理 |

### 3.3 虚拟终端

Lumina 在内存中维护一个虚拟终端 buffer：

```
┌────────────────────────────────────┐
│  VirtualTerminal (内存 2D Cell)    │
├────────────────────────────────────┤
│  前端 Buffer (显示中)              │
│  后端 Buffer (渲染中)              │
│  Diff Engine (差分计算)            │
└────────────────────────────────────┘
```

### 3.4 坐标系统

```
原点: 左上角 (0, 0)
X 轴: 左 → 右
Y 轴: 上 → 下
单位: 字符单元格 (cell)
```

---

## 4. 组件开发

### 4.1 定义组件

```lua
local MyComponent = lumina.defineComponent({
    -- 组件名称（用于调试）
    name = "MyComponent",
    
    -- 初始化函数（可选）
    init = function(props)
        -- 返回组件实例状态
        return {
            -- useState hooks
            value, setValue = lumina.useState(props.initial),
            -- 其他状态
            loading = false,
        }
    end,
    
    -- 渲染函数（必须）
    render = function(instance)
        -- 返回 VDOM 描述
        return {
            type = "vbox",
            children = {
                { type = "text", content = "Hello" },
                { type = "button", label = "Click", onClick = function() end }
            }
        }
    end,
    
    -- 清理函数（可选）
    cleanup = function(instance)
        -- 取消订阅、清理资源
    end
})
```

### 4.2 内置组件

#### Box

```lua
{
    type = "box",
    -- 布局
    direction = "horizontal",  -- "horizontal" | "vertical"
    align = "center",         -- "start" | "center" | "end" | "stretch"
    gap = 4,                   -- 间距
    flex = 1,                 -- 弹性比例
    -- 尺寸
    width = 20,               -- 字符数
    height = 10,              -- 行数
    minWidth = 10,
    maxWidth = 40,
    -- 样式
    padding = {4, 8},        -- {上下, 左右}
    border = "rounded",      -- "single" | "double" | "rounded"
    color = "primary",        -- 前景色
    background = "bg",        -- 背景色
    children = {...}
}
```

#### Text

```lua
{
    type = "text",
    content = "Hello, World!",  -- 文本内容
    color = "text",             -- 前景色
    bold = true,                -- 加粗
    dim = false,                -- 暗淡
    align = "left"             -- "left" | "center" | "right"
}
```

#### Button

```lua
{
    type = "button",
    label = "Click me",        -- 按钮文字
    onClick = function() end, -- 点击回调
    disabled = false,           -- 禁用状态
    variant = "default"        -- "default" | "primary" | "danger"
}
```

#### Input

```lua
{
    type = "input",
    value = "",                -- 输入值
    placeholder = "Enter...",  -- 占位符
    onChange = function(value) end,  -- 值变化回调
    onSubmit = function(value) end,  -- 回车提交
    maxLength = 100            -- 最大长度
}
```

#### Dialog

```lua
{
    type = "dialog",
    title = "Confirm",         -- 对话框标题
    content = "Are you sure?", -- 内容
    buttons = {
        { label = "OK", onClick = function() end },
        { label = "Cancel", onClick = function() end }
    }
}
```

### 4.3 组件组合

```lua
-- 布局组件
local Layout = lumina.defineComponent({
    name = "Layout",
    
    render = function()
        return {
            type = "vbox",
            children = {
                { type = "header" },  -- 头部
                {
                    type = "hbox",
                    children = {
                        { type = "sidebar" },  -- 侧边栏
                        { type = "content" }   -- 内容
                    }
                },
                { type = "footer" }  -- 底部
            }
        }
    end
})
```

### 4.4 条件渲染

```lua
render = function(instance)
    if instance.loading then
        return { type = "spinner", label = "Loading..." }
    end
    
    if #instance.items == 0 then
        return { type = "text", content = "No items" }
    end
    
    return {
        type = "vbox",
        children = lumina.map(instance.items, function(item)
            return { type = "item", data = item }
        end)
    }
end
```

### 4.5 列表渲染

```lua
render = function(instance)
    return {
        type = "vbox",
        children = lumina.map(instance.items, function(item, index)
            return {
                type = "listItem",
                children = {
                    { type = "text", content = index .. ". " .. item.title },
                    { type = "button", label = "Delete", onClick = function()
                        instance.setItems(lumina.filter(instance.items, function(i)
                            return i ~= item
                        end))
                    end}
                }
            }
        end)
    }
end
```

---

## 5. Hooks 参考

### 5.1 useState

状态管理 hook。

```lua
local count, setCount = lumina.useState(initialValue)
```

**参数**:
- `initialValue`: 初始状态值

**返回值**:
- `value`: 当前状态值
- `setValue`: 更新状态的函数

**示例**:

```lua
init = function(props)
    local count, setCount = lumina.useState(0)
    local name, setName = lumina.useState("")
    
    return { count = count, setCount = setCount, name = setName }
end

render = function(instance)
    -- 使用状态
    instance.setCount(instance.count + 1)  -- 更新
end
```

### 5.2 useEffect

副作用 hook，在渲染后执行。

```lua
lumina.useEffect(function()
    -- 副作用逻辑
    doSomething()
    
    -- 返回清理函数（可选）
    return function()
        cleanup()
    end
end, { dependency1, dependency2 })
```

**参数**:
- `effect`: 执行函数
- `deps`: 依赖数组

**示例**:

```lua
init = function(props)
    local data, setData = lumina.useState(nil)
    
    -- 数据获取
    lumina.useEffect(function()
        fetch("/api/data")
            :then(setData)
            :catch(function(err) print(err) end)
    end, {})
    
    return { data = data }
end
```

### 5.3 useMemo

记忆化计算。

```lua
local computed = lumina.useMemo(function()
    -- expensive computation
    return expensiveFunction(value)
end, { value })
```

### 5.4 useCallback

记忆化回调函数。

```lua
local handler = lumina.useCallback(function(arg)
    doSomething(arg)
end, { dep1, dep2 })
```

### 5.5 useContext

获取上下文值。

```lua
local theme = lumina.useContext("theme")
```

### 5.6 useRef

存储可变引用。

```lua
local ref = lumina.useRef(initialValue)
ref.current  -- 访问当前值
```

---

## 6. 样式系统

### 6.1 样式定义

```lua
lumina.defineStyle("button", {
    padding = {4, 8},
    color = "primary",
    border = "rounded",
    background = "transparent"
})
```

### 6.2 函数式样式

```lua
lumina.defineStyle("button", function(theme)
    return {
        padding = theme.spacing.md,
        color = theme.colors.primary,
        border = theme.borders.rounded
    }
end)
```

### 6.3 组件内样式

```lua
local MyButton = lumina.defineComponent({
    name = "MyButton",
    
    style = function(theme)
        return {
            layout = { direction = "horizontal", gap = 8 },
            visual = {
                padding = theme.spacing.sm,
                border = theme.borders.rounded,
                color = theme.colors.text
            }
        }
    end,
    
    render = function(instance)
        return {
            type = "hbox",
            style = instance.style,
            children = instance.children
        }
    end
})
```

### 6.4 样式优先级

```
内联 style > 组件 style > extends > 全局样式 > 默认样式
```

### 6.5 样式继承

```lua
lumina.defineStyle("primaryButton", {
    extends = "button",
    color = "accent"
})
```

### 6.6 全局样式

```lua
lumina.defineGlobalStyles({
    button = {
        padding = {4, 8},
        border = "rounded"
    },
    container = {
        padding = {0, 16},
        maxWidth = 80
    }
})
```

---

## 7. 事件处理

### 7.1 组件事件

```lua
{
    type = "button",
    label = "Click",
    onClick = function(event)
        print("Clicked!")
        -- event 包含事件信息
        -- event.x, event.y 终端坐标
        -- event.target 目标组件
    end
}
```

### 7.2 事件对象

```lua
Event = {
    type = "click",           -- 事件类型
    timestamp = 1700000000,  -- 时间戳
    target = component,       -- 目标组件
    
    -- 键盘事件
    key = "Enter",           -- 键名
    code = "Enter",          -- 物理键码
    modifiers = {
        ctrl = false,
        shift = false,
        alt = false,
        meta = false
    },
    
    -- 鼠标事件
    x = 10,                  -- 终端坐标
    y = 5,
    button = "left",
    
    -- 方法
    preventDefault = function() end,
    stopPropagation = function() end
}
```

### 7.3 键盘事件

```lua
{
    type = "input",
    onKeyDown = function(event)
        if event.key == "Enter" then
            print("Submit!")
        elseif event.key == "Escape" then
            print("Cancel")
        end
    end
}
```

### 7.4 全局快捷键

```lua
-- 在应用初始化时注册
lumina.registerShortcut({
    key = "ctrl+c",
    handler = function()
        lumina.exit()
    end,
    global = true
})

lumina.registerShortcut({
    key = "ctrl+l",
    handler = function()
        -- 清屏
    end,
    global = true
})
```

### 7.5 焦点管理

```lua
-- 可聚焦组件
{
    type = "input",
    tabIndex = 0,  -- 可聚焦
    onFocus = function() end,
    onBlur = function() end
}

-- 不可聚焦
{
    type = "text",
    tabIndex = -1  -- 不可聚焦
}
```

### 7.6 焦点陷阱

```lua
lumina.withFocusTrap(true, function()
    return {
        type = "dialog",
        children = {...}
    }
end)
```

---

## 8. 主题系统

### 8.1 定义主题

```lua
local darkTheme = lumina.defineTheme("dark", {
    colors = {
        primary = "cyan",
        secondary = "white",
        background = "black",
        text = "white",
        surface = "gray",
        error = "red",
        success = "green",
        warning = "yellow"
    },
    spacing = {
        xs = 2,
        sm = 4,
        md = 8,
        lg = 16,
        xl = 32
    },
    borders = {
        rounded = "round",
        single = "single",
        double = "double"
    }
})

local lightTheme = lumina.defineTheme("light", {
    colors = {
        primary = "blue",
        background = "white",
        text = "black"
    }
})
```

### 8.2 使用主题

```lua
-- 应用级别设置
lumina.render(App, {
    theme = darkTheme
})

-- 运行时切换
lumina.setTheme("dark")
lumina.setTheme("light")
```

### 8.3 在组件中访问主题

```lua
render = function(instance)
    local theme = lumina.useTheme()
    
    return {
        type = "box",
        color = theme.colors.primary,
        background = theme.colors.surface,
        children = {...}
    }
end
```

### 8.4 内置主题

```lua
lumina.themes.dark    -- 深色主题
lumina.themes.light   -- 浅色主题
lumina.themes.monokai -- 代码高亮风格
```

---

## 9. MCP 协议

### 9.1 协议概述

MCP (Model Context Protocol) 是 Lumina 与 AI Agent 通信的协议。

### 9.2 连接方式

```bash
# 启动 MCP 服务器
lumina run --mcp --port 8080

# AI Agent 连接
mcp connect localhost:8080
```

### 9.3 核心命令

| 命令 | 用途 |
|------|------|
| `lumina/create` | 创建组件 |
| `lumina/update` | 更新组件属性 |
| `lumina/delete` | 删除组件 |
| `lumina/query` | 查询状态 |
| `lumina/batch` | 批量操作 |

### 9.4 创建组件

```protobuf
message CreateRequest {
    string parent_id = 1;
    string component_type = 2;
    map<string, Value> props = 3;
}
```

### 9.5 更新组件

```protobuf
message UpdateRequest {
    string component_id = 1;
    map<string, Value> props = 2;
}
```

### 9.6 调试命令

| 命令 | 用途 |
|------|------|
| `lumina/debug/tree` | 组件树结构 |
| `lumina/debug/inspect` | 组件详情 |
| `lumina/debug/vm` | Lua VM 状态 |
| `lumina/hotreload/reload` | 热重载 |

### 9.7 事件通知

AI Agent 可以接收事件：

```protobuf
message EventNotification {
    string component_id = 1;
    string event_type = 2;
    map<string, Value> event_data = 3;
}
```

---

## 10. 性能优化

### 10.1 渲染预算

目标：60Hz (16.67ms 帧预算)

| 阶段 | 预算 | 目标 |
|------|------|------|
| 脏区收集 | 2ms | 1ms |
| Lua 渲染 | 5ms | 4ms |
| 布局计算 | 3ms | 2ms |
| 样式计算 | 3ms | 2ms |
| VT 更新 | 2ms | 1.5ms |
| 输出 | 1.67ms | 1.5ms |

### 10.2 避免不必要的重渲染

```lua
-- 坏：每次渲染都创建新函数
{
    type = "button",
    onClick = function() doSomething() end  -- 每次 render 都创建新函数
}

-- 好：使用 useCallback
init = function()
    local handler = lumina.useCallback(function()
        doSomething()
    end, {})
    
    return { handler = handler }
end

-- 好：使用
{
    type = "button",
    onClick = instance.handler
}
```

### 10.3 避免不必要的状态更新

```lua
-- 坏：每次点击都更新不相关的状态
onClick = function()
    setA(a + 1)
    setB(b + 1)  -- 不需要
end

-- 好：只更新需要的
onClick = function()
    setA(a + 1)
end
```

### 10.4 使用 useMemo 缓存计算

```lua
-- 坏：每次渲染都重新计算
render = function(instance)
    local filtered = expensiveFilter(instance.items)
end

-- 好：使用 useMemo
init = function()
    local filtered = lumina.useMemo(function()
        return expensiveFilter(instance.items)
    end, { instance.items })
end
```

### 10.5 列表渲染优化

```lua
-- 对于大型列表，考虑虚拟列表
{
    type = "virtualList",
    items = largeArray,
    itemHeight = 1,
    visibleCount = 20,
    renderItem = function(item)
        return { type = "text", content = item }
    end
}
```

---

## 11. 测试指南

### 11.1 单元测试

```lua
-- test/counter_test.lua
local lumina = require("lumina")
local Counter = require("../components/counter")

describe("Counter", function()
    it("should start at initial value", function()
        local instance = Counter.init({ initial = 10 })
        assert.equal(instance.count, 10)
    end)
    
    it("should increment", function()
        local instance = Counter.init({ initial = 0 })
        instance.setCount(instance.count + 1)
        assert.equal(instance.count, 1)
    end)
end)
```

### 11.2 集成测试

```lua
-- test/integration_test.lua
describe("App", function()
    it("should render counter", function()
        local app = lumina.createApp(Counter, { initial = 0 })
        local vdom = app:render()
        
        assert.equal(vdom.type, "vbox")
        assert.is_not_nil(vdom.children)
    end)
end)
```

### 11.3 事件测试

```lua
it("should handle click", function()
    local clicked = false
    local button = {
        type = "button",
        onClick = function() clicked = true end
    }
    
    button.onClick()
    assert.is_true(clicked)
end)
```

### 11.4 样式测试

```lua
it("should apply theme colors", function()
    local theme = lumina.themes.dark
    local style = lumina.computeStyle({
        color = "primary",
        background = "surface"
    }, theme)
    
    assert.is_not_nil(style.color)
    assert.is_not_nil(style.background)
end)
```

### 11.5 运行测试

```bash
# 运行所有测试
lumina test

# 运行特定测试
lumina test ./test/counter_test.lua

# 带覆盖率
lumina test --coverage
```

---

## 12. 常见问题

### Q: 如何调试 Lua 组件？

```lua
-- 添加日志
print("Debug: value = " .. tostring(value))

-- 使用 MCP 调试
lumina.debug.inspect(componentId)

-- 查看组件树
lumina.debug.tree()
```

### Q: 如何处理错误？

```lua
-- try-catch
local ok, err = pcall(function()
    riskyOperation()
end)

if not ok then
    print("Error: " .. tostring(err))
end

-- 错误边界
lumina.defineErrorBoundary({
    onError = function(error)
        return {
            type = "text",
            content = "Error: " .. tostring(error),
            color = "error"
        }
    end,
    children = {...}
})
```

### Q: 如何处理异步操作？

```lua
init = function()
    local data, setData = lumina.useState(nil)
    local loading, setLoading = lumina.useState(false)
    
    lumina.useEffect(function()
        setLoading(true)
        
        lumina.async(function()
            local result = fetchData()
            setData(result)
            setLoading(false)
        end)
    end, {})
    
    return { data = data, loading = loading }
end
```

### Q: 如何实现动画？

```lua
-- 简单动画
lumina.animate({
    from = 0,
    to = 100,
    duration = 1000,
    easing = "easeInOut",
    onUpdate = function(value)
        instance.progress = value
    end
})
```

### Q: 如何实现拖拽？

```lua
{
    type = "draggable",
    onDragStart = function(event) end,
    onDrag = function(event)
        instance.x = event.x
        instance.y = event.y
    end,
    onDragEnd = function(event) end
}
```

---

## 附录

### A. API 速查表

```lua
-- 组件
lumina.defineComponent(config)
lumina.createElement(type, props)
lumina.render(component, props)

-- Hooks
lumina.useState(initial)
lumina.useEffect(effect, deps)
lumina.useMemo(fn, deps)
lumina.useCallback(fn, deps)
lumina.useContext(name)
lumina.useRef(initial)

-- 样式
lumina.defineStyle(name, style)
lumina.defineGlobalStyles(styles)
lumina.getStyle(name)

-- 主题
lumina.defineTheme(name, config)
lumina.setTheme(name)
lumina.useTheme()

-- 事件
lumina.on(event, handler)
lumina.off(event, handler)
lumina.emit(target, event, data)
lumina.registerShortcut(config)

-- 输出
lumina.setOutputMode("ansi")  -- 或 "json"
```

### B. 内置组件列表

| 组件 | 说明 |
|------|------|
| `box` | 容器组件 |
| `text` | 文本组件 |
| `button` | 按钮组件 |
| `input` | 输入框 |
| `dialog` | 对话框 |
| `spinner` | 加载动画 |
| `progress` | 进度条 |
| `list` | 列表 |
| `scrollable` | 可滚动容器 |

### C. 快捷键参考

| 快捷键 | 功能 |
|--------|------|
| `Tab` | 下一个可聚焦元素 |
| `Shift+Tab` | 上一个可聚焦元素 |
| `Enter` | 确认/选择 |
| `Escape` | 取消/关闭 |
| `↑↓←→` | 方向导航 |

---

## 变更日志

### v0.1 (2024-04-24)
- 初始开发文档
- Phase 1 功能规划