package main

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/backend/gpg"
	"github.com/justwatchcom/gopass/config"
)

func TestInitContext(t *testing.T) {
	ctx := context.Background()
	cfg := config.New()

	ctx = initContext(ctx, cfg)

	if !gpg.IsAlwaysTrust(ctx) {
		t.Errorf("AlwaysTrust should be true")
	}
}
