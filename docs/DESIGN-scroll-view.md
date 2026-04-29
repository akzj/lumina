# ScrollView 设计方案

> **状态**：设计文档。  
> **相关**：[DESIGN-widgets.md](./DESIGN-widgets.md)（Widget 与事件）、[DESIGN-window-manager.md](./DESIGN-window-manager.md)（窗口管理）、[render-engine-v2.md](./render-engine-v2.md)（引擎）。

## 1. 背景与问题

引擎已有完整的纵向滚动基础设施：

| 已有能力 | 位置 |
|----------|------|
| `overflow: "scroll"` 样式 → 自动裁剪 | `layout.go`, `painter.go` |
| `ScrollY` / `ScrollHeight` 节点字段 | `node.go` |
| `paintScrollChildren` / `paintScrollChildrenClipped` | `painter.go` |
| `hitTestWithOffset(scrollOffsetY)` | `events.go` |
| `autoScroll` / `findScrollableAncestor` | `events.go` |
| `onScroll` Lua 回调 | `events.go` |
| layout 自动为 scrollbar 预留 1 列宽度 | `layout.go:233` |

**但用户使用时需要手写原始 descriptor**：

```lua
local panel = {
    type = "vbox",
    style = { flex = 1, overflow = "scroll", border = "single" },
    children = items,
}
```

缺少的产品化封装：

| 缺口 | 说明 |
|------|------|
| **scrollbar 可视化** | 引擎预留了 1 列，但没有画 scrollbar（thumb/track）。用户看不到滚动位置 |
| **键盘滚动** | 只支持鼠标滚轮，没有 ↑/↓/PageUp/PageDown |
| **编程式滚动** | 无法 `scrollToLine(n)` 或 `scrollToOffset(y)` |
| **横向滚动** | 只有 scrollY，没有 scrollX（引擎层面不支持，需要引擎扩展） |
| **Window 内容区不可滚动** | Window widget 内容区用 `overflow: "hidden"`，内容溢出直接截断 |

## 2. 目标与非目标

### 2.1 目标

- **Go Widget 封装**：`lumina.ScrollView` 作为 Go Widget，管理内部状态（scrollY、scrollbar 拖拽）和事件拦截。
- **scrollbar 可视化**：在预留的 1 列宽度内画 thumb/track，反映当前滚动位置和内容比例。
- **键盘滚动**：焦点在 ScrollView 或其后代节点时，↑/↓/PageUp/PageDown/Home/End 可滚动。
- **语法糖**：`ScrollView { children... }` 即可，无需手写 `overflow: "scroll"`。
- **与 Window 集成（P1）**：Window 内容区可选用 ScrollView 替代 `overflow: "hidden"`。

### 2.2 非目标（首版不做）

- **横向滚动**（P2）：需要引擎层面扩展 `ScrollX`、`hitTestWithOffset` 横向参数、`paintScrollChildren` 横向偏移。工作量大，单独评估。
- **scrollbar 鼠标拖拽**（P1）：首版 scrollbar 仅为视觉指示器，不支持鼠标拖拽 thumb。
- **惯性滚动**（P2）：TUI 场景下优先级低。
- **scrollToLine / scrollToOffset API**（P1）：首版只做滚轮 + 键盘。

## 3. 实现方案

### 3.1 架构：Go Widget

ScrollView 作为 Go Widget（类似 Window），注册在 `pkg/widget/register.go`：

```
┌─────────────────────────────────────────────┐
│  ScrollView (Go Widget)                      │
│  • 内部状态：ScrollViewState { }             │
│  • Render：生成 overflow:"scroll" 的 vbox    │
│  • OnEvent：消费 keydown 做滚动              │
│  • scrollbar 由引擎 painter 绘制             │
└─────────────────────────────────────────────┘
```

### 3.2 Lua 语法

```lua
local ScrollView = lumina.ScrollView

-- 基本用法
ScrollView {
    style = { width = 30, height = 15, border = "single" },

    Text { "Line 1" },
    Text { "Line 2" },
    -- ... 100 行内容，自动滚动
}

-- 完整 props
ScrollView {
    style = { flex = 1, border = "single" },
    showScrollbar = true,    -- 默认 true，是否显示 scrollbar
    onScroll = function(scrollY) end,  -- 可选滚动回调

    -- children
    ...
}
```

### 3.3 Go Widget 实现

**文件**：`pkg/widget/scrollview.go`

