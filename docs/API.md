# Lumina Lua API 参考

> 完整的 Lua API 文档，对应 V2 渲染引擎。

---

## 核心 API

### lumina.createComponent(config)

创建并注册根组件。每个应用至少需要一个根组件。

```lua
lumina.createComponent({
    id = "my-app",          -- 必填: 唯一标识
    name = "MyApp",         -- 可选: 显示名称（默认同 id）
    render = function(props)
        -- 必填: 渲染函数，返回 createElement 结果
        return lumina.createElement("box", {}, ...)
    end,
})
```

**参数**:
- `id` (string, 必填) — 组件唯一标识
- `name` (string, 可选) — 显示名称
- `render` (function, 必填) — 渲染函数，接收 `props` 表，返回元素描述

---

### lumina.defineComponent(name, renderFn)

定义可复用的子组件工厂。返回工厂表，可传给 `createElement`。

```lua
local Button = lumina.defineComponent("Button", function(props)
    local hovered, setHovered = lumina.useState("h", false)
    
    return lumina.createElement("box", {
        style = {background = hovered and "#313244" or "#1E1E2E"},
        onMouseEnter = function() setHovered(true) end,
        onMouseLeave = function() setHovered(false) end,
    },
        lumina.createElement("text", {foreground = "#89B4FA"}, props.label)
    )
end)
```

**参数**:
- `name` (string) — 组件类型名（如 "Button", "Cell"）
- `renderFn` (function) — 渲染函数，接收 `props` 表

**返回**: 工厂表 `{_isFactory=true, _name=name}`

**使用**:
```lua
lumina.createElement(Button, {key = "btn1", label = "Click"})
```

> **注意**: 使用子组件时必须传 `key` 或 `id` 属性，用于协调匹配。

---

### lumina.createElement(type, props, ...children)

创建 UI 元素描述。这是构建 UI 的基本函数。

```lua
-- 基本元素
lumina.createElement("box", {
    style = {background = "#1E1E2E", flex = 1},
    onClick = function(e) print("clicked") end,
},
    lumina.createElement("text", {foreground = "#CDD6F4"}, "Hello"),
    lumina.createElement("text", {foreground = "#89B4FA"}, "World")
)

-- 子组件
lumina.createElement(MyComponent, {key = "comp1", someProp = "value"})
```

**参数**:
- `type` (string 或 factory) — 元素类型或子组件工厂
- `props` (table, 可选) — 属性表
- `...children` — 子元素（table 或 string）

**元素类型**:

| 类型 | 说明 | 布局方向 |
|------|------|----------|
| `"box"` | 通用容器 | 垂直（默认） |
| `"vbox"` | 垂直容器 | 垂直 |
| `"hbox"` | 水平容器 | 水平 |
| `"text"` | 文本节点 | — |
| `"input"` | 单行文本输入 | — |
| `"textarea"` | 多行文本输入 | — |

**字符串子节点**: 字符串参数自动合并为 `content` 属性。

```lua
lumina.createElement("text", {foreground = "#CDD6F4"}, "Hello ", "World")
-- 等价于: content = "Hello World"
```

---

### lumina.useState(key, defaultValue)

在当前组件中声明状态变量。

```lua
local count, setCount = lumina.useState("count", 0)
local name, setName = lumina.useState("name", "")
local items, setItems = lumina.useState("items", {})
```

**参数**:
- `key` (string) — 状态键（组件内唯一）
- `defaultValue` (any) — 初始值（仅首次调用时使用）

**返回**: `(currentValue, setterFn)`

**setter 行为**:
- 调用 `setCount(newValue)` → 更新状态 + 标记组件 Dirty
- 如果新值与旧值相同（`reflect.DeepEqual`）→ 不标记 Dirty，不重新渲染
- 下一帧 `RenderDirty()` 时组件会重新调用 renderFn

> **限制**: `useState` 只能在 renderFn 内部调用（需要 currentComponent 上下文）。

---

## 定时器 API

### lumina.setInterval(fn, ms)

设置重复定时器。

```lua
local id = lumina.setInterval(function()
    -- 每 ms 毫秒执行一次
end, 1000)
```

**返回**: timer ID (number)

### lumina.setTimeout(fn, ms)

设置一次性定时器。

```lua
local id = lumina.setTimeout(function()
    -- ms 毫秒后执行一次
end, 500)
```

**返回**: timer ID (number)

### lumina.clearInterval(id) / lumina.clearTimeout(id)

取消定时器。两个函数行为相同。

```lua
lumina.clearInterval(id)
lumina.clearTimeout(id)
```

---

## 应用控制

### lumina.quit()

退出应用（关闭事件循环）。

```lua
onKeyDown = function(e)
    if e.key == "q" then lumina.quit() end
end
```

---

## 属性参考

### 通用属性

| 属性 | 类型 | 说明 |
|------|------|------|
| `id` | string | 节点标识（用于协调匹配） |
| `key` | string | 列表项标识（用于列表协调） |
| `content` | string | 文本内容 |
| `value` | string | input/textarea 的值（等价于 content） |
| `placeholder` | string | input/textarea 占位文本 |
| `autoFocus` | boolean | 初始渲染后自动聚焦 |
| `scrollY` | number | 垂直滚动偏移量 |

### 事件属性

