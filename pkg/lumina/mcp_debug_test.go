package lumina

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRenderTimeline_RecordAndGet(t *testing.T) {
	ClearRenderTimeline()

	RecordRenderEvent("comp1", "state_change", 5*time.Millisecond, 3)
	RecordRenderEvent("comp2", "initial", 10*time.Millisecond, 0)

	timeline := GetRenderTimeline()
	if len(timeline) != 2 {
		t.Fatalf("got %d events, want 2", len(timeline))
	}
	if timeline[0].Component != "comp1" {
		t.Errorf("event[0].Component = %q, want comp1", timeline[0].Component)
	}
	if timeline[0].Trigger != "state_change" {
		t.Errorf("event[0].Trigger = %q, want state_change", timeline[0].Trigger)
	}
	if timeline[0].PatchCount != 3 {
		t.Errorf("event[0].PatchCount = %d, want 3", timeline[0].PatchCount)
	}
	if timeline[1].Component != "comp2" {
		t.Errorf("event[1].Component = %q, want comp2", timeline[1].Component)
	}
}

func TestRenderTimeline_CircularBuffer(t *testing.T) {
	ClearRenderTimeline()

	// Record more than maxTimelineSize events.
	for i := 0; i < maxTimelineSize+20; i++ {
		RecordRenderEvent("comp", "state_change", time.Millisecond, 1)
	}

	timeline := GetRenderTimeline()
	if len(timeline) != maxTimelineSize {
		t.Errorf("got %d events, want %d (circular buffer cap)", len(timeline), maxTimelineSize)
	}
}

func TestRenderTimeline_Clear(t *testing.T) {
	RecordRenderEvent("comp", "initial", time.Millisecond, 0)
	ClearRenderTimeline()
	if len(GetRenderTimeline()) != 0 {
		t.Errorf("timeline not empty after clear")
	}
}

func TestPerformanceMetrics_Empty(t *testing.T) {
	ClearRenderTimeline()
	ResetPerformanceMetrics()

	m := GetPerformanceMetrics()
	if m.TotalRenders != 0 {
		t.Errorf("TotalRenders = %d, want 0", m.TotalRenders)
	}
}

func TestPerformanceMetrics_Aggregation(t *testing.T) {
	ResetPerformanceMetrics()

	RecordRenderEvent("a", "state_change", 5*time.Millisecond, 2)
	RecordRenderEvent("b", "state_change", 20*time.Millisecond, 5) // slow (>16ms)
	RecordRenderEvent("c", "initial", 1*time.Millisecond, 0)

	m := GetPerformanceMetrics()
	if m.TotalRenders != 3 {
		t.Errorf("TotalRenders = %d, want 3", m.TotalRenders)
	}
	if m.TotalPatches != 7 {
		t.Errorf("TotalPatches = %d, want 7", m.TotalPatches)
	}
	if m.SlowRenders != 1 {
		t.Errorf("SlowRenders = %d, want 1", m.SlowRenders)
	}
	if m.MinDurationMs > 2.0 {
		t.Errorf("MinDurationMs = %f, want <= 2.0", m.MinDurationMs)
	}
	if m.MaxDurationMs < 19.0 {
		t.Errorf("MaxDurationMs = %f, want >= 19.0", m.MaxDurationMs)
	}
}

func TestSkippedRenders(t *testing.T) {
	ResetPerformanceMetrics()

	RecordSkippedRender()
	RecordSkippedRender()
	RecordRenderEvent("a", "state_change", time.Millisecond, 1)

	m := GetPerformanceMetrics()
	if m.SkippedRenders != 2 {
		t.Errorf("SkippedRenders = %d, want 2", m.SkippedRenders)
	}
}

func TestDebugLog_RecordAndGet(t *testing.T) {
	ClearDebugLog()

	DebugLog("info", "hello world", "comp1")
	DebugLog("error", "something broke", "")

	entries := GetDebugLog()
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].Level != "info" || entries[0].Message != "hello world" {
		t.Errorf("entry[0] = %+v, want info/hello world", entries[0])
	}
	if entries[1].Level != "error" || entries[1].Component != "" {
		t.Errorf("entry[1] = %+v, want error with empty component", entries[1])
	}
}

func TestDebugLog_CircularBuffer(t *testing.T) {
	ClearDebugLog()

	for i := 0; i < maxDebugEntries+50; i++ {
		DebugLog("debug", "msg", "")
	}

	entries := GetDebugLog()
	if len(entries) != maxDebugEntries {
		t.Errorf("got %d entries, want %d", len(entries), maxDebugEntries)
	}
}

