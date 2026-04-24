//go:build linux || darwin

package lumina

import (
	"os"
	"syscall"
	"unsafe"
)

// Terminal handles raw mode and input reading for the terminal.
type Terminal struct {
	fd       int
	origTerm syscall.Termios
	rawMode  bool
}

// NewTerminal creates a terminal reader for stdin.
func NewTerminal() (*Terminal, error) {
	fd := int(os.Stdin.Fd())
	var orig syscall.Termios
	if err := tcgetattr(fd, &orig); err != nil {
		return nil, err
	}
	return &Terminal{fd: fd, origTerm: orig}, nil
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

	// Enable mouse reporting (SGR extended mode)
	os.Stdout.WriteString("\x1b[?1000h") // basic mouse tracking
	os.Stdout.WriteString("\x1b[?1006h") // SGR extended mouse format

	return nil
}

// RestoreMode restores the original terminal mode.
func (t *Terminal) RestoreMode() {
	if t.rawMode {
		// Disable mouse reporting
		os.Stdout.WriteString("\x1b[?1006l")
		os.Stdout.WriteString("\x1b[?1000l")

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
