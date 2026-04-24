// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// -----------------------------------------------------------------------
// Render Timeline
// -----------------------------------------------------------------------

// RenderEvent records a single component render for the timeline.
type RenderEvent struct {
	Timestamp  time.Time     `json:"timestamp"`
	Component  string        `json:"component"`
	Trigger    string        `json:"trigger"` // "state_change", "props_change", "effect", "event", "initial", "resize"
	Duration   time.Duration `json:"duration_ns"`
	DurationMs float64       `json:"duration_ms"`
	PatchCount int           `json:"patch_count"`
}

const maxTimelineSize = 100

var (
	renderTimeline   []RenderEvent
	renderTimelineMu sync.Mutex
)

// RecordRenderEvent appends a render event to the timeline (circular buffer).
func RecordRenderEvent(component, trigger string, duration time.Duration, patchCount int) {
	renderTimelineMu.Lock()
	defer renderTimelineMu.Unlock()

	event := RenderEvent{
		Timestamp:  time.Now(),
		Component:  component,
		Trigger:    trigger,
		Duration:   duration,
		DurationMs: float64(duration.Nanoseconds()) / 1e6,
		PatchCount: patchCount,
	}
	renderTimeline = append(renderTimeline, event)
	if len(renderTimeline) > maxTimelineSize {
		renderTimeline = renderTimeline[len(renderTimeline)-maxTimelineSize:]
	}
}

// GetRenderTimeline returns a copy of the render timeline.
func GetRenderTimeline() []RenderEvent {
	renderTimelineMu.Lock()
	defer renderTimelineMu.Unlock()
	cp := make([]RenderEvent, len(renderTimeline))
	copy(cp, renderTimeline)
	return cp
}

// ClearRenderTimeline clears the render timeline.
func ClearRenderTimeline() {
	renderTimelineMu.Lock()
	renderTimeline = renderTimeline[:0]
	renderTimelineMu.Unlock()
}

// -----------------------------------------------------------------------
// Debug Snapshots (time-travel debugging)
// -----------------------------------------------------------------------

// DebugSnapshot captures a component's state at a point in time for debugging.
type DebugSnapshot struct {
	ID        int            `json:"id"`
	Timestamp time.Time      `json:"timestamp"`
	Component string         `json:"component"`
	State     map[string]any `json:"state"`
}

const maxDebugSnapshots = 50

var (
	debugSnapshots   []DebugSnapshot
	snapshotCounter  int
	snapshotMu       sync.Mutex
)

// CaptureDebugSnapshot takes a snapshot of a component's current state.
func CaptureDebugSnapshot(compID string) *DebugSnapshot {
	comp, ok := GetComponentByID(compID)
	if !ok {
		return nil
	}

	comp.mu.RLock()
	stateCopy := make(map[string]any, len(comp.State))
	for k, v := range comp.State {
		stateCopy[k] = v
	}
	comp.mu.RUnlock()

	snapshotMu.Lock()
	defer snapshotMu.Unlock()

	snapshotCounter++
	snap := DebugSnapshot{
		ID:        snapshotCounter,
		Timestamp: time.Now(),
		Component: compID,
		State:     stateCopy,
	}
	debugSnapshots = append(debugSnapshots, snap)
	if len(debugSnapshots) > maxDebugSnapshots {
		debugSnapshots = debugSnapshots[len(debugSnapshots)-maxDebugSnapshots:]
	}
	return &snap
}

// RestoreDebugSnapshot restores a component to a previously captured state.
func RestoreDebugSnapshot(snapshotID int) bool {
	snapshotMu.Lock()
	var snap *DebugSnapshot
	for i := range debugSnapshots {
		if debugSnapshots[i].ID == snapshotID {
			snap = &debugSnapshots[i]
			break
		}
	}
	snapshotMu.Unlock()

	if snap == nil {
		return false
	}

	comp, ok := GetComponentByID(snap.Component)
	if !ok {
		return false
	}

	comp.mu.Lock()
	for k, v := range snap.State {
		comp.State[k] = v
	}
	comp.mu.Unlock()
	comp.Dirty.Store(true)

	// Notify render loop.
	if comp.RenderNotify != nil {
		select {
		case comp.RenderNotify <- struct{}{}:
		default:
		}
	}
	return true
}

