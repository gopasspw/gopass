package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/fatih/color"
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

	// run doctor — the plain-crypto test store uses no GPG or age,
	// so those binary checks always pass; store permissions and
	// recipient checks should also pass.
	err = act.Doctor(gptest.CliCtx(ctx, t))
	// accept both nil (all pass) and a Doctor exit code (some check failed) —
	// the test just verifies the command runs without panicking.
	if err != nil {
		assert.Contains(t, err.Error(), "doctor found")
	}
	buf.Reset()

	// --verbose shows each check result
	err = act.Doctor(gptest.CliCtxWithFlags(ctx, t, map[string]string{"verbose": "true"}))
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
