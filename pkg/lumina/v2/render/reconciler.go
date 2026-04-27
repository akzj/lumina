package render

// Reconcile updates an existing Node tree to match a new Descriptor.
// It patches the node in-place, marking LayoutDirty/PaintDirty as needed.
// Returns true if any changes were made.
func Reconcile(node *Node, desc Descriptor) bool {
	changed := false

	// 1. Update content
	if desc.Content != node.Content {
		node.Content = desc.Content
		node.PaintDirty = true
		changed = true
	}

	// 1b. Update placeholder
	if desc.Placeholder != node.Placeholder {
		node.Placeholder = desc.Placeholder
		node.PaintDirty = true
		changed = true
	}

	// 1c. Update scroll position
	if desc.ScrollY != node.ScrollY {
		node.ScrollY = desc.ScrollY
		node.PaintDirty = true
		changed = true
	}

	// 2. Update ComponentType (for component placeholder nodes)
	if desc.ComponentType != node.ComponentType {
		node.ComponentType = desc.ComponentType
		changed = true
	}

	// 3. Update style (check each field that affects paint vs layout)
	if styleChanged := reconcileStyle(node, desc.Style); styleChanged {
		changed = true
	}

	// 4. Update event handlers (just swap refs)
	changed = updateRef(&node.OnClick, desc.OnClick) || changed
	changed = updateRef(&node.OnMouseEnter, desc.OnMouseEnter) || changed
	changed = updateRef(&node.OnMouseLeave, desc.OnMouseLeave) || changed
	changed = updateRef(&node.OnKeyDown, desc.OnKeyDown) || changed
	changed = updateRef(&node.OnChange, desc.OnChange) || changed
	changed = updateRef(&node.OnScroll, desc.OnScroll) || changed

	// 5. Reconcile children (skip for component nodes — children are grafted)
	if node.Type != "component" {
		if reconcileChildren(node, desc.Children) {
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
		old.Flex != newStyle.Flex ||
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
		old.Justify != newStyle.Justify ||
		old.Align != newStyle.Align ||
		old.Border != newStyle.Border ||
		old.Overflow != newStyle.Overflow ||
		old.Position != newStyle.Position ||
		old.Top != newStyle.Top ||
		old.Left != newStyle.Left ||
		old.Right != newStyle.Right ||
		old.Bottom != newStyle.Bottom

	node.Style = newStyle

	if layoutChanged {
		node.MarkLayoutDirty()
	}
	node.PaintDirty = true
	return true
}

// reconcileChildren matches new descriptors against existing children by key+type.
func reconcileChildren(parent *Node, descs []Descriptor) bool {
	oldChildren := parent.Children
	changed := false

	// Fast path: same length, same keys in order
	if len(oldChildren) == len(descs) && allKeysMatch(oldChildren, descs) {
		for i, desc := range descs {
			if Reconcile(oldChildren[i], desc) {
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
			oldByKey[key] = i
		}
	}

	// Match new descriptors to existing children
	newChildren := make([]*Node, 0, len(descs))
	usedOld := make(map[int]bool, len(oldChildren))

	for _, desc := range descs {
		key := descKey(desc)
		if idx, ok := oldByKey[key]; ok && oldChildren[idx].Type == desc.Type {
			// Reuse existing node
			usedOld[idx] = true
			Reconcile(oldChildren[idx], desc)
			newChildren = append(newChildren, oldChildren[idx])
		} else {
			// Create new node
			newNode := createNodeFromDesc(desc)
			newNode.Parent = parent
			newNode.PaintDirty = true
			newNode.LayoutDirty = true
			newChildren = append(newChildren, newNode)
			changed = true
		}
	}

	// Cleanup removed children
	for i, child := range oldChildren {
		if !usedOld[i] {
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

func updateRef(ref *LuaRef, newRef LuaRef) bool {
	if *ref != newRef {
		// TODO: unref old, ref new
		*ref = newRef
		return true
	}
	return false
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
	node.Placeholder = desc.Placeholder
	node.AutoFocus = desc.AutoFocus
	node.ScrollY = desc.ScrollY
	node.Style = desc.Style
	node.ComponentType = desc.ComponentType
	node.OnClick = desc.OnClick
	node.OnMouseEnter = desc.OnMouseEnter
	node.OnMouseLeave = desc.OnMouseLeave
	node.OnKeyDown = desc.OnKeyDown
	node.OnChange = desc.OnChange
	node.OnScroll = desc.OnScroll

	for _, childDesc := range desc.Children {
		child := createNodeFromDesc(childDesc)
		node.AddChild(child)
	}
	return node
}

// markRemovedRecursive sets Removed=true and Parent=nil on a node and all descendants.
func markRemovedRecursive(node *Node) {
	node.Removed = true
	node.Parent = nil
	for _, child := range node.Children {
		markRemovedRecursive(child)
	}
}
