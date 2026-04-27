//go:build linux || darwin

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/akzj/go-lua/pkg/lua"
	v2 "github.com/akzj/lumina/pkg/lumina/v2"
	"github.com/akzj/lumina/pkg/lumina/v2/output"
	"github.com/akzj/lumina/pkg/lumina/v2/terminal"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: lumina-v2 <script.lua>")
		os.Exit(1)
	}
	scriptPath := os.Args[1]

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
