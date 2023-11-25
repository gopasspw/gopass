package root

import (
	"context"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRCS(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewNoWrites().WithConfig(context.Background())
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	require.Error(t, rs.RCSStatus(ctx, ""))

	revs, err := rs.ListRevisions(ctx, "foo")
	require.NoError(t, err)
	assert.Len(t, revs, 1)
}
