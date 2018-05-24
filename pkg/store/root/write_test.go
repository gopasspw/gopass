package root

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store/secret"
	"github.com/justwatchcom/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, u)
	assert.NoError(t, err)

	assert.NoError(t, rs.Set(ctx, "zab", secret.New("foo", "bar")))
}
