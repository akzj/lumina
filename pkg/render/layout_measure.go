package render

// --- Two-Phase Layout: Measurement Pass ---
//
// The measure pass walks the tree bottom-up, computing the "natural" (intrinsic)
// size of each node given parent constraints. Results are cached in
// node.MeasuredW / node.MeasuredH and consumed by the layout (positioning) pass.
//
// This eliminates the old pattern of calling computeFlex(child, x, y, w, 99999, depth+1)
// to probe natural heights, which created cascading special cases.

// SizeMode describes how a dimension constraint should be interpreted.
type SizeMode int

const (
	// SizeModeExact means the parent dictates the exact size.
	SizeModeExact SizeMode = iota
	// SizeModeAtMost means the child can be at most this size (shrink-to-fit).
	SizeModeAtMost
	// SizeModeUnbounded means there is no constraint (e.g. scroll container's cross axis).
	SizeModeUnbounded
)

// Constraints describes the space available to a node during measurement.
type Constraints struct {
	Width      int
	Height     int
	WidthMode  SizeMode
	HeightMode SizeMode
}

// measure computes the intrinsic size of a node given constraints.
// It stores results in node.MeasuredW and node.MeasuredH.
// It does NOT write X/Y positions.
func measure(node *Node, c Constraints) (int, int) {
	if node == nil {
		return 0, 0
	}

	style := node.Style

	// display:none — takes no space
	if style.Display == "none" {
		node.MeasuredW = 0
		node.MeasuredH = 0
		return 0, 0
	}

	// Resolve the outer dimensions from explicit style values
	borderW := 0
	if hasBorder(style) {
		borderW = 1
	}
	padH := style.PaddingTop + style.PaddingBottom + 2*borderW
	padW := style.PaddingLeft + style.PaddingRight + 2*borderW

	// Determine available content width for children
	outerW := resolveConstrainedWidth(style, c)
	contentW := outerW - padW
	if contentW < 0 {
		contentW = 0
	}

	// Determine explicit height if any
	outerH := resolveConstrainedHeight(style, c)

	var measuredW, measuredH int

	switch node.Type {
	case "text":
		measuredW, measuredH = measureText(node, contentW, padW, padH)
	case "input":
		measuredW = outerW
		measuredH = 1 + padH
	case "textarea":
		// textarea: use explicit height or default to 3 rows
		if outerH > 0 {
			measuredH = outerH
		} else {
			measuredH = 3 + padH
		}
		measuredW = outerW
	case "component":
		measuredW, measuredH = measureComponent(node, c, contentW, padW, padH)
	case "fragment":
		measuredW, measuredH = measureFragment(node, c, contentW, padW, padH)
	default:
		// "box", "vbox", "hbox" and any other container
		measuredW, measuredH = measureContainer(node, c, contentW, padW, padH)
	}

	// Apply explicit outer dimensions (override intrinsic).
	// Only override if the node has an explicit size set in style.
	// For AtMost mode without explicit width, the node should shrink to content.
	if hasExplicitWidth(style) && outerW > 0 {
		measuredW = outerW
	} else if c.WidthMode == SizeModeExact && outerW > 0 {
		// Exact mode: node fills parent (standard block-level behavior)
		measuredW = outerW
	}
	if hasExplicitHeight(style) && outerH > 0 {
		measuredH = outerH
	}

	// Apply min/max constraints
	minW := resolveMinW(style, c.Width)
	maxW := resolveMaxW(style, c.Width)
	minH := resolveMinH(style, c.Height)
	maxH := resolveMaxH(style, c.Height)

	measuredW = clamp(measuredW, minW, maxW)
	measuredH = clamp(measuredH, minH, maxH)

	// Floor at 0
	if measuredW < 0 {
		measuredW = 0
	}
	if measuredH < 0 {
		measuredH = 0
	}

	// Include margin in measured size (margin is part of the space this node "takes")
	marginW := style.MarginLeft + style.MarginRight
	marginH := style.MarginTop + style.MarginBottom

	node.MeasuredW = measuredW + marginW
	node.MeasuredH = measuredH + marginH

	return node.MeasuredW, node.MeasuredH
}

