package layout

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
// Accounts for wide characters (CJK, emoji).
func stringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

// runeWidth returns the display width of a rune in terminal columns.
func runeWidth(r rune) int {
	if r == 0 {
		return 0
	}
	// Control characters
	if r < 0x20 || (r >= 0x7F && r < 0xA0) {
		return 0
	}
	// Combining characters (zero width)
	if r >= 0x0300 && r <= 0x036F {
		return 0
	}
	// CJK ranges (double width)
	if r >= 0x1100 && r <= 0x115F {
		return 2
	}
	if r >= 0x2E80 && r <= 0x303E {
		return 2
	}
	if r >= 0x3041 && r <= 0x33BF {
		return 2
	}
	if r >= 0x3400 && r <= 0x4DBF {
		return 2
	}
	if r >= 0x4E00 && r <= 0xA4CF {
		return 2
	}
	if r >= 0xAC00 && r <= 0xD7AF {
		return 2
	}
	if r >= 0xF900 && r <= 0xFAFF {
		return 2
	}
	if r >= 0xFE30 && r <= 0xFE6F {
		return 2
	}
	if r >= 0xFF01 && r <= 0xFF60 {
		return 2
	}
	if r >= 0xFFE0 && r <= 0xFFE6 {
		return 2
	}
	if r >= 0x1F000 && r <= 0x1FFFF {
		return 2
	}
	if r >= 0x20000 && r <= 0x2FFFF {
		return 2
	}
	if r >= 0x30000 && r <= 0x3FFFF {
		return 2
	}
	return 1
}

// resolveStyleSpacing copies s and expands Padding / Margin shorthand into
// per-side fields when shorthand > 0 and that side is still 0 (longhands win).
func resolveStyleSpacing(s Style) Style {
	out := s
	if s.Padding > 0 {
		if out.PaddingTop == 0 {
			out.PaddingTop = s.Padding
		}
		if out.PaddingBottom == 0 {
			out.PaddingBottom = s.Padding
		}
		if out.PaddingLeft == 0 {
			out.PaddingLeft = s.Padding
		}
		if out.PaddingRight == 0 {
			out.PaddingRight = s.Padding
		}
	}
	if s.Margin > 0 {
		if out.MarginTop == 0 {
			out.MarginTop = s.Margin
		}
		if out.MarginBottom == 0 {
			out.MarginBottom = s.Margin
		}
		if out.MarginLeft == 0 {
			out.MarginLeft = s.Margin
		}
		if out.MarginRight == 0 {
			out.MarginRight = s.Margin
		}
	}
	return out
}

func normalizeSpacingInTree(v *VNode) {
	if v == nil {
		return
	}
	v.Style = resolveStyleSpacing(v.Style)
	for _, c := range v.Children {
		normalizeSpacingInTree(c)
	}
}

// hasBorder returns true if the style specifies a visible border.
func hasBorder(s Style) bool {
	return s.Border == "single" || s.Border == "double" || s.Border == "rounded"
}

// isPositioned returns true if the style has absolute or fixed positioning.
func isPositioned(s Style) bool {
	return s.Position == "absolute" || s.Position == "fixed"
}

// applyRelativeOffset offsets a VNode from its normal flow position
// when position="relative".
func applyRelativeOffset(child *VNode, cs Style) {
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

// computeFlexLayout computes the layout for a VNode tree using flexbox semantics.
func computeFlexLayout(vnode *VNode, x, y, w, h int) {
	style := vnode.Style

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

	vnode.X = x
	vnode.Y = y
	vnode.W = w
	vnode.H = h

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

	switch vnode.Type {
	case "fragment":
		// Fragment: transparent container
		layoutVBox(vnode, x, y, w, h, style)
		return

	case "text":
		layoutText(vnode, layoutW)

	case "vbox":
		layoutVBox(vnode, contentX, contentY, layoutW, contentH, style)

	case "hbox":
		layoutHBox(vnode, contentX, contentY, layoutW, contentH, style)

	default:
		// Generic container (box, etc.) — stack children vertically like vbox
		layoutVBox(vnode, contentX, contentY, layoutW, contentH, style)
	}

	// After normal layout, handle absolute/fixed positioned children.
	for _, child := range vnode.Children {
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
			computeFlexLayout(child, cx, cy, cw, ch)

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
			child.X = cx
			child.Y = cy
			child.W = cw
			child.H = ch
			computeFlexLayout(child, cx, cy, cw, ch)
		}
	}
}

// layoutText measures a text node. It wraps text if it exceeds the available width.
func layoutText(vnode *VNode, availW int) {
	if availW <= 0 {
		vnode.H = 1
		return
	}
	contentW := stringWidth(vnode.Content)
	if contentW == 0 {
		vnode.H = 1
		return
	}
	// Calculate wrapped height based on display width
	lines := (contentW + availW - 1) / availW // ceiling division
	if lines < 1 {
		lines = 1
	}
	vnode.H = lines
}

// layoutVBox lays out children in a vertical stack with flex distribution.
func layoutVBox(vnode *VNode, contentX, contentY, contentW, contentH int, style Style) {
	if len(vnode.Children) == 0 {
		return
	}

	type childInfo struct {
		style      Style
		fixedH     int
		flexGrow   int
		finalH     int
		positioned bool
	}
	children := make([]childInfo, len(vnode.Children))

	// Count flow children (not absolute/fixed)
	flowCount := 0
	for i, child := range vnode.Children {
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
	availH := contentH - totalGaps
	if availH < 0 {
		availH = 0
	}

	fixedTotal := 0
	flexTotal := 0

	for i, child := range vnode.Children {
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

	// Calculate total used height for justify alignment
	totalUsed := 0
	for i := range children {
		if children[i].positioned {
			continue
		}
		totalUsed += children[i].finalH
	}
	totalUsed += totalGaps

	// Determine starting Y based on justify
	curY := contentY
	extraSpace := contentH - totalUsed
	if extraSpace < 0 {
		extraSpace = 0
	}

	gapSize := style.Gap
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

	// Position each child (skip absolute/fixed)
	flowIdx := 0
	for i, child := range vnode.Children {
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

		computeFlexLayout(child, childX, curY, childW, childH)
		applyRelativeOffset(child, children[i].style)
		curY += childH
		flowIdx++
		if flowIdx < flowCount {
			curY += gapSize
		}
	}
}

// layoutHBox lays out children in a horizontal row with flex distribution.
func layoutHBox(vnode *VNode, contentX, contentY, contentW, contentH int, style Style) {
	if len(vnode.Children) == 0 {
		return
	}

	type childInfo struct {
		style      Style
		fixedW     int
		flexGrow   int
		finalW     int
		positioned bool
	}
	children := make([]childInfo, len(vnode.Children))

	flowCount := 0
	for i, child := range vnode.Children {
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

	for i, child := range vnode.Children {
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
	for i, child := range vnode.Children {
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

		computeFlexLayout(child, curX, childY, childW, childH)
		applyRelativeOffset(child, children[i].style)
		curX += childW
		flowIdx++
		if flowIdx < flowCount {
			curX += gapSize
		}
	}
}
