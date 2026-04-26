// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// sortByZIndex sorts VNodes by Style.ZIndex ascending (lower z-index first).
func sortByZIndex(nodes []*VNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Style.ZIndex < nodes[j].Style.ZIndex
	})
}

// VNode represents a virtual DOM node.
type VNode struct {
	Type     string
	Props    map[string]any
	Style    Style // parsed layout/visual style
	Children []*VNode
	// Layout properties
	X, Y, W, H int
	// Content for text nodes
	Content string
	// Focused indicates if this component is currently focused
	Focused bool
	// ComponentRef holds the Component instance that produced this VNode (if any).
	ComponentRef *Component
	// ComponentKey is a stable identity for reconciliation of component children.
	ComponentKey string
	// Portal fields
	IsPortal     bool   // true if this is a portal node
	PortalTarget string // target container ID for portal rendering
	// Ref support (forwardRef)
	Ref string // ref ID for forwardRef
	// Style cache — avoids re-parsing props on every layout pass
	cachedStyle *Style
}

// NewVNode creates a new VNode.
func NewVNode(nodeType string) *VNode {
	return &VNode{
		Type:     nodeType,
		Props:    make(map[string]any),
		Children: make([]*VNode, 0),
	}
}

// AddChild adds a child VNode.
func (v *VNode) AddChild(child *VNode) {
	v.Children = append(v.Children, child)
}

// SetContent sets the content for text nodes.
func (v *VNode) SetContent(content string) {
	v.Content = content
}

// LuaVNodeToVNode converts a Lua table at stack index to a VNode.
func LuaVNodeToVNode(L *lua.State, idx int) *VNode {
	L.PushValue(idx)
	tableIdx := L.AbsIndex(-1)
	defer L.Pop(1)

	vnode := NewVNode("box")

	// Get type field
	L.GetField(tableIdx, "type")
	if !L.IsNoneOrNil(-1) {
		if t, ok := L.ToString(-1); ok {
			vnode.Type = t
		}
	}
	L.Pop(1)

	// Handle component type: instantiate and render the component
	if vnode.Type == "component" {
		return luaComponentToVNode(L, tableIdx)
	}

	// Handle portal type: store content and target for later rendering
	if vnode.Type == "portal" {
		return luaPortalToVNode(L, tableIdx)
	}

	// Handle profiler type: time children rendering, call onRender callback
	if vnode.Type == "profiler" {
		return luaProfilerToVNode(L, tableIdx)
	}

	// Handle strictmode type: double-render children to detect side effects
	if vnode.Type == "strictmode" {
		return luaStrictModeToVNode(L, tableIdx)
	}

	// Get content field (text nodes)
	L.GetField(tableIdx, "content")
	if !L.IsNoneOrNil(-1) {
		if c, ok := L.ToString(-1); ok {
			vnode.Content = c
		}
	}
	L.Pop(1)

	// Get children field
	L.GetField(tableIdx, "children")
	if L.Type(-1) == lua.TypeTable {
		childrenIdx := L.AbsIndex(-1)
		// Iterate children array using integer keys for stability
		n := int(L.RawLen(childrenIdx))
		for i := 1; i <= n; i++ {
			L.RawGetI(childrenIdx, int64(i))
			if L.Type(-1) == lua.TypeTable {
				child := LuaVNodeToVNode(L, -1)
				vnode.AddChild(child)
			}
			L.Pop(1) // pop child value
		}
	}
	L.Pop(1) // pop children table

	// Copy remaining fields as props
	L.PushNil()
	for L.Next(tableIdx) {
		// stack: [..., tableIdx, ..., key, value]
		if L.Type(-2) == lua.TypeString {
			key, _ := L.ToString(-2)
			// Skip known fields
			if key != "type" && key != "content" && key != "children" && key != "_focused" {
				// Store Lua functions as registry references so event handlers survive
				if L.Type(-1) == lua.TypeFunction {
					vnode.Props[key] = storeLuaFuncRef(L, -1)
				} else {
					vnode.Props[key] = L.ToAny(-1)
				}
			} else if key == "_focused" {
				if L.ToBoolean(-1) {
					vnode.Focused = true
				}
			}
		}
		L.Pop(1) // pop value, keep key for Next iteration
	}

	// Flatten props.id → Props["id"] for backward compatibility.
	// Some Lua code uses { props = { id = "..." } } instead of { id = "..." }.
	if _, hasID := vnode.Props["id"]; !hasID {
		if propsMap, ok := vnode.Props["props"].(map[string]any); ok {
			if id, ok := propsMap["id"].(string); ok && id != "" {
				vnode.Props["id"] = id
			}
		}
	}

	return vnode
}

