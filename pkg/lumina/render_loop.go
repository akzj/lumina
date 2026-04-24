// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"github.com/akzj/go-lua/pkg/lua"
)

// RenderLoop is retained for backward compatibility.
// In the new architecture, rendering is driven by the App's single-threaded
// event loop. This struct now just holds dimensions and delegates to App.
type RenderLoop struct {
	L           *lua.State
	frameWidth  int
	frameHeight int
}

// NewRenderLoop creates a new render loop (kept for API compatibility).
func NewRenderLoop(L *lua.State, width, height int) *RenderLoop {
	return &RenderLoop{
		L:           L,
		frameWidth:  width,
		frameHeight: height,
	}
}

// Start is a no-op in the new architecture.
// Rendering is now driven by the App's event loop.
func (rl *RenderLoop) Start() {}

// Stop is a no-op in the new architecture.
func (rl *RenderLoop) Stop() {}

// InitialRender renders all components once using the render loop's state.
// This is used for non-App usage (e.g., tests that create RenderLoop directly).
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
		frame.FocusedID = globalEventBus.GetFocused()
		adapter.Write(frame)
	}
	SetCurrentComponent(nil)
}

// IsRunning always returns false in the new architecture.
// The event loop is managed by App.
func (rl *RenderLoop) IsRunning() bool {
	return false
}
