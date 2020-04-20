// +build !windows

package notify

import (
	"context"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotify(t *testing.T) {
	ctx := context.Background()
	_ = os.Setenv("GOPASS_NO_NOTIFY", "true")
	assert.NoError(t, Notify(ctx, "foo", "bar"))
}

func TestIcon(t *testing.T) {
	icon, err := iconURI("")
	assert.NoError(t, err)
	fn := strings.TrimPrefix(icon, "file://")
	_ = os.Remove(fn)
	_, _ = iconURI("")
	fh, err := os.Open(fn)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, fh.Close())
	}()
	require.NotNil(t, fh)
	_, err = png.Decode(fh)
	assert.NoError(t, err)
}
