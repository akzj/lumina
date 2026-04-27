package hotreload

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcher_DetectsChange(t *testing.T) {
	// Create temp file.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lua")
	if err := os.WriteFile(path, []byte("-- v1"), 0644); err != nil {
		t.Fatal(err)
	}

	var called atomic.Int32
	var changedPath atomic.Value

	w := NewWatcher([]string{path}, 50*time.Millisecond)
	w.SetOnChange(func(p string) {
		called.Add(1)
		changedPath.Store(p)
	})
	w.Start()
	defer w.Stop()

	// Wait for initial poll to register the modtime.
	time.Sleep(100 * time.Millisecond)

	// Modify the file.
	time.Sleep(10 * time.Millisecond) // ensure modtime differs
	if err := os.WriteFile(path, []byte("-- v2"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for the watcher to detect the change.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if called.Load() > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if called.Load() == 0 {
		t.Fatal("expected onChange callback to be called")
	}
	if got := changedPath.Load().(string); got != path {
		t.Fatalf("expected path %q, got %q", path, got)
	}
}

func TestWatcher_NoChangeNoCallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lua")
	if err := os.WriteFile(path, []byte("-- v1"), 0644); err != nil {
		t.Fatal(err)
	}

	var called atomic.Int32

	w := NewWatcher([]string{path}, 50*time.Millisecond)
	w.SetOnChange(func(p string) {
		called.Add(1)
	})
	w.Start()
	defer w.Stop()

	// Wait several poll cycles without modifying the file.
	time.Sleep(300 * time.Millisecond)

	if called.Load() != 0 {
		t.Fatalf("expected no callback, got %d calls", called.Load())
	}
}

func TestWatcher_StartStop(t *testing.T) {
	w := NewWatcher(nil, 50*time.Millisecond)

	if w.IsRunning() {
		t.Fatal("should not be running before Start")
	}

	w.Start()
	if !w.IsRunning() {
		t.Fatal("should be running after Start")
	}

	// Start again should be a no-op.
	w.Start()
	if !w.IsRunning() {
		t.Fatal("should still be running after double Start")
	}

	w.Stop()
	if w.IsRunning() {
		t.Fatal("should not be running after Stop")
	}
}

func TestWatcher_AddPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "added.lua")
	if err := os.WriteFile(path, []byte("-- v1"), 0644); err != nil {
		t.Fatal(err)
	}

	var called atomic.Int32

	w := NewWatcher(nil, 50*time.Millisecond)
	w.SetOnChange(func(p string) {
		called.Add(1)
	})
	w.Start()
	defer w.Stop()

	// Add path after start.
	w.AddPath(path)

	time.Sleep(100 * time.Millisecond)

	// Modify the file.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(path, []byte("-- v2"), 0644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if called.Load() > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if called.Load() == 0 {
		t.Fatal("expected callback after AddPath + modify")
	}
}

func TestSnapshot_CaptureRestore(t *testing.T) {
	// Mock component.
	comp := &mockComponent{
		id:   "test-comp",
		name: "TestComp",
		state: map[string]any{
			"count": 5,
			"text":  "hello",
		},
		hookStore: map[string]any{
			"hook_0": "value0",
		},
	}

	snap := Snapshot(comp)

	// Verify snapshot captured values.
	if snap.ID != "test-comp" {
		t.Fatalf("expected ID 'test-comp', got %q", snap.ID)
	}
	if snap.Name != "TestComp" {
		t.Fatalf("expected Name 'TestComp', got %q", snap.Name)
	}
	if snap.State["count"] != 5 {
		t.Fatalf("expected count=5, got %v", snap.State["count"])
	}
	if snap.HookStore["hook_0"] != "value0" {
		t.Fatalf("expected hook_0='value0', got %v", snap.HookStore["hook_0"])
	}

	// Mutate original — snapshot should be independent.
	comp.state["count"] = 99
	comp.hookStore["hook_0"] = "mutated"

	if snap.State["count"] != 5 {
		t.Fatal("snapshot was affected by mutation of original state")
	}
	if snap.HookStore["hook_0"] != "value0" {
		t.Fatal("snapshot was affected by mutation of original hookStore")
	}
}

// mockComponent implements Snapshottable for testing.
type mockComponent struct {
	id        string
	name      string
	state     map[string]any
	hookStore map[string]any
}

func (m *mockComponent) ID() string              { return m.id }
func (m *mockComponent) Name() string            { return m.name }
func (m *mockComponent) State() map[string]any   { return m.state }

func TestWatcher_MissingFile(t *testing.T) {
	// Watcher should not panic or call onChange for a file that doesn't exist.
	var called atomic.Int32

	w := NewWatcher([]string{"/nonexistent/path/foo.lua"}, 50*time.Millisecond)
	w.SetOnChange(func(p string) {
		called.Add(1)
	})
	w.Start()
	defer w.Stop()

	time.Sleep(200 * time.Millisecond)

	if called.Load() != 0 {
		t.Fatal("should not call onChange for missing file")
	}
}

func TestWatcher_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.lua")
	path2 := filepath.Join(dir, "b.lua")
	if err := os.WriteFile(path1, []byte("-- a v1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path2, []byte("-- b v1"), 0644); err != nil {
		t.Fatal(err)
	}

	var paths []string
	var mu sync.Mutex

	w := NewWatcher([]string{path1, path2}, 50*time.Millisecond)
	w.SetOnChange(func(p string) {
		mu.Lock()
		paths = append(paths, p)
		mu.Unlock()
	})
	w.Start()
	defer w.Stop()

	time.Sleep(100 * time.Millisecond)

	// Modify both files.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(path1, []byte("-- a v2"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path2, []byte("-- b v2"), 0644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(paths)
		mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(paths) < 2 {
		t.Fatalf("expected at least 2 onChange calls, got %d", len(paths))
	}

	// Both paths should be present.
	got := map[string]bool{}
	for _, p := range paths {
		got[p] = true
	}
	if !got[path1] {
		t.Fatalf("expected callback for %s", path1)
	}
	if !got[path2] {
		t.Fatalf("expected callback for %s", path2)
	}
}

func TestSnapshot_NilMaps(t *testing.T) {
	comp := &mockComponent{
		id:        "nil-maps",
		name:      "NilMaps",
		state:     nil,
		hookStore: nil,
	}

	snap := Snapshot(comp)

	if snap.State == nil {
		t.Fatal("State should be non-nil even when source is nil")
	}
	if snap.HookStore == nil {
		t.Fatal("HookStore should be non-nil even when source is nil")
	}
	if len(snap.State) != 0 {
		t.Fatalf("expected empty State, got %d entries", len(snap.State))
	}
	if len(snap.HookStore) != 0 {
		t.Fatalf("expected empty HookStore, got %d entries", len(snap.HookStore))
	}
}

func TestSnapshot_EmptyMaps(t *testing.T) {
	comp := &mockComponent{
		id:        "empty",
		name:      "Empty",
		state:     map[string]any{},
		hookStore: map[string]any{},
	}

	snap := Snapshot(comp)

	if snap.ID != "empty" {
		t.Fatalf("expected ID 'empty', got %q", snap.ID)
	}
	if len(snap.State) != 0 {
		t.Fatalf("expected empty State, got %d entries", len(snap.State))
	}
}
func (m *mockComponent) HookStore() map[string]any { return m.hookStore }
