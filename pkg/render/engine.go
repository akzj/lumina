package render

import (
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/buffer"
	"github.com/akzj/lumina/pkg/perf"
)

// AnimationManager is the interface for animation management.
// Implemented by animation.Manager to avoid import cycles (render cannot import animation).
type AnimationManager interface {
	// StartAnim begins a new animation. Returns the created animation's initial value.
	StartAnim(id string, from, to float64, duration int64, easing string, loop bool, onUpdate func(float64), onDone func(), nowMs int64) float64
	// GetAnimValue returns the current value of an animation, or (0, false) if not found.
	GetAnimValue(id string) (float64, bool)
	// StopAnim stops and removes an animation by ID.
	StopAnim(id string)
}

// Engine is the new render engine that manages persistent RenderNode trees.
// It replaces the VNode-based rendering pipeline with direct Lua→Descriptor→Reconcile.
type Engine struct {
	L          *lua.State
	root       *Component       // root component (or nil)
	components map[string]*Component
	width      int
	height     int
	buffer     *CellBuffer

	// Hook context: which component is currently rendering
	currentComp *Component

	// Factory registry: name → Lua registry ref for render function
	factories map[string]int64 // factory name → renderFn Lua ref

	// Shared metatable ref for callable factory tables (__call → createElement)
	factoryMetaRef int64

	// Event state: currently hovered node for enter/leave tracking
	hoveredNode *Node

	// Focus state: currently focused input/textarea node
	focusedNode *Node

	// Mouse capture: node that has captured mouse events (drag/resize)
	capturedNode *Node

	// Click prevention: set by HandleMouseDown when handler calls preventDefault
	clickPrevented bool

	// Lua ref cleanup: refs to unref after reconcile
	pendingUnrefs []int64

	// Async coroutine scheduler
	scheduler *lua.Scheduler

	// Performance tracking
	tracker *perf.Tracker

	// Render flag: true when any component is dirty or any node needs layout/paint
	needsRender bool

	// Layer stack: [0] = main app layer, [1..n] = overlay layers
	layers []*Layer

	// ThemeGetter returns the current theme as a map of color tokens.
	// Set by the app layer to avoid import cycles (render cannot import widget).
	ThemeGetter func() map[string]string

	// ThemeSetter switches the active theme by name.
	// Set by the app layer to avoid import cycles (render cannot import widget).
	ThemeSetter func(name string) bool

	// customTheme holds a user-provided theme table (from Lua setTheme(table)).
	customTheme map[string]string

	// AnimManager is the animation manager (set by App after construction).
	// Nil if animations are not supported (e.g., in tests without App).
	AnimManager AnimationManager

	// NowMs returns the current time in milliseconds.
	// Set by App; defaults to 0 if not set.
	NowMs func() int64

	// Scrollbar drag state
	scrollbarDragNode        *Node // node being scrollbar-dragged (nil = no drag)
	scrollbarDragStartY      int   // mouse Y at drag start
	scrollbarDragStartScrollY int  // ScrollY at drag start

	// Cursor blink state
	cursorBlinkOn    bool // true = cursor visible, false = hidden
	cursorBlinkFrame int  // frame counter for blink timing
}

// SetTracker sets the performance tracker for recording render-engine metrics.
func (e *Engine) SetTracker(t *perf.Tracker) {
	e.tracker = t
}

// NeedsRender returns true when there is pending dirty work (components, layout, or paint).
func (e *Engine) NeedsRender() bool {
	return e.needsRender
}

// MarkNeedsRender sets the needsRender flag so the next RenderDirty call does work.
// Call this after externally marking a component dirty or modifying node dirty flags.
func (e *Engine) MarkNeedsRender() {
	e.needsRender = true
}

// MarkAllComponentsDirty marks all components as needing re-render.
// Used after hot reload when function prototypes have been swapped in-place.
func (e *Engine) MarkAllComponentsDirty() {
	for _, comp := range e.components {
		comp.Dirty = true
	}
	if e.root != nil {
		e.root.Dirty = true
	}
	e.needsRender = true
}

// drainPendingUnrefs frees all Lua registry refs collected during reconcile.
// For table refs (useRef), sets ref.current = nil before unreffing.
func (e *Engine) drainPendingUnrefs() {
	if len(e.pendingUnrefs) == 0 {
		return
	}
	L := e.L
	for _, ref := range e.pendingUnrefs {
		// If this ref is a table (useRef), set current = nil before unreffing
		L.RawGetI(lua.RegistryIndex, ref)
		if L.IsTable(-1) {
			L.PushNil()
			L.SetField(-2, "current")
		}
		L.Pop(1)
		L.Unref(lua.RegistryIndex, int(ref))
	}
	e.pendingUnrefs = e.pendingUnrefs[:0]
}

