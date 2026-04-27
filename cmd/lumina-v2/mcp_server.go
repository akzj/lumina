//go:build linux || darwin

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/akzj/lumina/pkg/lumina/v2/mcp"
)

// jsonrpcRequest is a JSON-RPC 2.0 request envelope.
type jsonrpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// jsonrpcResponse is a JSON-RPC 2.0 response envelope.
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

type mcpTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// serveMCP starts an HTTP server that exposes MCP tools via JSON-RPC 2.0.
// It blocks, so call it in a goroutine.
func serveMCP(handler *mcp.Handler, addr string) {
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResp(w, jsonrpcResponse{
				JSONRPC: "2.0",
				Error:   &jsonrpcErr{Code: -32700, Message: "Parse error"},
			})
			return
		}

		if req.ID == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		resp := handleMCPRequest(handler, req)
		writeJSONResp(w, resp)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("lumina v2 MCP HTTP listening on %s (POST /mcp)", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("MCP server error: %v", err)
	}
}

func handleMCPRequest(handler *mcp.Handler, req jsonrpcRequest) jsonrpcResponse {
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
		return handleToolsList(handler, req)
	case "tools/call":
		return handleToolsCall(handler, req)
	default:
		// Compatibility: allow calling Lumina methods directly via JSON-RPC.
		if strings.HasPrefix(req.Method, "lumina.") {
			method := strings.TrimPrefix(req.Method, "lumina.")
			mcpReq := mcp.Request{ID: req.ID, Method: method, Params: req.Params}
			mcpResp := handler.Handle(mcpReq)
			return wrapMCPResponse(req.ID, mcpResp)
		}
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcErr{Code: -32601, Message: "Method not found"},
		}
	}
}

func handleInitialize(req jsonrpcRequest) jsonrpcResponse {
	var params struct {
		ProtocolVersion string `json:"protocolVersion"`
	}
	_ = json.Unmarshal(req.Params, &params)

	result := map[string]any{
		"protocolVersion": "2025-06-18",
		"serverInfo": map[string]any{
			"name":    "lumina-v2",
			"version": "v2",
		},
		"capabilities": map[string]any{
			"tools": map[string]any{
				"listChanged": false,
			},
		},
	}

	if params.ProtocolVersion == "2025-06-18" {
		result["protocolVersion"] = params.ProtocolVersion
	}

	return jsonrpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
}

func handleToolsList(handler *mcp.Handler, req jsonrpcRequest) jsonrpcResponse {
	return jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]any{
			"tools": handler.Tools(),
		},
	}
}

func handleToolsCall(handler *mcp.Handler, req jsonrpcRequest) jsonrpcResponse {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil || params.Name == "" {
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcErr{Code: -32602, Message: "Invalid params"},
		}
	}

	method := strings.TrimPrefix(params.Name, "lumina.")
	if method == params.Name {
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &jsonrpcErr{Code: -32601, Message: "Unknown tool"},
		}
	}

	raw, _ := json.Marshal(params.Arguments)
	mcpReq := mcp.Request{ID: req.ID, Method: method, Params: raw}
	mcpResp := handler.Handle(mcpReq)
	return wrapToolContent(req.ID, mcpResp)
}

func wrapMCPResponse(id any, resp mcp.Response) jsonrpcResponse {
	if resp.Error != nil {
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error:   &jsonrpcErr{Code: resp.Error.Code, Message: resp.Error.Message},
		}
	}
	return wrapToolContent(id, resp)
}

func wrapToolContent(id any, resp mcp.Response) jsonrpcResponse {
	if resp.Error != nil {
		text := toPrettyJSON(map[string]any{"error": resp.Error.Message})
		return jsonrpcResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: map[string]any{
				"content": []mcpTextContent{{Type: "text", Text: text}},
				"isError": true,
			},
		}
	}
	text := toPrettyJSON(resp.Result)
	if text == "" {
		text = fmt.Sprintf("%v", resp.Result)
	}
	return jsonrpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]any{
			"content": []mcpTextContent{{Type: "text", Text: text}},
			"isError": false,
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

func writeJSONResp(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
