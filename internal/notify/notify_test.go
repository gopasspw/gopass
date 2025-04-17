package notify

import (
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNotify(t *testing.T) {
	ctx := config.NewContextInMemory()

	t.Setenv("GOPASS_NO_NOTIFY", "true")
	require.NoError(t, Notify(ctx, "foo", "bar"))
}

func TestIcon(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	fn := strings.TrimPrefix(iconURI(ctx), "file://")
	require.NoError(t, os.Remove(fn))
	_ = iconURI(ctx)
	fh, err := os.Open(fn)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, fh.Close())
	}()

	require.NotNil(t, fh)
	_, err = png.Decode(fh)
	require.NoError(t, err)
}
