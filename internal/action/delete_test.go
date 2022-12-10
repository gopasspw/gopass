package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

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
	sec := secrets.NewAKV()
	sec.SetPassword("123")
	_, err = sec.Write([]byte("---\nbar: zab"))
	require.NoError(t, err)
	assert.NoError(t, act.Store.Set(ctx, "foo", sec))

	c = gptest.CliCtx(ctx, t, "foo", "bar")
	assert.NoError(t, act.Delete(c))
	buf.Reset()

	// delete -r foo
	assert.NoError(t, act.Store.Set(ctx, "foo", sec))

	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"recursive": "true"}, "foo")
	assert.NoError(t, act.Delete(c))
	buf.Reset()

	// reject recursive flag when a key is given
	c = gptest.CliCtxWithFlags(ctx, t, map[string]string{"recursive": "true"}, "foo", "bar")
	assert.Error(t, act.Delete(c))
	buf.Reset()

	assert.NoError(t, act.Store.Set(ctx, "sec/1", sec))
	assert.NoError(t, act.Store.Set(ctx, "sec/2", sec))
	assert.NoError(t, act.Store.Set(ctx, "sec/3", sec))
	assert.NoError(t, act.Store.Set(ctx, "sec/4", sec))

	// warn if key matching a secret is given
	c = gptest.CliCtx(ctx, t, "sec/1", "sec/2")
	assert.Error(t, act.Delete(c))
	buf.Reset()

	// remove multiple secrets
	c = gptest.CliCtx(ctx, t, "sec/1", "sec/2", "sec/3", "sec/4")
	assert.NoError(t, act.Delete(c))
	buf.Reset()
}
