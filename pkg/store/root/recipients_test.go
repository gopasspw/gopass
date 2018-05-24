package root

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestRecipients(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, u)
	assert.NoError(t, err)

	assert.Equal(t, []string{"0xDEADBEEF"}, rs.ListRecipients(ctx, ""))
	rt, err := rs.RecipientsTree(ctx, false)
	assert.NoError(t, err)
	assert.Equal(t, "gopass\n└── 0xDEADBEEF\n", rt.Format(0))
}
