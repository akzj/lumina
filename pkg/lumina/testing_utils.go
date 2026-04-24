package lumina

import (
	"strings"
	"sync"
)

// TestRenderer provides a headless renderer for testing Lumina components.
type TestRenderer struct {
	mu     sync.Mutex
	root   *TestVNode
	events []TestEvent
}

// TestVNode is a simplified VNode for testing.
type TestVNode struct {
	Type     string
	Content  string
	Props    map[string]any
	Aria     AriaAttributes
	Style    map[string]any
	Children []*TestVNode
}

// TestEvent records a fired event.
type TestEvent struct {
	Target    string
	EventType string
	Data      any
}

// NewTestRenderer creates a new TestRenderer.
func NewTestRenderer() *TestRenderer {
	return &TestRenderer{}
}

// Render sets the root VNode from a map (as would come from Lua).
func (tr *TestRenderer) Render(vnodeMap map[string]any) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.root = mapToTestVNode(vnodeMap)
}

// Root returns the root TestVNode.
func (tr *TestRenderer) Root() *TestVNode {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.root
}

// GetByText finds the first node containing the given text.
func (tr *TestRenderer) GetByText(text string) *TestVNode {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return findByText(tr.root, text)
}

// GetByRole finds the first node with the given ARIA role.
func (tr *TestRenderer) GetByRole(role string) *TestVNode {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return findByRole(tr.root, role)
}

// GetByType finds the first node with the given type.
func (tr *TestRenderer) GetByType(nodeType string) *TestVNode {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return findByType(tr.root, nodeType)
}

// GetAllByText finds all nodes containing the given text.
func (tr *TestRenderer) GetAllByText(text string) []*TestVNode {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	var results []*TestVNode
	collectByText(tr.root, text, &results)
	return results
}

// GetAllByRole finds all nodes with the given ARIA role.
func (tr *TestRenderer) GetAllByRole(role string) []*TestVNode {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	var results []*TestVNode
	collectByRole(tr.root, role, &results)
	return results
}

// FireEvent records an event on the given target.
func (tr *TestRenderer) FireEvent(target string, eventType string, data any) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.events = append(tr.events, TestEvent{
		Target:    target,
		EventType: eventType,
		Data:      data,
	})
}

// Events returns all recorded events.
func (tr *TestRenderer) Events() []TestEvent {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.events
}

// Reset clears the renderer state.
func (tr *TestRenderer) Reset() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.root = nil
	tr.events = nil
}

// RenderToString renders the VNode tree to a flat text string.
func RenderToString(node *TestVNode) string {
	if node == nil {
		return ""
	}
	var sb strings.Builder
	renderToStringHelper(node, &sb)
	return sb.String()
}

func renderToStringHelper(node *TestVNode, sb *strings.Builder) {
	if node.Content != "" {
		sb.WriteString(node.Content)
	}
	for _, child := range node.Children {
		renderToStringHelper(child, sb)
	}
}

// mapToTestVNode converts a map to a TestVNode.
func mapToTestVNode(m map[string]any) *TestVNode {
	if m == nil {
		return nil
	}
	node := &TestVNode{
		Props: m,
	}
	if t, ok := m["type"].(string); ok {
		node.Type = t
	}
	if c, ok := m["content"].(string); ok {
		node.Content = c
	}
	if s, ok := m["style"].(map[string]any); ok {
		node.Style = s
	}
	if a, ok := m["aria"].(map[string]any); ok {
		node.Aria = ParseAriaFromMap(a)
	}
	if children, ok := m["children"].([]any); ok {
		for _, child := range children {
			if cm, ok := child.(map[string]any); ok {
				node.Children = append(node.Children, mapToTestVNode(cm))
			}
		}
	}
	return node
}

func findByText(node *TestVNode, text string) *TestVNode {
	if node == nil {
		return nil
	}
	if strings.Contains(node.Content, text) {
		return node
	}
	for _, child := range node.Children {
		if found := findByText(child, text); found != nil {
			return found
		}
	}
	return nil
}

func findByRole(node *TestVNode, role string) *TestVNode {
	if node == nil {
		return nil
	}
	if node.Aria.Role == role {
		return node
	}
	for _, child := range node.Children {
		if found := findByRole(child, role); found != nil {
			return found
		}
	}
	return nil
}

func findByType(node *TestVNode, nodeType string) *TestVNode {
	if node == nil {
		return nil
	}
	if node.Type == nodeType {
		return node
	}
	for _, child := range node.Children {
		if found := findByType(child, nodeType); found != nil {
			return found
		}
	}
	return nil
}

func collectByText(node *TestVNode, text string, results *[]*TestVNode) {
	if node == nil {
		return
	}
	if strings.Contains(node.Content, text) {
		*results = append(*results, node)
	}
	for _, child := range node.Children {
		collectByText(child, text, results)
	}
}

func collectByRole(node *TestVNode, role string, results *[]*TestVNode) {
	if node == nil {
		return
	}
	if node.Aria.Role == role {
		*results = append(*results, node)
	}
	for _, child := range node.Children {
		collectByRole(child, role, results)
	}
}
