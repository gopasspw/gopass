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
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	// print mounts
	assert.NoError(t, act.MountsPrint(clictx(ctx, t)))
	buf.Reset()

	// complete mounts
	act.MountsComplete(clictx(ctx, t))
	assert.Equal(t, buf.String(), "")
	buf.Reset()

	// remove no non-existing mount
	assert.Error(t, act.MountRemove(clictx(ctx, t)))
	buf.Reset()

	// remove non-existing mount
	assert.NoError(t, act.MountRemove(clictx(ctx, t, "foo")))
	buf.Reset()

	// add non-existing mount
	assert.Error(t, act.MountAdd(clictx(ctx, t, "foo", filepath.Join(u.Dir, "mount1"))))
	buf.Reset()

	// add some mounts
	assert.NoError(t, u.InitStore("mount1"))
	assert.NoError(t, u.InitStore("mount2"))
	assert.NoError(t, act.Store.AddMount(ctx, "mount1", u.StoreDir("mount1")))
	assert.NoError(t, act.Store.AddMount(ctx, "mount2", u.StoreDir("mount2")))

	// print mounts
	assert.NoError(t, act.MountsPrint(clictx(ctx, t)))
	buf.Reset()
}
