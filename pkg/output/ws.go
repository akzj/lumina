package output

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/akzj/lumina/pkg/buffer"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// WSMessage is the WebSocket protocol message envelope.
type WSMessage struct {
	Type string          `json:"type"`           // "full", "dirty", "init"
	Data json.RawMessage `json:"data,omitempty"` // payload varies by type
}

// WSInputMessage is a browser→server input event.
type WSInputMessage struct {
	Type string      `json:"type"` // "keydown", "mousedown", "mouseup", "mousemove", "scroll", "resize"
	Data WSInputData `json:"data"`
}

// WSInputData carries the input event fields.
type WSInputData struct {
	Key    string `json:"key,omitempty"`
	Char   string `json:"char,omitempty"`
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Button string `json:"button,omitempty"`
	Ctrl   bool   `json:"ctrl,omitempty"`
	Alt    bool   `json:"alt,omitempty"`
	Shift  bool   `json:"shift,omitempty"`
}

// WSEvent is the event type produced by the WebSocket adapter,
// matching the v2.InputEvent structure so the caller can convert directly.
type WSEvent struct {
	Type   string
	Key    string
	Char   string
	X, Y   int
	Button string
	Ctrl   bool
	Alt    bool
	Shift  bool
}

// WSAdapter implements Adapter over WebSocket connections.
// It broadcasts screen updates to all connected browser clients
// and receives input events from them.
type WSAdapter struct {
	mu      sync.Mutex
	conns   map[*websocket.Conn]context.CancelFunc
	eventCh chan<- WSEvent
	server  *http.Server
	addr    string
	width   int
	height  int

	// listener is stored so we can get the actual bound address (useful for :0)
	listener net.Listener
}

// NewWSAdapter creates a WebSocket adapter that listens on addr (e.g. ":8080").
// Input events from browser clients are sent to eventCh.
// width and height are the initial screen dimensions sent to clients on connect.
func NewWSAdapter(addr string, width, height int, eventCh chan<- WSEvent) *WSAdapter {
	ws := &WSAdapter{
		conns:   make(map[*websocket.Conn]context.CancelFunc),
		eventCh: eventCh,
		addr:    addr,
		width:   width,
		height:  height,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", ws.handleWS)
	mux.HandleFunc("/", ws.handleIndex)

	ws.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return ws
}

// Start begins listening for HTTP/WebSocket connections.
// It returns immediately; the server runs in a background goroutine.
func (ws *WSAdapter) Start() error {
	ln, err := net.Listen("tcp", ws.addr)
	if err != nil {
		return fmt.Errorf("ws listen: %w", err)
	}
	ws.listener = ln
	go func() {
		if err := ws.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("ws server: %v", err)
		}
	}()
	return nil
}

// Addr returns the listener address (useful when binding to :0).
func (ws *WSAdapter) Addr() net.Addr {
	if ws.listener != nil {
		return ws.listener.Addr()
	}
	return nil
}

// WriteFull sends the entire screen buffer to all connected clients.
func (ws *WSAdapter) WriteFull(screen *buffer.Buffer) error {
	result := bufferToRenderResult(screen)
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	msg := WSMessage{Type: "full", Data: data}
	ws.broadcast(msg)
	return nil
}

// WriteDirty sends only the changed regions to all connected clients.
func (ws *WSAdapter) WriteDirty(screen *buffer.Buffer, dirtyRects []buffer.Rect) error {
	result := bufferToRenderResult(screen)
	result.DirtyRects = make([]RectJSON, len(dirtyRects))
	for i, r := range dirtyRects {
		result.DirtyRects[i] = RectJSON{X: r.X, Y: r.Y, W: r.W, H: r.H}
	}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	msg := WSMessage{Type: "dirty", Data: data}
	ws.broadcast(msg)
	return nil
}

// Flush is a no-op for WebSocket adapter (messages are sent immediately).
func (ws *WSAdapter) Flush() error { return nil }

// Close shuts down the HTTP server and closes all WebSocket connections.
func (ws *WSAdapter) Close() error {
	ws.mu.Lock()
	for conn, cancel := range ws.conns {
		cancel()
		conn.Close(websocket.StatusNormalClosure, "server closing")
	}
	ws.conns = make(map[*websocket.Conn]context.CancelFunc)
	ws.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return ws.server.Shutdown(ctx)
}

