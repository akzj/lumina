# 调试经验：ScrollTo / Jump to Latest 失效问题

## 问题描述

- "Jump to latest" 按钮点击后没有反应，页面没有滚动到底部
- 鼠标滚轮滚动正常，只有 `lumina.scrollTo()` API 调用不生效
- 问题稳定复现：每次点击按钮 `scrollTo` 返回 `false`

## 根因分析（两层 bug）

这个问题由两个独立的 bug 叠加导致。任何一个单独都会使 scroll-to-bottom 失败。

---

### Bug 1：引擎层 — `scrollTo` 无法穿透 component placeholder 节点

**现象**：`lumina.scrollTo("chat-scroll-X", 999999)` 返回 `false`

**原因**：

`ScrollView` 是通过 `lumina.defineComponent("LuxScrollView", ...)` 定义的组件。在 Lumina 的组件模型中：

1. 父节点树中，`ScrollView` 被表示为一个 **component placeholder 节点**（`Type="component"`）
2. 该 placeholder 节点继承了 `id="chat-scroll-X"`，但 **没有** `overflow: "scroll"` 样式
3. 真正的 scroll 容器（`vbox` with `overflow: "scroll"`）是组件的 `RootNode`，被 graft 为 placeholder 的子节点
4. `FindNodeByID("chat-scroll-X")` 找到的是 placeholder（因为 ID 匹配，先命中）

```
[component placeholder]  ← FindNodeByID 返回这个
  id: "chat-scroll-X"
  type: "component"
  overflow: ""          ← 不是 "scroll"！
  └── [vbox]            ← 真正的 scroll 容器（RootNode graft 在这里）
        id: "chat-scroll-X"
        overflow: "scroll"
        scrollHeight: 5000
```

`scrollTo` 的旧逻辑：
```go
node := e.FindNodeByID(id)
if node.Style.Overflow != "scroll" {
    L.PushBoolean(false)
    return 1  // 直接返回 false！
}
```

而鼠标滚轮的 `scrollNode` 处理逻辑（在 `events.go` 中）已经有 `findScrollNodeByID` fallback，所以鼠标滚轮正常工作。这就解释了"滚轮能滚，按钮不能滚"的诡异表现。

**修复**（`pkg/render/engine_lua_api.go`）：

给 `scrollTo` 和 `getScrollInfo` 添加同样的 fallback 逻辑：

```go
node := e.FindNodeByID(id)
if node == nil {
    L.PushBoolean(false)
    return 1
}
// If the found node is a component placeholder (not overflow:scroll itself),
// search its children for the actual scroll container with the same ID.
if node.Style.Overflow != "scroll" {
    node = e.findScrollNodeByID(node, id)
    if node == nil {
        L.PushBoolean(false)
        return 1
    }
}
```

`findScrollNodeByID`（定义在 `events.go:1140`）递归搜索子树，找到同 ID 且 `overflow == "scroll"` 的节点：

```go
func (e *Engine) findScrollNodeByID(node *Node, id string) *Node {
    for _, child := range node.Children {
        if child.ID == id && child.Style.Overflow == "scroll" {
            return child
        }
        if found := e.findScrollNodeByID(child, id); found != nil {
            return found
        }
    }
    return nil
}
```

**教训**：

- 当 scroll 容器被包装在 `defineComponent` 中时，`FindNodeByID` 返回的是 placeholder，不是真正的 scroll 节点
- 所有操作 scroll 的 API（`scrollTo`、`getScrollInfo`、`scrollNode`、`scrollNodeH`）都必须有 placeholder fallback
- 新增 scroll 相关 API 时，必须考虑 component 嵌套的情况
- **验证方法**：对比 `FindNodeByID` 返回节点的 `Style.Overflow` 和预期值

---

### Bug 2：Lua 层 — `setTimeout(fn, 0)` 时机不对

**现象**：即使修复了 Bug 1，`setTimeout` 回调中 `scrollTo` 仍然可能失败（`FindNodeByID` 返回 nil 或者节点尚未完成 layout）

