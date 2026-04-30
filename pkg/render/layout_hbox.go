package render

import "sort"

func layoutHBox(node *Node, contentX, contentY, contentW, contentH int, style Style, depth int) {
	if len(node.Children) == 0 {
		return
	}

	// flex-wrap: delegate to wrap layout
	if style.FlexWrap == "wrap" || style.FlexWrap == "wrap-reverse" {
		layoutHBoxWrap(node, contentX, contentY, contentW, contentH, style, depth)
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
		children[i].positioned = isPositioned(cs) || cs.Display == "none"
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
		if rw := resolveWidth(cs, contentW); rw > 0 {
			w := clamp(rw, cs.MinWidth, cs.MaxWidth)
			children[i].fixedW = w + marginH
			fixedTotal += children[i].fixedW
		} else if cs.FlexBasis > 0 && cs.Flex > 0 {
			// flexBasis: use as initial main size, still participate in flex-grow
			children[i].fixedW = cs.FlexBasis + marginH
			fixedTotal += children[i].fixedW
			children[i].flexGrow = cs.Flex
			flexTotal += cs.Flex
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
			baseW := children[i].fixedW // flexBasis base (0 if no basis)
			if flexTotal > 0 {
				children[i].finalW = baseW + (remainW*children[i].flexGrow)/flexTotal
			} else {
				children[i].finalW = baseW
			}
			if children[i].finalW < 1 {
				children[i].finalW = 1
			}
			children[i].finalW = clamp(children[i].finalW, children[i].style.MinWidth, children[i].style.MaxWidth)
		} else {
			children[i].finalW = children[i].fixedW
		}
	}

	// flexShrink: if total sizes exceed available space, shrink proportionally
	totalAfterGrow := 0
	for i := range children {
		if !children[i].positioned {
			totalAfterGrow += children[i].finalW
		}
	}
	overflowW := totalAfterGrow - availW
	if overflowW > 0 {
		shrinkTotal := 0
		for i := range children {
			if !children[i].positioned && children[i].style.FlexShrink > 0 {
				shrinkTotal += children[i].style.FlexShrink
			}
		}
		if shrinkTotal > 0 {
			for i := range children {
				if !children[i].positioned && children[i].style.FlexShrink > 0 {
					shrink := (overflowW * children[i].style.FlexShrink) / shrinkTotal
					children[i].finalW -= shrink
					if children[i].finalW < 1 {
						children[i].finalW = 1
					}
					children[i].finalW = clamp(children[i].finalW, children[i].style.MinWidth, children[i].style.MaxWidth)
				}
			}
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

	// Build order-sorted index for positioning
	orderIndicesH := make([]int, 0, len(node.Children))
	for i := range node.Children {
		if !children[i].positioned {
			orderIndicesH = append(orderIndicesH, i)
		}
	}
	sort.SliceStable(orderIndicesH, func(a, b int) bool {
		return children[orderIndicesH[a]].style.Order < children[orderIndicesH[b]].style.Order
	})

	// Position each child in order-sorted sequence
	flowIdx := 0
	for _, i := range orderIndicesH {
		child := node.Children[i]

		childW := children[i].finalW
		childH := contentH

		childY := contentY
		cs := children[i].style
		rh := resolveHeight(cs, contentH)
		cMinH := resolveMinH(cs, contentH)
		cMaxH := resolveMaxH(cs, contentH)
		if rh > 0 {
			rh = clamp(rh, cMinH, cMaxH)
		} else if cMinH > 0 || cMaxH > 0 {
			rh = clamp(contentH, cMinH, cMaxH)
		}

		// Cross-axis alignment: check alignSelf first, fall back to parent's align
		align := style.Align
		if cs.AlignSelf != "" {
			align = cs.AlignSelf
		}
		if rh > 0 && rh < contentH {
			switch align {
			case "center":
				childY = contentY + (contentH-rh)/2
				childH = rh
			case "end":
				childY = contentY + contentH - rh
				childH = rh
			case "start":
				childH = rh
			default: // "stretch"
				childH = rh
			}
		} else if rh > 0 {
			childH = rh
		}

		// Clamp child width so it doesn't extend beyond the parent's content area
		maxChildW := (contentX + contentW) - curX
		if maxChildW < 0 {
			maxChildW = 0
		}
		if mw := minOuterWidthForBorderPaint(child); mw > 0 && childW < mw {
			childW = mw
		}
		if childW > maxChildW {
			if mw := minOuterWidthForBorderPaint(child); mw > maxChildW {
				// Row tail too narrow for border paint; keep drawable width (may extend past clip).
				childW = mw
			} else {
				childW = maxChildW
			}
		}

		computeFlex(child, curX, childY, childW, childH, depth+1)
		applyRelativeOffset(child, children[i].style)
		curX += childW
		flowIdx++
		if flowIdx < flowCount {
			curX += gapSize
		}
	}

	// Cross-axis: children may end taller than the slot (e.g. explicit style.Height on
	// text while parent passed a short contentH). Grow outer H so fill + border match
	// painted descendants (Lux SplitButton in flex rows).
	maxChildBottom := contentY
	for i, child := range node.Children {
		if children[i].positioned {
			continue
		}
		b := child.Y + child.H
		if b > maxChildBottom {
			maxChildBottom = b
		}
	}
	bw := 0
	if hasBorder(style) {
		bw = 1
	}
	pt, pb := style.PaddingTop, style.PaddingBottom
	if style.Padding > 0 {
		if pt == 0 {
			pt = style.Padding
		}
		if pb == 0 {
			pb = style.Padding
		}
	}
	innerUsed := maxChildBottom - contentY
	if innerUsed < 1 {
		innerUsed = 1
	}
	wantOuter := innerUsed + 2*bw + pt + pb
	if wantOuter > node.H {
		node.H = wantOuter
		node.PaintDirty = true
		if node.Parent != nil {
			node.Parent.PaintDirty = true
		}
	} else if style.Height <= 0 && flowCount > 0 {
		// Cross-axis flex slice can be taller than flow children (e.g. lone row in a
		// card vbox). Shrink outer H so background/border match content (Lux Button.Group).
		// Skip when parent is a flex container that controls this node's height.
		parentControlsHeight := false
		if node.Parent != nil {
			pt := node.Parent.Type
			if pt == "vbox" || pt == "hbox" || pt == "box" || pt == "component" {
				parentControlsHeight = true
			}
		}
		if !parentControlsHeight {
			mnh := resolveMinH(style, contentH)
			shrinkTo := wantOuter
			if shrinkTo < mnh {
				shrinkTo = mnh
			}
			if shrinkTo < node.H {
				node.H = shrinkTo
				node.PaintDirty = true
				if node.Parent != nil {
					node.Parent.PaintDirty = true
				}
			}
		}
	}

	// For scroll containers, store the total content height (cross-axis)
	if style.Overflow == "scroll" || style.Overflow == "auto" {
		maxBottom := 0
		for i, child := range node.Children {
			if children[i].positioned {
				continue
			}
			bottom := child.Y + child.H - contentY
			if bottom > maxBottom {
				maxBottom = bottom
			}
		}
		node.ScrollHeight = maxBottom
	}
}

// wrapHBoxItemDesiredWidth estimates a flex item's outer width along the main axis
// for flex-wrap row packing when the item has no explicit width or flex-basis.
// It includes the item's horizontal margin. Nested hbox widths sum children; stacked
// containers (vbox, box, component, …) use the max child width.
func wrapHBoxItemDesiredWidth(child *Node, parentContentW int) int {
	cs := child.Style
	marginH := cs.MarginLeft + cs.MarginRight
	if isPositioned(cs) || cs.Display == "none" {
		return 0
	}
	if rw := resolveWidth(cs, parentContentW); rw > 0 {
		return clamp(rw, resolveMinW(cs, parentContentW), resolveMaxW(cs, parentContentW)) + marginH
	}
	if cs.FlexBasis > 0 {
		w := cs.FlexBasis
		return clamp(w, resolveMinW(cs, parentContentW), resolveMaxW(cs, parentContentW)) + marginH
	}

	borderW := 0
	if hasBorder(cs) {
		borderW = 1
	}
	hpad := cs.PaddingLeft + cs.PaddingRight + 2*borderW

	switch child.Type {
	case "text":
		n := 1
		if child.Content != "" {
			n = stringWidth(child.Content)
			if n < 1 {
				n = 1
			}
		}
		return clamp(n+hpad, resolveMinW(cs, parentContentW), resolveMaxW(cs, parentContentW)) + marginH
	case "input", "textarea":
		return clamp(1+hpad, resolveMinW(cs, parentContentW), resolveMaxW(cs, parentContentW)) + marginH
	case "hbox":
		inner := 0
		flow := 0
		for _, ch := range child.Children {
			if isPositioned(ch.Style) || ch.Style.Display == "none" {
				continue
			}
			inner += wrapHBoxItemDesiredWidth(ch, parentContentW)
			flow++
		}
		if flow > 1 {
			inner += child.Style.Gap * (flow - 1)
		}
		out := inner + hpad
		return clamp(out, resolveMinW(cs, parentContentW), resolveMaxW(cs, parentContentW)) + marginH
	default:
		// vbox, box, fragment, component, scroll, etc. — vertical stack: width ≈ max child
		maxW := 0
		for _, ch := range child.Children {
			if isPositioned(ch.Style) || ch.Style.Display == "none" {
				continue
			}
			w := wrapHBoxItemDesiredWidth(ch, parentContentW)
			if w > maxW {
				maxW = w
			}
		}
		if maxW < 1 {
			maxW = 1
		}
		out := maxW + hpad
		return clamp(out, resolveMinW(cs, parentContentW), resolveMaxW(cs, parentContentW)) + marginH
	}
}

// --- flex-wrap layout ---

// layoutHBoxWrap lays out children in a horizontal row with wrapping.
// When children overflow the available width, they wrap to the next row.
func layoutHBoxWrap(node *Node, contentX, contentY, contentW, contentH int, style Style, depth int) {
	type rowItem struct {
		childIdx int
		desiredW int
		style    Style
	}

	// Collect flow children and their desired widths
	var allItems []rowItem
	for i, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			continue
		}
		marginH := cs.MarginLeft + cs.MarginRight

		desiredW := 0
		if rw := resolveWidth(cs, contentW); rw > 0 {
			desiredW = clamp(rw, resolveMinW(cs, contentW), resolveMaxW(cs, contentW)) + marginH
		} else if cs.FlexBasis > 0 {
			desiredW = cs.FlexBasis + marginH
		} else {
			// Auto width — intrinsic measure for nested flex rows (e.g. Lux Button hbox)
			desiredW = wrapHBoxItemDesiredWidth(child, contentW)
		}
		allItems = append(allItems, rowItem{childIdx: i, desiredW: desiredW, style: cs})
	}

	if len(allItems) == 0 {
		return
	}

	// Build rows greedily
	type row struct {
		items []rowItem
	}
	var rows []row
	var currentRow []rowItem
	currentRowW := 0

	for _, item := range allItems {
		gap := 0
		if len(currentRow) > 0 {
			gap = style.Gap
		}
		if currentRowW+gap+item.desiredW > contentW && len(currentRow) > 0 {
			rows = append(rows, row{items: currentRow})
			currentRow = nil
			currentRowW = 0
			gap = 0
		}
		currentRow = append(currentRow, item)
		currentRowW += gap + item.desiredW
	}
	if len(currentRow) > 0 {
		rows = append(rows, row{items: currentRow})
	}

	// Layout each row
	rowGap := style.Gap
	curY := contentY

	// For wrap-reverse, reverse the row order
	if style.FlexWrap == "wrap-reverse" {
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
	}

	for ri, r := range rows {
		// Determine row width usage and flex distribution
		rowFlexTotal := 0
		rowFixedTotal := 0
		for _, item := range r.items {
			if item.style.Flex > 0 && resolveWidth(item.style, contentW) == 0 && item.style.FlexBasis == 0 {
				rowFlexTotal += item.style.Flex
				// Flex items start with their minimum desired width
			}
			rowFixedTotal += item.desiredW
		}
		rowGaps := 0
		if len(r.items) > 1 {
			rowGaps = style.Gap * (len(r.items) - 1)
		}
		remainW := contentW - rowFixedTotal - rowGaps
		if remainW < 0 {
			remainW = 0
		}

		// Compute final widths
		finalWidths := make([]int, len(r.items))
		for idx, item := range r.items {
			if item.style.Flex > 0 && resolveWidth(item.style, contentW) == 0 && item.style.FlexBasis == 0 && rowFlexTotal > 0 {
				finalWidths[idx] = item.desiredW + (remainW*item.style.Flex)/rowFlexTotal
			} else {
				finalWidths[idx] = item.desiredW
			}
			if finalWidths[idx] < 1 {
				finalWidths[idx] = 1
			}
		}
		for idx, item := range r.items {
			child := node.Children[item.childIdx]
			if mw := minOuterWidthForBorderPaint(child); mw > finalWidths[idx] {
				finalWidths[idx] = mw
			}
		}

		// Position children in this row
		curX := contentX
		rowH := 1

		for idx, item := range r.items {
			child := node.Children[item.childIdx]
			childW := finalWidths[idx]
			childH := contentH
			if rh := resolveHeight(item.style, contentH); rh > 0 {
				childH = rh
			}

			maxChildW := (contentX + contentW) - curX
			if maxChildW < 0 {
				maxChildW = 0
			}
			if mw := minOuterWidthForBorderPaint(child); mw > 0 && childW < mw {
				childW = mw
			}
			if childW > maxChildW {
				if mw := minOuterWidthForBorderPaint(child); mw > maxChildW {
					childW = mw
				} else {
					childW = maxChildW
				}
			}

			computeFlex(child, curX, curY, childW, childH, depth+1)
			applyRelativeOffset(child, item.style)

			if child.H > rowH {
				rowH = child.H
			}
			curX += childW
			if idx < len(r.items)-1 {
				curX += style.Gap
			}
		}

		curY += rowH
		if ri < len(rows)-1 {
			curY += rowGap
		}
	}

	// Outer H from computeFlex can be too small (e.g. intrinsic measure returned 1 while
	// rows use explicit child heights). Grow to the laid-out wrap span so parent vboxes
	// and borders match painted content (Lux Card + flex-wrap button rows).
	bw := 0
	if hasBorder(style) {
		bw = 1
	}
	pt, pb := style.PaddingTop, style.PaddingBottom
	if style.Padding > 0 {
		if pt == 0 {
			pt = style.Padding
		}
		if pb == 0 {
			pb = style.Padding
		}
	}
	innerUsed := curY - contentY
	if innerUsed < 1 {
		innerUsed = 1
	}
	wantOuter := innerUsed + 2*bw + pt + pb
	if wantOuter > node.H {
		node.H = wantOuter
		node.PaintDirty = true
		if node.Parent != nil {
			node.Parent.PaintDirty = true
		}
	}
}

// layoutVBoxWrap lays out children in a vertical column with wrapping.
// When children overflow the available height, they wrap to the next column.
