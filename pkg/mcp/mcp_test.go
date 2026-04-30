package mcp

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// --- Mock AppInspector ---

type mockApp struct {
	components       []ComponentInfo
	detail           *ComponentDetail
	componentList    []ComponentSummary
	componentProps   map[string]map[string]any
	state            map[string]any
	focusedID        string
	focusableIDs     []string
	clickedID        string
	keyPressed       string
	evalResult       any
	evalErr          error
	devToolsVis      bool
	screenText       string
	version          string
	setStateCalls    []setStateCall
}

type setStateCall struct {
	compID string
	key    string
	value  any
}

func (m *mockApp) MCPInspectTree() []ComponentInfo { return m.components }
func (m *mockApp) MCPInspectComponent(id string) (*ComponentDetail, error) {
	if m.detail != nil && m.detail.ID == id {
		return m.detail, nil
	}
	return nil, errors.New("component not found: " + id)
}
func (m *mockApp) MCPInspectComponents(filter string) []ComponentSummary {
	if filter == "" {
		return m.componentList
	}
	var result []ComponentSummary
	for _, c := range m.componentList {
		if strings.Contains(strings.ToLower(c.Name), strings.ToLower(filter)) {
			result = append(result, c)
		}
	}
	return result
}
func (m *mockApp) MCPGetComponentProps(id string) (map[string]any, error) {
	if m.componentProps != nil {
		if props, ok := m.componentProps[id]; ok {
			return props, nil
		}
	}
	return nil, errors.New("component not found: " + id)
}
func (m *mockApp) MCPGetState(compID, key string) (any, error) {
	if m.state == nil {
		return nil, errors.New("component not found: " + compID)
	}
	if key != "" {
		val, ok := m.state[key]
		if !ok {
			return nil, errors.New("key not found: " + key)
		}
		return val, nil
	}
	return m.state, nil
}
func (m *mockApp) MCPSetState(compID, key string, value any) error {
	m.setStateCalls = append(m.setStateCalls, setStateCall{compID, key, value})
	return nil
}
func (m *mockApp) MCPSimulateClick(id string) error {
	m.clickedID = id
	return nil
}
func (m *mockApp) MCPSimulateKey(key string) error {
	m.keyPressed = key
	return nil
}
func (m *mockApp) MCPEval(code string) (any, error) {
	return m.evalResult, m.evalErr
}
func (m *mockApp) MCPFocusNext() string {
	return m.focusedID
}
func (m *mockApp) MCPFocusPrev() string {
	return m.focusedID
}
func (m *mockApp) MCPSetFocus(id string) {
	m.focusedID = id
}
func (m *mockApp) MCPGetFocusableIDs() []string {
	return m.focusableIDs
}
func (m *mockApp) MCPGetFocusedID() string {
	return m.focusedID
}
func (m *mockApp) MCPToggleDevTools() bool {
	m.devToolsVis = !m.devToolsVis
	return m.devToolsVis
}
func (m *mockApp) MCPGetScreenText() string {
	return m.screenText
}
func (m *mockApp) MCPGetVersion() string {
	return m.version
}

// --- Tests ---

func TestHandler_InspectTree(t *testing.T) {
	app := &mockApp{
		components: []ComponentInfo{
			{ID: "c1", Name: "Counter", Focused: true, Rect: [4]int{0, 0, 40, 10}},
			{ID: "c2", Name: "Timer", Focused: false, Rect: [4]int{0, 10, 40, 10}},
		},
		focusedID:    "c1",
		focusableIDs: []string{"c1", "c2"},
	}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 1, Method: "inspectTree"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]any)
	tree := result["tree"].([]ComponentInfo)
	if len(tree) != 2 {
		t.Fatalf("expected 2 components, got %d", len(tree))
	}
	if tree[0].ID != "c1" {
		t.Errorf("expected first component ID 'c1', got %q", tree[0].ID)
	}
	if result["focusedID"] != "c1" {
		t.Errorf("expected focusedID 'c1', got %v", result["focusedID"])
	}
	ids := result["focusableIDs"].([]string)
	if len(ids) != 2 {
		t.Errorf("expected 2 focusable IDs, got %d", len(ids))
	}
}

