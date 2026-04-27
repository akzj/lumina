package hotreload

// StateSnapshot captures a component's state for restore after hot reload.
type StateSnapshot struct {
	ID        string
	Name      string
	State     map[string]any
	HookStore map[string]any
}

// Snapshottable is the interface a component must satisfy for snapshotting.
// Matches v2/component.Component accessors.
type Snapshottable interface {
	ID() string
	Name() string
	State() map[string]any
	HookStore() map[string]any
}

// Snapshot captures state from a component. The returned snapshot contains
// shallow copies of State and HookStore so mutations to the original
// component do not affect the snapshot.
func Snapshot(comp Snapshottable) StateSnapshot {
	return StateSnapshot{
		ID:        comp.ID(),
		Name:      comp.Name(),
		State:     copyMap(comp.State()),
		HookStore: copyMap(comp.HookStore()),
	}
}

// copyMap creates a shallow copy of a map.
func copyMap(m map[string]any) map[string]any {
	if m == nil {
		return make(map[string]any)
	}
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
