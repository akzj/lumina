//go:build linux || darwin

package lumina

import (
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"
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
	origTerm syscall.Termios
	rawMode  bool
	resizeCh chan os.Signal // SIGWINCH channel
	output   io.Writer     // where to send escape sequences (default: os.Stdout)
}

// NewTerminal creates a terminal reader for stdin.
func NewTerminal() (*Terminal, error) {
	fd := int(os.Stdin.Fd())
	var orig syscall.Termios
	if err := tcgetattr(fd, &orig); err != nil {
		return nil, err
	}
	return &Terminal{fd: fd, origTerm: orig, output: os.Stdout}, nil
}

// NewTerminalWithOutput creates a terminal reader with a custom output writer.
func NewTerminalWithOutput(w io.Writer) (*Terminal, error) {
	fd := int(os.Stdin.Fd())
	var orig syscall.Termios
	if err := tcgetattr(fd, &orig); err != nil {
		return nil, err
	}
	return &Terminal{fd: fd, origTerm: orig, output: w}, nil
}

// EnableRawMode puts the terminal into raw mode.
func (t *Terminal) EnableRawMode() error {
	raw := t.origTerm

	// Disable input processing
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK |
		syscall.ISTRIP | syscall.INLCR | syscall.IGNCR |
		syscall.ICRNL | syscall.IXON

	// Disable output processing
	raw.Oflag &^= syscall.OPOST

	// Disable echo, canonical mode, signals, extended input
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON |
		syscall.ISIG | syscall.IEXTEN

	// Set character size to 8 bits
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8

	// Read returns after 1 byte, no timeout
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

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
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(t.fd),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)),
	)
	if errno != 0 {
		return 80, 24, nil // fallback defaults
	}
	return int(ws.Col), int(ws.Row), nil
}

// tcgetattr retrieves terminal attributes via ioctl.
func tcgetattr(fd int, termios *syscall.Termios) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(termios)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// tcsetattr sets terminal attributes via ioctl.
func tcsetattr(fd int, termios *syscall.Termios) error {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(termios)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// WatchResize starts a goroutine that calls callback whenever the terminal
// is resized (SIGWINCH). Call StopResize to stop watching.
func (t *Terminal) WatchResize(callback func(width, height int)) {
	t.resizeCh = make(chan os.Signal, 1)
	signal.Notify(t.resizeCh, syscall.SIGWINCH)
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
