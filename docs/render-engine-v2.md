# Lumina Render Engine v2 — Architecture Design

## 1. Design Goals

- **10,000+ elements** at 60 FPS (4K terminal fullscreen)
- **O(k)** render cost where k = number of dirty components (not total)
- **Zero allocation** in the hot render path (hover, scroll, drag)
- **Layer system** for overlays, drag, sub-windows without full recomposition
- **Preserve Lua API** — `defineComponent`, `useState`, `createElement`, etc.

---

## 2. Architecture Overview

```
┌──────────────────────────────────────────────────────┐
│  Lua Layer: defineComponent / useState / createElement│
│  User writes declarative UI descriptions              │
│  Output: Lua tables (lightweight descriptors)         │
└────────────────────┬─────────────────────────────────┘
                     │ Only called when component is dirty
                     ▼
┌──────────────────────────────────────────────────────┐
│  Reconciler: Lua descriptor → RenderNode update       │
│  Persistent RenderNode tree (not rebuilt each frame)  │
│  Diff descriptor vs existing node → patch in-place    │
│  Mark PaintDirty / LayoutDirty per node               │
└────────────────────┬─────────────────────────────────┘
                     │ Only dirty subtrees
                     ▼
┌──────────────────────────────────────────────────────┐
│  Layout Engine: incremental flex layout               │
│  Only recompute LayoutDirty subtrees                  │
│  Cache (x, y, w, h) on each RenderNode               │
└────────────────────┬─────────────────────────────────┘
                     │ Only paint-dirty nodes
                     ▼
┌──────────────────────────────────────────────────────┐
│  Painter: RenderNode → CellBuffer                     │
│  Each Layer has its own CellBuffer                    │
│  Only write cells for PaintDirty nodes                │
└────────────────────┬─────────────────────────────────┘
                     │ Compose layers, diff vs previous
                     ▼
┌──────────────────────────────────────────────────────┐
│  Compositor + Terminal Output                          │
│  Compose layers bottom-up into final CellBuffer       │
│  Diff final vs previous → minimal ANSI output         │
│  Single write() syscall per frame                     │
└──────────────────────────────────────────────────────┘
```

---

## 3. Core Data Structures

### 3.1 RenderNode — Persistent UI Node

Replaces VNode. Created once, updated in-place. Never garbage collected during normal operation.

```go
type RenderNode struct {
    // === Identity ===
    Type     string   // "box", "text", "hbox", "vbox"
    ID       string   // from props.id
    Key      string   // from props.key (for reconciliation)

    // === Tree (persistent, not rebuilt) ===
    Parent   *RenderNode
    Children []*RenderNode
    
    // === Layout (cached) ===
    X, Y, W, H    int      // absolute position + size (computed by layout engine)
    StyleW, StyleH int     // requested width/height from style (-1 = flex)
    Flex           int     // flex grow factor
    LayoutDirty    bool    // true → recompute this subtree's layout
    
    // === Paint ===
    Content    string    // text content
    Style      Style     // fg, bg, bold, etc.
    PaintDirty bool      // true → repaint this node to CellBuffer
    
    // === Component (if this is a component root) ===
    Component  *Component  // nil for plain elements
    
    // === Events (persistent, not re-registered per frame) ===
    OnClick      LuaRef    // Lua registry ref (0 = none)
    OnMouseEnter LuaRef    // Lua registry ref
    OnMouseLeave LuaRef    // Lua registry ref
    OnChange     LuaRef
    OnKey        LuaRef
    
    // === Object Pool ===
    pooled       bool      // true if returned to pool
}

type LuaRef = int  // Lua registry reference (0 = nil)
```

### 3.2 Component — Stateful UI Unit

Simplified from current. Removes mutex (single-threaded). Removes VNode cache (replaced by RenderNode).

```go
type Component struct {
    // Identity
    ID       string
    Type     string    // factory name
    Name     string
    
    // State
    Props    map[string]any
    State    map[string]any
    Dirty    bool       // needs re-render (not atomic — single thread)
    
    // Hooks (same as current)
    effectHooks       []*EffectHook
    memoHooks         []*MemoHook
    hookIndex         int
    
    // Lua
    renderFn   LuaRef   // Lua registry ref to render function
    
    // Tree
    Parent     *Component
    Children   []*Component
    childMap   map[string]*Component  // "type:key" → child
    
    // Render output
    RootNode   *RenderNode  // the RenderNode subtree this component owns
    
    // Lifecycle
    IsRoot     bool
    Mounted    bool
}
```

