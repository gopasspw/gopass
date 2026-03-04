//go:build gofuzz

package colons

import "bytes"

func Fuzz(data []byte) int {
	if kl := Parse(bytes.NewReader(data)); len(kl) != 0 {
		return 1
	}
	return 0
}
