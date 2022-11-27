package action

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gokyle/twofactor"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOTP(t *testing.T) { //nolint:paralleltest
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	t.Run("display non-otp secret", func(t *testing.T) { //nolint:paralleltest
		defer buf.Reset()
		assert.Error(t, act.OTP(gptest.CliCtx(ctx, t, "foo")))
	})

	t.Run("create and display valid OTP", func(t *testing.T) { //nolint:paralleltest
		defer buf.Reset()
		sec := secrets.NewAKV()
		sec.SetPassword("foo")
		_, err := sec.Write([]byte(twofactor.GenerateGoogleTOTP().URL("foo")))
		require.NoError(t, err)
		assert.NoError(t, act.Store.Set(ctx, "bar", sec))

		t.Logf("Secret: %q\n", string(sec.Bytes()))

		assert.NoError(t, act.OTP(gptest.CliCtx(ctx, t, "bar")))
	})

	t.Run("copy to clipboard", func(t *testing.T) { //nolint:paralleltest
		defer buf.Reset()
		assert.NoError(t, act.otp(ctx, "bar", "", true, false, false))
	})

	t.Run("write QR file", func(t *testing.T) { //nolint:paralleltest
		defer buf.Reset()
		fn := filepath.Join(u.Dir, "qr.png")
		assert.NoError(t, act.OTP(gptest.CliCtxWithFlags(ctx, t, map[string]string{"qr": fn}, "bar")))
		assert.FileExists(t, fn)
	})
}