// GetDebugSnapshots returns all captured debug snapshots.
func GetDebugSnapshots() []DebugSnapshot {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	cp := make([]DebugSnapshot, len(debugSnapshots))
	copy(cp, debugSnapshots)
	return cp
}

// ClearDebugSnapshots clears all debug snapshots.
func ClearDebugSnapshots() {
	snapshotMu.Lock()
	debugSnapshots = debugSnapshots[:0]
	snapshotMu.Unlock()
}

// -----------------------------------------------------------------------
// Performance Metrics
// -----------------------------------------------------------------------

// PerformanceMetrics aggregates render performance data.
type PerformanceMetrics struct {
	TotalRenders   int     `json:"total_renders"`
	AvgDurationMs  float64 `json:"avg_duration_ms"`
	MaxDurationMs  float64 `json:"max_duration_ms"`
	MinDurationMs  float64 `json:"min_duration_ms"`
	TotalPatches   int     `json:"total_patches"`
	SkippedRenders int     `json:"skipped_renders"` // renders skipped due to 0 patches
	SlowRenders    int     `json:"slow_renders_gt16ms"`
}

var (
	skippedRenders   int
	skippedRendersMu sync.Mutex
)

// RecordSkippedRender increments the skipped render counter.
func RecordSkippedRender() {
	skippedRendersMu.Lock()
	skippedRenders++
	skippedRendersMu.Unlock()
}

// GetPerformanceMetrics computes metrics from the render timeline.
func GetPerformanceMetrics() *PerformanceMetrics {
	timeline := GetRenderTimeline()
	m := &PerformanceMetrics{
		TotalRenders: len(timeline),
	}
	if len(timeline) == 0 {
		return m
	}

	var totalMs float64
	m.MinDurationMs = timeline[0].DurationMs
	for _, e := range timeline {
		totalMs += e.DurationMs
		m.TotalPatches += e.PatchCount
		if e.DurationMs > m.MaxDurationMs {
			m.MaxDurationMs = e.DurationMs
		}
		if e.DurationMs < m.MinDurationMs {
			m.MinDurationMs = e.DurationMs
		}
		if e.DurationMs > 16 {
			m.SlowRenders++
		}
	}
	m.AvgDurationMs = totalMs / float64(len(timeline))

	skippedRendersMu.Lock()
	m.SkippedRenders = skippedRenders
	skippedRendersMu.Unlock()

	return m
}

// ResetPerformanceMetrics clears all performance data.
func ResetPerformanceMetrics() {
	ClearRenderTimeline()
	skippedRendersMu.Lock()
	skippedRenders = 0
	skippedRendersMu.Unlock()
}

// -----------------------------------------------------------------------
// Debug Log
// -----------------------------------------------------------------------

// DebugEntry is a structured debug log entry.
type DebugEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // "debug", "info", "warn", "error"
	Message   string    `json:"message"`
	Component string    `json:"component,omitempty"`
}

const maxDebugEntries = 200

var (
	debugLog   []DebugEntry
	debugLogMu sync.Mutex
)

// DebugLog appends a structured debug entry.
func DebugLog(level, message, component string) {
	debugLogMu.Lock()
	defer debugLogMu.Unlock()
	debugLog = append(debugLog, DebugEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Component: component,
	})
	if len(debugLog) > maxDebugEntries {
		debugLog = debugLog[len(debugLog)-maxDebugEntries:]
	}
}

// GetDebugLog returns a copy of the debug log.
func GetDebugLog() []DebugEntry {
	debugLogMu.Lock()
	defer debugLogMu.Unlock()
	cp := make([]DebugEntry, len(debugLog))
	copy(cp, debugLog)
	return cp
}

// ClearDebugLog clears the debug log.
func ClearDebugLog() {
	debugLogMu.Lock()
	debugLog = debugLog[:0]
	debugLogMu.Unlock()
}

// -----------------------------------------------------------------------
// MCP Request Handling for Debug Actions
// -----------------------------------------------------------------------

