// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
)

// AppEvent represents an event posted to the main event loop.
type AppEvent struct {
	Type    string
	Payload any
}

// LuaCallbackEvent carries a Lua registry reference + event data.
type LuaCallbackEvent struct {
	RefID int
	Event *Event
}

// App represents an interactive Lumina application.
// All Lua State operations happen on the goroutine that calls Run().
type App struct {
	L       *lua.State
	sched   *lua.Scheduler // async coroutine scheduler
	events  chan AppEvent
	width   int
	height  int
	running bool
}

// NewApp creates a new interactive Lumina application.
func NewApp() *App {
	L := lua.NewState()

	app := &App{
		L:      L,
		sched:  lua.NewScheduler(L),
		events: make(chan AppEvent, 256),
		width:  80,
		height: 24,
	}

	// Store app reference on the Lua state for event handlers
	L.SetUserValue("lumina_app", app)

	// Open Lumina module
	Open(L)

	return app
}

// NewAppWithSize creates a new app with custom terminal size.
func NewAppWithSize(width, height int) *App {
	L := lua.NewState()

	app := &App{
		L:      L,
		sched:  lua.NewScheduler(L),
		events: make(chan AppEvent, 256),
		width:  width,
		height: height,
	}

	L.SetUserValue("lumina_app", app)
	Open(L)

	return app
}

// PostEvent sends an event to the main event loop (goroutine-safe).
func (app *App) PostEvent(event AppEvent) {
	select {
	case app.events <- event:
	default:
		// Drop if channel full (non-blocking)
	}
}

// Run executes a Lua script and starts the single-threaded event loop.
// This blocks until Stop() is called or an error occurs.
// ALL Lua State operations happen on this goroutine.
func (app *App) Run(scriptPath string) error {
	if scriptPath == "" {
		fmt.Println("Lumina v" + ModuleName + " - Terminal React for AI Agents")
		fmt.Println("Usage: lumina <script.lua>")
		os.Exit(0)
	}

	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", scriptPath)
	}

	// Execute the user script (on main thread — safe)
	if err := app.L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("script error: %w", err)
	}

	// Initial render of all components
	app.renderAllDirty()

	// Start the single-threaded event loop
	return app.eventLoop()
}

// RunInteractive runs the app with real terminal input (raw mode).
func (app *App) RunInteractive(scriptPath string) error {
	// Set up terminal
	term, err := NewTerminal()
	if err != nil {
		return fmt.Errorf("terminal init: %w", err)
	}

	// Get terminal size
	w, h, _ := term.GetSize()
	app.width = w
	app.height = h

	// Enable raw mode
	if err := term.EnableRawMode(); err != nil {
		return fmt.Errorf("raw mode: %w", err)
	}
	defer term.RestoreMode()

	// Clear screen
	fmt.Print("\x1b[2J\x1b[H")

	// Load script
	if scriptPath != "" {
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			term.RestoreMode()
			return fmt.Errorf("file not found: %s", scriptPath)
		}
		if err := app.L.DoFile(scriptPath); err != nil {
			term.RestoreMode()
			return fmt.Errorf("script error: %w", err)
		}
	}

	// Initial render
	app.renderAllDirty()

	// Start input reader (runs in goroutine, sends to app.events)
	input := NewInputReader(term, app.events)
	input.Start()

	// Event loop (same as eventLoop but inlined for clarity)
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	app.running = true
	for app.running {
		select {
		case <-ticker.C:
			app.sched.Tick()
			app.renderAllDirty()

		case event := <-app.events:
			app.handleEvent(event)
		}
	}

	return nil
}

// eventLoop is the single-threaded event loop.
// All Lua operations happen here on the calling goroutine.
func (app *App) eventLoop() error {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60fps
	defer ticker.Stop()

	app.running = true
	for app.running {
		select {
		case <-ticker.C:
			app.sched.Tick()    // process async coroutines
			app.renderAllDirty()

		case event := <-app.events:
			app.handleEvent(event)
		}
	}
	return nil
}

// handleEvent dispatches an event on the main thread.
func (app *App) handleEvent(event AppEvent) {
	switch event.Type {
	case "quit":
		app.running = false

	case "input_event":
		e, ok := event.Payload.(*Event)
		if !ok {
			return
		}

		// Handle Ctrl+C / Ctrl+Q to quit
		if e.Type == "keydown" && e.Modifiers.Ctrl && (e.Key == "c" || e.Key == "q") {
			app.running = false
			return
		}

		// Dispatch to EventBus (handles focus, shortcuts, registered handlers)
		globalEventBus.Emit(e)

		// Handle keyboard navigation (Tab, Enter, Escape, etc.)
		if e.Type == "keydown" {
			globalEventBus.HandleKeyEvent(e.Key, e.Modifiers)
		}

		// Handle scroll events (mouse wheel and PageUp/PageDown)
		app.handleScrollEvent(e)

		// Re-render after input handling
		app.renderAllDirty()

	case "lua_callback":
		cb, ok := event.Payload.(LuaCallbackEvent)
		if !ok {
			return
		}
		app.L.RawGetI(lua.RegistryIndex, int64(cb.RefID))
		if app.L.Type(-1) == lua.TypeFunction {
			pushEventToLua(app.L, cb.Event)
			status := app.L.PCall(1, 0, 0)
			if status != lua.OK {
				app.L.Pop(1) // pop error
			}
		} else {
			app.L.Pop(1) // pop non-function
		}
		// Re-render after event handling (state may have changed)
		app.renderAllDirty()
	}
}

