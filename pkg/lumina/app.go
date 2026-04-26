// Package lumina provides Lua bindings for the Lumina UI framework.
package lumina

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	lastFrame      *Frame    // previous render frame for incremental updates
	lastRenderTime time.Time // frame rate limiting
	hoveredID      string    // ID of element currently under mouse cursor
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
		// Mark all components dirty for re-render at new size
		globalRegistry.mu.RLock()
		for _, comp := range globalRegistry.components {
			comp.Dirty.Store(true)
		}
		globalRegistry.mu.RUnlock()
		// Emit resize event for onResize handlers
		globalEventBus.Emit(&Event{
			Type:      "resize",
			Bubbles:   false,
			Timestamp: time.Now().UnixMilli(),
		})
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
			// Mark all components dirty for re-render at new size
			globalRegistry.mu.RLock()
			for _, comp := range globalRegistry.components {
				comp.Dirty.Store(true)
			}
			globalRegistry.mu.RUnlock()
			// Emit resize event for onResize handlers
			globalEventBus.Emit(&Event{
				Type:      "resize",
				Bubbles:   false,
				Timestamp: time.Now().UnixMilli(),
			})
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
				globalRegistry.mu.RLock()
				if c, ok := globalRegistry.components[compID]; ok {
					c.Dirty.Store(true)
				}
				globalRegistry.mu.RUnlock()
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

	case "input_event":
		e, ok := event.Payload.(*Event)
		if !ok {
			return
		}

		// Handle Ctrl+C / Ctrl+Q to quit (always works, even with modal)
		if e.Type == "keydown" && e.Modifiers.Ctrl && (e.Key == "c" || e.Key == "q") {
			app.running = false
			return
		}

		// F12 toggles DevTools inspector
		if e.Type == "keydown" && e.Key == "F12" {
			ToggleInspector()
			// Force re-render
			globalRegistry.mu.RLock()
			for _, comp := range globalRegistry.components {
				comp.Dirty.Store(true)
			}
			globalRegistry.mu.RUnlock()
		}

		// DevTools inspector: update hover/selection on mouse events
		if IsInspectorVisible() {
			switch e.Type {
			case "mousemove":
				if app.lastFrame != nil && e.X >= 0 && e.Y >= 0 &&
					e.X < app.lastFrame.Width && e.Y < app.lastFrame.Height {
					node := app.lastFrame.Cells[e.Y][e.X].OwnerNode
					if node != nil {
						if id, ok := node.Props["id"].(string); ok && id != "" {
							SetInspectorHighlight(id)
						}
					}
				}
			case "mousedown":
				// Click selects element for inspection
				if app.lastFrame != nil && e.X >= 0 && e.Y >= 0 &&
					e.X < app.lastFrame.Width && e.Y < app.lastFrame.Height {
					node := app.lastFrame.Cells[e.Y][e.X].OwnerNode
					if node != nil {
						if id, ok := node.Props["id"].(string); ok && id != "" {
							SetInspectorSelected(id)
						}
					}
				}
			}
		}

		// Modal overlay input routing: when a modal overlay is active,
		// Escape closes it and other events are captured by the modal.
		if topModal := globalOverlayManager.GetTopModal(); topModal != nil {
			if e.Type == "keydown" && e.Key == KeyEscape {
				globalOverlayManager.Hide(topModal.ID)
				return
			}
			// Route event to modal's VNode tree if it has one
			// (target events to the modal overlay, not the base layer)
			if topModal.VNode != nil {
				e.Target = topModal.ID
			}
		}

		// Mouse event hit-testing: map (x,y) to target component
		if e.Type == "mousedown" || e.Type == "mouseup" || e.Type == "mousemove" {
			// O(1) hit-test using Cell.OwnerNode from last rendered frame
			var targetNode *VNode
			if app.lastFrame != nil && e.X >= 0 && e.Y >= 0 &&
				e.X < app.lastFrame.Width && e.Y < app.lastFrame.Height {
				targetNode = app.lastFrame.Cells[e.Y][e.X].OwnerNode
			}

			if targetNode != nil {
				if id, ok := targetNode.Props["id"].(string); ok && id != "" {
					e.Target = id
				} else {
					// Walk up the VNode tree to find nearest ancestor with an id
					tree := globalEventBus.GetVNodeTree()
					if tree != nil {
						for node := tree.Parents[targetNode]; node != nil; node = tree.Parents[node] {
							if id, ok := node.Props["id"].(string); ok && id != "" {
								e.Target = id
								targetNode = node
								break
							}
						}
					}
				}
				e.TargetNode = targetNode
				// Compute local coordinates (relative to target VNode top-left)
				e.LocalX = e.X - targetNode.X
				e.LocalY = e.Y - targetNode.Y
			}

			// Fall back to VNode tree walk if Cell has no owner
			if e.Target == "" {
				if root := app.findRootVNode(); root != nil {
					targetID := HitTestVNode(root, e.X, e.Y)
					if targetID != "" {
						e.Target = targetID
					}
				}
			}

			// Focus on mousedown
			if e.Type == "mousedown" && e.Target != "" {
				globalEventBus.SetFocus(e.Target)
			}

			// Window manager: handle clicks on window chrome (title bar / controls)
			// and mouse drag/resize operations
			if win := globalWindowManager.WindowAtPoint(e.X, e.Y); win != nil {
				localY := e.Y - win.Y
				localX := e.X - win.X

				switch e.Type {
				case "mousedown":
					// Title bar click (row 0)
					if localY == 0 {
						// Check which control button was clicked
						// Button region is right-aligned: [ _ ][ □ ][ X ] = 9 chars
						if localX >= win.W-9 {
							// Close button — last 3 chars " X "
							globalWindowManager.CloseWindow(win.ID)
							return // consume event, don't emit to bus
						} else if localX >= win.W-13 {
							// Maximize button — next 3 chars " □ "
							if win.Maximized {
								globalWindowManager.RestoreWindow(win.ID)
							} else {
								globalWindowManager.MaximizeWindow(win.ID)
							}
							return
						} else if localX >= win.W-16 {
							// Minimize button — next 3 chars " _ "
							globalWindowManager.MinimizeWindow(win.ID)
							return
						}
						// Otherwise: start drag
						globalWindowManager.StartDrag(win.ID, localX, localY)
						globalWindowManager.FocusWindow(win.ID)
						return
					}
					// Resize handle (bottom-right corner)
					if localY == win.H-1 && localX == win.W-1 {
						globalWindowManager.StartResize(win.ID, e.X, e.Y)
						return
					}

				case "mousemove":
					if globalWindowManager.IsDragging() {
						globalWindowManager.UpdateDrag(e.X, e.Y)
						return
					}
					if globalWindowManager.IsResizing() {
						globalWindowManager.UpdateResize(e.X, e.Y)
						return
					}

				case "mouseup":
					if globalWindowManager.IsDragging() {
						globalWindowManager.StopDrag()
						return
					}
					if globalWindowManager.IsResizing() {
						globalWindowManager.StopResize()
						return
					}
				}
			}

			// Drag-and-drop handling
			switch e.Type {
			case "mousedown":
				if e.TargetNode != nil {
					if draggable, ok := e.TargetNode.Props["draggable"].(bool); ok && draggable {
						dragType, _ := e.TargetNode.Props["dragType"].(string)
						globalDragState.StartDrag(e.Target, dragType, e.TargetNode.Props["dragData"])
						globalDragState.UpdatePosition(e.X, e.Y)
					}
				}

			case "mousemove":
				// Update hover state and synthesize mouseenter/mouseleave
				newHoverID := e.Target
				if newHoverID != app.hoveredID {
					oldHoverID := app.hoveredID
					app.hoveredID = newHoverID

					// Synthesize mouseleave for old element
					if oldHoverID != "" {
						globalEventBus.Emit(&Event{
							Type:      "mouseleave",
							Target:    oldHoverID,
							Bubbles:   false,
							X:         e.X,
							Y:         e.Y,
							Timestamp: e.Timestamp,
						})
					}

					// Synthesize mouseenter for new element
					if newHoverID != "" {
						globalEventBus.Emit(&Event{
							Type:      "mouseenter",
							Target:    newHoverID,
							Bubbles:   false,
							X:         e.X,
							Y:         e.Y,
							Timestamp: e.Timestamp,
						})
					}

					// Mark all components dirty to re-render with new hover state
					globalRegistry.mu.RLock()
					for _, comp := range globalRegistry.components {
						comp.Dirty.Store(true)
					}
					globalRegistry.mu.RUnlock()
				}

				if globalDragState.Dragging() {
					globalDragState.UpdatePosition(e.X, e.Y)
					// Check if hovering over a drop target
					if e.TargetNode != nil {
						if _, hasOnDrop := e.TargetNode.Props["onDrop"]; hasOnDrop {
							globalDragState.SetDropTarget(e.Target)
							globalEventBus.Emit(&Event{
								Type: "dragover", Target: e.Target, Bubbles: true,
								X: e.X, Y: e.Y, LocalX: e.LocalX, LocalY: e.LocalY,
								Timestamp: e.Timestamp,
							})
						}
					}
				}

			case "mouseup":
				if globalDragState.Dragging() {
					sourceID, dropTargetID, _ := globalDragState.EndDrag()
					if dropTargetID != "" {
						globalEventBus.Emit(&Event{
							Type: "drop", Target: dropTargetID, Bubbles: true,
							X: e.X, Y: e.Y,
							Timestamp: e.Timestamp,
						})
					}
					_ = sourceID // available for future use
				}
			}
		}

		// Dispatch to EventBus (handles focus, shortcuts, registered handlers)
		globalEventBus.Emit(e)

		// For mousedown with a target, also emit a "click" event
		if e.Type == "mousedown" && e.Target != "" {
			clickEvent := &Event{
				Type:       "click",
				Target:     e.Target,
				Bubbles:    true,
				X:          e.X,
				Y:          e.Y,
				LocalX:     e.LocalX,
				LocalY:     e.LocalY,
				Button:     e.Button,
				Modifiers:  e.Modifiers,
				Timestamp:  e.Timestamp,
				TargetNode: e.TargetNode,
			}
			globalEventBus.Emit(clickEvent)

			// Right-click → emit contextmenu
			if e.Button == "right" {
				globalEventBus.Emit(&Event{
					Type:      "contextmenu",
					Target:    e.Target,
					Bubbles:   true,
					X:         e.X,
					Y:         e.Y,
					Timestamp: e.Timestamp,
				})
			}
		}

		// Handle text input events first (if focused element is input/textarea)
		textHandled := false
		if e.Type == "keydown" {
			textHandled = app.handleTextInputEvent(e)
		}

		// Handle keyboard navigation (Tab, Enter, Escape, etc.)
		// Skip if text input consumed the event (except Tab which is not consumed)
		if e.Type == "keydown" && !textHandled {
			globalEventBus.HandleKeyEvent(e.Key, e.Modifiers)
		}

		// Dispatch lumina.onKey() bindings
		if e.Type == "keydown" {
			app.dispatchKeyBinding(e.Key)
		}

		// Handle scroll events (mouse wheel and PageUp/PageDown)
		app.handleScrollEvent(e)

		// Re-render deferred to EndBatch()

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
		// Re-render deferred to EndBatch()
	}
}

