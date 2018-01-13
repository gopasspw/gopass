package main

import (
	"context"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/backend/crypto/gpg"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestInitContext(t *testing.T) {
	ctx := context.Background()
	cfg := config.New()

	ctx = initContext(ctx, cfg)
	assert.Equal(t, true, gpg.IsAlwaysTrust(ctx))

	assert.NoError(t, os.Setenv("GOPASS_DEBUG", "true"))
	ctx = initContext(ctx, cfg)
	assert.Equal(t, true, ctxutil.IsDebug(ctx))

	assert.NoError(t, os.Setenv("GOPASS_NOCOLOR", "true"))
	ctx = initContext(ctx, cfg)
	assert.Equal(t, false, ctxutil.IsColor(ctx))
}
