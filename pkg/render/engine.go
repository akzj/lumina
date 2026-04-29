package render

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/buffer"
	"github.com/akzj/lumina/pkg/perf"
)

// WidgetEvent is the event passed to widget OnEvent handlers.
// Output fields (FireOnChange) are set by the widget and read by the engine.
type WidgetEvent struct {
	Type string // "click", "mousedown", "mouseup", "mouseenter", "mouseleave", "keydown", "focus", "blur"
	Key  string
	X, Y int
	// Output: set by widget, read by engine after OnEvent returns.
	// Non-nil → engine calls onChange(value) on the widget's root node.
	FireOnChange any

	// Input: Widget's screen bounds (set by engine BEFORE calling OnEvent).
	WidgetX, WidgetY, WidgetW, WidgetH int

	// Input: Screen dimensions (set by engine BEFORE calling OnEvent).
	ScreenW, ScreenH int

	// Output: Mouse capture (set by widget during mousedown to capture subsequent mouse events).
	CaptureMouse bool

	// Output: Layer management (set by widget, processed by engine AFTER OnEvent).
	CreateLayer *LayerRequest // non-nil → engine creates this layer
	RemoveLayer string        // non-empty → engine removes this layer ID

	// Output: Scroll request (set by widget, processed by engine AFTER OnEvent).
	// Non-zero → engine adjusts ScrollY on the widget's root node by this many lines.
	ScrollBy int

	// Input: Current scroll state of the widget's scroll container (set by engine).
	// Populated from the first overflow:"scroll" node in the widget's tree.
	ScrollY       int // current scroll offset
	ContentHeight int // total content height (ScrollHeight)
}

// LayerRequest describes a layer to create.
type LayerRequest struct {
	ID    string
	Root  *Node
	Modal bool
}

// WidgetDef is the interface for Go-native widgets.
// Implemented by widget.Widget (in pkg/widget) to avoid import cycles.
type WidgetDef interface {
	GetName() string
	GetNewState() any
	DoRender(props map[string]any, state any) any // returns *Node
	DoOnEvent(props map[string]any, state any, event *WidgetEvent) bool
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

	// Lua ref cleanup: refs to unref after reconcile
	pendingUnrefs []int64

	// Async coroutine scheduler
	scheduler *lua.Scheduler

	// Performance tracking
	tracker *perf.Tracker

	// Render flag: true when any component is dirty or any node needs layout/paint
	needsRender bool

	// Go widget registry: name → widget definition
	widgets map[string]WidgetDef

	// Go widget state: component.ID → widget state
	widgetStates map[string]any

	// Layer stack: [0] = main app layer, [1..n] = overlay layers
	layers []*Layer

	// Mouse capture: non-nil = this widget captures all mouse events (for dragging)
	capturedComp *Component

	// ThemeGetter returns the current theme as a map of color tokens.
	// Set by the app layer to avoid import cycles (render cannot import widget).
	ThemeGetter func() map[string]string
}

// SetTracker sets the performance tracker for recording V2 engine metrics.
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
func (e *Engine) drainPendingUnrefs() {
	if len(e.pendingUnrefs) == 0 {
		return
	}
	L := e.L
	for _, ref := range e.pendingUnrefs {
		L.Unref(lua.RegistryIndex, int(ref))
	}
	e.pendingUnrefs = e.pendingUnrefs[:0]
}


// NewEngine creates a new render engine.
func NewEngine(L *lua.State, width, height int) *Engine {
	return &Engine{
		L:            L,
		components:   make(map[string]*Component),
		factories:    make(map[string]int64),
		width:        width,
		height:       height,
		buffer:       NewCellBuffer(width, height),
		widgets:      make(map[string]WidgetDef),
		widgetStates: make(map[string]any),
		layers:       make([]*Layer, 0, 4),
	}
}

// goWidgetSentinel is stored in e.factories for Go widgets (not a real Lua ref).
const goWidgetSentinel LuaRef = -999

// RegisterWidget registers a Go widget as a component factory.
// It becomes available in Lua as lumina.<Name> (e.g., lumina.Button).
func (e *Engine) RegisterWidget(w WidgetDef) {
	name := w.GetName()
	e.factories[name] = goWidgetSentinel
	e.widgets[name] = w
}

// Buffer returns the engine's cell buffer.
func (e *Engine) Buffer() *CellBuffer { return e.buffer }

// Root returns the root component.
func (e *Engine) Root() *Component { return e.root }

// GetComponent returns a component by ID.
func (e *Engine) GetComponent(id string) *Component { return e.components[id] }

// CurrentComponent returns the component currently being rendered (for hooks).
func (e *Engine) CurrentComponent() *Component { return e.currentComp }
// AllComponents returns all registered components.
func (e *Engine) AllComponents() map[string]*Component { return e.components }


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
		e.tracker.Record(perf.V2ComponentsRendered, rendered)
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
			e.tracker.Record(perf.V2PaintCells, 0)
			e.tracker.Record(perf.V2PaintClearCells, 0)
			e.tracker.Record(perf.V2DirtyRectArea, 0)
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

	// 5. Paint all layers (bottom to top)
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
		e.tracker.Record(perf.V2PaintCells, stats.WriteCount)
		e.tracker.Record(perf.V2PaintClearCells, stats.ClearCount)
		e.tracker.Record(perf.V2DirtyRectArea, stats.DirtyW*stats.DirtyH)
	}

	// 7. Fire pending useEffect callbacks (after paint, like React)
	e.firePendingEffects()

	// 8. Auto-focus newly created nodes with autoFocus=true
	// (only if nothing is currently focused or focused node was removed)
	if e.focusedNode == nil || e.focusedNode.Removed {
		e.FocusAutoFocus()
	}
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
		e.tracker.Record(perf.V2ComponentsRendered, rendered)
	}

	// Graft child component RootNodes into parent tree
	e.graftChildComponents()

	// Sync main layer and do full layout + paint for all layers
	e.syncMainLayer()
	e.buffer.Clear()
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
		paintNode(e.buffer, layer.Root)
		clearPaintDirty(layer.Root)
	}

	// Auto-focus first node with autoFocus=true
	e.FocusAutoFocus()

	// Record paint stats from CellBuffer.
	if e.tracker != nil {
		stats := e.buffer.Stats()
		e.tracker.Record(perf.V2PaintCells, stats.WriteCount)
		e.tracker.Record(perf.V2PaintClearCells, stats.ClearCount)
		e.tracker.Record(perf.V2DirtyRectArea, stats.DirtyW*stats.DirtyH)
	}

	// Fire pending useEffect callbacks (after paint, like React)
	e.firePendingEffects()
}

// renderComponent calls the Lua render function and reconciles the result.
func (e *Engine) renderComponent(comp *Component) {
	// Check if this is a Go widget
	if w, ok := e.widgets[comp.Type]; ok {
		e.renderGoWidget(comp, w)
		return
	}

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
}

