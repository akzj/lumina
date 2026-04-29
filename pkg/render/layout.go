package render

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// LayoutFull computes layout for the entire Node tree.
// Used for initial render and after structural changes.
// The root is positioned at (x, y) with available size (w, h).
func LayoutFull(root *Node, x, y, w, h int) {
	if root == nil {
		return
	}
	normalizeSpacingInTree(root)
	computeFlex(root, x, y, w, h)
	clearLayoutDirty(root)
}

// LayoutIncremental only recomputes LayoutDirty subtrees.
// Nodes that are not LayoutDirty keep their cached (X, Y, W, H).
func LayoutIncremental(root *Node) {
	if root == nil {
		return
	}
	layoutDirtyWalk(root)
}

// layoutDirtyWalk walks the tree looking for LayoutDirty nodes.
// When found, it recomputes the subtree using the node's current cached position/size.
func layoutDirtyWalk(node *Node) {
	if !node.LayoutDirty {
		// This node's layout is cached and valid.
		// But check children in case a descendant is dirty.
		for _, child := range node.Children {
			layoutDirtyWalk(child)
		}
		return
	}

	// This node needs re-layout.
	// Recompute using its CURRENT (cached) position and size as the container.
	normalizeSpacing(node)
	computeFlex(node, node.X, node.Y, node.W, node.H)
	node.LayoutDirty = false
	// All children within this subtree are now re-laid-out.
	clearLayoutDirtyBelow(node)
}

// clearLayoutDirty clears LayoutDirty on the entire tree.
func clearLayoutDirty(node *Node) {
	node.LayoutDirty = false
	for _, child := range node.Children {
		clearLayoutDirty(child)
	}
}

// clearLayoutDirtyBelow clears LayoutDirty on all children (not the node itself).
func clearLayoutDirtyBelow(node *Node) {
	for _, child := range node.Children {
		child.LayoutDirty = false
		clearLayoutDirtyBelow(child)
	}
}

// normalizeSpacingInTree expands Padding/Margin shorthand into per-side fields
// for the entire tree.
func normalizeSpacingInTree(node *Node) {
	normalizeSpacing(node)
	for _, child := range node.Children {
		normalizeSpacingInTree(child)
	}
}

// normalizeSpacing expands Padding/Margin shorthand into per-side fields
// for a single node. Longhands (non-zero) win over shorthand.
func normalizeSpacing(node *Node) {
	s := &node.Style
	if s.Padding > 0 {
		if s.PaddingTop == 0 {
			s.PaddingTop = s.Padding
		}
		if s.PaddingBottom == 0 {
			s.PaddingBottom = s.Padding
		}
		if s.PaddingLeft == 0 {
			s.PaddingLeft = s.Padding
		}
		if s.PaddingRight == 0 {
			s.PaddingRight = s.Padding
		}
	}
	if s.Margin > 0 {
		if s.MarginTop == 0 {
			s.MarginTop = s.Margin
		}
		if s.MarginBottom == 0 {
			s.MarginBottom = s.Margin
		}
		if s.MarginLeft == 0 {
			s.MarginLeft = s.Margin
		}
		if s.MarginRight == 0 {
			s.MarginRight = s.Margin
		}
	}
}

// --- Core layout helpers ---

// clamp constrains v to [lo, hi]. If hi is 0, no upper bound is applied.
func clamp(v, lo, hi int) int {
	if v < lo {
		v = lo
	}
	if hi > 0 && v > hi {
		v = hi
	}
	return v
}

// stringWidth returns the display width of a string in terminal columns.
func stringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

// runeWidth returns the display width of a rune in terminal columns.
// Uses go-runewidth for complete Unicode support (CJK, emoji, etc.).
func runeWidth(r rune) int {
	return runewidth.RuneWidth(r)
}

// hasBorder returns true if the style specifies a visible border.
func hasBorder(s Style) bool {
	return s.Border == "single" || s.Border == "double" || s.Border == "rounded"
}

