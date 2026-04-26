// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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

// MCPCallEvent is an AppEvent payload used to execute an MCP request on the app's
// main goroutine (the one running the event loop), which is required for Lua safety.
type MCPCallEvent struct {
	Req  MCPRequest
	Resp chan MCPResponse
}

// App represents an interactive Lumina application.
// All Lua State operations happen on the goroutine that calls Run().
type App struct {
	L              *lua.State
	sched          *lua.Scheduler // async coroutine scheduler
	events         chan AppEvent
	width          atomic.Int32 // terminal width (atomic for concurrent resize)
	height         atomic.Int32 // terminal height (atomic for concurrent resize)
	running        bool
	batchDepth     int       // >0 means we're inside a batch (suppress per-setState renders)
	termIO         TermIO    // terminal I/O abstraction (nil = default local)
	lastFrame         *Frame    // previous render frame for incremental updates
	lastRenderTime    time.Time // frame rate limiting
	lastMouseMoveTime time.Time // for mousemove throttling
	lastMouseX        int       // last processed mousemove X
	lastMouseY        int       // last processed mousemove Y
	hoveredID         string    // ID of element currently under mouse cursor
	lastRenderWidth  int     // previous render width for detecting resize
	lastRenderHeight int     // previous render height for detecting resize
}

// getWidth returns the current terminal width (goroutine-safe).
func (app *App) getWidth() int { return int(app.width.Load()) }

// getHeight returns the current terminal height (goroutine-safe).
func (app *App) getHeight() int { return int(app.height.Load()) }

// setSize stores width and height atomically.
func (app *App) setSize(w, h int) {
	app.width.Store(int32(w))
	app.height.Store(int32(h))
}


// NewApp creates a new interactive Lumina application.
func NewApp() *App {
	return NewAppWithSize(80, 24)
}

// NewAppWithSize creates a new app with custom terminal size.
// Resets ALL global singletons for test isolation.
func NewAppWithSize(width, height int) *App {
	// Reset global state so tests don't leak between runs
	ClearComponents()
	globalEventBus = NewEventBus()
	globalAnimationManager = NewAnimationManager()
	globalOverlayManager = NewOverlayManager()
	globalRouter = NewRouter()
	globalDragState = &DragState{}
	globalWindowManager = NewWindowManager(width, height)
	globalStyles = make(map[string]map[string]any)
	globalInspector = &DevToolsInspector{panelWidth: 40}
	globalConsole = &Console{maxSize: 1000}
	globalDevTools = &DevTools{
		renderCounts: make(map[string]int),
		renderTimes:  make(map[string]time.Duration),
	}
	globalI18n = NewI18n("en")
	globalAnnouncer = &Announcer{}
	globalQueryCache = &QueryCache{entries: make(map[string]*QueryEntry)}
	globalHotReloader = NewHotReloader(HotReloadConfig{
		Enabled:  false,
		Interval: 500 * time.Millisecond,
	})
	globalPluginRegistry = NewPluginRegistry()
	globalProfile = &Profile{
		timings: make([]FrameTiming, 0, 1000),
		maxSize: 1000,
	}
	globalThemeManager = &ThemeManager{
		current: CatppuccinMocha,
		styles:  make(map[string]map[string]string),
	}
	ResetRenderRefs()
	scrollBehavior = "instant"

	L := lua.NewState()

	app := &App{
		L:      L,
		sched:  lua.NewScheduler(L),
		events: make(chan AppEvent, 256),
	}
	app.width.Store(int32(width))
	app.height.Store(int32(height))

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
		fmt.Printf("lumina: warning: event dropped (channel full): %s\n", event.Type)
	}
}

// BeginBatch starts a batch update cycle. While batching is active,
// setState calls mark components dirty but don't trigger immediate re-renders.
// Re-renders are deferred until EndBatch.
func (app *App) BeginBatch() {
	app.batchDepth++
}

// EndBatch ends a batch cycle. If this is the outermost batch (depth→0),
// all dirty components are re-rendered once.
func (app *App) EndBatch() {
	if app.batchDepth > 0 {
		app.batchDepth--
	}
	if app.batchDepth == 0 {
		app.renderAllDirty()
	}
}

