package render

import "sort"

func layoutVBox(node *Node, contentX, contentY, contentW, contentH int, style Style, depth int) {
	if len(node.Children) == 0 {
		return
	}

	// flex-wrap: delegate to wrap layout
	if style.FlexWrap == "wrap" || style.FlexWrap == "wrap-reverse" {
		layoutVBoxWrap(node, contentX, contentY, contentW, contentH, style, depth)
		return
	}

	isScroll := style.Overflow == "scroll"
	// Inner content vbox of a scroll container must stack children at natural heights
	// rather than flex-distributing the scroll parent's real height across them.
	scrollContentStack := !isScroll && nodeHasScrollAncestor(node)

	useNaturalHeights := isScroll || scrollContentStack

	type childInfo struct {
		style      Style
		fixedH     int
		flexGrow   int
		finalH     int
		positioned bool
	}
	children := make([]childInfo, len(node.Children))

	// Count flow children (not absolute/fixed, not display:none)
	flowCount := 0
	for i, child := range node.Children {
		cs := child.Style
		children[i].style = cs
		children[i].positioned = isPositioned(cs) || cs.Display == "none"
		if !children[i].positioned {
			flowCount++
		}
	}

	totalGaps := 0
	if flowCount > 1 {
		totalGaps = style.Gap * (flowCount - 1)
	}

	if useNaturalHeights {
		// Scroll slot, intrinsic probe, or scrollable main's inner column: children get
		// natural heights, no flex distribution.
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
			} else if rh := resolveHeight(cs, contentH); rh > 0 {
				children[i].finalH = clamp(rh, cs.MinHeight, cs.MaxHeight) + marginV
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
					// Use pre-computed measurement from measure pass.
					children[i].finalH = child.MeasuredH
					if children[i].finalH < 1+marginV {
						children[i].finalH = 1 + marginV
					}
				}
				if mnh := resolveMinH(cs, contentH); mnh > 0 && children[i].finalH < mnh+marginV {
					children[i].finalH = mnh + marginV
				}
			} else if len(child.Children) > 0 {
				// Use pre-computed measurement from measure pass.
				children[i].finalH = child.MeasuredH
				if children[i].finalH < 1+marginV {
					children[i].finalH = 1 + marginV
				}
				if mnh := resolveMinH(cs, contentH); mnh > 0 && children[i].finalH < mnh+marginV {
					children[i].finalH = mnh + marginV
				}
			} else if cs.MinHeight > 0 {
				children[i].finalH = cs.MinHeight + marginV
			} else if mnh := resolveMinH(cs, contentH); mnh > 0 {
				children[i].finalH = mnh + marginV
			} else {
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
			} else if rh := resolveHeight(cs, contentH); rh > 0 {
				h := clamp(rh, cs.MinHeight, cs.MaxHeight)
				children[i].fixedH = h + marginV
				fixedTotal += children[i].fixedH
			} else if cs.FlexBasis > 0 && cs.Flex > 0 {
				// flexBasis: use as initial main size, still participate in flex-grow
				children[i].fixedH = cs.FlexBasis + marginV
				fixedTotal += children[i].fixedH
				children[i].flexGrow = cs.Flex
				flexTotal += cs.Flex
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
				baseH := children[i].fixedH // flexBasis base (0 if no basis)
				if flexTotal > 0 {
					children[i].finalH = baseH + (remainH*children[i].flexGrow)/flexTotal
				} else {
					children[i].finalH = baseH
				}
				if children[i].finalH < 1 {
					children[i].finalH = 1
				}
				children[i].finalH = clamp(children[i].finalH, children[i].style.MinHeight, children[i].style.MaxHeight)
			} else {
				children[i].finalH = children[i].fixedH
			}
		}

		// flexShrink: if total sizes exceed available space, shrink proportionally
		totalAfterGrow := 0
		for i := range children {
			if !children[i].positioned {
				totalAfterGrow += children[i].finalH
			}
		}
		overflow := totalAfterGrow - availH
		if overflow > 0 {
			shrinkTotal := 0
			for i := range children {
				if !children[i].positioned && children[i].style.FlexShrink > 0 {
					shrinkTotal += children[i].style.FlexShrink
				}
			}
			if shrinkTotal > 0 {
				for i := range children {
					if !children[i].positioned && children[i].style.FlexShrink > 0 {
						shrink := (overflow * children[i].style.FlexShrink) / shrinkTotal
						children[i].finalH -= shrink
						if children[i].finalH < 1 {
							children[i].finalH = 1
						}
						children[i].finalH = clamp(children[i].finalH, children[i].style.MinHeight, children[i].style.MaxHeight)
					}
				}
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

	// Determine starting Y based on justify (skip for scroll / intrinsic / scroll-content stack)
	curY := contentY
	gapSize := style.Gap
	if !isScroll && !scrollContentStack {
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

	// Build order-sorted index for positioning
	orderIndices := make([]int, 0, len(node.Children))
	for i := range node.Children {
		if !children[i].positioned {
			orderIndices = append(orderIndices, i)
		}
	}
	sort.SliceStable(orderIndices, func(a, b int) bool {
		return children[orderIndices[a]].style.Order < children[orderIndices[b]].style.Order
	})

	// Position each child in order-sorted sequence
	flowIdx := 0
	var lastFlowChildNode *Node
	for _, i := range orderIndices {
		child := node.Children[i]

		childH := children[i].finalH
		childW := contentW
		childX := contentX

		// Cross-axis: resolve child's effective width (percent/viewport/absolute)
		cs := children[i].style
		rw := resolveWidth(cs, contentW)
		cMinW := resolveMinW(cs, contentW)
		cMaxW := resolveMaxW(cs, contentW)
		if rw > 0 {
			rw = clamp(rw, cMinW, cMaxW)
			childW = rw
		} else if cMinW > 0 || cMaxW > 0 {
			childW = clamp(contentW, cMinW, cMaxW)
		}

		// Cross-axis alignment: check alignSelf first, fall back to parent's align
		align := style.Align
		if cs.AlignSelf != "" {
			align = cs.AlignSelf
		}
		if childW < contentW {
			switch align {
			case "center":
				childX = contentX + (contentW-childW)/2
			case "end":
				childX = contentX + contentW - childW
			}
		}

		computeFlex(child, childX, curY, childW, childH, depth+1)
		applyRelativeOffset(child, children[i].style)
		lastFlowChildNode = child
		curY += child.H
		flowIdx++
		if flowIdx < flowCount {
			curY += gapSize
		}
	}

	// Scrollable main columns: grow outer height if flow children extend past the slot
	// (e.g. flex-wrap hbox grew node.H after layout; this node still had the old H).
	if scrollContentStack {
		bw := 0
		if hasBorder(style) {
			bw = 1
		}
		pb := style.PaddingBottom
		if style.Padding > 0 && pb == 0 {
			pb = style.Padding
		}
		maxBottom := contentY
		for _, ch := range node.Children {
			cs := ch.Style
			if isPositioned(cs) || cs.Display == "none" {
				continue
			}
			b := ch.Y + ch.H
			if b > maxBottom {
				maxBottom = b
			}
		}
		minOuterH := (maxBottom + pb + bw) - node.Y
		if minOuterH > node.H {
			node.H = minOuterH
			node.PaintDirty = true
			if node.Parent != nil {
				node.Parent.PaintDirty = true
			}
		}
	}

	// For scroll containers, store the total content height
	if isScroll && lastFlowChildNode != nil {
		node.ScrollHeight = (lastFlowChildNode.Y + lastFlowChildNode.H) - contentY
	}
}

// layoutHBox lays out children in a horizontal row with flex distribution.

func layoutVBoxWrap(node *Node, contentX, contentY, contentW, contentH int, style Style, depth int) {
	type colItem struct {
		childIdx int
		desiredH int
		style    Style
	}

	// Collect flow children and their desired heights
	var allItems []colItem
	for i, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			continue
		}
		marginV := cs.MarginTop + cs.MarginBottom

		desiredH := 0
		if rh := resolveHeight(cs, contentH); rh > 0 {
			desiredH = clamp(rh, resolveMinH(cs, contentH), resolveMaxH(cs, contentH)) + marginV
		} else if cs.FlexBasis > 0 {
			desiredH = cs.FlexBasis + marginV
		} else {
			// Auto height
			switch child.Type {
			case "text", "input", "textarea":
				desiredH = 1 + marginV
			default:
				if cs.Flex > 0 {
					desiredH = 1 + marginV
				} else {
					desiredH = 1 + marginV
				}
			}
		}
		allItems = append(allItems, colItem{childIdx: i, desiredH: desiredH, style: cs})
	}

	if len(allItems) == 0 {
		return
	}

	// Build columns greedily
	type col struct {
		items []colItem
	}
	var cols []col
	var currentCol []colItem
	currentColH := 0

	for _, item := range allItems {
		gap := 0
		if len(currentCol) > 0 {
			gap = style.Gap
		}
		if currentColH+gap+item.desiredH > contentH && len(currentCol) > 0 {
			cols = append(cols, col{items: currentCol})
			currentCol = nil
			currentColH = 0
			gap = 0
		}
		currentCol = append(currentCol, item)
		currentColH += gap + item.desiredH
	}
	if len(currentCol) > 0 {
		cols = append(cols, col{items: currentCol})
	}

	// For wrap-reverse, reverse the column order
	if style.FlexWrap == "wrap-reverse" {
		for i, j := 0, len(cols)-1; i < j; i, j = i+1, j-1 {
			cols[i], cols[j] = cols[j], cols[i]
		}
	}

	// Layout each column
	colGap := style.Gap
	curX := contentX

	for ci, c := range cols {
		// Determine column height usage and flex distribution
		colFlexTotal := 0
		colFixedTotal := 0
		for _, item := range c.items {
			if item.style.Flex > 0 && resolveHeight(item.style, contentH) == 0 && item.style.FlexBasis == 0 {
				colFlexTotal += item.style.Flex
			}
			colFixedTotal += item.desiredH
		}
		colGaps := 0
		if len(c.items) > 1 {
			colGaps = style.Gap * (len(c.items) - 1)
		}
		remainH := contentH - colFixedTotal - colGaps
		if remainH < 0 {
			remainH = 0
		}

		// Compute final heights
		finalHeights := make([]int, len(c.items))
		for idx, item := range c.items {
			if item.style.Flex > 0 && resolveHeight(item.style, contentH) == 0 && item.style.FlexBasis == 0 && colFlexTotal > 0 {
				finalHeights[idx] = item.desiredH + (remainH*item.style.Flex)/colFlexTotal
			} else {
				finalHeights[idx] = item.desiredH
			}
			if finalHeights[idx] < 1 {
				finalHeights[idx] = 1
			}
		}

		// Position children in this column
		curY := contentY
		colW := 1

		for idx, item := range c.items {
			child := node.Children[item.childIdx]
			childH := finalHeights[idx]
			childW := contentW
			if rw := resolveWidth(item.style, contentW); rw > 0 {
				childW = rw
			}

			computeFlex(child, curX, curY, childW, childH, depth+1)
			applyRelativeOffset(child, item.style)

			if child.W > colW {
				colW = child.W
			}
			curY += childH
			if idx < len(c.items)-1 {
				curY += style.Gap
			}
		}

		curX += colW
		if ci < len(cols)-1 {
			curX += colGap
		}
	}
}

