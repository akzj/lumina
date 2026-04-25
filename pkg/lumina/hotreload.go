package lumina

import (
	"os"
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
	ComponentID   string
	ComponentType string // component Type/Name for restore-by-name matching
	State         map[string]any
	Timestamp     int64
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
	// Read component state under its lock first
	comp.mu.RLock()
	stateCopy := copyMap(comp.State)
	compID := comp.ID
	compType := comp.Type
	comp.mu.RUnlock()

	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.snapshots[compID] = &StateSnapshot{
		ComponentID:   compID,
		ComponentType: compType,
		State:         stateCopy,
		Timestamp:     time.Now().UnixMilli(),
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
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.watchers = append(hr.watchers, callback)
}

// Notify notifies all watchers of a change.
func (hr *HotReloader) Notify(path string) {
	hr.mu.RLock()
	watchers := make([]func(string), len(hr.watchers))
	copy(watchers, hr.watchers)
	hr.mu.RUnlock()
	for _, w := range watchers {
		w(path)
	}
}

// EnableHotReload enables or disables hot reload.
func (hr *HotReloader) Enable(enabled bool) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.config.Enabled = enabled
}

// IsEnabled returns whether hot reload is enabled.
func (hr *HotReloader) IsEnabled() bool {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return hr.config.Enabled
}

// SetConfig updates the hot reloader configuration under lock.
func (hr *HotReloader) SetConfig(interval time.Duration, paths []string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	if interval > 0 {
		hr.config.Interval = interval
	}
	if paths != nil {
		hr.config.WatchPaths = paths
	}
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
func (hr *HotReloader) SnapshotAllComponents() {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	for _, comp := range globalRegistry.components {
		hr.SnapshotState(comp)
	}
}

// RestoreByType restores state to a component by matching ComponentType.
// Returns true if a matching snapshot was found and restored.
func (hr *HotReloader) RestoreByType(comp *Component) bool {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	for _, snap := range hr.snapshots {
		if snap.ComponentType == comp.Type {
			comp.mu.Lock()
			comp.State = copyMap(snap.State)
			comp.Dirty.Store(true)
			comp.mu.Unlock()
			return true
		}
	}
	return false
}

// RestoreAllByType restores state to all registered components by type matching.
func (hr *HotReloader) RestoreAllByType() {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	for _, comp := range globalRegistry.components {
		hr.RestoreByType(comp)
	}
}

// GetSnapshotByType returns a snapshot matching the given component type.
func (hr *HotReloader) GetSnapshotByType(compType string) *StateSnapshot {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	for _, snap := range hr.snapshots {
		if snap.ComponentType == compType {
			return snap
		}
	}
	return nil
}

// -----------------------------------------------------------------------
// FileWatcher — poll-based file change detection
// -----------------------------------------------------------------------

// FileWatcher polls files for modification time changes.
type FileWatcher struct {
	paths    []string
	modTimes map[string]time.Time
	interval time.Duration
	onChange func(path string)
	stopCh   chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewFileWatcher creates a new poll-based file watcher.
func NewFileWatcher(paths []string, interval time.Duration) *FileWatcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &FileWatcher{
		paths:    paths,
		modTimes: make(map[string]time.Time),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start begins polling for file changes in a background goroutine.
func (fw *FileWatcher) Start() {
	fw.mu.Lock()
	if fw.running {
		fw.mu.Unlock()
		return
	}
	fw.running = true

	// Initialize mod times under lock
	for _, path := range fw.paths {
		if info, err := os.Stat(path); err == nil {
			fw.modTimes[path] = info.ModTime()
		}
	}
	fw.mu.Unlock()

	go func() {
		ticker := time.NewTicker(fw.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fw.poll()
			case <-fw.stopCh:
				return
			}
		}
	}()
}

// poll checks all watched paths for changes.
func (fw *FileWatcher) poll() {
	fw.mu.Lock()
	onChange := fw.onChange
	fw.mu.Unlock()

	for _, path := range fw.paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		fw.mu.Lock()
		prev, ok := fw.modTimes[path]
		if ok && info.ModTime().After(prev) {
			fw.modTimes[path] = info.ModTime()
			fw.mu.Unlock()
			if onChange != nil {
				onChange(path)
			}
		} else {
			if !ok {
				fw.modTimes[path] = info.ModTime()
			}
			fw.mu.Unlock()
		}
	}
}

// Stop stops the file watcher.
func (fw *FileWatcher) Stop() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.running {
		close(fw.stopCh)
		fw.running = false
	}
}

// IsRunning returns whether the watcher is active.
func (fw *FileWatcher) IsRunning() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.running
}

// SetOnChange sets the callback for file changes.
func (fw *FileWatcher) SetOnChange(fn func(path string)) {
	fw.onChange = fn
}
