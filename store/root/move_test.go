package root

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestMove(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	assert.NoError(t, err)

	assert.NoError(t, rs.Copy(ctx, "foo", "bar/zab"))
	assert.NoError(t, rs.Move(ctx, "foo", "bar/zab2"))
	assert.Error(t, rs.Delete(ctx, "foo"))
	assert.NoError(t, rs.Copy(ctx, "bar/zab", "foo"))
	assert.NoError(t, rs.Delete(ctx, "foo"))
	assert.NoError(t, rs.Prune(ctx, "bar"))
}