// isPositioned returns true if the style has absolute or fixed positioning.
func isPositioned(s Style) bool {
	return s.Position == "absolute" || s.Position == "fixed"
}

// applyRelativeOffset offsets a Node from its normal flow position
// when position="relative".
func applyRelativeOffset(child *Node, cs Style) {
	if cs.Position != "relative" {
		return
	}
	child.X += cs.Left
	child.Y += cs.Top
	if cs.Right > 0 && cs.Left == 0 {
		child.X -= cs.Right
	}
	if cs.Bottom > 0 && cs.Top == 0 {
		child.Y -= cs.Bottom
	}
}

// --- Core flexbox layout ---

// computeFlex computes the layout for a Node tree using flexbox semantics.
func computeFlex(node *Node, x, y, w, h int) {
	style := node.Style

	// Apply margin — shrinks the area this node occupies
	x += style.MarginLeft
	y += style.MarginTop
	w -= style.MarginLeft + style.MarginRight
	h -= style.MarginTop + style.MarginBottom
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}

	// Apply fixed sizing with min/max constraints
	if style.Width > 0 {
		w = clamp(style.Width, style.MinWidth, style.MaxWidth)
	} else if style.MinWidth > 0 || style.MaxWidth > 0 {
		w = clamp(w, style.MinWidth, style.MaxWidth)
	}
	if style.Height > 0 {
		h = clamp(style.Height, style.MinHeight, style.MaxHeight)
	} else if style.MinHeight > 0 || style.MaxHeight > 0 {
		h = clamp(h, style.MinHeight, style.MaxHeight)
	}

	// Mark PaintDirty if position/size changed
	if node.X != x || node.Y != y || node.W != w || node.H != h {
		node.PaintDirty = true
		// Parent must repaint to clear the old position area
		// (ClearRect on old pos alone would leave a hole in parent's background)
		if node.Parent != nil {
			node.Parent.PaintDirty = true
		}
	}

	// Track position changes for clearing old region during paint
	if node.X != x || node.Y != y || node.W != w || node.H != h {
		node.OldX, node.OldY, node.OldW, node.OldH = node.X, node.Y, node.W, node.H
		node.PositionChanged = true
		node.PaintDirty = true
	}

	node.X = x
	node.Y = y
	node.W = w
	node.H = h

	// Calculate content area (inside border + padding)
	borderWidth := 0
	if hasBorder(style) {
		borderWidth = 1
	}

	contentX := x + borderWidth + style.PaddingLeft
	contentY := y + borderWidth + style.PaddingTop
	contentW := w - 2*borderWidth - style.PaddingLeft - style.PaddingRight
	contentH := h - 2*borderWidth - style.PaddingTop - style.PaddingBottom
	if contentW < 0 {
		contentW = 0
	}
	if contentH < 0 {
		contentH = 0
	}

	// If overflow=scroll, reserve 1 column for scrollbar
	layoutW := contentW
	if style.Overflow == "scroll" && contentW > 1 {
		layoutW = contentW - 1
	}

	switch node.Type {
	case "fragment":
		// Fragment: transparent container
		layoutVBox(node, x, y, w, h, style)
		return

	case "component":
		// Component placeholder: transparent container, passes through to children.
		// For absolute/fixed positioned children, position relative to the parent
		// container's content area, not the placeholder's stacked position.
		// This prevents double-offset when multiple component placeholders are
		// stacked in a vbox (each placeholder gets a different y from stacking).
		parentX, parentY, parentW, parentH := x, y, w, h
		if node.Parent != nil {
			p := node.Parent
			bw := 0
			if p.Style.Border != "" && p.Style.Border != "none" {
				bw = 1
			}
			parentX = p.X + bw + p.Style.PaddingLeft
			parentY = p.Y + bw + p.Style.PaddingTop
			parentW = p.W - 2*bw - p.Style.PaddingLeft - p.Style.PaddingRight
			parentH = p.H - 2*bw - p.Style.PaddingTop - p.Style.PaddingBottom
		}
		for _, child := range node.Children {
			cs := child.Style
			if isPositioned(cs) {
				// Position relative to parent container's content area
				cx, cy, cw, ch := parentX+cs.Left, parentY+cs.Top, parentW, parentH
				if cs.Width > 0 {
					cw = cs.Width
				}
				if cs.Height > 0 {
					ch = cs.Height
				}
				if cs.Right >= 0 && cs.Left == 0 {
					cx = parentX + parentW - cw - cs.Right
				}
				if cs.Bottom >= 0 && cs.Top == 0 {
					cy = parentY + parentH - ch - cs.Bottom
				}
				computeFlex(child, cx, cy, cw, ch)
			} else {
				computeFlex(child, x, y, w, h)
			}
		}
		return

	case "text":
		layoutText(node, layoutW)

	case "vbox":
		layoutVBox(node, contentX, contentY, layoutW, contentH, style)

	case "hbox":
		layoutHBox(node, contentX, contentY, layoutW, contentH, style)

	default:
		// Generic container (box, etc.) — stack children vertically like vbox
		layoutVBox(node, contentX, contentY, layoutW, contentH, style)
	}

	// After normal layout, handle absolute/fixed positioned children.
	for _, child := range node.Children {
		cs := child.Style
		switch cs.Position {
		case "absolute":
			cx := contentX + cs.Left
			cy := contentY + cs.Top
			cw := child.W
			ch := child.H
			if cs.Width > 0 {
				cw = cs.Width
			}
			if cs.Height > 0 {
				ch = cs.Height
			}
			// Default to remaining parent dimensions when no explicit size
			if cw <= 0 {
				cw = contentW - cs.Left
			}
			if ch <= 0 {
				ch = 1 // text nodes default to 1 row
			}
			if cs.Right >= 0 && cs.Left == 0 {
				cx = contentX + contentW - cw - cs.Right
			}
			if cs.Bottom >= 0 && cs.Top == 0 {
				cy = contentY + contentH - ch - cs.Bottom
			}
			child.X = cx
			child.Y = cy
			child.W = cw
			child.H = ch
			computeFlex(child, cx, cy, cw, ch)

		case "fixed":
			cx := cs.Left
			cy := cs.Top
			cw := child.W
			ch := child.H
			if cs.Width > 0 {
				cw = cs.Width
			}
			if cs.Height > 0 {
				ch = cs.Height
			}
			// Default to parent dimensions for fixed-positioned elements
			// that don't have explicit width/height set
			if cw <= 0 {
				cw = node.W - cs.Left // fill remaining width from left edge
			}
			if ch <= 0 {
				ch = 1 // text nodes default to 1 row
			}
			child.X = cx
			child.Y = cy
			child.W = cw
			child.H = ch
			computeFlex(child, cx, cy, cw, ch)
		}
	}
}

