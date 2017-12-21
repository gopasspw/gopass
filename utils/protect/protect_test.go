package protect

import "testing"

func TestProtect(t *testing.T) {
	if err := Pledge(""); err != nil {
		t.Errorf("Error: %s", err)
	}
}
