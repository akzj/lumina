# VList: Virtual Scrolling List Design

> **Status**: Complete (Phases 0-5 implemented)
> **Inspired by**: [ScrollU](https://github.com/akzj/scroll-u) — clip-on-idle strategy
> **Pattern**: React `ref` prop for container property access

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

### 3.2 Engine Change (Phase 0)

`HandleScroll` allows `onScroll` + `autoScroll` to coexist. `callLuaRefScroll` includes `scrollY` and `scrollHeight` in the event table:

```go
L.PushInteger(int64(scrollNode.ScrollY))
L.SetField(tblIdx, "scrollY")
L.PushInteger(int64(scrollNode.ScrollHeight))
L.SetField(tblIdx, "scrollHeight")
```

---

## 4. VList Component (Pure Lua)

### 4.1 API

```lua
local VList = require("lux.vlist")

lumina.createElement(VList, {
    -- Required
    totalCount = 50000,           -- total number of items
    renderItem = function(index)  -- called to create item element
        return lumina.createElement("hbox", {key = "msg-" .. index},
            lumina.createElement("text", {}, "Message " .. (index + 1)),
        )
    end,

    -- Optional
    overscan = 15,                -- items to keep beyond viewport (default: 10)
    estimateHeight = 1,           -- fallback height per item in rows (default: 1)
    height = 20,                  -- viewport height (default: 20)
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
  key "item-10".."item-35" → match → reuse existing Node (no re-render!)
  key "item-36".."item-40" → new → call renderItem() to create
  key "item-5".."item-9"   → removed → destroy Node
```

**This is the foundation**: already-rendered items are never re-rendered. Only newly visible items trigger `renderItem`.

### 4.4 Clip-on-Idle

Uses `isClipped` state + `lumina.setTimeout`/`clearTimeout`:

```
Active scroll:
  - isClipped = false
  - effectiveOverscan = overscan (e.g., 10)
  - Memory: ~(viewport + 2*overscan) Nodes

Scroll stops → clipDelay ms passes:
  - isClipped = true
  - effectiveOverscan = max(2, floor(overscan/2)) (e.g., 5)
  - Reconciler removes items outside new range
  - Memory: ~(viewport + 2*reduced_overscan) Nodes

Scroll resumes:
  - isClipped = false (immediately on scroll)
  - old timer cleared, new timer set
  - Items re-created via renderItem() as they enter overscan range
```

### 4.5 Variable Height

Items can have different actual heights. Uses height cache + ref-based measurement:

```
Frame N:
  1. Render using heightCache (actual for measured, estimateHeight for unknown)
  2. Layout → populateRefs fills item refs with {x, y, w, h}
  3. Paint
  4. useEffect reads ref.current.h, updates heightCache → triggers re-render

Frame N+1:
  1. Render using updated heightCache
```

Each visible item is wrapped in `<box ref={itemRefs[i]}>` for measurement. The wrapper has no explicit height, so it sizes to its child's actual height.

### 4.6 Prefix Sum Optimization

`findIndexAt` uses binary search (O(log n)), `heightUpTo` uses direct lookup (O(1)). Prefix sums built once per render (single O(n) pass).

```
Before: 5 × O(n) linear scans per render (~250K iterations for 50K items)
After:  1 × O(n) prefix sum build + O(log n) lookups (~50K iterations)
```

---

## 5. `ref` Prop (Phase 3)

### 5.1 API

```lua
local ref = lumina.useRef()

lumina.createElement("vbox", {
    ref = ref,
    style = {overflow = "scroll"},
}, ...)

-- After layout, engine populates automatically:
-- ref.current = {x, y, w, h, scrollY, scrollHeight, id, type}
```

### 5.2 Lifecycle

```
createElement({ref = myRef}) → Descriptor.RefTableRef = myRef
Reconcile: createNodeFromDesc() → Node.RefTableRef = myRef
After layout → populateRefs() → ref.current = {x, y, w, h, ...}
Node removed → drainPendingUnrefs() → ref.current = nil → Unref
```

### 5.3 Engine Changes

| File | Change |
|------|--------|
| `descriptor.go` | Added `RefTableRef LuaRef` field |
| `node.go` | Added `RefTableRef LuaRef` field |
| `engine_marshal.go` | Read `ref` prop in `readDescriptor`, store as registry ref |
| `reconciler.go` | Thread through `createNodeFromDesc`, `updateRef`, `collectNodeRefs` |
| `engine.go` | Call `populateRefs` after layout; set `ref.current=nil` in `drainPendingUnrefs` |
| `layout.go` | Added `populateRefs()` — walks tree, fills `ref.current` |

---

## 6. Implementation Summary

| Phase | Content | Commit | Status |
|-------|---------|--------|--------|
| 0 | `onScroll` + `autoScroll` coexist; `callLuaRefScroll` passes `scrollY`/`scrollHeight` | `ac6a236` | ✅ |
| 1 | VList v1: overscan + spacers + key-based children (pure Lua) | `2f17838` | ✅ |
| 2 | Clip-on-idle: `isClipped` state + `lumina.setTimeout`/`clearTimeout` | `af84578` | ✅ |
| 3 | `ref` prop: engine populates `ref.current` after layout | `99e51f9` | ✅ |
| 4 | Variable height: heightCache + ref-based measurement + useEffect | `5cc6f43` | ✅ |
| 5 | Prefix sum optimization: O(log n) findIndexAt + O(1) heightUpTo | `6d74f78` | ✅ |

### Files

| File | Purpose |
|------|---------|
| `lua/lux/vlist.lua` | VList component (152 lines Lua) |
| `pkg/testdata/lua_tests/lux/vlist_test.lua` | 15 test cases |
| `examples/vlist_demo.lua` | 50K item demo |
| `pkg/render/events.go` | Phase 0: HandleScroll fix |
| `pkg/render/descriptor.go` | Phase 3: RefTableRef field |
| `pkg/render/node.go` | Phase 3: RefTableRef field |
| `pkg/render/engine_marshal.go` | Phase 3: read `ref` prop |
| `pkg/render/reconciler.go` | Phase 3: reconcile `ref` |
| `pkg/render/engine.go` | Phase 3: populateRefs + drainPendingUnrefs |
| `pkg/render/layout.go` | Phase 3: populateRefs function |
| `pkg/testdata/lua_tests/hooks/use_ref_test.lua` | Phase 3: 4 ref prop tests |

---

## 7. Memory Analysis

```
Scenario: 50,000 messages, viewport=20 rows, overscan=10, estimateHeight=1

                    Nodes in Go tree    Lua tables (itemRefs + heightCache)
Traditional:        50,000 (~50MB)      N/A
Active scrolling:   ~40 (~40KB)         grows with items scrolled past
After clip (idle):  ~24 (~24KB)         same as above (not cleaned)

Lua-side accumulation: 50K items all scrolled past ≈ 2-4MB in itemRefs + heightCache.
Acceptable for TUI. Can be cleaned in clip callback if needed.
```

---

## 8. Known Limitations

1. **Lua-side memory accumulation**: `itemRefs.current` and `heightCache` grow with total items scrolled past. Not cleaned by clip-on-idle (~2-4MB for 50K items). Acceptable for TUI.

2. **Prefix sums rebuilt every render**: Single O(n) pass per render. Could be cached with `useMemo` but <1ms for 50K items — not worth the complexity.

3. **No callback refs**: Only object refs (`ref = myRef`) supported, not `ref = function(node) end`.

4. **No horizontal virtual scrolling**: VList is vertical-only.

---

## 9. Comparison with Other Approaches

| | react-window | ScrollU | Lumina VList |
|---|---|---|---|
| Visible-only rendering | ✅ | ❌ (accumulates during scroll) | ❌ (accumulates during scroll) |
| Clip-on-idle | ❌ | ✅ | ✅ |
| Variable height | ✅ (with measured) | ✅ (native DOM) | ✅ (height cache + ref) |
| Implementation | Complex | Simple | Simple (pure Lua) |
| Scroll smoothness | Can jank | Smooth | Smooth |
| Memory (idle, Go) | O(visible) | O(visible) | O(visible) |
| Memory (idle, Lua) | N/A | N/A | O(scrolled) — minor |
