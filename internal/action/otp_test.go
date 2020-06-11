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

	// display non-otp secret
	assert.Error(t, act.OTP(gptest.CliCtx(ctx, t, "foo")))
	buf.Reset()

	// create and display valid OTP
	sec := secret.New()
	sec.Set("password", "foo")
	sec.WriteString(twofactor.GenerateGoogleTOTP().URL("foo"))
	assert.NoError(t, act.Store.Set(ctx, "bar", sec))

	assert.NoError(t, act.OTP(gptest.CliCtx(ctx, t, "bar")))
	buf.Reset()

	// copy to clipboard
	assert.NoError(t, act.otp(ctx, "bar", "", true, false, false))
	buf.Reset()

	// write QR file
	fn := filepath.Join(u.Dir, "qr.png")
	assert.NoError(t, act.OTP(gptest.CliCtxWithFlags(ctx, t, map[string]string{"qr": fn}, "bar")))
	assert.FileExists(t, fn)
	buf.Reset()
}
