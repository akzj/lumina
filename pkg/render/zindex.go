package render

import "sort"

// hasNonZeroZIndex returns true if any child has a non-zero ZIndex.
// Used as a fast-path check to avoid allocation when no z-index is set.
func hasNonZeroZIndex(children []*Node) bool {
	for _, ch := range children {
		if ch != nil && ch.Style.ZIndex != 0 {
			return true
		}
	}
	return false
}

// zSortedChildren returns children sorted by ZIndex ascending (stable).
// If no child has a non-zero ZIndex, returns nil (caller should use original slice).
// The returned slice is a copy — the original Children slice is never modified.
func zSortedChildren(children []*Node) []*Node {
	if !hasNonZeroZIndex(children) {
		return nil
	}
	sorted := make([]*Node, len(children))
	copy(sorted, children)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Style.ZIndex < sorted[j].Style.ZIndex
	})
	return sorted
}

// paintOrderChildren returns children in paint order (ascending ZIndex, stable).
// If no child has non-zero ZIndex, returns the original slice (zero allocation).
func paintOrderChildren(children []*Node) []*Node {
	if sorted := zSortedChildren(children); sorted != nil {
		return sorted
	}
	return children
}

// hitTestOrderChildren returns children in hit-test order: highest ZIndex first,
// and for equal ZIndex, last-in-array first (reverse of paint order).
// If no child has non-zero ZIndex, returns nil (caller should use reverse iteration).
func hitTestOrderChildren(children []*Node) []*Node {
	if !hasNonZeroZIndex(children) {
		return nil
	}
	// Sort ascending by ZIndex (stable), then reverse the whole slice.
	// This gives: highest ZIndex first; for equal ZIndex, last-in-array first.
	sorted := make([]*Node, len(children))
	copy(sorted, children)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Style.ZIndex < sorted[j].Style.ZIndex
	})
	// Reverse
	for i, j := 0, len(sorted)-1; i < j; i, j = i+1, j-1 {
		sorted[i], sorted[j] = sorted[j], sorted[i]
	}
	return sorted
}