// renderGoWidget renders a component backed by a Go widget.
// It calls Widget.Render to get a *Node tree, then replaces the component's RootNode.
func (e *Engine) renderGoWidget(comp *Component, w WidgetDef) {
	// Get or create widget state
	state, ok := e.widgetStates[comp.ID]
	if !ok {
		state = w.GetNewState()
		e.widgetStates[comp.ID] = state
	}

	// Convert children descriptors to Node trees
	comp.ChildNodes = convertChildDescriptors(comp.Props)

	// Pass child nodes via props so Widget.Render can use them
	if len(comp.ChildNodes) > 0 {
		comp.Props["_childNodes"] = comp.ChildNodes
	}

	// Call Go render function (returns *Node as any)
	result := w.DoRender(comp.Props, state)
	if result == nil {
		comp.Dirty = false
		return
	}
	newRoot, ok := result.(*Node)
	if !ok {
		comp.Dirty = false
		return
	}

	if comp.RootNode == nil {
		// First mount
		comp.RootNode = newRoot
		comp.RootNode.Component = comp
		comp.RootNode.LayoutDirty = true
		comp.RootNode.PaintDirty = true
	} else {
		// Update: replace the root node tree entirely
		// Preserve scroll positions from old tree to new tree
		preserveScrollState(comp.RootNode, newRoot)
		parent := comp.RootNode.Parent
		markRemovedRecursive(comp.RootNode)
		newRoot.Parent = parent
		newRoot.Component = comp
		newRoot.LayoutDirty = true
		newRoot.PaintDirty = true
		comp.RootNode = newRoot
	}

	comp.Dirty = false
	comp.Mounted = true
}

// preserveScrollState copies ScrollY from the old node tree to the new node tree
// for matching scroll containers. This prevents scroll position from resetting
// when a Go widget re-renders (e.g., when a sibling window gets focus).
func preserveScrollState(oldNode, newNode *Node) {
	if oldNode == nil || newNode == nil {
		return
	}
	if oldNode.Style.Overflow == "scroll" && newNode.Style.Overflow == "scroll" {
		newNode.ScrollY = oldNode.ScrollY
	}
	// Recurse into children (match by index since Go widgets produce deterministic trees)
	minLen := len(oldNode.Children)
	if len(newNode.Children) < minLen {
		minLen = len(newNode.Children)
	}
	for i := 0; i < minLen; i++ {
		preserveScrollState(oldNode.Children[i], newNode.Children[i])
	}
}

// copyWidgetEventHandlers copies Lua event refs from the component placeholder
// node to the widget's rendered root node. This allows Lua callbacks (onClick,
// etc.) passed as props to fire through the normal event bubbling system.
func copyWidgetEventHandlers(placeholder *Node, root *Node) {
	if placeholder == nil || root == nil {
		return
	}
	root.OnClick = placeholder.OnClick
	root.OnMouseDown = placeholder.OnMouseDown
	root.OnMouseUp = placeholder.OnMouseUp
	root.OnFocus = placeholder.OnFocus
	root.OnBlur = placeholder.OnBlur
	root.OnKeyDown = placeholder.OnKeyDown
	root.OnChange = placeholder.OnChange
	root.OnScroll = placeholder.OnScroll
	root.OnSubmit = placeholder.OnSubmit
	root.OnOutsideClick = placeholder.OnOutsideClick
	// OnMouseEnter/Leave are handled by Widget.OnEvent for hover state,
	// but also copy them for Lua callbacks
	root.OnMouseEnter = placeholder.OnMouseEnter
	root.OnMouseLeave = placeholder.OnMouseLeave
}

