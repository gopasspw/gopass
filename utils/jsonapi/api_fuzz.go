// +build gofuzz

package jsonapi

import "bytes"

func Fuzz(data []byte) int {
	if b, err := readMessage(bytes.NewReader(data)); err != nil {
		if b != nil {
			panic("body != nil on error")
		}
		return 0
	}
	return 1
}
