// +build windows

package gpg

func umask(mask int) int {
	return -1
}
