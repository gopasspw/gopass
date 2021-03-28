// +build windows

package cli

func tty() string {
	return ""
}

func umask(mask int) int {
	return -1
}