// dispatchKeyBinding checks if a key has a lumina.onKey() binding and calls it directly.
func (app *App) dispatchKeyBinding(key string) {
	keyBindingsMu.Lock()
	ref, ok := keyBindings[key]
	keyBindingsMu.Unlock()

	if ok {
		app.L.RawGetI(lua.RegistryIndex, int64(ref))
		if app.L.IsFunction(-1) {
			if status := app.L.PCall(0, 0, 0); status != lua.OK {
				msg, _ := app.L.ToString(-1)
				app.L.Pop(1)
				fmt.Fprintf(os.Stderr, "onKey(%q) error: %s\n", key, msg)
			}
		} else {
			app.L.Pop(1)
		}
	}
}

// handleScrollEvent handles scroll-related events (mouse wheel, PageUp/PageDown).
// handleTextInputEvent routes key events to the focused text input/textarea.
// Returns true if the event was consumed by a text input.
func (app *App) handleTextInputEvent(e *Event) bool {
	if e.Type != "keydown" {
		return false
	}

	focusedID := globalEventBus.GetFocused()
	if focusedID == "" {
		return false
	}

	// Check if the focused element has a text input state
	textInputMu.RLock()
	state, ok := textInputRegistry[focusedID]
	textInputMu.RUnlock()
	if !ok {
		return false
	}

	// Handle Enter for single-line input (triggers onSubmit, not consumed as text)
	if !state.MultiLine && (e.Key == "Enter" || e.Key == "\n") {
		// Trigger onSubmit callback if registered
		// The callback is stored as a Lua registry ref in the component
		app.triggerTextInputCallback(focusedID, "onSubmit", state.Text)
		return true
	}

	consumed, changed := HandleTextInputKey(state, e.Key, e.Modifiers)
	if !consumed {
		return false
	}

	if changed {
		// Trigger onChange callback
		app.triggerTextInputCallback(focusedID, "onChange", state.Text)
	}

	return consumed
}

