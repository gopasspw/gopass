package root

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/require"
)

func TestFsck(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, rs)

	require.NoError(t, rs.Fsck(ctx, "", ""))
}
