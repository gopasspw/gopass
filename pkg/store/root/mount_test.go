package root

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMount(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

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

	assert.Error(t, rs.RemoveMount(ctx, "foo"))
}
