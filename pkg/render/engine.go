package render

import (
	"sort"
	"strings"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/buffer"
	"github.com/akzj/lumina/pkg/perf"
)

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
		L:          L,
		components: make(map[string]*Component),
		factories:  make(map[string]int64),
		width:      width,
		height:     height,
		buffer:     NewCellBuffer(width, height),
	}
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
	if e.root != nil && e.root.RootNode != nil {
		e.root.RootNode.MarkLayoutDirty()
		e.needsRender = true
	}
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

	// 3. Early exit: if nothing rendered and no dirty nodes, skip layout/paint
	if rendered == 0 && e.root != nil && e.root.RootNode != nil && !hasAnyDirty(e.root.RootNode) {
		if e.tracker != nil {
			e.tracker.Record(perf.V2PaintCells, 0)
			e.tracker.Record(perf.V2PaintClearCells, 0)
			e.tracker.Record(perf.V2DirtyRectArea, 0)
		}
		return
	}

	// 4. Layout: only the root tree (which now contains grafted children)
	if e.root != nil && e.root.RootNode != nil {
		if e.root.RootNode.LayoutDirty {
			LayoutFull(e.root.RootNode, 0, 0, e.width, e.height)
		} else {
			LayoutIncremental(e.root.RootNode)
		}
	}

	// 5. Paint: only the root tree (grafted children are painted in-tree)
	if e.root != nil && e.root.RootNode != nil {
		PaintDirty(e.buffer, e.root.RootNode)
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

	// Force full layout + paint on root (which now includes grafted children)
	if e.root != nil && e.root.RootNode != nil {
		LayoutFull(e.root.RootNode, 0, 0, e.width, e.height)
		PaintFull(e.buffer, e.root.RootNode)
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
	// For input/textarea, also check "value" field
	if desc.Content == "" {
		if v := getStringField(L, absIdx, "value"); v != "" {
			desc.Content = v
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
			parent.AddChild(child, lookupKey)
			e.components[childID] = child
			child.Dirty = true
		}
		// Existing child: do NOT mark dirty unless props changed.
		// Child state changes (hover, click) mark themselves dirty via setState.
		node.Component = child
		return
	}

	// Recurse into children
	for _, ch := range node.Children {
		e.reconcileChildComponents(parent, ch)
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
	// Cleanup hook refs (effects, memos, refs) — runs effect cleanups
	e.cleanupComponentHooks(comp)
	// Unref the component's render function
	if comp.RenderFn != 0 {
		e.L.Unref(lua.RegistryIndex, int(comp.RenderFn))
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
	// Collect dirty non-root components, sort by depth (parent before child)
	var dirty []*Component
	for _, comp := range e.components {
		if !comp.Dirty || comp.IsRoot {
			continue
		}
		dirty = append(dirty, comp)
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

	L.SetGlobal("lumina")
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

	return 1
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

func pushMap(L *lua.State, m map[string]any) {
	if m == nil {
		L.NewTable()
		return
	}
	L.NewTableFrom(m)
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
			m[key] = L.ToAny(-1)
		}
		return true
	})
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
