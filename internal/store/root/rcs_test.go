package root

import (
	"context"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRCS(t *testing.T) {
	t.Parallel()

	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	assert.Error(t, rs.RCSStatus(ctx, ""))

	revs, err := rs.ListRevisions(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(revs))
}