// luaComponentToVNode handles a VNode table with type="component".
// It extracts _factory and _props, creates/reuses a Component instance,
// calls its render function, and recursively converts the result.
// The Lua table is expected at the given stack index (already pushed by caller).
func luaComponentToVNode(L *lua.State, idx int) *VNode {
	absIdx := L.AbsIndex(idx)

	// Get _factory table
	L.GetField(absIdx, "_factory")
	if L.Type(-1) != lua.TypeTable {
		L.Pop(1)
		return NewVNode("box")
	}
	factoryIdx := L.GetTop()

	// Get factory name
	L.GetField(factoryIdx, "name")
	factoryName, _ := L.ToString(-1)
	L.Pop(1)

	// Check for error boundary flag
	L.GetField(factoryIdx, "_isErrorBoundary")
	isErrorBoundary := L.ToBoolean(-1)
	L.Pop(1)

	// Check for memoized flag
	L.GetField(factoryIdx, "_memoized")
	isMemoized := L.ToBoolean(-1)
	L.Pop(1)

	// Check for forwardRef flag
	L.GetField(factoryIdx, "_forwardRef")
	isForwardRef := L.ToBoolean(-1)
	L.Pop(1)
	_ = isForwardRef // used for tagging only

	// Get _props as Go map for prop comparison/storage.
	// Walk the table manually to preserve Lua function refs (L.ToMap drops them).
	L.GetField(absIdx, "_props")
	props := make(map[string]any)
	if L.Type(-1) == lua.TypeTable {
		propsIdx := L.AbsIndex(-1)
		L.PushNil()
		for L.Next(propsIdx) {
			if L.Type(-2) == lua.TypeString {
				key, _ := L.ToString(-2)
				if L.Type(-1) == lua.TypeFunction {
					// Store Lua function as a registry ref so it survives
					props[key] = storeLuaFuncRef(L, -1)
				} else {
					props[key] = L.ToAny(-1)
				}
			}
			L.Pop(1) // pop value, keep key for next iteration
		}
	}
	L.Pop(1) // pop _props

	// Get optional key from props for reconciliation.
	// Fall back to props["id"] so each createElement with a unique id
	// gets its own Component instance (and thus its own useState state).
	compKey := ""
	if k, ok := props["key"].(string); ok {
		compKey = k
	} else if id, ok := props["id"].(string); ok {
		compKey = id
	} else if k, ok := props["key"].(int64); ok {
		compKey = fmt.Sprintf("%d", k)
	}

	// Look up existing component by key or factory name
	parentComp := GetCurrentComponent()
	var comp *Component
	if parentComp != nil {
		comp = findChildComponent(parentComp, factoryName, compKey)
	}

	if comp != nil {
		// Existing component — update props if changed
		comp.UpdateProps(props)
	} else {
		// New component — create instance
		var err error
		comp, err = NewComponent(L, factoryIdx, props)
		if err != nil {
			L.Pop(1) // pop _factory
			return NewVNode("box")
		}

		// Set special flags
		if isErrorBoundary {
			comp.IsErrorBoundary = true
			// Store fallback function ref
			L.GetField(factoryIdx, "_fallback")
			if L.Type(-1) == lua.TypeFunction {
				refID := L.Ref(lua.RegistryIndex)
				comp.FallbackFn = &luaFunctionRef{RefID: refID}
			} else {
				L.Pop(1)
			}
		}
		if isMemoized {
			comp.Memoized = true
		}

		// Call init function if present
		if comp.PushInitFn(L) {
			L.PushAny(props)
			prevComp := GetCurrentComponent()
			SetCurrentComponent(comp)
			status := L.PCall(1, lua.MultiRet, 0)
			if status == lua.OK && L.GetTop() > factoryIdx {
				if initState, ok := L.ToMap(-1); ok {
					for k, v := range initState {
						comp.State[k] = v
					}
				}
				L.Pop(1)
			} else if status != lua.OK {
				L.Pop(1) // pop error
			}
			SetCurrentComponent(prevComp)
		}

		// Register in parent's child list
		if parentComp != nil {
			parentComp.AddChild(comp)
		}
	}

	L.Pop(1) // pop _factory

	// React.memo: skip render if props unchanged
	if comp.Memoized && comp.LastVNode != nil && comp.LastProps != nil {
		if propsEqual(comp.LastProps, props) {
			return comp.LastVNode
		}
	}

	// Render the component
	prevComp := GetCurrentComponent()
	SetCurrentComponent(comp)
	comp.ResetHookIndex()

	if !comp.PushRenderFn(L) {
		SetCurrentComponent(prevComp)
		return NewVNode("box")
	}
	// DEBUG: log what function we're about to call
	fmt.Fprintf(os.Stderr, "[RENDER-CALL] comp=%s factory=%s renderRef=%d fnType=%d\n",
		comp.ID, comp.Name, comp.RenderRefID(), L.Type(-1))

	// Build instance table: state + props
	fields := map[string]any{
		"_instance": comp.ID,
	}
	comp.mu.RLock()
	for k, v := range comp.State {
		fields[k] = v
	}
	comp.mu.RUnlock()
	fields["props"] = props

	L.NewTableFrom(fields)
	for k, v := range props {
		if k != "children" {
			if ref, ok := v.(LuaFuncRef); ok {
				// Push Lua function from registry ref
				L.RawGetI(lua.RegistryIndex, int64(ref.Ref))
			} else {
				L.PushAny(v)
			}
			L.SetField(-2, k)
		}
	}
	// Copy children directly from the original Lua _props table (preserves Lua table refs
	// like createElement results that contain _factory with function references)
	instanceIdx := L.AbsIndex(-1)
	L.GetField(absIdx, "_props")
	if L.Type(-1) == lua.TypeTable {
		L.GetField(-1, "children")
		if !L.IsNoneOrNil(-1) {
			L.SetField(instanceIdx, "children")
		} else {
			L.Pop(1) // pop nil children
		}
		L.Pop(1) // pop _props
	} else {
		L.Pop(1) // pop non-table _props
	}

	status := L.PCall(1, 1, 0)
	// NOTE: Don't restore prevComp yet — child components in render output
	// need to see 'comp' as their parent for error boundary + tree tracking.

	if status != lua.OK {
		// Render error — try to find an error boundary
		errMsg, _ := L.ToString(-1)
		L.Pop(1)
		SetCurrentComponent(prevComp)

		boundary := findErrorBoundary(comp)
		if boundary != nil {
			boundary.CaughtError = errMsg
			return renderErrorFallback(L, boundary, errMsg)
		}
		// No boundary — return error text node
		errNode := NewVNode("text")
		errNode.Content = "Render error: " + errMsg
		return errNode
	}

	// Recursively convert the rendered VNode (comp is still current component,
	// so child components will be registered as children of comp)
	var resultVNode *VNode
	if L.Type(-1) == lua.TypeTable {
		resultVNode = LuaVNodeToVNode(L, -1)
	} else {
		resultVNode = NewVNode("box")
	}
	L.Pop(1) // pop render result
	SetCurrentComponent(prevComp)

	// Tag the VNode with the component reference
	resultVNode.ComponentRef = comp
	resultVNode.ComponentKey = compKey

	// Store for diffing and memo
	comp.LastVNode = resultVNode
	if comp.Memoized {
		comp.LastProps = copyProps(props)
	}

	return resultVNode
}