// layoutText measures a text node. It wraps text if it exceeds the available width.
func layoutText(node *Node, availW int) {
	if availW <= 0 {
		node.H = 1
		return
	}
	if node.Content == "" {
		node.H = 1
		return
	}
	// Split by newlines and calculate height for each line (with wrapping)
	lines := strings.Split(node.Content, "\n")
	totalH := 0
	for _, line := range lines {
		lineW := stringWidth(line)
		if lineW == 0 {
			totalH += 1 // empty line still takes 1 row
		} else {
			totalH += (lineW + availW - 1) / availW // ceiling division for wrapping
		}
	}
	if totalH < 1 {
		totalH = 1
	}
	node.H = totalH
}

// layoutVBox lays out children in a vertical stack with flex distribution.
func layoutVBox(node *Node, contentX, contentY, contentW, contentH int, style Style) {
	if len(node.Children) == 0 {
		return
	}

	isScroll := style.Overflow == "scroll"

	type childInfo struct {
		style      Style
		fixedH     int
		flexGrow   int
		finalH     int
		positioned bool
	}
	children := make([]childInfo, len(node.Children))

	// Count flow children (not absolute/fixed)
	flowCount := 0
	for i, child := range node.Children {
		cs := child.Style
		children[i].style = cs
		children[i].positioned = isPositioned(cs)
		if !children[i].positioned {
			flowCount++
		}
	}

	totalGaps := 0
	if flowCount > 1 {
		totalGaps = style.Gap * (flowCount - 1)
	}

	if isScroll {
		// Scroll container: children get natural heights, no flex distribution.
		// Children may overflow the container's contentH.
		for i, child := range node.Children {
			if children[i].positioned {
				continue
			}
			cs := children[i].style
			marginV := cs.MarginTop + cs.MarginBottom

			if child.Type == "fragment" {
				fragH := len(child.Children)
				if fragH < 1 {
					fragH = 1
				}
				children[i].finalH = fragH
			} else if cs.Height > 0 {
				children[i].finalH = clamp(cs.Height, cs.MinHeight, cs.MaxHeight) + marginV
			} else if cs.MinHeight > 0 {
				children[i].finalH = cs.MinHeight + marginV
			} else if child.Type == "component" && len(child.Children) > 0 {
				// Component placeholder with no explicit height: use the grafted
				// child's height if it has one. This lets defineComponent children
				// inside scroll containers size correctly based on their rendered content.
				graftedH := 0
				for _, gc := range child.Children {
					if gc.Style.Height > 0 {
						graftedH = gc.Style.Height
						break
					}
				}
				if graftedH > 0 {
					children[i].finalH = graftedH + marginV
				} else {
					children[i].finalH = 1 + marginV
				}
			} else {
				// Natural height: 1 for all types (flex children get 1, not distributed)
				children[i].finalH = 1 + marginV
			}
		}
	} else {
		// Normal (non-scroll) container: distribute available height via flex.
		availH := contentH - totalGaps
		if availH < 0 {
			availH = 0
		}

		fixedTotal := 0
		flexTotal := 0

		for i, child := range node.Children {
			cs := children[i].style
			if children[i].positioned {
				continue
			}

			marginV := cs.MarginTop + cs.MarginBottom

			// Fragment: natural height = number of children
			if child.Type == "fragment" {
				fragH := len(child.Children)
				if fragH < 1 {
					fragH = 1
				}
				children[i].fixedH = fragH
				fixedTotal += children[i].fixedH
			} else if cs.Height > 0 {
				h := clamp(cs.Height, cs.MinHeight, cs.MaxHeight)
				children[i].fixedH = h + marginV
				fixedTotal += children[i].fixedH
			} else if cs.MinHeight > 0 && cs.Flex == 0 {
				children[i].fixedH = cs.MinHeight + marginV
				fixedTotal += children[i].fixedH
			} else if cs.Flex > 0 {
				children[i].flexGrow = cs.Flex
				flexTotal += cs.Flex
			} else {
				// No flex, no fixed height.
				// Leaf types (text, input, textarea) get minimum 1 row.
				// Container types get implicit flex=1.
				switch child.Type {
				case "text", "input", "textarea":
					children[i].fixedH = 1 + marginV
					fixedTotal += children[i].fixedH
				default:
					children[i].flexGrow = 1
					flexTotal += 1
				}
			}
		}

		// Distribute remaining space to flex children
		remainH := availH - fixedTotal
		if remainH < 0 {
			remainH = 0
		}

		for i := range children {
			if children[i].positioned {
				continue
			}
			if children[i].flexGrow > 0 {
				if flexTotal > 0 {
					children[i].finalH = (remainH * children[i].flexGrow) / flexTotal
				}
				if children[i].finalH < 1 {
					children[i].finalH = 1
				}
				children[i].finalH = clamp(children[i].finalH, children[i].style.MinHeight, children[i].style.MaxHeight)
			} else {
				children[i].finalH = children[i].fixedH
			}
		}
	}

	// Calculate total used height for justify alignment
	totalUsed := 0
	for i := range children {
		if children[i].positioned {
			continue
		}
		totalUsed += children[i].finalH
	}
	totalUsed += totalGaps

	// Determine starting Y based on justify (skip for scroll containers)
	curY := contentY
	gapSize := style.Gap
	if !isScroll {
		extraSpace := contentH - totalUsed
		if extraSpace < 0 {
			extraSpace = 0
		}

		switch style.Justify {
		case "center":
			curY += extraSpace / 2
		case "end":
			curY += extraSpace
		case "space-between":
			if flowCount > 1 {
				gapSize = 0
				totalBetween := flowCount - 1
				usedNoGap := totalUsed - totalGaps
				spaceBetween := contentH - usedNoGap
				if spaceBetween > 0 && totalBetween > 0 {
					gapSize = spaceBetween / totalBetween
				}
			}
		case "space-around":
			if flowCount > 0 {
				gapSize = 0
				usedNoGap := totalUsed - totalGaps
				totalSlots := flowCount * 2
				spaceAround := contentH - usedNoGap
				if spaceAround > 0 && totalSlots > 0 {
					halfGap := spaceAround / totalSlots
					curY += halfGap
					gapSize = halfGap * 2
				}
			}
		}
	}

	// Position each child (skip absolute/fixed)
	flowIdx := 0
	var lastFlowChildNode *Node
	for i, child := range node.Children {
		if children[i].positioned {
			continue
		}

		childH := children[i].finalH
		childW := contentW

		// Cross-axis alignment (align)
		childX := contentX
		switch style.Align {
		case "center":
			cs := children[i].style
			if cs.Width > 0 {
				cw := clamp(cs.Width, cs.MinWidth, cs.MaxWidth)
				childX = contentX + (contentW-cw)/2
				childW = cw
			}
		case "end":
			cs := children[i].style
			if cs.Width > 0 {
				cw := clamp(cs.Width, cs.MinWidth, cs.MaxWidth)
				childX = contentX + contentW - cw
				childW = cw
			}
		case "start":
			cs := children[i].style
			if cs.Width > 0 {
				cw := clamp(cs.Width, cs.MinWidth, cs.MaxWidth)
				childW = cw
			}
		default: // "stretch" — use full width
		}

		computeFlex(child, childX, curY, childW, childH)
		applyRelativeOffset(child, children[i].style)
		lastFlowChildNode = child
		curY += childH
		flowIdx++
		if flowIdx < flowCount {
			curY += gapSize
		}
	}

	// For scroll containers, store the total content height
	if isScroll && lastFlowChildNode != nil {
		node.ScrollHeight = (lastFlowChildNode.Y + lastFlowChildNode.H) - contentY
	}
}

