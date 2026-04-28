// Package hotreload provides poll-based file watching and state-preserving
// hot reload for Lumina v2. When a watched .lua file changes on disk, the
// watcher triggers a callback so the app can re-execute the script and
// restore component state.
package hotreload

import (
	"os"
	"sync"
	"time"
)

// Watcher polls files for modification time changes and triggers a callback.
type Watcher struct {
	paths    []string
	modTimes map[string]time.Time
	interval time.Duration
	onChange  func(path string)
	stopCh   chan struct{}
	mu       sync.Mutex
	running  bool
}

// NewWatcher creates a new file watcher that polls at the given interval.
// If interval <= 0, defaults to 500ms.
func NewWatcher(paths []string, interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &Watcher{
		paths:    append([]string(nil), paths...), // defensive copy
		modTimes: make(map[string]time.Time),
		interval: interval,
		stopCh:   make(chan struct{}),
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
	for {
		select {
		case <-ticker.C:
			w.poll()
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
func (w *Watcher) AddPath(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.paths = append(w.paths, path)
	if info, err := os.Stat(path); err == nil {
		w.modTimes[path] = info.ModTime()
	}
}
