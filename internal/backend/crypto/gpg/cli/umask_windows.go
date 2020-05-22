// +build windows

package cli

func umask(mask int) int {
	return -1
}