// Destroy releases all Lua registry refs held by the engine.
// Call before discarding the engine (app shutdown, full hot-reload).
func (e *Engine) Destroy() {
	L := e.L
	if L == nil {
		return
	}

	// Free all component trees (hooks, event handlers, propFuncRefs, node refs).
	if e.root != nil {
		e.cleanupComponentTree(e.root)
		e.root = nil
	}

	// Clear ALL components (not just root — sub-components may have stale entries).
	for id := range e.components {
		delete(e.components, id)
	}

	// Free all Lua factory refs.
	for name, ref := range e.factories {
		L.Unref(lua.RegistryIndex, int(ref))
		delete(e.factories, name)
	}

	// Free factory metatable ref.
	if e.factoryMetaRef != 0 {
		L.Unref(lua.RegistryIndex, int(e.factoryMetaRef))
		e.factoryMetaRef = 0
	}

	// Drain any pending unrefs accumulated during cleanup.
	e.drainPendingUnrefs()
}

// NewEngine creates a new render engine.
func NewEngine(L *lua.State, width, height int) *Engine {
	return &Engine{
		L:              L,
		components:     make(map[string]*Component),
		factories:      make(map[string]int64),
		width:          width,
		height:         height,
		buffer:         NewCellBuffer(width, height),
		layers:         make([]*Layer, 0, 4),
		cursorBlinkOn:  true,
	}
}

// Buffer returns the engine's cell buffer.
func (e *Engine) Buffer() *CellBuffer { return e.buffer }

// Root returns the root component.
func (e *Engine) Root() *Component { return e.root }

// GetComponent returns a component by ID.
func (e *Engine) GetComponent(id string) *Component { return e.components[id] }

// HasComponent returns true if a component with the given ID exists.
func (e *Engine) HasComponent(id string) bool {
	_, exists := e.components[id]
	return exists
}

// CurrentComponent returns the component currently being rendered (for hooks).
func (e *Engine) CurrentComponent() *Component { return e.currentComp }

// AllComponents returns all registered components.
func (e *Engine) AllComponents() map[string]*Component { return e.components }

// CursorPosition returns the screen position of the hardware cursor.
// If a focused input/textarea exists, returns its cursor coordinates and visible=true.
// Otherwise returns (0, 0, false).
func (e *Engine) CursorPosition() (x, y int, visible bool) {
	node := e.focusedNode
	if node == nil {
		return 0, 0, false
	}

	// Lua-based input with cursor hint — disabled (software cursor handles display).
	// Hardware cursor is only used for native input/textarea nodes.

	if node.Type != "input" && node.Type != "textarea" {
		return 0, 0, false
	}

	// Multi-line textarea: compute line/col cursor position
	if node.Type == "textarea" && strings.Contains(node.Content, "\n") {
		runes := []rune(node.Content)
		cursorLine, cursorCol := computeCursorLineCol(runes, node.CursorPos)
		// Vertical scroll
		scrollY := 0
		if cursorLine >= node.H {
			scrollY = cursorLine - node.H + 1
		}
		visRow := cursorLine - scrollY
		if visRow < 0 || visRow >= node.H {
			return 0, 0, false
		}
		return node.X + cursorCol, node.Y + visRow, true
	}

	// Single-line input: horizontal scroll
	cursorOffset := inputCursorScreenOffset(node)
	availW := node.W
	scrollX := 0
	if cursorOffset >= availW {
		scrollX = cursorOffset - availW + 1
	}
	return node.X + cursorOffset - scrollX, node.Y, true
}

// TickCursorBlink should be called once per frame (~60fps).
// Toggles cursor visibility every 30 frames (~500ms at 60fps).
// Returns true if blink state changed (caller should trigger repaint).
func (e *Engine) TickCursorBlink() bool {
	e.cursorBlinkFrame++
	if e.cursorBlinkFrame >= 30 {
		e.cursorBlinkFrame = 0
		e.cursorBlinkOn = !e.cursorBlinkOn
		return true
	}
	return false
}

// CursorBlinkOn returns whether the cursor should be visible in the current blink phase.
func (e *Engine) CursorBlinkOn() bool {
	return e.cursorBlinkOn
}

// ResetCursorBlink resets the blink timer (call on any keypress to keep cursor visible).
func (e *Engine) ResetCursorBlink() {
	e.cursorBlinkOn = true
	e.cursorBlinkFrame = 0
}

