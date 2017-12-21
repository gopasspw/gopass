package xkcdgen

import (
	"strings"
	"testing"
)

func TestRandom(t *testing.T) {
	pw := Random()
	if len(pw) < 4 {
		t.Errorf("too short")
	}
	if len(strings.Fields(pw)) < 4 {
		t.Errorf("too few words")
	}
}
