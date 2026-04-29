package render

import (
	"fmt"
	"testing"
)

// BenchmarkReconciler_NoChange measures reconciler cost when nothing changed.
func BenchmarkReconciler_NoChange(b *testing.B) {
	// Build a tree with 200 nodes (10 rows × 20 cols)
	root := buildBenchTree(10, 20)
	desc := nodeToDescriptor(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reconcile(root, desc)
	}
}

// BenchmarkReconciler_ContentChange measures reconciler cost when one leaf changes content.
func BenchmarkReconciler_ContentChange(b *testing.B) {
	root := buildBenchTree(10, 20)
	desc := nodeToDescriptor(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate content on one leaf to force a diff
		leaf := desc.Children[5].Children[10]
		if i%2 == 0 {
			leaf.Content = "X"
		} else {
			leaf.Content = "."
		}
		desc.Children[5].Children[10] = leaf
		Reconcile(root, desc)
	}
}

// BenchmarkReconciler_StyleChange measures reconciler cost when styles change.
func BenchmarkReconciler_StyleChange(b *testing.B) {
	root := buildBenchTree(10, 20)
	desc := nodeToDescriptor(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Toggle background on a row
		row := &desc.Children[i%10]
		if i%2 == 0 {
			row.Style.Background = "#313244"
		} else {
			row.Style.Background = ""
		}
		Reconcile(root, desc)
	}
}

// BenchmarkReconciler_ChildReorder measures reconciler cost for child list reorder.
func BenchmarkReconciler_ChildReorder(b *testing.B) {
	root := buildBenchTree(10, 20)
	desc := nodeToDescriptor(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Swap first and last row
		desc.Children[0], desc.Children[9] = desc.Children[9], desc.Children[0]
		Reconcile(root, desc)
		// Swap back for next iteration
		desc.Children[0], desc.Children[9] = desc.Children[9], desc.Children[0]
	}
}

// BenchmarkReconciler_LargeTree benchmarks reconciler on a 50×40 (2000 node) tree.
func BenchmarkReconciler_LargeTree(b *testing.B) {
	root := buildBenchTree(50, 40)
	desc := nodeToDescriptor(root)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Change one leaf
		desc.Children[25].Children[20].Content = fmt.Sprintf("%d", i)
		Reconcile(root, desc)
	}
}

// --- helpers ---

func buildBenchTree(rows, cols int) *Node {
	root := &Node{
		ID:   "root",
		Type: "vbox",
		Style: Style{
			Width:  cols,
			Height: rows,
		},
	}
	for r := 0; r < rows; r++ {
		row := &Node{
			ID:   fmt.Sprintf("row-%d", r),
			Type: "hbox",
			Style: Style{
				Width:  cols,
				Height: 1,
			},
		}
		for c := 0; c < cols; c++ {
			cell := &Node{
				ID:      fmt.Sprintf("cell-%d-%d", r, c),
				Type:    "text",
				Content: ".",
				Key:     fmt.Sprintf("cell-%d-%d", r, c),
				Style: Style{
					Width:  1,
					Height: 1,
				},
			}
			row.Children = append(row.Children, cell)
		}
		root.Children = append(root.Children, row)
	}
	return root
}

func nodeToDescriptor(n *Node) Descriptor {
	d := Descriptor{
		ID:      n.ID,
		Type:    n.Type,
		Content: n.Content,
		Key:     n.Key,
		Style:   n.Style,
	}
	for _, child := range n.Children {
		d.Children = append(d.Children, nodeToDescriptor(child))
	}
	return d
}
