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
	t.Run("empty store", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.Grep(c))
	})

	t.Run("add some secret", func(t *testing.T) {
		defer buf.Reset()
		sec := secret.New()
		sec.Set("password", "foobar")
		sec.WriteString("foobar")
		assert.NoError(t, act.Store.Set(ctx, "foo", sec))
	})

	t.Run("should find existing", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.Grep(c))
	})

	t.Run("RE2", func(t *testing.T) {
		defer buf.Reset()
		c := gptest.CliCtxWithFlags(ctx, t, map[string]string{"regexp": "true"}, "f..bar")
		assert.NoError(t, act.Grep(c))
	})
}