// handleWS handles WebSocket upgrade requests.
func (ws *WSAdapter) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow any origin for development
	})
	if err != nil {
		log.Printf("ws accept: %v", err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())

	ws.mu.Lock()
	ws.conns[conn] = cancel
	ws.mu.Unlock()

	// Send initial screen dimensions so the client knows the grid size.
	initData, _ := json.Marshal(map[string]int{
		"width":  ws.width,
		"height": ws.height,
	})
	_ = wsjson.Write(ctx, conn, WSMessage{Type: "init", Data: initData})

	// Read input events from this client.
	ws.readLoop(ctx, conn)
}

// readLoop reads input messages from a single WebSocket client.
func (ws *WSAdapter) readLoop(ctx context.Context, conn *websocket.Conn) {
	defer func() {
		ws.mu.Lock()
		if cancel, ok := ws.conns[conn]; ok {
			cancel()
			delete(ws.conns, conn)
		}
		ws.mu.Unlock()
		conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		var msg WSInputMessage
		err := wsjson.Read(ctx, conn, &msg)
		if err != nil {
			return // client disconnected or context cancelled
		}

		evt := WSEvent{
			Type:   msg.Type,
			Key:    msg.Data.Key,
			Char:   msg.Data.Char,
			X:      msg.Data.X,
			Y:      msg.Data.Y,
			Button: msg.Data.Button,
			Ctrl:   msg.Data.Ctrl,
			Alt:    msg.Data.Alt,
			Shift:  msg.Data.Shift,
		}

		select {
		case ws.eventCh <- evt:
		default:
			// drop if channel full — don't block the read loop
		}
	}
}

// broadcast sends a message to all connected clients.
func (ws *WSAdapter) broadcast(msg WSMessage) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for conn, cancel := range ws.conns {
		ctx, writeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := wsjson.Write(ctx, conn, msg)
		writeCancel()
		if err != nil {
			// Remove dead connection.
			cancel()
			delete(ws.conns, conn)
			conn.Close(websocket.StatusGoingAway, "write error")
		}
	}
}

// handleIndex serves the embedded HTML frontend.
func (ws *WSAdapter) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

// SetSize updates the screen dimensions (called on resize).
func (ws *WSAdapter) SetSize(width, height int) {
	ws.mu.Lock()
	ws.width = width
	ws.height = height
	ws.mu.Unlock()
}

// indexHTML is the self-contained browser frontend.
const indexHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Lumina v2</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { background: #1E1E2E; overflow: hidden; }
        canvas { display: block; margin: 0 auto; }
        #status {
            position: fixed; top: 5px; right: 10px;
            color: #6C7086; font: 12px monospace; z-index: 10;
        }
    </style>
</head>
<body>
    <canvas id="screen" tabindex="0"></canvas>
    <div id="status">Connecting...</div>
