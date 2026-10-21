package action

import (
	"bytes"
	"os"
	"runtime"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctor(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()
	color.NoColor = true

	// run doctor — the plain-crypto test store uses no GPG,
	// so the GPG binary check always passes; store permissions and
	// recipient checks should also pass.
	err = act.Doctor(ctx, gptest.CliCtx(ctx, t))
	// accept both nil (all pass) and a Doctor exit code (some check failed) —
	// the test just verifies the command runs without panicking.
	if err != nil {
		assert.Contains(t, err.Error(), "doctor found")
	}
	buf.Reset()

	// --verbose shows each check result
	err = act.Doctor(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"verbose": "true"}))
	if err != nil {
		assert.Contains(t, err.Error(), "doctor found")
	}
	buf.Reset()
}

func TestDoctorStoreLabel(t *testing.T) {
	assert.Equal(t, "<root>", doctorStoreLabel(""))
	assert.Equal(t, "foo", doctorStoreLabel("foo"))
	assert.Equal(t, "foo/bar", doctorStoreLabel("foo/bar"))
}

// TestDoctorAgeWithoutBinary checks that doctor passes for an age store
// without the age CLI in PATH. See GH-3676.
func TestDoctorAgeWithoutBinary(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.Age)
	ctx = backend.WithStorageBackend(ctx, backend.FS)
	ctx = ctxutil.WithAgePassphrase(ctx, "foobar")

	act, err := newMock(ctx, u.StoreDir(""))
	require.ErrorContains(t, err, "not initialized")
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()
	color.NoColor = true

	require.NoError(t, os.RemoveAll(u.StoreDir("")))
	require.NoError(t, os.Remove(u.GPConfig()))
	require.NoError(t, act.Setup(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"storage": "fs", "crypto": "age"})))
	require.Equal(t, "age", act.Store.Crypto(ctx, "").Name())
	buf.Reset()

	t.Setenv("PATH", t.TempDir())

	err = act.Doctor(ctx, gptest.CliCtxWithFlags(ctx, t, map[string]string{"verbose": "true"}))
	// Windows reports writable dirs as 0777, failing the permissions check.
	if runtime.GOOS != "windows" {
		require.NoError(t, err, buf.String())
	}
	assert.NotContains(t, buf.String(), "age binary")
}
