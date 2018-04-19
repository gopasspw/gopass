package twofactor

import (
	"strings"
)

// Pad calculates the number of '='s to add to our encoded string
// to make base32.StdEncoding.DecodeString happy
func Pad(s string) string {
	if !strings.HasSuffix(s, "=") && len(s)%8 != 0 {
		for len(s)%8 != 0 {
			s += "="
		}
	}
	return s
}
