// Package mcp implements the Model Context Protocol (MCP) handler for Lumina v2.
// It provides JSON-RPC 2.0 request dispatch, tool definitions, and an
// AppInspector interface that decouples the handler from the v2 App.
package mcp

import "encoding/json"

// --- JSON-RPC types ---

// Request is a JSON-RPC 2.0 request.
type Request struct {
	ID     any             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	ID     any    `json:"id"`
	Result any    `json:"result,omitempty"`
	Error  *Error `json:"error,omitempty"`
}

// Error is a JSON-RPC error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool describes an MCP tool for the tools/list response.
type Tool struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// --- Domain types ---

// ComponentInfo is a summary of a component for tree listing.
type ComponentInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Focused bool   `json:"focused"`
	Rect    [4]int `json:"rect"` // x, y, w, h
}

// ComponentDetail is a detailed view of a single component.
type ComponentDetail struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Props       map[string]any `json:"props"`
	State       map[string]any `json:"state"`
	Focused     bool           `json:"focused"`
	Dirty       bool           `json:"dirty"`
	Rect        [4]int         `json:"rect"` // x, y, w, h
	ZIndex      int            `json:"zIndex"`
	RenderCount int            `json:"renderCount"`
	Hooks       HookSummary    `json:"hooks"`
	Children    []string       `json:"children"` // child component IDs
}

// HookSummary summarizes the hooks used by a component.
type HookSummary struct {
	Effects int `json:"effects"`
	Memos   int `json:"memos"`
	Refs    int `json:"refs"`
}

// ComponentSummary is a lightweight summary for listing all components.
type ComponentSummary struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Props       map[string]any `json:"props,omitempty"`
	State       map[string]any `json:"state,omitempty"`
	HookCount   int            `json:"hookCount"`
	RenderCount int            `json:"renderCount"`
	Children    []string       `json:"children"`
}

// --- AppInspector interface ---

// AppInspector is the interface that the v2 App must implement for MCP.
// This decouples the mcp package from the v2 App type.
type AppInspector interface {
	MCPInspectTree() []ComponentInfo
	MCPInspectComponent(id string) (*ComponentDetail, error)
	MCPInspectComponents(filter string) []ComponentSummary
	MCPGetComponentProps(id string) (map[string]any, error)
	MCPGetState(compID, key string) (any, error)
	MCPSetState(compID, key string, value any) error
	MCPSimulateClick(id string) error
	MCPSimulateKey(key string) error
	MCPEval(code string) (any, error)
	MCPFocusNext() string
	MCPFocusPrev() string
	MCPSetFocus(id string)
	MCPGetFocusableIDs() []string
	MCPGetFocusedID() string
	MCPToggleDevTools() bool
	MCPGetScreenText() string
	MCPGetVersion() string
	MCPHotReload() error
}

// --- Handler ---

// Handler dispatches MCP requests to an AppInspector.
type Handler struct {
	app AppInspector
}

// NewHandler creates a new MCP handler backed by the given AppInspector.
func NewHandler(app AppInspector) *Handler {
	return &Handler{app: app}
}