func TestHandler_InspectComponent(t *testing.T) {
	app := &mockApp{
		detail: &ComponentDetail{
			ID:      "c1",
			Name:    "Counter",
			State:   map[string]any{"count": 5},
			Focused: true,
			Dirty:   false,
			Rect:    [4]int{0, 0, 40, 10},
			ZIndex:  0,
		},
	}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 2, Method: "inspectComponent", Params: json.RawMessage(`{"id":"c1"}`)})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	detail := resp.Result.(*ComponentDetail)
	if detail.ID != "c1" {
		t.Errorf("expected ID 'c1', got %q", detail.ID)
	}
	if detail.Name != "Counter" {
		t.Errorf("expected Name 'Counter', got %q", detail.Name)
	}
	if !detail.Focused {
		t.Error("expected Focused=true")
	}

	// Not found
	resp2 := h.Handle(Request{ID: 3, Method: "inspectComponent", Params: json.RawMessage(`{"id":"nope"}`)})
	if resp2.Error == nil {
		t.Fatal("expected error for missing component")
	}
}

func TestHandler_GetSetState(t *testing.T) {
	app := &mockApp{
		state: map[string]any{"count": 42},
	}
	h := NewHandler(app)

	// Get specific key
	resp := h.Handle(Request{ID: 4, Method: "getState", Params: json.RawMessage(`{"id":"c1","key":"count"}`)})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]any)
	if result["value"] != 42 {
		t.Errorf("expected value 42, got %v", result["value"])
	}

	// Get all state (empty key)
	resp2 := h.Handle(Request{ID: 5, Method: "getState", Params: json.RawMessage(`{"id":"c1","key":""}`)})
	if resp2.Error != nil {
		t.Fatalf("unexpected error: %v", resp2.Error)
	}
	result2 := resp2.Result.(map[string]any)
	stateMap := result2["state"].(map[string]any)
	if stateMap["count"] != 42 {
		t.Errorf("expected count=42 in full state, got %v", stateMap["count"])
	}

	// Set state
	resp3 := h.Handle(Request{ID: 6, Method: "setState", Params: json.RawMessage(`{"id":"c1","key":"count","value":99}`)})
	if resp3.Error != nil {
		t.Fatalf("unexpected error: %v", resp3.Error)
	}
	if len(app.setStateCalls) != 1 {
		t.Fatalf("expected 1 setState call, got %d", len(app.setStateCalls))
	}
	call := app.setStateCalls[0]
	if call.compID != "c1" || call.key != "count" {
		t.Errorf("unexpected setState call: %+v", call)
	}
}

func TestHandler_SimulateClick(t *testing.T) {
	app := &mockApp{}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 7, Method: "simulateClick", Params: json.RawMessage(`{"id":"btn1"}`)})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if app.clickedID != "btn1" {
		t.Errorf("expected clickedID 'btn1', got %q", app.clickedID)
	}
}

func TestHandler_SimulateKey(t *testing.T) {
	app := &mockApp{}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 8, Method: "simulateKey", Params: json.RawMessage(`{"key":"Enter"}`)})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if app.keyPressed != "Enter" {
		t.Errorf("expected keyPressed 'Enter', got %q", app.keyPressed)
	}
}

func TestHandler_Eval(t *testing.T) {
	app := &mockApp{
		evalResult: map[string]any{"ok": true},
	}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 9, Method: "eval", Params: json.RawMessage(`{"code":"return 1+1"}`)})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}

	// Eval with error
	app.evalErr = errors.New("syntax error")
	app.evalResult = nil
	resp2 := h.Handle(Request{ID: 10, Method: "eval", Params: json.RawMessage(`{"code":"bad code"}`)})
	if resp2.Error == nil {
		t.Fatal("expected error for bad eval")
	}
	if resp2.Error.Message != "syntax error" {
		t.Errorf("expected 'syntax error', got %q", resp2.Error.Message)
	}
}

