package lumina

// inspectSystem provides debugging capabilities for AI agents.
type inspectSystem struct {
	frameHistory  []*Frame
	maxHistory    int
	componentTree []*ComponentSnapshot
}

// ComponentSnapshot captures component state for inspection.
type ComponentSnapshot struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Name     string         `json:"name"`
	X        int            `json:"x"`
	Y        int            `json:"y"`
	W        int            `json:"w"`
	H        int            `json:"h"`
	Props    map[string]any `json:"props"`
	State    map[string]any `json:"state"`
	Children []string       `json:"children,omitempty"`
}

// ComputedStyles holds the final calculated styles.
type ComputedStyles struct {
	ComponentID string         `json:"component_id"`
	Layout      map[string]any `json:"layout"` // x, y, w, h, flex
	Visual      map[string]any `json:"visual"` // color, bg, border, padding
	Raw         map[string]any `json:"raw"`    // original props
}

// Global inspector
var inspector = &inspectSystem{
	frameHistory:  make([]*Frame, 0, 100),
	maxHistory:    100,
	componentTree: make([]*ComponentSnapshot, 0),
}

// RecordFrame adds a frame to history (for diff).
func RecordFrame(frame *Frame) {
	if frame.Timestamp == 0 {
		frame.Timestamp = CurrentTimestamp()
	}
	inspector.frameHistory = append(inspector.frameHistory, frame)
	if len(inspector.frameHistory) > inspector.maxHistory {
		inspector.frameHistory = inspector.frameHistory[1:]
	}
}

// GetFrameHistory returns recent frames.
func GetFrameHistory() []*Frame {
	return inspector.frameHistory
}

// InspectTree returns the component tree structure.
func InspectTree() []*ComponentSnapshot {

	snapshots := make([]*ComponentSnapshot, 0, len(globalRegistry.components))
	for _, comp := range globalRegistry.components {
		snapshot := &ComponentSnapshot{
			ID:    comp.ID,
			Type:  comp.Type,
			Name:  comp.Name,
			Props: copyMap(comp.Props),
			State: copyMap(comp.State),
		}

		// Get layout from props
		if x, ok := comp.Props["x"].(int); ok {
			snapshot.X = x
		}
		if y, ok := comp.Props["y"].(int); ok {
			snapshot.Y = y
		}
		if w, ok := comp.Props["width"].(int); ok {
			snapshot.W = w
		}
		if h, ok := comp.Props["height"].(int); ok {
			snapshot.H = h
		}

		snapshots = append(snapshots, snapshot)
	}
	return snapshots
}

// InspectComponent returns detailed info for a specific component.
func InspectComponent(id string) *ComponentSnapshot {

	if comp, ok := globalRegistry.components[id]; ok {
		return &ComponentSnapshot{
			ID:       comp.ID,
			Type:     comp.Type,
			Name:     comp.Name,
			X:        getIntProp(comp.Props, "x"),
			Y:        getIntProp(comp.Props, "y"),
			W:        getIntProp(comp.Props, "width"),
			H:        getIntProp(comp.Props, "height"),
			Props:    copyMap(comp.Props),
			State:    copyMap(comp.State),
			Children: getChildren(comp),
		}
	}
	return nil
}

// getChildren returns child component IDs.
func getChildren(comp *Component) []string {
	children := make([]string, 0)
	if childIDs, ok := comp.Props["children"].([]string); ok {
		return childIDs
	}
	return children
}

// InspectStyles returns computed styles for a component.
func InspectStyles(id string) *ComputedStyles {

	if comp, ok := globalRegistry.components[id]; ok {
		return computeStyles(comp)
	}
	return nil
}

// computeStyles calculates final styles from component props.
func computeStyles(comp *Component) *ComputedStyles {
	styles := &ComputedStyles{
		ComponentID: comp.ID,
		Layout:      make(map[string]any),
		Visual:      make(map[string]any),
		Raw:         make(map[string]any),
	}

	// Layout from props
	if v, ok := comp.Props["x"].(int); ok {
		styles.Layout["x"] = v
	}
	if v, ok := comp.Props["y"].(int); ok {
		styles.Layout["y"] = v
	}
	if v, ok := comp.Props["width"].(int); ok {
		styles.Layout["w"] = v
	}
	if v, ok := comp.Props["height"].(int); ok {
		styles.Layout["h"] = v
	}
	if v, ok := comp.Props["flex"].(int); ok {
		styles.Layout["flex"] = v
	}
	if v, ok := comp.Props["minWidth"].(int); ok {
		styles.Layout["minWidth"] = v
	}
	if v, ok := comp.Props["minHeight"].(int); ok {
		styles.Layout["minHeight"] = v
	}

	// Visual from props
	if v, ok := comp.Props["foreground"].(string); ok {
		styles.Visual["foreground"] = v
	}
	if v, ok := comp.Props["color"].(string); ok {
		styles.Visual["color"] = v
	}
	if v, ok := comp.Props["background"].(string); ok {
		styles.Visual["background"] = v
	}
	if v, ok := comp.Props["border"].(string); ok {
		styles.Visual["border"] = v
	}
	if v, ok := comp.Props["padding"].(int); ok {
		styles.Visual["padding"] = v
	}
	if v, ok := comp.Props["bold"].(bool); ok {
		styles.Visual["bold"] = v
	}
	if v, ok := comp.Props["dim"].(bool); ok {
		styles.Visual["dim"] = v
	}
	if v, ok := comp.Props["underline"].(bool); ok {
		styles.Visual["underline"] = v
	}

	// Copy all props as raw
	for k, v := range comp.Props {
		styles.Raw[k] = v
	}

	return styles
}

// GetState returns component state.
func GetState(id string) (map[string]any, bool) {

	if comp, ok := globalRegistry.components[id]; ok {
		return copyStringAnyMap(comp.State), true
	}
	return nil, false
}

// Helper functions
func getIntProp(props map[string]any, key string) int {
	if v, ok := props[key].(int); ok {
		return v
	}
	return 0
}

func copyStringAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// GetAllComponentIDs returns all registered component IDs.
func GetAllComponentIDs() []string {

	ids := make([]string, 0, len(globalRegistry.components))
	for id := range globalRegistry.components {
		ids = append(ids, id)
	}
	return ids
}
