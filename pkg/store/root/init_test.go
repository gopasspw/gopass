package root

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/config"
	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	ctx = backend.WithCryptoBackend(ctx, backend.Plain)

	cfg := config.New()
	cfg.Root.Path = backend.FromPath(u.StoreDir("rs"))
	rs, err := New(ctx, cfg)
	assert.NoError(t, err)

	assert.Equal(t, false, rs.Initialized(ctx))
	assert.NoError(t, rs.Init(ctx, "", u.StoreDir("rs"), "0xDEADBEEF"))
	assert.Equal(t, true, rs.Initialized(ctx))
	assert.NoError(t, rs.Init(ctx, "rs2", u.StoreDir("rs2"), "0xDEADBEEF"))
}