// findErrorBoundary walks up the component tree to find the nearest error boundary.
func findErrorBoundary(comp *Component) *Component {
	for c := comp.Parent; c != nil; c = c.Parent {
		if c.IsErrorBoundary {
			return c
		}
	}
	return nil
}

// renderErrorFallback calls an error boundary's fallback function.
func renderErrorFallback(L *lua.State, boundary *Component, errMsg string) *VNode {
	if boundary.FallbackFn == nil {
		node := NewVNode("text")
		node.Content = "Error: " + errMsg
		return node
	}

	L.RawGetI(lua.RegistryIndex, int64(boundary.FallbackFn.RefID))
	if L.Type(-1) != lua.TypeFunction {
		L.Pop(1)
		node := NewVNode("text")
		node.Content = "Error: " + errMsg
		return node
	}

	L.PushString(errMsg)
	status := L.PCall(1, 1, 0)
	if status != lua.OK {
		L.Pop(1)
		node := NewVNode("text")
		node.Content = "Error: " + errMsg
		return node
	}

	var resultVNode *VNode
	if L.Type(-1) == lua.TypeTable {
		resultVNode = LuaVNodeToVNode(L, -1)
	} else {
		resultVNode = NewVNode("text")
		resultVNode.Content = "Error: " + errMsg
	}
	L.Pop(1)

	resultVNode.ComponentRef = boundary
	return resultVNode
}

// copyProps creates a shallow copy of a props map.
func copyProps(props map[string]any) map[string]any {
	if props == nil {
		return nil
	}
	cp := make(map[string]any, len(props))
	for k, v := range props {
		cp[k] = v
	}
	return cp
}

// findChildComponent looks for an existing child component by factory name and key.
func findChildComponent(parent *Component, factoryName, key string) *Component {
	parent.mu.RLock()
	defer parent.mu.RUnlock()
	for _, child := range parent.ChildComps {
		if child.Type == factoryName {
			if key == "" {
				return child
			}
			// Match by explicit "key" prop, or fall back to "id" prop
			childKey, _ := child.Props["key"].(string)
			if childKey == "" {
				childKey, _ = child.Props["id"].(string)
			}
			if childKey == key {
				return child
			}
		}
	}
	return nil
}



// luaPortalToVNode handles a VNode table with type="portal".
// It extracts _content and _target, converts content to VNode, and marks it as portal.
func luaPortalToVNode(L *lua.State, tableIdx int) *VNode {
	absIdx := L.AbsIndex(tableIdx)

	// Get target ID
	L.GetField(absIdx, "_target")
	targetID, _ := L.ToString(-1)
	L.Pop(1)

	// Get content VNode
	L.GetField(absIdx, "_content")
	var contentVNode *VNode
	if L.Type(-1) == lua.TypeTable {
		contentVNode = LuaVNodeToVNode(L, -1)
	} else {
		contentVNode = NewVNode("box")
	}
	L.Pop(1)

	// Mark as portal
	contentVNode.IsPortal = true
	contentVNode.PortalTarget = targetID

	return contentVNode
}


// VNodeToFrame converts a VNode tree to a Frame with layout.
func VNodeToFrame(vnode *VNode, width, height int) *Frame {
	frame := NewFrame(width, height)

	// Compute layout using flexbox engine
	computeFlexLayout(vnode, 0, 0, width, height)

	// Render VNode to frame
	fullClip := Rect{X: 0, Y: 0, W: width, H: height}
	renderVNode(frame, vnode, fullClip)

	// Mark entire frame dirty
	frame.MarkDirty()

	return frame
}

// VNodeToFrameWithFocus converts a VNode tree to a Frame with focus indication.
func VNodeToFrameWithFocus(vnode *VNode, width, height int, focusedID string) *Frame {
	frame := NewFrame(width, height)
	frame.FocusedID = focusedID

	// Compute layout using flexbox engine
	computeFlexLayout(vnode, 0, 0, width, height)

	// Mark focused VNode
	markFocusedVNode(vnode, focusedID)

	// Render VNode to frame
	fullClip := Rect{X: 0, Y: 0, W: width, H: height}
	renderVNode(frame, vnode, fullClip)

	// Mark entire frame dirty
	frame.MarkDirty()

	return frame
}


