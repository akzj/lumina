//go:build linux

package terminal

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

// tcgetattr retrieves terminal attributes via ioctl on Linux.
func tcgetattr(fd int, termios *unix.Termios) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.TCGETS),
		uintptr(unsafe.Pointer(termios)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

// tcsetattr sets terminal attributes via ioctl on Linux.
func tcsetattr(fd int, termios *unix.Termios) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.TCSETS),
		uintptr(unsafe.Pointer(termios)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}
