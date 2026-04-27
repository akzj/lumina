package output

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/akzj/lumina/pkg/lumina/v2/buffer"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// helper: create a WSAdapter on a random port and start it.
func newTestWSAdapter(t *testing.T) (*WSAdapter, chan WSEvent) {
	t.Helper()
	eventCh := make(chan WSEvent, 64)
	ws := NewWSAdapter("127.0.0.1:0", 40, 10, eventCh)
	if err := ws.Start(); err != nil {
		t.Fatalf("ws.Start: %v", err)
	}
	t.Cleanup(func() { ws.Close() })
	return ws, eventCh
}

// helper: connect a WebSocket client to the test server.
func connectClient(t *testing.T, ws *WSAdapter) *websocket.Conn {
	t.Helper()
	addr := ws.Addr().String()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws://"+addr+"/ws", nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "") })
	return conn
}

// helper: read the init message from a freshly connected client.
func readInitMessage(t *testing.T, conn *websocket.Conn) WSMessage {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg WSMessage
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read init: %v", err)
	}
	if msg.Type != "init" {
		t.Fatalf("expected init message, got %q", msg.Type)
	}
	return msg
}

func TestWSAdapter_InitMessage(t *testing.T) {
	ws, _ := newTestWSAdapter(t)
	conn := connectClient(t, ws)

	msg := readInitMessage(t, conn)

	var initData struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := json.Unmarshal(msg.Data, &initData); err != nil {
		t.Fatalf("unmarshal init: %v", err)
	}
	if initData.Width != 40 || initData.Height != 10 {
		t.Fatalf("init size = %dx%d, want 40x10", initData.Width, initData.Height)
	}
}

func TestWSAdapter_WriteFull(t *testing.T) {
	ws, _ := newTestWSAdapter(t)
	conn := connectClient(t, ws)
	readInitMessage(t, conn) // consume init

	// Create a buffer with some content.
	buf := buffer.New(4, 2)
	buf.Set(0, 0, buffer.Cell{Char: 'H', Foreground: "#FF0000"})
	buf.Set(1, 0, buffer.Cell{Char: 'i'})

	if err := ws.WriteFull(buf); err != nil {
		t.Fatalf("WriteFull: %v", err)
	}

	// Read the message from the client.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg WSMessage
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read full: %v", err)
	}

	if msg.Type != "full" {
		t.Fatalf("type = %q, want full", msg.Type)
	}

	var result RenderResult
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if result.Width != 4 || result.Height != 2 {
		t.Fatalf("size = %dx%d, want 4x2", result.Width, result.Height)
	}
	if result.Cells[0][0].Char != "H" {
		t.Fatalf("cell[0][0].Char = %q, want H", result.Cells[0][0].Char)
	}
	if result.Cells[0][0].Fg != "#FF0000" {
		t.Fatalf("cell[0][0].Fg = %q, want #FF0000", result.Cells[0][0].Fg)
	}
}