// triggerTextInputCallback calls a Lua callback (onChange/onSubmit) for a text input.
func (app *App) triggerTextInputCallback(id, callbackName, value string) {
	// The callback refs are stored in the text input callback registry
	textCallbackMu.RLock()
	refID, ok := textCallbacks[id+":"+callbackName]
	textCallbackMu.RUnlock()
	if !ok || refID == 0 {
		return
	}

	app.L.RawGetI(lua.RegistryIndex, int64(refID))
	if app.L.Type(-1) == lua.TypeFunction {
		app.L.PushString(value)
		status := app.L.PCall(1, 0, 0)
		if status != lua.OK {
			app.L.Pop(1) // pop error
		}
	} else {
		app.L.Pop(1) // pop non-function
	}
}

// Text input callback registry — stores Lua function refs for onChange/onSubmit.
var (
	textCallbacks  = make(map[string]int) // "id:onChange" -> Lua registry ref
	textCallbackMu sync.RWMutex
)

// RegisterTextCallback registers a Lua callback for a text input event.
func RegisterTextCallback(id, callbackName string, refID int) {
	textCallbackMu.Lock()
	defer textCallbackMu.Unlock()
	textCallbacks[id+":"+callbackName] = refID
}

// ClearTextCallbacks removes all text input callbacks (for testing).
func ClearTextCallbacks() {
	textCallbackMu.Lock()
	defer textCallbackMu.Unlock()
	textCallbacks = make(map[string]int)
}