// layoutHBox lays out children in a horizontal row with flex distribution.
func layoutHBox(node *Node, contentX, contentY, contentW, contentH int, style Style) {
	if len(node.Children) == 0 {
		return
	}

	type childInfo struct {
		style      Style
		fixedW     int
		flexGrow   int
		finalW     int
		positioned bool
	}
	children := make([]childInfo, len(node.Children))

	flowCount := 0
	for i, child := range node.Children {
		cs := child.Style
		children[i].style = cs
		children[i].positioned = isPositioned(cs)
		if !children[i].positioned {
			flowCount++
		}
		_ = child
	}

	totalGaps := 0
	if flowCount > 1 {
		totalGaps = style.Gap * (flowCount - 1)
	}
	availW := contentW - totalGaps
	if availW < 0 {
		availW = 0
	}

	fixedTotal := 0
	flexTotal := 0

	for i, child := range node.Children {
		if children[i].positioned {
			continue
		}
		cs := children[i].style

		marginH := cs.MarginLeft + cs.MarginRight
		if cs.Width > 0 {
			w := clamp(cs.Width, cs.MinWidth, cs.MaxWidth)
			children[i].fixedW = w + marginH
			fixedTotal += children[i].fixedW
		} else if cs.MinWidth > 0 && cs.Flex == 0 {
			children[i].fixedW = cs.MinWidth + marginH
			fixedTotal += children[i].fixedW
		} else if cs.Flex > 0 {
			children[i].flexGrow = cs.Flex
			flexTotal += cs.Flex
		} else {
			// No flex, no fixed width.
			// Text nodes use natural content width.
			// Container types get implicit flex=1.
			switch child.Type {
			case "text":
				naturalW := 1
				if child.Content != "" {
					naturalW = stringWidth(child.Content)
					if naturalW < 1 {
						naturalW = 1
					}
				}
				children[i].fixedW = naturalW + marginH
				fixedTotal += children[i].fixedW
			case "input", "textarea":
				children[i].fixedW = 1 + marginH
				fixedTotal += children[i].fixedW
			default:
				// Container — treat as flex=1
				children[i].flexGrow = 1
				flexTotal += 1
			}
		}
	}

	remainW := availW - fixedTotal
	if remainW < 0 {
		remainW = 0
	}

	for i := range children {
		if children[i].positioned {
			continue
		}
		if children[i].flexGrow > 0 {
			if flexTotal > 0 {
				children[i].finalW = (remainW * children[i].flexGrow) / flexTotal
			}
			if children[i].finalW < 1 {
				children[i].finalW = 1
			}
			children[i].finalW = clamp(children[i].finalW, children[i].style.MinWidth, children[i].style.MaxWidth)
		} else {
			children[i].finalW = children[i].fixedW
		}
	}

	totalUsed := 0
	for i := range children {
		if children[i].positioned {
			continue
		}
		totalUsed += children[i].finalW
	}
	totalUsed += totalGaps

	curX := contentX
	extraSpace := contentW - totalUsed
	if extraSpace < 0 {
		extraSpace = 0
	}

	gapSize := style.Gap
	switch style.Justify {
	case "center":
		curX += extraSpace / 2
	case "end":
		curX += extraSpace
	case "space-between":
		if flowCount > 1 {
			gapSize = 0
			totalBetween := flowCount - 1
			usedNoGap := totalUsed - totalGaps
			spaceBetween := contentW - usedNoGap
			if spaceBetween > 0 && totalBetween > 0 {
				gapSize = spaceBetween / totalBetween
			}
		}
	case "space-around":
		if flowCount > 0 {
			gapSize = 0
			usedNoGap := totalUsed - totalGaps
			totalSlots := flowCount * 2
			spaceAround := contentW - usedNoGap
			if spaceAround > 0 && totalSlots > 0 {
				halfGap := spaceAround / totalSlots
				curX += halfGap
				gapSize = halfGap * 2
			}
		}
	}

	flowIdx := 0
	for i, child := range node.Children {
		if children[i].positioned {
			continue
		}

		childW := children[i].finalW
		childH := contentH

		childY := contentY
		switch style.Align {
		case "center":
			cs := children[i].style
			if cs.Height > 0 {
				ch := clamp(cs.Height, cs.MinHeight, cs.MaxHeight)
				childY = contentY + (contentH-ch)/2
				childH = ch
			}
		case "end":
			cs := children[i].style
			if cs.Height > 0 {
				ch := clamp(cs.Height, cs.MinHeight, cs.MaxHeight)
				childY = contentY + contentH - ch
				childH = ch
			}
		case "start":
			cs := children[i].style
			if cs.Height > 0 {
				ch := clamp(cs.Height, cs.MinHeight, cs.MaxHeight)
				childH = ch
			}
		default: // "stretch"
		}

		// Clamp child width so it doesn't extend beyond the parent's content area
		maxChildW := (contentX + contentW) - curX
		if maxChildW < 0 {
			maxChildW = 0
		}
		if childW > maxChildW {
			childW = maxChildW
		}

		computeFlex(child, curX, childY, childW, childH)
		applyRelativeOffset(child, children[i].style)
		curX += childW
		flowIdx++
		if flowIdx < flowCount {
			curX += gapSize
		}
	}
}