<script>
(function() {
    const CELL_W = 10;
    const CELL_H = 20;
    const FONT_SIZE = 16;
    const FONT = FONT_SIZE + 'px monospace';

    const canvas = document.getElementById('screen');
    const ctx = canvas.getContext('2d');
    const statusEl = document.getElementById('status');

    let ws = null;
    let screenData = null;
    let gridW = 80, gridH = 24;

    function connect() {
        ws = new WebSocket('ws://' + location.host + '/ws');

        ws.onopen = function() {
            statusEl.textContent = 'Connected';
            canvas.focus();
        };

        ws.onclose = function() {
            statusEl.textContent = 'Disconnected — reconnecting...';
            setTimeout(connect, 1000);
        };

        ws.onerror = function() {
            statusEl.textContent = 'Connection error';
        };

        ws.onmessage = function(e) {
            var msg = JSON.parse(e.data);

            if (msg.type === 'init') {
                gridW = msg.data.width;
                gridH = msg.data.height;
                canvas.width = gridW * CELL_W;
                canvas.height = gridH * CELL_H;
                return;
            }

            if (msg.type === 'full' || msg.type === 'dirty') {
                var data = msg.data;
                screenData = data;

                // Resize canvas if dimensions changed.
                if (canvas.width !== data.width * CELL_W || canvas.height !== data.height * CELL_H) {
                    canvas.width = data.width * CELL_W;
                    canvas.height = data.height * CELL_H;
                }

                if (msg.type === 'dirty' && data.dirty_rects && data.dirty_rects.length > 0) {
                    for (var i = 0; i < data.dirty_rects.length; i++) {
                        var r = data.dirty_rects[i];
                        renderRect(data, r.x, r.y, r.w, r.h);
                    }
                } else {
                    renderRect(data, 0, 0, data.width, data.height);
                }
            }
        };
    }

    function renderRect(data, rx, ry, rw, rh) {
        ctx.textBaseline = 'top';

        for (var y = ry; y < ry + rh && y < data.height; y++) {
            for (var x = rx; x < rx + rw && x < data.width; x++) {
                var cell = data.cells[y][x];
                var px = x * CELL_W;
                var py = y * CELL_H;

                // Background
                ctx.fillStyle = cell.bg || '#1E1E2E';
                ctx.fillRect(px, py, CELL_W, CELL_H);

                // Character
                if (cell.char && cell.char !== ' ') {
                    ctx.fillStyle = cell.fg || '#CDD6F4';
                    ctx.font = cell.bold ? ('bold ' + FONT) : FONT;
                    ctx.fillText(cell.char, px + 1, py + 2);
                }
            }
        }
    }

    // Map browser key names to Lumina terminal key names.
    function mapKey(e) {
        var key = e.key;

        // Special keys map directly.
        var specials = {
            'ArrowUp': 'ArrowUp', 'ArrowDown': 'ArrowDown',
            'ArrowLeft': 'ArrowLeft', 'ArrowRight': 'ArrowRight',
            'Enter': 'Enter', 'Backspace': 'Backspace', 'Delete': 'Delete',
            'Tab': 'Tab', 'Escape': 'Escape', 'Home': 'Home', 'End': 'End',
            'PageUp': 'PageUp', 'PageDown': 'PageDown',
            'F1': 'F1', 'F2': 'F2', 'F3': 'F3', 'F4': 'F4',
            'F5': 'F5', 'F6': 'F6', 'F7': 'F7', 'F8': 'F8',
            'F9': 'F9', 'F10': 'F10', 'F11': 'F11', 'F12': 'F12',
            ' ': 'Space'
        };

        if (specials[key]) return specials[key];

        // Single character keys pass through.
        if (key.length === 1) return key;

        return key;
    }

    // Keyboard events
    canvas.addEventListener('keydown', function(e) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        e.preventDefault();

        var key = mapKey(e);
        ws.send(JSON.stringify({
            type: 'keydown',
            data: {
                key: key,
                ctrl: e.ctrlKey,
                alt: e.altKey,
                shift: e.shiftKey
            }
        }));
    });

    canvas.addEventListener('keyup', function(e) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        e.preventDefault();

        var key = mapKey(e);
        ws.send(JSON.stringify({
            type: 'keyup',
            data: {
                key: key,
                ctrl: e.ctrlKey,
                alt: e.altKey,
                shift: e.shiftKey
            }
        }));
    });

    // Mouse events
    function cellPos(e) {
        var rect = canvas.getBoundingClientRect();
        return {
            x: Math.floor((e.clientX - rect.left) / CELL_W),
            y: Math.floor((e.clientY - rect.top) / CELL_H)
        };
    }

    var buttonMap = ['left', 'middle', 'right'];

    canvas.addEventListener('mousedown', function(e) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        var pos = cellPos(e);
        ws.send(JSON.stringify({
            type: 'mousedown',
            data: { x: pos.x, y: pos.y, button: buttonMap[e.button] || 'left' }
        }));
    });

    canvas.addEventListener('mouseup', function(e) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        var pos = cellPos(e);
        ws.send(JSON.stringify({
            type: 'mouseup',
            data: { x: pos.x, y: pos.y, button: buttonMap[e.button] || 'left' }
        }));
    });

    var lastMoveX = -1, lastMoveY = -1;
    canvas.addEventListener('mousemove', function(e) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        var pos = cellPos(e);
        // Throttle: only send if cell position changed.
        if (pos.x === lastMoveX && pos.y === lastMoveY) return;
        lastMoveX = pos.x;
        lastMoveY = pos.y;
        ws.send(JSON.stringify({
            type: 'mousemove',
            data: { x: pos.x, y: pos.y }
        }));
    });

    canvas.addEventListener('wheel', function(e) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        e.preventDefault();
        var pos = cellPos(e);
        var dir = e.deltaY < 0 ? 'up' : 'down';
        ws.send(JSON.stringify({
            type: 'scroll',
            data: { x: pos.x, y: pos.y, button: dir }
        }));
    }, { passive: false });

    // Prevent context menu on canvas.
    canvas.addEventListener('contextmenu', function(e) { e.preventDefault(); });

    // Auto-focus canvas.
    canvas.focus();
    connect();
})();
</script>
</body>
</html>`