// IsBatching returns true if we're inside a batch update cycle.
func (app *App) IsBatching() bool {
	return app.batchDepth > 0
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

	// Create local TermIO
	localIO := NewLocalTermIO()

	// Get terminal size
	w, h, _ := term.GetSize()
	app.setSize(w, h)
	localIO.SetSize(w, h)
	app.termIO = localIO

	// Enable raw mode
	if err := term.EnableRawMode(); err != nil {
		return fmt.Errorf("raw mode: %w", err)
	}
	defer term.RestoreMode()

	// Set output adapter to write through TermIO
	SetOutputAdapter(NewANSIAdapter(localIO))

	// Clear screen with theme background color (not terminal default)
	localIO.Write([]byte(clearScreenWithBg()))

	// Load script
	if scriptPath != "" {
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			term.RestoreMode()
			return fmt.Errorf("file not found: %s", scriptPath)
		}
		addScriptDirToPackagePath(app.L, scriptPath)
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

	// Watch for terminal resize (SIGWINCH)
	term.WatchResize(func(w, h int) {
		app.setSize(w, h)
		localIO.SetSize(w, h)
		// Post to main thread — don't touch components/EventBus from SIGWINCH goroutine
		app.PostEvent(AppEvent{Type: "resize"})
	})
	defer term.StopResize()

	// Event loop (same as eventLoop but inlined for clarity)
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	app.running = true
	for app.running {
		select {
		case <-ticker.C:
			app.sched.Tick()
			// Tick smooth scrolling
			TickAllViewports()
			app.renderAllDirty()

		case event := <-app.events:
			app.handleEvent(event)
		}
	}

	return nil
}

// RunWithTermIO runs the app using a custom TermIO (e.g. WebSocket).
// This enables the same Lua app to run over a network connection.
func (app *App) RunWithTermIO(tio TermIO, scriptPath string) error {
	w, h := tio.Size()
	app.setSize(w, h)
	app.termIO = tio

	// Set output adapter to write through the TermIO
	SetOutputAdapter(NewANSIAdapter(tio))

	// Wire resize callback for WebTerminal
	if wt, ok := tio.(*WebTerminal); ok {
		wt.SetOnResize(func(newW, newH int) {
			app.setSize(newW, newH)
			// Post to main thread — don't touch components/EventBus from resize goroutine
			app.PostEvent(AppEvent{Type: "resize"})
		})
	}

	// Load script
	if scriptPath != "" {
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", scriptPath)
		}
		if err := app.L.DoFile(scriptPath); err != nil {
			return fmt.Errorf("script error: %w", err)
		}
	}

	// Clear screen with theme background color (not terminal default)
	tio.Write([]byte(clearScreenWithBg()))

	// Initial render
	app.renderAllDirty()

	// Start input reader from TermIO
	input := NewInputReaderFromIO(tio, app.events)
	input.Start()

	// Event loop
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	app.running = true
	for app.running {
		select {
		case <-ticker.C:
			app.sched.Tick()
			TickAllViewports()
			app.renderAllDirty()

		case event := <-app.events:
			app.handleEvent(event)
		}
	}

	return nil
}

// GetTermIO returns the app's TermIO (may be nil if not running interactively).
func (app *App) GetTermIO() TermIO {
	return app.termIO
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
			app.sched.Tick() // process async coroutines
			// Tick animations — mark owning components dirty
			dirtyComps := globalAnimationManager.Tick(time.Now().UnixMilli())
			for _, compID := range dirtyComps {
				if c, ok := globalRegistry.components[compID]; ok {
					c.Dirty.Store(true)
				}
			}
			// Tick smooth scrolling
			TickAllViewports()
			app.renderAllDirty()

		case event := <-app.events:
			app.handleEvent(event)
		}
	}
	return nil
}

