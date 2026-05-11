package render

// Reconcile updates an existing Node tree to match a new Descriptor.
// It patches the node in-place, marking LayoutDirty/PaintDirty as needed.
// Returns true if any changes were made.
// Old Lua refs replaced during reconcile are not tracked (use ReconcileCollectRefs
// when you need to unref them).
func Reconcile(node *Node, desc Descriptor) bool {
	return reconcileImpl(node, desc, nil)
}

// ReconcileCollectRefs is like Reconcile but appends replaced/removed Lua refs
// to freedRefs so the caller can call L.Unref on them.
func ReconcileCollectRefs(node *Node, desc Descriptor, freedRefs *[]int64) bool {
	return reconcileImpl(node, desc, freedRefs)
}

func reconcileImpl(node *Node, desc Descriptor, freedRefs *[]int64) bool {
	changed := false

	// 1. Update content
	if desc.Content != node.Content {
		oldContent := node.Content
		node.Content = desc.Content
		node.PaintDirty = true
		// If display width changed, mark layout dirty so parent re-lays out.
		if stringWidth(oldContent) != stringWidth(node.Content) {
			node.MarkLayoutDirty()
		}
		changed = true
	}

	// 1a. Update spans
	if !spansEqual(desc.Spans, node.Spans) {
		oldWidth := spansWidth(node.Spans)
		node.Spans = desc.Spans
		node.PaintDirty = true
		if oldWidth != spansWidth(node.Spans) {
			node.MarkLayoutDirty()
		}
		changed = true
	}

	// 1b. Update placeholder
	if desc.Placeholder != node.Placeholder {
		node.Placeholder = desc.Placeholder
		node.PaintDirty = true
		changed = true
	}

	// 1c. Update scroll position (only when Lua explicitly sets scrollY)
	if desc.ScrollYSet && desc.ScrollY != node.ScrollY {
		node.ScrollY = desc.ScrollY
		node.PaintDirty = true
		changed = true
	}

	// 2. Update ComponentType and ComponentProps (for component placeholder nodes)
	if desc.ComponentType != node.ComponentType {
		node.ComponentType = desc.ComponentType
		changed = true
	}
	node.ComponentProps = desc.ComponentProps

	// 3. Update style (check each field that affects paint vs layout)
	if styleChanged := reconcileStyle(node, desc.Style); styleChanged {
		changed = true
	}

	// 3b. Update hoverStyle
	if !hoverStyleEqual(node.HoverStyle, desc.HoverStyle) {
		node.HoverStyle = desc.HoverStyle
		node.PaintDirty = true
		changed = true
		// If HoverStyle was removed, clear Hovered flag to prevent stale hover painting
		if node.HoverStyle == nil && node.Hovered {
			node.Hovered = false
		}
	}

	// 4. Update event handlers (just swap refs, collect old refs for cleanup)
	changed = updateRef(&node.OnClick, desc.OnClick, freedRefs) || changed
	changed = updateRef(&node.OnRightClick, desc.OnRightClick, freedRefs) || changed
	changed = updateRef(&node.OnMouseEnter, desc.OnMouseEnter, freedRefs) || changed
	changed = updateRef(&node.OnMouseLeave, desc.OnMouseLeave, freedRefs) || changed
	changed = updateRef(&node.OnKeyDown, desc.OnKeyDown, freedRefs) || changed
	changed = updateRef(&node.OnChange, desc.OnChange, freedRefs) || changed
	changed = updateRef(&node.OnScroll, desc.OnScroll, freedRefs) || changed
	changed = updateRef(&node.OnMouseDown, desc.OnMouseDown, freedRefs) || changed
	changed = updateRef(&node.OnMouseUp, desc.OnMouseUp, freedRefs) || changed
	changed = updateRef(&node.OnMouseMove, desc.OnMouseMove, freedRefs) || changed
	changed = updateRef(&node.OnFocus, desc.OnFocus, freedRefs) || changed
	changed = updateRef(&node.OnBlur, desc.OnBlur, freedRefs) || changed
	changed = updateRef(&node.OnSubmit, desc.OnSubmit, freedRefs) || changed
	changed = updateRef(&node.OnOutsideClick, desc.OnOutsideClick, freedRefs) || changed
	changed = updateRef(&node.RefTableRef, desc.RefTableRef, freedRefs) || changed

	// 4b. Update node fields
	if desc.Focusable != node.Focusable {
		node.Focusable = desc.Focusable
		changed = true
	}
	if desc.Disabled != node.Disabled {
		node.Disabled = desc.Disabled
		node.PaintDirty = true
		changed = true
	}
	if desc.Role != node.Role {
		node.Role = desc.Role
	}

	// Cursor hint (always update — no dirty flag needed, just position tracking)
	node.CursorHintCol = desc.CursorHintCol
	node.CursorHintRow = desc.CursorHintRow

	// 5. Reconcile children (skip for component nodes — children are grafted)
	if node.Type != "component" {
		if reconcileChildrenImpl(node, desc.Children, freedRefs) {
			changed = true
		}
	}

	return changed
}

