# Lumina CSS Properties Reference

Lumina 支持的 CSS 属性完整参考。所有属性通过 Lua table 的 `style` 字段设置。

```lua
local el = lumina.createElement
el("box", {
  style = {
    width = 40,
    height = 10,
    background = "#1a1a2e",
    border = "rounded",
  }
}, children)
```

---

## Box Model（盒模型）

### Sizing（尺寸）

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `width` | int \| string | `0` (auto) | 固定宽度（cell 数），或 `"50%"`、`"80vw"` |
| `height` | int \| string | `0` (auto) | 固定高度（cell 数），或 `"50%"`、`"100vh"` |
| `minWidth` | int \| string | `0` | 最小宽度，支持 `"25%"` |
| `maxWidth` | int \| string | `0` | 最大宽度，支持 `"75%"` |
| `minHeight` | int \| string | `0` | 最小高度，支持 `"30%"` |
| `maxHeight` | int \| string | `0` | 最大高度，支持 `"80%"` |

```lua
-- 固定尺寸
{ type = "box", style = { width = 40, height = 10 } }

-- 百分比（相对父容器内容区域）
{ type = "box", style = { width = "50%", height = "100%" } }

-- 视口单位（相对终端窗口尺寸）
{ type = "box", style = { width = "80vw", height = "50vh" } }

-- 约束
{ type = "box", style = { minWidth = "25%", maxWidth = "75%" } }
```

### Spacing（间距）

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `padding` | int | `0` | 四边内边距 |
| `paddingTop` | int | `0` | 上内边距（覆盖 padding） |
| `paddingBottom` | int | `0` | 下内边距（覆盖 padding） |
| `paddingLeft` | int | `0` | 左内边距（覆盖 padding） |
| `paddingRight` | int | `0` | 右内边距（覆盖 padding） |
| `margin` | int | `0` | 四边外边距 |
| `marginTop` | int | `0` | 上外边距（覆盖 margin） |
| `marginBottom` | int | `0` | 下外边距（覆盖 margin） |
| `marginLeft` | int | `0` | 左外边距（覆盖 margin） |
| `marginRight` | int | `0` | 右外边距（覆盖 margin） |
| `gap` | int | `0` | 子元素之间的间距 |

```lua
-- 统一内边距
{ type = "box", style = { padding = 2 } }

-- 分别设置
{ type = "box", style = { paddingTop = 1, paddingLeft = 2, paddingRight = 2 } }

-- 外边距 + 子元素间距
{ type = "vbox", style = { margin = 1, gap = 1 } }
```

### Border（边框）

| 属性 | 类型 | 默认值 | 有效值 |
|------|------|--------|--------|
| `border` | string | `""` (无边框) | `"none"`, `"single"`, `"double"`, `"rounded"` |
| `borderColor` | string | `""` (使用 foreground) | `"#RRGGBB"` 格式 |

```lua
-- 圆角边框
{ type = "box", style = { border = "rounded", borderColor = "#61afef" } }

-- 双线边框
{ type = "box", style = { border = "double", borderColor = "#e06c75" } }
```

> **注意**: 边框固定占 1 cell 宽度，会减少内容区域。

---

## Flexbox 布局

Lumina 默认所有容器使用 flex 布局。`vbox`/`box` = flex-direction: column，`hbox` = flex-direction: row。

### Container Properties（容器属性）

| 属性 | 类型 | 默认值 | 有效值 |
|------|------|--------|--------|
| `justify` | string | `""` (start) | `"start"`, `"center"`, `"end"`, `"space-between"`, `"space-around"` |
| `align` | string | `""` (stretch) | `"stretch"`, `"start"`, `"center"`, `"end"` |
| `gap` | int | `0` | 子元素间距（cell 数） |
| `flexWrap` | string | `""` (nowrap) | `"wrap"`, `"wrap-reverse"` |

```lua
-- 水平居中 + 垂直居中
{ type = "box", style = { justify = "center", align = "center" } }

-- 两端对齐
{ type = "hbox", style = { justify = "space-between" } }

-- 自动换行
{ type = "hbox", style = { flexWrap = "wrap", gap = 1 } }
```

### Item Properties（子项属性）

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `flex` | int | `0` | flex-grow 因子，>0 时按比例分配剩余空间 |
| `flexShrink` | int | `0` (不收缩) | flex-shrink 因子，>0 时按比例收缩溢出空间 |
| `flexBasis` | int | `0` (auto) | 初始主轴尺寸（优先于 width/height），>0 时生效 |
| `alignSelf` | string | `""` (继承父 align) | `"start"`, `"center"`, `"end"`, `"stretch"` |
| `order` | int | `0` | 排列顺序，值小的排前面 |

