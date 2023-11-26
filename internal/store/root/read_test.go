package root

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	_, err = rs.Get(ctx, "foo")
	require.NoError(t, err)
}
