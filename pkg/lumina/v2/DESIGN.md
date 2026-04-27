# Lumina v2 — Component Buffer Rendering Engine

**Version**: 2.0 (post-review)  
**Status**: Design  
**Purpose**: Replace v1's full-repaint pipeline with an incremental Component Buffer architecture.

---

## §1 Overview

### Problem

Lumina v1 repaints ALL VNodes every frame. With 12,993 cells, a single hover event
triggers `renderVNode` on all 12,993 nodes (O(N)), even though only 1 cell changed.
CPU usage: 13% idle → spikes on mouse movement.

### Solution

Each Component owns a private buffer. Only dirty Components repaint their buffer.
Compositing assembles buffers to the screen using an occlusion map (z-index aware).
Hover on 1 cell = O(1) repaint + O(1) blit.

### Goals

1. **O(1) hover**: Repaint only the dirty Component's buffer, blit only changed cells.
2. **Correct z-index**: Higher layers occlude lower layers for both rendering and events.
3. **Testable without terminal**: JSON output adapter; all tests run headless.
4. **Modular**: Each module has `api.go`, testable in isolation.
5. **Multi-window**: Independent windows with dirty-rect compositing.
6. **Sub-component events**: Multiple interactive VNodes per Component (forms, toolbars).
7. **Focus + keyboard**: Full keyboard navigation, text input support.

---

## §2 Architecture

### Data Flow

```
Lua Script
    │
    ▼
Component.render()  →  VNode tree  →  ComputeLayout(vnode, rect)  →  VNode with (X,Y,W,H)
                                                                          │
                                                                          ▼
                                                               paint(VNode, comp.buffer, offset)
                                                                          │
                                                                          ▼
                                                    compositor.ComposeDirty(dirtyLayers)
                                                                          │
                                                                          ▼
                                                               OutputAdapter.Write(screen)
                                                                  ┌───────┼───────┐
                                                                  │       │       │
                                                                 TUI    JSON    Test
```

### Module Dependency Graph

```
                         ┌──────┐
                         │  app │  (composition root)
                         └──┬───┘
              ┌─────────────┼──────────────┐
              ▼             ▼              ▼
        ┌──────────┐  ┌────────┐    ┌──────────┐
        │component │  │ bridge │    │  output  │
        └────┬─────┘  └───┬────┘    └──────────┘
             │             │
        ┌────┴─────────────┴────┐
        ▼                       ▼
  ┌───────────┐          ┌──────────┐
  │   paint   │          │  event   │
  └─────┬─────┘          └────┬─────┘
        │                     │
        ▼                     ▼
  ┌──────────┐         ┌───────────┐
  │  layout  │         │compositor │
  └────┬─────┘         └─────┬─────┘
       │                     │
       ▼                     ▼
  ┌──────────┐         ┌──────────┐
  │  buffer  │         │  buffer  │
  └──────────┘         └──────────┘
```

### Directory Structure

```
pkg/lumina/v2/
├── DESIGN.md
├── buffer/
│   ├── api.go              # Buffer, Cell, Rect types
│   ├── buffer.go
│   └── buffer_test.go
├── layout/
│   ├── api.go              # VNode, Style, ComputeLayout
│   ├── layout.go
│   └── layout_test.go
├── paint/
│   ├── api.go              # Painter interface
│   ├── paint.go
│   └── paint_test.go
├── compositor/
│   ├── api.go              # Compositor, OcclusionMap
│   ├── compositor.go
│   └── compositor_test.go
├── event/
│   ├── api.go              # Dispatcher, HitTester, Focus
│   ├── event.go
│   └── event_test.go
├── component/
│   ├── api.go              # Component, Manager
│   ├── component.go
│   └── component_test.go
├── bridge/
│   ├── api.go              # Lua↔Go bridge
│   ├── bridge.go
│   └── bridge_test.go
├── output/
│   ├── api.go              # OutputAdapter interface
│   ├── tui.go
│   ├── json.go
│   └── output_test.go
├── app.go                  # Composition root
└── app_test.go             # Integration + scenario tests
```

---

## §3 Module Specifications

### §3.1 buffer — Cell Grid

**Responsibility**: Provide a 2D grid of terminal cells with basic operations.

**Dependencies**: None.

```go
// buffer/api.go

// Cell represents a single terminal character cell. Pure visual data.
type Cell struct {
    Char       rune
    Foreground string // "#rrggbb" or "" (inherit)
    Background string // "#rrggbb" or "" (transparent)
    Bold       bool
    Dim        bool
    Underline  bool
}

// Zero returns true if this cell has never been written to.
func (c Cell) Zero() bool

// Rect represents a rectangular region.
type Rect struct {
    X, Y, W, H int
}

func (r Rect) Contains(px, py int) bool
func (r Rect) Intersect(other Rect) Rect
func (r Rect) Union(other Rect) Rect

// Buffer is a 2D grid of Cells, backed by a flat []Cell slice (single allocation).
type Buffer struct {
    cells  []Cell // flat array, row-major: cells[y*width + x]
    width  int
    height int
}

func New(w, h int) *Buffer
func (b *Buffer) Width() int
func (b *Buffer) Height() int
func (b *Buffer) Get(x, y int) Cell
func (b *Buffer) Set(x, y int, c Cell)
func (b *Buffer) Fill(r Rect, c Cell)
func (b *Buffer) Resize(w, h int)
func (b *Buffer) Clear()

// Blit copies src buffer into dst buffer at offset (dx, dy).
// Only copies non-zero cells (transparent skip).
// clip limits the destination area.
// Returns the dirty rect (area actually written).
func Blit(dst, src *Buffer, dx, dy int, clip Rect) Rect

// Equal returns true if two buffers have identical content.
func Equal(a, b *Buffer) bool
```

