package root

import (
	"context"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	es, err := rs.List(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{"foo"}, es)

	sd, err := rs.HasSubDirs(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, false, sd)

	str, err := rs.Format(ctx, -1)
	assert.NoError(t, err)
	assert.Equal(t, `gopass
└── foo
`, str)
}
