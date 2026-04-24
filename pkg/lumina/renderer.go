// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
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
	defer L.Pop(1)

	vnode := NewVNode("box")

	// Get type field
	L.GetField(-1, "type")
	if !L.IsNoneOrNil(-1) {
		if t, ok := L.ToString(-1); ok {
			vnode.Type = t
		}
	}
	L.Pop(1)

	// Get content field (text nodes)
	L.GetField(-1, "content")
	if !L.IsNoneOrNil(-1) {
		if c, ok := L.ToString(-1); ok {
			vnode.Content = c
		}
	}
	L.Pop(1)

	// Get children field
	L.GetField(-1, "children")
	if L.Type(-1) == lua.TypeTable {
		// Iterate children array
		L.PushNil()
		for L.Next(-2) {
			// stack: [..., children, key, childValue]
			if L.Type(-1) == lua.TypeTable {
				child := LuaVNodeToVNode(L, -1)
				vnode.AddChild(child)
			}
			L.Pop(1) // pop value, keep key for iteration
		}
		L.Pop(1) // pop nil
	}
	L.Pop(1) // pop children table

	// Copy remaining fields as props
	L.PushNil()
	for L.Next(-2) {
		// stack: [..., vnode, key, value]
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
	case "text":
		// Render text content with wrapping support
		renderText(frame, vnode, cell)

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
