// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"fmt"
	"reflect"
)

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
	// Build key maps, detecting duplicates
	dupKeys := make(map[string]bool)
	oldKeyMap := make(map[string]int, len(oldChildren))
	for i, child := range oldChildren {
		if key := childKey(child); key != "" {
			if _, exists := oldKeyMap[key]; exists {
				dupKeys[key] = true // duplicate key — skip keyed reconciliation for this key
			}
			oldKeyMap[key] = i
		}
	}

	newKeyMap := make(map[string]int, len(newChildren))
	for i, child := range newChildren {
		if key := childKey(child); key != "" {
			if _, exists := newKeyMap[key]; exists {
				dupKeys[key] = true
			}
			newKeyMap[key] = i
		}
	}

	// Remove duplicate keys from maps — fall back to index-based for those
	for key := range dupKeys {
		delete(oldKeyMap, key)
		delete(newKeyMap, key)
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
// ApplyPatches applies VNode diff patches to a frame by re-rendering affected parent containers.
//
// LIMITATION: This function re-renders the full parent container of each patched node,
// not just the patched node itself. This is necessary because layout changes in one child
// (e.g., content size change) can shift sibling positions. A more granular approach would
// require incremental layout, which is not yet implemented. For most TUI use cases,
// re-rendering the parent container is fast enough.
func ApplyPatches(frame *Frame, root *VNode, patches []Patch, width, height int) {
	if len(patches) == 0 {
		return
	}

	// Clear the default full-frame dirty rect — we'll add precise ones
	frame.DirtyRects = frame.DirtyRects[:0]

	// Approach A: when a child has a patch, re-render its parent container.
	// This ensures sibling nodes whose positions shifted (due to content size
	// changes in a sibling) are also re-rendered, even if they have no patch.
	//
	// We collect unique parent containers by their region key to avoid
	// re-rendering the same container multiple times.
	type parentRegion struct {
		node *VNode
		// Also track old nodes for clearing stale regions
		oldNodes []*VNode
	}
	rerendered := make(map[*VNode]*parentRegion)

	for _, p := range patches {
		// Find the parent container in the NEW tree via path
		parent := findParentContainer(root, p.Path)

		pr, exists := rerendered[parent]
		if !exists {
			pr = &parentRegion{node: parent}
			rerendered[parent] = pr
		}
		if p.OldNode != nil {
			pr.oldNodes = append(pr.oldNodes, p.OldNode)
		}
	}

	// Re-render each unique parent container
	for _, pr := range rerendered {
		parent := pr.node

		// Clear old node regions that may not overlap with new layout
		for _, old := range pr.oldNodes {
			clearRegion(frame, old)
			frame.AddDirtyRect(old.X, old.Y, old.W, old.H)
		}

		// Clear the parent's region
		clearRect(frame, parent.X, parent.Y, parent.W, parent.H)

		// Re-render the entire parent (including all children)
		clip := nodeClipRect(parent, frame)
		renderVNode(frame, parent, clip)
		frame.AddDirtyRect(parent.X, parent.Y, parent.W, parent.H)
	}
}

// findParentContainer walks the VNode tree via the patch path and returns
// the parent container (one level up from the patched node). If the path
// is empty or has only one element, returns root.
func findParentContainer(root *VNode, path []int) *VNode {
	if len(path) <= 1 {
		return root
	}
	// Walk path up to second-to-last element to find the parent
	node := root
	for _, idx := range path[:len(path)-1] {
		if idx < len(node.Children) {
			node = node.Children[idx]
		} else {
			return root
		}
	}
	return node
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
	case LuaFuncRef:
		bv, ok := b.(LuaFuncRef)
		return ok && av.Ref == bv.Ref
	default:
		// For non-comparable types (maps, slices), use reflect.DeepEqual.
		return reflect.DeepEqual(a, b)
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
