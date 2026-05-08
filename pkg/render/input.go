package render

// FocusedNode returns the currently focused node.
func (e *Engine) FocusedNode() *Node { return e.focusedNode }

// SetFocusedNode sets the currently focused node.
func (e *Engine) SetFocusedNode(n *Node) { e.focusedNode = n }

// setFocus changes focus from the current node to newNode, firing onBlur/onFocus.
func (e *Engine) setFocus(newNode *Node) {
	old := e.focusedNode
	if old == newNode {
		return
	}

	// Blur old
	if old != nil && !old.Removed {
		old.Focused = false
		old.PaintDirty = true
		blurRef := old.OnBlur
		// For component placeholders: search children for onBlur handler
		if blurRef == 0 && old.Type == "component" && len(old.Children) > 0 {
			if inner := findFocusHandlerInSubtree(old, "blur"); inner != nil {
				blurRef = inner.OnBlur
			}
		}
		if blurRef != 0 {
			e.callLuaRefSimple(blurRef)
		}
	}

	e.focusedNode = newNode

	// Focus new
	if newNode != nil {
		newNode.Focused = true
		newNode.PaintDirty = true
		focusRef := newNode.OnFocus
		// For component placeholders: search children for onFocus handler
		if focusRef == 0 && newNode.Type == "component" && len(newNode.Children) > 0 {
			if inner := findFocusHandlerInSubtree(newNode, "focus"); inner != nil {
				focusRef = inner.OnFocus
			}
		}
		if focusRef != 0 {
			e.callLuaRefSimple(focusRef)
		}
	}

	e.needsRender = true
}

// findFocusHandlerInSubtree finds the first node with the specified handler in the subtree.
// Used for component placeholders where onFocus/onBlur are on the rendered child.
func findFocusHandlerInSubtree(node *Node, handlerType string) *Node {
	if node == nil {
		return nil
	}
	for _, child := range node.Children {
		switch handlerType {
		case "focus":
			if child.OnFocus != 0 {
				return child
			}
		case "blur":
			if child.OnBlur != 0 {
				return child
			}
		}
		if found := findFocusHandlerInSubtree(child, handlerType); found != nil {
			return found
		}
	}
	return nil
}

// FocusNext cycles focus to the next focusable node.
// If nothing is focused, focuses the first focusable node.
func (e *Engine) FocusNext() {
	if len(e.layers) == 0 {
		return
	}

	// Find active root: topmost modal layer, or main layer
	activeRoot := e.layers[0].Root
	for i := len(e.layers) - 1; i >= 0; i-- {
		if e.layers[i].Root != nil && e.layers[i].Modal {
			activeRoot = e.layers[i].Root
			break
		}
	}
	focusable := collectFocusable(activeRoot)
	if len(focusable) == 0 {
		return
	}

	// Find current index
	idx := -1
	for i, n := range focusable {
		if n == e.focusedNode {
			idx = i
			break
		}
	}

	// Advance to next (wrap around)
	nextIdx := (idx + 1) % len(focusable)
	e.setFocus(focusable[nextIdx])
}

// FocusPrev cycles focus to the previous focusable node.
// If nothing is focused, focuses the last focusable node.
func (e *Engine) FocusPrev() {
	if len(e.layers) == 0 {
		return
	}

	// Find active root: topmost modal layer, or main layer
	activeRoot := e.layers[0].Root
	for i := len(e.layers) - 1; i >= 0; i-- {
		if e.layers[i].Root != nil && e.layers[i].Modal {
			activeRoot = e.layers[i].Root
			break
		}
	}
	focusable := collectFocusable(activeRoot)
	if len(focusable) == 0 {
		return
	}

	// Find current index
	idx := -1
	for i, n := range focusable {
		if n == e.focusedNode {
			idx = i
			break
		}
	}

	// Go to previous (wrap around)
	prevIdx := idx - 1
	if prevIdx < 0 {
		prevIdx = len(focusable) - 1
	}
	e.setFocus(focusable[prevIdx])
}

