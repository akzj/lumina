package v2

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/akzj/go-lua/pkg/lua"
	"github.com/akzj/lumina/pkg/animation"
	"github.com/akzj/lumina/pkg/event"
	"github.com/akzj/lumina/pkg/hotreload"
	"github.com/akzj/lumina/pkg/router"
)

// InputEvent represents a terminal input event. This is the interface that
// the terminal package will provide. Defined here so the event loop can be
// compiled and tested before the terminal package exists.
type InputEvent struct {
	Type      string // "keydown", "mousedown", "mouseup", "mousemove", "scroll", "resize"
	Key       string // key name (e.g. "a", "Enter", "Tab", "Escape")
	Char      string // printable character (if any)
	X, Y      int    // mouse position (screen coordinates)
	Button    string // mouse button or scroll direction ("left", "right", "up", "down")
	Modifiers InputModifiers
}

// InputModifiers represents keyboard modifiers.
type InputModifiers struct {
	Ctrl  bool
	Alt   bool
	Shift bool
}

// RunConfig configures the App runtime.
type RunConfig struct {
	// ScriptPath is the Lua script to load and execute.
	ScriptPath string

	// Events is the channel of input events from the terminal.
	// If nil, no input events are processed (useful for headless/test mode).
	Events <-chan InputEvent

	// FrameRate is the target frame rate in Hz. Default: 60.
	FrameRate int

	// Watch enables hot reload: when the ScriptPath .lua file changes on
	// disk, the app re-executes it and restores component state.
	Watch bool
}

// Run starts the application event loop. It loads the Lua script (if configured),
// performs an initial full render, then enters the event loop. Returns when
// Stop() is called or the event channel is closed.
func (a *App) Run(cfg RunConfig) error {
	if a.luaState == nil {
		// Non-Lua mode: just run the event loop without script loading.
		return a.eventLoop(cfg)
	}

	// Tune Lua GC for UI workloads: less frequent collections, smaller steps.
	a.luaState.SetGCParam("pause", 200)
	a.luaState.SetGCParam("stepmul", 100)

	// Load and execute the Lua script.
	if cfg.ScriptPath != "" {
		addScriptDirToPackagePath(a.luaState, cfg.ScriptPath)
		if err := a.luaState.DoFile(cfg.ScriptPath); err != nil {
			return err
		}
	}

	// Initial full render.
	a.RenderAll()

	// Enter the event loop.
	return a.eventLoop(cfg)
}

// RunScript loads and executes a Lua script without starting the event loop.
// Useful for testing: load the script, then manually call RenderAll/RenderDirty.
func (a *App) RunScript(path string) error {
	if a.luaState == nil {
		return nil
	}
	addScriptDirToPackagePath(a.luaState, path)
	return a.luaState.DoFile(path)
}

// RunString executes a Lua code string without starting the event loop.
// Useful for testing: execute inline Lua, then manually call RenderAll/RenderDirty.
func (a *App) RunString(code string) error {
	if a.luaState == nil {
		return nil
	}
	return a.luaState.DoString(code)
}

// Stop signals the event loop to exit.
func (a *App) Stop() {
	// Cancel pending async coroutines.
	if a.scheduler != nil {
		a.scheduler.Destroy()
	}
	if a.quit != nil {
		select {
		case <-a.quit:
			// Already closed.
		default:
			close(a.quit)
		}
	}
}

// IsRunning returns true if the event loop is active.
func (a *App) IsRunning() bool {
	return a.running
}

// AnimationManager returns the animation manager (for testing).
func (a *App) AnimationManager() *animation.Manager {
	return a.animMgr
}

// RouterManager returns the router (for testing).
func (a *App) RouterManager() *router.Router {
	return a.routerMgr
}

