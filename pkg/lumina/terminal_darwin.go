//go:build darwin

package lumina

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

// tcgetattr retrieves terminal attributes via ioctl on macOS/Darwin.
func tcgetattr(fd int, termios *unix.Termios) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.TIOCGETA),
		uintptr(unsafe.Pointer(termios)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// tcsetattr sets terminal attributes via ioctl on macOS/Darwin.
func tcsetattr(fd int, termios *unix.Termios) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.TIOCSETA),
		uintptr(unsafe.Pointer(termios)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}
