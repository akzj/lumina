//go:build linux || darwin

package lumina

import (
	"io"
	"os"
	"os/signal"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ColorMode represents the terminal's color capability.
type ColorMode int

const (
	// ColorNone means no color support (e.g., TERM=dumb).
	ColorNone ColorMode = iota
	// Color16 means basic 16-color ANSI support.
	Color16
	// Color256 means 256-color xterm support.
	Color256
	// ColorTrue means 24-bit true color support.
	ColorTrue
)

// String returns a human-readable name for ColorMode.
func (c ColorMode) String() string {
	switch c {
	case ColorNone:
		return "none"
	case Color16:
		return "16"
	case Color256:
		return "256"
	case ColorTrue:
		return "truecolor"
	default:
		return "unknown"
	}
}

// DetectColorMode detects the terminal's color capability from environment variables.
func DetectColorMode() ColorMode {
	// COLORTERM is the most reliable indicator for true color.
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return ColorTrue
	}

	// TERM variable.
	term := os.Getenv("TERM")
	if term == "dumb" || term == "" {
		return ColorNone
	}
	if strings.Contains(term, "256color") {
		return Color256
	}
	// Most terminals with a TERM value support at least 16 colors.
	return Color16
}

// Terminal handles raw mode and input reading for the terminal.
type Terminal struct {
	fd       int
	origTerm unix.Termios
	rawMode  bool
	resizeCh chan os.Signal // SIGWINCH channel
	output   io.Writer      // where to send escape sequences (default: os.Stdout)
}

// NewTerminal creates a terminal reader for stdin.
func NewTerminal() (*Terminal, error) {
	fd := int(os.Stdin.Fd())
	var orig unix.Termios
	if err := tcgetattr(fd, &orig); err != nil {
		return nil, err
	}
	return &Terminal{fd: fd, origTerm: orig, output: os.Stdout}, nil
}

// NewTerminalWithOutput creates a terminal reader with a custom output writer.
func NewTerminalWithOutput(w io.Writer) (*Terminal, error) {
	fd := int(os.Stdin.Fd())
	var orig unix.Termios
	if err := tcgetattr(fd, &orig); err != nil {
		return nil, err
	}
	return &Terminal{fd: fd, origTerm: orig, output: w}, nil
}

// EnableRawMode puts the terminal into raw mode.
func (t *Terminal) EnableRawMode() error {
	raw := t.origTerm

	// Disable input processing
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK |
		unix.ISTRIP | unix.INLCR | unix.IGNCR |
		unix.ICRNL | unix.IXON

	// Disable output processing
	raw.Oflag &^= unix.OPOST

	// Disable echo, canonical mode, signals, extended input
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON |
		unix.ISIG | unix.IEXTEN

	// Set character size to 8 bits
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8

	// Read returns after 1 byte, no timeout
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := tcsetattr(t.fd, &raw); err != nil {
		return err
	}
	t.rawMode = true

	// Enter alternate screen buffer (preserves main screen on exit)
	t.output.Write([]byte("\x1b[?1049h"))
	// Hide cursor during rendering
	t.output.Write([]byte("\x1b[?25l"))

	// Enable mouse reporting (SGR extended mode)
	t.output.Write([]byte("\x1b[?1003h")) // ANY_EVENT mouse tracking (motion + click)
	t.output.Write([]byte("\x1b[?1006h")) // SGR extended mouse format

	return nil
}

// RestoreMode restores the original terminal mode.
func (t *Terminal) RestoreMode() {
	if t.rawMode {
		// Disable mouse reporting
		t.output.Write([]byte("\x1b[?1006l"))
		t.output.Write([]byte("\x1b[?1003l"))

		// Show cursor
		t.output.Write([]byte("\x1b[?25h"))
		// Leave alternate screen buffer (restores main screen)
		t.output.Write([]byte("\x1b[?1049l"))

		tcsetattr(t.fd, &t.origTerm)
		t.rawMode = false
	}
}

// GetSize returns the terminal width and height.
func (t *Terminal) GetSize() (width, height int, err error) {
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
		return 80, 24, nil // fallback defaults
	}
	return int(ws.Col), int(ws.Row), nil
}

// WatchResize starts a goroutine that calls callback whenever the terminal
// is resized (SIGWINCH). Call StopResize to stop watching.
func (t *Terminal) WatchResize(callback func(width, height int)) {
	t.resizeCh = make(chan os.Signal, 1)
	signal.Notify(t.resizeCh, unix.SIGWINCH)
	go func() {
		for range t.resizeCh {
			w, h, err := t.GetSize()
			if err == nil {
				callback(w, h)
			}
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
