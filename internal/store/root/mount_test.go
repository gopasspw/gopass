package root

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMount(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	assert.Equal(t, map[string]string{}, rs.Mounts())
	assert.Equal(t, []string{}, rs.MountPoints())

	sub, err := rs.GetSubStore("")
	require.NoError(t, err)
	require.NotNil(t, sub)

	sub, err = rs.GetSubStore("foo")
	assert.Error(t, err)
	assert.Nil(t, sub)

	// removing mounts should never fail
	assert.NoError(t, rs.RemoveMount(ctx, "foo"))
}

func TestMountPoint(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	assert.NoError(t, u.InitStore("sub1"))
	assert.NoError(t, u.InitStore("sub2"))
	assert.NoError(t, rs.AddMount(ctx, "sub1", u.StoreDir("sub1")))
	assert.NoError(t, rs.AddMount(ctx, "sub2", u.StoreDir("sub2")))

	assert.Equal(t, "sub1", rs.MountPoint("sub1"))
}
