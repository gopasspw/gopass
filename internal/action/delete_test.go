package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/tests/gptest"

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
	c := clictx(ctx, t)
	assert.Error(t, act.Delete(c))
	buf.Reset()

	// delete foo
	c = clictx(ctx, t, "foo")
	assert.NoError(t, act.Delete(c))
	buf.Reset()

	// delete foo bar
	assert.NoError(t, act.Store.Set(ctx, "foo", secret.New("123", "---\nbar: zab")))

	c = clictx(ctx, t, "foo", "bar")
	assert.NoError(t, act.Delete(c))
	buf.Reset()

	// delete -r foo
	assert.NoError(t, act.Store.Set(ctx, "foo", secret.New("123", "---\nbar: zab")))

	c = clictxf(ctx, t, map[string]string{"recursive": "true"}, "foo")
	assert.NoError(t, act.Delete(c))
	buf.Reset()
}
