//go:build linux || darwin

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/akzj/go-lua/pkg/lua"
	v2 "github.com/akzj/lumina/pkg/lumina/v2"
	"github.com/akzj/lumina/pkg/lumina/v2/mcp"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/terminal"
)

func main() {
	// Parse simple flags: --web :8080, --watch, --mcp :8088
	var webAddr string
	var mcpAddr string
	var watchMode bool
	var scriptPath string
	var args []string

	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--web" && i+1 < len(os.Args) {
			webAddr = os.Args[i+1]
			i++ // skip value
		} else if strings.HasPrefix(os.Args[i], "--web=") {
			webAddr = strings.TrimPrefix(os.Args[i], "--web=")
		} else if os.Args[i] == "--mcp" && i+1 < len(os.Args) {
			mcpAddr = os.Args[i+1]
			i++ // skip value
		} else if strings.HasPrefix(os.Args[i], "--mcp=") {
			mcpAddr = strings.TrimPrefix(os.Args[i], "--mcp=")
		} else if os.Args[i] == "--watch" {
			watchMode = true
		} else {
			args = append(args, os.Args[i])
		}
	}

	if len(args) < 1 {
		fmt.Println("Usage: lumina-v2 [--web :8080] [--mcp :8088] [--watch] <script.lua>")
		os.Exit(1)
	}
	scriptPath = args[0]

	if webAddr != "" {
		runWeb(webAddr, scriptPath, watchMode, mcpAddr)
	} else {
		runTerminal(scriptPath, watchMode, mcpAddr)
	}
}

// runWeb starts the WebSocket server mode — renders to browser instead of terminal.
func runWeb(addr string, scriptPath string, watch bool, mcpAddr string) {
	const defaultW, defaultH = 80, 24

	// 1. Create Lua state.
	L := lua.NewState()
	defer L.Close()

	addScriptDirToPackagePath(L, scriptPath)

	// 2. Create WebSocket adapter.
	wsEvents := make(chan output.WSEvent, 64)
	wsAdapter := output.NewWSAdapter(addr, defaultW, defaultH, wsEvents)

	if err := wsAdapter.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "websocket server: %v\n", err)
		os.Exit(1)
	}
	defer wsAdapter.Close()

	listenAddr := wsAdapter.Addr()
	fmt.Fprintf(os.Stderr, "Lumina v2 web mode — open http://%s in your browser\n", listenAddr)

	// 3. Create App.
	app := v2.NewAppWithLua(L, defaultW, defaultH, wsAdapter)

	// 3b. Start MCP HTTP server if requested.
	if mcpAddr != "" {
		mcpHandler := mcp.NewHandler(app)
		go serveMCP(mcpHandler, mcpAddr)
	}

	// 4. Bridge WSEvent → v2.InputEvent.
	appEvents := make(chan v2.InputEvent, 64)
	go func() {
		for we := range wsEvents {
			appEvents <- v2.InputEvent{
				Type:   we.Type,
				Key:    we.Key,
				Char:   we.Char,
				X:      we.X,
				Y:      we.Y,
				Button: we.Button,
				Modifiers: v2.InputModifiers{
					Ctrl:  we.Ctrl,
					Alt:   we.Alt,
					Shift: we.Shift,
				},
			}
		}
		close(appEvents)
	}()

	// 5. Run the app event loop.
	if err := app.Run(v2.RunConfig{
		ScriptPath: scriptPath,
		Events:     appEvents,
		Watch:      watch,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runTerminal runs in standard TUI terminal mode (existing behavior).
func runTerminal(scriptPath string, watch bool, mcpAddr string) {
	// 1. Create terminal.
	term, err := terminal.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "terminal init: %v\n", err)
		os.Exit(1)
	}

	// 2. Enable raw mode (alt screen, mouse tracking, hide cursor).
	if err := term.EnableRawMode(); err != nil {
		fmt.Fprintf(os.Stderr, "raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.RestoreMode()

	// 3. Get terminal size.
	w, h := term.Size()

	// 4. Create Lua state (standard libraries are loaded by NewState).
	L := lua.NewState()
	defer L.Close()

	// 5. Add script directory to package.path so require() finds siblings.
	addScriptDirToPackagePath(L, scriptPath)

	// 6. Create TUI adapter writing to the terminal output.
	adapter := output.NewTUIAdapter(term.Output())

	// 7. Create App with Lua wiring.
	app := v2.NewAppWithLua(L, w, h, adapter)

	// 7b. Start MCP HTTP server if requested.
	if mcpAddr != "" {
		mcpHandler := mcp.NewHandler(app)
		go serveMCP(mcpHandler, mcpAddr)
	}

	// 8. Start input reader — reads raw bytes from stdin, parses into events.
	termEvents := make(chan terminal.InputEvent, 64)
	reader := terminal.NewInputReader(termEvents)
	reader.Start()
	defer reader.Stop()

	// 9. Bridge terminal events → app events via a converter goroutine.
	appEvents := make(chan v2.InputEvent, 64)
	go func() {
		for te := range termEvents {
			appEvents <- v2.InputEvent{
				Type:   te.Type,
				Key:    te.Key,
				Char:   te.Char,
				X:      te.X,
				Y:      te.Y,
				Button: te.Button,
				Modifiers: v2.InputModifiers{
					Ctrl:  te.Modifiers.Ctrl,
					Alt:   te.Modifiers.Alt,
					Shift: te.Modifiers.Shift,
				},
			}
		}
		close(appEvents)
	}()

	// 10. Watch for terminal resize (SIGWINCH).
	term.WatchResize(func(newW, newH int) {
		appEvents <- v2.InputEvent{
			Type: "resize",
			X:    newW,
			Y:    newH,
		}
	})
	defer term.StopResize()

	// 11. Run the app event loop with the Lua script.
	if err := app.Run(v2.RunConfig{
		ScriptPath: scriptPath,
		Events:     appEvents,
		Watch:      watch,
	}); err != nil {
		// Terminal is restored by deferred RestoreMode before os.Exit.
		term.RestoreMode()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// addScriptDirToPackagePath prepends the script's directory to Lua's
// package.path so that require("foo") finds foo.lua next to the script.
func addScriptDirToPackagePath(L *lua.State, scriptPath string) {
	dir := filepath.Dir(scriptPath)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}
	// Prepend: <dir>/?.lua;<dir>/?/init.lua; to existing package.path
	code := fmt.Sprintf(`package.path = %q .. ";" .. package.path`,
		absDir+"/?.lua;"+absDir+"/?/init.lua")
	_ = L.DoString(code)
}
