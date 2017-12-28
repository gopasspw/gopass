package pwgen

import (
	"os"
	"testing"
)

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

func TestPwgenCharset(t *testing.T) {
	_ = os.Setenv("GOPASS_CHARACTER_SET", "a")
	pw := GeneratePassword(4, true)
	if pw != "aaaa" {
		t.Errorf("Wrong password: %s", pw)
	}
}