// Handle dispatches a single MCP request and returns a response.
func (h *Handler) Handle(req Request) Response {
	var result any
	var err error

	switch req.Method {
	case "inspectTree":
		result = h.handleInspectTree()
	case "inspectComponent":
		result, err = h.handleInspectComponent(req.Params)
	case "inspectComponents":
		result = h.handleInspectComponents(req.Params)
	case "getComponentProps":
		result, err = h.handleGetComponentProps(req.Params)
	case "getState":
		result, err = h.handleGetState(req.Params)
	case "setState":
		result, err = h.handleSetState(req.Params)
	case "simulateClick":
		result, err = h.handleSimulateClick(req.Params)
	case "simulateKey":
		result, err = h.handleSimulateKey(req.Params)
	case "eval":
		result, err = h.handleEval(req.Params)
	case "focusNext":
		focused := h.app.MCPFocusNext()
		result = map[string]string{"focused": focused}
	case "focusPrev":
		focused := h.app.MCPFocusPrev()
		result = map[string]string{"focused": focused}
	case "setFocus":
		result, err = h.handleSetFocus(req.Params)
	case "getFocusableIDs":
		result = map[string]any{"ids": h.app.MCPGetFocusableIDs()}
	case "getFrame":
		result = map[string]any{
			"focusedID": h.app.MCPGetFocusedID(),
			"screen":    h.app.MCPGetScreenText(),
		}
	case "debug.toggleDevTools":
		visible := h.app.MCPToggleDevTools()
		result = map[string]any{"visible": visible}
	case "getVersion":
		result = map[string]string{"version": h.app.MCPGetVersion()}
	case "hotReload":
		err = h.app.MCPHotReload()
		if err == nil {
			result = map[string]any{"ok": true}
		}
	default:
		return Response{
			ID:    req.ID,
			Error: &Error{Code: -32601, Message: "unknown method: " + req.Method},
		}
	}

	if err != nil {
		return Response{
			ID:    req.ID,
			Error: &Error{Code: -32602, Message: err.Error()},
		}
	}
	return Response{ID: req.ID, Result: result}
}

// --- method handlers ---

func (h *Handler) handleInspectTree() map[string]any {
	comps := h.app.MCPInspectTree()
	return map[string]any{
		"tree":         comps,
		"focusedID":    h.app.MCPGetFocusedID(),
		"focusableIDs": h.app.MCPGetFocusableIDs(),
	}
}

func (h *Handler) handleInspectComponent(params json.RawMessage) (any, error) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	return h.app.MCPInspectComponent(p.ID)
}

func (h *Handler) handleInspectComponents(params json.RawMessage) map[string]any {
	var p struct {
		Filter string `json:"filter"`
	}
	if params != nil {
		_ = json.Unmarshal(params, &p)
	}
	comps := h.app.MCPInspectComponents(p.Filter)
	return map[string]any{
		"components": comps,
		"total":      len(comps),
	}
}

func (h *Handler) handleGetComponentProps(params json.RawMessage) (any, error) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	props, err := h.app.MCPGetComponentProps(p.ID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"props": props}, nil
}