**Key design**: Buffer uses flat `[]Cell` (not `[][]Cell`) — single allocation, cache-friendly,
minimal GC pressure for 12,993 tiny buffers.

**Test cases**:
- `TestBuffer_NewAndSize` — New(10, 5) → Width()=10, Height()=5
- `TestBuffer_SetGet` — Set(3,2,cell) → Get(3,2) returns same cell
- `TestBuffer_OutOfBounds` — Get(-1,0) returns zero; Set(100,0,c) is no-op
- `TestBuffer_Fill` — Fill rect → all cells in rect match
- `TestBuffer_Clear` — After Clear(), all cells are zero
- `TestBuffer_Resize_Grow` — Resize larger, old content preserved
- `TestBuffer_Resize_Shrink` — Resize smaller, content clipped
- `TestBuffer_Blit` — Copy 3x2 src to dst at (5,3), verify content
- `TestBuffer_Blit_Clip` — Blit with clip rect, only clipped area written
- `TestBuffer_Blit_Transparent` — Zero cells in src are not copied
- `TestBuffer_Equal` — Identical → true; different → false
- `TestBuffer_FlatBacking` — Verify single allocation (no inner slices)

---

### §3.2 layout — Flexbox Layout Engine

**Responsibility**: Compute absolute positions (X, Y, W, H) for a VNode tree using flexbox semantics.

**Dependencies**: `buffer` (for Rect type).

```go
// layout/api.go

// Style holds layout and visual properties for a VNode.
type Style struct {
    // Sizing
    Width, Height                int // fixed size (0 = auto)
    MinWidth, MaxWidth           int
    MinHeight, MaxHeight         int
    Flex                         int // flex grow factor

    // Spacing
    Padding                      int
    PaddingTop, PaddingBottom    int
    PaddingLeft, PaddingRight    int
    Margin                       int
    MarginTop, MarginBottom      int
    MarginLeft, MarginRight      int
    Gap                          int

    // Alignment
    Justify string // "start", "center", "end", "space-between", "space-around"
    Align   string // "stretch", "start", "center", "end"

    // Visual (used by paint, stored here for convenience)
    Border     string // "none", "single", "double", "rounded"
    Foreground string
    Background string
    Bold, Dim, Underline bool

    // Positioning
    Position string // "", "relative", "absolute", "fixed"
    Top, Left, Right, Bottom int
    ZIndex   int
}

// VNode is a virtual DOM node — a drawing instruction.
type VNode struct {
    Type     string         // "box", "vbox", "hbox", "text", "input", "textarea", "fragment"
    ID       string         // unique identifier (for events, hit-test)
    Props    map[string]any
    Style    Style
    Children []*VNode
    Content  string         // for text nodes

    // Layout results (set by ComputeLayout)
    X, Y, W, H int
}

// ComputeLayout computes absolute positions for the VNode tree.
// The root is positioned at (x, y) with available size (w, h).
// All descendant positions are set as absolute screen coordinates.
//
// For a Component's VNode tree, call:
//   ComputeLayout(comp.VNodeTree, comp.Rect.X, comp.Rect.Y, comp.Rect.W, comp.Rect.H)
// This positions the VNode subtree within the component's screen rect.
// Paint then translates back to buffer-local via offsetX/offsetY.
func ComputeLayout(root *VNode, x, y, w, h int)
```

**Coordinate system clarification**:

ComputeLayout ALWAYS sets absolute screen coordinates on VNodes. There is no
separate "local layout" function. When a Component re-renders:

1. `comp.render()` → fresh VNode tree (all X=Y=W=H=0)
2. `ComputeLayout(vnodeTree, comp.Rect.X, comp.Rect.Y, comp.Rect.W, comp.Rect.H)`
   → sets absolute positions within the component's screen rect
3. `Paint(comp.Buffer, vnodeTree, comp.Rect.X, comp.Rect.Y)`
   → translates absolute → buffer-local: `bufX = vnode.X - offsetX`

This is simple and correct. No "local layout" magic needed.

**Test cases**:
- `TestLayout_SingleBox` — box with width=10, height=5 → X=0,Y=0,W=10,H=5
- `TestLayout_VBox_EqualFlex` — vbox with 3 children flex=1 in 30×12 → each 30×4
- `TestLayout_HBox_EqualFlex` — hbox with 3 children flex=1 in 30×12 → each 10×12
- `TestLayout_VBox_FixedAndFlex` — child1 height=2, child2 flex=1 in 10×10 → child2 gets 8
- `TestLayout_Text_Wrap` — text "hello world" in width=5 → height=2
- `TestLayout_Padding` — box with padding=1, child gets reduced space
- `TestLayout_Gap` — vbox gap=1, 3 children → 2 gaps inserted
- `TestLayout_Absolute` — absolute child at top=2, left=3 → X=3, Y=2
- `TestLayout_Nested` — vbox > hbox > box → correct nested positions
- `TestLayout_ComponentRect` — layout at offset (10, 5) → all positions offset by (10, 5)

---

### §3.3 paint — VNode to Buffer Painting

**Responsibility**: Paint a VNode tree into a Buffer. Handles text rendering,
background fill, border drawing.

**Dependencies**: `buffer`, `layout` (for VNode, Style types).

```go
// paint/api.go

// Painter paints a VNode tree into a Buffer.
type Painter interface {
    // Paint renders the VNode tree into the given buffer.
    // The VNode must have layout computed (X, Y, W, H set as absolute screen coords).
    // offsetX, offsetY translate absolute VNode coords to buffer-local coords:
    //   bufferX = vnode.X - offsetX
    //   bufferY = vnode.Y - offsetY
    //
    // For a Component at screen position (10, 5):
    //   Paint(comp.Buffer, comp.VNodeTree, 10, 5)
    // A VNode at absolute (12, 7) paints to buffer position (2, 2).
    Paint(buf *buffer.Buffer, root *layout.VNode, offsetX, offsetY int)
}

func NewPainter() Painter
```

