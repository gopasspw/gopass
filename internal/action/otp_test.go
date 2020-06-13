package action

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secret"

	"github.com/gokyle/twofactor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOTP(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	t.Run("display non-otp secret", func(t *testing.T) {
		defer buf.Reset()
		assert.Error(t, act.OTP(gptest.CliCtx(ctx, t, "foo")))
	})

	t.Run("create and display valid OTP", func(t *testing.T) {
		defer buf.Reset()
		sec := secret.New()
		sec.Set("password", "foo")
		sec.WriteString(twofactor.GenerateGoogleTOTP().URL("foo"))
		assert.NoError(t, act.Store.Set(ctx, "bar", sec))

		assert.NoError(t, act.OTP(gptest.CliCtx(ctx, t, "bar")))
	})

	t.Run("copy to clipboard", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.otp(ctx, "bar", "", true, false, false))
	})

	t.Run("write QR file", func(t *testing.T) {
		defer buf.Reset()
		fn := filepath.Join(u.Dir, "qr.png")
		assert.NoError(t, act.OTP(gptest.CliCtxWithFlags(ctx, t, map[string]string{"qr": fn}, "bar")))
		assert.FileExists(t, fn)
	})
}
