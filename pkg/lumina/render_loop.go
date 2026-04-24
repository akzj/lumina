// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"context"
	"sync"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// RenderLoop manages the continuous rendering for interactive applications.
type RenderLoop struct {
	L           *lua.State
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool
	mu          sync.Mutex
	frameWidth  int
	frameHeight int
}

// NewRenderLoop creates a new render loop.
func NewRenderLoop(L *lua.State, width, height int) *RenderLoop {
	ctx, cancel := context.WithCancel(context.Background())
	return &RenderLoop{
		L:           L,
		ctx:         ctx,
		cancel:      cancel,
		frameWidth:  width,
		frameHeight: height,
	}
}

// Start starts the render loop in a goroutine.
func (rl *RenderLoop) Start() {
	rl.mu.Lock()
	if rl.running {
		rl.mu.Unlock()
		return
	}
	rl.running = true
	rl.mu.Unlock()

	go rl.loop()
}

// Stop stops the render loop.
func (rl *RenderLoop) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if !rl.running {
		return
	}
	rl.running = false
	rl.cancel()
}

// loop is the main render loop that watches for dirty components.
func (rl *RenderLoop) loop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60fps
	defer ticker.Stop()

	for {
		select {
		case <-rl.ctx.Done():
			return
		case <-ticker.C:
			rl.checkAndRender()
		}
	}
}

// checkAndRender checks all components for dirty state and re-renders.
func (rl *RenderLoop) checkAndRender() {
	globalRegistry.mu.RLock()
	components := make([]*Component, 0, len(globalRegistry.components))
	for _, comp := range globalRegistry.components {
		components = append(components, comp)
	}
	globalRegistry.mu.RUnlock()

	for _, comp := range components {
		if comp.Dirty.Load() {
			rl.renderComponent(comp)
		}
	}
}

// renderComponent re-renders a single component.
func (rl *RenderLoop) renderComponent(comp *Component) {
	// Get the adapter - use ANSI adapter for terminal
	adapter := GetOutputAdapter()
	if adapter == nil {
		return
	}

	// Set current component for hooks
	SetCurrentComponent(comp)

	// Call render function
	if !comp.PushRenderFn(rl.L) {
		return
	}

	// Execute render and get frame
	status := rl.L.PCall(0, 1, 0)
	if status != lua.OK {
		rl.L.Pop(1)
		return
	}

	// Convert Lua VNode to frame
	frame := RenderLuaVNode(rl.L, -1, rl.frameWidth, rl.frameHeight)
	rl.L.Pop(1)

	// Set focused component ID from event bus
	frame.FocusedID = globalEventBus.GetFocused()

	// Write to terminal
	adapter.Write(frame)

	// Mark clean
	comp.MarkClean()
}

// InitialRender renders all components once.
func (rl *RenderLoop) InitialRender() {
	globalRegistry.mu.RLock()
	components := make([]*Component, 0, len(globalRegistry.components))
	for _, comp := range globalRegistry.components {
		components = append(components, comp)
	}
	globalRegistry.mu.RUnlock()

	adapter := GetOutputAdapter()
	if adapter == nil {
		return
	}

	for _, comp := range components {
		SetCurrentComponent(comp)

		if !comp.PushRenderFn(rl.L) {
			continue
		}

		status := rl.L.PCall(0, 1, 0)
		if status != lua.OK {
			rl.L.Pop(1)
			continue
		}

		frame := RenderLuaVNode(rl.L, -1, rl.frameWidth, rl.frameHeight)
		rl.L.Pop(1)

		// Set focused component ID from event bus
		frame.FocusedID = globalEventBus.GetFocused()

		adapter.Write(frame)
	}
}

// IsRunning returns true if the render loop is running.
func (rl *RenderLoop) IsRunning() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.running
}
