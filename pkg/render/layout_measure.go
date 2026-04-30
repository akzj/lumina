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

	// Apply explicit outer dimensions (override intrinsic)
	if outerW > 0 {
		measuredW = outerW
	}
	if outerH > 0 {
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
// and the parent constraints.
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
	// Default: use constraint width (fill parent)
	if c.WidthMode == SizeModeExact || c.WidthMode == SizeModeAtMost {
		return c.Width
	}
	return 0
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
func measureComponent(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	if len(node.Children) == 0 {
		return contentW + padW, 1 + padH
	}
	// Component placeholder is transparent — measure the grafted root
	if len(node.Children) == 1 {
		childC := Constraints{
			Width:      c.Width,
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
	totalH := 0
	flowCount := 0

	for _, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			measure(child, Constraints{Width: contentW, WidthMode: SizeModeExact, Height: 0, HeightMode: SizeModeUnbounded})
			continue
		}

		// Child constraints: full width, unbounded height (measure natural height)
		childC := Constraints{
			Width:     contentW,
			WidthMode: SizeModeExact,
		}
		// If parent has unbounded height, children also get unbounded
		if c.HeightMode == SizeModeUnbounded {
			childC.Height = 0
			childC.HeightMode = SizeModeUnbounded
		} else {
			childC.Height = 0
			childC.HeightMode = SizeModeUnbounded
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
func measureHBox(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	maxH := 0
	flowCount := 0
	totalFlexW := 0

	for _, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			measure(child, Constraints{Width: contentW, WidthMode: SizeModeAtMost, Height: 0, HeightMode: SizeModeUnbounded})
			continue
		}

		// For hbox children, measure with AtMost width to get natural width
		childC := Constraints{
			Width:      contentW,
			WidthMode:  SizeModeAtMost,
			Height:     0,
			HeightMode: SizeModeUnbounded,
		}
		_, childH := measure(child, childC)

		if childH > maxH {
			maxH = childH
		}
		totalFlexW += child.MeasuredW
		flowCount++
	}

	// Add gaps
	if flowCount > 1 {
		totalFlexW += node.Style.Gap * (flowCount - 1)
	}

	if maxH < 1 {
		maxH = 1
	}

	return contentW + padW, maxH + padH
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

// measureGrid measures a grid container.
func measureGrid(node *Node, c Constraints, contentW, padW, padH int) (int, int) {
	// Grid measurement: for now, use a simple heuristic.
	// The grid layout is complex; we approximate by measuring children
	// and using the grid template to determine row heights.
	// This will be refined in a later pass.
	totalH := 0
	flowCount := 0

	for _, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			continue
		}
		childC := Constraints{
			Width:      contentW,
			WidthMode:  SizeModeAtMost,
			Height:     0,
			HeightMode: SizeModeUnbounded,
		}
		measure(child, childC)
		if child.MeasuredH > totalH {
			totalH = child.MeasuredH
		}
		flowCount++
	}

	if totalH < 1 {
		totalH = 1
	}

	return contentW + padW, totalH + padH
}