func TestWSAdapter_WriteDirty(t *testing.T) {
	ws, _ := newTestWSAdapter(t)
	conn := connectClient(t, ws)
	readInitMessage(t, conn)

	buf := buffer.New(4, 2)
	buf.Set(0, 0, buffer.Cell{Char: 'A'})

	dirtyRects := []buffer.Rect{{X: 0, Y: 0, W: 2, H: 1}}
	if err := ws.WriteDirty(buf, dirtyRects); err != nil {
		t.Fatalf("WriteDirty: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg WSMessage
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read dirty: %v", err)
	}

	if msg.Type != "dirty" {
		t.Fatalf("type = %q, want dirty", msg.Type)
	}

	var result RenderResult
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(result.DirtyRects) != 1 {
		t.Fatalf("dirty_rects len = %d, want 1", len(result.DirtyRects))
	}
	if result.DirtyRects[0].W != 2 || result.DirtyRects[0].H != 1 {
		t.Fatalf("dirty rect = %+v, want {0 0 2 1}", result.DirtyRects[0])
	}
}

func TestWSAdapter_InputEvents(t *testing.T) {
	ws, eventCh := newTestWSAdapter(t)
	conn := connectClient(t, ws)
	readInitMessage(t, conn)

	// Send a keydown event from the "browser".
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inputMsg := WSInputMessage{
		Type: "keydown",
		Data: WSInputData{
			Key:  "Enter",
			Ctrl: true,
		},
	}
	if err := wsjson.Write(ctx, conn, inputMsg); err != nil {
		t.Fatalf("write input: %v", err)
	}

	// Read the event from the channel.
	select {
	case evt := <-eventCh:
		if evt.Type != "keydown" {
			t.Fatalf("event type = %q, want keydown", evt.Type)
		}
		if evt.Key != "Enter" {
			t.Fatalf("event key = %q, want Enter", evt.Key)
		}
		if !evt.Ctrl {
			t.Fatal("event ctrl = false, want true")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for input event")
	}
}

func TestWSAdapter_MouseEvent(t *testing.T) {
	ws, eventCh := newTestWSAdapter(t)
	conn := connectClient(t, ws)
	readInitMessage(t, conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inputMsg := WSInputMessage{
		Type: "mousedown",
		Data: WSInputData{
			X:      5,
			Y:      3,
			Button: "left",
		},
	}
	if err := wsjson.Write(ctx, conn, inputMsg); err != nil {
		t.Fatalf("write mouse: %v", err)
	}

	select {
	case evt := <-eventCh:
		if evt.Type != "mousedown" {
			t.Fatalf("event type = %q, want mousedown", evt.Type)
		}
		if evt.X != 5 || evt.Y != 3 {
			t.Fatalf("event pos = (%d,%d), want (5,3)", evt.X, evt.Y)
		}
		if evt.Button != "left" {
			t.Fatalf("event button = %q, want left", evt.Button)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for mouse event")
	}
}

func TestWSAdapter_MultipleClients(t *testing.T) {
	ws, _ := newTestWSAdapter(t)

	conn1 := connectClient(t, ws)
	readInitMessage(t, conn1)

	conn2 := connectClient(t, ws)
	readInitMessage(t, conn2)

	// WriteFull should broadcast to both.
	buf := buffer.New(2, 1)
	buf.Set(0, 0, buffer.Cell{Char: 'X'})
	if err := ws.WriteFull(buf); err != nil {
		t.Fatalf("WriteFull: %v", err)
	}

	// Both clients should receive the message.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i, conn := range []*websocket.Conn{conn1, conn2} {
		var msg WSMessage
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			t.Fatalf("client %d read: %v", i, err)
		}
		if msg.Type != "full" {
			t.Fatalf("client %d type = %q, want full", i, msg.Type)
		}
	}
}

func TestWSAdapter_ClientDisconnect(t *testing.T) {
	ws, _ := newTestWSAdapter(t)

	conn1 := connectClient(t, ws)
	readInitMessage(t, conn1)

	conn2 := connectClient(t, ws)
	readInitMessage(t, conn2)

	// Disconnect conn1.
	conn1.Close(websocket.StatusNormalClosure, "bye")

	// Give the server a moment to process the disconnect.
	time.Sleep(100 * time.Millisecond)

	// WriteFull should still work (conn2 still connected).
	buf := buffer.New(2, 1)
	buf.Set(0, 0, buffer.Cell{Char: 'Y'})
	if err := ws.WriteFull(buf); err != nil {
		t.Fatalf("WriteFull after disconnect: %v", err)
	}

	// conn2 should receive the message.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var msg WSMessage
	if err := wsjson.Read(ctx, conn2, &msg); err != nil {
		t.Fatalf("conn2 read: %v", err)
	}
	if msg.Type != "full" {
		t.Fatalf("type = %q, want full", msg.Type)
	}
}

func TestWSAdapter_SetSize(t *testing.T) {
	ws, _ := newTestWSAdapter(t)

	ws.SetSize(120, 40)

	conn := connectClient(t, ws)
	msg := readInitMessage(t, conn)

	var initData struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}
	if err := json.Unmarshal(msg.Data, &initData); err != nil {
		t.Fatalf("unmarshal init: %v", err)
	}
	if initData.Width != 120 || initData.Height != 40 {
		t.Fatalf("init size = %dx%d, want 120x40", initData.Width, initData.Height)
	}
}