func (app *App) handleScrollEvent(e *Event) {
	markAllDirty := func() {
		globalRegistry.mu.RLock()
		for _, comp := range globalRegistry.components {
			comp.Dirty.Store(true)
		}
		globalRegistry.mu.RUnlock()
	}

	switch e.Type {
	case "scroll":
		// Mouse wheel scroll — e.Button is "up" or "down", NOT e.Y (which is cursor position)
		scrollAmount := 3 // lines per wheel tick
		if e.Button == "up" {
			scrollAmount = -3
		}

		// Find the scrollable container under the cursor using VNode tree walk
		targetID := ""
		if root := app.findRootVNode(); root != nil {
			targetID, _ = findScrollableVNode(root, e.X, e.Y)
		}

		// Fallback: try focused element's viewport
		if targetID == "" {
			focusedID := globalEventBus.GetFocused()
			viewportMu.RLock()
			_, hasFocusedVP := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if hasFocusedVP {
				targetID = focusedID
			}
		}

		// Last resort: find any viewport that needs scroll
		if targetID == "" {
			viewportMu.RLock()
			for id, vp := range viewportRegistry {
				if vp.NeedsScroll() {
					targetID = id
					break
				}
			}
			viewportMu.RUnlock()
		}

		if targetID != "" {
			ScrollViewport(targetID, scrollAmount)
			markAllDirty()
		}

		// Also emit wheel event for onWheel handlers
		globalEventBus.Emit(&Event{
			Type:      "wheel",
			Target:    e.Target,
			Bubbles:   true,
			X:         e.X,
			Y:         e.Y,
			Button:    e.Button, // "up" or "down"
			Timestamp: e.Timestamp,
		})

	case "keydown":
		focusedID := globalEventBus.GetFocused()
		if focusedID == "" {
			return
		}

		scrolled := false
		switch e.Key {
		case "PageUp":
			viewportMu.RLock()
			vp, ok := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if ok {
				vp.ScrollUp(vp.ViewH)
				scrolled = true
			}
		case "PageDown":
			viewportMu.RLock()
			vp, ok := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if ok {
				vp.ScrollDown(vp.ViewH)
				scrolled = true
			}
		case "Home":
			viewportMu.RLock()
			vp, ok := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if ok {
				vp.ScrollToTop()
				scrolled = true
			}
		case "End":
			viewportMu.RLock()
			vp, ok := viewportRegistry[focusedID]
			viewportMu.RUnlock()
			if ok {
				vp.ScrollToBottom()
				scrolled = true
			}
		}
		if scrolled {
			markAllDirty()
		}
	}
}

