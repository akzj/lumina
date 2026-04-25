// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import "fmt"

// PatchType identifies the kind of VNode tree mutation.
type PatchType int

const (
	// PatchReplace replaces an entire subtree.
	PatchReplace PatchType = iota
	// PatchUpdate updates props/style of an existing node.
	PatchUpdate
	// PatchInsert inserts a new child at a given index.
	PatchInsert
	// PatchRemove removes a child at a given index.
	PatchRemove
	// PatchText updates text content only.
	PatchText
	// PatchReorder reorders children (keyed reconciliation).
	PatchReorder
)

// String returns a human-readable name for a PatchType.
func (p PatchType) String() string {
	switch p {
	case PatchReplace:
		return "Replace"
	case PatchUpdate:
		return "Update"
	case PatchInsert:
		return "Insert"
	case PatchRemove:
		return "Remove"
	case PatchText:
		return "Text"
	case PatchReorder:
		return "Reorder"
	default:
		return fmt.Sprintf("PatchType(%d)", int(p))
	}
}

// Patch represents a single mutation to apply to a VNode tree.
type Patch struct {
	Type    PatchType
	Path    []int  // child-index path from root to the target node
	OldNode *VNode // previous node (nil for Insert)
	NewNode *VNode // new node (nil for Remove)
	Index   int    // child index for Insert/Remove/Reorder
	// MoveFrom is the original index before reorder (only for PatchReorder).
	MoveFrom int
}

// DiffVNode compares old and new VNode trees and returns a minimal patch set.
// Either or both nodes may be nil.
func DiffVNode(oldNode, newNode *VNode) []Patch {
	var patches []Patch
	diffVNode(oldNode, newNode, nil, &patches)
	return patches
}

// diffVNode is the recursive core of the diff algorithm.
func diffVNode(oldNode, newNode *VNode, path []int, patches *[]Patch) {
	// Both nil → nothing to do.
	if oldNode == nil && newNode == nil {
		return
	}

	// Old nil, new exists → insert at root (replace).
	if oldNode == nil {
		*patches = append(*patches, Patch{
			Type:    PatchReplace,
			Path:    copyPath(path),
			NewNode: newNode,
		})
		return
	}

	// Old exists, new nil → remove.
	if newNode == nil {
		*patches = append(*patches, Patch{
			Type:    PatchRemove,
			Path:    copyPath(path),
			OldNode: oldNode,
		})
		return
	}

	// Different types → replace entire subtree.
	if oldNode.Type != newNode.Type {
		*patches = append(*patches, Patch{
			Type:    PatchReplace,
			Path:    copyPath(path),
			OldNode: oldNode,
			NewNode: newNode,
		})
		return
	}

	// Same type — check for prop/style/content changes.
	propsChanged := !propsEqual(oldNode.Props, newNode.Props)
	if propsChanged {
		*patches = append(*patches, Patch{
			Type:    PatchUpdate,
			Path:    copyPath(path),
			OldNode: oldNode,
			NewNode: newNode,
		})
	}

	// Text content change.
	if oldNode.Content != newNode.Content {
		*patches = append(*patches, Patch{
			Type:    PatchText,
			Path:    copyPath(path),
			OldNode: oldNode,
			NewNode: newNode,
		})
	}

	// Diff children.
	diffChildren(oldNode.Children, newNode.Children, path, patches)
}

// diffChildren compares two children slices using keyed or index-based reconciliation.
func diffChildren(oldChildren, newChildren []*VNode, parentPath []int, patches *[]Patch) {
	// If any children have keys, use keyed reconciliation.
	if hasKeys(oldChildren) || hasKeys(newChildren) {
		diffKeyedChildren(oldChildren, newChildren, parentPath, patches)
		return
	}

	// Index-based comparison.
	maxLen := len(oldChildren)
	if len(newChildren) > maxLen {
		maxLen = len(newChildren)
	}

	for i := 0; i < maxLen; i++ {
		childPath := append(copyPath(parentPath), i)
		if i < len(oldChildren) && i < len(newChildren) {
			diffVNode(oldChildren[i], newChildren[i], childPath, patches)
		} else if i >= len(oldChildren) {
			// New child — insert.
			*patches = append(*patches, Patch{
				Type:    PatchInsert,
				Path:    copyPath(parentPath),
				NewNode: newChildren[i],
				Index:   i,
			})
		} else {
			// Old child — remove.
			*patches = append(*patches, Patch{
				Type:    PatchRemove,
				Path:    copyPath(parentPath),
				OldNode: oldChildren[i],
				Index:   i,
			})
		}
	}
}

