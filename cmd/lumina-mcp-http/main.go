// Package main runs Lumina with an MCP Streamable HTTP endpoint.
//
// This is meant for remote / multi-client debugging: the Lumina process stays running,
// and MCP clients connect over HTTP to call tools.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	var (
		addr       = flag.String("addr", ":8088", "listen address")
		scriptPath = flag.String("script", "", "optional lua entry script path")
	)
	flag.Parse()

	app := lumina.NewApp()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})

	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		var req jsonrpcRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			writeJSON(w, jsonrpcResponse{JSONRPC: "2.0", ID: nil, Error: &jsonrpcErr{Code: -32700, Message: "Parse error"}})
			return
		}

		if req.ID == nil {
			// Ignore notifications, but respond 204 to keep clients happy.
			w.WriteHeader(http.StatusNoContent)
			return
		}

		resp := handleRequest(app, req)

		writeJSON(w, resp)
	})

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("lumina MCP HTTP listening on %s (POST /mcp)", *addr)
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	// Run interactive UI in the current terminal.
	if *scriptPath == "" {
		log.Fatal("missing -script")
	}
	if _, err := os.Stat(*scriptPath); err != nil {
		log.Fatalf("script not found: %s (%v)", *scriptPath, err)
	}
	if err := app.RunInteractive(*scriptPath); err != nil {
		// Fallback to non-interactive mode (headless).
		log.Printf("interactive terminal unavailable (%v); falling back to headless event loop", err)
		if err2 := app.Run(*scriptPath); err2 != nil {
			log.Fatalf("run headless: %v", err2)
		}
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
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
				"listChanged": false,
			},
		},
	}

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

	if strings.HasPrefix(method, "debug.") {
		return callLuminaDebug(app, req, method, params.Arguments)
	}
	return callLuminaMethodWithArgs(app, req, method, params.Arguments)
}

func callLuminaMethod(app *lumina.App, req jsonrpcRequest, method string) jsonrpcResponse {
	lreq := lumina.MCPRequest{Method: method, Params: req.Params, ID: req.ID}
	lresp := callOnAppThread(app, lreq)
	return wrapLuminaResult(req.ID, lresp.Result, lresp.Error)
}

func callLuminaMethodWithArgs(app *lumina.App, req jsonrpcRequest, method string, args map[string]any) jsonrpcResponse {
	raw, _ := json.Marshal(args)
	lreq := lumina.MCPRequest{Method: method, Params: raw, ID: req.ID}
	lresp := callOnAppThread(app, lreq)
	return wrapLuminaResult(req.ID, lresp.Result, lresp.Error)
}

func callLuminaDebug(app *lumina.App, req jsonrpcRequest, method string, args map[string]any) jsonrpcResponse {
	raw, _ := json.Marshal(args)
	lreq := lumina.MCPRequest{Method: method, Params: raw, ID: req.ID}
	lresp := callOnAppThread(app, lreq)
	if lresp.Error != nil {
		return jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &jsonrpcErr{Code: lresp.Error.Code, Message: lresp.Error.Message}}
	}
	return wrapToolContent(req.ID, lresp.Result, false)
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

func callOnAppThread(app *lumina.App, req lumina.MCPRequest) lumina.MCPResponse {
	respCh := make(chan lumina.MCPResponse, 1)
	app.PostEvent(lumina.AppEvent{
		Type:    "mcp_request",
		Payload: lumina.MCPCallEvent{Req: req, Resp: respCh},
	})

	select {
	case resp := <-respCh:
		return resp
	case <-time.After(10 * time.Second):
		return lumina.MCPResponse{ID: req.ID, Error: &lumina.MCPError{Code: -32000, Message: "MCP request timeout"}}
	}
}

