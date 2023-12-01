package root

import (
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	es, err := rs.List(ctx, tree.INF)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo"}, es)

	sd, err := rs.HasSubDirs(ctx, "foo")
	require.NoError(t, err)
	assert.False(t, sd)

	str, err := rs.Format(ctx, -1)
	require.NoError(t, err)
	assert.Equal(t, `gopass
└── foo
`, str)
}