```lua
-- 弹性子项：2:1 比例分配空间
{ type = "box", style = { flex = 2 } }
{ type = "box", style = { flex = 1 } }

-- 固定基础尺寸 + 可收缩
{ type = "box", style = { flexBasis = 30, flexShrink = 1 } }

-- 单独对齐
{ type = "box", style = { alignSelf = "center" } }

-- 自定义排列顺序
{ type = "box", style = { order = 2 } }  -- 排在后面
{ type = "box", style = { order = -1 } } -- 排在前面
```

---

## Grid 布局

通过 `display = "grid"` 启用 CSS Grid 布局。

### Container Properties（容器属性）

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `display` | string | `""` (flex) | 设为 `"grid"` 启用 Grid |
| `gridTemplateColumns` | string | `""` | 列轨道定义，如 `"1fr 2fr 1fr"`, `"100 200"`, `"1fr 100 2fr"` |
| `gridTemplateRows` | string | `""` | 行轨道定义，如 `"auto 1fr 2fr"` |
| `gridColumnGap` | int | `0` | 列间距（0 时使用 `gap`） |
| `gridRowGap` | int | `0` | 行间距（0 时使用 `gap`） |

```lua
-- 3 列等宽 Grid
{ type = "box", style = {
    display = "grid",
    gridTemplateColumns = "1fr 1fr 1fr",
    gap = 1,
} }

-- 混合 fr + 固定宽度
{ type = "box", style = {
    display = "grid",
    gridTemplateColumns = "20 1fr 2fr",
    gridTemplateRows = "3 1fr",
    gridColumnGap = 2,
    gridRowGap = 1,
} }
```

### Item Properties（子项属性）

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `gridColumn` | string | `""` | 列位置，如 `"1"`, `"1 / 3"`（起始 / 结束） |
| `gridRow` | string | `""` | 行位置，如 `"1"`, `"1 / 3"` |
| `gridColumnStart` | int | `0` | 列起始位置（1-based） |
| `gridColumnEnd` | int | `0` | 列结束位置（1-based，不包含） |
| `gridRowStart` | int | `0` | 行起始位置（1-based） |
| `gridRowEnd` | int | `0` | 行结束位置（1-based，不包含） |

```lua
-- 跨 2 列
{ type = "box", style = { gridColumn = "1 / 3" } }

-- 指定位置
{ type = "box", style = { gridColumnStart = 2, gridColumnEnd = 4, gridRowStart = 1, gridRowEnd = 2 } }

-- 自动放置（不指定位置，按顺序填充）
{ type = "text", content = "cell 1" }
{ type = "text", content = "cell 2" }
```

---

## Positioning（定位）

| 属性 | 类型 | 默认值 | 有效值 |
|------|------|--------|--------|
| `position` | string | `""` (static/relative) | `"relative"`, `"absolute"`, `"fixed"` |
| `top` | int | `0` | 顶部偏移 |
| `left` | int | `0` | 左侧偏移 |
| `right` | int | `-1` (unset) | 右侧偏移 |
| `bottom` | int | `-1` (unset) | 底部偏移 |
| `zIndex` | int | `0` | 堆叠顺序，值大的在上层 |

```lua
-- 绝对定位（相对最近定位祖先）
{ type = "box", style = {
    position = "absolute",
    top = 5,
    left = 10,
    width = 30,
    height = 8,
    zIndex = 10,
} }

-- 固定定位（相对视口）
{ type = "box", style = {
    position = "fixed",
    bottom = 0,
    left = 0,
    right = 0,
    height = 3,
} }
```

> **注意**: `right` 和 `bottom` 默认值为 `-1`（未设置）。设为 `0` 或正数时生效。

---

## Display & Visibility（显示与可见性）

| 属性 | 类型 | 默认值 | 有效值 |
|------|------|--------|--------|
| `display` | string | `""` (flex) | `"none"`, `"grid"` |
| `visibility` | string | `""` (visible) | `"hidden"` |

```lua
-- 隐藏元素（不参与布局，不占空间）
{ type = "box", style = { display = "none" } }

-- 不可见但保留空间（参与布局，不渲染）
{ type = "box", style = { visibility = "hidden" } }
```

**区别**:
- `display = "none"`: 完全从布局中移除，不占空间
- `visibility = "hidden"`: 保留布局空间，但不可见

---

## Overflow（溢出）

| 属性 | 类型 | 默认值 | 有效值 |
|------|------|--------|--------|
| `overflow` | string | `""` (visible) | `"hidden"`, `"scroll"` |

```lua
-- 裁剪溢出内容
{ type = "box", style = { overflow = "hidden", height = 5 } }

-- 可滚动容器
{ type = "vbox", style = { overflow = "scroll", height = 10 } }
```