**Algorithm**: Ported from v1 `renderVNode`. Walks VNode tree recursively:
- `box`/`vbox`/`hbox`: fill background, draw border, recurse children
- `text`: render text content with word wrap, foreground/background colors
- `input`/`textarea`: render input content with cursor
- `fragment`: just recurse children (no visual output)

**Test cases**:
- `TestPaint_Box_Background` — box with bg="#FF0000" → all cells have bg="#FF0000"
- `TestPaint_Text_Simple` — text "Hi" → cells[0]='H', cells[1]='i'
- `TestPaint_Text_Foreground` — text with fg="#00FF00" → cells have correct fg
- `TestPaint_Box_Border_Single` — border="single" → ┌┐└┘─│
- `TestPaint_Box_Border_Rounded` — border="rounded" → ╭╮╰╯─│
- `TestPaint_Nested` — box > text → text painted inside box area
- `TestPaint_VBox_Children` — vbox with 2 text children → stacked vertically
- `TestPaint_Offset` — VNode at abs (10,5), offset=(10,5) → painted at buffer (0,0)
- `TestPaint_Clip` — VNode extends beyond buffer → clipped, no panic

---

### §3.4 compositor — Buffer Compositing + Occlusion

**Responsibility**: Compose multiple Component buffers into a screen buffer using
z-index ordering. Maintain an occlusion map for hit-testing and selective blitting.

**Dependencies**: `buffer`.

```go
// compositor/api.go

// Layer represents a Component's buffer positioned on screen.
type Layer struct {
    ID       string         // unique component/window ID
    Buffer   *buffer.Buffer
    Rect     buffer.Rect    // screen position
    ZIndex   int
    DirtyRect *buffer.Rect  // sub-region that changed (nil = entire buffer dirty)
}

// OcclusionMap maps each screen cell to the Layer that owns it (highest z-index).
type OcclusionMap struct { ... }

func NewOcclusionMap(w, h int) *OcclusionMap

// Build rebuilds the occlusion map from layers.
// Layers are processed from highest z-index to lowest.
// Caches sorted order; only re-sorts when layer set changes.
func (om *OcclusionMap) Build(layers []*Layer)

// Owner returns the layer ID that owns cell (x, y). Empty if no owner.
func (om *OcclusionMap) Owner(x, y int) string

// OwnerLayer returns the Layer that owns cell (x, y). Nil if no owner.
func (om *OcclusionMap) OwnerLayer(x, y int) *Layer

// Compositor composes layers into a screen buffer.
type Compositor struct { ... }

func NewCompositor(w, h int) *Compositor

// SetLayers sets the full layer stack. Rebuilds the occlusion map.
func (c *Compositor) SetLayers(layers []*Layer)

// ComposeAll composes all layers into the screen buffer. Returns screen.
func (c *Compositor) ComposeAll() *buffer.Buffer

// ComposeDirty composes only changed regions.
// For each dirty layer:
//   - If layer.DirtyRect is set, only blit that sub-region
//   - Otherwise blit the entire layer rect
//   - Only write cells where occlusionMap says this layer owns the cell
// Returns dirty rects on screen.
func (c *Compositor) ComposeDirty(dirtyLayers []*Layer) []buffer.Rect

// ComposeRects recomposes specific screen regions (used after window move).
// For each cell in the rects, look up occlusionMap → blit from owning layer.
func (c *Compositor) ComposeRects(rects []buffer.Rect) []buffer.Rect

// Screen returns the current screen buffer.
func (c *Compositor) Screen() *buffer.Buffer

// OcclusionMap returns the current occlusion map.
func (c *Compositor) OcclusionMap() *OcclusionMap
```

**Algorithm — ComposeDirty (hot path)**:
```
for each dirtyLayer:
    rect := dirtyLayer.DirtyRect or dirtyLayer.Rect  // use sub-rect if available
    for y := rect.Y; y < rect.Y + rect.H; y++:
        for x := rect.X; x < rect.X + rect.W; x++:
            if occlusionMap.Owner(x, y) == dirtyLayer.ID:
                localX := x - dirtyLayer.Rect.X
                localY := y - dirtyLayer.Rect.Y
                screen.Set(x, y, dirtyLayer.Buffer.Get(localX, localY))
    dirtyRects = append(dirtyRects, rect)
```

**Algorithm — Build occlusion map**:
```
sort layers by ZIndex descending (cached, only re-sort on layer set change)
clear map
for each layer:
    for y := layer.Rect.Y; y < layer.Rect.Y + layer.Rect.H; y++:
        for x := layer.Rect.X; x < layer.Rect.X + layer.Rect.W; x++:
            if map[y][x] == nil:
                map[y][x] = layer
```

**Test cases**:
- `TestOcclusion_SingleLayer` — 1 layer → all cells owned by it
- `TestOcclusion_TwoLayers_Overlap` — A(z=0) full, B(z=100) partial → correct ownership
- `TestOcclusion_ThreeLayers` — A(z=0), B(z=100), C(z=200) → correct ownership
- `TestOcclusion_NoOverlap` — two layers side by side → each owns its rect
- `TestCompositor_ComposeAll` — 2 layers → correct screen content
- `TestCompositor_ComposeDirty_SingleCell` — 1 cell dirty → only that cell updated
- `TestCompositor_ComposeDirty_SubRect` — DirtyRect set → only sub-region blitted
- `TestCompositor_ComposeDirty_Occluded` — dirty layer occluded → screen unchanged there
- `TestCompositor_ComposeRects_WindowMove` — window moves, old+new rects recomposed
- `TestCompositor_ManyLayers` — 100 layers → correct compositing
- `TestCompositor_TransparentCells` — zero cells → lower layer shows through
- `TestCompositor_CachedSort` — add/remove layer → re-sort; no change → no re-sort

