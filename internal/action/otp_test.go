package action

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/gokyle/twofactor"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOTP(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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

	t.Run("display non-otp secret", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.OTP(gptest.CliCtx(ctx, t, "foo")))
	})

	t.Run("create and display valid OTP", func(t *testing.T) {
		defer buf.Reset()
		sec := secrets.NewAKV()
		sec.SetPassword("foo")
		_, err := sec.Write([]byte(twofactor.GenerateGoogleTOTP().URL("foo") + "\n"))
		require.NoError(t, err)
		require.NoError(t, act.Store.Set(ctx, "bar", sec))

		require.NoError(t, act.OTP(gptest.CliCtx(ctx, t, "bar")))

		// add some unrelated body material, it should still work
		_, err = sec.Write([]byte("more body content, unrelated to otp"))
		require.NoError(t, err)
		require.NoError(t, act.Store.Set(ctx, "bar", sec))

		require.NoError(t, act.OTP(gptest.CliCtx(ctx, t, "bar")))
	})

	t.Run("copy to clipboard", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.otp(ctx, "bar", "", true, false, false, false))
	})

	t.Run("copy to clipboard chained", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.otp(ctx, "bar", "", true, false, false, true))
	})

	t.Run("write QR file", func(t *testing.T) {
		defer buf.Reset()
		fn := filepath.Join(u.Dir, "qr.png")
		require.NoError(t, act.OTP(gptest.CliCtxWithFlags(ctx, t, map[string]string{"qr": fn}, "bar")))
		assert.FileExists(t, fn)
	})
}
