package lumina

import (
	"encoding/json"
	"io"
	"sync"
)

// WebTerminal implements TermIO over a WebSocket connection.
// Terminal output (escape sequences) is sent as binary WebSocket messages.
// Terminal input (keystrokes) is received from WebSocket messages.
type WebTerminal struct {
	ws     *WSConn
	mu     sync.Mutex
	width  int
	height int
	closed bool

	// Input pipe: WebSocket read loop writes here, Read() reads from here.
	inputR *io.PipeReader
	inputW *io.PipeWriter
}

// resizeMsg is the JSON message sent by the browser on resize.
type resizeMsg struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

// NewWebTerminal creates a TermIO backed by a WebSocket connection.
func NewWebTerminal(ws *WSConn) *WebTerminal {
	pr, pw := io.Pipe()
	wt := &WebTerminal{
		ws:     ws,
		width:  80,
		height: 24,
		inputR: pr,
		inputW: pw,
	}
	go wt.readLoop()
	return wt
}

// Write sends terminal output to the browser via WebSocket binary message.
func (wt *WebTerminal) Write(p []byte) (int, error) {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	if wt.closed {
		return 0, io.ErrClosedPipe
	}
	err := wt.ws.WriteBinary(p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// Read reads terminal input (keystrokes from browser).
func (wt *WebTerminal) Read(p []byte) (int, error) {
	return wt.inputR.Read(p)
}

// Size returns the current terminal dimensions.
func (wt *WebTerminal) Size() (int, int) {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	return wt.width, wt.height
}

// SetSize updates terminal dimensions.
func (wt *WebTerminal) SetSize(w, h int) {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	wt.width = w
	wt.height = h
}

// Close closes the WebTerminal and underlying WebSocket.
func (wt *WebTerminal) Close() error {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	if wt.closed {
		return nil
	}
	wt.closed = true
	wt.inputW.Close()
	return wt.ws.Close()
}

// readLoop reads WebSocket messages and routes them:
// - Text messages starting with '{' are parsed as JSON control messages (resize)
// - Other messages are raw terminal input (keystrokes)
func (wt *WebTerminal) readLoop() {
	defer wt.inputW.Close()
	for {
		op, data, err := wt.ws.ReadMessage()
		if err != nil {
			return
		}

		switch op {
		case wsOpText:
			// Try to parse as JSON control message
			if len(data) > 0 && data[0] == '{' {
				var msg resizeMsg
				if json.Unmarshal(data, &msg) == nil && msg.Type == "resize" {
					if msg.Cols > 0 && msg.Rows > 0 {
						wt.SetSize(msg.Cols, msg.Rows)
					}
					continue
				}
			}
			// Otherwise treat as text input (keystrokes)
			wt.inputW.Write(data)

		case wsOpBinary:
			// Binary = raw terminal input
			wt.inputW.Write(data)
		}
	}
}

// Ensure WebTerminal implements TermIO at compile time.
var _ TermIO = (*WebTerminal)(nil)