---

### §3.5 event — Event Dispatch + Focus

**Responsibility**: Dispatch input events to the correct handler using the occlusion map
for hit-testing. Manage focus state for keyboard navigation.

**Dependencies**: `buffer` (Rect), `layout` (VNode for sub-component hit-test), `compositor` (OcclusionMap).

**Key design decision**: Events are dispatched to **VNode IDs** (not Component IDs).
A single Component can have multiple interactive VNodes (e.g., a form with buttons).
The occlusion map resolves screen (x,y) → Layer (Component), then within the Component,
the VNode's rect determines which VNode was hit.

```go
// event/api.go

// Event represents an input event.
type Event struct {
    Type      string // "mousedown", "mouseup", "mousemove", "mouseenter",
                     // "mouseleave", "click", "keydown", "keyup"
    X, Y      int    // mouse position (screen coordinates)
    LocalX    int    // mouse position relative to target VNode
    LocalY    int
    Key       string // key name for keyboard events
    Target    string // VNode ID that should handle this event
    Bubbles   bool
    Timestamp int64
}

// EventHandler is a function that handles an event.
type EventHandler func(e *Event)

// HandlerMap maps event types to handlers.
// Key = "click", "mouseenter", "mouseleave", "keydown", etc.
type HandlerMap map[string]EventHandler

// HitTester resolves screen coordinates to a target ID.
type HitTester interface {
    // HitTest returns the VNode ID at screen position (x, y).
    // First finds the owning Layer via occlusion map, then finds
    // the deepest VNode within that layer's VNode tree that contains (x, y).
    HitTest(x, y int) string
}

// VNodeHitTester implements HitTester using occlusion map + VNode tree walk.
type VNodeHitTester struct { ... }

// NewVNodeHitTester creates a hit tester.
// layers: component layers (each has a VNode tree for sub-component hit-test)
// om: occlusion map for layer-level hit-test
func NewVNodeHitTester(layers []*ComponentLayer, om *compositor.OcclusionMap) *VNodeHitTester

// ComponentLayer extends compositor.Layer with VNode tree for sub-component events.
type ComponentLayer struct {
    compositor.Layer
    VNodeTree *layout.VNode // for sub-component hit-test within this layer
}

// Dispatcher dispatches events to handlers.
type Dispatcher struct { ... }

func NewDispatcher() *Dispatcher

// SetHitTester sets the hit-test provider.
func (d *Dispatcher) SetHitTester(ht HitTester)

// RegisterHandlers registers event handlers for a VNode ID.
func (d *Dispatcher) RegisterHandlers(vnodeID string, handlers HandlerMap)

// UnregisterHandlers removes all handlers for a VNode ID.
func (d *Dispatcher) UnregisterHandlers(vnodeID string)

// Dispatch processes an input event.
// Mouse events: HitTest → target → synthesize enter/leave → call handler → bubble.
// Key events: dispatch to focused VNode ID.
func (d *Dispatcher) Dispatch(e *Event)

// HoveredID returns the currently hovered VNode ID.
func (d *Dispatcher) HoveredID() string

// --- Focus Management ---

// SetFocus sets the focused VNode ID.
func (d *Dispatcher) SetFocus(vnodeID string)

// FocusedID returns the currently focused VNode ID.
func (d *Dispatcher) FocusedID() string

// FocusNext moves focus to the next focusable VNode (Tab).
func (d *Dispatcher) FocusNext()

// FocusPrev moves focus to the previous focusable VNode (Shift+Tab).
func (d *Dispatcher) FocusPrev()

// RegisterFocusable registers a VNode as focusable (in tab order).
func (d *Dispatcher) RegisterFocusable(vnodeID string, tabIndex int)

// UnregisterFocusable removes a VNode from the focus chain.
func (d *Dispatcher) UnregisterFocusable(vnodeID string)
```

**Algorithm — Dispatch mouse event**:
```
targetID := hitTester.HitTest(e.X, e.Y)
e.Target = targetID

if e.Type == "mousemove":
    if targetID != d.hoveredID:
        if d.hoveredID != "":
            emit("mouseleave", d.hoveredID)
        if targetID != "":
            emit("mouseenter", targetID)
        d.hoveredID = targetID

if e.Type == "mousedown":
    emit("click", targetID)
    // Set focus to target if it's focusable
    if isFocusable(targetID):
        d.SetFocus(targetID)

if e.Type == "keydown":
    if e.Key == "Tab":
        d.FocusNext()
    elif e.Key == "Shift+Tab":
        d.FocusPrev()
    else:
        emit("keydown", d.FocusedID())

func emit(eventType, targetID):
    // Direct handler
    if handler := d.handlers[targetID][eventType]; handler != nil:
        handler(e)
    // Bubble to parent (if Bubbles=true)
    if e.Bubbles:
        parentID := d.parentMap[targetID]
        if parentID != "":
            emit(eventType, parentID)
```

**Algorithm — VNodeHitTester.HitTest(x, y)**:
```
// Step 1: Layer-level hit-test via occlusion map
layer := om.OwnerLayer(x, y)
if layer == nil: return ""

// Step 2: Within the layer's VNode tree, find deepest VNode at (x, y) with an ID
func findDeepest(vnode, x, y) string:
    if !vnode.Rect().Contains(x, y): return ""
    // Check children in reverse order (last child = highest z)
    for i := len(vnode.Children)-1; i >= 0; i--:
        if id := findDeepest(vnode.Children[i], x, y); id != "":
            return id
    if vnode.ID != "":
        return vnode.ID
    return ""

return findDeepest(layer.VNodeTree, x, y)
```