// VNodeToFrameWithOverlays renders the base VNode tree, then composites overlay layers on top.
func VNodeToFrameWithOverlays(vnode *VNode, width, height int, overlays []*Overlay) *Frame {
	// 1. Render base layer
	frame := VNodeToFrame(vnode, width, height)

	if len(overlays) == 0 {
		return frame
	}

	// 2. Composite overlays using the layer compositor
	compositor := NewCompositor(width, height)
	return compositor.Compose(frame, overlays)
}

// renderBackdrop dims the entire screen for a modal overlay.
func renderBackdrop(frame *Frame) {
	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			frame.Cells[y][x].Dim = true
			frame.Cells[y][x].Transparent = false
		}
	}
}

// markFocusedVNode recursively marks the VNode that matches the focusedID.
func markFocusedVNode(vnode *VNode, focusedID string) {
	if vnode == nil {
		return
	}

	// Check if this node is focused (by ID in props)
	if id, ok := vnode.Props["id"].(string); ok && id == focusedID {
		vnode.Focused = true
	}

	// Recurse into children
	for _, child := range vnode.Children {
		markFocusedVNode(child, focusedID)
	}
}


// getVNodeID returns the id prop of a VNode, or "" if not set.
func getVNodeID(vnode *VNode) string {
	if vnode == nil {
		return ""
	}
	if id, ok := vnode.Props["id"].(string); ok {
		return id
	}
	return ""
}

// renderVNode renders a VNode into the frame, clipped to the given rect.
// All cell writes go through frame.SetCellClipped — content NEVER overflows.
func renderVNode(frame *Frame, vnode *VNode, clip Rect) {
	// Quick reject: if vnode is entirely outside clip
	if vnode.W > 0 && vnode.H > 0 {
		nodeRect := Rect{X: vnode.X, Y: vnode.Y, W: vnode.W, H: vnode.H}
		if IntersectRect(nodeRect, clip).W == 0 || IntersectRect(nodeRect, clip).H == 0 {
			return
		}
	}

	style := vnode.Style

	cell := Cell{Char: ' '}
	cell.Foreground = style.Foreground
	cell.Background = style.Background
	cell.Bold = style.Bold
	cell.Dim = style.Dim
	cell.Underline = style.Underline

	switch vnode.Type {
	case "fragment":
		for _, child := range vnode.Children {
			renderVNode(frame, child, clip)
		}
		return

	case "text":
		renderText(frame, vnode, cell, clip)

	case "input":
		renderInput(frame, vnode, style, clip)

	case "textarea":
		renderTextArea(frame, vnode, style, clip)

	default:
		// Container types: box, vbox, hbox, etc.

		// Render background (clipped to parent clip)
		if style.Background != "" {
			frame.FillClipped(vnode.X, vnode.Y, vnode.W, vnode.H,
				Cell{Char: ' ', Background: style.Background,
					OwnerNode: vnode, OwnerID: getVNodeID(vnode), CellRole: "background"}, clip)
		}

		// Render border (clipped to parent clip)
		if style.Border != "" {
			renderBorder(frame, vnode, style.Border, clip)
		}

		// Render focus indicator (in border area, uses parent clip)
		if vnode.Focused {
			renderFocusIndicator(frame, vnode, clip)
		}

		// Child clip = intersection of parent clip AND this node's inner area
		// (inside border only — the layout engine already positions children
		// with padding offsets, so we don't subtract padding from the clip).
		borderW := 0
		if style.Border != "" {
			borderW = 1
		}
		innerRect := Rect{
			X: vnode.X + borderW,
			Y: vnode.Y + borderW,
			W: vnode.W - 2*borderW,
			H: vnode.H - 2*borderW,
		}
		if innerRect.W < 0 {
			innerRect.W = 0
		}
		if innerRect.H < 0 {
			innerRect.H = 0
		}
		childClip := IntersectRect(clip, innerRect)

		// Render children with child clip.
		// Positioned children (absolute/fixed) render after flow children,
		// sorted by z-index so higher z-index paints on top.
		if style.Overflow == "scroll" {
			renderScrollableChildren(frame, vnode, style, childClip)
		} else {
			var positioned []*VNode
			for _, child := range vnode.Children {
				cs := child.Style
				if cs.Position == "absolute" || cs.Position == "fixed" {
					positioned = append(positioned, child)
				} else {
					renderVNode(frame, child, childClip)
				}
			}
			// Render positioned children sorted by z-index (ascending = lower first)
			if len(positioned) > 0 {
				sortByZIndex(positioned)
				for _, child := range positioned {
					renderVNode(frame, child, childClip)
				}
			}
		}
	}
}