// RepaintCursor repaints just the cursor cell based on current blink state.
// Call after TickCursorBlink returns true to update the buffer.
func (e *Engine) RepaintCursor() {
	node := e.focusedNode
	if node == nil || (node.Type != "input" && node.Type != "textarea") {
		return
	}

	var cursorX, cursorY int

	// Multi-line textarea
	if node.Type == "textarea" && strings.Contains(node.Content, "\n") {
		runes := []rune(node.Content)
		cursorLine, cursorCol := computeCursorLineCol(runes, node.CursorPos)
		scrollY := 0
		if cursorLine >= node.H {
			scrollY = cursorLine - node.H + 1
		}
		visRow := cursorLine - scrollY
		if visRow < 0 || visRow >= node.H {
			return
		}
		cursorX = node.X + cursorCol
		cursorY = node.Y + visRow
	} else {
		// Single-line input
		cursorOffset := inputCursorScreenOffset(node)
		availW := node.W
		scrollX := 0
		if cursorOffset >= availW {
			scrollX = cursorOffset - availW + 1
		}
		cursorX = node.X + cursorOffset - scrollX
		cursorY = node.Y
	}

	if e.cursorBlinkOn {
		paintInputCursor(e.buffer, node, cursorX, cursorY)
	} else {
		// Restore the cell under the cursor (paint with input's normal colors)
		bg := node.Style.Background
		if bg == "" {
			bg = "#1E1E2E"
		}
		fg := node.Style.Foreground
		if fg == "" {
			fg = "#CDD6F4"
		}
		ch := cursorCharAt(node)
		e.buffer.Set(cursorX, cursorY, Cell{Ch: ch, FG: fg, BG: bg})
		// Restore padding cell for wide characters
		if runeWidth(ch) == 2 && cursorX+1 < node.X+node.W {
			e.buffer.Set(cursorX+1, cursorY, Cell{Wide: true, BG: bg})
		}
	}
}

// cursorCharAt returns the character at the cursor position in the node's content.
func cursorCharAt(node *Node) rune {
	if node.Content == "" {
		return ' '
	}
	runes := []rune(node.Content)
	if node.CursorPos >= len(runes) {
		return ' '
	}
	ch := runes[node.CursorPos]
	if ch == '\n' {
		return ' '
	}
	return ch
}

// Resize updates the engine dimensions and buffer.
func (e *Engine) Resize(width, height int) {
	e.width = width
	e.height = height
	e.buffer.Resize(width, height)
	// Mark all layers for re-layout
	for _, layer := range e.layers {
		if layer.Root != nil {
			layer.Root.MarkLayoutDirty()
		}
	}
	if e.root != nil && e.root.RootNode != nil {
		e.root.RootNode.MarkLayoutDirty()
	}
	e.needsRender = true
}

// syncMainLayer ensures layers[0] points to the root component's RootNode.
// Called after rendering to keep the main layer in sync.
func (e *Engine) syncMainLayer() {
	var mainRoot *Node
	if e.root != nil {
		mainRoot = e.root.RootNode
	}
	if len(e.layers) == 0 {
		e.layers = append(e.layers, &Layer{ID: "_main", Root: mainRoot})
	} else {
		e.layers[0].Root = mainRoot
	}
}

// CreateLayer creates a new overlay layer and pushes it onto the stack.
// The root node should have position/size set via its Style (Left, Top, Width, Height).
func (e *Engine) CreateLayer(id string, root *Node, modal bool) *Layer {
	layer := &Layer{ID: id, Root: root, Modal: modal}
	e.layers = append(e.layers, layer)
	if root != nil {
		root.LayoutDirty = true
		root.PaintDirty = true
	}
	e.needsRender = true
	return layer
}

// RemoveLayer removes a layer by ID and marks the covered area for repaint.
func (e *Engine) RemoveLayer(id string) {
	for i, l := range e.layers {
		if l.ID == id && i > 0 { // Never remove layer 0 (main app)
			// Mark the area this layer covered as dirty on layers below
			if l.Root != nil && l.Root.W > 0 && l.Root.H > 0 {
				for j := 0; j < i; j++ {
					if e.layers[j].Root != nil {
						markOverlappingDirty(e.layers[j].Root, l.Root.X, l.Root.Y, l.Root.W, l.Root.H)
					}
				}
			}

			// Clean up the layer's node tree
			if l.Root != nil {
				// Clear focus/hover if they belong to this layer
				if e.focusedNode != nil && isDescendantOf(e.focusedNode, l.Root) {
					e.focusedNode = nil
				}
				if e.hoveredNode != nil && isDescendantOf(e.hoveredNode, l.Root) {
					e.hoveredNode = nil
				}
				// Mark nodes as removed and collect refs to free
				markRemovedRecursive(l.Root)
				collectNodeRefsRecursive(l.Root, &e.pendingUnrefs)
			}

			e.layers = append(e.layers[:i], e.layers[i+1:]...)
			e.needsRender = true
			return
		}
	}
}

// BringToFront moves a layer to the top of the stack.
func (e *Engine) BringToFront(id string) {
	for i, l := range e.layers {
		if l.ID == id && i > 0 { // Don't move layer 0
			e.layers = append(e.layers[:i], e.layers[i+1:]...)
			e.layers = append(e.layers, l)
			if l.Root != nil {
				l.Root.PaintDirty = true
			}
			e.needsRender = true
			return
		}
	}
}

