package action

import (
	"bytes"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
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
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	// git init
	c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"name": "foobar", "email": "foo.bar@example.org"})
	require.NoError(t, act.RCSInit(c))
	buf.Reset()

	// getUserData
	name, email := act.getUserData(ctx, "", "", "")
	assert.Equal(t, "0xDEADBEEF", name)
	assert.Equal(t, "0xDEADBEEF", email)

	// GitAddRemote
	require.Error(t, act.RCSAddRemote(c))
	buf.Reset()

	// GitRemoveRemote
	require.Error(t, act.RCSRemoveRemote(c))
	buf.Reset()

	// GitPull
	require.Error(t, act.RCSPull(c))
	buf.Reset()

	// GitPush
	require.NoError(t, act.RCSPush(c))
	buf.Reset()
}