// convertChildDescriptors converts raw children data from Lua props into Node trees.
// Children come from Lua as []any (each element is map[string]any descriptor).
func convertChildDescriptors(props map[string]any) []*Node {
	raw, ok := props["children"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}

	nodes := make([]*Node, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		desc := descriptorFromMap(m)
		node := createNodeFromDesc(desc)
		if node != nil {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// descriptorFromMap converts a raw map (from Lua ToAny) back into a Descriptor.
// This is the inverse of readDescriptor but works on Go maps instead of Lua stack.
func descriptorFromMap(m map[string]any) Descriptor {
	var desc Descriptor

	// Type
	if t, ok := m["type"].(string); ok {
		desc.Type = t
	}
	if desc.Type == "" {
		desc.Type = "box"
	}

	// Content
	if c, ok := m["content"].(string); ok {
		desc.Content = c
		desc.ContentSet = true
	}

	// ID
	if id, ok := m["id"].(string); ok {
		desc.ID = id
	}

	// Key
	if k, ok := m["key"].(string); ok {
		desc.Key = k
	}

	// Component type (for sub-components)
	if ft, ok := m["_factoryName"].(string); ok {
		desc.Type = "component"
		desc.ComponentType = ft
		if p, ok := m["_props"].(map[string]any); ok {
			desc.ComponentProps = p
		}
	}

	// Style
	if s, ok := m["style"].(map[string]any); ok {
		desc.Style = styleFromMap(s)
	} else {
		desc.Style.Right = -1
		desc.Style.Bottom = -1
	}

	// Foreground/Background at top level (Lua shorthand)
	if fg, ok := m["foreground"].(string); ok {
		desc.Style.Foreground = fg
	}
	if bg, ok := m["background"].(string); ok {
		desc.Style.Background = bg
	}

	// Bold, Dim, Underline at top level
	if b, ok := m["bold"].(bool); ok {
		desc.Style.Bold = b
	}
	if d, ok := m["dim"].(bool); ok {
		desc.Style.Dim = d
	}
	if u, ok := m["underline"].(bool); ok {
		desc.Style.Underline = u
	}

	// Focusable / Disabled
	if f, ok := m["focusable"].(bool); ok {
		desc.Focusable = f
	}
	if d, ok := m["disabled"].(bool); ok {
		desc.Disabled = d
	}

	// Placeholder
	if p, ok := m["placeholder"].(string); ok {
		desc.Placeholder = p
	}

	// AutoFocus
	if af, ok := m["autoFocus"].(bool); ok {
		desc.AutoFocus = af
	}

	// Children (recursive)
	if children, ok := m["children"].([]any); ok {
		for _, child := range children {
			if cm, ok := child.(map[string]any); ok {
				childDesc := descriptorFromMap(cm)
				desc.Children = append(desc.Children, childDesc)
			}
		}
	}

	return desc
}

// styleFromMap converts a raw map to a Style struct.
func styleFromMap(m map[string]any) Style {
	var s Style
	s.Right = -1
	s.Bottom = -1

	if w, ok := m["width"].(int64); ok {
		s.Width = int(w)
	} else if w, ok := m["width"].(float64); ok {
		s.Width = int(w)
	}
	if h, ok := m["height"].(int64); ok {
		s.Height = int(h)
	} else if h, ok := m["height"].(float64); ok {
		s.Height = int(h)
	}
	if bg, ok := m["background"].(string); ok {
		s.Background = bg
	}
	if fg, ok := m["foreground"].(string); ok {
		s.Foreground = fg
	}
	if b, ok := m["border"].(string); ok {
		s.Border = b
	}
	if p, ok := m["padding"].(int64); ok {
		s.Padding = int(p)
	} else if p, ok := m["padding"].(float64); ok {
		s.Padding = int(p)
	}
	if g, ok := m["gap"].(int64); ok {
		s.Gap = int(g)
	} else if g, ok := m["gap"].(float64); ok {
		s.Gap = int(g)
	}
	if f, ok := m["flex"].(int64); ok {
		s.Flex = int(f)
	} else if f, ok := m["flex"].(float64); ok {
		s.Flex = int(f)
	}
	if b, ok := m["bold"].(bool); ok {
		s.Bold = b
	}
	if d, ok := m["dim"].(bool); ok {
		s.Dim = d
	}
	if u, ok := m["underline"].(bool); ok {
		s.Underline = u
	}

	// Margin
	if v, ok := m["margin"].(int64); ok {
		s.Margin = int(v)
	} else if v, ok := m["margin"].(float64); ok {
		s.Margin = int(v)
	}
	if v, ok := m["marginTop"].(int64); ok {
		s.MarginTop = int(v)
	} else if v, ok := m["marginTop"].(float64); ok {
		s.MarginTop = int(v)
	}
	if v, ok := m["marginBottom"].(int64); ok {
		s.MarginBottom = int(v)
	} else if v, ok := m["marginBottom"].(float64); ok {
		s.MarginBottom = int(v)
	}
	if v, ok := m["marginLeft"].(int64); ok {
		s.MarginLeft = int(v)
	} else if v, ok := m["marginLeft"].(float64); ok {
		s.MarginLeft = int(v)
	}
	if v, ok := m["marginRight"].(int64); ok {
		s.MarginRight = int(v)
	} else if v, ok := m["marginRight"].(float64); ok {
		s.MarginRight = int(v)
	}

	// Padding individual
	if v, ok := m["paddingTop"].(int64); ok {
		s.PaddingTop = int(v)
	} else if v, ok := m["paddingTop"].(float64); ok {
		s.PaddingTop = int(v)
	}
	if v, ok := m["paddingBottom"].(int64); ok {
		s.PaddingBottom = int(v)
	} else if v, ok := m["paddingBottom"].(float64); ok {
		s.PaddingBottom = int(v)
	}
	if v, ok := m["paddingLeft"].(int64); ok {
		s.PaddingLeft = int(v)
	} else if v, ok := m["paddingLeft"].(float64); ok {
		s.PaddingLeft = int(v)
	}
	if v, ok := m["paddingRight"].(int64); ok {
		s.PaddingRight = int(v)
	} else if v, ok := m["paddingRight"].(float64); ok {
		s.PaddingRight = int(v)
	}

	// Alignment
	if j, ok := m["justify"].(string); ok {
		s.Justify = j
	}
	if a, ok := m["align"].(string); ok {
		s.Align = a
	}

	// Overflow
	if o, ok := m["overflow"].(string); ok {
		s.Overflow = o
	}

	// Positioning
	if p, ok := m["position"].(string); ok {
		s.Position = p
	}
	if v, ok := m["top"].(int64); ok {
		s.Top = int(v)
	} else if v, ok := m["top"].(float64); ok {
		s.Top = int(v)
	}
	if v, ok := m["left"].(int64); ok {
		s.Left = int(v)
	} else if v, ok := m["left"].(float64); ok {
		s.Left = int(v)
	}
	if v, ok := m["right"].(int64); ok {
		s.Right = int(v)
	} else if v, ok := m["right"].(float64); ok {
		s.Right = int(v)
	}
	if v, ok := m["bottom"].(int64); ok {
		s.Bottom = int(v)
	} else if v, ok := m["bottom"].(float64); ok {
		s.Bottom = int(v)
	}
	if v, ok := m["zIndex"].(int64); ok {
		s.ZIndex = int(v)
	} else if v, ok := m["zIndex"].(float64); ok {
		s.ZIndex = int(v)
	}

	// Min/Max sizing
	if v, ok := m["minWidth"].(int64); ok {
		s.MinWidth = int(v)
	} else if v, ok := m["minWidth"].(float64); ok {
		s.MinWidth = int(v)
	}
	if v, ok := m["maxWidth"].(int64); ok {
		s.MaxWidth = int(v)
	} else if v, ok := m["maxWidth"].(float64); ok {
		s.MaxWidth = int(v)
	}
	if v, ok := m["minHeight"].(int64); ok {
		s.MinHeight = int(v)
	} else if v, ok := m["minHeight"].(float64); ok {
		s.MinHeight = int(v)
	}
	if v, ok := m["maxHeight"].(int64); ok {
		s.MaxHeight = int(v)
	} else if v, ok := m["maxHeight"].(float64); ok {
		s.MaxHeight = int(v)
	}

	return s
}


// readDescriptor reads a Lua table at stack index and converts to Descriptor.
func (e *Engine) readDescriptor(L *lua.State, idx int) Descriptor {
	absIdx := L.AbsIndex(idx)

	var desc Descriptor
	desc.Type = getStringField(L, absIdx, "type")
	if desc.Type == "" {
		desc.Type = "box"
	}
	desc.ID = getStringField(L, absIdx, "id")
	desc.Key = getStringField(L, absIdx, "key")
	desc.Content = getStringField(L, absIdx, "content")
	if desc.Content != "" {
		desc.ContentSet = true
	}
	// For input/textarea, also check "value" field
	if desc.Content == "" {
		if v := getStringField(L, absIdx, "value"); v != "" {
			desc.Content = v
			desc.ContentSet = true
		}
	}
	desc.Placeholder = getStringField(L, absIdx, "placeholder")
	desc.AutoFocus = getBoolField(L, absIdx, "autoFocus")
	L.GetField(absIdx, "scrollY")
	if !L.IsNoneOrNil(-1) {
		n, _ := L.ToInteger(-1)
		desc.ScrollY = int(n)
		desc.ScrollYSet = true
	}
	L.Pop(1)

	// Read style — check for nested "style" table first, then top-level style fields
	L.GetField(absIdx, "style")
	if L.IsTable(-1) {
		desc.Style = e.readStyle(L, -1)
	}
	L.Pop(1)

	// Also read top-level style fields (they override if style sub-table didn't set them)
	e.readStyleFields(L, absIdx, &desc.Style)

	// Read event handlers (store as Lua refs)
	desc.OnClick = getRefField(L, absIdx, "onClick")
	desc.OnMouseEnter = getRefField(L, absIdx, "onMouseEnter")
	desc.OnMouseLeave = getRefField(L, absIdx, "onMouseLeave")
	desc.OnKeyDown = getRefField(L, absIdx, "onKeyDown")
	desc.OnChange = getRefField(L, absIdx, "onChange")
	desc.OnScroll = getRefField(L, absIdx, "onScroll")
	desc.OnMouseDown = getRefField(L, absIdx, "onMouseDown")
	desc.OnMouseUp = getRefField(L, absIdx, "onMouseUp")
	desc.OnFocus = getRefField(L, absIdx, "onFocus")
	desc.OnBlur = getRefField(L, absIdx, "onBlur")
	desc.OnSubmit = getRefField(L, absIdx, "onSubmit")
	desc.OnOutsideClick = getRefField(L, absIdx, "onOutsideClick")
	desc.Focusable = getBoolField(L, absIdx, "focusable")
	desc.Disabled = getBoolField(L, absIdx, "disabled")

	// Read children
	L.GetField(absIdx, "children")
	if L.IsTable(-1) {
		childrenIdx := L.AbsIndex(-1)
		n := int(L.RawLen(childrenIdx))
		desc.Children = make([]Descriptor, 0, n)
		for i := 1; i <= n; i++ {
			L.RawGetI(childrenIdx, int64(i))
			if L.IsTable(-1) {
				child := e.readDescriptor(L, -1)
				desc.Children = append(desc.Children, child)
			} else if L.IsString(-1) {
				// String child → text descriptor
				s, _ := L.ToString(-1)
				desc.Children = append(desc.Children, Descriptor{
					Type:    "text",
					Content: s,
				})
			}
			L.Pop(1)
		}
	}
	L.Pop(1)

	// Check if this is a component type
	factoryName := getStringField(L, absIdx, "_factoryName")
	if factoryName != "" {
		desc.Type = "component"
		desc.ComponentType = factoryName
		L.GetField(absIdx, "_props")
		if L.IsTable(-1) {
			desc.ComponentProps = readMapFromTable(L, -1)
		}
		L.Pop(1)
	}

	// Backward compat: input/textarea are always focusable
	if desc.Type == "input" || desc.Type == "textarea" {
		desc.Focusable = true
	}

	return desc
}

// readStyle reads a style table from the Lua stack.
func (e *Engine) readStyle(L *lua.State, idx int) Style {
	absIdx := L.AbsIndex(idx)
	var s Style
	s.Width = int(getIntField(L, absIdx, "width"))
	s.Height = int(getIntField(L, absIdx, "height"))
	s.Flex = int(getIntField(L, absIdx, "flex"))
	s.Padding = int(getIntField(L, absIdx, "padding"))
	s.PaddingTop = int(getIntField(L, absIdx, "paddingTop"))
	s.PaddingBottom = int(getIntField(L, absIdx, "paddingBottom"))
	s.PaddingLeft = int(getIntField(L, absIdx, "paddingLeft"))
	s.PaddingRight = int(getIntField(L, absIdx, "paddingRight"))
	s.Margin = int(getIntField(L, absIdx, "margin"))
	s.MarginTop = int(getIntField(L, absIdx, "marginTop"))
	s.MarginBottom = int(getIntField(L, absIdx, "marginBottom"))
	s.MarginLeft = int(getIntField(L, absIdx, "marginLeft"))
	s.MarginRight = int(getIntField(L, absIdx, "marginRight"))
	s.Gap = int(getIntField(L, absIdx, "gap"))
	s.MinWidth = int(getIntField(L, absIdx, "minWidth"))
	s.MaxWidth = int(getIntField(L, absIdx, "maxWidth"))
	s.MinHeight = int(getIntField(L, absIdx, "minHeight"))
	s.MaxHeight = int(getIntField(L, absIdx, "maxHeight"))
	s.Justify = getStringField(L, absIdx, "justify")
	s.Align = getStringField(L, absIdx, "align")
	s.Border = getStringField(L, absIdx, "border")
	s.Foreground = getStringField(L, absIdx, "foreground")
	if fg := getStringField(L, absIdx, "fg"); fg != "" && s.Foreground == "" {
		s.Foreground = fg
	}
	s.Background = getStringField(L, absIdx, "background")
	if bg := getStringField(L, absIdx, "bg"); bg != "" && s.Background == "" {
		s.Background = bg
	}
	s.Bold = getBoolField(L, absIdx, "bold")
	s.Dim = getBoolField(L, absIdx, "dim")
	s.Underline = getBoolField(L, absIdx, "underline")
	s.Overflow = getStringField(L, absIdx, "overflow")
	s.Position = getStringField(L, absIdx, "position")
	s.Top = int(getIntField(L, absIdx, "top"))
	s.Left = int(getIntField(L, absIdx, "left"))
	s.Right = int(getIntFieldDefault(L, absIdx, "right", -1))
	s.Bottom = int(getIntFieldDefault(L, absIdx, "bottom", -1))
	s.ZIndex = int(getIntField(L, absIdx, "zIndex"))
	return s
}

// readStyleFields reads style fields from the top-level table (not a nested "style" sub-table).
// Only sets fields that are still at their zero/default value.
func (e *Engine) readStyleFields(L *lua.State, idx int, s *Style) {
	absIdx := L.AbsIndex(idx)

	if s.Width == 0 {
		s.Width = int(getIntField(L, absIdx, "width"))
	}
	if s.Height == 0 {
		s.Height = int(getIntField(L, absIdx, "height"))
	}
	if s.MinWidth == 0 {
		s.MinWidth = int(getIntField(L, absIdx, "minWidth"))
	}
	if s.MinHeight == 0 {
		s.MinHeight = int(getIntField(L, absIdx, "minHeight"))
	}
	if s.MaxWidth == 0 {
		s.MaxWidth = int(getIntField(L, absIdx, "maxWidth"))
	}
	if s.MaxHeight == 0 {
		s.MaxHeight = int(getIntField(L, absIdx, "maxHeight"))
	}
	if s.Flex == 0 {
		s.Flex = int(getIntField(L, absIdx, "flex"))
	}
	if s.Gap == 0 {
		s.Gap = int(getIntField(L, absIdx, "gap"))
	}
	if s.Padding == 0 {
		s.Padding = int(getIntField(L, absIdx, "padding"))
	}
	if s.PaddingTop == 0 {
		s.PaddingTop = int(getIntField(L, absIdx, "paddingTop"))
	}
	if s.PaddingRight == 0 {
		s.PaddingRight = int(getIntField(L, absIdx, "paddingRight"))
	}
	if s.PaddingBottom == 0 {
		s.PaddingBottom = int(getIntField(L, absIdx, "paddingBottom"))
	}
	if s.PaddingLeft == 0 {
		s.PaddingLeft = int(getIntField(L, absIdx, "paddingLeft"))
	}
	if s.Margin == 0 {
		s.Margin = int(getIntField(L, absIdx, "margin"))
	}
	if s.MarginTop == 0 {
		s.MarginTop = int(getIntField(L, absIdx, "marginTop"))
	}
	if s.MarginRight == 0 {
		s.MarginRight = int(getIntField(L, absIdx, "marginRight"))
	}
	if s.MarginBottom == 0 {
		s.MarginBottom = int(getIntField(L, absIdx, "marginBottom"))
	}
	if s.MarginLeft == 0 {
		s.MarginLeft = int(getIntField(L, absIdx, "marginLeft"))
	}
	if s.Foreground == "" {
		s.Foreground = getStringField(L, absIdx, "foreground")
		if s.Foreground == "" {
			s.Foreground = getStringField(L, absIdx, "fg")
		}
	}
	if s.Background == "" {
		s.Background = getStringField(L, absIdx, "background")
		if s.Background == "" {
			s.Background = getStringField(L, absIdx, "bg")
		}
	}
	if s.Border == "" {
		s.Border = getStringField(L, absIdx, "border")
	}
	if s.Justify == "" {
		s.Justify = getStringField(L, absIdx, "justify")
	}
	if s.Align == "" {
		s.Align = getStringField(L, absIdx, "align")
	}
	if s.Overflow == "" {
		s.Overflow = getStringField(L, absIdx, "overflow")
	}
	if s.Position == "" {
		s.Position = getStringField(L, absIdx, "position")
	}
	if !s.Bold {
		s.Bold = getBoolField(L, absIdx, "bold")
	}
	if !s.Dim {
		s.Dim = getBoolField(L, absIdx, "dim")
	}
	if !s.Underline {
		s.Underline = getBoolField(L, absIdx, "underline")
	}
	if s.Top == 0 {
		s.Top = int(getIntField(L, absIdx, "top"))
	}
	if s.Left == 0 {
		s.Left = int(getIntField(L, absIdx, "left"))
	}
	if s.Right == -1 {
		s.Right = int(getIntFieldDefault(L, absIdx, "right", -1))
	}
	if s.Bottom == -1 {
		s.Bottom = int(getIntFieldDefault(L, absIdx, "bottom", -1))
	}
	if s.ZIndex == 0 {
		s.ZIndex = int(getIntField(L, absIdx, "zIndex"))
	}
}

// reconcileChildComponents walks the RenderNode tree looking for component-type
// nodes and reconciles child components.
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

	// Recurse into children
	for _, ch := range node.Children {
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
			// Reconstruct the map key from type + lookup key
			lookupKey := ""
			if child.ID != "" {
				parts := splitAfterColon(child.ID, parent.ID)
				if parts != "" {
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

	// Clear capturedComp if it points to this component (prevent zombie capture)
	if e.capturedComp == comp {
		e.capturedComp = nil
	}

	// Clean up widget state for this component (prevent memory leak + stale state)
	delete(e.widgetStates, comp.ID)

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
	for iterations := 0; iterations < 10; iterations++ {
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
		sort.Slice(dirty, func(i, j int) bool {
			return componentDepth(dirty[i]) < componentDepth(dirty[j])
		})
		for _, comp := range dirty {
			if !comp.Dirty {
				continue // may have been rendered as side effect
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

// graftChildComponents walks the root tree and connects child component
// RootNodes as children of their placeholder nodes. This allows layout and
// paint to naturally traverse into sub-components.
func (e *Engine) graftChildComponents() {
	if e.root == nil || e.root.RootNode == nil {
		return
	}
	e.graftWalk(e.root.RootNode)
}

// graftWalk recursively finds component placeholder nodes and grafts the
// child component's RootNode as the placeholder's child.
// Only marks dirty when the graft actually changes (new or different RootNode).
func (e *Engine) graftWalk(node *Node) {
	if node == nil {
		return
	}

	// Handle the case where node itself is a component placeholder
	// (happens when root render returns a defineComponent directly)
	if node.Type == "component" && node.Component != nil {
		comp := node.Component
		if comp.RootNode != nil {
			alreadyGrafted := len(node.Children) == 1 && node.Children[0] == comp.RootNode
			if !alreadyGrafted {
				node.Children = []*Node{comp.RootNode}
				comp.RootNode.Parent = node
				node.LayoutDirty = true
				node.PaintDirty = true
			}
			if _, isWidget := e.widgets[comp.Type]; isWidget {
				copyWidgetEventHandlers(node, comp.RootNode)
			}
		}
	}

	for _, child := range node.Children {
		if child.Type == "component" && child.Component != nil {
			comp := child.Component
			if comp.RootNode != nil {
				// Only mark dirty if the grafted child actually changed
				alreadyGrafted := len(child.Children) == 1 && child.Children[0] == comp.RootNode
				if !alreadyGrafted {
					child.Children = []*Node{comp.RootNode}
					comp.RootNode.Parent = child
					child.LayoutDirty = true
					child.PaintDirty = true
				}
				// For Go widgets, copy event handlers from placeholder to root node
				if _, isWidget := e.widgets[comp.Type]; isWidget {
					copyWidgetEventHandlers(child, comp.RootNode)
				}
			}
		}
		// Always recurse (component children may contain nested components)
		e.graftWalk(child)
	}
}


// RegisterLuaAPI registers lumina.createElement, lumina.useState,
// lumina.defineComponent, lumina.createComponent on the Lua global table.
func (e *Engine) RegisterLuaAPI() {
	L := e.L

	// Create or get the "lumina" global table
	L.GetGlobal("lumina")
	if !L.IsTable(-1) {
		L.Pop(1)
		L.NewTable()
	}
	tblIdx := L.AbsIndex(-1)

	// lumina.createElement(type, props, children...)
	L.PushFunction(e.luaCreateElement)
	L.SetField(tblIdx, "createElement")

	// lumina.defineComponent(name, renderFn) → factory table
	L.PushFunction(e.luaDefineComponent)
	L.SetField(tblIdx, "defineComponent")

	// lumina.createComponent(config) — root component
	L.PushFunction(e.luaCreateComponent)
	L.SetField(tblIdx, "createComponent")

	// lumina.useState(key, initial) → value, setter
	L.PushFunction(e.luaUseState)
	L.SetField(tblIdx, "useState")

	// lumina.useEffect(callback, deps?)
	L.PushFunction(e.luaUseEffect)
	L.SetField(tblIdx, "useEffect")

	// lumina.useRef(initialValue?)
	L.PushFunction(e.luaUseRef)
	L.SetField(tblIdx, "useRef")

	// lumina.useMemo(factory, deps)
	L.PushFunction(e.luaUseMemo)
	L.SetField(tblIdx, "useMemo")

	// lumina.useCallback(fn, deps)
	L.PushFunction(e.luaUseCallback)
	L.SetField(tblIdx, "useCallback")

	// lumina.spawn(fn) — start async coroutine
	L.PushFunction(e.luaSpawn)
	L.SetField(tblIdx, "spawn")

	// lumina.cancel(handle) — cancel a spawned coroutine
	L.PushFunction(e.luaCancel)
	L.SetField(tblIdx, "cancel")

	// lumina.sleep(ms) — returns Future
	L.PushFunction(e.luaSleep)
	L.SetField(tblIdx, "sleep")

	// lumina.exec(cmd) — returns Future
	L.PushFunction(e.luaExec)
	L.SetField(tblIdx, "exec")

	// lumina.readFile(path) — returns Future
	L.PushFunction(e.luaReadFile)
	L.SetField(tblIdx, "readFile")

	// Create shared callable metatable for factory tables (__call → createElement)
	L.NewTable()
	L.PushFunction(e.luaFactoryCall)
	L.SetField(-2, "__call")
	sharedMetaIdx := L.AbsIndex(-1)

	// Register Go widgets as Lua-accessible factories (e.g., lumina.Button)
	for name := range e.widgets {
		L.NewTable()
		factoryIdx := L.AbsIndex(-1)
		L.PushBoolean(true)
		L.SetField(factoryIdx, "_isFactory")
		L.PushString(name)
		L.SetField(factoryIdx, "_name")
		// Set callable metatable so lumina.Button { props } works
		L.PushValue(sharedMetaIdx)
		L.SetMetatable(factoryIdx)
		L.SetField(tblIdx, name) // lumina.Button = {_isFactory=true, _name="Button"}
	}

	// Store shared metatable as registry ref for reuse by defineComponent
	e.factoryMetaRef = int64(L.Ref(lua.RegistryIndex)) // pops metatable

	// lumina.getTheme() → returns theme color table
	L.PushFunction(e.luaGetTheme)
	L.SetField(tblIdx, "getTheme")

	L.SetGlobal("lumina")
}

// luaGetTheme implements lumina.getTheme() → returns theme color table.
// Uses Engine.ThemeGetter to avoid import cycles (render cannot import widget).
func (e *Engine) luaGetTheme(L *lua.State) int {
	if e.ThemeGetter == nil {
		L.NewTable()
		return 1
	}
	theme := e.ThemeGetter()
	L.NewTable()
	for k, v := range theme {
		L.PushString(v)
		L.SetField(-2, k)
	}
	return 1
}

// luaDefineComponent implements lumina.defineComponent(name, renderFn)
// Returns a factory table: {_isFactory=true, _name=name}
func (e *Engine) luaDefineComponent(L *lua.State) int {
	name := L.CheckString(1)
	L.CheckType(2, lua.TypeFunction)

	// Store render function as registry ref
	L.PushValue(2)
	ref := L.Ref(lua.RegistryIndex)
	e.factories[name] = int64(ref)

	// Return a factory table that createElement can detect
	L.NewTable()
	resultIdx := L.AbsIndex(-1)
	L.PushBoolean(true)
	L.SetField(resultIdx, "_isFactory")
	L.PushString(name)
	L.SetField(resultIdx, "_name")

	// Set callable metatable so Factory { props } works
	e.setFactoryMetatable(L, resultIdx)

	return 1
}

// luaFactoryCall implements the __call metamethod for factory tables.
// Supports two calling patterns:
//
// Pattern 1 (standard): Factory(props, child1, child2)
//
//	__call receives: self, props, child1, child2
//	→ delegate to createElement(self, props, child1, child2)
//
// Pattern 2 (mixed single table): Factory { prop1=v1, Child1{}, Child2{} }
//
//	__call receives: self, mixedTable
//	→ split mixedTable: string keys → props, integer keys → children
//	→ rebuild stack as createElement(self, extractedProps, child1, child2, ...)
func (e *Engine) luaFactoryCall(L *lua.State) int {
	nArgs := L.GetTop()

	// Pattern 1: no args, or multiple args → standard delegation
	if nArgs != 2 || !L.IsTable(2) {
		return e.luaCreateElement(L)
	}

	// nArgs == 2, arg2 is a table. Check for integer keys (children).
	mixedIdx := 2
	childCount := int(L.LenI(mixedIdx))

	if childCount == 0 {
		// No integer keys → pure props table → standard delegation
		return e.luaCreateElement(L)
	}

	// Check if integer values are tables (descriptors/children) vs strings (content).
	// If first integer value is NOT a table, treat as standard (e.g., Text { "hello" }).
	L.RawGetI(mixedIdx, 1)
	firstIsTable := L.IsTable(-1)
	L.Pop(1)

	if !firstIsTable {
		// Integer values are strings/numbers → not mixed children pattern
		return e.luaCreateElement(L)
	}

	// === Pattern 2: mixed table ===
	// Split mixed table into props (string keys) and children (integer keys).
	// Build new stack: [1]=self, [2]=cleanProps, [3..]=children
	// without using registry refs (avoids go-lua Ref slot reuse bug).

	// Stack: [1]=self, [2]=mixed

	// Step 1: Build clean props table (string keys only) on top of stack
	L.NewTable() // [3] = newProps
	newPropsIdx := L.AbsIndex(-1)
	L.PushNil()
	for L.Next(mixedIdx) {
		// key at -2, value at -1
		if L.Type(-2) == lua.TypeString {
			L.PushValue(-2) // push key copy
			L.PushValue(-2) // push value copy
			L.SetTable(newPropsIdx)
		}
		L.Pop(1) // pop value, keep key for Next
	}
	// Stack: [1]=self, [2]=mixed, [3]=newProps

	// Step 2: Push children from mixed table
	for i := 1; i <= childCount; i++ {
		L.RawGetI(mixedIdx, int64(i)) // push child[i]
	}
	// Stack: [1]=self, [2]=mixed, [3]=newProps, [4..3+childCount]=children

	// Step 3: Remove mixed table at [2], shifting everything down
	L.Remove(2)
	// Stack: [1]=self, [2]=newProps, [3..2+childCount]=children

	// Stack is now: [1]=factory, [2]=props, [3..N]=children
	return e.luaCreateElement(L)
}

// setFactoryMetatable sets the shared __call metatable on a factory table at the given index.
// This enables the syntax: Factory { props } or Factory(props, child1, child2)
func (e *Engine) setFactoryMetatable(L *lua.State, tableIdx int) {
	absIdx := L.AbsIndex(tableIdx)
	L.RawGetI(lua.RegistryIndex, e.factoryMetaRef)
	L.SetMetatable(absIdx)
}

// luaCreateElement implements lumina.createElement(type_or_factory, props, children...)
func (e *Engine) luaCreateElement(L *lua.State) int {
	nArgs := L.GetTop()

	// Check if first arg is a factory table (from defineComponent)
	if L.IsTable(1) {
		L.GetField(1, "_isFactory")
		isFactory := L.ToBoolean(-1)
		L.Pop(1)

		if isFactory {
			return e.luaCreateComponentElement(L, nArgs)
		}
	}

	// Normal element: type is a string
	nodeType := L.CheckString(1)

	// Create result table
	L.NewTable()
	resultIdx := L.AbsIndex(-1)

	L.PushString(nodeType)
	L.SetField(resultIdx, "type")

	// Copy props
	if nArgs >= 2 && L.IsTable(2) {
		L.ForEach(2, func(L *lua.State) bool {
			if L.Type(-2) == lua.TypeString {
				key, _ := L.ToString(-2)
				L.PushValue(-1)
				L.SetField(resultIdx, key)
			}
			return true
		})
	}

	// Handle children (args 3+)
	if nArgs > 2 {
		hasTable := false
		for i := 3; i <= nArgs; i++ {
			if L.Type(i) == lua.TypeTable {
				hasTable = true
				break
			}
		}

		if !hasTable {
			// String children → content
			var parts []string
			for i := 3; i <= nArgs; i++ {
				if L.Type(i) == lua.TypeString {
					s, _ := L.ToString(i)
					parts = append(parts, s)
				}
			}
			if len(parts) > 0 {
				L.PushString(strings.Join(parts, ""))
				L.SetField(resultIdx, "content")
			}
		} else {
			// Table children → children array
			L.CreateTable(nArgs-2, 0)
			childrenIdx := L.AbsIndex(-1)
			for i := 3; i <= nArgs; i++ {
				L.PushValue(i)
				L.RawSetI(childrenIdx, int64(i-2))
			}
			L.SetField(resultIdx, "children")
		}
	}

	return 1
}

// luaCreateComponentElement handles createElement(Factory, props)
func (e *Engine) luaCreateComponentElement(L *lua.State, nArgs int) int {
	// Get factory name
	L.GetField(1, "_name")
	factoryName, _ := L.ToString(-1)
	L.Pop(1)

	// Create a component descriptor table
	L.NewTable()
	resultIdx := L.AbsIndex(-1)

	L.PushString("component")
	L.SetField(resultIdx, "type")

	L.PushString(factoryName)
	L.SetField(resultIdx, "_factoryName")

	// Copy props (including key)
	if nArgs >= 2 && L.IsTable(2) {
		// Store full props table
		L.PushValue(2)
		L.SetField(resultIdx, "_props")

		// Extract key for reconciliation
		L.GetField(2, "key")
		if L.IsString(-1) {
			s, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString(s)
			L.SetField(resultIdx, "key")
		} else {
			L.Pop(1)
		}

		// Extract id for reconciliation
		L.GetField(2, "id")
		if L.IsString(-1) {
			s, _ := L.ToString(-1)
			L.Pop(1)
			L.PushString(s)
			L.SetField(resultIdx, "id")
		} else {
			L.Pop(1)
		}

		// Copy event handler functions to top level so readDescriptor picks them up.
		// This allows onClick etc. to be set on the placeholder Node and fire via
		// the normal event bubbling system.
		eventKeys := []string{
			"onClick", "onMouseEnter", "onMouseLeave", "onKeyDown",
			"onChange", "onScroll", "onMouseDown", "onMouseUp",
			"onFocus", "onBlur", "onSubmit", "onOutsideClick",
		}
		for _, key := range eventKeys {
			L.GetField(2, key)
			if L.IsFunction(-1) {
				L.SetField(resultIdx, key) // pops the function
			} else {
				L.Pop(1)
			}
		}

		// Copy disabled/focusable fields if present
		L.GetField(2, "disabled")
		if L.IsBoolean(-1) {
			L.SetField(resultIdx, "disabled")
		} else {
			L.Pop(1)
		}
		L.GetField(2, "focusable")
		if L.IsBoolean(-1) {
			L.SetField(resultIdx, "focusable")
		} else {
			L.Pop(1)
		}
	}

	// Collect children (args 3+) and inject into _props.children
	if nArgs > 2 {
		// Count non-nil children
		childCount := 0
		for i := 3; i <= nArgs; i++ {
			if !L.IsNoneOrNil(i) {
				childCount++
			}
		}

		if childCount > 0 {
			// Ensure we have a _props table
			L.GetField(resultIdx, "_props")
			if L.IsNil(-1) {
				L.Pop(1)
				L.NewTable()
				L.PushValue(-1) // dup for SetField
				L.SetField(resultIdx, "_props")
			}
			propsIdx := L.AbsIndex(-1)

			// Create children array
			L.CreateTable(childCount, 0)
			childrenIdx := L.AbsIndex(-1)
			idx := int64(1)
			for i := 3; i <= nArgs; i++ {
				if !L.IsNoneOrNil(i) {
					L.PushValue(i)
					L.RawSetI(childrenIdx, idx)
					idx++
				}
			}

			// Set props.children = childrenArray
			L.SetField(propsIdx, "children")

			L.Pop(1) // pop _props
		}
	}

	return 1
}

// luaCreateComponent implements lumina.createComponent(config) for root components
func (e *Engine) luaCreateComponent(L *lua.State) int {
	L.CheckType(1, lua.TypeTable)
	absIdx := L.AbsIndex(1)

	id := getStringField(L, absIdx, "id")
	if id == "" {
		L.PushString("createComponent: 'id' is required")
		L.Error()
		return 0
	}

	name := getStringField(L, absIdx, "name")
	if name == "" {
		name = id
	}

	// Get render function ref
	L.GetField(absIdx, "render")
	if !L.IsFunction(-1) {
		L.Pop(1)
		L.PushString("createComponent: 'render' function is required")
		L.Error()
		return 0
	}
	ref := L.Ref(lua.RegistryIndex)

	e.CreateRootComponent(id, name, int64(ref))
	return 0
}

// luaUseState implements lumina.useState(key, initial) → value, setter
func (e *Engine) luaUseState(L *lua.State) int {
	comp := e.currentComp
	if comp == nil {
		L.PushString("useState: no current component")
		L.Error()
		return 0
	}

	key := L.CheckString(1)

	// Initialize if not exists
	if _, exists := comp.State[key]; !exists {
		var initial any
		if L.GetTop() >= 2 && !L.IsNoneOrNil(2) {
			initial = L.ToAny(2)
		}
		comp.State[key] = initial
	}

	// Push current value
	L.PushAny(comp.State[key])

	// Push setter function
	compID := comp.ID
	L.PushFunction(func(L *lua.State) int {
		newValue := L.ToAny(1)
		e.SetState(compID, key, newValue)
		return 0
	})

	return 2
}

// --- Helper functions for reading Lua tables ---

// propFuncRef is a Lua registry reference for a function stored in ComponentProps.
// Plain int64 in props would round-trip as a Lua number via PushAny; propFuncRef
// is restored in pushMap via RawGetI(registry, ref).
type propFuncRef int64

// collectPropFuncRefsFromAny walks values produced by readMapFromTable / readPropTable
// and collects nested propFuncRef ids (registry indices).
func collectPropFuncRefsFromAny(v any, out *[]int64) {
	switch x := v.(type) {
	case propFuncRef:
		if x != 0 {
			*out = append(*out, int64(x))
		}
	case map[string]any:
		for _, vv := range x {
			collectPropFuncRefsFromAny(vv, out)
		}
	case []any:
		for _, vv := range x {
			collectPropFuncRefsFromAny(vv, out)
		}
	}
}

// unrefPropFuncRefsInProps releases Lua registry refs held in a props map. Must be
// called before dropping the map: each render builds a new readMapFromTable result
// with new Ref() ids for the same logical functions, so propsEqual is usually false
// every frame and child.Props is reassigned without otherwise freeing old refs.
func unrefPropFuncRefsInProps(L *lua.State, m map[string]any) {
	if L == nil || m == nil {
		return
	}
	var refs []int64
	collectPropFuncRefsFromAny(m, &refs)
	for _, ref := range refs {
		L.Unref(lua.RegistryIndex, int(ref))
	}
}

func pushMap(L *lua.State, m map[string]any) {
	if m == nil {
		L.NewTable()
		return
	}
	L.CreateTable(0, len(m))
	for k, v := range m {
		L.PushString(k)
		switch vv := v.(type) {
		case propFuncRef:
			if vv != 0 {
				L.RawGetI(lua.RegistryIndex, int64(vv))
			} else {
				L.PushNil()
			}
		default:
			pushPropValue(L, v)
		}
		L.SetTable(-3)
	}
}

// pushPropValue pushes a value stored in ComponentProps, preserving nested
// propFuncRef (Lua functions) that plain PushAny cannot represent.
func pushPropValue(L *lua.State, v any) {
	if v == nil {
		L.PushNil()
		return
	}
	if pf, ok := v.(propFuncRef); ok {
		if pf != 0 {
			L.RawGetI(lua.RegistryIndex, int64(pf))
		} else {
			L.PushNil()
		}
		return
	}
	switch vv := v.(type) {
	case map[string]any:
		L.CreateTable(0, len(vv))
		for k, val := range vv {
			L.PushString(k)
			pushPropValue(L, val)
			L.RawSet(-3)
		}
	case []any:
		L.CreateTable(len(vv), 0)
		for i, val := range vv {
			pushPropValue(L, val)
			L.RawSetI(-2, int64(i+1))
		}
	default:
		L.PushAny(v)
	}
}

func getStringField(L *lua.State, idx int, field string) string {
	return L.GetFieldString(idx, field)
}

func getIntField(L *lua.State, idx int, field string) int64 {
	return L.GetFieldInt(idx, field)
}

func getIntFieldDefault(L *lua.State, idx int, field string, def int64) int64 {
	L.GetField(idx, field)
	if L.IsNoneOrNil(-1) {
		L.Pop(1)
		return def
	}
	n, _ := L.ToInteger(-1)
	L.Pop(1)
	return n
}

func getBoolField(L *lua.State, idx int, field string) bool {
	return L.GetFieldBool(idx, field)
}

func getRefField(L *lua.State, idx int, field string) int64 {
	L.GetField(idx, field)
	if L.IsFunction(-1) {
		ref := L.Ref(lua.RegistryIndex)
		return int64(ref)
	}
	L.Pop(1)
	return 0
}

func readMapFromTable(L *lua.State, idx int) map[string]any {
	m := make(map[string]any)
	absIdx := L.AbsIndex(idx)
	L.ForEach(absIdx, func(L *lua.State) bool {
		if L.Type(-2) == lua.TypeString {
			key, _ := L.ToString(-2)
			m[key] = readPropValueFromStack(L)
		}
		return true
	})
	return m
}

func luaPropTableKeyString(L *lua.State, keyIdx int) string {
	switch L.Type(keyIdx) {
	case lua.TypeString:
		s, _ := L.ToString(keyIdx)
		return s
	case lua.TypeNumber:
		if L.IsInteger(keyIdx) {
			v, _ := L.ToInteger(keyIdx)
			return strconv.FormatInt(v, 10)
		}
		v, _ := L.ToNumber(keyIdx)
		return strconv.FormatFloat(v, 'g', -1, 64)
	default:
		s, _ := L.ToString(keyIdx)
		return s
	}
}

// readPropValueFromStack reads the Lua value at stack index -1 (without popping it).
// Used for ComponentProps so nested descriptor tables keep onClick etc. as propFuncRef.
// (L.ToAny maps Lua functions to nil.)
func readPropValueFromStack(L *lua.State) any {
	switch L.Type(-1) {
	case lua.TypeNil:
		return nil
	case lua.TypeBoolean:
		return L.ToBoolean(-1)
	case lua.TypeNumber:
		if L.IsInteger(-1) {
			v, _ := L.ToInteger(-1)
			return v
		}
		v, _ := L.ToNumber(-1)
		return v
	case lua.TypeString:
		s, _ := L.ToString(-1)
		return s
	case lua.TypeFunction:
		L.PushValue(-1)
		ref := L.Ref(lua.RegistryIndex)
		return propFuncRef(ref)
	case lua.TypeTable:
		L.PushValue(-1)
		tIdx := L.AbsIndex(-1)
		out := readPropTable(L, tIdx)
		L.Pop(1)
		return out
	default:
		return L.ToAny(-1)
	}
}

func readPropTable(L *lua.State, idx int) any {
	idx = L.AbsIndex(idx)
	length := int(L.LenI(idx))
	if length > 0 {
		var count int64
		L.PushNil()
		for L.Next(idx) {
			count++
			L.Pop(1)
		}
		if count == int64(length) {
			arr := make([]any, length)
			for i := 1; i <= length; i++ {
				L.RawGetI(idx, int64(i))
				arr[i-1] = readPropValueFromStack(L)
				L.Pop(1)
			}
			return arr
		}
	}
	m := make(map[string]any)
	L.PushNil()
	for L.Next(idx) {
		key := luaPropTableKeyString(L, -2)
		m[key] = readPropValueFromStack(L)
		L.Pop(1)
	}
	return m
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
				Char:       c.Ch,
				Foreground: c.FG,
				Background: c.BG,
				Bold:       c.Bold,
				Dim:        c.Dim,
				Underline:  c.Underline,
				Wide:       isWideChar, // Wide on MAIN cell, not padding
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
