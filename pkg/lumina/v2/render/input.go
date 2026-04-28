package render

import "github.com/akzj/go-lua/pkg/lua"

// FocusedNode returns the currently focused node.
func (e *Engine) FocusedNode() *Node { return e.focusedNode }

// SetFocusedNode sets the currently focused node.
func (e *Engine) SetFocusedNode(n *Node) { e.focusedNode = n }

// FocusNext cycles focus to the next focusable node (input/textarea).
// If nothing is focused, focuses the first focusable node.
func (e *Engine) FocusNext() {
	if e.root == nil || e.root.RootNode == nil {
		return
	}

	focusable := collectFocusable(e.root.RootNode)
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
	old := e.focusedNode
	e.focusedNode = focusable[nextIdx]

	if old != nil {
		old.PaintDirty = true
	}
	e.focusedNode.PaintDirty = true
	e.needsRender = true
}

// FocusAutoFocus focuses the first node with AutoFocus=true.
// Called after initial render.
func (e *Engine) FocusAutoFocus() {
	if e.root == nil || e.root.RootNode == nil {
		return
	}
	node := findAutoFocus(e.root.RootNode)
	if node != nil {
		e.focusedNode = node
	}
}

// HandleInputKeyDown handles a keydown event on the focused input/textarea.
// Returns true if the event was consumed by the input system.
func (e *Engine) HandleInputKeyDown(key string) bool {
	node := e.focusedNode
	if node == nil {
		return false
	}
	if node.Type != "input" && node.Type != "textarea" {
		return false
	}

	switch key {
	case "Tab":
		// Tab cycles focus — don't consume, let caller handle
		return false

	case "Backspace":
		if len(node.Content) > 0 && node.CursorPos > 0 {
			runes := []rune(node.Content)
			if node.CursorPos > len(runes) {
				node.CursorPos = len(runes)
			}
			runes = append(runes[:node.CursorPos-1], runes[node.CursorPos:]...)
			node.Content = string(runes)
			node.CursorPos--
			node.PaintDirty = true
			e.needsRender = true
			e.fireOnChange(node)
		}
		return true

	case "Enter":
		if node.Type == "textarea" {
			// Insert newline
			runes := []rune(node.Content)
			if node.CursorPos > len(runes) {
				node.CursorPos = len(runes)
			}
			runes = append(runes[:node.CursorPos], append([]rune{'\n'}, runes[node.CursorPos:]...)...)
			node.Content = string(runes)
			node.CursorPos++
			node.PaintDirty = true
			e.needsRender = true
			e.fireOnChange(node)
			return true
		}
		// For input, Enter fires onSubmit (via onChange with current value)
		// The test expects the onChange callback to be called
		e.fireOnChange(node)
		return true

	case "ArrowLeft", "Left":
		if node.CursorPos > 0 {
			node.CursorPos--
			node.PaintDirty = true
			e.needsRender = true
		}
		return true

	case "ArrowRight", "Right":
		runes := []rune(node.Content)
		if node.CursorPos < len(runes) {
			node.CursorPos++
			node.PaintDirty = true
			e.needsRender = true
		}
		return true

	case "ArrowUp", "Up":
		if node.Type == "textarea" {
			e.moveCursorVertical(node, -1)
			return true
		}
		return false

	case "ArrowDown", "Down":
		if node.Type == "textarea" {
			e.moveCursorVertical(node, 1)
			return true
		}
		return false

	default:
		// Single printable character
		if len(key) == 1 && key[0] >= 0x20 && key[0] <= 0x7E {
			runes := []rune(node.Content)
			if node.CursorPos > len(runes) {
				node.CursorPos = len(runes)
			}
			ch := []rune(key)
			runes = append(runes[:node.CursorPos], append(ch, runes[node.CursorPos:]...)...)
			node.Content = string(runes)
			node.CursorPos++
			node.PaintDirty = true
			e.needsRender = true
			e.fireOnChange(node)
			return true
		}
		return false
	}
}

// fireOnChange calls the node's onChange handler with the current content.
func (e *Engine) fireOnChange(node *Node) {
	if node.OnChange == 0 {
		return
	}
	L := e.L
	L.RawGetI(lua.RegistryIndex, node.OnChange)
	if !L.IsFunction(-1) {
		L.Pop(1)
		return
	}
	L.PushString(node.Content)
	if status := L.PCall(1, 0, 0); status != lua.OK {
		L.Pop(1) // pop error message to prevent stack pollution
	}

	// Mark the component owning this node dirty so it re-renders
	e.markOwnerDirty(node)
}

// markOwnerDirty finds the component that owns this node and marks it dirty.
func (e *Engine) markOwnerDirty(node *Node) {
	for n := node; n != nil; n = n.Parent {
		if n.Component != nil {
			n.Component.Dirty = true
			e.needsRender = true
			return
		}
	}
}

// moveCursorVertical moves the cursor up/down lines in a textarea.
func (e *Engine) moveCursorVertical(node *Node, direction int) {
	runes := []rune(node.Content)
	if len(runes) == 0 {
		return
	}

	// Find current line and column
	line, col := 0, 0
	for i := 0; i < node.CursorPos && i < len(runes); i++ {
		if runes[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	targetLine := line + direction
	if targetLine < 0 {
		return
	}

	// Find the start of the target line
	currentLine := 0
	lineStart := 0
	for i, r := range runes {
		if currentLine == targetLine {
			lineStart = i
			break
		}
		if r == '\n' {
			currentLine++
			if currentLine == targetLine {
				lineStart = i + 1
			}
		}
	}
	if currentLine < targetLine {
		return // target line doesn't exist
	}

	// Find the position at the same column in the target line
	newPos := lineStart
	targetCol := col
	for i := lineStart; i < len(runes) && runes[i] != '\n' && (i-lineStart) < targetCol; i++ {
		newPos = i + 1
	}

	node.CursorPos = newPos
	node.PaintDirty = true
	e.needsRender = true
}

// collectFocusable walks the tree and collects all focusable nodes (input/textarea).
func collectFocusable(node *Node) []*Node {
	if node == nil {
		return nil
	}
	var result []*Node
	if node.Type == "input" || node.Type == "textarea" {
		result = append(result, node)
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
	if node.AutoFocus && (node.Type == "input" || node.Type == "textarea") {
		return node
	}
	for _, child := range node.Children {
		if found := findAutoFocus(child); found != nil {
			return found
		}
	}
	return nil
}