// HandleDebugMCPRequest handles debug-specific MCP requests.
// Returns (result, handled). If handled is false, the caller should try other handlers.
func HandleDebugMCPRequest(req MCPRequest) (interface{}, bool) {
	switch req.Method {
	case "debug.timeline":
		return map[string]interface{}{
			"events": GetRenderTimeline(),
			"count":  len(GetRenderTimeline()),
		}, true

	case "debug.performance":
		return GetPerformanceMetrics(), true

	case "debug.snapshot":
		var params struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(req.Params, &params) != nil || params.ID == "" {
			return map[string]interface{}{"error": "missing component id"}, true
		}
		snap := CaptureDebugSnapshot(params.ID)
		if snap == nil {
			return map[string]interface{}{"error": "component not found"}, true
		}
		return snap, true

	case "debug.snapshots":
		snaps := GetDebugSnapshots()
		return map[string]interface{}{
			"snapshots": snaps,
			"count":     len(snaps),
		}, true

	case "debug.restore":
		var params struct {
			SnapshotID int `json:"snapshot_id"`
		}
		if json.Unmarshal(req.Params, &params) != nil {
			return map[string]interface{}{"error": "missing snapshot_id"}, true
		}
		ok := RestoreDebugSnapshot(params.SnapshotID)
		return map[string]interface{}{"restored": ok}, true

	case "debug.log":
		return map[string]interface{}{
			"entries": GetDebugLog(),
			"count":   len(GetDebugLog()),
		}, true

	case "debug.reset":
		ResetPerformanceMetrics()
		ClearDebugSnapshots()
		ClearDebugLog()
		return map[string]interface{}{"ok": true}, true
	}

	return nil, false
}

// -----------------------------------------------------------------------
// Lua API for Debug
// -----------------------------------------------------------------------

// debugLogLua — lumina.debug.log(message) or lumina.debug.log(level, message)
func debugLogLua(L *lua.State) int {
	var level, message string
	if L.GetTop() >= 2 {
		level = L.CheckString(1)
		message = L.CheckString(2)
	} else {
		level = "debug"
		message = L.CheckString(1)
	}

	compID := ""
	if comp := GetCurrentComponent(); comp != nil {
		compID = comp.ID
	}
	DebugLog(level, message, compID)
	return 0
}

// debugInspectLua — lumina.debug.inspect(componentId) → table
func debugInspectLua(L *lua.State) int {
	id := L.CheckString(1)
	comp, ok := GetComponentByID(id)
	if !ok {
		L.PushNil()
		return 1
	}

	comp.mu.RLock()
	state := make(map[string]any, len(comp.State))
	for k, v := range comp.State {
		state[k] = v
	}
	comp.mu.RUnlock()

	L.NewTableFrom(map[string]any{
		"id":    comp.ID,
		"type":  comp.Type,
		"name":  comp.Name,
		"dirty": comp.Dirty.Load(),
		"state": state,
	})
	return 1
}

// debugTimelineLua — lumina.debug.timeline() → JSON string
func debugTimelineLua(L *lua.State) int {
	timeline := GetRenderTimeline()
	data, err := json.Marshal(timeline)
	if err != nil {
		L.PushString("[]")
		return 1
	}
	L.PushString(string(data))
	return 1
}

// debugPerformanceLua — lumina.debug.performance() → JSON string
func debugPerformanceLua(L *lua.State) int {
	metrics := GetPerformanceMetrics()
	data, err := json.Marshal(metrics)
	if err != nil {
		L.PushString("{}")
		return 1
	}
	L.PushString(string(data))
	return 1
}

// debugSnapshotLua — lumina.debug.snapshot(componentId) → snapshot table
func debugSnapshotLua(L *lua.State) int {
	id := L.CheckString(1)
	snap := CaptureDebugSnapshot(id)
	if snap == nil {
		L.PushNil()
		return 1
	}
	L.NewTableFrom(map[string]any{
		"id":        int64(snap.ID),
		"component": snap.Component,
		"state":     snap.State,
	})
	return 1
}

// debugRestoreLua — lumina.debug.restore(snapshotId) → boolean
func debugRestoreLua(L *lua.State) int {
	id, ok := L.ToInteger(1)
	if !ok {
		L.PushBoolean(false)
		return 1
	}
	L.PushBoolean(RestoreDebugSnapshot(int(id)))
	return 1
}

// RegisterDebugAPI registers the lumina.debug sub-table.
func RegisterDebugAPI(L *lua.State) {
	L.SetFuncs(map[string]lua.Function{
		"log":         debugLogLua,
		"inspect":     debugInspectLua,
		"timeline":    debugTimelineLua,
		"performance": debugPerformanceLua,
		"snapshot":    debugSnapshotLua,
		"restore":     debugRestoreLua,
	}, 0)
}
