package render

import (
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
)

// layoutViewportW and layoutViewportH hold the root layout dimensions (viewport).
// Set at the start of LayoutFull. Used to resolve vw/vh units.
// Thread-safe because layout is single-threaded.
var layoutViewportW, layoutViewportH int

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
// No real UI tree exceeds this depth.
const maxLayoutDepth = 500

// LayoutFull computes layout for the entire Node tree.
// Used for initial render and after structural changes.
// The root is positioned at (x, y) with available size (w, h).
func LayoutFull(root *Node, x, y, w, h int) {
	if root == nil {
		return
	}
	layoutViewportW = w
	layoutViewportH = h
	normalizeSpacingInTree(root)
	computeFlex(root, x, y, w, h, 0)
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
	visited := make(map[*Node]bool)
	layoutDirtyWalkImpl(node, visited)
}

func layoutDirtyWalkImpl(node *Node, visited map[*Node]bool) {
	if node == nil || visited[node] {
		return
	}
	visited[node] = true

	if !node.LayoutDirty {
		// This node's layout is cached and valid.
		// But check children in case a descendant is dirty.
		for _, child := range node.Children {
			layoutDirtyWalkImpl(child, visited)
		}
		return
	}

	// This node needs re-layout.
	// Recompute using its CURRENT (cached) position and size as the container.
	normalizeSpacing(node)
	computeFlex(node, node.X, node.Y, node.W, node.H, 0)
	node.LayoutDirty = false
	// All children within this subtree are now re-laid-out.
	clearLayoutDirtyBelow(node)
}

// clearLayoutDirty clears LayoutDirty on the entire tree.
func clearLayoutDirty(node *Node) {
	visited := make(map[*Node]bool)
	clearLayoutDirtyImpl(node, visited)
}

func clearLayoutDirtyImpl(node *Node, visited map[*Node]bool) {
	if node == nil || visited[node] {
		return
	}
	visited[node] = true
	node.LayoutDirty = false
	for _, child := range node.Children {
		clearLayoutDirtyImpl(child, visited)
	}
}

// clearLayoutDirtyBelow clears LayoutDirty on all children (not the node itself).
func clearLayoutDirtyBelow(node *Node) {
	visited := make(map[*Node]bool)
	clearLayoutDirtyBelowImpl(node, visited)
}

func clearLayoutDirtyBelowImpl(node *Node, visited map[*Node]bool) {
	if visited[node] {
		return
	}
	visited[node] = true
	for _, child := range node.Children {
		if child == nil || visited[child] {
			continue
		}
		child.LayoutDirty = false
		clearLayoutDirtyBelowImpl(child, visited)
	}
}

// normalizeSpacingInTree expands Padding/Margin shorthand into per-side fields
// for the entire tree. Uses a visited set to prevent infinite recursion from
// cycles in the node tree (which can occur after hot reload with stale modules).
func normalizeSpacingInTree(node *Node) {
	visited := make(map[*Node]bool)
	normalizeSpacingWalk(node, visited)
}

