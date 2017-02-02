package pwgen

import "testing"

func TestPwgen(t *testing.T) {
	for _, sym := range []bool{true, false} {
		for i := 0; i < 50; i++ {
			sec := GeneratePassword(i, sym)
			if len(sec) != i {
				t.Errorf("Length mismatch")
			}
		}
	}
}
