package action

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMounts(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
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
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	t.Run("print mounts", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.MountsPrint(gptest.CliCtx(ctx, t)))
	})

	t.Run("complete mounts", func(t *testing.T) {
		defer buf.Reset()
		act.MountsComplete(gptest.CliCtx(ctx, t))
		assert.Equal(t, "", buf.String())
	})

	t.Run("remove no non-existing mount", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.MountRemove(gptest.CliCtx(ctx, t)))
	})

	t.Run("remove non-existing mount", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.MountRemove(gptest.CliCtx(ctx, t, "foo")))
	})

	t.Run("add non-existing mount", func(t *testing.T) {
		defer buf.Reset()
		require.Error(t, act.MountAdd(gptest.CliCtx(ctx, t, "foo", filepath.Join(u.Dir, "mount1"))))
	})

	t.Run("add some mounts", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, u.InitStore("mount1"))
		require.NoError(t, u.InitStore("mount2"))
		require.NoError(t, act.Store.AddMount(ctx, "mount1", u.StoreDir("mount1")))
		require.NoError(t, act.Store.AddMount(ctx, "mount2", u.StoreDir("mount2")))
	})

	t.Run("print mounts", func(t *testing.T) {
		defer buf.Reset()
		require.NoError(t, act.MountsPrint(gptest.CliCtx(ctx, t)))
	})
}
