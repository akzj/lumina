// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"testing"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

func TestSetStateMarksDirty(t *testing.T) {
	// Create component directly
	comp := &Component{
		ID:           "test_comp",
		Type:         "Test",
		Name:         "Test",
		Props:        make(map[string]any),
		State:        make(map[string]any),
		RenderNotify: make(chan struct{}, 1),
	}

	// Verify initial state
	if comp.Dirty.Load() {
		t.Error("Component should not be dirty initially")
	}

	// Set state
	comp.SetState("count", 42)

	// Verify dirty
	if !comp.Dirty.Load() {
		t.Error("Component should be dirty after SetState")
	}

	// Verify state updated
	count, ok := comp.GetState("count")
	if !ok || count != 42 {
		t.Errorf("Expected count=42, got %v", count)
	}

	t.Log("SetState correctly marks dirty and updates state")
}

func TestRenderLoopStartStop(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	Open(L)

	rl := NewRenderLoop(L, 80, 24)

	// Test start
	rl.Start()
	if !rl.IsRunning() {
		t.Error("RenderLoop should be running after Start()")
	}

	// Test stop
	rl.Stop()
	time.Sleep(50 * time.Millisecond) // Give time for goroutine to stop
	if rl.IsRunning() {
		t.Error("RenderLoop should not be running after Stop()")
	}
}

func TestAppCreation(t *testing.T) {
	app := NewApp()
	defer app.Close()

	if app.L == nil {
		t.Error("App should have Lua state")
	}

	if app.RenderLoop == nil {
		t.Error("App should have RenderLoop")
	}

	if !app.RenderLoop.IsRunning() {
		// Start and verify
		app.RenderLoop.Start()
		if !app.RenderLoop.IsRunning() {
			t.Error("RenderLoop should be running after Start()")
		}
		app.RenderLoop.Stop()
	}
}

func TestReactiveIntegration(t *testing.T) {
	// This test verifies the full reactive chain works
	L := lua.NewState()
	defer L.Close()
	Open(L)

	// Create a component
	comp := &Component{
		ID:           "test_comp",
		Type:         "Test",
		Name:         "Test",
		Props:        make(map[string]any),
		State:        map[string]any{"count": 0},
		RenderNotify: make(chan struct{}, 1),
	}

	// Register it
	globalRegistry.mu.Lock()
	globalRegistry.components[comp.ID] = comp
	globalRegistry.mu.Unlock()

	// Create render loop
	_ = NewRenderLoop(L, 80, 24)

	// Simulate state change
	comp.SetState("count", 1)

	// Verify dirty is set
	if !comp.Dirty.Load() {
		t.Error("Component should be dirty after SetState")
	}

	// Clean up
	globalRegistry.mu.Lock()
	delete(globalRegistry.components, comp.ID)
	globalRegistry.mu.Unlock()

	t.Log("Reactive chain verified: SetState -> Dirty flag set")
}
