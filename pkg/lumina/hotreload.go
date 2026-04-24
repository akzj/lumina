package lumina

import (
	"sync"
	"time"
)

// HotReloadConfig configures the hot reload system.
type HotReloadConfig struct {
	Enabled    bool
	Interval   time.Duration
	WatchPaths []string
}

// StateSnapshot captures component state for reload.
type StateSnapshot struct {
	ComponentID string
	State       map[string]any
	Timestamp   int64
}

// HotReloader manages component hot reloading.
type HotReloader struct {
	config    HotReloadConfig
	snapshots map[string]*StateSnapshot
	mu        sync.RWMutex
	watchers  []func(string)
}

// Global hot reloader instance
var globalHotReloader *HotReloader

func init() {
	globalHotReloader = NewHotReloader(HotReloadConfig{
		Enabled:  false,
		Interval: 500 * time.Millisecond,
	})
}

// NewHotReloader creates a new hot reloader.
func NewHotReloader(config HotReloadConfig) *HotReloader {
	if config.Interval == 0 {
		config.Interval = 500 * time.Millisecond
	}
	return &HotReloader{
		config:    config,
		snapshots: make(map[string]*StateSnapshot),
	}
}

// SnapshotState saves the current state of a component.
func (hr *HotReloader) SnapshotState(comp *Component) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.snapshots[comp.ID] = &StateSnapshot{
		ComponentID: comp.ID,
		State:       copyMap(comp.State),
		Timestamp:   time.Now().UnixMilli(),
	}
}

// RestoreState restores a snapshot to a component.
func (hr *HotReloader) RestoreState(comp *Component) bool {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	if snap, ok := hr.snapshots[comp.ID]; ok {
		comp.mu.Lock()
		comp.State = copyMap(snap.State)
		comp.mu.Unlock()
		return true
	}
	return false
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

// Watch registers a callback for file changes.
func (hr *HotReloader) Watch(callback func(path string)) {
	hr.watchers = append(hr.watchers, callback)
}

// Notify notifies all watchers of a change.
func (hr *HotReloader) Notify(path string) {
	for _, w := range hr.watchers {
		w(path)
	}
}

// EnableHotReload enables or disables hot reload.
func (hr *HotReloader) Enable(enabled bool) {
	hr.config.Enabled = enabled
}

// IsEnabled returns whether hot reload is enabled.
func (hr *HotReloader) IsEnabled() bool {
	return hr.config.Enabled
}

// GetSnapshot returns a snapshot for a component.
func (hr *HotReloader) GetSnapshot(compID string) *StateSnapshot {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.snapshots[compID]
}

// ClearSnapshots removes all snapshots.
func (hr *HotReloader) ClearSnapshots() {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.snapshots = make(map[string]*StateSnapshot)
}

// SnapshotAllComponents snapshots all registered components.
// Note: This requires a component registry with instance tracking.
func (hr *HotReloader) SnapshotAllComponents() {
	// Component instance tracking would be added here
	// For now, this is a placeholder for future implementation
}