// renderText renders text content into the frame, supporting wrapping and clipping.
func renderText(frame *Frame, vnode *VNode, cell Cell, clip Rect) {
	if vnode.Content == "" {
		return
	}

	// Calculate content area (inside border + padding)
	style := vnode.Style
	borderW := 0
	if style.Border == "single" || style.Border == "double" || style.Border == "rounded" {
		borderW = 1
	}

	startX := vnode.X + borderW + style.PaddingLeft
	startY := vnode.Y + borderW + style.PaddingTop
	availW := vnode.W - 2*borderW - style.PaddingLeft - style.PaddingRight
	if availW <= 0 {
		availW = vnode.W
	}

	col := 0
	row := 0
	for _, ch := range vnode.Content {
		// Handle newlines
		if ch == '\n' {
			col = 0
			row++
			continue
		}

		w := RuneWidth(ch)
		if w == 0 {
			continue
		}

		// Wrap if this character would exceed available width
		if availW > 0 && col+w > availW {
			col = 0
			row++
		}

		px := startX + col
		py := startY + row
		frame.SetCellClipped(px, py, Cell{
			Char:       ch,
			Foreground: cell.Foreground,
			Background: cell.Background,
			Bold:       cell.Bold,
			Dim:        cell.Dim,
			Underline:  cell.Underline,
			OwnerNode:  vnode, OwnerID: getVNodeID(vnode), CellRole: "content",
		}, clip)
		// For wide chars, place a padding cell in the next column
		if w == 2 {
			frame.SetCellClipped(px+1, py, Cell{
				Char:       0,
				Foreground: cell.Foreground,
				Background: cell.Background,
				OwnerNode:  vnode, OwnerID: getVNodeID(vnode), CellRole: "padding",
			}, clip)
		}
		col += w
	}
}

// renderFocusIndicator draws a focus indicator around a VNode.
// Uses reverse video (swap foreground/background) for the border.
func renderFocusIndicator(frame *Frame, vnode *VNode, clip Rect) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H

	if w < 2 || h < 2 {
		return
	}

	style := vnode.Style

	// Use focus-specific border chars if available, else defaults
	focusTl := '['
	focusTr := ']'
	focusBl := '['
	focusBr := ']'
	focusH := '-'
	focusV := '|'

	// If the node has a border, use matching border chars for focus indicator
	if style.FocusBorder != "" || style.Border != "" {
		bt := style.FocusBorder
		if bt == "" {
			bt = style.Border
		}
		switch bt {
		case "single":
			focusTl, focusTr, focusBl, focusBr, focusH, focusV = '┌', '┐', '└', '┘', '─', '│'
		case "double":
			focusTl, focusTr, focusBl, focusBr, focusH, focusV = '╔', '╗', '╚', '╝', '═', '║'
		case "rounded":
			focusTl, focusTr, focusBl, focusBr, focusH, focusV = '╭', '╮', '╰', '╯', '─', '│'
		}
	}

	// Focus color: prefer FocusForeground, then Foreground, then fallback
	fc := style.FocusForeground
	if fc == "" {
		fc = style.Foreground
		if fc == "" {
			fc = "#FFFF00" // last resort fallback
		}
	}
	oid := getVNodeID(vnode)

	bc := func(ch rune) Cell {
		return Cell{Char: ch, Foreground: fc, OwnerNode: vnode, OwnerID: oid, CellRole: "border"}
	}

	// Top border
	frame.SetCellClipped(x, y, bc(focusTl), clip)
	for i := 1; i < w-1; i++ {
		frame.SetCellClipped(x+i, y, bc(focusH), clip)
	}
	frame.SetCellClipped(x+w-1, y, bc(focusTr), clip)

	// Side borders
	for i := 1; i < h-1; i++ {
		frame.SetCellClipped(x, y+i, bc(focusV), clip)
		frame.SetCellClipped(x+w-1, y+i, bc(focusV), clip)
	}

	// Bottom border
	frame.SetCellClipped(x, y+h-1, bc(focusBl), clip)
	for i := 1; i < w-1; i++ {
		frame.SetCellClipped(x+i, y+h-1, bc(focusH), clip)
	}
	frame.SetCellClipped(x+w-1, y+h-1, bc(focusBr), clip)
}

// renderBorder draws a border around a VNode.
func renderBorder(frame *Frame, vnode *VNode, borderType string, clip Rect) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H

	if w < 2 || h < 2 {
		return
	}

	var tl, tr, bl, br, hLine, vLine rune
	style := vnode.Style

	// Check for custom border chars first
	if style.BorderChars != nil {
		tl = firstRune(style.BorderChars["tl"], '┌')
		tr = firstRune(style.BorderChars["tr"], '┐')
		bl = firstRune(style.BorderChars["bl"], '└')
		br = firstRune(style.BorderChars["br"], '┘')
		hLine = firstRune(style.BorderChars["h"], '─')
		vLine = firstRune(style.BorderChars["v"], '│')
	} else {
		switch borderType {
		case "single":
			tl, tr, bl, br, hLine, vLine = '┌', '┐', '└', '┘', '─', '│'
		case "double":
			tl, tr, bl, br, hLine, vLine = '╔', '╗', '╚', '╝', '═', '║'
		case "rounded":
			tl, tr, bl, br, hLine, vLine = '╭', '╮', '╰', '╯', '─', '│'
		default:
			return
		}
	}

	// Border foreground color from style
	borderFg := style.Foreground // borders inherit foreground color
	oid := getVNodeID(vnode)
	bc := func(ch rune) Cell {
		return Cell{Char: ch, Foreground: borderFg, OwnerNode: vnode, OwnerID: oid, CellRole: "border"}
	}

	// Top border
	frame.SetCellClipped(x, y, bc(tl), clip)
	for i := 1; i < w-1; i++ {
		frame.SetCellClipped(x+i, y, bc(hLine), clip)
	}
	frame.SetCellClipped(x+w-1, y, bc(tr), clip)

	// Bottom border
	frame.SetCellClipped(x, y+h-1, bc(bl), clip)
	for i := 1; i < w-1; i++ {
		frame.SetCellClipped(x+i, y+h-1, bc(hLine), clip)
	}
	frame.SetCellClipped(x+w-1, y+h-1, bc(br), clip)

	// Vertical borders
	for i := 1; i < h-1; i++ {
		frame.SetCellClipped(x, y+i, bc(vLine), clip)
		frame.SetCellClipped(x+w-1, y+i, bc(vLine), clip)
	}
}