**Note on sub-component hit-test**: For a 1×1 Cell component, the VNode tree walk
is trivial (2 nodes: box + text). For a larger component (form with buttons),
the walk is O(VNodes in that component) — typically small (<50). The expensive
layer-level occlusion lookup is O(1).

**Test cases**:
- `TestDispatcher_HitTest_Click` — click at (5,3) → correct VNode receives "click"
- `TestDispatcher_HoverEnterLeave` — move A→B → A gets "mouseleave", B gets "mouseenter"
- `TestDispatcher_HoverReenter` — leave A, enter B, return A → A gets "mouseenter" again
- `TestDispatcher_ClickOccluded` — click occluded area → only top layer's VNode handles
- `TestDispatcher_SubComponentHitTest` — form with 2 buttons → correct button gets click
- `TestDispatcher_EventBubbling` — child click bubbles to parent handler
- `TestDispatcher_NoHandler` — click on VNode without handler → no panic
- `TestDispatcher_Focus_Tab` — Tab cycles through focusable VNodes
- `TestDispatcher_Focus_ShiftTab` — Shift+Tab cycles backwards
- `TestDispatcher_Focus_Click` — click focusable VNode → receives focus
- `TestDispatcher_Keyboard_ToFocused` — keydown dispatched to focused VNode

---

### §3.6 component — Component Lifecycle

**Responsibility**: Manage Component instances — state, hooks, buffer ownership,
dirty tracking, and the render cycle.

**Dependencies**: `buffer`, `layout`, `paint`, `event`.

```go
// component/api.go

// RenderFunc is the component's render function.
// Called with (state, props), returns a VNode tree.
// In Go tests, this is a Go function.
// In production, the bridge module wraps Lua render functions as RenderFunc.
type RenderFunc func(state map[string]any, props map[string]any) *layout.VNode

// Component represents a stateful rendering unit.
type Component struct {
    ID          string
    Name        string
    Buffer      *buffer.Buffer      // private buffer (W×H of this component)
    Rect        buffer.Rect         // screen absolute position
    PrevRect    buffer.Rect         // previous rect (for change detection)
    ZIndex      int
    DirtyPaint  bool                // buffer needs repaint
    RectChanged bool                // rect changed → needs recompose + occlusionMap rebuild

    State       map[string]any      // component state (for hooks)
    Props       map[string]any      // props from parent
    RenderFn    RenderFunc          // render function
    VNodeTree   *layout.VNode       // last render result

    Parent      *Component
    Children    []*Component        // child component instances
    ChildMap    map[string]*Component // keyed lookup for reconciliation

    // Event handlers extracted from VNode tree after render
    Handlers    map[string]event.HandlerMap // vnodeID → {eventType → handler}
    Focusables  []string                    // VNode IDs that are focusable (tab order)
}

// Manager manages the component tree.
type Manager struct { ... }

func NewManager() *Manager

// Register adds a component to the tree.
func (m *Manager) Register(comp *Component)

// Unregister removes a component and cleans up (lifecycle: unmount).
func (m *Manager) Unregister(id string)

// Get returns a component by ID.
func (m *Manager) Get(id string) *Component

// SetState updates a component's state and marks it dirty.
func (m *Manager) SetState(compID string, key string, value any)

// GetDirtyPaint returns all components with DirtyPaint=true.
func (m *Manager) GetDirtyPaint() []*Component

// GetRectChanged returns all components with RectChanged=true.
func (m *Manager) GetRectChanged() []*Component

// ClearDirty clears DirtyPaint and RectChanged on all components.
func (m *Manager) ClearDirty()

// AllLayers returns all components as compositor Layers.
func (m *Manager) AllLayers() []*event.ComponentLayer

// Reconcile reconciles a parent's child list after re-render.
// newChildren: child descriptors from the new VNode tree.
// Creates new Components for new children, removes old ones, updates existing.
func (m *Manager) Reconcile(parent *Component, newChildren []ChildDescriptor)

// ChildDescriptor describes a child component to create/update.
type ChildDescriptor struct {
    Key       string         // stable key for reconciliation
    Name      string         // component type name
    Props     map[string]any
    RenderFn  RenderFunc
}
```

**Render cycle for a dirty component**:
```
1. comp.DirtyPaint = true (from SetState or props change)
2. comp.VNodeTree = comp.RenderFn(comp.State, comp.Props)
3. ComputeLayout(comp.VNodeTree, comp.Rect.X, comp.Rect.Y, comp.Rect.W, comp.Rect.H)
4. Check: did comp need a new rect? (e.g., content grew)
   - If component has fixed size (from parent layout): rect unchanged
   - If component size depends on content: check if VNodeTree root size differs
   - Rect changed → comp.RectChanged = true → parent notified
5. painter.Paint(comp.Buffer, comp.VNodeTree, comp.Rect.X, comp.Rect.Y)
6. Extract event handlers from VNode tree:
   - Walk VNodeTree, collect onClick/onMouseEnter/etc from VNode.Props
   - Store in comp.Handlers[vnodeID] = HandlerMap
   - Register with event.Dispatcher
7. comp.DirtyPaint = false
```