### 3.3 CellBuffer — Pre-allocated Terminal Grid

Replaces Frame. Allocated once at startup, never reallocated (except on terminal resize).

```go
type CellBuffer struct {
    cells  []Cell     // flat array: cells[y*width + x]
    width  int
    height int
}

type Cell struct {
    Ch     rune
    FG     uint32    // packed RGB (3 bytes) + flags (1 byte)
    BG     uint32    // packed RGB
    // Bit flags packed into FG high byte:
    //   bit 0: bold
    //   bit 1: dim
    //   bit 2: underline
    //   bit 3: reverse
    //   bit 4: transparent
}

// Set writes a cell. Zero allocation.
func (cb *CellBuffer) Set(x, y int, ch rune, fg, bg uint32) {
    if x >= 0 && x < cb.width && y >= 0 && y < cb.height {
        cb.cells[y*cb.width+x] = Cell{Ch: ch, FG: fg, BG: bg}
    }
}
```

### 3.4 Layer — Independent Rendering Surface

```go
type Layer struct {
    ID       string
    ZIndex   int
    Buffer   *CellBuffer   // pre-allocated
    Root     *RenderNode    // render tree for this layer
    Dirty    bool           // any node in this layer changed
    Visible  bool
    // Bounds (for floating windows/overlays)
    X, Y, W, H int         // position within terminal
}

type Compositor struct {
    layers  []*Layer       // sorted by ZIndex
    final   *CellBuffer    // composed result
    prev    *CellBuffer    // previous frame (for diff)
    output  *bufio.Writer  // terminal stdout
}
```

---

## 4. Render Pipeline (New)

### 4.1 Event → State Change

```
Input event (mouse/key)
  → Hit-test on layer stack (top to bottom)
  → Find target RenderNode
  → Call handler directly (node.OnClick, node.OnMouseEnter, etc.)
  → Handler calls setHovered(true) → Component.SetState
  → Component.Dirty = true
  → Schedule render
```

No EventBus. No bridgeVNodeEvents. No RegisterBridgedHandler. No RegisterFocusable.

### 4.2 Render Dirty Components

```
renderFrame():
  for each dirty Component (not just roots):
    1. Call Lua render function → get Lua table descriptor
    2. Reconcile descriptor against component.RootNode
       - Same type+key → update props/style in-place, mark PaintDirty
       - Different type → replace subtree
       - Added/removed children → insert/remove RenderNodes
    3. If any child size changed → mark LayoutDirty up to nearest fixed-size ancestor
    4. Clear component.Dirty
```

**Key: Only dirty components call Lua. Non-dirty components are completely skipped — no traversal, no checking.**

### 4.3 Incremental Layout

```
layoutDirty():
  Walk RenderNode tree (only LayoutDirty subtrees):
    - Recompute (x, y, w, h) for LayoutDirty nodes and their children
    - Mark affected nodes PaintDirty (position changed)
    - Clear LayoutDirty
```

For hover (color change only): **no layout needed at all.** PaintDirty without LayoutDirty.

### 4.4 Paint Dirty Nodes

```
paintDirty(layer):
  Walk RenderNode tree (only PaintDirty nodes):
    - Write cell content to layer.Buffer at (node.X, node.Y, node.W, node.H)
    - Clear PaintDirty
  layer.Dirty = true
```

### 4.5 Compose + Output

```
compose():
  if no layer is dirty: return (skip frame entirely)
  
  Clear final buffer
  For each visible layer (bottom to top):
    Blit layer.Buffer onto final buffer (skip transparent cells)
  
  Diff final vs prev → write ANSI for changed cells
  Swap final ↔ prev
```

### 4.6 Cost Analysis

| Scenario | Lua PCall | Reconcile | Layout | Paint | Compose | Total |
|----------|-----------|-----------|--------|-------|---------|-------|
| 1 cell hover (10000 cells) | 1 | 3 nodes | 0 | 1 cell | 1 cell diff | **~5 ops** |
| Window drag (60fps) | 0 | 0 | 1 subtree | 1 window | 1 layer blit | **~100 ops** |
| Full screen redraw | N | N subtrees | full | full | full | **O(N)** |

---

## 5. Reconciler Design

The reconciler is the most critical new component. It bridges Lua descriptors and the persistent RenderNode tree.

### 5.1 Lua Descriptor Format

