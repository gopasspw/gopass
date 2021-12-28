package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	// git init
	c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"username": "foobar", "useremail": "foo.bar@example.org"})
	assert.NoError(t, act.RCSInit(c))
	buf.Reset()

	// getUserData
	name, email := act.getUserData(ctx, "", "", "")
	assert.Equal(t, "0xDEADBEEF", name)
	assert.Equal(t, "0xDEADBEEF", email)

	// GitAddRemote
	assert.Error(t, act.RCSAddRemote(c))
	buf.Reset()

	// GitRemoveRemote
	assert.Error(t, act.RCSRemoveRemote(c))
	buf.Reset()

	// GitPull
	assert.Error(t, act.RCSPull(c))
	buf.Reset()

	// GitPush
	assert.Error(t, act.RCSPush(c))
	buf.Reset()
}