// FocusByID finds a node by ID and focuses it, firing blur/focus callbacks.
// Returns true if the node was found and focused.
func (e *Engine) FocusByID(id string) bool {
	node := e.FindNodeByID(id)
	if node == nil {
		return false
	}
	// If node is directly focusable, focus it
	if node.Focusable && !node.Disabled {
		e.setFocus(node)
		return true
	}
	// If it's a non-focusable component placeholder (e.g. Textarea component
	// where the caller didn't pass focusable=true), search its children for
	// the first focusable node (the rendered inner vbox with onKeyDown).
	if node.Type == "component" {
		if inner := findFirstFocusable(node); inner != nil {
			e.setFocus(inner)
			return true
		}
	}
	return false
}

// findFirstFocusable finds the first focusable, non-disabled node in a subtree (DFS).
func findFirstFocusable(node *Node) *Node {
	if node == nil {
		return nil
	}
	for _, child := range node.Children {
		if child.Focusable && !child.Disabled {
			return child
		}
		if found := findFirstFocusable(child); found != nil {
			return found
		}
	}
	return nil
}

// FindNodeByID searches all layers for a node with the given ID.
func (e *Engine) FindNodeByID(id string) *Node {
	if id == "" {
		return nil
	}
	for _, layer := range e.layers {
		if layer.Root != nil {
			if found := findNodeByID(layer.Root, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func findNodeByID(node *Node, id string) *Node {
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := findNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// FocusableIDs returns the ordered list of focusable node IDs in the active layer.
func (e *Engine) FocusableIDs() []string {
	if len(e.layers) == 0 {
		return nil
	}
	activeRoot := e.layers[0].Root
	for i := len(e.layers) - 1; i >= 0; i-- {
		if e.layers[i].Root != nil && e.layers[i].Modal {
			activeRoot = e.layers[i].Root
			break
		}
	}
	focusable := collectFocusable(activeRoot)
	ids := make([]string, 0, len(focusable))
	for _, n := range focusable {
		if n.ID != "" {
			ids = append(ids, n.ID)
		}
	}
	return ids
}


// FocusAutoFocus focuses the first node with AutoFocus=true.
// Called after initial render.
func (e *Engine) FocusAutoFocus() {
	if len(e.layers) == 0 {
		return
	}
	// Search all layers for autoFocus node (topmost first)
	for i := len(e.layers) - 1; i >= 0; i-- {
		if e.layers[i].Root != nil {
			node := findAutoFocus(e.layers[i].Root)
			if node != nil {
				e.setFocus(node)
				return
			}
		}
	}
}

// focusAutoFocusOnly steals focus ONLY if an autoFocus=true node exists.
// Unlike FocusAutoFocus, this has no fallback — if no autoFocus node is found,
// the current focus is left as-is.
func (e *Engine) focusAutoFocusOnly() {
	for i := len(e.layers) - 1; i >= 0; i-- {
		if e.layers[i].Root != nil {
			node := findAutoFocus(e.layers[i].Root)
			if node != nil {
				e.setFocus(node)
				return
			}
		}
	}
	// No autoFocus node found — keep current focus as-is
}

// collectFocusable walks the tree and collects all focusable, non-disabled nodes.
func collectFocusable(node *Node) []*Node {
	if node == nil {
		return nil
	}
	var result []*Node
	if node.Focusable && !node.Disabled {
		result = append(result, node)
		// If this is a component placeholder, don't recurse into rendered children.
		// The component's inner focusable nodes should not appear separately in the
		// focus list — key dispatch handles reaching them via findKeyHandlerInSubtree.
		if node.Type == "component" {
			return result
		}
	}
	for _, child := range node.Children {
		result = append(result, collectFocusable(child)...)
	}
	return result
}

// findAutoFocus walks the tree and returns the first node with AutoFocus=true.
func findAutoFocus(node *Node) *Node {
	if node == nil {
		return nil
	}
	// Skip hidden subtrees (display:none)
	if node.Style.Display == "none" {
		return nil
	}
	if node.AutoFocus && node.Focusable {
		return node
	}
	for _, child := range node.Children {
		if found := findAutoFocus(child); found != nil {
			return found
		}
	}
	return nil
}