// handleScrollEvent handles scroll-related events (mouse wheel, PageUp/PageDown).
func (app *App) handleScrollEvent(e *Event) {
	switch e.Type {
	case "scroll":
		// Mouse wheel scroll — find the scrollable container under the cursor
		// e.Y contains the scroll direction: negative = up, positive = down
		// For mouse scroll, we use the focused container or find by position
		focusedID := globalEventBus.GetFocused()
		if focusedID != "" {
			ScrollViewport(focusedID, e.Y)
		}

	case "keydown":
		focusedID := globalEventBus.GetFocused()
		if focusedID == "" {
			return
		}

		switch e.Key {
		case "PageUp":
			viewportMu.RLock()
			vp, ok := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if ok {
				vp.ScrollUp(vp.ViewH)
			}
		case "PageDown":
			viewportMu.RLock()
			vp, ok := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if ok {
				vp.ScrollDown(vp.ViewH)
			}
		case "Home":
			viewportMu.RLock()
			vp, ok := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if ok {
				vp.ScrollToTop()
			}
		case "End":
			viewportMu.RLock()
			vp, ok := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if ok {
				vp.ScrollToBottom()
			}
		}
	}
}

// renderAllDirty checks all components for dirty state and re-renders.
func (app *App) renderAllDirty() {
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
		if comp.Dirty.Load() {
			app.renderComponent(comp, adapter)
		}
	}
}

// renderComponent re-renders a single component on the main thread.
func (app *App) renderComponent(comp *Component, adapter OutputAdapter) {
	SetCurrentComponent(comp)
	defer SetCurrentComponent(nil)

	if !comp.PushRenderFn(app.L) {
		return
	}

	status := app.L.PCall(0, 1, 0)
	if status != lua.OK {
		app.L.Pop(1)
		return
	}

	frame := RenderLuaVNode(app.L, -1, app.width, app.height)
	app.L.Pop(1)

	frame.FocusedID = globalEventBus.GetFocused()
	adapter.Write(frame)
	comp.MarkClean()
}

// InitialRender renders all components once (for testing/compatibility).
func (app *App) InitialRender() {
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

		if !comp.PushRenderFn(app.L) {
			continue
		}

		status := app.L.PCall(0, 1, 0)
		if status != lua.OK {
			app.L.Pop(1)
			continue
		}

		frame := RenderLuaVNode(app.L, -1, app.width, app.height)
		app.L.Pop(1)
		frame.FocusedID = globalEventBus.GetFocused()
		adapter.Write(frame)
	}
	SetCurrentComponent(nil)
}

// Stop stops the application by posting a quit event.
func (app *App) Stop() {
	app.PostEvent(AppEvent{Type: "quit"})
}

// Close closes the application and cleans up resources.
func (app *App) Close() {
	if app.L != nil {
		app.L.Close()
	}
}

// Scheduler returns the App's async coroutine scheduler.
func (app *App) Scheduler() *lua.Scheduler {
	return app.sched
}

// GetApp retrieves the App from a Lua State's user values.
// Works correctly even from Go functions called by the Lua VM because
// UserValue is stored on the internal api.State (survives wrapFunction).
func GetApp(L *lua.State) *App {
	if v := L.UserValue("lumina_app"); v != nil {
		if app, ok := v.(*App); ok {
			return app
		}
	}
	return nil
}

// MCPRequest represents a JSON-RPC style request from an AI agent.
type MCPRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
	ID     interface{}     `json:"id,omitempty"`
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
	app.L.GetGlobal("lumina")
	app.L.SetGlobal("lumina")

	if err := app.L.DoString(code); err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	n := app.L.GetTop()
	if n == 0 {
		return map[string]interface{}{"ok": true}
	}

	results := make([]interface{}, n)
	for i := 1; i <= n; i++ {
		results[i-1] = luaValueToInterface(app.L, i)
	}
	app.L.Pop(n)

	return map[string]interface{}{"results": results}
}

func luaValueToInterface(L *lua.State, index int) interface{} {
	switch L.Type(index) {
	case lua.TypeString:
		if v, ok := L.ToString(index); ok {
			return v
		}
		return nil
	case lua.TypeNumber:
		if L.IsInteger(index) {
			v, _ := L.ToInteger(index)
			return v
		}
		v, _ := L.ToNumber(index)
		return v
	case lua.TypeBoolean:
		return L.ToBoolean(index)
	case lua.TypeTable:
		return "table"
	case lua.TypeNil:
		return nil
	default:
		return fmt.Sprintf("unknown(%s)", L.TypeName(L.Type(index)))
	}
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
		"focusedID":      globalEventBus.GetFocused(),
		"componentCount": len(globalRegistry.components),
	}
}
