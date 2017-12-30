package notify

import (
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotify(t *testing.T) {
	_ = os.Setenv("GOPASS_NO_NOTIFY", "true")
	err := Notify("foo", "bar")
	t.Logf("Error: %s", err)
}

func TestIcon(t *testing.T) {
	fn := strings.TrimPrefix(iconURI(), "file://")
	fh, err := os.Open(fn)
	if err != nil {
		t.Fatalf("Failed to open file %s: %s", fn, err)
	}
	defer func() {
		assert.NoError(t, fh.Close())
	}()
	_, err = png.Decode(fh)
	if err != nil {
		t.Fatalf("Failed to decode icon: %s", err)
	}
}
