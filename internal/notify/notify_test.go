package notify

import (
	"context"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNotify(t *testing.T) {
	ctx := config.NewNoWrites().WithConfig(context.Background())

	t.Setenv("GOPASS_NO_NOTIFY", "true")
	require.NoError(t, Notify(ctx, "foo", "bar"))
}

func TestIcon(t *testing.T) {
	t.Parallel()

	fn := strings.TrimPrefix(iconURI(), "file://")
	require.NoError(t, os.Remove(fn))
	_ = iconURI()
	fh, err := os.Open(fn)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, fh.Close())
	}()

	require.NotNil(t, fh)
	_, err = png.Decode(fh)
	require.NoError(t, err)
}