// Layers returns the current layer stack (read-only view).
func (e *Engine) Layers() []*Layer {
	return e.layers
}

// DefineComponent registers a component factory.
// Called from Lua: lumina.defineComponent("Cell", renderFn)
func (e *Engine) DefineComponent(name string, renderFnRef int64) {
	// Free the old factory ref if redefining (e.g. module-level hot-reload).
	if old, exists := e.factories[name]; exists {
		e.L.Unref(lua.RegistryIndex, int(old))
	}
	e.factories[name] = renderFnRef
}

// CreateRootComponent creates and registers a root component.
func (e *Engine) CreateRootComponent(id, name string, renderFnRef int64) {
	comp := NewComponent(id, name, name)
	comp.RenderFn = renderFnRef
	comp.IsRoot = true
	comp.Dirty = true
	e.components[id] = comp
	e.root = comp
	e.needsRender = true
}

// SetState sets a state value on a component and marks it dirty.
func (e *Engine) SetState(compID, key string, value any) {
	comp := e.components[compID]
	if comp == nil {
		return
	}
	oldDirty := comp.Dirty
	comp.SetState(key, value)
	if !oldDirty && comp.Dirty {
		e.needsRender = true
	}
}

// RenderDirty renders all dirty components, reconciles, layouts, and paints.
// This is the main frame function.
func (e *Engine) RenderDirty() {
	// Always reset stats so callers see accurate per-frame numbers.
	e.buffer.ResetStats()

	if !e.needsRender {
		return // Nothing dirty — skip all tree walks
	}
	e.needsRender = false

	// 1. Render dirty components in dependency order (parents first)
	rendered := e.renderInOrder()
	if e.tracker != nil {
		e.tracker.Record(perf.ComponentsRendered, rendered)
	}

	// 2. Graft child component RootNodes into parent tree
	e.graftChildComponents()

	// Sync main layer
	e.syncMainLayer()

	// 3. Early exit: check all layers for dirty nodes
	anyDirty := false
	for _, layer := range e.layers {
		if layer.Root != nil && hasAnyDirty(layer.Root) {
			anyDirty = true
			break
		}
	}
	if rendered == 0 && !anyDirty {
		if e.tracker != nil {
			e.tracker.Record(perf.PaintCells, 0)
			e.tracker.Record(perf.PaintClearCells, 0)
			e.tracker.Record(perf.DirtyRectArea, 0)
		}
		return
	}

	// 4. Layout all layers
	for i, layer := range e.layers {
		if layer.Root == nil {
			continue
		}
		if layer.Root.LayoutDirty {
			if i == 0 {
				LayoutFull(layer.Root, 0, 0, e.width, e.height)
			} else {
				// Overlay layers: use their root node's style for position/size
				lx := layer.Root.Style.Left
				ly := layer.Root.Style.Top
				lw := layer.Root.Style.Width
				lh := layer.Root.Style.Height
				if lw <= 0 {
					lw = layer.Root.W
				}
				if lh <= 0 {
					lh = layer.Root.H
				}
				if lw <= 0 {
					lw = e.width
				}
				if lh <= 0 {
					lh = e.height
				}
				LayoutFull(layer.Root, lx, ly, lw, lh)
			}
		} else {
			LayoutIncremental(layer.Root)
		}
	}

	// Populate ref.current for nodes with ref prop (after layout, before paint)
	for _, layer := range e.layers {
		if layer.Root != nil {
			populateRefs(layer.Root, e.L)
		}
	}

	// 5. Paint all layers (bottom to top)
	e.buffer.CursorBlinkOn = e.cursorBlinkOn
	for i, layer := range e.layers {
		if layer.Root != nil {
			if i == 0 {
				// Main layer: full clear + repaint when dirty
				PaintDirty(e.buffer, layer.Root)
			} else {
				// Overlay layers: repaint without clearing (paint on top)
				PaintDirtyOverlay(e.buffer, layer.Root)
			}
		}
	}

	// 6. Record paint stats from CellBuffer.
	if e.tracker != nil {
		stats := e.buffer.Stats()
		e.tracker.Record(perf.PaintCells, stats.WriteCount)
		e.tracker.Record(perf.PaintClearCells, stats.ClearCount)
		e.tracker.Record(perf.DirtyRectArea, stats.DirtyW*stats.DirtyH)
	}

	// 7. Fire pending useEffect callbacks (after paint, like React)
	e.firePendingEffects()

	// 8. Auto-focus newly created nodes with autoFocus=true
	// Run if: no focus, focus removed, or focused node is in a hidden subtree (display:none)
	if e.focusedNode == nil || e.focusedNode.Removed || isNodeHidden(e.focusedNode) {
		e.FocusAutoFocus()
	} else if e.focusedNode.Type != "input" && e.focusedNode.Type != "textarea" {
		// Only steal focus when a component actually re-rendered this frame
		// (e.g., on mount or tab switch). Don't steal on idle frames.
		if rendered > 0 {
			e.focusAutoFocusOnly()
		}
	}
}

