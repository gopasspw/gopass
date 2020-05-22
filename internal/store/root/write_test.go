package root

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	assert.NoError(t, rs.Set(ctx, "zab", secret.New("foo", "bar")))

	ctx, err = rs.SetContext(ctx, "zab2", secret.New("foo", "bar"))
	assert.NoError(t, err)
	assert.NotNil(t, ctx)
}
