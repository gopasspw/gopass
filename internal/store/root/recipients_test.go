package root

import (
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipients(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, u)
	require.NoError(t, err)

	assert.Equal(t, []string{"0xDEADBEEF"}, rs.ListRecipients(ctx, ""))
	rt, err := rs.RecipientsTree(ctx, false)
	require.NoError(t, err)
	assert.Equal(t, "gopass\n└── 0xDEADBEEF\n", rt.Format(0))
}