> `overflow = "scroll"` 启用垂直滚动，通过鼠标滚轮或 `onScroll` 事件控制。

---

## Text（文本）

| 属性 | 类型 | 默认值 | 有效值 |
|------|------|--------|--------|
| `textAlign` | string | `""` (left) | `"center"`, `"right"` |
| `textOverflow` | string | `""` (clip) | `"ellipsis"` |
| `whiteSpace` | string | `""` (normal/wrap) | `"nowrap"` |

```lua
-- 居中文本
{ type = "text", content = "Hello", style = { textAlign = "center" } }

-- 文本溢出省略
{ type = "text", content = "Very long text...", style = {
    textOverflow = "ellipsis",
    whiteSpace = "nowrap",
    width = 20,
} }

-- 不换行
{ type = "text", content = "Single line", style = { whiteSpace = "nowrap" } }
```

---

## Colors（颜色）

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `foreground` | string | `""` (终端默认) | 前景色，`"#RRGGBB"` 格式 |
| `background` | string | `""` (终端默认) | 背景色，`"#RRGGBB"` 格式 |

**别名**: `fg` = `foreground`, `bg` = `background`

```lua
-- 使用完整名
{ type = "text", style = { foreground = "#61afef", background = "#282c34" } }

-- 使用别名
{ type = "text", style = { fg = "#e06c75", bg = "#1a1a2e" } }
```

> **注意**: 仅支持 `#RRGGBB` 格式，不支持 named colors（如 "red"）或 `rgb()` 函数。

---

## Text Decoration（文本装饰）

| 属性 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `bold` | bool | `false` | 粗体 |
| `dim` | bool | `false` | 暗淡 |
| `underline` | bool | `false` | 下划线 |
| `italic` | bool | `false` | 斜体 |
| `strikethrough` | bool | `false` | 删除线 |
| `inverse` | bool | `false` | 反色（前景/背景互换） |

```lua
-- 粗体 + 下划线
{ type = "text", content = "Important", style = { bold = true, underline = true } }

-- 斜体 + 暗淡
{ type = "text", content = "Hint", style = { italic = true, dim = true } }

-- 删除线
{ type = "text", content = "Deprecated", style = { strikethrough = true } }

-- 反色高亮
{ type = "text", content = "Selected", style = { inverse = true } }
```

---

## Units（单位）

| 单位 | 示例 | 说明 |
|------|------|------|
| 绝对值 | `40` | 终端列数（宽度）或行数（高度） |
| 百分比 | `"50%"` | 相对父容器内容区域（去除 padding/border） |
| vw | `"80vw"` | 相对视口宽度百分比 |
| vh | `"100vh"` | 相对视口高度百分比 |
| fr | `"1fr"` | Grid 弹性单位（仅 `gridTemplateColumns`/`gridTemplateRows`） |

### 百分比解析规则

- `width = "50%"` → 父容器内容宽度的 50%
- `minWidth = "25%"` → 同上，用于约束
- 百分比仅在 `vbox`/`hbox` 布局中解析（避免双重解析）

### 视口单位解析规则

- `width = "80vw"` → 终端宽度的 80%
- `height = "50vh"` → 终端高度的 50%
- 视口尺寸在 `LayoutFull` 时设置为根节点尺寸

### fr 单位

- 仅用于 Grid 轨道定义
- `"1fr 2fr 1fr"` → 按 1:2:1 比例分配剩余空间
- 可与固定值混合：`"100 1fr 2fr"` → 第一列固定 100，剩余按 1:2 分配

---

## Node Types（节点类型）

| 类型 | 说明 | 默认布局方向 |
|------|------|--------------|
| `box` / `vbox` | 垂直布局容器 | flex-direction: column |
| `hbox` | 水平布局容器 | flex-direction: row |
| `text` | 文本节点 | — |
| `input` | 单行输入框 | — |
| `textarea` | 多行输入框 | — |
| `component` | 组件占位节点 | 继承子树布局 |

```lua
-- 垂直布局
el("vbox", { style = { gap = 1 } }, {
    el("text", { content = "Line 1" }),
    el("text", { content = "Line 2" }),
})

-- 水平布局
el("hbox", { style = { gap = 2 } }, {
    el("box", { style = { flex = 1 } }),
    el("box", { style = { flex = 1 } }),
})
```

---

## 与 Web CSS 的差异