**Reconciliation algorithm** (with props shallow comparison):
```
func Reconcile(parent, newChildren):
    oldMap := parent.ChildMap
    newMap := {}
    
    for each child in newChildren:
        if existing := oldMap[child.Key]; existing != nil:
            // Update existing: only mark dirty if props actually changed
            if !shallowEqual(existing.Props, child.Props):
                existing.Props = child.Props
                existing.DirtyPaint = true
            // else: props unchanged → skip child render entirely
            newMap[child.Key] = existing
        else:
            // Create new component
            comp := NewComponent(child)
            comp.Parent = parent
            newMap[child.Key] = comp
            Register(comp)
    
    // Remove components no longer in children
    for key, old := range oldMap:
        if newMap[key] == nil:
            Unregister(old.ID)  // cleanup: release Lua refs, remove handlers
    
    parent.Children = newMap values (ordered)
    parent.ChildMap = newMap

// shallowEqual compares two prop maps.
// Returns true if same keys with identical values (== comparison).
// Function values are compared by reference (pointer equality).
func shallowEqual(a, b map[string]any) bool:
    if len(a) != len(b): return false
    for k, va := range a:
        vb, ok := b[k]
        if !ok || va != vb: return false
    return true
```

**§3.6.1 Stable Key Best Practice**:

Keys should use data-inherent stable IDs (e.g., database ID), NOT array indices.
Props should only contain data that affects visual output — no index or position metadata.

```lua
-- ❌ WRONG: key depends on array order, props include non-visual data
createElement("Button", { key = tostring(i), text = item.text, index = i })

-- ✅ CORRECT: key uses data ID, props only contain visual data
createElement("Button", { key = item.id, text = item.text })
```

**Effect of stable keys** (delete 1 of 100 buttons):

| | key=index | key=data.id |
|--|-----------|-------------|
| Parent Lua render | O(N) VNode creation | O(N) VNode creation (unavoidable) |
| Child Lua render | O(N) all re-render | **O(0) zero re-render** |
| Child paint | O(N) all re-paint | **O(0) zero re-paint** |
| Rect update | O(N) | O(N) |
| Compositing | O(affected_area) | O(affected_area) |

With stable keys + props shallow comparison, removing/adding a child only costs:
- Parent render (creates VNode descriptors)
- Reconcile (matches keys, skips unchanged children)
- Layout update (positions shift for remaining children)
- Compositor recompose (affected area only)

No child Lua render is called unless its props actually changed.

**Test cases**:
- `TestComponent_SetState_MarksDirty` — SetState → DirtyPaint=true
- `TestComponent_Render_UpdatesBuffer` — render → buffer has new content
- `TestComponent_RectUnchanged_NoRectChanged` — content change → RectChanged=false
- `TestComponent_RectChanged` — size change → RectChanged=true
- `TestComponent_GetDirtyPaint` — 3 comps, 1 dirty → returns 1
- `TestComponent_ParentChild` — parent with 2 children → correct tree
- `TestComponent_AllLayers` — components → layers with correct rect/z
- `TestComponent_Reconcile_AddChild` — new child → created and registered
- `TestComponent_Reconcile_RemoveChild` — missing child → unregistered
- `TestComponent_Reconcile_UpdateChild` — existing child → props updated, marked dirty
- `TestComponent_Reconcile_StableKeys` — keyed children reorder → same instances
- `TestComponent_ExtractHandlers` — VNode with onClick → handler registered

---

### §3.7 output — Output Adapters

**Responsibility**: Write the screen buffer to various outputs.

**Dependencies**: `buffer`.

```go
// output/api.go

// Adapter writes screen content to an output destination.
type Adapter interface {
    // WriteFull writes the entire screen buffer.
    WriteFull(screen *buffer.Buffer) error

    // WriteDirty writes only the changed regions.
    WriteDirty(screen *buffer.Buffer, dirtyRects []buffer.Rect) error

    // Flush flushes buffered output.
    Flush() error

    // Close closes the adapter.
    Close() error
}

// --- TUI Adapter (ANSI terminal) ---
func NewTUIAdapter(w io.Writer) Adapter

// --- JSON Adapter (testing + WebUI) ---

type RenderResult struct {
    Width      int          `json:"width"`
    Height     int          `json:"height"`
    Cells      [][]CellJSON `json:"cells"`
    DirtyRects []RectJSON   `json:"dirty_rects"`
}

type CellJSON struct {
    Char string `json:"char"`
    Fg   string `json:"fg,omitempty"`
    Bg   string `json:"bg,omitempty"`
    Bold bool   `json:"bold,omitempty"`
}

type RectJSON struct {
    X, Y, W, H int `json:"x,y,w,h"`
}

func NewJSONAdapter(w io.Writer) Adapter

// --- Test Adapter (direct assertion) ---

type TestAdapter struct {
    LastScreen *buffer.Buffer
    DirtyRects []buffer.Rect
    WriteCount int
}

func NewTestAdapter() *TestAdapter
```

**Test cases**:
- `TestTUIAdapter_WriteDirty` — 1 dirty rect → only that region's ANSI emitted
- `TestJSONAdapter_WriteFull` — full screen → valid JSON with correct cells
- `TestJSONAdapter_DirtyRects` — dirty rects included in JSON
- `TestTestAdapter_Captures` — WriteFull → LastScreen accessible

---

### §3.8 bridge — Lua Integration

**Responsibility**: Bridge between Lua render functions and Go Component system.
Convert Lua tables to VNode trees, manage Lua state references, wire hooks.

**Dependencies**: `layout` (VNode), `component` (RenderFunc), `go-lua`.