// renderAllDirty checks all components for dirty state and re-renders.
// ReloadScript performs a hot reload of the given Lua script.
// It snapshots all component states, clears the registry, re-executes
// the script, then restores states by component type name matching.
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

func (app *App) renderAllDirty() {
	// Frame rate limiting: skip if less than 16ms since last render
	now := time.Now()
	if !app.lastRenderTime.IsZero() && now.Sub(app.lastRenderTime) < 16*time.Millisecond {
		return // will catch on next tick
	}

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
	comp.ResetHookIndex()
	defer SetCurrentComponent(nil)

	// Cache dimensions locally (atomic reads)
	w, h := app.getWidth(), app.getHeight()

	if !comp.PushRenderFn(app.L) {
		return
	}

	status := app.L.PCall(0, 1, 0)
	if status != lua.OK {
		app.L.Pop(1)
		return
	}

	newVNode := LuaVNodeToVNode(app.L, -1)
	app.L.Pop(1)

	// Diff against previous render to detect no-change case.
	var frame *Frame
	sizeChanged := (w != app.lastRenderWidth) || (h != app.lastRenderHeight)
	inspectorDirty := needsInspectorRerender.Load()
	needsInspectorRerender.Store(false)
	if comp.LastVNode != nil {
		patches := DiffVNode(comp.LastVNode, newVNode)
		scrollDirty := AnyViewportScrollDirty()
		if len(patches) == 0 && !scrollDirty && !sizeChanged && !inspectorDirty {
			// Nothing changed — skip rendering.
			comp.MarkClean()
			ClearAllScrollDirty()
			return
		}

		// Incremental rendering: if few patches and we have a previous frame,
		// re-layout the new tree and apply only changed subtrees via parent
		// container re-rendering (handles sibling position shifts correctly).
		// Scroll-dirty or size-change forces full re-render since layout positions change.
		// Also need full re-render if inspector visibility changed.
		if len(patches) <= 10 && app.lastFrame != nil && !ShouldFullRerender(patches, newVNode) && !scrollDirty && !sizeChanged && !inspectorDirty {
			frame = app.lastFrame
			// Re-layout the entire new tree (cheap) so positions are correct
			computeFlexLayout(newVNode, 0, 0, w, h)
			ApplyPatches(frame, newVNode, patches, w, h)
			// Reconcile components: cleanup any removed in incremental update
			ReconcileComponents(app.L, comp.LastVNode, newVNode)
			comp.LastVNode = newVNode
			app.lastFrame = frame
			app.lastRenderWidth = w
			app.lastRenderHeight = h
			goto compositeAndWrite
		}
	}
	// Full re-render (first render, large change, or no previous frame)
	frame = VNodeToFrame(newVNode, w, h)

	// Reconcile components: cleanup any that were in old tree but not in new
	if comp.LastVNode != nil {
		ReconcileComponents(app.L, comp.LastVNode, newVNode)
	}

	comp.LastVNode = newVNode
	app.lastFrame = frame
	app.lastRenderWidth = w
	app.lastRenderHeight = h

compositeAndWrite:
	// Clear scroll dirty flags after re-render (layout applied new scroll positions)
	ClearAllScrollDirty()
	// Bridge VNode event handlers to EventBus
	app.bridgeVNodeEvents(newVNode)

	// Composite overlays on top of the base frame using the layer compositor
	overlays := globalOverlayManager.GetVisible()
	if len(overlays) > 0 {
		compositor := NewCompositor(w, h)
		frame = compositor.Compose(frame, overlays)
	}

	// Composite managed windows on top of overlays
	windows := globalWindowManager.GetVisible()
	if len(windows) > 0 {
		compositor := NewCompositor(w, h)
		var winOverlays []*Overlay
		for _, win := range windows {
			winVNode := BuildWindowVNode(win)
			winOverlays = append(winOverlays, &Overlay{
				ID:      "window-" + win.ID,
				VNode:   winVNode,
				X:       win.X,
				Y:       win.Y,
				W:       win.W,
				H:       win.H,
				ZIndex:  win.ZIndex,
				Visible: true,
			})
		}
		frame = compositor.Compose(frame, winOverlays)
	}

	frame.FocusedID = globalEventBus.GetFocused()

	// DevTools inspector overlay
	if IsInspectorVisible() && app.lastFrame != nil {
		// Highlight the hovered/selected element
		var highlightNode *VNode
		if globalInspector.selectedID != "" {
			highlightNode = findVNodeByID(newVNode, globalInspector.selectedID)
		} else if globalInspector.highlightID != "" {
			highlightNode = findVNodeByID(newVNode, globalInspector.highlightID)
		}
		if highlightNode != nil {
			RenderHighlight(frame, highlightNode)
		}

		// Render inspector panel as overlay on right side
		panelVNode := BuildInspectorVNode(newVNode, w, h)
		if panelVNode != nil {
			panelW := globalInspector.panelWidth
			if panelW > w/2 {
				panelW = w / 2
			}
			panelOverlay := &Overlay{
				ID:      "devtools-panel",
				VNode:   panelVNode,
				X:       w - panelW,
				Y:       0,
				W:       panelW,
				H:       h,
				ZIndex:  9999,
				Visible: true,
			}
			dtCompositor := NewCompositor(w, h)
			frame = dtCompositor.Compose(frame, []*Overlay{panelOverlay})
		}
	}

	adapter.Write(frame)
	app.lastRenderTime = time.Now()
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

		frame := RenderLuaVNode(app.L, -1, app.getWidth(), app.getHeight())
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

// RenderOnce renders all dirty components once and returns.
// Useful for headless testing — render a single frame without an event loop.
func (app *App) RenderOnce() {
	app.renderAllDirty()
}

// Close closes the application and cleans up resources.
func (app *App) Close() {
	if app.L != nil {
		app.L.Close()
	}
}

// GetWindowManager returns the app's window manager.
func (app *App) GetWindowManager() *WindowManager {
	return globalWindowManager
}

// GetDevTools returns the app's DevTools instance.
func (app *App) GetDevTools() *DevTools {
	return globalDevTools
}

// Scheduler returns the App's async coroutine scheduler.
func (app *App) Scheduler() *lua.Scheduler {
	return app.sched
}

// HitTestVNode finds the deepest VNode containing point (px, py).
// Returns the VNode's ID (from props["id"]) or "" if no match.
func HitTestVNode(vnode *VNode, px, py int) string {
	if vnode == nil {
		return ""
	}
	// Check if point is within this node's bounds
	if px < vnode.X || px >= vnode.X+vnode.W || py < vnode.Y || py >= vnode.Y+vnode.H {
		return ""
	}
	// Check children (deepest match wins — reverse order for z-order)
	for i := len(vnode.Children) - 1; i >= 0; i-- {
		if id := HitTestVNode(vnode.Children[i], px, py); id != "" {
			return id
		}
	}
	// Return this node's ID if it has one
	if id, ok := vnode.Props["id"].(string); ok && id != "" {
		return id
	}
	return ""
}

// findRootVNode returns the last rendered VNode tree from the root component.
func (app *App) findRootVNode() *VNode {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	for _, comp := range globalRegistry.components {
		if comp.LastVNode != nil {
			return comp.LastVNode
		}
	}
	return nil
}

// GetGlobalEventBus returns the global event bus (for testing).
func GetGlobalEventBus() *EventBus {
	return globalEventBus
}

// ProcessPendingEvents drains and processes all pending events in the channel.
// Used in tests to process lua_callback events posted by event handlers.
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
		Type:    "click",
		Target:  id,
		Bubbles: true,
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

// clearScreenWithBg returns an ANSI escape sequence that sets the theme
// background color before clearing the screen. This ensures the cleared
// area uses our theme color instead of the terminal's default background.
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