// reconcileStyle updates the node's style, marking appropriate dirty flags.
func reconcileStyle(node *Node, newStyle Style) bool {
	old := node.Style
	if old == newStyle {
		return false
	}

	// Check if layout-affecting properties changed
	layoutChanged := old.Width != newStyle.Width ||
		old.Height != newStyle.Height ||
		old.WidthPercent != newStyle.WidthPercent ||
		old.HeightPercent != newStyle.HeightPercent ||
		old.WidthVW != newStyle.WidthVW ||
		old.HeightVH != newStyle.HeightVH ||
		old.Flex != newStyle.Flex ||
		old.FlexShrink != newStyle.FlexShrink ||
		old.FlexBasis != newStyle.FlexBasis ||
		old.FlexWrap != newStyle.FlexWrap ||
		old.Padding != newStyle.Padding ||
		old.PaddingTop != newStyle.PaddingTop ||
		old.PaddingBottom != newStyle.PaddingBottom ||
		old.PaddingLeft != newStyle.PaddingLeft ||
		old.PaddingRight != newStyle.PaddingRight ||
		old.Margin != newStyle.Margin ||
		old.MarginTop != newStyle.MarginTop ||
		old.MarginBottom != newStyle.MarginBottom ||
		old.MarginLeft != newStyle.MarginLeft ||
		old.MarginRight != newStyle.MarginRight ||
		old.Gap != newStyle.Gap ||
		old.MinWidth != newStyle.MinWidth ||
		old.MaxWidth != newStyle.MaxWidth ||
		old.MinHeight != newStyle.MinHeight ||
		old.MaxHeight != newStyle.MaxHeight ||
		old.MinWidthPercent != newStyle.MinWidthPercent ||
		old.MaxWidthPercent != newStyle.MaxWidthPercent ||
		old.MinHeightPercent != newStyle.MinHeightPercent ||
		old.MaxHeightPercent != newStyle.MaxHeightPercent ||
		old.Justify != newStyle.Justify ||
		old.Align != newStyle.Align ||
		old.AlignSelf != newStyle.AlignSelf ||
		old.Order != newStyle.Order ||
		old.Border != newStyle.Border ||
		old.Overflow != newStyle.Overflow ||
		old.Scrollbar != newStyle.Scrollbar ||
		old.Position != newStyle.Position ||
		old.Top != newStyle.Top ||
		old.Left != newStyle.Left ||
		old.Right != newStyle.Right ||
		old.Bottom != newStyle.Bottom ||
		old.Display != newStyle.Display ||
		old.TextAlign != newStyle.TextAlign ||
		old.WhiteSpace != newStyle.WhiteSpace ||
		old.GridTemplateColumns != newStyle.GridTemplateColumns ||
		old.GridTemplateRows != newStyle.GridTemplateRows ||
		old.GridColumnGap != newStyle.GridColumnGap ||
		old.GridRowGap != newStyle.GridRowGap ||
		old.GridColumn != newStyle.GridColumn ||
		old.GridRow != newStyle.GridRow ||
		old.GridColumnStart != newStyle.GridColumnStart ||
		old.GridColumnEnd != newStyle.GridColumnEnd ||
		old.GridRowStart != newStyle.GridRowStart ||
		old.GridRowEnd != newStyle.GridRowEnd

	node.Style = newStyle

	if layoutChanged {
		node.MarkLayoutDirty()
	}
	node.PaintDirty = true
	return true
}

// reconcileChildren matches new descriptors against existing children by key+type.
func reconcileChildren(parent *Node, descs []Descriptor) bool {
	return reconcileChildrenImpl(parent, descs, nil)
}