```go
// bridge/api.go

// Bridge connects Lua scripts to the Go component system.
type Bridge struct { ... }

// NewBridge creates a bridge with the given Lua state.
func NewBridge(L *lua.State) *Bridge

// WrapRenderFn wraps a Lua render function as a Go RenderFunc.
// The Lua function is stored as a registry ref.
// When called, it pushes state+props to Lua, calls the function,
// and converts the returned Lua table to a VNode tree.
func (b *Bridge) WrapRenderFn(luaFuncRef int) component.RenderFunc

// LuaTableToVNode converts a Lua table (on stack) to a VNode tree.
// Handles: type, id, style, props, children, content.
// For type="component" children: calls luaComponentToVNode (reconciliation).
func (b *Bridge) LuaTableToVNode(L *lua.State, idx int) *layout.VNode

// ExtractHandlers extracts event handler functions from VNode Props.
// Returns handlers map: vnodeID → {eventType → handler}.
// Handler functions are Lua function refs wrapped as Go EventHandlers.
func (b *Bridge) ExtractHandlers(root *layout.VNode) map[string]event.HandlerMap

// RegisterHooks registers Lua-callable hooks: useState, useEffect, useStore, etc.
func (b *Bridge) RegisterHooks(L *lua.State)

// ReleaseRefs releases all Lua registry refs from the previous render cycle.
func (b *Bridge) ReleaseRefs()
```

**Note**: This module is ported from v1's `renderer.go` (LuaVNodeToVNode,
luaComponentToVNode), `hooks.go` (useState, useEffect, useStore), and
`event_bridge.go` (walkVNodeForEvents). The key difference: v1 registered events
via a global EventBus walk; v2 extracts handlers from VNode Props and registers
them per-component via the event.Dispatcher.

**This module is implemented in Phase 8** (after all Go-native modules are tested).
Phases 1-7 use Go RenderFunc directly for testing.

**Test cases** (using Go mock Lua state or real go-lua):
- `TestBridge_LuaTableToVNode` — Lua table → correct VNode tree
- `TestBridge_WrapRenderFn` — Lua render function → callable RenderFunc
- `TestBridge_ExtractHandlers` — VNode with onClick → handler map
- `TestBridge_UseState` — useState hook → state management works
- `TestBridge_ReleaseRefs` — refs released after render cycle

---

## §4 Data Flow Scenarios

### Scenario 1: Cell Hover (O(1) hot path)

```
1. Terminal sends mousemove(x=5, y=3)
2. event.Dispatcher.Dispatch(mousemove)
   → hitTester.HitTest(5, 3)
     → occlusionMap[3][5] → Layer "cell-5-3"
     → VNode tree walk (2 nodes: box + text) → VNode ID "5,3"
   → hoveredID changed: "4,3" → "5,3"
   → synthesize mouseleave for "4,3" → handler: setState("hovered", false)
   → synthesize mouseenter for "5,3" → handler: setState("hovered", true)

3. cell-4-3.DirtyPaint = true
   cell-5-3.DirtyPaint = true

4. Render cycle (next tick):
   → GetDirtyPaint() returns [cell-4-3, cell-5-3]
   → For each:
     a. render() → new VNode tree (1 box + 1 text, new color)
     b. ComputeLayout(vnode, rect.X, rect.Y, 1, 1)
     c. Paint(buffer, vnode, rect.X, rect.Y) → 1 cell updated
   → compositor.ComposeDirty([cell-4-3-layer, cell-5-3-layer])
     → each layer: DirtyRect = {0,0,1,1}, blit 1 cell (occlusion check)
     → returns dirtyRects = [{4,3,1,1}, {5,3,1,1}]
   → outputAdapter.WriteDirty(screen, dirtyRects)
```

**Total work**: 2 Lua PCall + 2×ComputeLayout(2 nodes) + 2×Paint(1 cell) + 2×blit(1 cell).

### Scenario 2: Click Cell

```
1. mousedown(5, 3) → hitTest → "5,3" → handler: store.dispatch("clickCell")
2. Store: clickCount++ → Grid selector fires → Grid.DirtyPaint = true
3. Grid.render() → new VNode tree (vbox > hbox > createElement(Cell) × N)
4. Reconcile Grid's children:
   - Cell "5,3": props.clicked changed → DirtyPaint = true
   - Other cells: props unchanged → skip
5. Cell "5,3" render + paint (1 cell)
6. StatusBar.render() + paint (1 line)
7. ComposeDirty → blit dirty cells + status bar
8. outputAdapter.WriteDirty
```

### Scenario 3: Window Move

```
1. Dialog moves from (10,5) to (20,8)
2. Dialog.Rect = {20,8,30,10}, Dialog.PrevRect = {10,5,30,10}
3. Dialog.RectChanged = true
4. compositor.SetLayers(updated) → rebuild occlusionMap
5. compositor.ComposeRects([oldRect, newRect])
   → each cell: owner = occlusionMap → blit from owner's buffer
6. outputAdapter.WriteDirty(screen, dirtyRects)
```

### Scenario 4: Form with Multiple Buttons (sub-component events)

```
Component "form" renders:
  vbox
    ├── input#email (focusable)
    ├── input#password (focusable)
    ├── button#ok {onClick: submit}
    └── button#cancel {onClick: close}

Hit-test at button#ok position:
  → occlusionMap → Layer "form"
  → VNode tree walk in "form" → deepest VNode at (x,y) with ID → "ok"
  → dispatch "click" to "ok" → submit handler called

Tab navigation:
  → Tab → FocusNext() → email → password → ok → cancel → email...
```

---

## §5 Testing Strategy

### Principle: No Terminal Required

All tests use `TestAdapter`. Tests assert on `buffer.Buffer` content directly.

### Test Levels

| Level | What | Example |
|-------|------|---------|
| Unit | Single module | `TestBuffer_Blit`, `TestOcclusion_TwoLayers` |
| Integration | Multiple modules | `TestPipeline_HoverUpdatesScreen` |
| Scenario | Full app lifecycle | `TestScenario_MultiWindowOverlap` |

