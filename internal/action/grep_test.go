package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secret"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGrep(t *testing.T) {
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

	c := gptest.CliCtx(ctx, t, "foo")
	assert.Error(t, act.Grep(c))
	buf.Reset()

	// add some secret
	sec := secret.New()
	sec.Set("password", "foobar")
	sec.WriteString("foobar")
	assert.NoError(t, act.Store.Set(ctx, "foo", sec))
	assert.Error(t, act.Grep(c))
	buf.Reset()
}