| 属性 | 回调签名 | 说明 |
|------|----------|------|
| `onClick` | `function(e)` | 鼠标点击。`e.x`, `e.y` |
| `onMouseEnter` | `function(e)` | 鼠标进入。`e.x`, `e.y` |
| `onMouseLeave` | `function(e)` | 鼠标离开。`e.x`, `e.y` |
| `onKeyDown` | `function(e)` | 键盘按下。`e.key` |
| `onScroll` | `function(e)` | 滚动。`e.delta`(-1/+1), `e.key`("up"/"down") |
| `onChange` | `function(value)` | input/textarea 值变化。`value` 为当前内容 |

### 样式属性

样式可通过 `style` 子表或直接作为顶层属性传入。`style` 子表优先级更高。

#### 尺寸

| 属性 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `width` | number | 0(自动) | 固定宽度 |
| `height` | number | 0(自动) | 固定高度 |
| `minWidth` | number | 0 | 最小宽度 |
| `maxWidth` | number | 0(无限) | 最大宽度 |
| `minHeight` | number | 0 | 最小高度 |
| `maxHeight` | number | 0(无限) | 最大高度 |
| `flex` | number | 0 | Flex 增长因子 |

#### 间距

| 属性 | 类型 | 说明 |
|------|------|------|
| `padding` | number | 四边内边距（简写） |
| `paddingTop` | number | 上内边距（覆盖简写） |
| `paddingBottom` | number | 下内边距 |
| `paddingLeft` | number | 左内边距 |
| `paddingRight` | number | 右内边距 |
| `margin` | number | 四边外边距（简写） |
| `marginTop` | number | 上外边距（覆盖简写） |
| `marginBottom` | number | 下外边距 |
| `marginLeft` | number | 左外边距 |
| `marginRight` | number | 右外边距 |
| `gap` | number | 子元素间距 |

#### 对齐

| 属性 | 值 | 说明 |
|------|-----|------|
| `justify` | `"start"` `"center"` `"end"` `"space-between"` `"space-around"` | 主轴对齐 |
| `align` | `"stretch"` `"start"` `"center"` `"end"` | 交叉轴对齐 |

#### 视觉

| 属性 | 类型 | 说明 |
|------|------|------|
| `foreground` / `fg` | string | 前景色（如 `"#89B4FA"`） |
| `background` / `bg` | string | 背景色 |
| `bold` | boolean | 粗体 |
| `dim` | boolean | 暗淡 |
| `underline` | boolean | 下划线 |
| `border` | string | 边框: `"single"`, `"double"`, `"rounded"` |

#### 定位

| 属性 | 类型 | 说明 |
|------|------|------|
| `position` | string | `"relative"`, `"absolute"`, `"fixed"` |
| `top` | number | 上偏移 |
| `left` | number | 左偏移 |
| `right` | number | 右偏移（-1=未设置） |
| `bottom` | number | 下偏移（-1=未设置） |
| `zIndex` | number | 层叠顺序 |

#### 溢出

| 属性 | 值 | 说明 |
|------|-----|------|
| `overflow` | `"hidden"`, `"scroll"` | 溢出处理 |

---

## 键盘按键名

事件中 `e.key` 的可能值：

| 按键 | e.key |
|------|-------|
| 字母/数字 | `"a"`, `"A"`, `"1"`, ... |
| 回车 | `"Enter"` |
| 退格 | `"Backspace"` |
| Tab | `"Tab"` |
| Escape | `"Escape"` |
| 方向键 | `"ArrowUp"`, `"ArrowDown"`, `"ArrowLeft"`, `"ArrowRight"` |
| 空格 | `" "` |
| F 键 | `"F1"` ~ `"F12"` |

---

## 原始表语法

除了 `createElement`，也可以直接使用 Lua 表描述 UI：

```lua
-- createElement 方式
lumina.createElement("vbox", {style = {background = "#1E1E2E"}},
    lumina.createElement("text", {}, "Hello")
)

-- 原始表方式（等价）
{
    type = "vbox",
    style = {background = "#1E1E2E"},
    children = {
        {type = "text", content = "Hello"},
    },
}
```

两种方式可以混用。原始表方式在动态生成子节点列表时更方便。

---

## 完整示例

### 计数器

```lua
lumina.createComponent({
    id = "counter",
    render = function(props)
        local count, setCount = lumina.useState("count", 0)
        
        return lumina.createElement("box", {
            style = {background = "#1E1E2E"},
            onClick = function() setCount(count + 1) end,
        },
            lumina.createElement("text", {
                foreground = "#89B4FA",
                bold = true,
            }, "Count: " .. tostring(count))
        )
    end,
})
```

### 带子组件的网格

```lua
local Cell = lumina.defineComponent("Cell", function(props)
    local hovered, setHovered = lumina.useState("h", false)
    
    return lumina.createElement("box", {
        style = {width = 1, height = 1, background = hovered and "#313244" or "#1E1E2E"},
        onMouseEnter = function() setHovered(true) end,
        onMouseLeave = function() setHovered(false) end,
    },
        lumina.createElement("text", {
            foreground = hovered and "#A6E3A1" or "#585B70",
        }, hovered and "█" or "·")
    )
end)

lumina.createComponent({
    id = "grid",
    render = function(props)
        local cells = {}
        for x = 0, 9 do
            cells[#cells + 1] = lumina.createElement(Cell, {key = tostring(x)})
        end
        return {
            type = "hbox",
            style = {height = 1},
            children = cells,
        }
    end,
})
```
