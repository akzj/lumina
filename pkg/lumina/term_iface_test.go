package lumina

import (
	"bytes"
	"io"
	"sync"
	"testing"
)

// MockTermIO is a TermIO backed by buffers for testing.
type MockTermIO struct {
	mu       sync.Mutex
	outBuf   bytes.Buffer
	inBuf    bytes.Buffer
	width    int
	height   int
	closed   bool
}

func NewMockTermIO(w, h int) *MockTermIO {
	return &MockTermIO{width: w, height: h}
}

func (m *MockTermIO) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.outBuf.Write(p)
}

func (m *MockTermIO) Read(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.EOF
	}
	return m.inBuf.Read(p)
}

func (m *MockTermIO) Size() (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.width, m.height
}

func (m *MockTermIO) SetSize(w, h int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.width = w
	m.height = h
}

func (m *MockTermIO) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

// WriteInput simulates user input into the TermIO.
func (m *MockTermIO) WriteInput(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inBuf.Write(data)
}

// OutputBytes returns all bytes written to the TermIO.
func (m *MockTermIO) OutputBytes() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.outBuf.Bytes()
}

var _ TermIO = (*MockTermIO)(nil)

func TestTermIO_LocalTermIO(t *testing.T) {
	lt := NewLocalTermIO()
	w, h := lt.Size()
	if w != 80 || h != 24 {
		t.Fatalf("expected 80x24, got %dx%d", w, h)
	}
	lt.SetSize(120, 40)
	w, h = lt.Size()
	if w != 120 || h != 40 {
		t.Fatalf("after SetSize: expected 120x40, got %dx%d", w, h)
	}
	if err := lt.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestTermIO_MockWrite(t *testing.T) {
	m := NewMockTermIO(80, 24)
	n, err := m.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != 5 {
		t.Fatalf("expected 5 bytes written, got %d", n)
	}
	if string(m.OutputBytes()) != "hello" {
		t.Fatalf("output mismatch: %q", m.OutputBytes())
	}
}

func TestTermIO_MockRead(t *testing.T) {
	m := NewMockTermIO(80, 24)
	m.WriteInput([]byte("key"))
	buf := make([]byte, 10)
	n, err := m.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if string(buf[:n]) != "key" {
		t.Fatalf("read mismatch: %q", buf[:n])
	}
}

func TestTermIO_MockResize(t *testing.T) {
	m := NewMockTermIO(80, 24)
	m.SetSize(100, 50)
	w, h := m.Size()
	if w != 100 || h != 50 {
		t.Fatalf("expected 100x50, got %dx%d", w, h)
	}
}

func TestTermIO_MockClose(t *testing.T) {
	m := NewMockTermIO(80, 24)
	m.Close()
	_, err := m.Write([]byte("x"))
	if err != io.ErrClosedPipe {
		t.Fatalf("expected ErrClosedPipe after close, got %v", err)
	}
}

func TestTermIO_Interface(t *testing.T) {
	// Verify both types implement TermIO
	var _ TermIO = (*LocalTermIO)(nil)
	var _ TermIO = (*MockTermIO)(nil)
}

func TestTermIO_ANSIAdapterWithMock(t *testing.T) {
	m := NewMockTermIO(40, 10)
	adapter := NewANSIAdapter(m)
	frame := NewFrame(40, 10)
	frame.SetCell(0, 0, Cell{Char: 'A', Foreground: "#FFFFFF"})
	frame.MarkDirty()
	err := adapter.Write(frame)
	if err != nil {
		t.Fatalf("adapter.Write: %v", err)
	}
	out := m.OutputBytes()
	if len(out) == 0 {
		t.Fatal("expected output from ANSI adapter, got nothing")
	}
	// Should contain ANSI escape sequences
	if !bytes.Contains(out, []byte("\x1b[")) {
		t.Fatal("expected ANSI escape sequences in output")
	}
}