**原因**：

`setTimeout(fn, 0)` 注册的回调在 `fireTimers()` 中执行。Lumina 事件循环的执行顺序：

```
每个 tick:
  1. fireTimers()       ← setTimeout 回调在这里执行
  2. RenderDirty()      ← 渲染在这之后！
```

当初始加载 effect 中调用 `setMessages()` + `setTimeout(scrollTo, 0)` 时：

1. `setMessages(msgs)` — 标记组件 dirty（需要重新渲染）
2. `setTimeout(scrollTo, 0)` — 注册下一个 tick 执行的回调
3. 当前 `RenderDirty` 完成，本 tick 结束
4. **下一个 tick**：`fireTimers()` → `scrollTo` 执行 → 但组件还没重新渲染！
5. 组件的新节点树还没有被 graft/sync/layout，scroll 容器不存在或 `ScrollHeight = 0`

**`RenderDirty` 的内部执行顺序**：

```
RenderDirty():
  1. renderInOrder()          — 渲染 dirty 组件，生成新 VNode 树
  2. graftChildComponents()   — 将子组件 RootNode 嫁接到父树
  3. syncMainLayer()          — layers[0].Root = root.RootNode（完整树）
  4. layoutAll()              — 计算布局（ScrollHeight, H, W 等）
  5. paintAll()               — 绘制到 buffer
  6. firePendingEffects()     — 执行 useEffect 回调 ← 正确的执行时机！
```

关键洞察：**`useEffect` 回调在 `firePendingEffects`（步骤 6）中执行**，此时节点树已经完成 graft + sync + layout，所有信息就绪。

**修复**（`zerofas/tui/components/organisms/chat_panel.lua`）：

用 `useEffect`（无 deps）+ ref flag 替代 `setTimeout`：

```lua
local scrollRequestedRef = lumina.useRef(false)
local _, setForceScrollTrigger = lumina.useState("_scrollTrig", 0)

-- 请求在下一次渲染后 scroll to bottom
local function scrollToBottomNextFrame()
    scrollRequestedRef.current = true
    setForceScrollTrigger(os.clock())  -- 强制触发重新渲染
end

-- 无 deps → 每次渲染都执行，在 firePendingEffects 中（layout 之后）
lumina.useEffect(function()
    if scrollRequestedRef.current then
        scrollRequestedRef.current = false
        lumina.scrollTo(scrollId, 999999)
    end
end)  -- 注意：没有 deps 参数 → 每次渲染后都执行
```

工作原理：
1. `scrollToBottomNextFrame()` 设置 ref flag + 触发状态变更
2. 状态变更导致组件被标记 dirty
3. `RenderDirty()` 渲染组件 → graft → sync → layout → **firePendingEffects**
4. 在 `firePendingEffects` 中，无 deps 的 `useEffect` 被执行
5. 此时检查 ref flag，调用 `scrollTo`，节点树已就绪，成功！

**教训**：

- `setTimeout(fn, 0)` **不等于** "下一帧渲染后执行"，它在 `fireTimers` 中执行，**早于** `RenderDirty`
- 需要在渲染后执行逻辑时，应使用 `useEffect`（在 `firePendingEffects` 中执行，此时节点树已完成 graft + layout）
- 无 deps 的 `useEffect` 每次渲染都会执行，配合 ref flag 可以实现"仅在需要时执行一次"
- 这个模式类似于 React 中 `useLayoutEffect` 的用法

---

## 调试方法论

### 第一步：加诊断日志

在关键路径上加 `print("[SCROLL-DIAG] ...")` 追踪调用链：