// isNodeHidden returns true if the node or any ancestor has display:none.
func isNodeHidden(node *Node) bool {
	for n := node; n != nil; n = n.Parent {
		if n.Style.Display == "none" {
			return true
		}
	}
	return false
}

// RenderAll does a full render of everything (initial mount).
func (e *Engine) RenderAll() {
	// Reset CellBuffer stats for this frame.
	e.buffer.ResetStats()

	for _, comp := range e.components {
		comp.Dirty = true
	}
	e.needsRender = false // RenderAll handles everything inline; clear the flag

	// Render all components in dependency order (parents first)
	rendered := e.renderInOrder()
	if e.tracker != nil {
		e.tracker.Record(perf.ComponentsRendered, rendered)
	}

	// Graft child component RootNodes into parent tree
	e.graftChildComponents()

	// Sync main layer and do full layout + paint for all layers
	e.syncMainLayer()
	e.buffer.Clear()
	e.buffer.CursorBlinkOn = e.cursorBlinkOn
	for i, layer := range e.layers {
		if layer.Root == nil {
			continue
		}
		if i == 0 {
			LayoutFull(layer.Root, 0, 0, e.width, e.height)
		} else {
			lx := layer.Root.Style.Left
			ly := layer.Root.Style.Top
			lw := layer.Root.Style.Width
			lh := layer.Root.Style.Height
			if lw <= 0 {
				lw = layer.Root.W
			}
			if lh <= 0 {
				lh = layer.Root.H
			}
			if lw <= 0 {
				lw = e.width
			}
			if lh <= 0 {
				lh = e.height
			}
			LayoutFull(layer.Root, lx, ly, lw, lh)
		}
		populateRefs(layer.Root, e.L)
		paintNode(e.buffer, layer.Root)
		clearPaintDirty(layer.Root)
	}

	// Auto-focus first node with autoFocus=true
	e.FocusAutoFocus()

	// Record paint stats from CellBuffer.
	if e.tracker != nil {
		stats := e.buffer.Stats()
		e.tracker.Record(perf.PaintCells, stats.WriteCount)
		e.tracker.Record(perf.PaintClearCells, stats.ClearCount)
		e.tracker.Record(perf.DirtyRectArea, stats.DirtyW*stats.DirtyH)
	}

	// Fire pending useEffect callbacks (after paint, like React)
	e.firePendingEffects()
}

// renderComponent calls the Lua render function and reconciles the result.
func (e *Engine) renderComponent(comp *Component) {
	L := e.L

	// Stop GC during render
	L.SetGCStopped(true)
	defer func() {
		L.SetGCStopped(false)
		L.GCStepAPI()
	}()

	// Set current component (for hooks like useState)
	e.currentComp = comp
	defer func() { e.currentComp = nil }()

	// Reset hook index for this render cycle
	comp.hookIdx = 0

	// Push render function from registry
	L.RawGetI(lua.RegistryIndex, comp.RenderFn)
	if !L.IsFunction(-1) {
		L.Pop(1)
		comp.Dirty = false
		return
	}

	// Push props table
	pushMap(L, comp.Props)

	// PCall(1 arg = props, 1 result, 0 error handler)
	if status := L.PCall(1, 1, 0); status != lua.OK {
		errMsg, _ := L.ToString(-1)
		L.Pop(1) // pop error
		comp.Dirty = false
		comp.LastError = errMsg
		log.Printf("[lumina] render error in component %q (id=%s): %s", comp.Name, comp.ID, errMsg)
		return
	}
	comp.LastError = "" // clear on success

	// Read descriptor from Lua stack (the returned table)
	if !L.IsTable(-1) {
		L.Pop(1)
		comp.Dirty = false
		return
	}

	desc := e.readDescriptor(L, -1)
	L.Pop(1)

	// Reconcile against existing RenderNode tree
	if comp.RootNode == nil {
		// First mount: create tree from descriptor
		comp.RootNode = createNodeFromDesc(desc)
		comp.RootNode.Component = comp
		comp.RootNode.LayoutDirty = true
		comp.RootNode.PaintDirty = true
	} else {
		// Update: reconcile (diff + patch in-place), collect freed refs
		ReconcileCollectRefs(comp.RootNode, desc, &e.pendingUnrefs)
	}

	// Handle sub-component children
	e.reconcileChildComponents(comp, comp.RootNode)

	// Cleanup child components that are no longer in the tree
	e.cleanupRemovedChildComponents(comp, comp.RootNode)

	// Unref all freed Lua refs from this reconcile
	e.drainPendingUnrefs()

	comp.Dirty = false
	comp.Mounted = true
	comp.RenderCount++
}

