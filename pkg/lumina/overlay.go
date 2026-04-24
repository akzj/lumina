// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"sort"
	"sync"
)

// Overlay represents a popup/overlay layer rendered on top of the base UI.
type Overlay struct {
	ID      string // unique identifier
	VNode   *VNode // content to render
	X, Y    int    // screen position (top-left corner)
	W, H    int    // dimensions
	ZIndex  int    // render order (higher = on top)
	Visible bool   // whether the overlay is currently shown
	Modal   bool   // if true, captures all input and dims background
}

// OverlayManager manages popup/overlay layers.
type OverlayManager struct {
	overlays map[string]*Overlay
	mu       sync.RWMutex
}

// NewOverlayManager creates a new OverlayManager.
func NewOverlayManager() *OverlayManager {
	return &OverlayManager{
		overlays: make(map[string]*Overlay),
	}
}

// globalOverlayManager is the singleton overlay manager.
var globalOverlayManager = NewOverlayManager()

// Show adds or updates an overlay and makes it visible.
func (om *OverlayManager) Show(overlay *Overlay) {
	om.mu.Lock()
	defer om.mu.Unlock()
	overlay.Visible = true
	om.overlays[overlay.ID] = overlay
}

// Hide hides an overlay (does not remove it).
func (om *OverlayManager) Hide(id string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	if ov, ok := om.overlays[id]; ok {
		ov.Visible = false
	}
}

// Remove removes an overlay entirely.
func (om *OverlayManager) Remove(id string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	delete(om.overlays, id)
}

// Get returns an overlay by ID, or nil if not found.
func (om *OverlayManager) Get(id string) *Overlay {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.overlays[id]
}

// IsVisible returns whether an overlay is currently visible.
func (om *OverlayManager) IsVisible(id string) bool {
	om.mu.RLock()
	defer om.mu.RUnlock()
	if ov, ok := om.overlays[id]; ok {
		return ov.Visible
	}
	return false
}

// Toggle toggles an overlay's visibility.
func (om *OverlayManager) Toggle(id string) bool {
	om.mu.Lock()
	defer om.mu.Unlock()
	if ov, ok := om.overlays[id]; ok {
		ov.Visible = !ov.Visible
		return ov.Visible
	}
	return false
}

// GetVisible returns all visible overlays sorted by ZIndex (ascending).
func (om *OverlayManager) GetVisible() []*Overlay {
	om.mu.RLock()
	defer om.mu.RUnlock()
	var result []*Overlay
	for _, ov := range om.overlays {
		if ov.Visible {
			result = append(result, ov)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ZIndex < result[j].ZIndex
	})
	return result
}

// GetTopModal returns the highest-ZIndex visible modal overlay, or nil.
func (om *OverlayManager) GetTopModal() *Overlay {
	om.mu.RLock()
	defer om.mu.RUnlock()
	var top *Overlay
	for _, ov := range om.overlays {
		if ov.Visible && ov.Modal {
			if top == nil || ov.ZIndex > top.ZIndex {
				top = ov
			}
		}
	}
	return top
}

// Clear removes all overlays.
func (om *OverlayManager) Clear() {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.overlays = make(map[string]*Overlay)
}

// Count returns the total number of overlays (visible + hidden).
func (om *OverlayManager) Count() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return len(om.overlays)
}