```go
// ScrollViewState is the internal state for a ScrollView widget.
type ScrollViewState struct {
    // 首版无需额外状态，scrollY 由引擎 Node.ScrollY 管理
    // P1: ScrollbarDragging bool, DragStartY int, etc.
}

var ScrollView = &Widget{
    Name: "ScrollView",
    NewState: func() any { return &ScrollViewState{} },
    Render: func(props map[string]any, state any) any {
        // 生成 vbox 节点，overflow = "scroll"
        // 将 _childNodes 作为 children
        // 样式从 props["style"] 继承，强制 overflow = "scroll"
    },
    OnEvent: func(props map[string]any, state any, event *WidgetEvent) bool {
        // 消费键盘事件做滚动
        // ↑/↓: 滚动 1 行
        // PageUp/PageDown: 滚动 1 页
        // Home/End: 滚动到顶/底
    },
}
```

### 3.4 scrollbar 可视化

**方案**：在引擎 `painter.go` 的 `paintScrollChildren` 中，绘制完 children 后，在预留的 1 列宽度内画 scrollbar。

```
┌─ ScrollView ──────────┐
│ Line 1                ▲│  ← track (dim)
│ Line 2                █│  ← thumb (bright)
│ Line 3                █│
│ Line 4                ░│  ← track
│ Line 5                ▼│
└───────────────────────┘
```

**scrollbar 绘制逻辑**：
- track：整列用 `░` 或空格 + dim 背景
- thumb：根据 `scrollY / maxScrollY` 和 `visibleH / totalH` 计算位置和长度
- thumb 最小高度 = 1 行
- 当 `maxScrollY == 0`（内容不超出）时，不画 scrollbar

**实现位置**：在 `paintScrollChildren` / `paintScrollChildrenClipped` 末尾追加 scrollbar 绘制。这样 scrollbar 不参与布局（layout 已预留空间），只在绘制阶段渲染。

### 3.5 键盘滚动

**焦点契约**：
- ScrollView Widget 消费 `keydown` 事件的前提是：**焦点在 ScrollView 自身或其后代节点上**。
- 这由 `dispatchWidgetEvent` 的 `findOwnerComponent` 机制保证：只有焦点节点的祖先链上有 ScrollView 组件时，keydown 才会分发到 ScrollView。
- 如果焦点在兄弟 input 上，ScrollView 不会收到 keydown。

**键盘映射**：

| 按键 | 行为 |
|------|------|
| `↑` / `k` | 滚动上 1 行 |
| `↓` / `j` | 滚动下 1 行 |
| `PageUp` | 滚动上 1 页（visible height） |
| `PageDown` | 滚动下 1 页 |
| `Home` / `g` | 滚动到顶部 |
| `End` / `G` | 滚动到底部 |

**注意**：`j/k/g/G` 仅在 ScrollView 有焦点且没有 input 子节点获得焦点时生效。如果焦点在 input 上，keydown 会被 input 先消费（`HandleInputKeyDown` 优先级更高）。

### 3.6 与 Window 集成（P1）

Window widget 当前内容区：
```go
contentBox := &render.Node{
    Type:     "vbox",
    Children: childNodes,
    Style: render.Style{
        Flex:     1,
        Overflow: "hidden",  // ← 截断，不可滚动
    },
}
```

P1 改为：Window 内部自动用 ScrollView 包裹 children，或提供 `scrollable` prop：
```lua
Window {
    title = "Editor",
    scrollable = true,  -- 内容区可滚动（默认 false，保持向后兼容）
    -- children
}
```

## 4. 分阶段计划

| 阶段 | 内容 | 工作量 |
|------|------|--------|
| **P0** | Go Widget（`scrollview.go`）+ scrollbar 可视化（`painter.go`）+ 键盘滚动 + 注册 + 示例 | ~300 行 Go |
| **P1** | `scrollToLine(n)` / `scrollToOffset(y)` API + scrollbar 鼠标拖拽 + Window 集成 | ~200 行 Go |
| **P2** | 横向滚动（引擎扩展 `ScrollX` + layout + painter + events）| ~500 行 Go |

## 5. 验收要点

- `ScrollView { ... }` 语法可用，children 超出容器高度时自动可滚动。
- scrollbar 可视化：thumb 位置和大小正确反映滚动状态。
- 键盘 ↑/↓/PageUp/PageDown/Home/End 可滚动（焦点在 ScrollView 或其后代时）。
- 鼠标滚轮滚动正常（复用引擎已有 `autoScroll`）。
- 内容不超出时，scrollbar 不显示。
- 嵌套 ScrollView 正常工作（外层/内层独立滚动）。
- 所有现有测试通过。
