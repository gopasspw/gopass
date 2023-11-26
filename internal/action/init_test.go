package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithStorageBackend(ctx, backend.FS)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	c := gptest.CliCtx(ctx, t, "foo.bar@example.org")
	require.NoError(t, act.IsInitialized(c))
	require.Error(t, act.Init(c))
	require.NoError(t, act.Setup(c))

	crypto := act.Store.Crypto(ctx, "")
	require.NotNil(t, crypto)
	assert.Equal(t, "plain", crypto.Name())
	assert.True(t, act.initHasUseablePrivateKeys(ctx, crypto))
	require.Error(t, act.initGenerateIdentity(ctx, crypto, "foo bar", "foo.bar@example.org"))
	buf.Reset()

	act.printRecipients(ctx, "")
	assert.Contains(t, buf.String(), "0xDEADBEEF")
	buf.Reset()

	// un-initialize the store
	require.NoError(t, os.Remove(filepath.Join(u.StoreDir(""), plain.IDFile)))
	require.Error(t, act.IsInitialized(c))
	buf.Reset()
}

func TestInitParseContext(t *testing.T) {
	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	for _, tc := range []struct {
		name  string
		flags map[string]string
		check func(context.Context) error
	}{
		{
			name:  "crypto age",
			flags: map[string]string{"crypto": "age"},
			check: func(ctx context.Context) error {
				if be := backend.GetCryptoBackend(ctx); be != backend.Age {
					return fmt.Errorf("wrong backend: %d", be)
				}

				return nil
			},
		},
		{
			name: "default",
			check: func(ctx context.Context) error {
				if backend.GetStorageBackend(ctx) != backend.GitFS {
					return fmt.Errorf("wrong backend")
				}

				return nil
			},
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			c := gptest.CliCtxWithFlags(config.NewContextReadOnly(), t, tc.flags)
			require.NoError(t, tc.check(initParseContext(c.Context, c)), tc.name)
			buf.Reset()
		})
	}
}
