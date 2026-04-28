//go:build linux || darwin

// Package terminal provides raw mode control, resize detection, and input
// parsing for Lumina v2. It has no dependencies on other v2 packages.
package terminal

import (
	"io"
	"os"
	"os/signal"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Terminal manages raw mode, mouse reporting, and resize detection.
type Terminal struct {
	fd       int
	origTerm unix.Termios
	rawMode  bool
	output   io.Writer
	resizeCh chan os.Signal
}

// New creates a Terminal that reads from stdin and writes to stdout.
func New() (*Terminal, error) {
	return NewWithOutput(os.Stdout)
}

// NewWithOutput creates a Terminal that reads from stdin and writes to w.
func NewWithOutput(w io.Writer) (*Terminal, error) {
	fd := int(os.Stdin.Fd())
	var orig unix.Termios
	if err := tcgetattr(fd, &orig); err != nil {
		return nil, err
	}
	return &Terminal{fd: fd, origTerm: orig, output: w}, nil
}

// EnableRawMode puts the terminal into raw mode and enables alternate screen,
// cursor hiding, and mouse reporting.
func (t *Terminal) EnableRawMode() error {
	raw := t.origTerm

	// Disable input processing.
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK |
		unix.ISTRIP | unix.INLCR | unix.IGNCR |
		unix.ICRNL | unix.IXON

	// Disable output processing.
	raw.Oflag &^= unix.OPOST

	// Disable echo, canonical mode, signals, extended input.
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON |
		unix.ISIG | unix.IEXTEN

	// Set character size to 8 bits, no parity.
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8

	// Read returns after 1 byte, no timeout.
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := tcsetattr(t.fd, &raw); err != nil {
		return err
	}
	t.rawMode = true

	// Enter alternate screen buffer.
	t.output.Write([]byte("\x1b[?1049h"))
	// Hide cursor.
	t.output.Write([]byte("\x1b[?25l"))
	// Enable ANY_EVENT mouse tracking.
	t.output.Write([]byte("\x1b[?1003h"))
	// Enable SGR extended mouse format.
	t.output.Write([]byte("\x1b[?1006h"))

	return nil
}

// RestoreMode restores the original terminal mode and disables mouse
// reporting, shows the cursor, and leaves the alternate screen buffer.
func (t *Terminal) RestoreMode() {
	if !t.rawMode {
		return
	}
	// Disable SGR mouse format.
	t.output.Write([]byte("\x1b[?1006l"))
	// Disable mouse tracking.
	t.output.Write([]byte("\x1b[?1003l"))
	// Show cursor.
	t.output.Write([]byte("\x1b[?25h"))
	// Leave alternate screen buffer.
	t.output.Write([]byte("\x1b[?1049l"))

	tcsetattr(t.fd, &t.origTerm)
	t.rawMode = false
}

// Size returns the terminal width and height via TIOCGWINSZ.
// Falls back to 80×24 on error.
func (t *Terminal) Size() (width, height int) {
	var ws struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(t.fd),
		uintptr(unix.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	)
	if errno != 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}

// Output returns the writer used for escape sequences.
func (t *Terminal) Output() io.Writer {
	return t.output
}

// WatchResize starts a goroutine that calls callback(w, h) whenever the
// terminal is resized (SIGWINCH). Call StopResize to stop.
func (t *Terminal) WatchResize(callback func(w, h int)) {
	t.resizeCh = make(chan os.Signal, 1)
	signal.Notify(t.resizeCh, unix.SIGWINCH)
	go func() {
		for range t.resizeCh {
			w, h := t.Size()
			callback(w, h)
		}
	}()
}

// StopResize stops watching for terminal resize signals.
func (t *Terminal) StopResize() {
	if t.resizeCh != nil {
		signal.Stop(t.resizeCh)
		close(t.resizeCh)
		t.resizeCh = nil
	}
}
