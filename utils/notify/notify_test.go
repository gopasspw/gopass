package notify

import (
	"os"
	"testing"
)

func TestNotify(t *testing.T) {
	_ = os.Setenv("GOPASS_NO_NOTIFY", "true")
	err := Notify("foo", "bar")
	t.Logf("Error: %s", err)
}