// resolveConstrainedWidth determines the outer width of a node from its style
// and the parent constraints. Returns (width, explicit) where explicit is true
// if the width came from a style property (not from defaulting to constraint).
func resolveConstrainedWidth(style Style, c Constraints) int {
	// Explicit absolute width
	if style.Width > 0 {
		return style.Width
	}
	// Percentage width
	if style.WidthPercent > 0 && c.WidthMode != SizeModeUnbounded {
		return (c.Width * style.WidthPercent) / 100
	}
	// Viewport width
	if style.WidthVW > 0 {
		return (layoutViewportW * style.WidthVW) / 100
	}
	// Default: use constraint width (fill parent) — but only for Exact mode.
	// For AtMost mode, the node should shrink to content unless explicitly sized.
	if c.WidthMode == SizeModeExact {
		return c.Width
	}
	if c.WidthMode == SizeModeAtMost {
		return c.Width // still use as upper bound for content measurement
	}
	return 0
}

// hasExplicitWidth returns true if the node has an explicit width set in style.
func hasExplicitWidth(style Style) bool {
	return style.Width > 0 || style.WidthPercent > 0 || style.WidthVW > 0
}

// hasExplicitHeight returns true if the node has an explicit height set in style.
func hasExplicitHeight(style Style) bool {
	return style.Height > 0 || style.HeightPercent > 0 || style.HeightVH > 0
}

// resolveConstrainedHeight determines the outer height of a node from its style
// and the parent constraints.
func resolveConstrainedHeight(style Style, c Constraints) int {
	// Explicit absolute height
	if style.Height > 0 {
		return style.Height
	}
	// Percentage height — only resolve if parent height is known (Exact mode)
	if style.HeightPercent > 0 && c.HeightMode == SizeModeExact {
		return (c.Height * style.HeightPercent) / 100
	}
	// Viewport height
	if style.HeightVH > 0 {
		return (layoutViewportH * style.HeightVH) / 100
	}
	// No explicit height — will be determined by content
	return 0
}

// measureText computes the intrinsic size of a text node.
func measureText(node *Node, contentW, padW, padH int) (int, int) {
	if node.Content == "" {
		return contentW + padW, 1 + padH
	}

	// If whiteSpace=nowrap, text is one line
	if node.Style.WhiteSpace == "nowrap" {
		textW := stringWidth(node.Content)
		return textW + padW, 1 + padH
	}

	// Wrap text to contentW
	if contentW <= 0 {
		contentW = 1
	}
	lines := wrapTextLines(node.Content, contentW)
	h := len(lines)
	if h < 1 {
		h = 1
	}
	return contentW + padW, h + padH
}

