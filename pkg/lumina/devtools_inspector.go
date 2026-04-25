// Package lumina — DevTools Element Inspector overlay.
// Provides a visual element inspector that renders as an overlay panel.
// Toggle with F12. Mouse hover highlights elements via Cell.OwnerNode.
package lumina

import (
	"fmt"
	"strings"
)

// DevToolsInspector manages the element inspector overlay.
type DevToolsInspector struct {
	enabled     bool
	highlightID string   // ID of element being highlighted (hover)
	selectedID  string   // ID of element selected for inspection
	scrollY     int      // scroll offset in element tree
	panelWidth  int      // width of inspector panel
}

// globalInspector is the singleton inspector.
var globalInspector = &DevToolsInspector{
	panelWidth: 40,
}

// ToggleInspector toggles the devtools inspector panel.
func ToggleInspector() {
	globalInspector.enabled = !globalInspector.enabled
	if globalInspector.enabled {
		globalDevTools.Enable()
	}
}

// IsInspectorVisible returns whether the inspector is visible.
func IsInspectorVisible() bool {
	return globalInspector.enabled
}

// SetInspectorHighlight sets the highlighted element (from mouse hover).
func SetInspectorHighlight(id string) {
	globalInspector.highlightID = id
}

// SetInspectorSelected sets the selected element for detailed inspection.
func SetInspectorSelected(id string) {
	globalInspector.selectedID = id
}

// BuildInspectorVNode creates the VNode for the devtools inspector panel.
// It renders on the right side of the screen as a vertical panel.
func BuildInspectorVNode(rootVNode *VNode, screenW, screenH int) *VNode {
	if !globalInspector.enabled {
		return nil
	}

	panelW := globalInspector.panelWidth
	if panelW > screenW/2 {
		panelW = screenW / 2
	}

	// Build element tree section
	treeLines := buildElementTree(rootVNode, 0)

	// Build style inspector section
	styleLines := buildStyleInspector(rootVNode)

	// Combine into panel
	var children []*VNode

	// Header
	children = append(children, &VNode{
		Type:    "text",
		Content: " 🔍 DevTools Inspector ",
		Props: map[string]any{
			"style": map[string]any{
				"bold":       true,
				"foreground": "#FFFFFF",
				"background": "#1E40AF",
				"height":     1,
			},
		},
	})

	// Element tree header
	children = append(children, &VNode{
		Type:    "text",
		Content: " ─── Element Tree ───",
		Props: map[string]any{
			"style": map[string]any{
				"foreground": "#93C5FD",
				"height":     1,
			},
		},
	})

	// Element tree lines (scrollable region)
	maxTreeLines := screenH/2 - 4
	startLine := globalInspector.scrollY
	if startLine > len(treeLines)-maxTreeLines {
		startLine = len(treeLines) - maxTreeLines
	}
	if startLine < 0 {
		startLine = 0
	}
	endLine := startLine + maxTreeLines
	if endLine > len(treeLines) {
		endLine = len(treeLines)
	}

	for _, line := range treeLines[startLine:endLine] {
		fg := "#D1D5DB"
		if line.id == globalInspector.selectedID {
			fg = "#FCD34D" // yellow for selected
		} else if line.id == globalInspector.highlightID {
			fg = "#34D399" // green for highlighted
		}
		children = append(children, &VNode{
			Type:    "text",
			Content: line.text,
			Props: map[string]any{
				"id": "dt-tree-" + line.id,
				"style": map[string]any{
					"foreground": fg,
					"height":     1,
				},
			},
		})
	}

	// Style inspector header
	children = append(children, &VNode{
		Type:    "text",
		Content: " ─── Computed Styles ───",
		Props: map[string]any{
			"style": map[string]any{
				"foreground": "#93C5FD",
				"height":     1,
			},
		},
	})

	// Style lines
	maxStyleLines := screenH/2 - 2
	if len(styleLines) > maxStyleLines {
		styleLines = styleLines[:maxStyleLines]
	}
	for _, line := range styleLines {
		children = append(children, &VNode{
			Type:    "text",
			Content: line,
			Props: map[string]any{
				"style": map[string]any{
					"foreground": "#D1D5DB",
					"height":     1,
				},
			},
		})
	}

	// Close hint
	children = append(children, &VNode{
		Type:    "text",
		Content: " [F12] Close",
		Props: map[string]any{
			"style": map[string]any{
				"foreground": "#6B7280",
				"height":     1,
			},
		},
	})

	return &VNode{
		Type: "vbox",
		Props: map[string]any{
			"style": map[string]any{
				"width":      panelW,
				"height":     screenH,
				"background": "#111827",
				"foreground": "#D1D5DB",
				"border":     "single",
			},
		},
		Children: children,
	}
}

// treeLine represents a line in the element tree display.
type treeLine struct {
	text string
	id   string
}

