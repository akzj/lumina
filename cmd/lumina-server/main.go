// Package main is an MCP (Model Context Protocol) stdio server for Lumina.
// It exposes Lumina debugging/control operations as MCP tools.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/akzj/lumina/pkg/lumina"
)

type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id,omitempty"`
	Result  any         `json:"result,omitempty"`
	Error   *jsonrpcErr `json:"error,omitempty"`
}

type jsonrpcErr struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

type mcpTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	// Create Lumina app
	app := lumina.NewApp()

	// Load script if provided (non-blocking load, no render loop)
	scriptPath := ""
	if len(os.Args) > 1 {
		scriptPath = os.Args[1]
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", scriptPath)
			os.Exit(1)
		}
		if err := app.L.DoFile(scriptPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading script: %v\n", err)
			os.Exit(1)
		}
	}

	// Process MCP JSON-RPC requests (one per line over stdio).
	scanner := bufio.NewScanner(os.Stdin)
	// Increase maximum message size for large frames / inspection payloads.
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			sendResponse(jsonrpcResponse{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &jsonrpcErr{Code: -32700, Message: "Parse error"},
			})
			continue
		}

		// Notifications have no id; ignore any we don't recognize.
		if req.ID == nil {
			continue
		}

		resp := handleRequest(app, req)
		sendResponse(resp)
	}
}

func handleRequest(app *lumina.App, req jsonrpcRequest) jsonrpcResponse {
	if req.JSONRPC != "" && req.JSONRPC != "2.0" {
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcErr{Code: -32600, Message: "Invalid Request: unsupported jsonrpc version"},
		}
	}

	switch req.Method {
	case "initialize":
		return handleInitialize(req)
	case "ping":
		return jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
	case "tools/list":
		return handleToolsList(req)
	case "tools/call":
		return handleToolsCall(app, req)
	default:
		// Compatibility: allow calling Lumina custom methods directly via JSON-RPC.
		if strings.HasPrefix(req.Method, "lumina.") {
			return callLuminaMethod(app, req, strings.TrimPrefix(req.Method, "lumina."))
		}
		return jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcErr{Code: -32601, Message: "Method not found"}}
	}
}

func handleInitialize(req jsonrpcRequest) jsonrpcResponse {
	// Per MCP lifecycle: client calls initialize, then sends notifications/initialized (which we ignore).
	// We accept any client protocolVersion and respond with a server protocol version supported by us.
	var params struct {
		ProtocolVersion string         `json:"protocolVersion"`
		Capabilities    map[string]any `json:"capabilities,omitempty"`
		ClientInfo      map[string]any `json:"clientInfo,omitempty"`
	}
	_ = json.Unmarshal(req.Params, &params)

	_ = params.Capabilities
	_ = params.ClientInfo

	result := map[string]any{
		"protocolVersion": "2025-06-18",
		"serverInfo": map[string]any{
			"name":    "lumina",
			"version": lumina.ModuleName,
		},
		"capabilities": map[string]any{
			"tools": map[string]any{
				// We keep tool set static for now.
				"listChanged": false,
			},
		},
	}

	// Echo client protocolVersion if it matches the server one.
	if params.ProtocolVersion != "" && params.ProtocolVersion == "2025-06-18" {
		result["protocolVersion"] = params.ProtocolVersion
	}

	return jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func handleToolsList(req jsonrpcRequest) jsonrpcResponse {
	tools := []mcpTool{
		{
			Name:        "lumina.inspectTree",
			Title:       "Inspect component tree",
			Description: "List all registered components and focus state.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.inspectComponent",
			Title:       "Inspect component",
			Description: "Get details (state/props/focus/dirty) for a component by id.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []string{"id"}, "additionalProperties": false},
		},
		{
			Name:        "lumina.inspectStyles",
			Title:       "Inspect component styles",
			Description: "Get computed style-like props for a component by id.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []string{"id"}, "additionalProperties": false},
		},
		{
			Name:        "lumina.simulateClick",
			Title:       "Simulate click",
			Description: "Simulate a click event on a component by id.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []string{"id"}, "additionalProperties": false},
		},
		{
			Name:        "lumina.getState",
			Title:       "Get component state",
			Description: "Read a state key from a component.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}, "key": map[string]any{"type": "string"}}, "required": []string{"id", "key"}, "additionalProperties": false},
		},
		{
			Name:        "lumina.setState",
			Title:       "Set component state",
			Description: "Set a state key on a component.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}, "key": map[string]any{"type": "string"}, "value": map[string]any{}}, "required": []string{"id", "key", "value"}, "additionalProperties": false},
		},
		{
			Name:        "lumina.focusNext",
			Title:       "Focus next",
			Description: "Move focus to the next focusable component.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.focusPrev",
			Title:       "Focus previous",
			Description: "Move focus to the previous focusable component.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.getFocusableIDs",
			Title:       "List focusable component ids",
			Description: "Return the current focusable ids list.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.getFrame",
			Title:       "Get current frame",
			Description: "Get the latest rendered frame as JSON.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.eval",
			Title:       "Eval Lua (unsafe)",
			Description: "Execute arbitrary Lua code inside the Lumina VM. Use only in trusted dev environments.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{"code": map[string]any{"type": "string"}}, "required": []string{"code"}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.timeline",
			Title:       "Debug timeline",
			Description: "Get recent render events timeline.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.performance",
			Title:       "Debug performance metrics",
			Description: "Get aggregate render performance metrics.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.snapshot",
			Title:       "Take snapshot",
			Description: "Capture a component state snapshot by id.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []string{"id"}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.snapshots",
			Title:       "List snapshots",
			Description: "List captured snapshots.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.restore",
			Title:       "Restore snapshot",
			Description: "Restore component state from a snapshot id.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{"snapshot_id": map[string]any{"type": "integer"}}, "required": []string{"snapshot_id"}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.log",
			Title:       "Debug log",
			Description: "Get structured debug log entries.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.reset",
			Title:       "Reset debug data",
			Description: "Reset performance metrics, snapshots, and debug log.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.checkInspectorBounds",
			Title:       "Check inspector panel bounds",
			Description: "Scan the last rendered frame and report any writes that break the inspector right border.",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
		{
			Name:        "lumina.debug.toggleInspector",
			Title:       "Toggle inspector panel",
			Description: "Toggle the DevTools Inspector panel visibility (same as F12).",
			InputSchema: map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false},
		},
	}

	return jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"tools": tools,
		},
	}
}

