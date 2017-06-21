// build !windows

package termutil

import (
	"syscall"
	"unsafe"
)

type termSize struct {
	Rows   uint16
	Cols   uint16
	XPixel uint16
	YPixel uint16
}

// GetTermsize returns the size of the current terminal
func GetTermsize() (int, int) {
	ts := termSize{}
	ret, _, _ := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ts)),
	)
	if int(ret) == -1 {
		return -1, -1
	}
	return int(ts.Rows), int(ts.Cols)
}