// RenderLuaVNode converts a Lua VDOM table to a Frame and writes it.
func RenderLuaVNode(L *lua.State, vnodeIdx int, width, height int) *Frame {
	vnode := LuaVNodeToVNode(L, vnodeIdx)
	return VNodeToFrame(vnode, width, height)
}

// RenderToTerminal renders a VDOM to the terminal using the global adapter.
func RenderToTerminal(L *lua.State, vnodeIdx int) error {
	frame := RenderLuaVNode(L, vnodeIdx, 80, 24)

	adapter := GetOutputAdapter()
	if adapter == nil {
		adapter = NewANSIAdapter(new(noopWriter))
	}

	return adapter.Write(frame)
}

// noopWriter is a writer that discards output.
type noopWriter struct{}

func (n *noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

// SetTerminalTitle sets the terminal window title.
func SetTerminalTitle(title string) error {
	// OSC sequence to set title: \x1b]0;title\x07
	// This requires using the raw writer, not the buffered adapter
	return nil // TODO: implement if needed
}

// renderScrollableChildren renders children of a scrollable container,
// clipping them to the container's visible area and rendering a scrollbar.
func renderScrollableChildren(frame *Frame, vnode *VNode, style Style, clip Rect) {
	if clip.W <= 0 || clip.H <= 0 {
		return
	}

	// Reserve 1 column for scrollbar
	contentClip := clip
	if contentClip.W > 1 {
		contentClip.W--
	}

	// Render children with clipping
	for _, child := range vnode.Children {
		renderVNode(frame, child, contentClip)
	}

	// Render scrollbar if needed
	nodeID, _ := vnode.Props["id"].(string)
	if nodeID != "" {
		vp := GetViewport(nodeID)
		if vp.NeedsScroll() {
			scrollbarX := clip.X + contentClip.W
			renderScrollbar(frame, scrollbarX, clip.Y, clip.H, vp, style, clip)
		}
	}
}
// renderScrollbar draws a vertical scrollbar at the given position.
// Reads styling from the parent VNode's style if available.
func renderScrollbar(frame *Frame, x, y, trackH int, vp *Viewport, style Style, clip Rect) {
	if trackH <= 0 {
		return
	}

	thumbStart, thumbSize := vp.ScrollbarThumb(trackH)

	// Use style-specified chars/colors, or defaults
	trackChar := firstRune(style.ScrollbarTrackChar, '│')
	thumbChar := firstRune(style.ScrollbarThumbChar, '█')
	fg := style.ScrollbarForeground // empty = terminal default

	for i := 0; i < trackH; i++ {
		py := y + i
		var ch rune
		if i >= thumbStart && i < thumbStart+thumbSize {
			ch = thumbChar
		} else {
			ch = trackChar
		}
		frame.SetCellClipped(x, py, Cell{
			Char:       ch,
			Foreground: fg,
		}, clip)
	}
}

// firstRune returns the first rune of s, or the fallback if s is empty.
func firstRune(s string, fallback rune) rune {
	if s == "" {
		return fallback
	}
	for _, r := range s {
		return r
	}
	return fallback
}

// renderInput renders a single-line text input VNode.
func renderInput(frame *Frame, vnode *VNode, style Style, clip Rect) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H
	if w <= 0 || h <= 0 {
		return
	}

	// Get or create text input state
	nodeID, _ := vnode.Props["id"].(string)
	var state *TextInputState
	if nodeID != "" {
		state = GetTextInput(nodeID)
		state.Width = w
		state.Height = 1
		state.MultiLine = false
		state.Focused = vnode.Focused
		if val, ok := vnode.Props["value"].(string); ok {
			if val != state.Text {
				state.Text = val
				state.EnsureCursorVisible()
			}
		}
		if ph, ok := vnode.Props["placeholder"].(string); ok {
			state.Placeholder = ph
		}
		if ml, ok := vnode.Props["maxLength"]; ok {
			state.MaxLength = getInt(vnode.Props, "maxLength")
			_ = ml
		}
		if ro := getBool(vnode.Props, "readOnly"); ro {
			state.ReadOnly = true
		}
	}

	// Determine what to render
	text := ""
	isPlaceholder := false
	scrollX := 0

	if state != nil {
		if state.Text == "" && state.Placeholder != "" {
			text = state.Placeholder
			isPlaceholder = true
		} else {
			text = state.Text
			scrollX = state.ScrollX
		}
	} else if val, ok := vnode.Props["value"].(string); ok {
		text = val
	}

	// Render background
	bg := style.Background // empty = transparent (terminal default)
	oid := getVNodeID(vnode)
	for cx := x; cx < x+w; cx++ {
		frame.SetCellClipped(cx, y, Cell{Char: ' ', Background: bg,
			OwnerNode: vnode, OwnerID: oid, CellRole: "background"}, clip)
	}

	// Render border first (before content, so content draws inside the border)
	if style.Border != "" {
		renderBorder(frame, vnode, style.Border, clip)
	}

	// Render text (with horizontal scroll)
	runes := []rune(text)
	fg := style.Foreground // empty = terminal default
	if isPlaceholder {
		if style.PlaceholderColor != "" {
			fg = style.PlaceholderColor
		} else if fg == "" {
			fg = "#888888" // last resort fallback for placeholder
		}
	}

	for col := 0; col < w; col++ {
		runeIdx := col + scrollX
		if runeIdx >= len(runes) {
			break
		}
		px := x + col

		cellFg := fg
		cellBg := bg
		reverse := false

		if state != nil && state.HasSelection() {
			lo := state.selMin()
			hi := state.selMax()
			if runeIdx >= lo && runeIdx < hi {
				reverse = true
			}
		}

		frame.SetCellClipped(px, y, Cell{
			Char:       runes[runeIdx],
			Foreground: cellFg,
			Background: cellBg,
			Bold:       style.Bold,
			Reverse:    reverse,
			OwnerNode:  vnode, OwnerID: oid, CellRole: "content",
		}, clip)
	}

	// Render cursor
	if state != nil && state.Focused {
		cursorCol := state.CursorPos - scrollX
		if cursorCol >= 0 && cursorCol < w {
			px := x + cursorCol
			ch := ' '
			if state.CursorPos < len(runes) {
				ch = runes[state.CursorPos]
			}
			frame.SetCellClipped(px, y, Cell{
				Char:       ch,
				Foreground: bg,
				Background: fg,
				Reverse:    true,
				OwnerNode:  vnode, OwnerID: oid, CellRole: "content",
			}, clip)
		}
	}

}