func TestDebugSnapshot_CaptureNonExistent(t *testing.T) {
	snap := CaptureDebugSnapshot("nonexistent-component")
	if snap != nil {
		t.Errorf("expected nil for nonexistent component, got %+v", snap)
	}
}

func TestDebugSnapshot_RestoreNonExistent(t *testing.T) {
	ok := RestoreDebugSnapshot(99999)
	if ok {
		t.Errorf("expected false for nonexistent snapshot")
	}
}

func TestDebugSnapshots_ClearAndGet(t *testing.T) {
	ClearDebugSnapshots()
	snaps := GetDebugSnapshots()
	if len(snaps) != 0 {
		t.Errorf("got %d snapshots after clear, want 0", len(snaps))
	}
}

func TestHandleDebugMCPRequest_Timeline(t *testing.T) {
	ClearRenderTimeline()
	RecordRenderEvent("comp", "initial", time.Millisecond, 0)

	req := MCPRequest{Method: "debug.timeline"}
	result, handled := HandleDebugMCPRequest(req)
	if !handled {
		t.Fatal("debug.timeline not handled")
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("result type = %T, want map", result)
	}
	if m["count"].(int) != 1 {
		t.Errorf("count = %v, want 1", m["count"])
	}
}

func TestHandleDebugMCPRequest_Performance(t *testing.T) {
	ResetPerformanceMetrics()
	RecordRenderEvent("comp", "state_change", 5*time.Millisecond, 2)

	req := MCPRequest{Method: "debug.performance"}
	result, handled := HandleDebugMCPRequest(req)
	if !handled {
		t.Fatal("debug.performance not handled")
	}
	m, ok := result.(*PerformanceMetrics)
	if !ok {
		t.Fatalf("result type = %T, want *PerformanceMetrics", result)
	}
	if m.TotalRenders != 1 {
		t.Errorf("TotalRenders = %d, want 1", m.TotalRenders)
	}
}

func TestHandleDebugMCPRequest_Log(t *testing.T) {
	ClearDebugLog()
	DebugLog("warn", "test warning", "comp1")

	req := MCPRequest{Method: "debug.log"}
	result, handled := HandleDebugMCPRequest(req)
	if !handled {
		t.Fatal("debug.log not handled")
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("result type = %T, want map", result)
	}
	if m["count"].(int) != 1 {
		t.Errorf("count = %v, want 1", m["count"])
	}
}

func TestHandleDebugMCPRequest_Reset(t *testing.T) {
	RecordRenderEvent("comp", "initial", time.Millisecond, 0)
	DebugLog("info", "msg", "")

	req := MCPRequest{Method: "debug.reset"}
	result, handled := HandleDebugMCPRequest(req)
	if !handled {
		t.Fatal("debug.reset not handled")
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("result type = %T, want map", result)
	}
	if m["ok"] != true {
		t.Errorf("ok = %v, want true", m["ok"])
	}

	// Verify everything was cleared.
	if len(GetRenderTimeline()) != 0 {
		t.Errorf("timeline not empty after reset")
	}
	if len(GetDebugLog()) != 0 {
		t.Errorf("debug log not empty after reset")
	}
}

func TestHandleDebugMCPRequest_SnapshotMissingID(t *testing.T) {
	params, _ := json.Marshal(map[string]string{})
	req := MCPRequest{Method: "debug.snapshot", Params: params}
	result, handled := HandleDebugMCPRequest(req)
	if !handled {
		t.Fatal("debug.snapshot not handled")
	}
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("result type = %T, want map", result)
	}
	if _, hasErr := m["error"]; !hasErr {
		t.Errorf("expected error for missing id, got %v", m)
	}
}

func TestHandleDebugMCPRequest_UnknownMethod(t *testing.T) {
	req := MCPRequest{Method: "unknown.method"}
	_, handled := HandleDebugMCPRequest(req)
	if handled {
		t.Errorf("unknown method should not be handled")
	}
}

func TestRenderEventJSON(t *testing.T) {
	event := RenderEvent{
		Timestamp:  time.Now(),
		Component:  "test",
		Trigger:    "state_change",
		Duration:   5 * time.Millisecond,
		DurationMs: 5.0,
		PatchCount: 3,
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if decoded["component"] != "test" {
		t.Errorf("component = %v, want test", decoded["component"])
	}
	if decoded["trigger"] != "state_change" {
		t.Errorf("trigger = %v, want state_change", decoded["trigger"])
	}
}
