// +build windows

package termutil

// GetTermsize is not available on windows
func GetTermsize() (int, int) {
	return -1, -1
}
