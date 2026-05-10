package render

// paint_helpers.go — shared paint utility functions used by V3 paint engine and engine.go.

// clearPaintDirty recursively clears PaintDirty flags on all nodes.
func clearPaintDirty(node *Node) {
	if node == nil {
		return
	}
	node.PaintDirty = false
	node.PositionChanged = false
	for _, child := range node.Children {
		clearPaintDirty(child)
	}
}

// clearPaintDirtyBelow clears PaintDirty on all descendants (not the node itself).
func clearPaintDirtyBelow(node *Node) {
	for _, child := range node.Children {
		child.PaintDirty = false
		child.PositionChanged = false
		clearPaintDirtyBelow(child)
	}
}

// mergeHoverStyle merges hover style overrides into a base style.
func mergeHoverStyle(base Style, hover *Style) Style {
	if hover == nil {
		return base
	}
	result := base
	if hover.Foreground != "" {
		result.Foreground = hover.Foreground
	}
	if hover.Background != "" {
		result.Background = hover.Background
	}
	if hover.Bold {
		result.Bold = true
	}
	if hover.Dim {
		result.Dim = true
	}
	if hover.Underline {
		result.Underline = true
	}
	if hover.Italic {
		result.Italic = true
	}
	if hover.Strikethrough {
		result.Strikethrough = true
	}
	if hover.Inverse {
		result.Inverse = true
	}
	if hover.Border != "" {
		result.Border = hover.Border
	}
	if hover.BorderColor != "" {
		result.BorderColor = hover.BorderColor
	}
	return result
}

// splitLines splits text by newline characters.
func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	lines := make([]string, 0, 4)
	start := 0
	for i, ch := range s {
		if ch == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

// truncateRunesForWidth returns the number of runes from the slice that fit in maxW display columns.
func truncateRunesForWidth(runes []rune, maxW int) int {
	w := 0
	for i, r := range runes {
		rw := runeWidth(r)
		if w+rw > maxW {
			return i
		}
		w += rw
	}
	return len(runes)
}

// alignedX computes the starting X position for text alignment.
func alignedX(nodeX, availW, lineW int, textAlign string) int {
	x := nodeX
	switch textAlign {
	case "center":
		if offset := (availW - lineW) / 2; offset > 0 {
			x += offset
		}
	case "right":
		if offset := availW - lineW; offset > 0 {
			x += offset
		}
	}
	return x
}

// resolveSpanStyle resolves the effective style for a text span, falling back to node style.
func resolveSpanStyle(span *Span, nodeStyle *Style) (fg, bg string, bold, dim, underline, italic, strikethrough, inverse bool) {
	fg = span.Foreground
	if fg == "" {
		fg = nodeStyle.Foreground
	}
	bg = span.Background
	if bg == "" {
		bg = nodeStyle.Background
	}
	if span.Bold != nil {
		bold = *span.Bold
	} else {
		bold = nodeStyle.Bold
	}
	if span.Dim != nil {
		dim = *span.Dim
	} else {
		dim = nodeStyle.Dim
	}
	if span.Underline != nil {
		underline = *span.Underline
	} else {
		underline = nodeStyle.Underline
	}
	if span.Italic != nil {
		italic = *span.Italic
	} else {
		italic = nodeStyle.Italic
	}
	if span.Strikethrough != nil {
		strikethrough = *span.Strikethrough
	} else {
		strikethrough = nodeStyle.Strikethrough
	}
	if span.Inverse != nil {
		inverse = *span.Inverse
	} else {
		inverse = nodeStyle.Inverse
	}
	return
}

// findAncestorBackground walks up the tree to find the nearest ancestor with a background color.
func findAncestorBackground(node *Node) string {
	for n := node.Parent; n != nil; n = n.Parent {
		if n.Style.Background != "" {
			return n.Style.Background
		}
	}
	return ""
}
