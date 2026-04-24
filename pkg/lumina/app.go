// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/akzj/go-lua/pkg/lua"
)

// App represents an interactive Lumina application.
type App struct {
	L          *lua.State
	RenderLoop *RenderLoop
}

// NewApp creates a new interactive Lumina application.
func NewApp() *App {
	// Create Lua state with standard libraries
	L := lua.NewState()

	// Open Lumina module
	Open(L)

	// Create render loop with default terminal size
	app := &App{
		L:          L,
		RenderLoop: NewRenderLoop(L, 80, 24),
	}

	return app
}

// NewAppWithSize creates a new app with custom terminal size.
func NewAppWithSize(width, height int) *App {
	L := lua.NewState()
	Open(L)

	return &App{
		L:          L,
		RenderLoop: NewRenderLoop(L, width, height),
	}
}

// Run executes a Lua script and starts the render loop.
// This blocks until Stop() is called or an error occurs.
func (app *App) Run(scriptPath string) error {
	// Check for script argument
	if scriptPath == "" {
		fmt.Println("Lumina v" + ModuleName + " - Terminal React for AI Agents")
		fmt.Println("Usage: lumina <script.lua>")
		os.Exit(0)
	}

	// Verify the file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", scriptPath)
	}

	// Execute the user script
	if err := app.L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("script error: %w", err)
	}

	// Initial render of all components
	app.RenderLoop.InitialRender()

	// Start the render loop (this blocks)
	app.RenderLoop.Start()

	return nil
}

// RunInteractive runs the app with stdin event handling.
// This is the main entry point for interactive apps.
func (app *App) RunInteractive(scriptPath string) error {
	// Load script
	if scriptPath != "" {
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", scriptPath)
		}
		if err := app.L.DoFile(scriptPath); err != nil {
			return fmt.Errorf("script error: %w", err)
		}
	}

	// Initial render
	app.RenderLoop.InitialRender()

	// Start render loop
	app.RenderLoop.Start()

	// TODO: Start event handling loop for keyboard/mouse
	// For now, just keep the render loop running
	select {}
}

// Stop stops the application.
func (app *App) Stop() {
	if app.RenderLoop != nil {
		app.RenderLoop.Stop()
	}
}

// Close closes the application and cleans up resources.
func (app *App) Close() {
	app.Stop()
	if app.L != nil {
		app.L.Close()
	}
}

// MCPRequest represents a JSON-RPC style request from an AI agent.
type MCPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
	ID     interface{}      `json:"id,omitempty"`
}

// MCPResponse represents a JSON-RPC style response.
type MCPResponse struct {
	ID     interface{} `json:"id,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an error response.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HandleMCPRequest handles an MCP request and returns a response.
func (app *App) HandleMCPRequest(req MCPRequest) MCPResponse {
	var result interface{}
	var errMsg string

	switch req.Method {
	case "inspectTree":
		result = app.mcpInspectTree()
	case "inspectComponent":
		var params struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpInspectComponent(params.ID)
		} else {
			errMsg = "invalid params for inspectComponent"
		}
	case "inspectStyles":
		var params struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpInspectStyles(params.ID)
		} else {
			errMsg = "invalid params for inspectStyles"
		}
	case "simulateClick":
		var params struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpSimulateClick(params.ID)
		} else {
			errMsg = "invalid params for simulateClick"
		}
	case "eval":
		var params struct {
			Code string `json:"code"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpEval(params.Code)
		} else {
			errMsg = "invalid params for eval"
		}
	case "getState":
		var params struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpGetState(params.ID, params.Key)
		} else {
			errMsg = "invalid params for getState"
		}
	case "setState":
		var params struct {
			ID    string      `json:"id"`
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		}
		if json.Unmarshal(req.Params, &params) == nil {
			result = app.mcpSetState(params.ID, params.Key, params.Value)
		} else {
			errMsg = "invalid params for setState"
		}
	case "focusNext":
		app.mcpFocusNext()
		result = map[string]string{"focused": globalEventBus.GetFocused()}
	case "focusPrev":
		app.mcpFocusPrev()
		result = map[string]string{"focused": globalEventBus.GetFocused()}
	case "getFocusableIDs":
		result = map[string]interface{}{"ids": globalEventBus.GetFocusableIDs()}
	case "getFrame":
		result = app.mcpGetFrame()
	case "getVersion":
		result = map[string]string{"version": ModuleName}
	default:
		errMsg = "unknown method: " + req.Method
	}

	if errMsg != "" {
		return MCPResponse{
			ID:    req.ID,
			Error: &MCPError{Code: -32601, Message: errMsg},
		}
	}

	return MCPResponse{
		ID:     req.ID,
		Result: result,
	}
}

