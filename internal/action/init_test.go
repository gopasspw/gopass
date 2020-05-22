package action

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithDebug(ctx, false)
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

	c := clictx(ctx, t, "foo.bar@example.org")
	assert.NoError(t, act.Initialized(c))
	assert.Error(t, act.Init(c))
	assert.Error(t, act.InitOnboarding(c))

	crypto := act.Store.Crypto(ctx, "")
	assert.Equal(t, true, act.initHasUseablePrivateKeys(ctx, crypto, ""))
	assert.Error(t, act.initCreatePrivateKey(ctx, crypto, "", "foo bar", "foo.bar@example.org"))
	buf.Reset()

	act.printRecipients(ctx, "")
	assert.Contains(t, buf.String(), "0xDEADBEEF")
	buf.Reset()

	// un-initialize the store
	assert.NoError(t, os.Remove(filepath.Join(u.StoreDir(""), ".gpg-id")))
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
			name:  "crypto xc",
			flags: map[string]string{"crypto": "xc"},
			check: func(ctx context.Context) error {
				if backend.GetCryptoBackend(ctx) != backend.XC {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
		{
			name:  "rcs noop",
			flags: map[string]string{"rcs": "noop"},
			check: func(ctx context.Context) error {
				if backend.GetRCSBackend(ctx) != backend.Noop {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
		{
			name:  "nogit",
			flags: map[string]string{"nogit": "true"},
			check: func(ctx context.Context) error {
				if backend.GetRCSBackend(ctx) != backend.Noop {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
		{
			name: "default",
			check: func(ctx context.Context) error {
				if backend.GetRCSBackend(ctx) != backend.GitCLI {
					return fmt.Errorf("wrong backend")
				}
				return nil
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			c := clictxf(context.Background(), t, tc.flags)
			assert.NoError(t, tc.check(initParseContext(context.Background(), c)), tc.name)
			buf.Reset()
		})
	}
}
