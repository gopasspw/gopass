package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/secrets"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
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

	// delete
	c := gptest.CliCtx(ctx, t)
	assert.Error(t, act.Delete(c))
	buf.Reset()

	// delete foo
	c = gptest.CliCtx(ctx, t, "foo")
	assert.NoError(t, act.Delete(c))
	buf.Reset()

	// delete foo bar
	sec := &secrets.Plain{}
	sec.SetPassword("123")
	sec.WriteString("---\nbar: zab")
	assert.NoError(t, act.Store.Set(ctx, "foo", sec))

	c = gptest.CliCtx(ctx, t, "foo", "bar")
	assert.NoError(t, act.Delete(c))
	buf.Reset()

	// delete -r foo
	assert.NoError(t, act.Store.Set(ctx, "foo", sec))

	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"recursive": "true"}, "foo")
	assert.NoError(t, act.Delete(c))
	buf.Reset()
}
