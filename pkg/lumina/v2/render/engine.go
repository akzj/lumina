package render

import (
	"strings"

	"github.com/akzj/go-lua/pkg/lua"
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

// Resize updates the engine dimensions and buffer.
func (e *Engine) Resize(width, height int) {
	e.width = width
	e.height = height
	e.buffer.Resize(width, height)
	if e.root != nil && e.root.RootNode != nil {
		e.root.RootNode.MarkLayoutDirty()
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
}

// SetState sets a state value on a component and marks it dirty.
func (e *Engine) SetState(compID, key string, value any) {
	comp := e.components[compID]
	if comp == nil {
		return
	}
	comp.SetState(key, value)
}

// RenderDirty renders all dirty components, reconciles, layouts, and paints.
// This is the main frame function.
func (e *Engine) RenderDirty() {
	// 1. Render dirty components (call Lua, reconcile)
	for _, comp := range e.components {
		if !comp.Dirty {
			continue
		}
		e.renderComponent(comp)
	}

	// 2. Incremental layout (only LayoutDirty subtrees)
	for _, comp := range e.components {
		if comp.RootNode == nil {
			continue
		}
		if comp.IsRoot {
			if comp.RootNode.LayoutDirty {
				LayoutFull(comp.RootNode, 0, 0, e.width, e.height)
			}
		} else {
			LayoutIncremental(comp.RootNode)
		}
	}

	// 3. Incremental paint (only PaintDirty nodes)
	for _, comp := range e.components {
		if comp.RootNode == nil {
			continue
		}
		PaintDirty(e.buffer, comp.RootNode)
	}
}

// RenderAll does a full render of everything (initial mount).
func (e *Engine) RenderAll() {
	for _, comp := range e.components {
		comp.Dirty = true
	}

	// Render all components
	for _, comp := range e.components {
		if !comp.Dirty {
			continue
		}
		e.renderComponent(comp)
	}

	// Force full layout + paint on root
	if e.root != nil && e.root.RootNode != nil {
		LayoutFull(e.root.RootNode, 0, 0, e.width, e.height)
		PaintFull(e.buffer, e.root.RootNode)
	}
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
		L.Pop(1) // pop error
		comp.Dirty = false
		return
	}

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
		// Update: reconcile (diff + patch in-place)
		Reconcile(comp.RootNode, desc)
	}

	// Handle sub-component children
	e.reconcileChildComponents(comp, comp.RootNode)

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
	if s.Flex == 0 {
		s.Flex = int(getIntField(L, absIdx, "flex"))
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
	if !s.Bold {
		s.Bold = getBoolField(L, absIdx, "bold")
	}
}

// reconcileChildComponents walks the RenderNode tree looking for component-type
// nodes and reconciles child components.
func (e *Engine) reconcileChildComponents(parent *Component, node *Node) {
	if node == nil {
		return
	}

	// If this node represents a sub-component, handle it
	if node.Type == "component" && node.Key != "" {
		factoryName := node.Key
		child := parent.FindChild(factoryName, node.ID)
		if child == nil {
			// Create new child component
			renderRef, ok := e.factories[factoryName]
			if !ok {
				return
			}
			childID := parent.ID + ":" + node.ID
			if node.ID == "" {
				childID = parent.ID + ":" + factoryName
			}
			child = NewComponent(childID, factoryName, factoryName)
			child.RenderFn = renderRef
			child.Parent = parent
			parent.AddChild(child)
			e.components[childID] = child
			child.Dirty = true
		}
		node.Component = child
		return
	}

	// Recurse into children
	for _, ch := range node.Children {
		e.reconcileChildComponents(parent, ch)
	}
}

// --- Lua API Registration ---

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
	if n == 0 {
		return def
	}
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