// mcpInspectTree returns the component tree.
func (app *App) mcpInspectTree() map[string]interface{} {
	tree := []map[string]interface{}{}

	globalRegistry.mu.RLock()
	for id, comp := range globalRegistry.components {
		tree = append(tree, map[string]interface{}{
			"id":      id,
			"type":    comp.Type,
			"name":    comp.Name,
			"focused": id == globalEventBus.GetFocused(),
		})
	}
	globalRegistry.mu.RUnlock()

	return map[string]interface{}{
		"tree":         tree,
		"focusedID":    globalEventBus.GetFocused(),
		"focusableIDs": globalEventBus.GetFocusableIDs(),
	}
}

// mcpInspectComponent returns details of a specific component.
func (app *App) mcpInspectComponent(id string) map[string]interface{} {
	comp, ok := GetComponentByID(id)
	if !ok {
		return map[string]interface{}{"error": "component not found"}
	}

	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	return map[string]interface{}{
		"id":      comp.ID,
		"type":    comp.Type,
		"name":    comp.Name,
		"state":   comp.State,
		"props":   comp.Props,
		"focused": id == globalEventBus.GetFocused(),
		"dirty":   comp.Dirty.Load(),
	}
}

// mcpInspectStyles returns computed styles for a component.
func (app *App) mcpInspectStyles(id string) map[string]interface{} {
	comp, ok := GetComponentByID(id)
	if !ok {
		return map[string]interface{}{"error": "component not found"}
	}

	// Return component props as styles (simplified)
	return map[string]interface{}{
		"id":     id,
		"styles": comp.Props,
	}
}

// mcpSimulateClick simulates a click on a component.
func (app *App) mcpSimulateClick(id string) map[string]interface{} {
	globalEventBus.Emit(&Event{
		Type:   "click",
		Target: id,
	})
	return map[string]interface{}{"clicked": id}
}

// mcpEval evaluates Lua code.
func (app *App) mcpEval(code string) map[string]interface{} {
	if err := app.L.DoString(code); err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return map[string]interface{}{"ok": true}
}

// mcpGetState returns component state.
func (app *App) mcpGetState(id, key string) map[string]interface{} {
	comp, ok := GetComponentByID(id)
	if !ok {
		return map[string]interface{}{"error": "component not found"}
	}

	if key != "" {
		value, exists := comp.GetState(key)
		if !exists {
			return map[string]interface{}{"error": "key not found"}
		}
		return map[string]interface{}{"value": value}
	}

	return map[string]interface{}{"state": comp.State}
}

// mcpSetState sets component state.
func (app *App) mcpSetState(id, key string, value interface{}) map[string]interface{} {
	comp, ok := GetComponentByID(id)
	if !ok {
		return map[string]interface{}{"error": "component not found"}
	}

	comp.SetState(key, value)
	return map[string]interface{}{"ok": true}
}

// mcpFocusNext moves focus to next component.
func (app *App) mcpFocusNext() {
	globalEventBus.FocusNext()
}

// mcpFocusPrev moves focus to previous component.
func (app *App) mcpFocusPrev() {
	globalEventBus.FocusPrev()
}

// mcpGetFrame returns the current frame.
func (app *App) mcpGetFrame() map[string]interface{} {
	return map[string]interface{}{
		"focusedID":  globalEventBus.GetFocused(),
		"componentCount": len(globalRegistry.components),
	}
}
