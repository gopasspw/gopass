//go:build windows
// +build windows

package gpgconf

func TTY() string {
	return ""
}

func Umask(mask int) int {
	return -1
}