func TestHandler_Focus(t *testing.T) {
	app := &mockApp{
		focusedID:    "c1",
		focusableIDs: []string{"c1", "c2", "c3"},
	}
	h := NewHandler(app)

	// focusNext
	resp := h.Handle(Request{ID: 11, Method: "focusNext"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]string)
	if result["focused"] != "c1" {
		t.Errorf("expected focused 'c1', got %q", result["focused"])
	}

	// focusPrev
	resp2 := h.Handle(Request{ID: 12, Method: "focusPrev"})
	if resp2.Error != nil {
		t.Fatalf("unexpected error: %v", resp2.Error)
	}

	// setFocus
	resp3 := h.Handle(Request{ID: 13, Method: "setFocus", Params: json.RawMessage(`{"id":"c3"}`)})
	if resp3.Error != nil {
		t.Fatalf("unexpected error: %v", resp3.Error)
	}
	if app.focusedID != "c3" {
		t.Errorf("expected focusedID 'c3', got %q", app.focusedID)
	}

	// getFocusableIDs
	resp4 := h.Handle(Request{ID: 14, Method: "getFocusableIDs"})
	if resp4.Error != nil {
		t.Fatalf("unexpected error: %v", resp4.Error)
	}
	result4 := resp4.Result.(map[string]any)
	ids := result4["ids"].([]string)
	if len(ids) != 3 {
		t.Errorf("expected 3 focusable IDs, got %d", len(ids))
	}
}

func TestHandler_UnknownMethod(t *testing.T) {
	h := NewHandler(&mockApp{})

	resp := h.Handle(Request{ID: 99, Method: "nonexistent"})
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected error code -32601, got %d", resp.Error.Code)
	}
}

func TestHandler_GetFrame(t *testing.T) {
	app := &mockApp{
		focusedID:  "c1",
		screenText: "Hello World\n",
	}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 15, Method: "getFrame"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]any)
	if result["focusedID"] != "c1" {
		t.Errorf("expected focusedID 'c1', got %v", result["focusedID"])
	}
	if result["screen"] != "Hello World\n" {
		t.Errorf("unexpected screen text: %q", result["screen"])
	}
}

func TestHandler_ToggleDevTools(t *testing.T) {
	app := &mockApp{}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 16, Method: "debug.toggleDevTools"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]any)
	if result["visible"] != true {
		t.Error("expected visible=true after toggle")
	}
}

func TestHandler_GetVersion(t *testing.T) {
	app := &mockApp{version: "lumina-v2-test"}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 17, Method: "getVersion"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]string)
	if result["version"] != "lumina-v2-test" {
		t.Errorf("expected version 'lumina-v2-test', got %q", result["version"])
	}
}

func TestHandler_Tools(t *testing.T) {
	h := NewHandler(&mockApp{})
	tools := h.Tools()
	if len(tools) == 0 {
		t.Fatal("expected non-empty tool list")
	}
	// Verify all tools have required fields.
	for _, tool := range tools {
		if tool.Name == "" {
			t.Error("tool missing Name")
		}
		if tool.Description == "" {
			t.Errorf("tool %q missing Description", tool.Name)
		}
		if tool.InputSchema == nil {
			t.Errorf("tool %q missing InputSchema", tool.Name)
		}
	}
	// Verify specific tools exist.
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	required := []string{
		"lumina.inspectTree", "lumina.inspectComponent",
		"lumina.inspectComponents", "lumina.getComponentProps",
		"lumina.getState", "lumina.setState",
		"lumina.simulateClick", "lumina.simulateKey",
		"lumina.eval", "lumina.focusNext", "lumina.focusPrev",
		"lumina.setFocus", "lumina.getFocusableIDs",
		"lumina.getFrame", "lumina.getVersion",
	}
	for _, name := range required {
		if !names[name] {
			t.Errorf("missing required tool: %s", name)
		}
	}
}