func reconcileChildrenImpl(parent *Node, descs []Descriptor, freedRefs *[]int64) bool {
	oldChildren := parent.Children
	changed := false

	// Fast path: same length, same keys in order
	if len(oldChildren) == len(descs) && allKeysMatch(oldChildren, descs) {
		for i, desc := range descs {
			if reconcileImpl(oldChildren[i], desc, freedRefs) {
				changed = true
			}
		}
		return changed
	}

	// Build key→index map for existing children
	oldByKey := make(map[string]int, len(oldChildren))
	for i, child := range oldChildren {
		key := childKey(child)
		if key != "" {
			// Note: duplicate keys silently overwrite (last one wins).
			// This matches React behavior but may cause unexpected reconciliation.
			oldByKey[key] = i
		}
	}

	// Match new descriptors to existing children
	newChildren := make([]*Node, 0, len(descs))
	usedOld := make(map[int]bool, len(oldChildren))

	for i, desc := range descs {
		key := descKey(desc)
		if key != "" {
			// Keyed matching: look up by key+type
			if idx, ok := oldByKey[key]; ok && oldChildren[idx].Type == desc.Type {
				usedOld[idx] = true
				reconcileImpl(oldChildren[idx], desc, freedRefs)
				newChildren = append(newChildren, oldChildren[idx])
				continue
			}
		} else {
			// Keyless: positional fallback — try same index first, then scan forward
			matched := false
			if i < len(oldChildren) && !usedOld[i] && childKey(oldChildren[i]) == "" && oldChildren[i].Type == desc.Type {
				usedOld[i] = true
				reconcileImpl(oldChildren[i], desc, freedRefs)
				newChildren = append(newChildren, oldChildren[i])
				matched = true
			}
			if !matched {
				// Scan for first available keyless child of same type
				for j := 0; j < len(oldChildren); j++ {
					if !usedOld[j] && childKey(oldChildren[j]) == "" && oldChildren[j].Type == desc.Type {
						usedOld[j] = true
						reconcileImpl(oldChildren[j], desc, freedRefs)
						newChildren = append(newChildren, oldChildren[j])
						matched = true
						break
					}
				}
			}
			if matched {
				continue
			}
		}
		// No match found — create new node
		newNode := createNodeFromDesc(desc)
		newNode.Parent = parent
		newNode.PaintDirty = true
		newNode.LayoutDirty = true
		newChildren = append(newChildren, newNode)
		changed = true
	}

	// Cleanup removed children
	for i, child := range oldChildren {
		if !usedOld[i] {
			collectNodeRefsRecursive(child, freedRefs)
			markRemovedRecursive(child)
			changed = true
			parent.PaintDirty = true // Force parent repaint to clear ghost pixels
		}
	}

	// Detect reorder: same nodes but different positions
	reordered := false
	if len(newChildren) == len(oldChildren) && !changed {
		for i, child := range newChildren {
			if child != oldChildren[i] {
				reordered = true
				break
			}
		}
	}

	if changed || reordered || len(newChildren) != len(oldChildren) {
		parent.Children = newChildren
		parent.LayoutDirty = true
		changed = true
	}

	return changed
}

// Helper functions

func updateRef(ref *LuaRef, newRef LuaRef, freedRefs *[]int64) bool {
	if *ref != newRef {
		if *ref != 0 && freedRefs != nil {
			*freedRefs = append(*freedRefs, *ref)
		}
		*ref = newRef
		return true
	}
	return false
}

// collectNodeRefsRecursive appends all non-zero Lua refs from a node and its
// descendants to freedRefs. Used when removing nodes from the tree.
func collectNodeRefsRecursive(node *Node, freedRefs *[]int64) {
	if freedRefs == nil {
		return
	}
	collectNodeRefs(node, freedRefs)
	for _, child := range node.Children {
		collectNodeRefsRecursive(child, freedRefs)
	}
}

// collectNodeRefs appends all non-zero Lua refs from a single node.
func collectNodeRefs(node *Node, refs *[]int64) {
	if node.OnClick != 0 {
		*refs = append(*refs, node.OnClick)
	}
	if node.OnRightClick != 0 {
		*refs = append(*refs, node.OnRightClick)
	}
	if node.OnMouseEnter != 0 {
		*refs = append(*refs, node.OnMouseEnter)
	}
	if node.OnMouseLeave != 0 {
		*refs = append(*refs, node.OnMouseLeave)
	}
	if node.OnKeyDown != 0 {
		*refs = append(*refs, node.OnKeyDown)
	}
	if node.OnChange != 0 {
		*refs = append(*refs, node.OnChange)
	}
	if node.OnScroll != 0 {
		*refs = append(*refs, node.OnScroll)
	}
	if node.OnMouseDown != 0 {
		*refs = append(*refs, node.OnMouseDown)
	}
	if node.OnMouseUp != 0 {
		*refs = append(*refs, node.OnMouseUp)
	}
	if node.OnMouseMove != 0 {
		*refs = append(*refs, node.OnMouseMove)
	}
	if node.OnFocus != 0 {
		*refs = append(*refs, node.OnFocus)
	}
	if node.OnBlur != 0 {
		*refs = append(*refs, node.OnBlur)
	}
	if node.OnSubmit != 0 {
		*refs = append(*refs, node.OnSubmit)
	}
	if node.OnOutsideClick != 0 {
		*refs = append(*refs, node.OnOutsideClick)
	}
	if node.RefTableRef != 0 {
		*refs = append(*refs, node.RefTableRef)
	}
}