func (h *Handler) handleGetState(params json.RawMessage) (any, error) {
	var p struct {
		ID  string `json:"id"`
		Key string `json:"key"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	val, err := h.app.MCPGetState(p.ID, p.Key)
	if err != nil {
		return nil, err
	}
	if p.Key != "" {
		return map[string]any{"value": val}, nil
	}
	return map[string]any{"state": val}, nil
}

func (h *Handler) handleSetState(params json.RawMessage) (any, error) {
	var p struct {
		ID    string `json:"id"`
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if err := h.app.MCPSetState(p.ID, p.Key, p.Value); err != nil {
		return nil, err
	}
	return map[string]any{"ok": true}, nil
}

func (h *Handler) handleSimulateClick(params json.RawMessage) (any, error) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if err := h.app.MCPSimulateClick(p.ID); err != nil {
		return nil, err
	}
	return map[string]any{"clicked": p.ID}, nil
}

func (h *Handler) handleSimulateKey(params json.RawMessage) (any, error) {
	var p struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	if err := h.app.MCPSimulateKey(p.Key); err != nil {
		return nil, err
	}
	return map[string]any{"key": p.Key}, nil
}

func (h *Handler) handleEval(params json.RawMessage) (any, error) {
	var p struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	return h.app.MCPEval(p.Code)
}

func (h *Handler) handleSetFocus(params json.RawMessage) (any, error) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	h.app.MCPSetFocus(p.ID)
	return map[string]any{"focused": p.ID}, nil
}

// --- Tool definitions ---

// Tools returns the list of available MCP tools.
func (h *Handler) Tools() []Tool {
	noParams := map[string]any{
		"type":                 "object",
		"properties":          map[string]any{},
		"additionalProperties": false,
	}
	idParam := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{"type": "string", "description": "Component ID"},
		},
		"required":             []string{"id"},
		"additionalProperties": false,
	}

	return []Tool{
		{
			Name:        "lumina.inspectTree",
			Title:       "Inspect component tree",
			Description: "List all registered components with IDs, names, focus state, and rects.",
			InputSchema: noParams,
		},
		{
			Name:        "lumina.inspectComponent",
			Title:       "Inspect component",
			Description: "Get detailed component info: props, state, hooks, children, render count, dirty/focused state.",
			InputSchema: idParam,
		},
		{
			Name:        "lumina.inspectComponents",
			Title:       "List all components",
			Description: "List all rendered Lux components with props, state, hooks, and render metrics. Optional name filter.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"filter": map[string]any{"type": "string", "description": "Optional component name filter (substring match)"},
				},
				"additionalProperties": false,
			},
		},
		{
			Name:        "lumina.getComponentProps",
			Title:       "Get component props",
			Description: "Get the current props of a specific component by ID.",
			InputSchema: idParam,
		},
		{
			Name:        "lumina.getState",
			Title:       "Get component state",
			Description: "Read a state key from a component. If key is empty, returns all state.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":  map[string]any{"type": "string", "description": "Component ID"},
					"key": map[string]any{"type": "string", "description": "State key (empty = all)"},
				},
				"required":             []string{"id"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "lumina.setState",
			Title:       "Set component state",
			Description: "Set a state key on a component. Marks the component dirty for re-render.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"id":    map[string]any{"type": "string", "description": "Component ID"},
					"key":   map[string]any{"type": "string", "description": "State key"},
					"value": map[string]any{"description": "Value to set"},
				},
				"required":             []string{"id", "key", "value"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "lumina.simulateClick",
			Title:       "Simulate click",
			Description: "Simulate a click event on a VNode by ID.",
			InputSchema: idParam,
		},
		{
			Name:        "lumina.simulateKey",
			Title:       "Simulate key press",
			Description: "Simulate a keydown event with the given key name.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key": map[string]any{"type": "string", "description": "Key name (e.g. 'Enter', 'Tab', 'a')"},
				},
				"required":             []string{"key"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "lumina.eval",
			Title:       "Eval Lua",
			Description: "Execute Lua code inside the Lumina VM. Use only in trusted dev environments.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"code": map[string]any{"type": "string", "description": "Lua code to execute"},
				},
				"required":             []string{"code"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "lumina.focusNext",
			Title:       "Focus next",
			Description: "Move focus to the next focusable VNode.",
			InputSchema: noParams,
		},
		{
			Name:        "lumina.focusPrev",
			Title:       "Focus previous",
			Description: "Move focus to the previous focusable VNode.",
			InputSchema: noParams,
		},
		{
			Name:        "lumina.setFocus",
			Title:       "Set focus",
			Description: "Set focus to a specific VNode by ID.",
			InputSchema: idParam,
		},
		{
			Name:        "lumina.getFocusableIDs",
			Title:       "List focusable IDs",
			Description: "Return the ordered list of focusable VNode IDs.",
			InputSchema: noParams,
		},
		{
			Name:        "lumina.getFrame",
			Title:       "Get current frame",
			Description: "Get the current screen content as text plus focus info.",
			InputSchema: noParams,
		},
		{
			Name:        "lumina.debug.toggleDevTools",
			Title:       "Toggle DevTools",
			Description: "Toggle the DevTools panel visibility.",
			InputSchema: noParams,
		},
		{
			Name:        "lumina.getVersion",
			Title:       "Get version",
			Description: "Get the Lumina version string.",
			InputSchema: noParams,
		},
		{
			Name:        "lumina.hotReload",
			Title:       "Hot reload",
			Description: "Trigger immediate hot reload of the current script (same as file watcher detecting a change).",
			InputSchema: noParams,
		},
	}
}
