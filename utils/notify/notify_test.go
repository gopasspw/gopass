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
	assert.NoError(t, Notify("foo", "bar"))
}

func TestIcon(t *testing.T) {
	fn := strings.TrimPrefix(iconURI(), "file://")
	_ = os.Remove(fn)
	_ = iconURI()
	fh, err := os.Open(fn)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, fh.Close())
	}()
	_, err = png.Decode(fh)
	assert.NoError(t, err)
}
