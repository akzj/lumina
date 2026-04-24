package lumina

import (
	"fmt"
	"sync"
	"time"
)

// DevTools provides a built-in developer tools panel for inspecting
// component trees, props, state, and performance.
type DevTools struct {
	mu            sync.RWMutex
	enabled       bool
	showPanel     bool
	selectedID    string
	componentTree []*DevToolsNode
	renderCounts  map[string]int
	renderTimes   map[string]time.Duration
	lastUpdate    time.Time
}

// DevToolsNode represents a component in the devtools tree.
type DevToolsNode struct {
	ID       string
	Name     string
	Props    map[string]any
	State    map[string]any
	Hooks    []DevToolsHook
	Children []*DevToolsNode
	Depth    int
}

// DevToolsHook represents a hook used by a component.
type DevToolsHook struct {
	Type  string // "useState", "useEffect", "useMemo", etc.
	Value any    // current value (for useState)
	Deps  []any  // dependency array (for useEffect/useMemo)
}

var globalDevTools = &DevTools{
	renderCounts: make(map[string]int),
	renderTimes:  make(map[string]time.Duration),
}

// GetDevTools returns the global DevTools instance.
func GetDevTools() *DevTools {
	return globalDevTools
}

// Enable enables the devtools.
func (dt *DevTools) Enable() {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.enabled = true
}

// Disable disables the devtools.
func (dt *DevTools) Disable() {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.enabled = false
	dt.showPanel = false
}

// IsEnabled returns whether devtools are enabled.
func (dt *DevTools) IsEnabled() bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.enabled
}

// Toggle toggles the devtools panel visibility.
func (dt *DevTools) Toggle() {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	if !dt.enabled {
		dt.enabled = true
	}
	dt.showPanel = !dt.showPanel
}

// IsVisible returns whether the panel is visible.
func (dt *DevTools) IsVisible() bool {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.enabled && dt.showPanel
}

// SetSelected sets the selected component ID.
func (dt *DevTools) SetSelected(id string) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.selectedID = id
}

// GetSelected returns the selected component ID.
func (dt *DevTools) GetSelected() string {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.selectedID
}

// RecordRender records a component render event.
func (dt *DevTools) RecordRender(componentName string, duration time.Duration) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	if !dt.enabled {
		return
	}
	dt.renderCounts[componentName]++
	dt.renderTimes[componentName] = duration
	dt.lastUpdate = time.Now()
}

// GetRenderCount returns the render count for a component.
func (dt *DevTools) GetRenderCount(componentName string) int {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.renderCounts[componentName]
}

// GetRenderTime returns the last render time for a component.
func (dt *DevTools) GetRenderTime(componentName string) time.Duration {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.renderTimes[componentName]
}

// UpdateTree updates the component tree snapshot.
func (dt *DevTools) UpdateTree(tree []*DevToolsNode) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.componentTree = tree
}

// GetTree returns the current component tree.
func (dt *DevTools) GetTree() []*DevToolsNode {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return dt.componentTree
}

// GetNodeByID finds a node in the tree by ID.
func (dt *DevTools) GetNodeByID(id string) *DevToolsNode {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	return findNode(dt.componentTree, id)
}

func findNode(nodes []*DevToolsNode, id string) *DevToolsNode {
	for _, n := range nodes {
		if n.ID == id {
			return n
		}
		if found := findNode(n.Children, id); found != nil {
			return found
		}
	}
	return nil
}

// RenderTree returns a string representation of the component tree.
func (dt *DevTools) RenderTree() string {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	if len(dt.componentTree) == 0 {
		return "  (no components mounted)"
	}
	result := ""
	for _, node := range dt.componentTree {
		result += dt.renderNode(node, 0)
	}
	return result
}

func (dt *DevTools) renderNode(node *DevToolsNode, depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	prefix := "├─"
	if depth == 0 {
		prefix = "▸"
	}

	selected := ""
	if node.ID == dt.selectedID {
		selected = " ◄"
	}

	count := dt.renderCounts[node.Name]
	renderTime := dt.renderTimes[node.Name]

	line := fmt.Sprintf("%s%s %s (renders: %d, %v)%s\n",
		indent, prefix, node.Name, count, renderTime, selected)

	for _, child := range node.Children {
		line += dt.renderNode(child, depth+1)
	}
	return line
}

// RenderInspector returns a string representation of the selected component.
func (dt *DevTools) RenderInspector() string {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	if dt.selectedID == "" {
		return "  Select a component to inspect"
	}
	node := findNode(dt.componentTree, dt.selectedID)
	if node == nil {
		return "  Component not found: " + dt.selectedID
	}

	result := fmt.Sprintf("  Component: %s\n", node.Name)
	result += fmt.Sprintf("  ID: %s\n", node.ID)
	result += fmt.Sprintf("  Renders: %d\n", dt.renderCounts[node.Name])
	result += fmt.Sprintf("  Last render: %v\n", dt.renderTimes[node.Name])
	result += "\n  Props:\n"
	if len(node.Props) == 0 {
		result += "    (none)\n"
	}
	for k, v := range node.Props {
		result += fmt.Sprintf("    %s: %v\n", k, v)
	}
	result += "\n  State:\n"
	if len(node.State) == 0 {
		result += "    (none)\n"
	}
	for k, v := range node.State {
		result += fmt.Sprintf("    %s: %v\n", k, v)
	}
	result += "\n  Hooks:\n"
	if len(node.Hooks) == 0 {
		result += "    (none)\n"
	}
	for _, h := range node.Hooks {
		result += fmt.Sprintf("    %s: %v\n", h.Type, h.Value)
	}
	return result
}

// Reset resets devtools state (for testing).
func (dt *DevTools) Reset() {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.enabled = false
	dt.showPanel = false
	dt.selectedID = ""
	dt.componentTree = nil
	dt.renderCounts = make(map[string]int)
	dt.renderTimes = make(map[string]time.Duration)
}

// Summary returns a brief summary of devtools state.
func (dt *DevTools) Summary() map[string]any {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	totalRenders := 0
	for _, c := range dt.renderCounts {
		totalRenders += c
	}
	return map[string]any{
		"enabled":        dt.enabled,
		"visible":        dt.showPanel,
		"components":     len(dt.componentTree),
		"total_renders":  totalRenders,
		"selected":       dt.selectedID,
	}
}
