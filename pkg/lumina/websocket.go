package lumina

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

// wsGUID is the WebSocket magic GUID per RFC 6455.
const wsGUID = "258EAFA5-E914-47DA-95CA-5AB5DC30CE7A"

// wsOpcode represents WebSocket frame opcodes.
type wsOpcode byte

const (
	wsOpContinuation wsOpcode = 0x0
	wsOpText         wsOpcode = 0x1
	wsOpBinary       wsOpcode = 0x2
	wsOpClose        wsOpcode = 0x8
	wsOpPing         wsOpcode = 0x9
	wsOpPong         wsOpcode = 0xA
)

// WSConn is a minimal WebSocket connection.
type WSConn struct {
	conn   net.Conn
	br     *bufio.Reader
	mu     sync.Mutex // protects writes
	closed bool
}

// UpgradeWebSocket performs the HTTP→WebSocket upgrade handshake.
// Returns a WSConn on success, or writes an error response and returns nil.
func UpgradeWebSocket(w http.ResponseWriter, r *http.Request) *WSConn {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		http.Error(w, "not a websocket request", http.StatusBadRequest)
		return nil
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "missing Sec-WebSocket-Key", http.StatusBadRequest)
		return nil
	}

	// Compute accept key
	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Hijack the connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server doesn't support hijacking", http.StatusInternalServerError)
		return nil
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	// Write upgrade response
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	bufrw.WriteString(resp)
	bufrw.Flush()

	return &WSConn{
		conn: conn,
		br:   bufrw.Reader,
	}
}

// ReadMessage reads a complete WebSocket message (handles fragmentation).
// Returns the opcode and payload.
func (ws *WSConn) ReadMessage() (wsOpcode, []byte, error) {
	var payload []byte
	var firstOp wsOpcode

	for {
		fin, op, data, err := ws.readFrame()
		if err != nil {
			return 0, nil, err
		}

		// Handle control frames inline
		switch op {
		case wsOpPing:
			ws.writeFrame(wsOpPong, data)
			continue
		case wsOpPong:
			continue
		case wsOpClose:
			ws.writeFrame(wsOpClose, nil)
			return wsOpClose, nil, io.EOF
		}

		if firstOp == 0 && op != wsOpContinuation {
			firstOp = op
		}

		payload = append(payload, data...)

		if fin {
			return firstOp, payload, nil
		}
	}
}

// WriteMessage writes a complete WebSocket message.
func (ws *WSConn) WriteMessage(op wsOpcode, data []byte) error {
	return ws.writeFrame(op, data)
}

// WriteText writes a text message.
func (ws *WSConn) WriteText(data []byte) error {
	return ws.WriteMessage(wsOpText, data)
}

// WriteBinary writes a binary message.
func (ws *WSConn) WriteBinary(data []byte) error {
	return ws.WriteMessage(wsOpBinary, data)
}

// Close closes the WebSocket connection.
func (ws *WSConn) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	if ws.closed {
		return nil
	}
	ws.closed = true
	// Send close frame (best effort)
	ws.writeFrameUnlocked(wsOpClose, nil)
	return ws.conn.Close()
}

// readFrame reads a single WebSocket frame.
func (ws *WSConn) readFrame() (fin bool, op wsOpcode, payload []byte, err error) {
	// Read first 2 bytes
	var header [2]byte
	if _, err = io.ReadFull(ws.br, header[:]); err != nil {
		return
	}

	fin = header[0]&0x80 != 0
	op = wsOpcode(header[0] & 0x0F)
	masked := header[1]&0x80 != 0
	length := uint64(header[1] & 0x7F)

	// Extended payload length
	switch length {
	case 126:
		var ext [2]byte
		if _, err = io.ReadFull(ws.br, ext[:]); err != nil {
			return
		}
		length = uint64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err = io.ReadFull(ws.br, ext[:]); err != nil {
			return
		}
		length = binary.BigEndian.Uint64(ext[:])
	}

	// Masking key (client→server frames are always masked)
	var mask [4]byte
	if masked {
		if _, err = io.ReadFull(ws.br, mask[:]); err != nil {
			return
		}
	}

	// Payload
	if length > 16*1024*1024 { // 16MB limit
		err = fmt.Errorf("websocket: frame too large (%d bytes)", length)
		return
	}
	payload = make([]byte, length)
	if _, err = io.ReadFull(ws.br, payload); err != nil {
		return
	}

	// Unmask
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}

	return
}

// writeFrame writes a single WebSocket frame (server→client, unmasked).
func (ws *WSConn) writeFrame(op wsOpcode, data []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.writeFrameUnlocked(op, data)
}

func (ws *WSConn) writeFrameUnlocked(op wsOpcode, data []byte) error {
	length := len(data)

	// First byte: FIN + opcode
	var header []byte
	header = append(header, 0x80|byte(op))

	// Length encoding (server→client: no mask bit)
	if length <= 125 {
		header = append(header, byte(length))
	} else if length <= 65535 {
		header = append(header, 126)
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(length))
		header = append(header, ext[:]...)
	} else {
		header = append(header, 127)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(length))
		header = append(header, ext[:]...)
	}

	// Write header + payload atomically
	buf := make([]byte, 0, len(header)+length)
	buf = append(buf, header...)
	buf = append(buf, data...)
	_, err := ws.conn.Write(buf)
	return err
}