// wrapTextLines simulates text wrapping and returns the number of lines.
func wrapTextLines(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 1
	}
	var lines []string
	currentWidth := 0
	lineStart := 0
	runes := []rune(text)

	for i, r := range runes {
		if r == '\n' {
			lines = append(lines, string(runes[lineStart:i]))
			lineStart = i + 1
			currentWidth = 0
			continue
		}
		rw := runeWidth(r)
		if currentWidth+rw > maxWidth && currentWidth > 0 {
			lines = append(lines, string(runes[lineStart:i]))
			lineStart = i
			currentWidth = rw
		} else {
			currentWidth += rw
		}
	}
	// Last line
	if lineStart <= len(runes) {
		lines = append(lines, string(runes[lineStart:]))
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

// measureComponent measures a component placeholder by measuring its grafted children.
// Component placeholders are transparent — they pass the available content area
// (after subtracting their own border/padding) to the grafted root.
func measureComponent(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	if len(node.Children) == 0 {
		return contentW + padW, 1 + padH
	}
	// Component placeholder is transparent — measure the grafted root
	// Use contentW (accounts for border/padding of the component node itself)
	if len(node.Children) == 1 {
		childC := Constraints{
			Width:      contentW + padW, // outer width available to the child
			WidthMode:  c.WidthMode,
			Height:     c.Height,
			HeightMode: c.HeightMode,
		}
		w, h := measure(node.Children[0], childC)
		return w, h
	}
	// Multiple children (rare): treat as vbox
	return measureVBox(node, c, contentW, padW, padH)
}

// measureFragment measures a fragment (transparent container).
func measureFragment(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	if len(node.Children) == 0 {
		return contentW + padW, 1 + padH
	}
	// Fragment acts like vbox
	return measureVBox(node, c, contentW, padW, padH)
}

// measureContainer dispatches to the appropriate container measurement.
func measureContainer(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	if len(node.Children) == 0 {
		// Leaf container: min height 1
		return contentW + padW, 1 + padH
	}

	style := node.Style

	// Grid layout
	if style.Display == "grid" {
		return measureGrid(node, c, contentW, padW, padH)
	}

	// Determine direction
	isHBox := node.Type == "hbox"
	isWrap := style.FlexWrap == "wrap" || style.FlexWrap == "wrap-reverse"

	if isHBox && isWrap {
		return measureHBoxWrap(node, c, contentW, padW, padH)
	}
	if isHBox {
		return measureHBox(node, c, contentW, padW, padH)
	}
	if isWrap {
		return measureVBoxWrap(node, c, contentW, padW, padH)
	}
	// Default: vbox
	return measureVBox(node, c, contentW, padW, padH)
}

// measureVBox measures children stacked vertically.
func measureVBox(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	// For scroll containers, reserve 1 column for scrollbar
	layoutW := contentW
	if node.Style.Overflow == "scroll" && contentW > 1 {
		layoutW = contentW - 1
	}

	totalH := 0
	flowCount := 0

	for _, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			measure(child, Constraints{Width: layoutW, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
			continue
		}

		// Child constraints: full width, unbounded height (measure natural height)
		childC := Constraints{
			Width:      layoutW,
			WidthMode:  SizeModeExact,
			Height:     0,
			HeightMode: SizeModeUnbounded,
		}

		_, childH := measure(child, childC)
		totalH += childH
		flowCount++
	}

	// Add gaps
	if flowCount > 1 {
		totalH += node.Style.Gap * (flowCount - 1)
	}

	return contentW + padW, totalH + padH
}

// measureHBox measures children laid out horizontally.
// Mirrors layoutHBox's flex distribution to give each child its correct width.
func measureHBox(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	if contentW <= 0 {
		contentW = 1
	}

	type childMInfo struct {
		style      Style
		fixedW     int
		flexGrow   int
		finalW     int
		positioned bool
	}
	children := make([]childMInfo, len(node.Children))

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
		totalGaps = node.Style.Gap * (flowCount - 1)
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
			baseW := children[i].fixedW
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

	// Now measure each child with its allocated width to get correct height
	maxH := 0
	usedW := 0
	for i, child := range node.Children {
		if children[i].positioned {
			measure(child, Constraints{Width: contentW, WidthMode: SizeModeAtMost, Height: 0, HeightMode: SizeModeUnbounded})
			continue
		}
		childC := Constraints{
			Width:      children[i].finalW,
			WidthMode:  SizeModeExact,
			Height:     0,
			HeightMode: SizeModeUnbounded,
		}
		_, childH := measure(child, childC)
		if childH > maxH {
			maxH = childH
		}
		usedW += children[i].finalW
	}

	// Add gaps to used width
	if flowCount > 1 {
		usedW += node.Style.Gap * (flowCount - 1)
	}

	if maxH < 1 {
		maxH = 1
	}

	// For AtMost mode, shrink to content width; for Exact mode, fill parent
	returnW := contentW + padW
	if c.WidthMode == SizeModeAtMost && usedW+padW < returnW {
		returnW = usedW + padW
	}
	return returnW, maxH + padH
}

// measureHBoxWrap measures children in a wrapping horizontal layout.
func measureHBoxWrap(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	if contentW <= 0 {
		contentW = 1
	}

	// Measure each child's natural size
	type childMeasure struct {
		w, h int
	}
	var items []childMeasure

	for _, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			measure(child, Constraints{Width: contentW, WidthMode: SizeModeAtMost, Height: 0, HeightMode: SizeModeUnbounded})
			continue
		}
		childC := Constraints{
			Width:      contentW,
			WidthMode:  SizeModeAtMost,
			Height:     0,
			HeightMode: SizeModeUnbounded,
		}
		measure(child, childC)
		items = append(items, childMeasure{w: child.MeasuredW, h: child.MeasuredH})
	}

	// Simulate wrapping
	gap := node.Style.Gap
	rowH := 0
	totalH := 0
	rowW := 0
	rowCount := 0

	for _, item := range items {
		itemW := item.w
		if rowW > 0 && rowW+gap+itemW > contentW {
			// Wrap to new row
			totalH += rowH
			if rowCount > 0 {
				totalH += gap
			}
			rowCount++
			rowW = itemW
			rowH = item.h
		} else {
			if rowW > 0 {
				rowW += gap
			}
			rowW += itemW
			if item.h > rowH {
				rowH = item.h
			}
		}
	}
	// Last row
	if rowH > 0 {
		if rowCount > 0 {
			totalH += gap
		}
		totalH += rowH
	}

	if totalH < 1 {
		totalH = 1
	}

	return contentW + padW, totalH + padH
}

