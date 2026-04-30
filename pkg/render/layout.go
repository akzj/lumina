package render

// layoutViewportW and layoutViewportH hold the root layout dimensions (viewport).
// Set at the start of LayoutFull. Used to resolve vw/vh units.
// Thread-safe because layout is single-threaded.
var layoutViewportW, layoutViewportH int


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
	// Phase 1: Measure pass — compute intrinsic sizes bottom-up.
	// Results cached in node.MeasuredW / node.MeasuredH.
	measure(root, Constraints{
		Width:      w,
		WidthMode:  SizeModeExact,
		Height:     h,
		HeightMode: SizeModeExact,
	})
	// Phase 2: Layout pass — position nodes top-down using measurements.
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
	// Normalize entire subtree — newly-added children may have shorthand spacing only.
	normalizeSpacingInTree(node)
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

	// Mark PaintDirty and track position changes for clearing old region during paint
	if node.X != x || node.Y != y || node.W != w || node.H != h {
		node.OldX, node.OldY, node.OldW, node.OldH = node.X, node.Y, node.W, node.H
		node.PositionChanged = true
		node.PaintDirty = true
		// Parent must repaint to clear the old position area
		if node.Parent != nil {
			node.Parent.PaintDirty = true
		}
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
			if hasBorder(p.Style) {
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
