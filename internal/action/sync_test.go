package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	t.Run("default", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.Sync(gptest.CliCtx(ctx, t)))
	})

	t.Run("sync --store=root", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.Sync(gptest.CliCtxWithFlags(ctx, t, map[string]string{"store": "root"})))
	})
}
