package lumina

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHotReloader(t *testing.T) {
	config := HotReloadConfig{
		Enabled:  true,
		Interval: 500 * time.Millisecond,
	}
	hr := NewHotReloader(config)
	assert.NotNil(t, hr)
	assert.True(t, hr.IsEnabled())
}

func TestHotReloaderEnable(t *testing.T) {
	config := HotReloadConfig{Enabled: false}
	hr := NewHotReloader(config)
	assert.False(t, hr.IsEnabled())

	hr.Enable(true)
	assert.True(t, hr.IsEnabled())

	hr.Enable(false)
	assert.False(t, hr.IsEnabled())
}

func TestSnapshotState(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	comp := &Component{
		ID:    "test-comp",
		State: map[string]any{"count": 42, "name": "test"},
	}

	hr.SnapshotState(comp)

	snap := hr.GetSnapshot("test-comp")
	assert.NotNil(t, snap)
	assert.Equal(t, "test-comp", snap.ComponentID)
	assert.Equal(t, 42, snap.State["count"])
}

func TestRestoreState(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	comp := &Component{
		ID:    "test-comp",
		State: map[string]any{"count": 0},
	}

	// Modify state
	comp.State["count"] = 42

	// Snapshot the state
	hr.SnapshotState(comp)

	// Reset state
	comp.State["count"] = 0

	// Restore should bring back 42
	restored := hr.RestoreState(comp)
	assert.True(t, restored)
	assert.Equal(t, 42, comp.State["count"])
}

func TestRestoreStateNotFound(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	comp := &Component{
		ID:    "nonexistent",
		State: map[string]any{"count": 0},
	}

	// Restore should fail for non-existent snapshot
	restored := hr.RestoreState(comp)
	assert.False(t, restored)
}

func TestClearSnapshots(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	comp := &Component{ID: "test", State: map[string]any{"n": 1}}
	hr.SnapshotState(comp)

	assert.NotNil(t, hr.GetSnapshot("test"))

	hr.ClearSnapshots()
	assert.Nil(t, hr.GetSnapshot("test"))
}

func TestWatch(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	called := false
	hr.Watch(func(path string) {
		called = true
	})

	hr.Notify("/path/to/file.lua")
	assert.True(t, called)
}

func TestWatchMultiple(t *testing.T) {
	hr := NewHotReloader(HotReloadConfig{})

	count := 0
	hr.Watch(func(path string) { count++ })
	hr.Watch(func(path string) { count++ })

	hr.Notify("test")
	assert.Equal(t, 2, count)
}

func TestCopyMap(t *testing.T) {
	original := map[string]any{
		"string": "value",
		"number": 42,
		"nested": map[string]any{"a": 1},
	}

	copied := copyMap(original)
	assert.Equal(t, original["string"], copied["string"])
	assert.Equal(t, original["number"], copied["number"])

	// Verify it's a copy, not a reference
	original["string"] = "changed"
	assert.Equal(t, "value", copied["string"])
}

func TestCopyMapNil(t *testing.T) {
	copied := copyMap(nil)
	assert.NotNil(t, copied)
	assert.Equal(t, 0, len(copied))
}
