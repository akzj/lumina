// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

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
				vnode.Props[key] = L.ToAny(-1)
			} else if key == "_focused" {
				if L.ToBoolean(-1) {
					vnode.Focused = true
				}
			}
		}
		L.Pop(1) // pop value, keep key for Next iteration
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

	// Get _props as Go map for prop comparison/storage
	L.GetField(absIdx, "_props")
	var props map[string]any
	if L.Type(-1) == lua.TypeTable {
		if m, ok := L.ToMap(-1); ok {
			props = m
		}
	}
	if props == nil {
		props = make(map[string]any)
	}
	L.Pop(1) // pop _props

	// Get optional key from props for reconciliation
	compKey := ""
	if k, ok := props["key"].(string); ok {
		compKey = k
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
			L.PushAny(v)
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
			if key == "" || child.Props["key"] == key {
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
	renderVNode(frame, vnode)

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
	renderVNode(frame, vnode)

	// Mark entire frame dirty
	frame.MarkDirty()

	return frame
}

// VNodeToFrameWithOverlays renders the base VNode tree, then renders overlay layers on top.
func VNodeToFrameWithOverlays(vnode *VNode, width, height int, overlays []*Overlay) *Frame {
	// 1. Render base layer
	frame := VNodeToFrame(vnode, width, height)

	if len(overlays) == 0 {
		return frame
	}

	// 2. Render each overlay on top (overlays should already be sorted by ZIndex)
	for _, ov := range overlays {
		if !ov.Visible {
			continue
		}

		// If modal, render semi-transparent backdrop first
		if ov.Modal {
			renderBackdrop(frame)
		}

		// Layout the overlay VNode at its absolute position
		if ov.VNode != nil {
			computeFlexLayout(ov.VNode, ov.X, ov.Y, ov.W, ov.H)
			renderVNode(frame, ov.VNode)
		}
	}

	return frame
}

// renderBackdrop dims the entire screen for a modal overlay.
func renderBackdrop(frame *Frame) {
	for y := 0; y < frame.Height; y++ {
		for x := 0; x < frame.Width; x++ {
			frame.Cells[y][x].Dim = true
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


// renderVNode renders a VNode into the frame.
func renderVNode(frame *Frame, vnode *VNode) {
	// Use parsed style (populated by computeFlexLayout)
	style := vnode.Style

	cell := Cell{Char: ' '}
	cell.Foreground = style.Foreground
	cell.Background = style.Background
	cell.Bold = style.Bold
	cell.Dim = style.Dim
	cell.Underline = style.Underline

	switch vnode.Type {
	case "fragment":
		// Fragment: render children directly, no background/border/styling
		for _, child := range vnode.Children {
			renderVNode(frame, child)
		}
		return

	case "text":
		// Render text content with wrapping support
		renderText(frame, vnode, cell)

	case "input":
		renderInput(frame, vnode, style)

	case "textarea":
		renderTextArea(frame, vnode, style)

	default:
		// Container types: box, vbox, hbox, container, etc.

		// Render background if specified
		if style.Background != "" {
			for y := vnode.Y; y < vnode.Y+vnode.H && y < frame.Height; y++ {
				for x := vnode.X; x < vnode.X+vnode.W && x < frame.Width; x++ {
					frame.Cells[y][x] = Cell{Char: ' ', Background: style.Background}
				}
			}
		}

		// Render border if specified
		if style.Border != "" {
			renderBorder(frame, vnode, style.Border)
		}

		// Render focus indicator if focused
		if vnode.Focused {
			renderFocusIndicator(frame, vnode)
		}

		// Determine content clipping area for scrollable containers
		scrollable := style.Overflow == "scroll"
		if scrollable {
			renderScrollableChildren(frame, vnode, style)
		} else {
			// Render children normally
			for _, child := range vnode.Children {
				renderVNode(frame, child)
			}
		}
	}
}

// renderText renders text content into the frame, supporting wrapping.
func renderText(frame *Frame, vnode *VNode, cell Cell) {
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
		px := startX + col
		py := startY + row
		if px < frame.Width && py < frame.Height && px >= 0 && py >= 0 {
			frame.Cells[py][px] = Cell{
				Char:       ch,
				Foreground: cell.Foreground,
				Background: cell.Background,
				Bold:       cell.Bold,
				Dim:        cell.Dim,
				Underline:  cell.Underline,
			}
		}
		col++
		if col >= availW {
			col = 0
			row++
		}
	}
}

// renderFocusIndicator draws a focus indicator around a VNode.
// Uses reverse video (swap foreground/background) for the border.
func renderFocusIndicator(frame *Frame, vnode *VNode) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H

	if w < 2 || h < 2 {
		return
	}

	// Focus border characters (using bracket style for clarity)
	focusTl := '['
	focusTr := ']'
	focusBl := '['
	focusBr := ']'
	focusH := '-'
	focusV := '|'

	// Top-left corner
	if y < frame.Height && x < frame.Width {
		frame.Cells[y][x] = Cell{Char: focusTl, Foreground: "#FFFF00"} // Yellow
	}
	// Top border
	for i := 1; i < w-1 && x+i < frame.Width; i++ {
		if y < frame.Height {
			frame.Cells[y][x+i] = Cell{Char: focusH, Foreground: "#FFFF00"}
		}
	}
	// Top-right corner
	if y < frame.Height && x+w-1 < frame.Width {
		frame.Cells[y][x+w-1] = Cell{Char: focusTr, Foreground: "#FFFF00"}
	}

	// Side borders
	for i := 1; i < h-1 && y+i < frame.Height; i++ {
		if x < frame.Width {
			frame.Cells[y+i][x] = Cell{Char: focusV, Foreground: "#FFFF00"}
		}
		if x+w-1 < frame.Width {
			frame.Cells[y+i][x+w-1] = Cell{Char: focusV, Foreground: "#FFFF00"}
		}
	}

	// Bottom-left corner
	if y+h-1 < frame.Height && x < frame.Width {
		frame.Cells[y+h-1][x] = Cell{Char: focusBl, Foreground: "#FFFF00"}
	}
	// Bottom border
	for i := 1; i < w-1 && x+i < frame.Width; i++ {
		if y+h-1 < frame.Height {
			frame.Cells[y+h-1][x+i] = Cell{Char: focusH, Foreground: "#FFFF00"}
		}
	}
	// Bottom-right corner
	if y+h-1 < frame.Height && x+w-1 < frame.Width {
		frame.Cells[y+h-1][x+w-1] = Cell{Char: focusBr, Foreground: "#FFFF00"}
	}
}

// renderBorder draws a border around a VNode.
func renderBorder(frame *Frame, vnode *VNode, borderType string) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H

	if w < 2 || h < 2 {
		return
	}

	// Box drawing characters
	var tl, tr, bl, br, hLine, vLine rune
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

	// Top border
	if y < frame.Height && x < frame.Width {
		frame.Cells[y][x] = Cell{Char: tl}
	}
	for i := 1; i < w-1 && x+i < frame.Width; i++ {
		frame.Cells[y][x+i] = Cell{Char: hLine}
	}
	if y < frame.Height && x+w-1 < frame.Width {
		frame.Cells[y][x+w-1] = Cell{Char: tr}
	}

	// Bottom border
	if y+h-1 < frame.Height && x < frame.Width {
		frame.Cells[y+h-1][x] = Cell{Char: bl}
	}
	for i := 1; i < w-1 && x+i < frame.Width; i++ {
		if y+h-1 < frame.Height {
			frame.Cells[y+h-1][x+i] = Cell{Char: hLine}
		}
	}
	if y+h-1 < frame.Height && x+w-1 < frame.Width {
		frame.Cells[y+h-1][x+w-1] = Cell{Char: br}
	}

	// Vertical borders
	for i := 1; i < h-1 && y+i < frame.Height; i++ {
		if x < frame.Width {
			frame.Cells[y+i][x] = Cell{Char: vLine}
		}
		if x+w-1 < frame.Width {
			frame.Cells[y+i][x+w-1] = Cell{Char: vLine}
		}
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
func renderScrollableChildren(frame *Frame, vnode *VNode, style Style) {
	// Calculate the clip region (content area inside border + padding)
	borderW := 0
	if style.Border == "single" || style.Border == "double" || style.Border == "rounded" {
		borderW = 1
	}

	clipX := vnode.X + borderW + style.PaddingLeft
	clipY := vnode.Y + borderW + style.PaddingTop
	clipW := vnode.W - 2*borderW - style.PaddingLeft - style.PaddingRight
	clipH := vnode.H - 2*borderW - style.PaddingTop - style.PaddingBottom

	// Reserve 1 column for scrollbar
	if clipW > 1 {
		clipW-- // scrollbar takes rightmost column
	}

	if clipW <= 0 || clipH <= 0 {
		return
	}

	// Render children with clipping
	for _, child := range vnode.Children {
		renderVNodeClipped(frame, child, clipX, clipY, clipW, clipH)
	}

	// Render scrollbar if needed
	nodeID, _ := vnode.Props["id"].(string)
	if nodeID != "" {
		vp := GetViewport(nodeID)
		if vp.NeedsScroll() {
			scrollbarX := clipX + clipW // right edge of content area
			renderScrollbar(frame, scrollbarX, clipY, clipH, vp)
		}
	}
}

// renderVNodeClipped renders a VNode and its children, clipping to the given rect.
func renderVNodeClipped(frame *Frame, vnode *VNode, clipX, clipY, clipW, clipH int) {
	// Quick reject: if the node is entirely outside the clip region, skip it
	if vnode.Y+vnode.H <= clipY || vnode.Y >= clipY+clipH {
		return // entirely above or below
	}
	if vnode.X+vnode.W <= clipX || vnode.X >= clipX+clipW {
		return // entirely left or right
	}

	style := vnode.Style

	switch vnode.Type {
	case "fragment":
		// Fragment: render children directly, no background/border
		for _, child := range vnode.Children {
			renderVNodeClipped(frame, child, clipX, clipY, clipW, clipH)
		}
		return

	case "text":
		// Render text with clipping
		renderTextClipped(frame, vnode, style, clipX, clipY, clipW, clipH)

	default:
		// Container: render background clipped
		if style.Background != "" {
			for y := vnode.Y; y < vnode.Y+vnode.H; y++ {
				if y < clipY || y >= clipY+clipH {
					continue
				}
				if y < 0 || y >= frame.Height {
					continue
				}
				for x := vnode.X; x < vnode.X+vnode.W; x++ {
					if x < clipX || x >= clipX+clipW {
						continue
					}
					if x < 0 || x >= frame.Width {
						continue
					}
					frame.Cells[y][x] = Cell{Char: ' ', Background: style.Background}
				}
			}
		}

		// Render border clipped (only if visible)
		if style.Border != "" {
			renderBorderClipped(frame, vnode, style.Border, clipX, clipY, clipW, clipH)
		}

		// Render children clipped
		for _, child := range vnode.Children {
			renderVNodeClipped(frame, child, clipX, clipY, clipW, clipH)
		}
	}
}

// renderTextClipped renders text content clipped to the given rect.
func renderTextClipped(frame *Frame, vnode *VNode, style Style, clipX, clipY, clipW, clipH int) {
	if vnode.Content == "" {
		return
	}

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
		px := startX + col
		py := startY + row

		// Clip check
		if py >= clipY && py < clipY+clipH && px >= clipX && px < clipX+clipW {
			if px >= 0 && px < frame.Width && py >= 0 && py < frame.Height {
				frame.Cells[py][px] = Cell{
					Char:       ch,
					Foreground: style.Foreground,
					Background: style.Background,
					Bold:       style.Bold,
					Dim:        style.Dim,
					Underline:  style.Underline,
				}
			}
		}

		col++
		if col >= availW {
			col = 0
			row++
		}
	}
}

// renderBorderClipped draws a border clipped to the given rect.
func renderBorderClipped(frame *Frame, vnode *VNode, borderType string, clipX, clipY, clipW, clipH int) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H

	if w < 2 || h < 2 {
		return
	}

	var tl, tr, bl, br, hLine, vLine rune
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

	setClipped := func(px, py int, ch rune) {
		if px >= clipX && px < clipX+clipW && py >= clipY && py < clipY+clipH {
			if px >= 0 && px < frame.Width && py >= 0 && py < frame.Height {
				frame.Cells[py][px] = Cell{Char: ch}
			}
		}
	}

	// Top border
	setClipped(x, y, tl)
	for i := 1; i < w-1; i++ {
		setClipped(x+i, y, hLine)
	}
	setClipped(x+w-1, y, tr)

	// Bottom border
	setClipped(x, y+h-1, bl)
	for i := 1; i < w-1; i++ {
		setClipped(x+i, y+h-1, hLine)
	}
	setClipped(x+w-1, y+h-1, br)

	// Vertical borders
	for i := 1; i < h-1; i++ {
		setClipped(x, y+i, vLine)
		setClipped(x+w-1, y+i, vLine)
	}
}

// renderScrollbar draws a vertical scrollbar at the given position.
// Uses │ for the track and █ for the thumb.
func renderScrollbar(frame *Frame, x, y, trackH int, vp *Viewport) {
	if x < 0 || x >= frame.Width || trackH <= 0 {
		return
	}

	thumbStart, thumbSize := vp.ScrollbarThumb(trackH)

	for i := 0; i < trackH; i++ {
		py := y + i
		if py < 0 || py >= frame.Height {
			continue
		}

		var ch rune
		if i >= thumbStart && i < thumbStart+thumbSize {
			ch = '█' // thumb
		} else {
			ch = '│' // track
		}

		frame.Cells[py][x] = Cell{
			Char:       ch,
			Foreground: "#666666", // dim gray for scrollbar
		}
	}
}

// renderInput renders a single-line text input VNode.
func renderInput(frame *Frame, vnode *VNode, style Style) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H
	if w <= 0 || h <= 0 || y >= frame.Height || x >= frame.Width {
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
		// Sync value from Lua props if provided
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
	bg := style.Background
	if bg == "" {
		bg = "#1a1a2e" // default input background (dark)
	}
	for cx := x; cx < x+w && cx < frame.Width; cx++ {
		if y < frame.Height {
			frame.Cells[y][cx] = Cell{Char: ' ', Background: bg}
		}
	}

	// Render text (with horizontal scroll)
	runes := []rune(text)
	fg := style.Foreground
	if isPlaceholder {
		fg = "#666666" // dim placeholder
	}

	for col := 0; col < w; col++ {
		runeIdx := col + scrollX
		if runeIdx >= len(runes) {
			break
		}
		px := x + col
		if px >= frame.Width {
			break
		}

		cellFg := fg
		cellBg := bg
		reverse := false

		// Check if this position is in the selection
		if state != nil && state.HasSelection() {
			lo := state.selMin()
			hi := state.selMax()
			if runeIdx >= lo && runeIdx < hi {
				reverse = true
			}
		}

		frame.Cells[y][px] = Cell{
			Char:       runes[runeIdx],
			Foreground: cellFg,
			Background: cellBg,
			Bold:       style.Bold,
			Reverse:    reverse,
		}
	}

	// Render cursor (reverse video at cursor position)
	if state != nil && state.Focused {
		cursorCol := state.CursorPos - scrollX
		if cursorCol >= 0 && cursorCol < w {
			px := x + cursorCol
			if px < frame.Width && y < frame.Height {
				ch := ' '
				if state.CursorPos < len(runes) {
					ch = runes[state.CursorPos]
				}
				frame.Cells[y][px] = Cell{
					Char:       ch,
					Foreground: bg,  // swap fg/bg for cursor
					Background: fg,
					Reverse:    true,
				}
			}
		}
	}

	// Render border if specified
	if style.Border != "" {
		renderBorder(frame, vnode, style.Border)
	}
}

// renderTextArea renders a multi-line text area VNode.
func renderTextArea(frame *Frame, vnode *VNode, style Style) {
	x, y := vnode.X, vnode.Y
	w, h := vnode.W, vnode.H
	if w <= 0 || h <= 0 || y >= frame.Height || x >= frame.Width {
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

	// Render background
	bg := style.Background
	if bg == "" {
		bg = "#1a1a2e"
	}
	for ry := y; ry < y+h && ry < frame.Height; ry++ {
		for cx := x; cx < x+w && cx < frame.Width; cx++ {
			frame.Cells[ry][cx] = Cell{Char: ' ', Background: bg}
		}
	}

	// Render border
	if style.Border != "" {
		renderBorder(frame, vnode, style.Border)
	}

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

	fg := style.Foreground
	if isPlaceholder {
		fg = "#666666"
	}

	// Render visible lines
	for row := 0; row < contentH; row++ {
		lineIdx := row + scrollY
		if lineIdx >= len(lines) {
			break
		}
		line := []rune(lines[lineIdx])

		py := contentY + row
		if py >= frame.Height {
			break
		}

		for col := 0; col < contentW; col++ {
			runeIdx := col + scrollX
			if runeIdx >= len(line) {
				break
			}
			px := contentX + col
			if px >= frame.Width {
				break
			}

			cellFg := fg
			reverse := false

			// Check selection
			if state != nil && state.HasSelection() && !isPlaceholder {
				// Convert line/col to absolute rune position
				absPos := lineColToPos(lines, lineIdx, runeIdx)
				lo := state.selMin()
				hi := state.selMax()
				if absPos >= lo && absPos < hi {
					reverse = true
				}
			}

			frame.Cells[py][px] = Cell{
				Char:       line[runeIdx],
				Foreground: cellFg,
				Background: bg,
				Bold:       style.Bold,
				Reverse:    reverse,
			}
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
			if px < frame.Width && py < frame.Height {
				ch := ' '
				if state.CursorPos < len(runes) {
					ch = runes[state.CursorPos]
				}
				frame.Cells[py][px] = Cell{
					Char:       ch,
					Foreground: bg,
					Background: fg,
					Reverse:    true,
				}
			}
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
		_ = L.PCall(6, 0, 0)
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