Lua `render()` returns a table:
```lua
{
  type = "box",
  id = "cell-1",
  key = "cell-1",
  style = { background = "#333", width = 1, height = 1 },
  onClick = function() ... end,
  onMouseEnter = function() ... end,
  children = {
    { type = "text", content = "·", style = { foreground = "#888" } }
  }
}
```

### 5.2 Reconcile Algorithm

```go
func reconcile(L *lua.State, node *RenderNode, desc LuaDesc) {
    // 1. Update props
    if desc.Style != node.Style {
        node.Style = desc.Style
        node.PaintDirty = true
    }
    if desc.Content != node.Content {
        node.Content = desc.Content
        node.PaintDirty = true
    }
    
    // 2. Update event handlers (just swap Lua refs, no allocation)
    updateLuaRef(L, &node.OnClick, desc.OnClick)
    updateLuaRef(L, &node.OnMouseEnter, desc.OnMouseEnter)
    updateLuaRef(L, &node.OnMouseLeave, desc.OnMouseLeave)
    
    // 3. Reconcile children (keyed matching)
    reconcileChildren(L, node, desc.Children)
}

func reconcileChildren(L *lua.State, parent *RenderNode, descs []LuaDesc) {
    // Build key→index map for existing children
    oldByKey := map[string]int{}
    for i, child := range parent.Children {
        if child.Key != "" {
            oldByKey[child.Key] = i
        }
    }
    
    // Match new descriptors to existing children
    newChildren := make([]*RenderNode, len(descs))  // reuse slice if same length
    for i, desc := range descs {
        if idx, ok := oldByKey[desc.Key]; ok && parent.Children[idx].Type == desc.Type {
            // Reuse existing node
            newChildren[i] = parent.Children[idx]
            reconcile(L, newChildren[i], desc)
        } else if desc.IsComponent {
            // Component child — find or create Component, render it
            comp := findOrCreateComponent(parent.Component, desc)
            if comp.Dirty || !comp.Mounted {
                renderComponent(L, comp)  // Lua PCall → reconcile recursively
            }
            newChildren[i] = comp.RootNode
        } else {
            // New plain element
            newChildren[i] = createRenderNode(desc)
            newChildren[i].PaintDirty = true
            newChildren[i].LayoutDirty = true
        }
    }
    
    // Cleanup removed children
    // ...
    
    if childrenChanged(parent.Children, newChildren) {
        parent.Children = newChildren
        parent.LayoutDirty = true
    }
}
```

### 5.3 Component Rendering

```go
func renderComponent(L *lua.State, comp *Component) {
    // 1. Set current component (for hooks)
    SetCurrentComponent(comp)
    comp.ResetHookIndex()
    
    // 2. Call Lua render function
    L.RawGetI(lua.RegistryIndex, int64(comp.renderFn))
    pushProps(L, comp.Props)
    status := L.PCall(1, 1, 0)
    if status != lua.OK {
        // error handling
        return
    }
    
    // 3. Read Lua descriptor (don't create Go objects — read directly from Lua stack)
    desc := readLuaDesc(L, -1)
    L.Pop(1)
    
    // 4. Reconcile against existing RenderNode tree
    if comp.RootNode == nil {
        // First mount
        comp.RootNode = createRenderNodeTree(L, desc)
        comp.RootNode.Component = comp
    } else {
        // Update
        reconcile(L, comp.RootNode, desc)
    }
    
    comp.Dirty = false
    comp.Mounted = true
    SetCurrentComponent(nil)
}
```

---

## 6. Event System (New)

### 6.1 Hit-Test Dispatch

```go
func (app *App) dispatchMouse(ev MouseEvent) {
    // Search layers top-to-bottom
    for i := len(app.layers) - 1; i >= 0; i-- {
        layer := app.layers[i]
        if !layer.Visible { continue }
        
        // Translate coordinates to layer-local
        lx, ly := ev.X - layer.X, ev.Y - layer.Y
        if lx < 0 || lx >= layer.W || ly < 0 || ly >= layer.H { continue }
        
        node := hitTest(layer.Root, lx, ly)
        if node == nil { continue }
        
        switch ev.Type {
        case "click":
            if node.OnClick != 0 {
                callLuaRef(app.L, node.OnClick)
                return
            }
        case "mousemove":
            app.updateHover(node)
            return
        }
    }
}

func (app *App) updateHover(newNode *RenderNode) {
    if newNode == app.hoveredNode { return }
    
    old := app.hoveredNode
    app.hoveredNode = newNode
    
    // Direct handler calls — no EventBus, no registration
    if old != nil && old.OnMouseLeave != 0 {
        callLuaRef(app.L, old.OnMouseLeave)
    }
    if newNode != nil && newNode.OnMouseEnter != 0 {
        callLuaRef(app.L, newNode.OnMouseEnter)
    }
}
```