// handleEvent dispatches an event on the main thread.
func (app *App) handleEvent(event AppEvent) {
	app.BeginBatch()
	defer app.EndBatch()

	switch event.Type {
	case "quit":
		app.running = false

	case "resize":
		// Mark all components dirty for re-render at new size
		markAllComponentsDirty()
		// Invalidate ANSI adapter to force full frame rewrite (not diff)
		if adapter := GetOutputAdapter(); adapter != nil {
			if ansi, ok := adapter.(*ANSIAdapter); ok {
				ansi.Invalidate()
			}
		}
		// Clear hover state — old cell IDs may be invalid after resize
		app.hoveredID = ""
		// Invalidate lastFrame — old frame's OwnerNode pointers are stale
		// after resize (grid dimensions changed, cells removed/added).
		// Hit-test must wait for a fresh render.
		app.lastFrame = nil
		// Emit resize event for onResize handlers (now safely on main thread)
		globalEventBus.Emit(&Event{
			Type:      "resize",
			Bubbles:   false,
			Timestamp: time.Now().UnixMilli(),
		})

	case "mcp_request":
		call, ok := event.Payload.(MCPCallEvent)
		if !ok || call.Resp == nil {
			return
		}
		// Debug methods are handled by HandleDebugMCPRequest (stateless).
		if res, handled := HandleDebugMCPRequest(call.Req); handled {
			call.Resp <- MCPResponse{ID: call.Req.ID, Result: res}
			return
		}
		// All other methods go through HandleMCPRequest (needs app/Lua State).
		call.Resp <- app.HandleMCPRequest(call.Req)

	case "input_event":
		e, ok := event.Payload.(*Event)
		if !ok {
			return
		}
		app.runInputPipeline(e)

	case "lua_callback":
		cb, ok := event.Payload.(LuaCallbackEvent)
		if !ok {
			return
		}
		app.invokeLuaCallback(cb.RefID, cb.Event)
		// Re-render deferred to EndBatch()
	}
}

func (app *App) ReloadScript(scriptPath string) error {
	// 1. Snapshot all component states
	globalHotReloader.SnapshotAllComponents()

	// 2. Clear component registry (but keep snapshots)
	ClearComponents()

	// 3. Re-execute Lua script
	if err := app.L.DoFile(scriptPath); err != nil {
		return err
	}

	// 4. Restore component states by type name matching
	globalHotReloader.RestoreAllByType()

	// 5. Re-render
	app.renderAllDirty()
	return nil
}

func (app *App) Stop() {
	app.PostEvent(AppEvent{Type: "quit"})
}

// LoadScript loads and executes a Lua script using the given TermIO for output,
// but does NOT start the event loop. Useful for testing: load script, render
// one frame, inspect output.
func (app *App) LoadScript(scriptPath string, tio TermIO) error {
	if scriptPath == "" {
		return fmt.Errorf("empty script path")
	}
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", scriptPath)
	}

	// Configure terminal I/O
	w, h := tio.Size()
	app.setSize(w, h)
	app.termIO = tio

	// Set output adapter to write through the TermIO
	SetOutputAdapter(NewANSIAdapter(tio))

	// Add script directory to package.path for multi-file require
	addScriptDirToPackagePath(app.L, scriptPath)

	// Execute the Lua script (on caller goroutine — safe)
	if err := app.L.DoFile(scriptPath); err != nil {
		return fmt.Errorf("script error: %w", err)
	}

	return nil
}

func (app *App) Close() {
	if app.L != nil {
		app.L.Close()
	}
}

func (app *App) ProcessPendingEvents() {
	for {
		select {
		case event := <-app.events:
			app.handleEvent(event)
		default:
			return
		}
	}
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

func clearScreenWithBg() string {
	bg := "#1E1E2E" // Catppuccin Mocha base (fallback)
	if theme := GetCurrentTheme(); theme != nil {
		if tbg, ok := theme.Colors["background"]; ok && tbg != "" {
			bg = tbg
		}
	}
	r, g, b := hexToRGB(bg)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm\x1b[2J\x1b[H", r, g, b)
}

// addScriptDirToPackagePath prepends the script's directory to Lua's package.path
// so that multi-file require() calls resolve relative to the script, not the CWD.
func addScriptDirToPackagePath(L *lua.State, scriptPath string) {
	scriptDir := filepath.Dir(scriptPath)
	absDir, err := filepath.Abs(scriptDir)
	if err != nil {
		return
	}
	// Escape backslashes for Windows paths in Lua string
	escaped := strings.ReplaceAll(absDir, `\`, `\\`)
	code := fmt.Sprintf(`package.path = "%s/?.lua;%s/?/init.lua;" .. package.path`, escaped, escaped)
	L.DoString(code)
}
