package render

// Layer is an independent rendering layer.
// Each Layer owns a Node tree that is laid out and painted independently.
// Layers are stacked: later layers paint over earlier ones.
type Layer struct {
	ID    string
	Root  *Node
	Modal bool // true = blocks events from reaching layers below
}

// markOverlappingDirty marks all nodes in the tree that overlap with the given rect as PaintDirty.
func markOverlappingDirty(node *Node, rx, ry, rw, rh int) {
	if node == nil || node.W <= 0 || node.H <= 0 {
		return
	}
	// Check if node overlaps with rect
	if node.X < rx+rw && node.X+node.W > rx && node.Y < ry+rh && node.Y+node.H > ry {
		node.PaintDirty = true
	}
	for _, child := range node.Children {
		markOverlappingDirty(child, rx, ry, rw, rh)
	}
}