// renderTextArea renders a multi-line text area VNode.
func renderTextArea(frame *Frame, vnode *VNode, style Style, clip Rect) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H
	if w <= 0 || h <= 0 {
		return
	}

	// Account for border
	borderW := 0
	if style.Border == "single" || style.Border == "double" || style.Border == "rounded" {
		borderW = 1
	}

	contentX := x + borderW + style.PaddingLeft
	contentY := y + borderW + style.PaddingTop
	contentW := w - 2*borderW - style.PaddingLeft - style.PaddingRight
	contentH := h - 2*borderW - style.PaddingTop - style.PaddingBottom
	if contentW <= 0 || contentH <= 0 {
		return
	}

	// Get or create text input state
	nodeID, _ := vnode.Props["id"].(string)
	var state *TextInputState
	if nodeID != "" {
		state = GetTextInput(nodeID)
		state.Width = contentW
		state.Height = contentH
		state.MultiLine = true
		state.Focused = vnode.Focused
		if val, ok := vnode.Props["value"].(string); ok {
			if val != state.Text {
				state.Text = val
				state.EnsureCursorVisible()
			}
		}
		if ph, ok := vnode.Props["placeholder"].(string); ok {
			state.Placeholder = ph
		}
		if ml, ok := vnode.Props["maxLength"]; ok {
			state.MaxLength = getInt(vnode.Props, "maxLength")
			_ = ml
		}
		if ro := getBool(vnode.Props, "readOnly"); ro {
			state.ReadOnly = true
		}
	}

	// Render background (clipped)
	bg := style.Background // empty = transparent (terminal default)
	oid := getVNodeID(vnode)
	frame.FillClipped(x, y, w, h, Cell{Char: ' ', Background: bg,
		OwnerNode: vnode, OwnerID: oid, CellRole: "background"}, clip)

	// Render border (clipped)
	if style.Border != "" {
		renderBorder(frame, vnode, style.Border, clip)
	}

	// Content clip = intersection of parent clip and content area
	contentClip := IntersectRect(clip, Rect{X: contentX, Y: contentY, W: contentW, H: contentH})

	// Determine text and scroll
	text := ""
	isPlaceholder := false
	scrollX := 0
	scrollY := 0

	if state != nil {
		if state.Text == "" && state.Placeholder != "" {
			text = state.Placeholder
			isPlaceholder = true
		} else {
			text = state.Text
			scrollX = state.ScrollX
			scrollY = state.ScrollY
		}
	} else if val, ok := vnode.Props["value"].(string); ok {
		text = val
	}

	// Split into lines
	lines := splitLines(text)

	fg := style.Foreground // empty = terminal default
	if isPlaceholder {
		if style.PlaceholderColor != "" {
			fg = style.PlaceholderColor
		} else if fg == "" {
			fg = "#888888" // last resort fallback for placeholder
		}
	}

	// Render visible lines
	for row := 0; row < contentH; row++ {
		lineIdx := row + scrollY
		if lineIdx >= len(lines) {
			break
		}
		line := []rune(lines[lineIdx])

		py := contentY + row

		for col := 0; col < contentW; col++ {
			runeIdx := col + scrollX
			if runeIdx >= len(line) {
				break
			}
			px := contentX + col

			cellFg := fg
			reverse := false

			if state != nil && state.HasSelection() && !isPlaceholder {
				absPos := lineColToPos(lines, lineIdx, runeIdx)
				lo := state.selMin()
				hi := state.selMax()
				if absPos >= lo && absPos < hi {
					reverse = true
				}
			}

			frame.SetCellClipped(px, py, Cell{
				Char:       line[runeIdx],
				Foreground: cellFg,
				Background: bg,
				Bold:       style.Bold,
				Reverse:    reverse,
				OwnerNode:  vnode, OwnerID: oid, CellRole: "content",
			}, contentClip)
		}
	}

	// Render cursor
	if state != nil && state.Focused && !isPlaceholder {
		runes := []rune(state.Text)
		curLine, curCol := state.lineCol(runes)
		screenRow := curLine - scrollY
		screenCol := curCol - scrollX

		if screenRow >= 0 && screenRow < contentH && screenCol >= 0 && screenCol < contentW {
			px := contentX + screenCol
			py := contentY + screenRow
			ch := ' '
			if state.CursorPos < len(runes) {
				ch = runes[state.CursorPos]
			}
			frame.SetCellClipped(px, py, Cell{
				Char:       ch,
				Foreground: bg,
				Background: fg,
				Reverse:    true,
				OwnerNode:  vnode, OwnerID: oid, CellRole: "content",
			}, contentClip)
		}
	}
}