```lua
-- 在 onClick 中
print("[SCROLL-DIAG] button clicked")

-- 在 scrollToBottomNextFrame 中
print("[SCROLL-DIAG] scrollToBottomNextFrame called")

-- 在实际 scrollTo 调用处
local ok = lumina.scrollTo(scrollId, 999999)
print("[SCROLL-DIAG] scrollTo result:", ok)

-- 检查 getScrollInfo
local info = lumina.getScrollInfo(scrollId)
if info then
    print("[SCROLL-DIAG] scrollY:", info.scrollY, "maxScroll:", info.maxScroll,
          "scrollHeight:", info.scrollHeight, "visibleH:", info.visibleH)
else
    print("[SCROLL-DIAG] getScrollInfo returned nil!")
end
```

当 `scrollTo` 返回 `false` 或 `getScrollInfo` 返回 `nil` 时，问题出在引擎层（节点查找/类型判断）。当 `scrollTo` 返回 `true` 但不生效时，可能是时机问题（执行时 scrollHeight 还是 0）。

### 第二步：写 E2E 测试

不要依赖手动测试！使用已有的测试基础设施验证：

```go
// 使用 setupSessionApp() 获得完整 app 环境
app := setupSessionApp(t)

// 模拟操作
clickAt(app, x, y)

// 触发 timer 和渲染
app.FireTimers()
app.Render()

// 直接检查节点树状态
node := engine.Root().RootNode
scrollNode := findNodeByID(layers[0].Root, "chat-scroll-X")
assert(scrollNode.ScrollY == scrollNode.ScrollHeight - scrollNode.H)
```

关键对比：
- `engine.Root().RootNode` vs `layers[0].Root` — 在 `syncMainLayer` 前可能不一致
- `FindNodeByID` 返回的节点类型 — 是 `component` 还是真正的 scroll 容器

### 第三步：理解节点树的生命周期

```
Component render → VNode tree
  ↓ graftChildComponents
Parent tree ← child RootNode inserted as child of placeholder
  ↓ syncMainLayer
layers[0].Root = root.RootNode（完整树，可被 FindNodeByID 搜索）
  ↓ layoutAll
ScrollHeight, H, W 等数值计算完成
  ↓ firePendingEffects
useEffect 回调在这里执行（此时一切就绪）
```

**关键**：`FindNodeByID` 搜索的是 `layers[0].Root`。在 `syncMainLayer` 之前，新渲染的子组件可能还不在搜索范围内。

---

## 通用规则：何时使用 `useEffect` vs `setTimeout`

| 场景 | 正确方案 | 原因 |
|------|----------|------|
| 渲染后操作 DOM（scroll、focus） | `useEffect` | 在 `firePendingEffects` 中执行，layout 已完成 |
| 延迟执行（真正的时间延迟） | `setTimeout(fn, ms)` | ms > 0 时是正确用法 |
| "下一帧"执行 | `useEffect` + ref flag | `setTimeout(fn, 0)` 在 render 前执行，不是"下一帧后" |
| 依赖其他 state 变更后的 DOM | `useEffect(fn, {dep})` | deps 变化时在渲染后执行 |

---

## 相关文件

| 文件 | 说明 |
|------|------|
| `lumina/pkg/render/engine_lua_api.go` | `scrollTo`, `getScrollInfo` 实现（含 placeholder fallback） |
| `lumina/pkg/render/events.go:1140` | `findScrollNodeByID` 定义 |
| `lumina/lua/lux/scrollview.lua` | ScrollView 组件（`defineComponent` 包装，产生 placeholder） |
| `zerofas/tui/components/organisms/chat_panel.lua` | ChatPanel 的 scroll-to-bottom 逻辑 |

## 相关测试

| 测试文件 | 覆盖内容 |
|----------|----------|
| `lumina/pkg/scroll_jump_test.go` | 纯引擎 scroll 测试（component placeholder fallback） |
| `zerofas/cmd/zerofas-tui/scroll_jump_e2e_test.go` | E2E 滚动测试（完整 app 环境） |

## 相关 Commits

| Commit | 仓库 | 说明 |
|--------|------|------|
| `e32680f` | lumina | fix(render): scrollTo/getScrollInfo fall through component placeholder nodes |
| `4bd0c5d` | zerofas | fix(tui): use useEffect + ref flag for scroll-to-bottom, add E2E test |