### Integration Test Examples

```go
func TestScenario_CellHover(t *testing.T) {
    app := NewTestApp(20, 10)
    app.MountGrid(20, 9)  // Go render functions, no Lua
    app.RenderAll()
    
    screen := app.Screen()
    assert.Equal(t, '·', screen.Get(5, 3).Char)
    
    app.HandleEvent(&Event{Type: "mousemove", X: 5, Y: 3})
    app.RenderDirty()
    
    assert.Equal(t, '█', app.Screen().Get(5, 3).Char)
    assert.Equal(t, []Rect{{5,3,1,1}}, app.DirtyRects())
}

func TestScenario_DialogOccludesGrid(t *testing.T) {
    app := NewTestApp(40, 20)
    app.MountGrid(40, 19)
    app.OpenDialog("dlg1", Rect{10,5,20,10}, 100)
    app.RenderAll()
    
    // Dialog area shows dialog content
    assert.Equal(t, 'D', app.Screen().Get(15, 8).Char)
    // Grid visible outside dialog
    assert.Equal(t, '·', app.Screen().Get(5, 3).Char)
    
    // Click in dialog → dialog handles, not grid
    app.HandleEvent(&Event{Type: "mousedown", X: 15, Y: 8})
    assert.Equal(t, "dlg1-ok", app.LastClickTarget())
}

func TestScenario_WindowMoveReveals(t *testing.T) {
    app := NewTestApp(40, 20)
    app.MountGrid(40, 20)
    app.OpenDialog("dlg1", Rect{5,5,15,8}, 100)
    app.RenderAll()
    
    assert.Equal(t, 'D', app.Screen().Get(10, 8).Char) // occluded by dialog
    
    app.MoveDialog("dlg1", Rect{25,5,15,8})
    app.RenderDirty()
    
    assert.Equal(t, '·', app.Screen().Get(10, 8).Char) // grid visible again
}

func TestScenario_FormTabNavigation(t *testing.T) {
    app := NewTestApp(40, 20)
    app.MountForm()  // form with email, password, ok, cancel
    app.RenderAll()
    
    // Tab through focusable elements
    app.HandleEvent(&Event{Type: "keydown", Key: "Tab"})
    assert.Equal(t, "email", app.FocusedID())
    
    app.HandleEvent(&Event{Type: "keydown", Key: "Tab"})
    assert.Equal(t, "password", app.FocusedID())
    
    app.HandleEvent(&Event{Type: "keydown", Key: "Tab"})
    assert.Equal(t, "ok", app.FocusedID())
}
```

---

## §6 Concurrency Model

**Single-threaded render loop** (same as v1):

```
Main goroutine:
    for {
        select {
        case event := <-inputCh:
            dispatcher.Dispatch(event)    // may trigger setState → dirty
        case <-ticker.C:
            renderDirty()                 // process all dirty components
            outputAdapter.WriteDirty(...)
        }
    }
```

All Component state, Lua state, and rendering happen on the main goroutine.
No locks needed. `DirtyPaint` can be a simple `bool` (not `atomic.Bool`).

Input parsing runs on a separate goroutine and sends events via channel.

---

## §7 Implementation Order

| Phase | Module | Dependencies | Est. Tests | Est. Lines |
|-------|--------|-------------|------------|------------|
| 1 | `buffer` | none | 12 | ~200 |
| 2 | `layout` | buffer | 10 | ~500 (port from v1) |
| 3 | `paint` | buffer, layout | 9 | ~400 (port from v1) |
| 4 | `compositor` | buffer | 12 | ~300 |
| 5 | `output` | buffer | 4 | ~200 |
| 6 | `event` | buffer, compositor | 11 | ~300 |
| 7 | `component` | buffer, layout, paint, event | 12 | ~400 |
| 8 | `bridge` + `app` | all | 10+ | ~500 (port from v1) |

Each phase = one builder fork. Each fork reads only the relevant `api.go` files.
Each phase committed independently with all tests passing.

---

## §8 Migration from v1

### Reusable Code (port, don't rewrite)

| v1 File | v2 Module | What to port |
|---------|-----------|-------------|
| `layout.go` | `layout/` | `computeFlexLayout` algorithm |
| `renderer.go` (renderVNode) | `paint/` | VNode painting logic |
| `output.go` (ANSIAdapter) | `output/` | ANSI diff output |
| `hooks.go` | `bridge/` | useState, useEffect, useStore |
| `renderer.go` (LuaVNodeToVNode) | `bridge/` | Lua table → VNode conversion |

### Replaced by new architecture

| v1 Code | Replaced By |
|---------|-------------|
| `bridgeVNodeEvents` + `VNodeTree.Parents` | `event.Dispatcher` + `occlusionMap` + VNode tree walk |
| `renderAllDirty` + `renderDirtyChildren` | Component buffer + `compositor.ComposeDirty` |
| `reRenderDirtySubtree` (in-place VNode update) | Component-level dirty + reconciliation |
| `Frame` (global render target) | Per-component `Buffer` + compositor `Screen` |

---

## §9 Performance Expectations

| Scenario | v1 | v2 (expected) |
|----------|-----|---------------|
| Cell hover | O(12993) renderVNode | O(1) paint + O(1) blit |
| Cell click | O(12993) renderVNode | O(1) paint + O(few) blit |
| Window move | O(N) full repaint | O(dirty_rect_area) recompose |
| Terminal resize | O(N) full repaint | O(N) full repaint (unavoidable) |
| Memory (12993 cells) | 1 Frame (200×50) | 12993 flat Buffers (1 cell each) + Screen ≈ +2MB |
| OcclusionMap rebuild | N/A | O(screen_area) = O(25600) ≈ 50µs |