// diffKeyedChildren performs keyed reconciliation of children.
// It builds a map of old key→index, then walks new children to detect
// inserts, removes, moves, and in-place updates.
func diffKeyedChildren(oldChildren, newChildren []*VNode, parentPath []int, patches *[]Patch) {
	oldKeyMap := make(map[string]int, len(oldChildren))
	for i, child := range oldChildren {
		if key := childKey(child); key != "" {
			oldKeyMap[key] = i
		}
	}

	newKeyMap := make(map[string]int, len(newChildren))
	for i, child := range newChildren {
		if key := childKey(child); key != "" {
			newKeyMap[key] = i
		}
	}

	// Track which old children have been matched.
	matched := make(map[int]bool, len(oldChildren))

	for newIdx, newChild := range newChildren {
		key := childKey(newChild)
		if key == "" {
			// Unkeyed child in a keyed list — treat by index.
			childPath := append(copyPath(parentPath), newIdx)
			if newIdx < len(oldChildren) {
				diffVNode(oldChildren[newIdx], newChild, childPath, patches)
				matched[newIdx] = true
			} else {
				*patches = append(*patches, Patch{
					Type:    PatchInsert,
					Path:    copyPath(parentPath),
					NewNode: newChild,
					Index:   newIdx,
				})
			}
			continue
		}

		oldIdx, found := oldKeyMap[key]
		if !found {
			// Key not in old → insert.
			*patches = append(*patches, Patch{
				Type:    PatchInsert,
				Path:    copyPath(parentPath),
				NewNode: newChild,
				Index:   newIdx,
			})
			continue
		}

		matched[oldIdx] = true

		// Key exists in both — diff the pair.
		childPath := append(copyPath(parentPath), newIdx)
		diffVNode(oldChildren[oldIdx], newChild, childPath, patches)

		// If position changed → emit reorder.
		if oldIdx != newIdx {
			*patches = append(*patches, Patch{
				Type:     PatchReorder,
				Path:     copyPath(parentPath),
				OldNode:  oldChildren[oldIdx],
				NewNode:  newChild,
				Index:    newIdx,
				MoveFrom: oldIdx,
			})
		}
	}

	// Old children not matched → remove.
	for oldIdx, oldChild := range oldChildren {
		if matched[oldIdx] {
			continue
		}
		key := childKey(oldChild)
		if key != "" {
			if _, inNew := newKeyMap[key]; inNew {
				continue // already handled above
			}
		}
		*patches = append(*patches, Patch{
			Type:    PatchRemove,
			Path:    copyPath(parentPath),
			OldNode: oldChild,
			Index:   oldIdx,
		})
	}
}

// ApplyPatches applies a set of patches to a Frame via incremental re-render.
// It re-renders only the affected subtrees by walking the patch paths.
// The root VNode should already have layout computed before calling this.
// Dirty rects are added to the frame for each affected region.
func ApplyPatches(frame *Frame, root *VNode, patches []Patch, width, height int) {
	if len(patches) == 0 {
		return
	}

	// Clear the default full-frame dirty rect — we'll add precise ones
	frame.DirtyRects = frame.DirtyRects[:0]

	// For each patch, find the affected VNode and re-render its region.
	for _, p := range patches {
		switch p.Type {
		case PatchRemove:
			// Clear the old node's region.
			if p.OldNode != nil {
				clearRegion(frame, p.OldNode)
				frame.AddDirtyRect(p.OldNode.X, p.OldNode.Y, p.OldNode.W, p.OldNode.H)
			}

		case PatchReplace:
			// Clear old region first, then render new.
			if p.OldNode != nil {
				clearRegion(frame, p.OldNode)
				frame.AddDirtyRect(p.OldNode.X, p.OldNode.Y, p.OldNode.W, p.OldNode.H)
			}
			if p.NewNode != nil {
				clip := nodeClipRect(p.NewNode, frame)
				renderVNode(frame, p.NewNode, clip)
				frame.AddDirtyRect(p.NewNode.X, p.NewNode.Y, p.NewNode.W, p.NewNode.H)
			}

		case PatchUpdate, PatchText:
			// Clear old area first (content may have shrunk), then re-render.
			if p.OldNode != nil {
				clearRegion(frame, p.OldNode)
				frame.AddDirtyRect(p.OldNode.X, p.OldNode.Y, p.OldNode.W, p.OldNode.H)
			}
			if p.NewNode != nil {
				clip := nodeClipRect(p.NewNode, frame)
				renderVNode(frame, p.NewNode, clip)
				frame.AddDirtyRect(p.NewNode.X, p.NewNode.Y, p.NewNode.W, p.NewNode.H)
			}

		case PatchInsert:
			// Render the new child — it already has layout from the full tree relayout.
			if p.NewNode != nil {
				clip := nodeClipRect(p.NewNode, frame)
				renderVNode(frame, p.NewNode, clip)
				frame.AddDirtyRect(p.NewNode.X, p.NewNode.Y, p.NewNode.W, p.NewNode.H)
			}

		case PatchReorder:
			// Reorder is handled by the full tree relayout; individual
			// patches are already emitted for content changes.
		}
	}
}

