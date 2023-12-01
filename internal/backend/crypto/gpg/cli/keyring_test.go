package cli

import (
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pubkey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

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
oLGNPe8bErLNfny6AWU0Enam6a13BxwbBrtr
=AmFu
-----END PGP PUBLIC KEY BLOCK-----
`

func TestReadNamesFromKey(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	g, err := New(ctx, Config{})
	require.NoError(t, err)
	assert.NotEqual(t, "", g.Binary())

	names, err := g.ReadNamesFromKey(ctx, []byte(pubkey))
	require.NoError(t, err)
	assert.Equal(t, []string{"Gopass Archive Signing Key <gopass@justwatch.com>"}, names)
}

func TestExportPublicKey(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	g, err := New(ctx, Config{})
	require.NoError(t, err)

	_, err = g.ExportPublicKey(ctx, "foobar")
	require.Error(t, err)
}

func TestImport(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	g := &GPG{}
	g.binary = "true"
	if runtime.GOOS == "windows" {
		g.binary = "rundll32"
	}

	require.NoError(t, g.ImportPublicKey(ctx, []byte("foobar")))

	g.binary = ""
	require.Error(t, g.ImportPublicKey(ctx, []byte("foobar")))
}
