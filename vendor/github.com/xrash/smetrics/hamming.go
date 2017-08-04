package smetrics

import (
	"fmt"
)

func Hamming(a, b string) (int, error) {
	al := len(a)
	bl := len(b)

	if al != bl {
		return -1, fmt.Errorf("strings are not equal (len(a)=%d, len(b)=%d)", al, bl)
	}

	var difference = 0

	for i := range a {
		if a[i] != b[i] {
			difference = difference + 1
		}
	}

	return difference, nil
}
