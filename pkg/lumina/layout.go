package lumina

// Style holds layout and visual properties for a VNode.
type Style struct {
	// Sizing
	Width     int // fixed width (0 = auto)
	Height    int // fixed height (0 = auto)
	MinWidth  int // minimum width
	MaxWidth  int // maximum width (0 = unlimited)
	MinHeight int
	MaxHeight int // maximum height (0 = unlimited)

	// Flex
	Flex int // flex grow factor (0 = no flex)

	// Spacing
	Padding       int // shorthand for all sides
	PaddingTop    int
	PaddingBottom int
	PaddingLeft   int
	PaddingRight  int
	Margin        int // shorthand for all sides
	MarginTop     int
	MarginBottom  int
	MarginLeft    int
	MarginRight   int
	Gap           int // space between children

	// Alignment (for container types)
	Justify string // "start" (default), "center", "end", "space-between", "space-around"
	Align   string // "stretch" (default), "start", "center", "end"

	// Visual
	Border     string // "none", "single", "double", "rounded"
	Foreground string
	Background string
	Bold       bool
	Dim        bool
	Underline  bool

	// Overflow
	Overflow string // "hidden" (default), "scroll" (future)
}

// getInt extracts an int from a map value, handling int64 (from go-lua ToAny)
// and float64 (from JSON).
func getInt(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return int(n)
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

// getString extracts a string from a map value with a default.
func getString(m map[string]any, key string, def string) string {
	v, ok := m[key]
	if !ok {
		return def
	}
	if s, ok := v.(string); ok {
		return s
	}
	return def
}

// getBool extracts a bool from a map value.
func getBool(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// parseStyle extracts a Style from VNode.Props.
// It checks the "style" sub-table first, then falls back to top-level props
// for backward compatibility.
func parseStyle(props map[string]any) Style {
	s := Style{Justify: "start", Align: "stretch"}

	// First, check for "style" sub-table
	if styleMap, ok := props["style"].(map[string]any); ok {
		s.Width = getInt(styleMap, "width")
		s.Height = getInt(styleMap, "height")
		s.MinWidth = getInt(styleMap, "minWidth")
		s.MaxWidth = getInt(styleMap, "maxWidth")
		s.MinHeight = getInt(styleMap, "minHeight")
		s.MaxHeight = getInt(styleMap, "maxHeight")
		s.Flex = getInt(styleMap, "flex")

		s.Padding = getInt(styleMap, "padding")
		s.PaddingTop = getInt(styleMap, "paddingTop")
		s.PaddingBottom = getInt(styleMap, "paddingBottom")
		s.PaddingLeft = getInt(styleMap, "paddingLeft")
		s.PaddingRight = getInt(styleMap, "paddingRight")
		s.Margin = getInt(styleMap, "margin")
		s.MarginTop = getInt(styleMap, "marginTop")
		s.MarginBottom = getInt(styleMap, "marginBottom")
		s.MarginLeft = getInt(styleMap, "marginLeft")
		s.MarginRight = getInt(styleMap, "marginRight")
		s.Gap = getInt(styleMap, "gap")

		s.Justify = getString(styleMap, "justify", "start")
		s.Align = getString(styleMap, "align", "stretch")

		s.Border = getString(styleMap, "border", "")
		s.Foreground = getString(styleMap, "foreground", "")
		s.Background = getString(styleMap, "background", "")
		s.Bold = getBool(styleMap, "bold")
		s.Dim = getBool(styleMap, "dim")
		s.Underline = getBool(styleMap, "underline")
		s.Overflow = getString(styleMap, "overflow", "hidden")
	}

	// Backward compat: check top-level props (lower priority — don't override style sub-table)
	if s.Flex == 0 {
		s.Flex = getInt(props, "flex")
	}
	if s.MinHeight == 0 {
		s.MinHeight = getInt(props, "minHeight")
	}
	if s.MinWidth == 0 {
		s.MinWidth = getInt(props, "minWidth")
	}
	if s.Border == "" {
		s.Border = getString(props, "border", "")
	}
	if s.Foreground == "" {
		s.Foreground = getString(props, "foreground", "")
	}
	if s.Background == "" {
		s.Background = getString(props, "background", "")
	}
	if !s.Bold {
		s.Bold = getBool(props, "bold")
	}
	if !s.Dim {
		s.Dim = getBool(props, "dim")
	}

	// Expand shorthand padding
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

	// Expand shorthand margin
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

	return s
}

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

// max returns the larger of a and b.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// computeFlexLayout computes the layout for a VNode tree using flexbox semantics.
// It replaces the old computeLayout function.
func computeFlexLayout(vnode *VNode, x, y, w, h int) {
	style := parseStyle(vnode.Props)
	vnode.Style = style

	// Apply margin — shrinks the area this node occupies from the parent's perspective
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
	if style.Border == "single" || style.Border == "double" || style.Border == "rounded" {
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
	scrollable := style.Overflow == "scroll"
	layoutW := contentW
	if scrollable && contentW > 1 {
		layoutW = contentW - 1 // reserve rightmost column for scrollbar
	}

	switch vnode.Type {
	case "fragment":
		// Fragment: transparent container — no border/padding/margin.
		// Layout each child using computeFlexLayout directly.
		// The parent's layout function handles positioning.
		layoutVBox(vnode, x, y, w, h, style)
		return // skip the rest of computeFlexLayout

	case "text":
		layoutText(vnode, layoutW)

	case "vbox":
		if scrollable {
			layoutScrollableVBox(vnode, contentX, contentY, layoutW, contentH, style)
		} else {
			layoutVBox(vnode, contentX, contentY, layoutW, contentH, style)
		}

	case "hbox":
		if scrollable {
			layoutScrollableHBox(vnode, contentX, contentY, layoutW, contentH, style)
		} else {
			layoutHBox(vnode, contentX, contentY, layoutW, contentH, style)
		}

	default:
		// Generic container (box, etc.) — stack children vertically like vbox
		if scrollable {
			layoutScrollableVBox(vnode, contentX, contentY, layoutW, contentH, style)
		} else {
			layoutVBox(vnode, contentX, contentY, layoutW, contentH, style)
		}
	}
}

// layoutText measures a text node. It wraps text if it exceeds the available width.
func layoutText(vnode *VNode, availW int) {
	if availW <= 0 {
		vnode.H = 1
		return
	}
	contentLen := len(vnode.Content)
	if contentLen == 0 {
		vnode.H = 1
		vnode.W = maxInt(vnode.W, 0)
		return
	}
	// Calculate wrapped height
	lines := (contentLen + availW - 1) / availW // ceiling division
	if lines < 1 {
		lines = 1
	}
	// Text node width is min(contentLen, availW)
	if contentLen < availW {
		// Don't shrink the node below content width if parent hasn't constrained it
		// but also don't exceed the allocated width
	}
	vnode.H = lines
}

// layoutVBox lays out children in a vertical stack with flex distribution.
func layoutVBox(vnode *VNode, contentX, contentY, contentW, contentH int, style Style) {
	if len(vnode.Children) == 0 {
		return
	}

	// Parse child styles
	type childInfo struct {
		style    Style
		fixedH   int  // resolved fixed height (0 = flex)
		flexGrow int  // flex factor
		finalH   int  // computed height after distribution
	}
	children := make([]childInfo, len(vnode.Children))

	totalGaps := style.Gap * (len(vnode.Children) - 1)
	availH := contentH - totalGaps
	if availH < 0 {
		availH = 0
	}

	fixedTotal := 0
	flexTotal := 0

	for i, child := range vnode.Children {
		cs := parseStyle(child.Props)
		children[i].style = cs

		marginV := cs.MarginTop + cs.MarginBottom

		// Fragment: natural height = number of children (each at least 1 row)
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
			// No flex, no fixed height — give minimum 1 row
			children[i].fixedH = 1 + marginV
			fixedTotal += children[i].fixedH
		}
	}

	// Distribute remaining space to flex children
	remainH := availH - fixedTotal
	if remainH < 0 {
		remainH = 0
	}

	for i := range children {
		if children[i].flexGrow > 0 {
			if flexTotal > 0 {
				children[i].finalH = (remainH * children[i].flexGrow) / flexTotal
			}
			if children[i].finalH < 1 {
				children[i].finalH = 1
			}
			// Apply min/max constraints
			children[i].finalH = clamp(children[i].finalH, children[i].style.MinHeight, children[i].style.MaxHeight)
		} else {
			children[i].finalH = children[i].fixedH
		}
	}

	// Calculate total used height for justify alignment
	totalUsed := 0
	for i := range children {
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
		if len(vnode.Children) > 1 {
			gapSize = 0 // we'll distribute all extra space as gaps
			totalBetween := len(vnode.Children) - 1
			usedNoGap := totalUsed - totalGaps
			spaceBetween := contentH - usedNoGap
			if spaceBetween > 0 && totalBetween > 0 {
				gapSize = spaceBetween / totalBetween
			}
		}
	case "space-around":
		if len(vnode.Children) > 0 {
			gapSize = 0
			usedNoGap := totalUsed - totalGaps
			totalSlots := len(vnode.Children)*2
			spaceAround := contentH - usedNoGap
			if spaceAround > 0 && totalSlots > 0 {
				halfGap := spaceAround / totalSlots
				curY += halfGap
				gapSize = halfGap * 2
			}
		}
	}

	// Position each child
	for i, child := range vnode.Children {
		childH := children[i].finalH
		childW := contentW

		// Cross-axis alignment (align)
		childX := contentX
		switch style.Align {
		case "center":
			// Center child in cross axis — but we need child's preferred width
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
		curY += childH
		if i < len(vnode.Children)-1 {
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
		style    Style
		fixedW   int
		flexGrow int
		finalW   int
	}
	children := make([]childInfo, len(vnode.Children))

	totalGaps := style.Gap * (len(vnode.Children) - 1)
	availW := contentW - totalGaps
	if availW < 0 {
		availW = 0
	}

	fixedTotal := 0
	flexTotal := 0

	for i, child := range vnode.Children {
		cs := parseStyle(child.Props)
		children[i].style = cs

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
			// No flex, no fixed width — give minimum 1 col
			children[i].fixedW = 1 + marginH
			fixedTotal += children[i].fixedW
		}
	}

	remainW := availW - fixedTotal
	if remainW < 0 {
		remainW = 0
	}

	for i := range children {
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

	// Calculate total used width for justify alignment
	totalUsed := 0
	for i := range children {
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
		if len(vnode.Children) > 1 {
			gapSize = 0
			totalBetween := len(vnode.Children) - 1
			usedNoGap := totalUsed - totalGaps
			spaceBetween := contentW - usedNoGap
			if spaceBetween > 0 && totalBetween > 0 {
				gapSize = spaceBetween / totalBetween
			}
		}
	case "space-around":
		if len(vnode.Children) > 0 {
			gapSize = 0
			usedNoGap := totalUsed - totalGaps
			totalSlots := len(vnode.Children) * 2
			spaceAround := contentW - usedNoGap
			if spaceAround > 0 && totalSlots > 0 {
				halfGap := spaceAround / totalSlots
				curX += halfGap
				gapSize = halfGap * 2
			}
		}
	}

	for i, child := range vnode.Children {
		childW := children[i].finalW
		childH := contentH

		// Cross-axis alignment (align) — vertical for hbox
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

		computeFlexLayout(child, curX, childY, childW, childH)
		curX += childW
		if i < len(vnode.Children)-1 {
			curX += gapSize
		}
	}
}

// layoutScrollableVBox lays out children in a vertical stack without clamping
// to the container height. Children get their natural size, and a Viewport
// is created/updated to track the scroll state.
func layoutScrollableVBox(vnode *VNode, contentX, contentY, contentW, contentH int, style Style) {
	if len(vnode.Children) == 0 {
		return
	}

	// Get VNode ID for viewport registry
	nodeID, _ := vnode.Props["id"].(string)

	// Calculate total natural height of all children
	type childInfo struct {
		style  Style
		height int // natural height
	}
	children := make([]childInfo, len(vnode.Children))

	totalGaps := style.Gap * (len(vnode.Children) - 1)
	totalH := totalGaps

	for i, child := range vnode.Children {
		cs := parseStyle(child.Props)
		children[i].style = cs

		marginV := cs.MarginTop + cs.MarginBottom
		var h int
		if cs.Height > 0 {
			h = clamp(cs.Height, cs.MinHeight, cs.MaxHeight) + marginV
		} else if cs.MinHeight > 0 {
			h = cs.MinHeight + marginV
		} else {
			// Default: 1 row for text, or estimate for containers
			h = 1 + marginV
			if child.Type != "text" && len(child.Children) > 0 {
				// Estimate based on number of children
				h = len(child.Children) + marginV
			}
		}
		children[i].height = h
		totalH += h
	}

	// Get or create viewport
	vp := &Viewport{
		ContentH: totalH,
		ContentW: contentW,
		ViewW:    contentW,
		ViewH:    contentH,
	}

	if nodeID != "" {
		existing := GetViewport(nodeID)
		// Preserve scroll position, update dimensions
		vp.ScrollX = existing.ScrollX
		vp.ScrollY = existing.ScrollY
		vp.ContentH = totalH
		vp.ContentW = contentW
		vp.ViewW = contentW
		vp.ViewH = contentH
		vp.clampScroll()
		SetViewport(nodeID, vp)
	}

	// Position children at their natural positions (virtual coordinates)
	// The renderer will apply scroll offset and clipping
	curY := contentY - vp.ScrollY
	for i, child := range vnode.Children {
		childH := children[i].height
		computeFlexLayout(child, contentX, curY, contentW, childH)
		curY += childH
		if i < len(vnode.Children)-1 {
			curY += style.Gap
		}
	}
}

// layoutScrollableHBox lays out children horizontally without clamping
// to the container width. Similar to layoutScrollableVBox but on the X axis.
func layoutScrollableHBox(vnode *VNode, contentX, contentY, contentW, contentH int, style Style) {
	if len(vnode.Children) == 0 {
		return
	}

	nodeID, _ := vnode.Props["id"].(string)

	type childInfo struct {
		style Style
		width int
	}
	children := make([]childInfo, len(vnode.Children))

	totalGaps := style.Gap * (len(vnode.Children) - 1)
	totalW := totalGaps

	for i, child := range vnode.Children {
		cs := parseStyle(child.Props)
		children[i].style = cs

		marginH := cs.MarginLeft + cs.MarginRight
		var w int
		if cs.Width > 0 {
			w = clamp(cs.Width, cs.MinWidth, cs.MaxWidth) + marginH
		} else if cs.MinWidth > 0 {
			w = cs.MinWidth + marginH
		} else {
			w = 1 + marginH
			if child.Type == "text" {
				w = len(child.Content) + marginH
			}
		}
		children[i].width = w
		totalW += w
	}

	vp := &Viewport{
		ContentW: totalW,
		ContentH: contentH,
		ViewW:    contentW,
		ViewH:    contentH,
	}

	if nodeID != "" {
		existing := GetViewport(nodeID)
		vp.ScrollX = existing.ScrollX
		vp.ScrollY = existing.ScrollY
		vp.ContentW = totalW
		vp.ContentH = contentH
		vp.ViewW = contentW
		vp.ViewH = contentH
		vp.clampScroll()
		SetViewport(nodeID, vp)
	}

	curX := contentX - vp.ScrollX
	for i, child := range vnode.Children {
		childW := children[i].width
		computeFlexLayout(child, curX, contentY, childW, contentH)
		curX += childW
		if i < len(vnode.Children)-1 {
			curX += style.Gap
		}
	}
}