// measureVBoxWrap measures children in a wrapping vertical layout.
func measureVBoxWrap(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	// For vbox wrap, children wrap into columns when they exceed height.
	// Since height is typically unbounded during measure, just sum all heights.
	totalH := 0
	flowCount := 0

	for _, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			measure(child, Constraints{Width: contentW, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
			continue
		}
		childC := Constraints{
			Width:      contentW,
			WidthMode:  SizeModeExact,
			Height:     0,
			HeightMode: SizeModeUnbounded,
		}
		_, childH := measure(child, childC)
		totalH += childH
		flowCount++
	}

	if flowCount > 1 {
		totalH += node.Style.Gap * (flowCount - 1)
	}
	if totalH < 1 {
		totalH = 1
	}

	return contentW + padW, totalH + padH
}

// measureGrid measures a grid container by simulating auto-placement into rows.
func measureGrid(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	style := node.Style

	// Determine column count from grid template
	colGap := style.GridColumnGap
	if colGap == 0 {
		colGap = style.Gap
	}
	rowGap := style.GridRowGap
	if rowGap == 0 {
		rowGap = style.Gap
	}

	cols := parseGridTemplate(style.GridTemplateColumns, contentW, colGap)
	numCols := len(cols)
	if numCols == 0 {
		numCols = 1
		cols = []int{contentW}
	}

	// Measure all flow children
	var flowChildren []*Node
	for _, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			continue
		}
		// Approximate column width for child measurement
		colW := cols[len(flowChildren)%numCols]
		childC := Constraints{
			Width:      colW,
			WidthMode:  SizeModeAtMost,
			Height:     0,
			HeightMode: SizeModeUnbounded,
		}
		measure(child, childC)
		flowChildren = append(flowChildren, child)
	}

	// Simulate auto-placement to determine row heights
	numRows := (len(flowChildren) + numCols - 1) / numCols
	if numRows < 1 {
		numRows = 1
	}

	totalH := 0
	for row := 0; row < numRows; row++ {
		maxRowH := 0
		for col := 0; col < numCols; col++ {
			idx := row*numCols + col
			if idx >= len(flowChildren) {
				break
			}
			if flowChildren[idx].MeasuredH > maxRowH {
				maxRowH = flowChildren[idx].MeasuredH
			}
		}
		totalH += maxRowH
		if row < numRows-1 {
			totalH += rowGap
		}
	}

	if totalH < 1 {
		totalH = 1
	}

	return contentW + padW, totalH + padH
}