### 6.2 Keyboard Events

```go
func (app *App) dispatchKey(key string) {
    // 1. Global key bindings (lumina.onKey)
    if handler, ok := app.keyBindings[key]; ok {
        callLuaRef(app.L, handler)
        return
    }
    
    // 2. Focused component's onKey handler
    if app.focusedNode != nil && app.focusedNode.OnKey != 0 {
        callLuaRef(app.L, app.focusedNode.OnKey, key)
    }
}
```

---

## 7. What to Keep vs Rewrite

### Keep (copy from current codebase)
- `app.go`: App struct, event loop, PostEvent, terminal setup — **adapt, don't rewrite**
- `hooks.go`: useState, useEffect, useMemo, useCallback, useRef, useReducer — **keep logic, remove locks**
- `store.go`: createStore, useStore, dispatch — **keep as-is**
- `terminal.go`: Terminal I/O, raw mode, mouse tracking — **keep as-is**
- `ansi_adapter.go`: ANSI color/style encoding — **extract color helpers, rewrite output**
- `component.go`: Component struct — **simplify, keep hooks state**

### Rewrite from scratch
- `renderer.go` (1437 lines) → `reconciler.go` (~300 lines)
- `layout.go` (1077 lines) → `layout.go` (~200 lines, incremental)
- `output.go` (Frame/Cell) → `cellbuffer.go` (~100 lines)
- `event.go` (760 lines) → `events.go` (~150 lines, hit-test only)
- `event_bridge.go` → **delete entirely**
- `vdom_diff.go` (569 lines) → **delete entirely**
- `app_render.go` → `render.go` (~150 lines)

### Delete (no longer needed)
- `vdom_diff.go` — no more VNode diff
- `event_bridge.go` — no more per-frame event registration
- `event_pipeline.go` — simplify into direct dispatch
- `concurrent.go` — single-threaded, no SafeMap/SafeSlice needed
- Most of `mcp_debug.go` — rebuild later if needed

---

## 8. Implementation Plan

### Step 1: Core Engine (Day 1)
1. `cellbuffer.go` — CellBuffer + Cell (flat array, pre-allocated)
2. `rendernode.go` — RenderNode struct + tree operations
3. `reconciler.go` — Lua descriptor → RenderNode reconciliation
4. `layout.go` — Incremental flex layout (hbox/vbox/box)
5. `painter.go` — RenderNode → CellBuffer
6. `compositor.go` — Layer composition + ANSI diff output

### Step 2: Lua Bridge (Day 1-2)
7. `lua_api.go` — defineComponent, createElement, mount, run (rewrite)
8. `hooks.go` — useState, useEffect (adapt from current)
9. `store.go` — createStore, useStore (keep)
10. `events.go` — Hit-test dispatch, hover tracking, key bindings

### Step 3: Integration (Day 2)
11. `app.go` — Event loop, terminal setup (adapt)
12. Wire everything together
13. Run `stress_test.lua` — verify 60 FPS at fullscreen

### Step 4: Feature Parity (Day 3+)
14. Overlays, windows, scroll, text input
15. Other examples (counter, todo, dashboard)
16. Test suite adaptation

---

## 9. File Structure (New)

```
pkg/lumina/
  // Core engine
  cellbuffer.go      // CellBuffer, Cell
  rendernode.go      // RenderNode, tree ops
  reconciler.go      // Lua desc → RenderNode update
  layout.go          // Incremental flex layout
  painter.go         // RenderNode → CellBuffer
  compositor.go      // Layer compose + ANSI output
  
  // Lua bridge
  lua_api.go         // Module registration, defineComponent, createElement
  hooks.go           // useState, useEffect, useMemo, etc.
  store.go           // createStore, useStore
  component.go       // Component struct, lifecycle
  
  // Runtime
  app.go             // App, event loop
  events.go          // Hit-test, hover, key dispatch
  terminal.go        // Terminal I/O (keep)
  
  // Features (later)
  overlay.go         // Overlay/popup layer management
  window.go          // Window manager
  scroll.go          // Viewport scrolling
  text_input.go      // Text input handling
```