# VList: Virtual Scrolling List Design

> **Status**: Draft (Revised after review)  
> **Inspired by**: [ScrollU](https://github.com/akzj/scroll-u) — clip-on-idle strategy  
> **Pattern**: React `ref` prop for container property access (Phase 3, independent)

---

## 1. Motivation

Lumina currently renders lists by creating a Node for every item:

```lua
for i = 1, 1000 do
    children[i] = lumina.createElement("text", {}, "Item " .. i)
end
```

This creates 1000 Nodes — all participate in layout, paint, and consume memory. For chat applications with 50,000+ messages, this is unacceptable.

**Goal**: O(visible) rendering and memory, not O(total).

---

## 2. Strategy: Clip-on-Idle

Traditional virtual lists (react-window) only render visible items — creating/destroying nodes during scroll. This causes jank during fast scrolling.

ScrollU's approach is different:

```
Active scrolling: items accumulate in both directions (smooth, no jank)
        ↓
Idle (no scroll for N ms): clip!
        ↓
Items far outside viewport are removed from memory
        ↓
Scroll back: items are re-created via renderItem()
```

**Trade-off**: Memory grows during active scroll, but drops back to O(visible) after idle. This is acceptable because:
- During scroll, the user is actively engaged — smoothness matters more than memory
- After idle, memory is reclaimed
- Peak memory is bounded by scroll speed × idle timeout, not total item count

---

## 3. Scroll Position Tracking

### 3.1 Design Decision: `useState` + `onScroll(e.scrollY)`

**NOT `ref.current.scrollY`**: The `ref` prop is populated AFTER layout/paint. But VList reads `scrollY` DURING render. Using `ref.current.scrollY` would always be one frame behind.

**Correct approach**: Track scrollY via `useState`, updated by `onScroll`:

```lua
local scrollY, setScrollY = lumina.useState("scrollY", 0)

onScroll = function(e)
    setScrollY(e.scrollY)  -- triggers re-render with current value
end
```

### 3.2 Engine Change Required (Phase 0)

`callLuaRefScroll` must include `scrollY` and `scrollHeight` in the event table:

```go
L.PushInteger(int64(scrollNode.ScrollY))
L.SetField(tblIdx, "scrollY")
L.PushInteger(int64(scrollNode.ScrollHeight))
L.SetField(tblIdx, "scrollHeight")
```

Also, `HandleScroll` must allow `onScroll` + `autoScroll` to coexist (see §7 Phase 0).

---

## 4. VList Component (Pure Lua)

### 4.1 API

```lua
local VList = require("lux.vlist")

lumina.createElement(VList, {
    -- Required
    totalCount = 50000,           -- total number of items (for scrollbar)
    renderItem = function(index)  -- called to create item element
        return lumina.createElement("hbox", {key = "msg-" .. index},
            lumina.createElement("text", {}, "Message " .. (index + 1)),
        )
    end,

    -- Optional
    overscan = 15,                -- items to keep beyond viewport (default: 10)
    estimateHeight = 1,           -- estimated height per item in rows (default: 1)
    height = 20,                  -- viewport height (default: parent height)
    clipDelay = 500,              -- ms idle before clipping (default: 500)
})
```

### 4.2 How It Works

```
totalCount=1000, estimateHeight=1, viewH=20, overscan=5

ScrollY=0:
  ┌─── viewport (20 rows) ───┐
  │ renderItem(0)             │  ← actual Node
  │ renderItem(1)             │
  │ ...                       │
  │ renderItem(24)            │  ← endIdx = 0+20+5 = 25
  ├───────────────────────────┤
  │ bottom spacer (975 rows)  │  ← no Nodes, just layout space
  └───────────────────────────┘

ScrollY=50 (scrolled down):
  ┌───────────────────────────┐
  │ top spacer (45 rows)      │
  ├─── viewport ──────────────┤
  │ renderItem(45)            │  ← startIdx = 50-5 = 45
  │ ...                       │
  │ renderItem(74)            │  ← endIdx = 50+20+5 = 75
  ├───────────────────────────┤
  │ bottom spacer (925 rows)  │
  └───────────────────────────┘
```

### 4.3 Key-Based Reuse

Lumina's reconciler matches children by `key`. When scroll changes:

```
Frame 1: items [5, 6, 7, ..., 35]
Frame 2: items [10, 11, 12, ..., 40]

Reconciler:
  key "i10".."i35" → match → reuse existing Node (no re-render!) ✨
  key "i36".."i40" → new → call renderItem() to create
  key "i5".."i9"   → removed → destroy Node
```

**This is the foundation**: already-rendered items are never re-rendered. Only newly visible items trigger `renderItem`.

### 4.4 Clip-on-Idle

```
Active scroll:
  - Items accumulate in both directions
  - overscan=15 items kept beyond viewport
  - Memory: ~50 Nodes (15+20+15)

Scroll stops → clipDelay ms passes:
  - Clip range: overscan*2 = 30 items beyond viewport
  - Remove items beyond clip range from children array
  - Spacer nodes fill the gaps
  - Memory: ~30 Nodes (10+20+10)

Scroll resumes:
  - Items re-created via renderItem() as they enter overscan range
```

### 4.5 Pseudo-Implementation (Revised)

```lua
local VList = lumina.defineComponent("LuxVList", function(props)
    local totalCount = props.totalCount or 0
    local renderItem = props.renderItem
    local overscan = props.overscan or 10
    local estimateHeight = props.estimateHeight or 1
    local clipDelay = props.clipDelay or 500

    -- Track scrollY via state (NOT ref — avoids frame lag)
    local scrollY, setScrollY = lumina.useState("scrollY", 0)
    local viewH = props.height or 20
    local clipTimerRef = lumina.useRef()

    -- Calculate visible range
    local startIdx = math.max(0, math.floor(scrollY / estimateHeight) - overscan)
    local endIdx = math.min(totalCount, math.ceil((scrollY + viewH) / estimateHeight) + overscan)

    -- Build children
    local children = {}

    -- Top spacer
    if startIdx > 0 then
        children[#children + 1] = lumina.createElement("box", {
            key = "vlist_top",
            style = {height = startIdx * estimateHeight}
        })
    end

    -- Visible items (reconciler reuses by key)
    for i = startIdx, endIdx - 1 do
        children[#children + 1] = renderItem(i)
    end

    -- Bottom spacer
    if endIdx < totalCount then
        children[#children + 1] = lumina.createElement("box", {
            key = "vlist_bottom",
            style = {height = (totalCount - endIdx) * estimateHeight}
        })
    end

    return lumina.createElement("vbox", {
        style = {height = viewH, overflow = "scroll"},
        onScroll = function(e)
            setScrollY(e.scrollY)  -- absolute scrollY from engine
            -- Clip-on-idle: clear old timer, set new one
            if clipTimerRef.current then
                lumina.clearTimeout(clipTimerRef.current)
            end
            clipTimerRef.current = lumina.setTimeout(function()
                -- Re-render with clipped range
                setScrollY(e.scrollY)  -- trigger re-render for clip
            end, clipDelay)
        end,
    }, table.unpack(children))
end)
```

---

## 5. Variable Height Items (Future)

Items may have different heights (e.g., multi-line messages, embedded images). The `estimateHeight` approach causes scrollbar inaccuracy.

### 5.1 Height Cache Strategy

```lua
-- Track actual heights of rendered items
local heights = lumina.useRef("heights", {})  -- {[index] = actualHeight}

-- After render, record actual heights
-- For unrendered items, use estimateHeight

-- Spacer height = sum of actual heights for hidden rendered items
--               + estimateHeight * count of hidden unrendered items
```

### 5.2 How to Get Actual Height

Requires the `ref` prop (Phase 3) to expose child layout info, or a separate mechanism to query rendered node dimensions from Lua.

This is a future enhancement. Initial implementation uses fixed `estimateHeight`.

---

## 6. `ref` Prop (Independent Feature — Phase 3)

`ref` prop is NOT a prerequisite for VList. It's a separate, generally useful feature following React's pattern.

### 6.1 API

```lua
local containerRef = lumina.useRef()

lumina.createElement("vbox", {
    ref = containerRef,
    style = {overflow = "scroll"},
}, ...)

-- After mount, engine populates automatically:
-- containerRef.current.width, .height
-- containerRef.current.contentWidth, .contentHeight
-- containerRef.current.scrollY, .scrollX
-- containerRef.current.isAtTop, .isAtBottom
```

### 6.2 Engine Changes

| File | Change |
|------|--------|
| `descriptor.go` | Read `ref` prop, store ref table reference |
| `node.go` | Add `RefTable LuaRef` field |
| `reconciler.go` | Reconcile `ref` prop |
| `engine.go` | After layout/paint: `populateRefs()` — fill `ref.current` |
| `engine.go` | On node removal: `clearRef()` |

### 6.3 Lifecycle

```
createElement({ref = myRef}) → Descriptor.RefTable = myRef
Reconcile: createNodeFromDesc() → Node.RefTable = myRef
After layout + paint → myRef.current = { width, height, scrollY, ... }
Node removed → myRef.current = nil
```

---

## 7. Implementation Plan

### Phase 0: Fix Scroll Event Handling (Engine — ~20 lines)

**Problem**: `onScroll` handler and `autoScroll` are mutually exclusive in `HandleScroll`. If a node has `onScroll`, the engine returns early and never calls `autoScroll`. VList needs both.

**Fix**:

| Step | File | Description |
|------|------|-------------|
| 0.1 | `events.go:HandleScroll` | Remove early `return` after `callLuaRefScroll`. Always do autoScroll first, then call onScroll handler with the updated scrollY |
| 0.2 | `events.go:callLuaRefScroll` | Add `scrollY` and `scrollHeight` fields to the Lua event table |

### Phase 1: VList v1 — Basic (Pure Lua)

No clip, no variable height. Uses `onScroll(e.scrollY)` + `useState`.

| Step | File | Description |
|------|------|-------------|
| 1.1 | `lua/lux/vlist.lua` | VList component: overscan + spacers + key-based children |
| 1.2 | `examples/vlist_demo.lua` | Demo with 50000 items |

### Phase 2: Clip-on-Idle (Lua + setTimeout)

`lumina.setTimeout` / `lumina.clearTimeout` already exist in `pkg/app_timer.go`.

| Step | File | Description |
|------|------|-------------|
| 2.1 | `lua/lux/vlist.lua` | Add clip-on-idle: debounced `setTimeout` removes far-offscreen items |

### Phase 3: `ref` Prop (Engine — Independent)

| Step | File | Description |
|------|------|-------------|
| 3.1 | `node.go` | Add `RefTable LuaRef` field |
| 3.2 | `descriptor.go` | Read `ref` prop, store ref table reference |
| 3.3 | `reconciler.go` | Reconcile `ref` prop |
| 3.4 | `engine.go` | After layout/paint: `populateRefs()` — fill `ref.current` |
| 3.5 | `engine.go` | On node removal: `clearRef()` |

### Phase 4: Variable Height (Future)

| Step | Description |
|------|-------------|
| 4.1 | Height cache in Lua |
| 4.2 | Accurate spacer heights |
| 4.3 | Accurate scrollbar |

---

## 8. Memory Analysis

```
Scenario: 50,000 messages, viewport=20 rows, overscan=15, estimateHeight=1

                    Nodes in memory
Traditional:        50,000 Nodes       (~50MB)
Active scrolling:   ~50 Nodes          (~50KB)     ← 1000x less
After clip (idle):  ~30 Nodes          (~30KB)     ← 1600x less
```

---

## 9. Comparison with Other Approaches

| | react-window | ScrollU | Lumina VList |
|---|---|---|---|
| Visible-only rendering | ✅ | ❌ (accumulates during scroll) | ❌ (accumulates during scroll) |
| Clip-on-idle | ❌ | ✅ | ✅ |
| Variable height | ✅ (with measured) | ✅ (native DOM) | 🔸 (estimate, future: cache) |
| Implementation | Complex | Simple | Simple (pure Lua) |
| Scroll smoothness | Can jank | Smooth | Smooth |
| Memory (idle) | O(visible) | O(visible) | O(visible) |

---

## 10. Review Notes

### Critical Design Decisions

1. **VList uses `useState` + `onScroll(e.scrollY)`, NOT `ref.current.scrollY`** — avoids frame-lag because ref is populated after layout/paint, but VList reads scrollY during render.

2. **`onScroll` + `autoScroll` must coexist** — current engine behavior is mutually exclusive (Phase 0 fixes this).

3. **`ref` prop is independent** — VList v1 does not need it. Develop in parallel.

4. **`setTimeout` already exists** — `lumina.setTimeout(fn, ms)` and `lumina.clearTimeout(id)` are available for clip-on-idle debounce.
