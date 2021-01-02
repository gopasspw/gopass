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
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)
	ctx = backend.WithStorageBackend(ctx, backend.FS)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	c := gptest.CliCtx(ctx, t, "foo.bar@example.org")
	assert.NoError(t, act.Initialized(c))
	assert.Error(t, act.Init(c))
	assert.NoError(t, act.InitOnboarding(c))

	crypto := act.Store.Crypto(ctx, "")
	require.NotNil(t, crypto)
	assert.Equal(t, "plain", crypto.Name())
	assert.True(t, act.initHasUseablePrivateKeys(ctx, crypto))
	assert.Error(t, act.initGenerateIdentity(ctx, crypto, "foo bar", "foo.bar@example.org"))
	buf.Reset()

	act.printRecipients(ctx, "")
	assert.Contains(t, buf.String(), "0xDEADBEEF")
	buf.Reset()

	// un-initialize the store
	assert.NoError(t, os.Remove(filepath.Join(u.StoreDir(""), plain.IDFile)))
	assert.Error(t, act.Initialized(c))
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
			c := gptest.CliCtxWithFlags(context.Background(), t, tc.flags)
			assert.NoError(t, tc.check(initParseContext(c.Context, c)), tc.name)
			buf.Reset()
		})
	}
}