// buildElementTree recursively builds tree display lines.
func buildElementTree(vnode *VNode, depth int) []treeLine {
	if vnode == nil {
		return nil
	}

	var lines []treeLine
	indent := strings.Repeat("  ", depth)
	prefix := "├─"
	if depth == 0 {
		prefix = "▸"
	}

	// Get element ID and type
	id := ""
	if idVal, ok := vnode.Props["id"].(string); ok {
		id = idVal
	}

	typeName := vnode.Type
	if typeName == "" {
		typeName = "unknown"
	}

	// Format: ├─ vbox#my-id (30×10)
	label := fmt.Sprintf("%s%s %s", indent, prefix, typeName)
	if id != "" {
		label += "#" + id
	}
	if vnode.W > 0 && vnode.H > 0 {
		label += fmt.Sprintf(" (%d×%d)", vnode.W, vnode.H)
	}

	// Truncate to fit panel (guard against zero/negative panelWidth)
	maxLen := globalInspector.panelWidth - 4
	if maxLen > 0 && len(label) > maxLen {
		label = label[:maxLen-1] + "…"
	}

	nodeID := id
	if nodeID == "" {
		nodeID = fmt.Sprintf("%s-%d-%d", typeName, vnode.X, vnode.Y)
	}

	lines = append(lines, treeLine{text: " " + label, id: nodeID})

	for _, child := range vnode.Children {
		lines = append(lines, buildElementTree(child, depth+1)...)
	}

	return lines
}

// buildStyleInspector builds the style display for the selected element.
func buildStyleInspector(rootVNode *VNode) []string {
	selectedID := globalInspector.selectedID
	if selectedID == "" {
		selectedID = globalInspector.highlightID
	}
	if selectedID == "" {
		return []string{" (hover or click to inspect)"}
	}

	// Find the selected VNode
	node := findVNodeByID(rootVNode, selectedID)
	if node == nil {
		return []string{" Element: " + selectedID, " (not found in tree)"}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf(" Element: %s", node.Type))
	if id, ok := node.Props["id"].(string); ok && id != "" {
		lines = append(lines, fmt.Sprintf(" ID: %s", id))
	}
	lines = append(lines, fmt.Sprintf(" Position: (%d, %d)", node.X, node.Y))
	lines = append(lines, fmt.Sprintf(" Size: %d × %d", node.W, node.H))

	// Show content for text nodes
	if node.Content != "" {
		content := node.Content
		if len(content) > 30 {
			content = content[:27] + "..."
		}
		lines = append(lines, fmt.Sprintf(" Content: %q", content))
	}

	// Show computed style
	s := node.Style
	lines = append(lines, " ─── Style ───")
	if s.Position != "" {
		lines = append(lines, fmt.Sprintf("  position: %s", s.Position))
	}

	if s.Flex > 0 {
		lines = append(lines, fmt.Sprintf("  flex: %d", s.Flex))
	}
	if s.Width > 0 {
		lines = append(lines, fmt.Sprintf("  width: %d", s.Width))
	}
	if s.Height > 0 {
		lines = append(lines, fmt.Sprintf("  height: %d", s.Height))
	}
	if s.Padding > 0 {
		lines = append(lines, fmt.Sprintf("  padding: %d", s.Padding))
	}
	if s.PaddingTop > 0 || s.PaddingBottom > 0 || s.PaddingLeft > 0 || s.PaddingRight > 0 {
		lines = append(lines, fmt.Sprintf("  padding: %d %d %d %d",
			s.PaddingTop, s.PaddingRight, s.PaddingBottom, s.PaddingLeft))
	}
	if s.Margin > 0 {
		lines = append(lines, fmt.Sprintf("  margin: %d", s.Margin))
	}
	if s.Border != "" {
		lines = append(lines, fmt.Sprintf("  border: %s", s.Border))
	}
	if s.Foreground != "" {
		lines = append(lines, fmt.Sprintf("  fg: %s", s.Foreground))
	}
	if s.Background != "" {
		lines = append(lines, fmt.Sprintf("  bg: %s", s.Background))
	}
	if s.Bold {
		lines = append(lines, "  bold: true")
	}
	if s.Overflow != "" {
		lines = append(lines, fmt.Sprintf("  overflow: %s", s.Overflow))
	}
	if s.ZIndex != 0 {
		lines = append(lines, fmt.Sprintf("  z-index: %d", s.ZIndex))
	}
	if s.Justify != "" {
		lines = append(lines, fmt.Sprintf("  justify: %s", s.Justify))
	}
	if s.Align != "" {
		lines = append(lines, fmt.Sprintf("  align: %s", s.Align))
	}

	// Truncate lines to fit panel (guard against zero/negative panelWidth)
	maxLen := globalInspector.panelWidth - 3
	if maxLen > 0 {
		for i, line := range lines {
			if len(line) > maxLen {
				lines[i] = line[:maxLen-1] + "…"
			}
		}
	}

	return lines
}

// findVNodeByID recursively searches for a VNode with the given ID.
func findVNodeByID(vnode *VNode, id string) *VNode {
	if vnode == nil {
		return nil
	}
	if nodeID, ok := vnode.Props["id"].(string); ok && nodeID == id {
		return vnode
	}
	for _, child := range vnode.Children {
		if found := findVNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// RenderHighlight draws a highlight border around the hovered/selected element.
func RenderHighlight(frame *Frame, node *VNode) {
	if node == nil || frame == nil {
		return
	}
	// Draw top and bottom edges
	for x := node.X; x < node.X+node.W; x++ {
		setHighlightCell(frame, x, node.Y)
		setHighlightCell(frame, x, node.Y+node.H-1)
	}
	// Draw left and right edges
	for y := node.Y; y < node.Y+node.H; y++ {
		setHighlightCell(frame, node.X, y)
		setHighlightCell(frame, node.X+node.W-1, y)
	}
}

// setHighlightCell changes the background of a cell to indicate highlight.
func setHighlightCell(frame *Frame, x, y int) {
	if x >= 0 && x < frame.Width && y >= 0 && y < frame.Height {
		frame.Cells[y][x].Background = "#3B82F6" // blue highlight
	}
}


