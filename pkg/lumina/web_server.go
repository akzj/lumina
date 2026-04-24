package lumina

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sync"
)

//go:embed web/index.html web/lumina-client.js
var webAssets embed.FS

// WebServer serves a Lumina app over HTTP + WebSocket.
// Each WebSocket connection gets its own App instance running the same script.
type WebServer struct {
	addr       string
	scriptPath string
	mux        *http.ServeMux
	listener   net.Listener
	mu         sync.Mutex
	sessions   map[*WebTerminal]bool
}

// NewWebServer creates a web server that serves the given Lua script.
func NewWebServer(addr, scriptPath string) *WebServer {
	ws := &WebServer{
		addr:       addr,
		scriptPath: scriptPath,
		mux:        http.NewServeMux(),
		sessions:   make(map[*WebTerminal]bool),
	}
	ws.setupRoutes()
	return ws
}

// setupRoutes configures HTTP routes.
func (ws *WebServer) setupRoutes() {
	// Serve embedded web assets
	webFS, _ := fs.Sub(webAssets, "web")
	ws.mux.Handle("/", http.FileServer(http.FS(webFS)))

	// WebSocket endpoint
	ws.mux.HandleFunc("/ws", ws.handleWebSocket)

	// Health check
	ws.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ws.mu.Lock()
		count := len(ws.sessions)
		ws.mu.Unlock()
		fmt.Fprintf(w, `{"status":"ok","sessions":%d}`, count)
	})
}

// Start starts the HTTP server. Blocks until the server is stopped.
func (ws *WebServer) Start() error {
	ln, err := net.Listen("tcp", ws.addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", ws.addr, err)
	}
	ws.mu.Lock()
	ws.listener = ln
	ws.mu.Unlock()

	log("Lumina web server listening on http://%s", ln.Addr())

	server := &http.Server{Handler: ws.mux}
	return server.Serve(ln)
}

// StartBackground starts the server in a goroutine and returns the address.
func (ws *WebServer) StartBackground() (string, error) {
	ln, err := net.Listen("tcp", ws.addr)
	if err != nil {
		return "", fmt.Errorf("listen %s: %w", ws.addr, err)
	}
	ws.mu.Lock()
	ws.listener = ln
	ws.mu.Unlock()

	server := &http.Server{Handler: ws.mux}
	go server.Serve(ln)

	return ln.Addr().String(), nil
}

// Stop stops the web server.
func (ws *WebServer) Stop() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.listener != nil {
		return ws.listener.Close()
	}
	return nil
}

// Addr returns the listener address (useful after StartBackground).
func (ws *WebServer) Addr() string {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.listener != nil {
		return ws.listener.Addr().String()
	}
	return ws.addr
}

// SessionCount returns the number of active WebSocket sessions.
func (ws *WebServer) SessionCount() int {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return len(ws.sessions)
}

// handleWebSocket upgrades HTTP to WebSocket and runs a Lumina session.
func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn := UpgradeWebSocket(w, r)
	if conn == nil {
		return
	}

	// Create WebTerminal for this connection
	wt := NewWebTerminal(conn)

	// Track session
	ws.mu.Lock()
	ws.sessions[wt] = true
	ws.mu.Unlock()

	defer func() {
		ws.mu.Lock()
		delete(ws.sessions, wt)
		ws.mu.Unlock()
		wt.Close()
	}()

	// Create a new App for this session
	app := NewApp()
	defer app.Close()

	// Run the app with the WebTerminal
	if err := app.RunWithTermIO(wt, ws.scriptPath); err != nil {
		log("session error: %v", err)
	}
}
