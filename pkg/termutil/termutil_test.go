package termutil

import "testing"

func TestGetTermsize(t *testing.T) {
	x, y := GetTermsize()
	if x < -1 || y < -1 {
		t.Errorf("Should be at least -1")
	}
}