| Web CSS | Lumina | 说明 |
|---------|--------|------|
| `display: flex` | 默认行为 | 所有容器默认 flex 布局 |
| `flex-direction` | 节点类型决定 | `vbox`/`box` = column, `hbox` = row |
| `display: grid` | `display = "grid"` | 相同语义 |
| `border-width` | 固定 1 cell | TUI 限制，无法设置边框宽度 |
| `border-radius` | 不支持 | TUI 不适用（`"rounded"` 是预设样式） |
| `box-shadow` | 不支持 | TUI 不适用 |
| `transform` | 不支持 | TUI 不适用 |
| `transition` / `animation` | 不支持 | TUI 不适用 |
| `float` | 不支持 | 使用 flex 布局替代 |
| `px` / `em` / `rem` | 不支持 | 使用绝对值（cell）或百分比 |
| `color` | `foreground` / `fg` | 属性名不同 |
| `background-color` | `background` / `bg` | 属性名不同 |
| 颜色格式 | 仅 `#RRGGBB` | 不支持 named colors / `rgb()` / `hsl()` |
| `flex-grow` | `flex` | 属性名不同 |
| `flex-shrink` | `flexShrink` | camelCase |
| `flex-basis` | `flexBasis` | camelCase |
| `flex-wrap` | `flexWrap` | camelCase |
| `align-items` | `align` | 属性名简化 |
| `align-self` | `alignSelf` | camelCase |
| `justify-content` | `justify` | 属性名简化 |
| `text-align` | `textAlign` | camelCase |
| `text-overflow` | `textOverflow` | camelCase |
| `white-space` | `whiteSpace` | camelCase |
| `z-index` | `zIndex` | camelCase |
| `grid-template-columns` | `gridTemplateColumns` | camelCase |
| `grid-template-rows` | `gridTemplateRows` | camelCase |
| `grid-column-gap` / `column-gap` | `gridColumnGap` | camelCase |
| `grid-row-gap` / `row-gap` | `gridRowGap` | camelCase |
| `overflow-x` / `overflow-y` | `overflow` | 仅支持统一设置 |

---

## 完整属性速查表

按字母顺序列出所有 style 属性：

| 属性名 | 类型 | 默认值 | 分类 |
|--------|------|--------|------|
| `align` | string | `""` | Flexbox |
| `alignSelf` | string | `""` | Flexbox |
| `background` / `bg` | string | `""` | Colors |
| `bold` | bool | `false` | Decoration |
| `border` | string | `""` | Border |
| `borderColor` | string | `""` | Border |
| `bottom` | int | `-1` | Positioning |
| `dim` | bool | `false` | Decoration |
| `display` | string | `""` | Display |
| `flex` | int | `0` | Flexbox |
| `flexBasis` | int | `0` | Flexbox |
| `flexShrink` | int | `0` | Flexbox |
| `flexWrap` | string | `""` | Flexbox |
| `foreground` / `fg` | string | `""` | Colors |
| `gap` | int | `0` | Spacing |
| `gridColumn` | string | `""` | Grid |
| `gridColumnEnd` | int | `0` | Grid |
| `gridColumnGap` | int | `0` | Grid |
| `gridColumnStart` | int | `0` | Grid |
| `gridRow` | string | `""` | Grid |
| `gridRowEnd` | int | `0` | Grid |
| `gridRowGap` | int | `0` | Grid |
| `gridRowStart` | int | `0` | Grid |
| `gridTemplateColumns` | string | `""` | Grid |
| `gridTemplateRows` | string | `""` | Grid |
| `height` | int \| string | `0` | Sizing |
| `inverse` | bool | `false` | Decoration |
| `italic` | bool | `false` | Decoration |
| `justify` | string | `""` | Flexbox |
| `left` | int | `0` | Positioning |
| `margin` | int | `0` | Spacing |
| `marginBottom` | int | `0` | Spacing |
| `marginLeft` | int | `0` | Spacing |
| `marginRight` | int | `0` | Spacing |
| `marginTop` | int | `0` | Spacing |
| `maxHeight` | int \| string | `0` | Sizing |
| `maxWidth` | int \| string | `0` | Sizing |
| `minHeight` | int \| string | `0` | Sizing |
| `minWidth` | int \| string | `0` | Sizing |
| `order` | int | `0` | Flexbox |
| `overflow` | string | `""` | Overflow |
| `padding` | int | `0` | Spacing |
| `paddingBottom` | int | `0` | Spacing |
| `paddingLeft` | int | `0` | Spacing |
| `paddingRight` | int | `0` | Spacing |
| `paddingTop` | int | `0` | Spacing |
| `position` | string | `""` | Positioning |
| `right` | int | `-1` | Positioning |
| `strikethrough` | bool | `false` | Decoration |
| `textAlign` | string | `""` | Text |
| `textOverflow` | string | `""` | Text |
| `top` | int | `0` | Positioning |
| `underline` | bool | `false` | Decoration |
| `visibility` | string | `""` | Display |
| `whiteSpace` | string | `""` | Text |
| `width` | int \| string | `0` | Sizing |
| `zIndex` | int | `0` | Positioning |
