package render

import (
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
)

// --- Parsing helpers ---

// parsePercent parses "50%" → (50, true).
func parsePercent(s string) (int, bool) {
	if strings.HasSuffix(s, "%") {
		v := strings.TrimSuffix(s, "%")
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil && n >= 0 {
			return n, true
		}
	}
	return 0, false
}

// parseViewport parses "50vw" → (50, "vw", true) or "50vh" → (50, "vh", true).
func parseViewport(s string) (int, string, bool) {
	if strings.HasSuffix(s, "vw") {
		v := strings.TrimSuffix(s, "vw")
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil && n >= 0 {
			return n, "vw", true
		}
	}
	if strings.HasSuffix(s, "vh") {
		v := strings.TrimSuffix(s, "vh")
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err == nil && n >= 0 {
			return n, "vh", true
		}
	}
	return 0, "", false
}

// maxLayoutDepth is a safety limit to prevent infinite recursion from cycles
// in the node tree (which can occur after hot reload with stale modules).

// --- Size resolution ---

// resolveWidth returns the effective width for a style, resolving percentage
// and viewport units against parentW. Returns 0 if no width is set.
func resolveWidth(s Style, parentW int) int {
	if s.Width > 0 {
		return s.Width
	}
	if s.WidthPercent > 0 {
		return (parentW * s.WidthPercent) / 100
	}
	if s.WidthVW > 0 {
		return (layoutViewportW * s.WidthVW) / 100
	}
	return 0
}

// resolveHeight returns the effective height for a style, resolving percentage
// and viewport units against parentH. Returns 0 if no height is set.
func resolveHeight(s Style, parentH int) int {
	if s.Height > 0 {
		return s.Height
	}
	if s.HeightPercent > 0 {
		return (parentH * s.HeightPercent) / 100
	}
	if s.HeightVH > 0 {
		return (layoutViewportH * s.HeightVH) / 100
	}
	return 0
}

// resolveMinW returns the effective minWidth, resolving percentage against parentW.
func resolveMinW(s Style, parentW int) int {
	if s.MinWidth > 0 {
		return s.MinWidth
	}
	if s.MinWidthPercent > 0 {
		return (parentW * s.MinWidthPercent) / 100
	}
	return 0
}

// resolveMaxW returns the effective maxWidth, resolving percentage against parentW.
func resolveMaxW(s Style, parentW int) int {
	if s.MaxWidth > 0 {
		return s.MaxWidth
	}
	if s.MaxWidthPercent > 0 {
		return (parentW * s.MaxWidthPercent) / 100
	}
	return 0
}

// resolveMinH returns the effective minHeight, resolving percentage against parentH.
func resolveMinH(s Style, parentH int) int {
	if s.MinHeight > 0 {
		return s.MinHeight
	}
	if s.MinHeightPercent > 0 {
		return (parentH * s.MinHeightPercent) / 100
	}
	return 0
}

// resolveMaxH returns the effective maxHeight, resolving percentage against parentH.
func resolveMaxH(s Style, parentH int) int {
	if s.MaxHeight > 0 {
		return s.MaxHeight
	}
	if s.MaxHeightPercent > 0 {
		return (parentH * s.MaxHeightPercent) / 100
	}
	return 0
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

// minBorderPaintW is the smallest outer width for which paintBorder draws a box
// (see painter.paintBorder). Flex rows must not clamp bordered widgets below this.
const minBorderPaintW = 2

// minOuterWidthForBorderPaint returns the minimum main-axis width a node needs so
// its visible border can render. Unwraps a single-graft component to the real root.
func minOuterWidthForBorderPaint(n *Node) int {
	if n == nil {
		return 0
	}
	if n.Type == "component" && len(n.Children) == 1 {
		return minOuterWidthForBorderPaint(n.Children[0])
	}
	if hasBorder(n.Style) {
		return minBorderPaintW
	}
	return 0
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

// layoutProbeHMin is the smallest node.H value treated as an intrinsic-measurement
// probe. DEPRECATED: The measure pass now provides intrinsic sizes via node.MeasuredH.
// This constant is kept only for backward compatibility with existing test assertions.
const layoutProbeHMin = 99900

// intrinsicFlowHeight returns the vertical extent of n for scroll / intrinsic
// sizing when n.H may be a probe value (>= layoutProbeHMin).
// DEPRECATED: Use node.MeasuredH from the measure pass instead.
// Kept for backward compatibility with existing test assertions.
func intrinsicFlowHeight(n *Node) int {
	if n == nil {
		return 1
	}
	if len(n.Children) == 0 {
		if n.H >= layoutProbeHMin {
			return 1
		}
		if n.H > 0 {
			return n.H
		}
		return 1
	}
	maxSpan := 1
	for _, c := range n.Children {
		cs := c.Style
		if isPositioned(cs) || cs.Display == "none" {
			continue
		}
		span := (c.Y - n.Y) + intrinsicFlowHeight(c)
		if span > maxSpan {
			maxSpan = span
		}
	}
	if n.H >= layoutProbeHMin {
		// Unlimited-measure probe height is not a real content box.
		return maxSpan
	}
	if n.H > 0 {
		// Use the larger of laid-out box height and descendant span: fixed Style.Height
		// / flex finalH can exceed a one-line text's span; a too-small outer H (e.g.
		// cross-axis hbox) must not hide taller overflow (Card rows below the border).
		if maxSpan > n.H {
			return maxSpan
		}
		return n.H
	}
	return maxSpan
}

// laidOutFlowContentHeight returns the vertical span from node.Y to the lowest
// bottom among flow (non-positioned) descendants after layout.
// DEPRECATED: Use node.MeasuredH from the measure pass instead.
// Kept for backward compatibility with existing test assertions.
func laidOutFlowContentHeight(node *Node) int {
	if node == nil {
		return 1
	}
	maxB := node.Y
	var walk func(*Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		for _, c := range n.Children {
			cs := c.Style
			if isPositioned(cs) || cs.Display == "none" {
				continue
			}
			if bb := c.Y + intrinsicFlowHeight(c); bb > maxB {
				maxB = bb
			}
			walk(c)
		}
	}
	walk(node)
	h := maxB - node.Y
	if h < 1 {
		return 1
	}
	return h
}

// nodeHasScrollAncestor is true when any ancestor has overflow:scroll.
// Scrollable main content is often not the direct parent of the inner column
// (e.g. component placeholder, plain box). Using only the immediate parent for
// scrollContentStack lets a post-intrinsic pass treat contentH as flex space and
// split it across stacked cards so only the first panel appears.
func nodeHasScrollAncestor(n *Node) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Style.Overflow == "scroll" {
			return true
		}
	}
	return false
}


// --- Text layout ---

func layoutText(node *Node, availW int) {
	if availW <= 0 {
		node.H = 1
		return
	}
	if node.Content == "" {
		node.H = 1
		return
	}

	// whiteSpace: "nowrap" — no wrapping, height = number of \n lines
	if node.Style.WhiteSpace == "nowrap" {
		lines := strings.Split(node.Content, "\n")
		totalH := len(lines)
		if totalH < 1 {
			totalH = 1
		}
		node.H = totalH
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
