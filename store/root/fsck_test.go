package root

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestFsck(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	assert.NoError(t, err)

	assert.NoError(t, u.InitStore("sub1"))
	assert.NoError(t, u.InitStore("sub2"))

	assert.NoError(t, rs.AddMount(ctx, "sub1", u.StoreDir("sub1")))
	assert.NoError(t, rs.AddMount(ctx, "sub2", u.StoreDir("sub2")))

	_, err = rs.Fsck(ctx, "")
	assert.NoError(t, err)
}
