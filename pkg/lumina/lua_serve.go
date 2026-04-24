package lumina

import (
	"fmt"
	"os"

	"github.com/akzj/go-lua/pkg/lua"
)

// luaServe implements lumina.serve(port) — starts the web server.
// Usage from Lua:
//
//	lumina.serve(8080)                    -- serve on port 8080
//	lumina.serve({ port = 8080 })         -- serve with options table
//	lumina.serve({ port = 8080, script = "app.lua" })
//
// This function blocks (runs the HTTP server).
func luaServe(L *lua.State) int {
	var port int
	var scriptPath string

	switch L.Type(1) {
	case lua.TypeNumber:
		p, _ := L.ToInteger(1)
		port = int(p)
	case lua.TypeTable:
		L.GetField(1, "port")
		if p, ok := L.ToInteger(-1); ok {
			port = int(p)
		}
		L.Pop(1)

		L.GetField(1, "script")
		if s, ok := L.ToString(-1); ok {
			scriptPath = s
		}
		L.Pop(1)
	default:
		port = 8080
	}

	if port <= 0 || port > 65535 {
		port = 8080
	}

	// If no script specified, try to find the currently running script.
	// The script path is typically set by the caller; if not, we serve
	// without a script (empty app — useful for testing).
	if scriptPath == "" {
		// Check if there's a script path stored on the Lua state
		L.GetGlobal("_LUMINA_SCRIPT")
		if s, ok := L.ToString(-1); ok && s != "" {
			scriptPath = s
		}
		L.Pop(1)
	}

	// Validate script exists (if specified)
	if scriptPath != "" {
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			L.PushString(fmt.Sprintf("lumina.serve: script not found: %s", scriptPath))
			L.Error()
			return 0
		}
	}

	addr := fmt.Sprintf(":%d", port)
	ws := NewWebServer(addr, scriptPath)

	// Print startup message
	fmt.Fprintf(os.Stderr, "🌐 Lumina web server starting on http://localhost:%d\n", port)

	// Start blocks until server stops
	if err := ws.Start(); err != nil {
		L.PushString(fmt.Sprintf("lumina.serve: %v", err))
		L.Error()
		return 0
	}

	return 0
}

// luaServeBackground implements lumina.serveBackground(port) for testing.
// Returns the address string.
func luaServeBackground(L *lua.State) int {
	var port int
	var scriptPath string

	switch L.Type(1) {
	case lua.TypeNumber:
		p, _ := L.ToInteger(1)
		port = int(p)
	case lua.TypeTable:
		L.GetField(1, "port")
		if p, ok := L.ToInteger(-1); ok {
			port = int(p)
		}
		L.Pop(1)

		L.GetField(1, "script")
		if s, ok := L.ToString(-1); ok {
			scriptPath = s
		}
		L.Pop(1)
	default:
		port = 0
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	ws := NewWebServer(addr, scriptPath)

	actualAddr, err := ws.StartBackground()
	if err != nil {
		L.PushString(fmt.Sprintf("lumina.serveBackground: %v", err))
		L.Error()
		return 0
	}

	L.PushString(actualAddr)

	// Store the server reference so it can be stopped later
	L.SetUserValue("lumina_web_server", ws)

	return 1
}
