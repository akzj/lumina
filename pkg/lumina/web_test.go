package lumina

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestWebServer_Start(t *testing.T) {
	ws := NewWebServer("127.0.0.1:0", "")
	addr, err := ws.StartBackground()
	if err != nil {
		t.Fatalf("StartBackground: %v", err)
	}
	defer ws.Stop()

	if addr == "" {
		t.Fatal("expected non-empty address")
	}

	// Verify HTTP serves index.html
	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET /: status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("Lumina")) {
		t.Fatal("index.html should contain 'Lumina'")
	}
}

func TestWebServer_LuminaClientJS(t *testing.T) {
	ws := NewWebServer("127.0.0.1:0", "")
	addr, err := ws.StartBackground()
	if err != nil {
		t.Fatalf("StartBackground: %v", err)
	}
	defer ws.Stop()

	resp, err := http.Get("http://" + addr + "/lumina-client.js")
	if err != nil {
		t.Fatalf("GET /lumina-client.js: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET /lumina-client.js: status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("WebSocket")) {
		t.Fatal("lumina-client.js should contain 'WebSocket'")
	}
}

func TestWebServer_HealthCheck(t *testing.T) {
	ws := NewWebServer("127.0.0.1:0", "")
	addr, err := ws.StartBackground()
	if err != nil {
		t.Fatalf("StartBackground: %v", err)
	}
	defer ws.Stop()

	resp, err := http.Get("http://" + addr + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("parse health: %v", err)
	}
	if health["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %v", health["status"])
	}
}

func TestWebServer_SessionCount(t *testing.T) {
	ws := NewWebServer("127.0.0.1:0", "")
	_, err := ws.StartBackground()
	if err != nil {
		t.Fatalf("StartBackground: %v", err)
	}
	defer ws.Stop()

	if ws.SessionCount() != 0 {
		t.Fatalf("expected 0 sessions, got %d", ws.SessionCount())
	}
}

func TestWebSocketUpgrade(t *testing.T) {
	ws := NewWebServer("127.0.0.1:0", "")
	addr, err := ws.StartBackground()
	if err != nil {
		t.Fatalf("StartBackground: %v", err)
	}
	defer ws.Stop()

	// Manual WebSocket handshake
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	key := base64.StdEncoding.EncodeToString([]byte("test-key-1234567"))
	req := "GET /ws HTTP/1.1\r\n" +
		"Host: " + addr + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n\r\n"
	conn.Write([]byte(req))

	// Read response
	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	resp := string(buf[:n])
	if !bytes.Contains([]byte(resp), []byte("101 Switching Protocols")) {
		t.Fatalf("expected 101, got: %s", resp)
	}

	// Verify accept key
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	expectedAccept := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if !bytes.Contains([]byte(resp), []byte(expectedAccept)) {
		t.Fatalf("missing Sec-WebSocket-Accept in response")
	}
}

func TestWebTerminalWrite(t *testing.T) {
	// Create a pipe-based mock WebSocket for testing
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	// Read the output from the server side in a goroutine
	done := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 4096)
		n, _ := serverConn.Read(buf)
		done <- buf[:n]
	}()

	// Create a WSConn wrapping the client side
	wsConn := &WSConn{
		conn: clientConn,
	}

	// Create WebTerminal
	wt := &WebTerminal{
		ws:     wsConn,
		width:  80,
		height: 24,
	}

	// Write terminal output
	testData := []byte("\x1b[2J\x1b[H Hello World")
	n, err := wt.Write(testData)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(testData) {
		t.Fatalf("expected %d bytes written, got %d", len(testData), n)
	}

	// Verify data was sent as WebSocket binary frame
	select {
	case raw := <-done:
		// First byte: 0x82 (FIN + binary opcode)
		if len(raw) < 2 {
			t.Fatal("frame too short")
		}
		if raw[0] != 0x82 {
			t.Fatalf("expected binary frame (0x82), got 0x%02x", raw[0])
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for WebSocket frame")
	}
}

func TestWebTerminalSize(t *testing.T) {
	wt := &WebTerminal{width: 80, height: 24}
	w, h := wt.Size()
	if w != 80 || h != 24 {
		t.Fatalf("expected 80x24, got %dx%d", w, h)
	}
	wt.SetSize(120, 40)
	w, h = wt.Size()
	if w != 120 || h != 40 {
		t.Fatalf("expected 120x40, got %dx%d", w, h)
	}
}

func TestResizeMessage(t *testing.T) {
	msg := resizeMsg{Type: "resize", Cols: 100, Rows: 50}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var parsed resizeMsg
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Type != "resize" || parsed.Cols != 100 || parsed.Rows != 50 {
		t.Fatalf("unexpected: %+v", parsed)
	}
}

func TestWSConn_WriteFrame(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	ws := &WSConn{conn: clientConn}

	done := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 4096)
		n, _ := serverConn.Read(buf)
		done <- buf[:n]
	}()

	// Write a text frame
	err := ws.WriteText([]byte("hello"))
	if err != nil {
		t.Fatalf("WriteText: %v", err)
	}

	select {
	case raw := <-done:
		// Byte 0: 0x81 (FIN + text opcode)
		if raw[0] != 0x81 {
			t.Fatalf("expected text frame (0x81), got 0x%02x", raw[0])
		}
		// Byte 1: length (5, no mask)
		if raw[1] != 5 {
			t.Fatalf("expected length 5, got %d", raw[1])
		}
		// Payload
		if string(raw[2:7]) != "hello" {
			t.Fatalf("expected 'hello', got %q", raw[2:7])
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestWSConn_WriteLargeFrame(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	ws := &WSConn{conn: clientConn}

	// Create a payload > 125 bytes (needs extended length)
	payload := make([]byte, 200)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	done := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 4096)
		total := 0
		for total < 204 { // 2 header + 2 extended length + 200 payload
			n, err := serverConn.Read(buf[total:])
			if err != nil {
				break
			}
			total += n
		}
		done <- buf[:total]
	}()

	err := ws.WriteBinary(payload)
	if err != nil {
		t.Fatalf("WriteBinary: %v", err)
	}

	select {
	case raw := <-done:
		if raw[0] != 0x82 {
			t.Fatalf("expected binary (0x82), got 0x%02x", raw[0])
		}
		if raw[1] != 126 {
			t.Fatalf("expected extended length marker 126, got %d", raw[1])
		}
		extLen := binary.BigEndian.Uint16(raw[2:4])
		if extLen != 200 {
			t.Fatalf("expected length 200, got %d", extLen)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
