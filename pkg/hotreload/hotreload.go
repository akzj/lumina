// Package hotreload provides poll-based file watching and state-preserving
// hot reload for Lumina v2. When a watched .lua file changes on disk, the
// watcher triggers a callback so the app can re-execute the script and
// restore component state.
package hotreload

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Watcher polls files for modification time changes and triggers a callback.
type Watcher struct {
	paths    []string
	dirs     []string // directories to periodically rescan for new .lua files
	modTimes map[string]time.Time
	pathSet  map[string]bool // fast dedup lookup for paths
	interval time.Duration
	onChange func(path string)
	stopCh   chan struct{}
	mu       sync.Mutex
	running  bool

	// RescanInterval controls how often directories are rescanned for new files.
	// Expressed as number of poll ticks between rescans. Default: 10 (= 5s at 500ms poll).
	RescanInterval int
}

// NewWatcher creates a new file watcher that polls at the given interval.
// If interval <= 0, defaults to 500ms.
func NewWatcher(paths []string, interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	pathSet := make(map[string]bool, len(paths))
	for _, p := range paths {
		pathSet[p] = true
	}
	return &Watcher{
		paths:          append([]string(nil), paths...), // defensive copy
		modTimes:       make(map[string]time.Time),
		pathSet:        pathSet,
		interval:       interval,
		stopCh:         make(chan struct{}),
		RescanInterval: 10, // every 10 polls = 5s at 500ms default
	}
}

// SetOnChange sets the callback invoked when a watched file changes.
func (w *Watcher) SetOnChange(fn func(path string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onChange = fn
}

// Start begins polling in a background goroutine. Safe to call multiple
// times — subsequent calls are no-ops while running.
func (w *Watcher) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true

	// Initialize mod times for all known paths.
	for _, path := range w.paths {
		if info, err := os.Stat(path); err == nil {
			w.modTimes[path] = info.ModTime()
		}
	}
	w.mu.Unlock()

	go w.pollLoop()
}

// pollLoop runs the ticker-driven poll cycle until Stop is called.
func (w *Watcher) pollLoop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	pollCount := 0
	for {
		select {
		case <-ticker.C:
			w.poll()
			pollCount++
			w.mu.Lock()
			rescanInterval := w.RescanInterval
			w.mu.Unlock()
			if rescanInterval > 0 && pollCount%rescanInterval == 0 {
				w.rescanDirs()
			}
		case <-w.stopCh:
			return
		}
	}
}

// poll checks each watched path for modification time changes.
func (w *Watcher) poll() {
	w.mu.Lock()
	paths := append([]string(nil), w.paths...)
	w.mu.Unlock()

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		w.mu.Lock()
		prev, ok := w.modTimes[path]
		onChange := w.onChange
		if ok && info.ModTime().After(prev) {
			w.modTimes[path] = info.ModTime()
			w.mu.Unlock()
			if onChange != nil {
				onChange(path)
			}
		} else {
			if !ok {
				w.modTimes[path] = info.ModTime()
			}
			w.mu.Unlock()
		}
	}
}

// Stop stops the polling goroutine. Safe to call multiple times.
func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		close(w.stopCh)
		w.running = false
	}
}

// IsRunning returns whether the watcher is currently polling.
func (w *Watcher) IsRunning() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.running
}

// AddPath adds a path to the watch list. Can be called before or after Start.
// Duplicate paths are ignored (idempotent).
func (w *Watcher) AddPath(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.pathSet[path] {
		return // already watching
	}
	w.pathSet[path] = true
	w.paths = append(w.paths, path)
	if info, err := os.Stat(path); err == nil {
		w.modTimes[path] = info.ModTime()
	}
}

// AddDir adds a directory to the periodic rescan list.
// New .lua files appearing in this directory (or subdirectories) will be
// automatically picked up on the next rescan cycle.
func (w *Watcher) AddDir(dir string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.dirs = append(w.dirs, dir)
}

// rescanDirs walks all registered directories and adds any new .lua files.
func (w *Watcher) rescanDirs() {
	w.mu.Lock()
	dirs := append([]string(nil), w.dirs...)
	w.mu.Unlock()

	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && filepath.Ext(path) == ".lua" {
				w.mu.Lock()
				if !w.pathSet[path] {
					w.pathSet[path] = true
					w.paths = append(w.paths, path)
					if info, statErr := os.Stat(path); statErr == nil {
						w.modTimes[path] = info.ModTime()
					}
				}
				w.mu.Unlock()
			}
			return nil
		})
	}
}