func (e *Engine) reconcileChildComponents(parent *Component, node *Node) {
	if node == nil {
		return
	}

	// If this node represents a sub-component, handle it
	if node.Type == "component" && node.ComponentType != "" {
		factoryName := node.ComponentType
		// Use ID for lookup; fall back to Key when ID is empty.
		lookupKey := node.ID
		if lookupKey == "" {
			lookupKey = node.Key
		}
		// If still no key, use the positional key set by the parent's recursion.
		if lookupKey == "" {
			lookupKey = node.positionalKey
		}
		child := parent.FindChild(factoryName, lookupKey)
		if child == nil {
			// Create new child component
			renderRef, ok := e.factories[factoryName]
			if !ok {
				return
			}
			childID := parent.ID + ":" + lookupKey
			if lookupKey == "" {
				childID = parent.ID + ":" + factoryName
			}
			child = NewComponent(childID, factoryName, factoryName)
			child.RenderFn = renderRef
			child.Parent = parent
			child.Props = node.ComponentProps
			parent.AddChild(child, lookupKey)
			e.components[childID] = child
			child.Dirty = true
		} else {
			// Existing child: update props and mark dirty if changed.
			if !propsEqual(child.Props, node.ComponentProps) {
				unrefPropFuncRefsInProps(e.L, child.Props)
				child.Props = node.ComponentProps
				child.Dirty = true
			}
		}
		node.Component = child
		return
	}

	// Recurse into children, assigning positional keys to same-type siblings
	// that lack explicit keys (like React's implicit array index keys).
	typeCount := make(map[string]int) // tracks occurrence index per component type
	for _, ch := range node.Children {
		if ch != nil && ch.Type == "component" && ch.ComponentType != "" {
			explicitKey := ch.ID
			if explicitKey == "" {
				explicitKey = ch.Key
			}
			if explicitKey == "" {
				idx := typeCount[ch.ComponentType]
				ch.positionalKey = "__pos_" + strconv.Itoa(idx)
				typeCount[ch.ComponentType] = idx + 1
			}
		}
		e.reconcileChildComponents(parent, ch)
	}
}

// propsEqual returns true if two props maps are equal (shallow comparison).
func propsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if !safeEqual(va, vb) {
			return false
		}
	}
	return true
}

// safeEqual compares two values safely, handling uncomparable types like slices and maps.
func safeEqual(a, b any) bool {
	// Fast path for common comparable types
	switch av := a.(type) {
	case nil:
		return b == nil
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case int:
		bv, ok := b.(int)
		return ok && av == bv
	case int64:
		bv, ok := b.(int64)
		return ok && av == bv
	case propFuncRef:
		bv, ok := b.(propFuncRef)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	default:
		// Uncomparable types (slices, maps): use reflect
		return reflect.DeepEqual(a, b)
	}
}

// cleanupRemovedChildComponents removes child components that are no longer
// referenced in the current render tree. This prevents component leaks.
func (e *Engine) cleanupRemovedChildComponents(parent *Component, rootNode *Node) {
	// Collect all component type:key pairs referenced in the current tree
	activeKeys := make(map[string]bool)
	collectActiveComponentKeys(rootNode, activeKeys)

	// Remove children not in activeKeys
	var kept []*Component
	for _, child := range parent.Children {
		mapKey := child.Type
		// Find the lookup key used in ChildMap
		for k, v := range parent.ChildMap {
			if v == child {
				mapKey = k
				break
			}
		}
		if activeKeys[mapKey] {
			kept = append(kept, child)
		} else {
			// Remove from engine
			delete(e.components, child.ID)
			// Recursively cleanup grandchildren
			e.cleanupComponentTree(child)
		}
	}

	if len(kept) != len(parent.Children) {
		parent.Children = kept
		// Rebuild ChildMap from kept children
		parent.ChildMap = make(map[string]*Component)
		for _, child := range kept {
			// Reconstruct the map key from type + lookup key.
			// AddChild uses: "Type:lookupKey" when lookupKey != "", else just "Type".
			// child.ID is "parentID:lookupKey" or "parentID:factoryName" (when no key).
			lookupKey := ""
			if child.ID != "" {
				parts := splitAfterColon(child.ID, parent.ID)
				// If the extracted part equals the factory name, it means no
				// explicit key was provided — AddChild used just child.Type.
				if parts != "" && parts != child.Type {
					lookupKey = parts
				}
			}
			mapKey := child.Type
			if lookupKey != "" {
				mapKey = child.Type + ":" + lookupKey
			}
			parent.ChildMap[mapKey] = child
		}
	}
}