func childKey(node *Node) string {
	if node.Key != "" {
		return node.Type + ":" + node.Key
	}
	if node.ID != "" {
		return node.Type + ":" + node.ID
	}
	return ""
}

func descKey(desc Descriptor) string {
	if desc.Key != "" {
		return desc.Type + ":" + desc.Key
	}
	if desc.ID != "" {
		return desc.Type + ":" + desc.ID
	}
	return ""
}

func allKeysMatch(nodes []*Node, descs []Descriptor) bool {
	for i, desc := range descs {
		if childKey(nodes[i]) != descKey(desc) || nodes[i].Type != desc.Type {
			return false
		}
	}
	return true
}

func createNodeFromDesc(desc Descriptor) *Node {
	node := NewNode(desc.Type)
	node.ID = desc.ID
	node.Key = desc.Key
	node.Content = desc.Content
	node.Spans = desc.Spans
	node.Placeholder = desc.Placeholder
	node.AutoFocus = desc.AutoFocus
	if desc.ScrollYSet {
		node.ScrollY = desc.ScrollY
	}
	node.Style = desc.Style
	node.HoverStyle = desc.HoverStyle
	node.ComponentType = desc.ComponentType
	node.ComponentProps = desc.ComponentProps
	node.Focusable = desc.Focusable
	node.Disabled = desc.Disabled
	node.Role = desc.Role
	node.CursorHintCol = desc.CursorHintCol
	node.CursorHintRow = desc.CursorHintRow
	node.OnClick = desc.OnClick
	node.OnRightClick = desc.OnRightClick
	node.OnMouseEnter = desc.OnMouseEnter
	node.OnMouseLeave = desc.OnMouseLeave
	node.OnKeyDown = desc.OnKeyDown
	node.OnChange = desc.OnChange
	node.OnScroll = desc.OnScroll
	node.OnMouseDown = desc.OnMouseDown
	node.OnMouseUp = desc.OnMouseUp
	node.OnMouseMove = desc.OnMouseMove
	node.OnFocus = desc.OnFocus
	node.OnBlur = desc.OnBlur
	node.OnSubmit = desc.OnSubmit
	node.OnOutsideClick = desc.OnOutsideClick
	node.RefTableRef = desc.RefTableRef

	for _, childDesc := range desc.Children {
		child := createNodeFromDesc(childDesc)
		node.AddChild(child)
	}
	return node
}

// isDescendantOf returns true if node is the root or a descendant of root.
func isDescendantOf(node, root *Node) bool {
	for n := node; n != nil; n = n.Parent {
		if n == root {
			return true
		}
	}
	return false
}

// markRemovedRecursive sets Removed=true and Parent=nil on a node and all descendants.
func markRemovedRecursive(node *Node) {
	node.Removed = true
	node.Parent = nil
	for _, child := range node.Children {
		markRemovedRecursive(child)
	}
}

// spansEqual compares two Span slices for equality.
func spansEqual(a, b []Span) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Text != b[i].Text ||
			a[i].Foreground != b[i].Foreground ||
			a[i].Background != b[i].Background ||
			!boolPtrEqual(a[i].Bold, b[i].Bold) ||
			!boolPtrEqual(a[i].Dim, b[i].Dim) ||
			!boolPtrEqual(a[i].Underline, b[i].Underline) ||
			!boolPtrEqual(a[i].Italic, b[i].Italic) ||
			!boolPtrEqual(a[i].Strikethrough, b[i].Strikethrough) ||
			!boolPtrEqual(a[i].Inverse, b[i].Inverse) {
			return false
		}
	}
	return true
}

// boolPtrEqual compares two *bool values.
func boolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// spansWidth returns the total display width of all spans concatenated.
func spansWidth(spans []Span) int {
	w := 0
	for _, s := range spans {
		w += stringWidth(s.Text)
	}
	return w
}

// hoverStyleEqual compares two *Style pointers for equality.
// Both nil = equal. One nil = not equal. Otherwise compare visual fields.
func hoverStyleEqual(a, b *Style) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Foreground == b.Foreground &&
		a.Background == b.Background &&
		a.Bold == b.Bold &&
		a.Dim == b.Dim &&
		a.Underline == b.Underline &&
		a.Italic == b.Italic &&
		a.Strikethrough == b.Strikethrough &&
		a.Inverse == b.Inverse &&
		a.Border == b.Border &&
		a.BorderColor == b.BorderColor
}