func TestHandler_InspectComponents(t *testing.T) {
	app := &mockApp{
		componentList: []ComponentSummary{
			{
				ID:          "c1",
				Name:        "Counter",
				Props:       map[string]any{"label": "clicks"},
				State:       map[string]any{"count": 5},
				HookCount:   2,
				RenderCount: 3,
				Children:    []string{"c1-child1"},
			},
			{
				ID:          "c2",
				Name:        "LuxToast",
				Props:       map[string]any{"items": []any{}},
				State:       map[string]any{},
				HookCount:   1,
				RenderCount: 1,
				Children:    []string{},
			},
		},
	}
	h := NewHandler(app)

	// No filter — returns all
	resp := h.Handle(Request{ID: 20, Method: "inspectComponents"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]any)
	comps := result["components"].([]ComponentSummary)
	if len(comps) != 2 {
		t.Fatalf("expected 2 components, got %d", len(comps))
	}
	total := result["total"].(int)
	if total != 2 {
		t.Errorf("expected total=2, got %d", total)
	}

	// With filter — only matching
	resp2 := h.Handle(Request{ID: 21, Method: "inspectComponents", Params: json.RawMessage(`{"filter":"toast"}`)})
	if resp2.Error != nil {
		t.Fatalf("unexpected error: %v", resp2.Error)
	}
	result2 := resp2.Result.(map[string]any)
	comps2 := result2["components"].([]ComponentSummary)
	if len(comps2) != 1 {
		t.Fatalf("expected 1 filtered component, got %d", len(comps2))
	}
	if comps2[0].Name != "LuxToast" {
		t.Errorf("expected LuxToast, got %q", comps2[0].Name)
	}

	// With filter — no match
	resp3 := h.Handle(Request{ID: 22, Method: "inspectComponents", Params: json.RawMessage(`{"filter":"nonexistent"}`)})
	if resp3.Error != nil {
		t.Fatalf("unexpected error: %v", resp3.Error)
	}
	result3 := resp3.Result.(map[string]any)
	comps3 := result3["components"].([]ComponentSummary)
	if len(comps3) != 0 {
		t.Fatalf("expected 0 components, got %d", len(comps3))
	}
}

func TestHandler_GetComponentProps(t *testing.T) {
	app := &mockApp{
		componentProps: map[string]map[string]any{
			"c1": {"label": "clicks", "max": 100},
		},
	}
	h := NewHandler(app)

	// Found
	resp := h.Handle(Request{ID: 23, Method: "getComponentProps", Params: json.RawMessage(`{"id":"c1"}`)})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]any)
	props := result["props"].(map[string]any)
	if props["label"] != "clicks" {
		t.Errorf("expected label='clicks', got %v", props["label"])
	}
	if props["max"] != 100 {
		t.Errorf("expected max=100, got %v", props["max"])
	}

	// Not found
	resp2 := h.Handle(Request{ID: 24, Method: "getComponentProps", Params: json.RawMessage(`{"id":"nope"}`)})
	if resp2.Error == nil {
		t.Fatal("expected error for missing component")
	}
}

func TestHandler_InspectComponent_Enhanced(t *testing.T) {
	app := &mockApp{
		detail: &ComponentDetail{
			ID:          "c1",
			Name:        "Counter",
			Props:       map[string]any{"label": "clicks"},
			State:       map[string]any{"count": 5},
			Focused:     true,
			Dirty:       false,
			Rect:        [4]int{0, 0, 40, 10},
			ZIndex:      0,
			RenderCount: 7,
			Hooks:       HookSummary{Effects: 2, Memos: 1, Refs: 0},
			Children:    []string{"c1-child1", "c1-child2"},
		},
	}
	h := NewHandler(app)

	resp := h.Handle(Request{ID: 25, Method: "inspectComponent", Params: json.RawMessage(`{"id":"c1"}`)})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	detail := resp.Result.(*ComponentDetail)
	if detail.RenderCount != 7 {
		t.Errorf("expected RenderCount=7, got %d", detail.RenderCount)
	}
	if detail.Hooks.Effects != 2 {
		t.Errorf("expected 2 effects, got %d", detail.Hooks.Effects)
	}
	if detail.Hooks.Memos != 1 {
		t.Errorf("expected 1 memo, got %d", detail.Hooks.Memos)
	}
	if len(detail.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(detail.Children))
	}
	if detail.Props["label"] != "clicks" {
		t.Errorf("expected props.label='clicks', got %v", detail.Props["label"])
	}
}