// splitAfterColon extracts the part after "parentID:" from childID.
func splitAfterColon(childID, parentID string) string {
	prefix := parentID + ":"
	if len(childID) > len(prefix) && childID[:len(prefix)] == prefix {
		return childID[len(prefix):]
	}
	return ""
}

// collectActiveComponentKeys walks the node tree and collects the ChildMap keys
// for all component placeholder nodes.
func collectActiveComponentKeys(node *Node, keys map[string]bool) {
	if node == nil {
		return
	}
	if node.Type == "component" && node.ComponentType != "" {
		lookupKey := node.ID
		if lookupKey == "" {
			lookupKey = node.Key
		}
		if lookupKey == "" {
			lookupKey = node.positionalKey
		}
		mapKey := node.ComponentType
		if lookupKey != "" {
			mapKey = node.ComponentType + ":" + lookupKey
		}
		keys[mapKey] = true
		return // Don't recurse into component children (they belong to the child component)
	}
	for _, child := range node.Children {
		collectActiveComponentKeys(child, keys)
	}
}

// cleanupComponentTree recursively removes a component and all its descendants
// from the engine's component map, and unrefs their Lua render functions and
// any refs on their render nodes.
func (e *Engine) cleanupComponentTree(comp *Component) {
	for _, child := range comp.Children {
		delete(e.components, child.ID)
		e.cleanupComponentTree(child)
	}

	// Drop Lua function refs held in props (nested propFuncRef from readMapFromTable).
	unrefPropFuncRefsInProps(e.L, comp.Props)
	comp.Props = nil

	// Cleanup hook refs (effects, memos, refs) — runs effect cleanups
	e.cleanupComponentHooks(comp)
	// Unref the component's render function — but only if it's NOT a shared
	// factory ref. Factory refs (from defineComponent) are shared across all
	// instances and must not be freed when an individual instance is removed.
	if comp.RenderFn != 0 {
		factoryRef, isFactory := e.factories[comp.Type]
		if !isFactory || comp.RenderFn != factoryRef {
			e.L.Unref(lua.RegistryIndex, int(comp.RenderFn))
		}
		comp.RenderFn = 0
	}
	// Collect and unref all node refs from the component's render tree
	if comp.RootNode != nil {
		collectNodeRefsRecursive(comp.RootNode, &e.pendingUnrefs)
		e.drainPendingUnrefs()
	}
	comp.Children = nil
	comp.ChildMap = nil
}

// --- Lua API Registration ---

// renderInOrder renders components in dependency order: parents before children.
// This ensures parent trees have component placeholders before child components render.
// Returns the number of components that were rendered.
func (e *Engine) renderInOrder() int {
	count := 0
	// Render root first (it creates the component placeholders)
	if e.root != nil && e.root.Dirty {
		e.renderComponent(e.root)
		count++
	}
	// Loop until no more newly-dirty components remain.
	// reconcileChildComponents inside renderComponent may create new dirty children,
	// so a single pass can miss them.
	const maxRenderPasses = 50
	for iterations := 0; iterations < maxRenderPasses; iterations++ {
		var dirty []*Component
		for _, comp := range e.components {
			if !comp.Dirty || comp.IsRoot {
				continue
			}
			dirty = append(dirty, comp)
		}
		if len(dirty) == 0 {
			break
		}
		if iterations == maxRenderPasses-1 {
			// Possible circular dependency: components keep dirtying each other.
			// Log remaining dirty components for debugging.
			_ = dirty // convergence not reached; last pass will render what it can
		}
		sort.Slice(dirty, func(i, j int) bool {
			return componentDepth(dirty[i]) < componentDepth(dirty[j])
		})
		for _, comp := range dirty {
			if !comp.Dirty {
				continue // may have been rendered as side effect
			}
			// Skip if any ancestor component is still dirty — this component
			// will receive stale props. It will be re-dirtied by the ancestor's
			// reconcileChildComponents and rendered in a subsequent iteration
			// with correct props.
			if ancestorDirty(comp) {
				continue
			}
			e.renderComponent(comp)
			count++
		}
	}
	return count
}

// componentDepth returns the depth of a component in the tree (0 = root).
func componentDepth(c *Component) int {
	depth := 0
	for p := c.Parent; p != nil; p = p.Parent {
		depth++
	}
	return depth
}

// ancestorDirty returns true if any ancestor component of c is still dirty.
// This means c would render with stale props — its ancestor hasn't propagated
// updated props yet via reconcileChildComponents.
func ancestorDirty(c *Component) bool {
	for p := c.Parent; p != nil; p = p.Parent {
		if p.Dirty {
			return true
		}
	}
	return false
}

// hasAnyDirty returns true if any node in the tree has LayoutDirty or PaintDirty set.
func hasAnyDirty(node *Node) bool {
	if node == nil {
		return false
	}
	if node.LayoutDirty || node.PaintDirty {
		return true
	}
	for _, child := range node.Children {
		if hasAnyDirty(child) {
			return true
		}
	}
	return false
}