func normalizeSpacingWalk(node *Node, visited map[*Node]bool) {
	if node == nil || visited[node] {
		return
	}
	visited[node] = true
	normalizeSpacing(node)
	for _, child := range node.Children {
		normalizeSpacingWalk(child, visited)
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
// probe (see computeFlex(..., 99999)). laidOutFlowContentHeight must not treat
// such heights as real content extent or scroll/card stacks get finalH≈99999 and
// only the first row paints.
const layoutProbeHMin = 99900

// intrinsicFlowHeight returns the vertical extent of n for scroll / intrinsic
// sizing when n.H may be a probe value (>= layoutProbeHMin) from unlimited-height
// measurement. Uses flow children and absolute Y offsets from a completed layout.
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
// bottom among flow (non-positioned) descendants after layout. computeFlex(..., huge H)
// sets node.H to the probe height; scroll sizing must not use that as content height
// or siblings stack at Y += 99999 and only the first panel is visible.
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

// --- Core flexbox layout ---

// computeFlex computes the layout for a Node tree using flexbox semantics.
func computeFlex(node *Node, x, y, w, h int, depth int) {
	if depth > maxLayoutDepth {
		return // cycle detected — break infinite recursion
	}
	style := node.Style

	// display:none — node takes no space, not rendered
	if style.Display == "none" {
		node.X = x
		node.Y = y
		node.W = 0
		node.H = 0
		return
	}

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

	// Apply fixed sizing with min/max constraints.
	// Percentage/viewport sizing is resolved by the parent layout functions
	// (layoutVBox/layoutHBox) which pass correctly resolved w/h values.
	// Here we only apply absolute constraints.
	// Exception: if the parent has applied flexShrink and passed a smaller h/w,
	// respect the parent's decision (don't override with explicit size).
	if style.Width > 0 {
		cw := clamp(style.Width, style.MinWidth, style.MaxWidth)
		if style.FlexShrink == 0 || cw <= w {
			w = cw
		}
	} else if style.MinWidth > 0 || style.MaxWidth > 0 {
		w = clamp(w, style.MinWidth, style.MaxWidth)
	}
	if style.Height > 0 {
		ch := clamp(style.Height, style.MinHeight, style.MaxHeight)
		if style.FlexShrink == 0 || ch <= h {
			h = ch
		}
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
		layoutVBox(node, x, y, w, h, style, depth)
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
				if rw := resolveWidth(cs, parentW); rw > 0 {
					cw = rw
				}
				if rh := resolveHeight(cs, parentH); rh > 0 {
					ch = rh
				}
				if cs.Right >= 0 && cs.Left == 0 {
					cx = parentX + parentW - cw - cs.Right
				}
				if cs.Bottom >= 0 && cs.Top == 0 {
					cy = parentY + parentH - ch - cs.Bottom
				}
				computeFlex(child, cx, cy, cw, ch, depth+1)
			} else {
				computeFlex(child, x, y, w, h, depth+1)
			}
		}
		// Graft is a single root laid out in the placeholder's box; the root may end
		// taller/wider than the slot (e.g. LuxButton style.Height=3 while flex-wrap
		// passes childH=1). Keep placeholder W/H in sync so layoutHBoxWrap row height,
		// parent vbox totals, and hit-testing match what paint draws.
		if len(node.Children) == 1 {
			root := node.Children[0]
			if root.X == node.X && root.Y == node.Y {
				// Match placeholder outer box to the grafted root (grow or shrink).
				// Only growing used to leave a taller slot than the real button; the
				// extra row sat under the bottom border and looked like a line drawn
				// on the border (e.g. outlined LuxButton in flex-wrap rows).
				if root.W != node.W {
					node.W = root.W
					node.PaintDirty = true
				}
				if root.H != node.H {
					node.H = root.H
					node.PaintDirty = true
				}
			}
		}
		return

	case "text":
		layoutText(node, layoutW)
		// Respect explicit height constraint — layoutText computes natural height
		// but explicit style.Height should override (clamp/truncate text)
		if style.Height > 0 {
			node.H = clamp(style.Height, style.MinHeight, style.MaxHeight)
		}

	case "vbox":
		if style.Display == "grid" {
			layoutGrid(node, contentX, contentY, layoutW, contentH, style, depth)
		} else {
			layoutVBox(node, contentX, contentY, layoutW, contentH, style, depth)
		}

	case "hbox":
		if style.Display == "grid" {
			layoutGrid(node, contentX, contentY, layoutW, contentH, style, depth)
		} else {
			layoutHBox(node, contentX, contentY, layoutW, contentH, style, depth)
		}

	default:
		// Generic container (box, etc.) — stack children vertically like vbox
		if style.Display == "grid" {
			layoutGrid(node, contentX, contentY, layoutW, contentH, style, depth)
		} else {
			layoutVBox(node, contentX, contentY, layoutW, contentH, style, depth)
		}
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
			if rw := resolveWidth(cs, contentW); rw > 0 {
				cw = rw
			}
			if rh := resolveHeight(cs, contentH); rh > 0 {
				ch = rh
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
			computeFlex(child, cx, cy, cw, ch, depth+1)

		case "fixed":
			cx := cs.Left
			cy := cs.Top
			cw := child.W
			ch := child.H
			if rw := resolveWidth(cs, node.W); rw > 0 {
				cw = rw
			}
			if rh := resolveHeight(cs, node.H); rh > 0 {
				ch = rh
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
			computeFlex(child, cx, cy, cw, ch, depth+1)
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
	// Scroll parents measure children with ~99999px height; nested non-scroll vboxes
	// must not treat that as flex free space (else every flex child grows huge and
	// stacked cards collapse visually to "first card only" with broken ScrollHeight).
	intrinsicMeasure := !isScroll && contentH >= layoutProbeHMin
	// Second pass: inner content vbox gets real (tall) height from scroll but must still
	// stack children at natural heights — not flex-split that height across cards.
	scrollContentStack := !isScroll && !intrinsicMeasure && nodeHasScrollAncestor(node)

	useNaturalHeights := isScroll || intrinsicMeasure || scrollContentStack

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
					// Intrinsic height: measure by laying out with unlimited height.
					computeFlex(child, contentX, contentY, contentW, 99999, depth+1)
					naturalH := laidOutFlowContentHeight(child)
					children[i].finalH = naturalH + marginV
				}
				if mnh := resolveMinH(cs, contentH); mnh > 0 && children[i].finalH < mnh+marginV {
					children[i].finalH = mnh + marginV
				}
			} else if len(child.Children) > 0 {
				// Intrinsic height for vbox/box/hbox/etc. Min/max height floors/ceil the
				// measured size — do not use minHeight alone or scroll content collapses
				// to one panel when the inner column has minHeight + many children.
				computeFlex(child, contentX, contentY, contentW, 99999, depth+1)
				naturalH := laidOutFlowContentHeight(child)
				children[i].finalH = naturalH + marginV
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
						children[i].finalH = clamp(children[i].finalH, children[i].style.MinHeight, 0)
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
	if !isScroll && !intrinsicMeasure && !scrollContentStack {
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
					children[i].finalW = clamp(children[i].finalW, children[i].style.MinWidth, 0)
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
}

// layoutVBoxWrap lays out children in a vertical column with wrapping.
// When children overflow the available height, they wrap to the next column.
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

// --- CSS Grid layout ---

// gridTrack represents a single track in a grid template.
type gridTrack struct {
	fr int // fractional unit (0 = not fr)
	px int // pixel/cell value (0 = auto)
}

// parseGridTemplate parses a grid template string like "1fr 2fr 100" into track sizes.
// available is the total space to distribute among tracks.
// gapSize is the gap between tracks.
func parseGridTemplate(template string, available int, gapSize int) []int {
	if template == "" {
		return nil
	}
	parts := strings.Fields(template)
	if len(parts) == 0 {
		return nil
	}

	tracks := make([]gridTrack, len(parts))
	fixedTotal := 0
	frTotal := 0

	for i, part := range parts {
		if strings.HasSuffix(part, "fr") {
			nStr := strings.TrimSuffix(part, "fr")
			n, err := strconv.Atoi(nStr)
			if err != nil || n < 1 {
				n = 1
			}
			tracks[i] = gridTrack{fr: n}
			frTotal += n
		} else if part == "auto" {
			tracks[i] = gridTrack{px: 0} // auto = 0, will get remaining space
			frTotal += 1
			tracks[i] = gridTrack{fr: 1} // treat auto as 1fr
		} else {
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 {
				n = 0
			}
			tracks[i] = gridTrack{px: n}
			fixedTotal += n
		}
	}

	// Subtract gaps from available space
	totalGaps := 0
	if len(tracks) > 1 {
		totalGaps = gapSize * (len(tracks) - 1)
	}
	remainW := available - fixedTotal - totalGaps
	if remainW < 0 {
		remainW = 0
	}

	sizes := make([]int, len(tracks))
	for i, t := range tracks {
		if t.fr > 0 && frTotal > 0 {
			sizes[i] = (remainW * t.fr) / frTotal
		} else {
			sizes[i] = t.px
		}
		if sizes[i] < 0 {
			sizes[i] = 0
		}
	}
	return sizes
}

// parseGridSpan parses "1 / 3" into (start=1, end=3) or "2" into (start=2, end=3).
// Returns 1-based start and exclusive end.
func parseGridSpan(s string) (int, int) {
	if s == "" {
		return 0, 0
	}
	parts := strings.Split(s, "/")
	if len(parts) == 2 {
		start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err1 == nil && err2 == nil && start >= 1 && end > start {
			return start, end
		}
	}
	// Single number: occupies one cell
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err == nil && n >= 1 {
		return n, n + 1
	}
	return 0, 0
}

// layoutGrid lays out children in a CSS Grid.
func layoutGrid(node *Node, contentX, contentY, contentW, contentH int, style Style, depth int) {
	if len(node.Children) == 0 {
		return
	}

	// Determine gaps
	colGap := style.GridColumnGap
	if colGap == 0 {
		colGap = style.Gap
	}
	rowGap := style.GridRowGap
	if rowGap == 0 {
		rowGap = style.Gap
	}

	// Parse column template
	colSizes := parseGridTemplate(style.GridTemplateColumns, contentW, colGap)
	if len(colSizes) == 0 {
		// Default: single column
		colSizes = []int{contentW}
	}
	numCols := len(colSizes)

	// Collect flow children
	type gridChild struct {
		childIdx int
		colStart int // 1-based
		colEnd   int // exclusive
		rowStart int // 1-based
		rowEnd   int // exclusive
	}
	var flowChildren []gridChild
	for i, child := range node.Children {
		cs := child.Style
		if isPositioned(cs) || cs.Display == "none" {
			continue
		}

		colStart, colEnd := 0, 0
		rowStart, rowEnd := 0, 0

		// Explicit placement via gridColumn/gridRow strings
		if cs.GridColumn != "" {
			colStart, colEnd = parseGridSpan(cs.GridColumn)
		}
		if cs.GridRow != "" {
			rowStart, rowEnd = parseGridSpan(cs.GridRow)
		}

		// Explicit placement via gridColumnStart/End fields
		if colStart == 0 && cs.GridColumnStart > 0 {
			colStart = cs.GridColumnStart
			colEnd = cs.GridColumnEnd
			if colEnd <= colStart {
				colEnd = colStart + 1
			}
		}
		if rowStart == 0 && cs.GridRowStart > 0 {
			rowStart = cs.GridRowStart
			rowEnd = cs.GridRowEnd
			if rowEnd <= rowStart {
				rowEnd = rowStart + 1
			}
		}

		flowChildren = append(flowChildren, gridChild{
			childIdx: i,
			colStart: colStart,
			colEnd:   colEnd,
			rowStart: rowStart,
			rowEnd:   rowEnd,
		})
	}

	if len(flowChildren) == 0 {
		return
	}

	// Auto-place children that don't have explicit positions
	// Track which cells are occupied
	// First pass: determine number of rows needed
	maxRow := 0
	for _, gc := range flowChildren {
		if gc.rowEnd > maxRow {
			maxRow = gc.rowEnd - 1
		}
	}
	// Estimate rows from auto-placement
	autoPlaceCount := 0
	for _, gc := range flowChildren {
		if gc.colStart == 0 {
			autoPlaceCount++
		}
	}
	estimatedRows := maxRow
	if autoPlaceCount > 0 {
		neededRows := (autoPlaceCount + numCols - 1) / numCols
		if neededRows > estimatedRows {
			estimatedRows = neededRows
		}
	}
	if estimatedRows < 1 {
		estimatedRows = 1
	}

	// Build occupancy grid and place explicitly positioned children
	numRows := estimatedRows + maxRow // overestimate to be safe
	if numRows < 1 {
		numRows = 1
	}
	occupied := make([][]bool, numRows)
	for r := range occupied {
		occupied[r] = make([]bool, numCols)
	}

	// Place explicitly positioned children
	for i := range flowChildren {
		gc := &flowChildren[i]
		if gc.colStart > 0 && gc.rowStart > 0 {
			// Mark cells as occupied
			for r := gc.rowStart - 1; r < gc.rowEnd-1 && r < numRows; r++ {
				for c := gc.colStart - 1; c < gc.colEnd-1 && c < numCols; c++ {
					occupied[r][c] = true
				}
			}
		}
	}

	// Auto-place remaining children
	autoRow, autoCol := 0, 0
	for i := range flowChildren {
		gc := &flowChildren[i]
		if gc.colStart > 0 && gc.rowStart > 0 {
			continue // already placed
		}

		// Find next available cell
		span := 1
		if gc.colStart > 0 && gc.colEnd > gc.colStart {
			span = gc.colEnd - gc.colStart
		}

		for {
			if autoCol+span <= numCols {
				// Check if cells are free
				free := true
				for c := autoCol; c < autoCol+span; c++ {
					if autoRow < numRows && occupied[autoRow][c] {
						free = false
						break
					}
				}
				if free {
					break
				}
			}
			autoCol++
			if autoCol+span > numCols {
				autoCol = 0
				autoRow++
				// Expand occupied grid if needed
				if autoRow >= numRows {
					numRows++
					occupied = append(occupied, make([]bool, numCols))
				}
			}
		}

		gc.colStart = autoCol + 1
		gc.colEnd = autoCol + span + 1
		if gc.colEnd > numCols+1 {
			gc.colEnd = numCols + 1
		}
		gc.rowStart = autoRow + 1
		gc.rowEnd = autoRow + 2
		if gc.rowEnd-1 > gc.rowStart-1 {
			// multi-row not supported in auto-place
		}

		// Mark occupied
		for c := autoCol; c < autoCol+span && c < numCols; c++ {
			if autoRow < numRows {
				occupied[autoRow][c] = true
			}
		}

		// Advance auto cursor
		autoCol += span
		if autoCol >= numCols {
			autoCol = 0
			autoRow++
			if autoRow >= numRows {
				numRows++
				occupied = append(occupied, make([]bool, numCols))
			}
		}
	}

	// Determine actual number of rows used
	actualRows := 0
	for _, gc := range flowChildren {
		if gc.rowEnd-1 > actualRows {
			actualRows = gc.rowEnd - 1
		}
	}
	if actualRows < 1 {
		actualRows = 1
	}

	// Parse row template
	rowSizes := parseGridTemplate(style.GridTemplateRows, contentH, rowGap)
	// If row template doesn't cover all rows, extend with equal distribution
	if len(rowSizes) < actualRows {
		// Calculate remaining height after defined rows
		definedH := 0
		for _, h := range rowSizes {
			definedH += h
		}
		definedGaps := 0
		if len(rowSizes) > 0 {
			definedGaps = rowGap * len(rowSizes)
		}
		remainH := contentH - definedH - definedGaps
		if remainH < 0 {
			remainH = 0
		}
		extraRows := actualRows - len(rowSizes)
		extraGaps := 0
		if extraRows > 1 {
			extraGaps = rowGap * (extraRows - 1)
		}
		perRow := 1
		if extraRows > 0 && remainH > extraGaps {
			perRow = (remainH - extraGaps) / extraRows
		}
		if perRow < 1 {
			perRow = 1
		}
		for len(rowSizes) < actualRows {
			rowSizes = append(rowSizes, perRow)
		}
	}

	// Compute column X positions
	colX := make([]int, numCols)
	cx := contentX
	for c := 0; c < numCols; c++ {
		colX[c] = cx
		cx += colSizes[c]
		if c < numCols-1 {
			cx += colGap
		}
	}

	// Compute row Y positions
	rowY := make([]int, actualRows)
	ry := contentY
	for r := 0; r < actualRows; r++ {
		rowY[r] = ry
		ry += rowSizes[r]
		if r < actualRows-1 {
			ry += rowGap
		}
	}

	// Position each child in its grid cell(s)
	for _, gc := range flowChildren {
		child := node.Children[gc.childIdx]

		c0 := gc.colStart - 1
		c1 := gc.colEnd - 2 // inclusive end column
		r0 := gc.rowStart - 1
		r1 := gc.rowEnd - 2 // inclusive end row

		// Clamp to grid bounds
		if c0 < 0 {
			c0 = 0
		}
		if c1 >= numCols {
			c1 = numCols - 1
		}
		if r0 < 0 {
			r0 = 0
		}
		if r1 >= actualRows {
			r1 = actualRows - 1
		}

		// Calculate cell position and size
		cellX := colX[c0]
		cellY := rowY[r0]

		// Width spans from colStart to colEnd (inclusive of gaps between spanned columns)
		cellW := 0
		for c := c0; c <= c1; c++ {
			cellW += colSizes[c]
			if c < c1 {
				cellW += colGap
			}
		}

		// Height spans from rowStart to rowEnd
		cellH := 0
		for r := r0; r <= r1; r++ {
			cellH += rowSizes[r]
			if r < r1 {
				cellH += rowGap
			}
		}

		if cellW < 1 {
			cellW = 1
		}
		if cellH < 1 {
			cellH = 1
		}

		computeFlex(child, cellX, cellY, cellW, cellH, depth+1)
		applyRelativeOffset(child, child.Style)
	}
}