func handleToolsCall(app *lumina.App, req jsonrpcRequest) jsonrpcResponse {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil || params.Name == "" {
		return jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcErr{Code: -32602, Message: "Invalid params"}}
	}

	method := strings.TrimPrefix(params.Name, "lumina.")
	if method == params.Name {
		return jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcErr{Code: -32601, Message: "Unknown tool"}}
	}

	// Route debug tools to debug handler.
	if strings.HasPrefix(method, "debug.") {
		return callLuminaDebug(app, req, method, params.Arguments)
	}
	return callLuminaMethodWithArgs(app, req, method, params.Arguments)
}

func callLuminaMethod(app *lumina.App, req jsonrpcRequest, method string) jsonrpcResponse {
	lreq := lumina.MCPRequest{Method: method, Params: req.Params, ID: req.ID}
	lresp := app.HandleMCPRequest(lreq)
	return wrapLuminaResult(req.ID, lresp.Result, lresp.Error)
}

func callLuminaMethodWithArgs(app *lumina.App, req jsonrpcRequest, method string, args map[string]any) jsonrpcResponse {
	raw, _ := json.Marshal(args)
	lreq := lumina.MCPRequest{Method: method, Params: raw, ID: req.ID}
	lresp := app.HandleMCPRequest(lreq)
	return wrapLuminaResult(req.ID, lresp.Result, lresp.Error)
}

func callLuminaDebug(app *lumina.App, req jsonrpcRequest, method string, args map[string]any) jsonrpcResponse {
	raw, _ := json.Marshal(args)
	lreq := lumina.MCPRequest{Method: method, Params: raw, ID: req.ID}
	if res, handled := lumina.HandleDebugMCPRequest(lreq); handled {
		return wrapToolContent(req.ID, res, false)
	}
	return jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcErr{Code: -32601, Message: "Unknown tool"}}
}

func wrapLuminaResult(id any, result any, err *lumina.MCPError) jsonrpcResponse {
	if err != nil {
		return jsonrpcResponse{JSONRPC: "2.0", ID: id, Error: &jsonrpcErr{Code: err.Code, Message: err.Message}}
	}
	return wrapToolContent(id, result, false)
}

func wrapToolContent(id any, v any, isError bool) jsonrpcResponse {
	text := toPrettyJSON(v)
	if text == "" {
		text = fmt.Sprintf("%v", v)
	}
	return jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]any{
			"content": []mcpTextContent{{Type: "text", Text: text}},
			"isError": isError,
		},
	}
}

func toPrettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// sendResponse sends a JSON-RPC response to stdout.
func sendResponse(resp jsonrpcResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}