// CountDirtyNodes counts nodes with PaintDirty or LayoutDirty set (exported for testing).
func CountDirtyNodes(node *Node) int {
	if node == nil {
		return 0
	}
	c := 0
	if node.PaintDirty || node.LayoutDirty {
		c++
	}
	for _, child := range node.Children {
		c += CountDirtyNodes(child)
	}
	return c
}

// graftChildComponents walks the root tree and connects child component
// RootNodes as children of their placeholder nodes. This allows layout and
// paint to naturally traverse into sub-components.
func (e *Engine) graftChildComponents() {
	if e.root == nil || e.root.RootNode == nil {
		return
	}
	visited := make(map[*Node]bool)
	e.graftWalk(e.root.RootNode, visited)
}

// graftWalk recursively finds component placeholder nodes and grafts the
// child component's RootNode as the placeholder's child.
// Only marks dirty when the graft actually changes (new or different RootNode).
// Uses a visited set to prevent infinite recursion from cycles.
func (e *Engine) graftWalk(node *Node, visited map[*Node]bool) {
	if node == nil {
		return
	}
	if visited[node] {
		return // cycle detected — break infinite recursion
	}
	visited[node] = true

	// Handle the case where node itself is a component placeholder
	// (happens when root render returns a defineComponent directly)
	if node.Type == "component" && node.Component != nil {
		comp := node.Component
		if comp.RootNode != nil && comp.RootNode != node {
			alreadyGrafted := len(node.Children) == 1 && node.Children[0] == comp.RootNode
			if !alreadyGrafted {
				node.Children = []*Node{comp.RootNode}
				comp.RootNode.Parent = node
				node.LayoutDirty = true
				node.PaintDirty = true
			}
		}
	}

	for _, child := range node.Children {
		if child.Type == "component" && child.Component != nil {
			comp := child.Component
			if comp.RootNode != nil && comp.RootNode != child {
				// Only mark dirty if the grafted child actually changed
				alreadyGrafted := len(child.Children) == 1 && child.Children[0] == comp.RootNode
				if !alreadyGrafted {
					child.Children = []*Node{comp.RootNode}
					comp.RootNode.Parent = child
					child.LayoutDirty = true
					child.PaintDirty = true
				}
			}
		}
		// Always recurse (component children may contain nested components)
		e.graftWalk(child, visited)
	}
}

// ToBuffer converts the engine's CellBuffer to a buffer.Buffer for output.
// Convention translation: in CellBuffer, Wide=true marks the PADDING cell (x+1).
// In buffer.Buffer, Wide=true marks the MAIN cell (the one with the character).
// The output adapter uses buffer.Buffer convention to skip padding cells.
func (e *Engine) ToBuffer() *buffer.Buffer {
	buf := buffer.New(e.width, e.height)
	cb := e.buffer
	for y := 0; y < e.height; y++ {
		for x := 0; x < e.width; x++ {
			c := cb.Get(x, y)
			if c.Ch == 0 && c.FG == "" && c.BG == "" && !c.Wide {
				continue // skip zero cells (but preserve Wide padding cells)
			}
			// Check if the NEXT cell is a Wide padding cell — if so, this is a wide char
			isWideChar := false
			if x+1 < e.width {
				next := cb.Get(x+1, y)
				if next.Wide {
					isWideChar = true
				}
			}
			buf.Set(x, y, buffer.Cell{
				Char:          c.Ch,
				Foreground:    c.FG,
				Background:    c.BG,
				Bold:          c.Bold,
				Dim:           c.Dim,
				Underline:     c.Underline,
				Italic:        c.Italic,
				Strikethrough: c.Strikethrough,
				Inverse:       c.Inverse,
				Wide:          isWideChar, // Wide on MAIN cell, not padding
			})
		}
	}
	return buf
}

// DirtyRect returns the bounding rect of cells that were written or cleared
// since the last ResetStats (i.e., during the most recent RenderDirty/RenderAll).
func (e *Engine) DirtyRect() buffer.Rect {
	stats := e.buffer.Stats()
	if stats.DirtyW == 0 || stats.DirtyH == 0 {
		return buffer.Rect{} // nothing dirty
	}
	return buffer.Rect{X: stats.DirtyX, Y: stats.DirtyY, W: stats.DirtyW, H: stats.DirtyH}
}

// VNodeTree returns the current render tree as a VNode (JSON-serializable).
func (e *Engine) VNodeTree() *VNode {
	if e.root == nil || e.root.RootNode == nil {
		return nil
	}
	return NodeToVNode(e.root.RootNode)
}

// Width returns the engine width.
func (e *Engine) Width() int { return e.width }

// Height returns the engine height.
func (e *Engine) Height() int { return e.height }
