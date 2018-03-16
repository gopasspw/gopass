// +build gofuzz

package cli

import "bytes"

func Fuzz(data []byte) int {
	if kl := parseColons(bytes.NewReader(data)); len(kl) != 0 {
		return 1
	}
	return 0
}