// splitLines splits text into lines (preserving empty lines).
func splitLines(text string) []string {
	if text == "" {
		return []string{""}
	}
	lines := make([]string, 0)
	start := 0
	for i, ch := range text {
		if ch == '\n' {
			lines = append(lines, text[start:i])
			start = i + 1
		}
	}
	lines = append(lines, text[start:])
	return lines
}

// lineColToPos converts a (line, col) position to an absolute rune position.
func lineColToPos(lines []string, line, col int) int {
	pos := 0
	for i := 0; i < line && i < len(lines); i++ {
		pos += len([]rune(lines[i])) + 1 // +1 for newline
	}
	pos += col
	return pos
}

// luaProfilerToVNode handles a VNode table with type="profiler".
// It wraps child rendering with timing and calls the onRender callback.
func luaProfilerToVNode(L *lua.State, idx int) *VNode {
	vnode := NewVNode("fragment") // profiler is transparent like fragment

	// Get profiler ID
	L.GetField(idx, "id")
	profilerID := "unknown"
	if s, ok := L.ToString(-1); ok && s != "" {
		profilerID = s
	}
	L.Pop(1)

	// Get onRender callback reference
	L.GetField(idx, "onRender")
	hasCallback := L.IsFunction(-1)
	var callbackIdx int
	if hasCallback {
		callbackIdx = L.AbsIndex(-1)
	}

	// Time the children rendering
	startTime := float64(time.Now().UnixMilli())

	// Render children
	L.GetField(idx, "children")
	if L.Type(-1) == lua.TypeTable {
		childrenIdx := L.AbsIndex(-1)
		n := int(L.RawLen(childrenIdx))
		for i := 1; i <= n; i++ {
			L.RawGetI(childrenIdx, int64(i))
			if L.Type(-1) == lua.TypeTable {
				child := LuaVNodeToVNode(L, -1)
				vnode.AddChild(child)
			}
			L.Pop(1)
		}
	}
	L.Pop(1) // pop children

	commitTime := float64(time.Now().UnixMilli())
	actualDuration := commitTime - startTime

	// Call onRender(id, phase, actualDuration, baseDuration, startTime, commitTime)
	if hasCallback {
		L.PushValue(callbackIdx)
		L.PushString(profilerID)
		L.PushString("mount") // phase: "mount" for first render, "update" for re-render
		L.PushNumber(actualDuration)
		L.PushNumber(actualDuration) // baseDuration ≈ actualDuration for now
		L.PushNumber(startTime)
		L.PushNumber(commitTime)
		if err := L.PCall(6, 0, 0); err != 0 {
			errMsg, _ := L.ToString(-1)
			L.Pop(1) // pop error message
			fmt.Printf("lumina: profiler callback error: %s\n", errMsg)
		}
	}

	L.Pop(1) // pop onRender

	vnode.Props["profiler_id"] = profilerID
	return vnode
}

// luaStrictModeToVNode handles a VNode table with type="strictmode".
// It double-renders children to detect side effects.
func luaStrictModeToVNode(L *lua.State, idx int) *VNode {
	vnode := NewVNode("fragment") // strictmode is transparent like fragment

	// Enable strict mode
	prevStrict := IsStrictMode()
	SetStrictMode(true)
	defer SetStrictMode(prevStrict)

	// First render of children
	L.GetField(idx, "children")
	if L.Type(-1) == lua.TypeTable {
		childrenIdx := L.AbsIndex(-1)
		n := int(L.RawLen(childrenIdx))
		for i := 1; i <= n; i++ {
			L.RawGetI(childrenIdx, int64(i))
			if L.Type(-1) == lua.TypeTable {
				child := LuaVNodeToVNode(L, -1)
				vnode.AddChild(child)
			}
			L.Pop(1)
		}
	}
	L.Pop(1) // pop children

	// Second render (strict mode double-render) — results are discarded
	// This helps detect side effects in render functions
	L.GetField(idx, "children")
	if L.Type(-1) == lua.TypeTable {
		childrenIdx := L.AbsIndex(-1)
		n := int(L.RawLen(childrenIdx))
		for i := 1; i <= n; i++ {
			L.RawGetI(childrenIdx, int64(i))
			if L.Type(-1) == lua.TypeTable {
				_ = LuaVNodeToVNode(L, -1) // discard second render result
			}
			L.Pop(1)
		}
	}
	L.Pop(1)

	return vnode
}