// nodeClipRect returns the clip rect for a VNode, bounded by the frame.
func nodeClipRect(node *VNode, frame *Frame) Rect {
	r := Rect{X: node.X, Y: node.Y, W: node.W, H: node.H}
	// Clamp to frame bounds
	if r.X < 0 {
		r.W += r.X
		r.X = 0
	}
	if r.Y < 0 {
		r.H += r.Y
		r.Y = 0
	}
	if r.X+r.W > frame.Width {
		r.W = frame.Width - r.X
	}
	if r.Y+r.H > frame.Height {
		r.H = frame.Height - r.Y
	}
	return r
}

// ShouldFullRerender returns true when the patch set is large enough that
// a full re-render is cheaper than incremental application.
// Threshold: if more than 50% of total nodes are patched, prefer full re-render.
func ShouldFullRerender(patches []Patch, root *VNode) bool {
	if root == nil {
		return true
	}
	total := countNodes(root)
	if total == 0 {
		return true
	}
	return len(patches) > total/2
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

// childKey returns the "key" prop of a VNode (empty string if absent).
func childKey(v *VNode) string {
	if v == nil {
		return ""
	}
	if k, ok := v.Props["key"].(string); ok {
		return k
	}
	// Also accept int64 keys (common from Lua).
	if k, ok := v.Props["key"].(int64); ok {
		return fmt.Sprintf("%d", k)
	}
	return ""
}

// hasKeys returns true if any child in the slice has a "key" prop.
func hasKeys(children []*VNode) bool {
	for _, c := range children {
		if childKey(c) != "" {
			return true
		}
	}
	return false
}

// propsEqual performs a shallow comparison of two prop maps.
// It ignores the "children" and "key" keys (handled separately).
func propsEqual(a, b map[string]any) bool {
	// Fast path: same length check.
	skipKeys := map[string]bool{"children": true, "key": true}

	countA := 0
	for k := range a {
		if !skipKeys[k] {
			countA++
		}
	}
	countB := 0
	for k := range b {
		if !skipKeys[k] {
			countB++
		}
	}
	if countA != countB {
		return false
	}

	for k, va := range a {
		if skipKeys[k] {
			continue
		}
		vb, ok := b[k]
		if !ok {
			return false
		}
		if !shallowEqual(va, vb) {
			return false
		}
	}
	return true
}

// shallowEqual compares two values with == semantics for primitive types
// and identity comparison for maps/slices.
func shallowEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// For comparable primitive types, use ==.
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case int:
		bv, ok := b.(int)
		return ok && av == bv
	case int64:
		bv, ok := b.(int64)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	default:
		// For non-comparable types (maps, slices), identity check via fmt.
		return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
	}
}

// copyPath returns a copy of a path slice (to avoid aliasing).
func copyPath(path []int) []int {
	if path == nil {
		return nil
	}
	cp := make([]int, len(path))
	copy(cp, path)
	return cp
}

// countNodes returns the total number of nodes in a VNode tree.
func countNodes(v *VNode) int {
	if v == nil {
		return 0
	}
	n := 1
	for _, c := range v.Children {
		n += countNodes(c)
	}
	return n
}

// clearRegion fills the region occupied by a VNode with empty transparent cells.
func clearRegion(frame *Frame, node *VNode) {
	if node == nil || frame == nil {
		return
	}
	clearRect(frame, node.X, node.Y, node.W, node.H)
}

// clearRect fills a rectangular region with empty transparent cells.
func clearRect(frame *Frame, x, y, w, h int) {
	if frame == nil {
		return
	}
	for cy := y; cy < y+h && cy < frame.Height; cy++ {
		if cy < 0 {
			continue
		}
		for cx := x; cx < x+w && cx < frame.Width; cx++ {
			if cx < 0 {
				continue
			}
			frame.Cells[cy][cx] = Cell{Char: ' ', Transparent: true}
		}
	}
}