// eventLoop is the core event loop. It processes input events, ticks
// animations, handles hot reload, and renders dirty components at the
// target frame rate.
func (a *App) eventLoop(cfg RunConfig) error {
	frameRate := cfg.FrameRate
	if frameRate <= 0 {
		frameRate = 60
	}
	frameDuration := time.Second / time.Duration(frameRate)

	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	a.running = true
	defer func() { a.running = false }()

	events := cfg.Events

	// Set up hot reload watcher if enabled.
	var reloadCh chan string
	if cfg.Watch && cfg.ScriptPath != "" {
		reloadCh = make(chan string, 1)
		watcher := hotreload.NewWatcher([]string{cfg.ScriptPath}, 500*time.Millisecond)
		watcher.SetOnChange(func(path string) {
			select {
			case reloadCh <- path:
			default: // drop if already queued
			}
		})
		watcher.Start()
		defer watcher.Stop()
	}
	// Provide a nil-safe channel that never receives when watch is disabled.
	if reloadCh == nil {
		reloadCh = make(chan string)
	}

	for {
		select {
		case <-a.quit:
			return nil

		case ie, ok := <-events:
			if !ok {
				return nil // input channel closed
			}
			a.handleInputEvent(ie)

		case path := <-reloadCh:
			a.reloadScript(path)

		case <-ticker.C:
			// Tick animations.
			if a.animMgr != nil && a.animMgr.IsRunning() {
				nowMs := time.Now().UnixMilli()
				completed := a.animMgr.Tick(nowMs)
				_ = completed
			}

			// Fire due timers (setInterval/setTimeout callbacks).
			a.fireTimers()

			// Tick async scheduler — resume completed coroutines.
			if a.scheduler != nil {
				a.scheduler.Tick()
			}

			// Tick FPS counter and auto-refresh devtools.
			a.tickDevTools()

			// Render dirty components.
			a.RenderDirty()
		}
	}
}

// reloadScript performs a hot reload: re-executes the Lua script and re-renders.
func (a *App) reloadScript(path string) {
	// Reset timers.
	if a.timerMgr != nil {
		a.timerMgr.releaseAll(a.luaState)
	}

	// Cancel pending async coroutines.
	if a.scheduler != nil {
		a.scheduler.Destroy()
	}

	// Re-set package.path for the reloaded script.
	addScriptDirToPackagePath(a.luaState, path)
	if err := a.luaState.DoFile(path); err != nil {
		log.Printf("[hotreload] error reloading %s: %v", path, err)
		return
	}

	// Full re-render.
	a.RenderAll()

	log.Printf("[hotreload] reloaded %s", path)
}

// handleInputEvent converts a terminal InputEvent to an event.Event and
// dispatches it through the event system.
func (a *App) handleInputEvent(ie InputEvent) {
	switch ie.Type {
	case "keydown", "keyup":
		key := ie.Key
		// Build modifier-prefixed key name.
		if ie.Modifiers.Shift && key == "Tab" {
			key = "Shift+Tab"
		}

		// Check for quit keys (Ctrl+C, Ctrl+Q).
		if ie.Type == "keydown" && ie.Modifiers.Ctrl {
			if key == "c" || key == "q" {
				a.Stop()
				return
			}
		}

		a.HandleEvent(&event.Event{
			Type: ie.Type,
			Key:  key,
		})

	case "mousedown", "mouseup", "mousemove":
		a.HandleEvent(&event.Event{
			Type: ie.Type,
			X:    ie.X,
			Y:    ie.Y,
		})

	case "scroll":
		a.HandleEvent(&event.Event{
			Type: "scroll",
			X:    ie.X,
			Y:    ie.Y,
			Key:  ie.Button, // "up" or "down"
		})

	case "resize":
		a.Resize(ie.X, ie.Y)
		a.RenderAll()
	}
}

// addScriptDirToPackagePath prepends the script's directory to Lua's package.path
// so that require("lib.components") resolves to <script_dir>/lib/components.lua.
func addScriptDirToPackagePath(L *lua.State, scriptPath string) {
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return
	}
	dir := filepath.Dir(absPath)
	// Escape backslashes for Windows paths in Lua string.
	escaped := strings.ReplaceAll(dir, `\`, `\\`)
	code := fmt.Sprintf(`package.path = "%s/?.lua;%s/?/init.lua;" .. package.path`, escaped, escaped)
	_ = L.DoString(code)
}
