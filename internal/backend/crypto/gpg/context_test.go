package gpg

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
)

func TestAlwaysTrust(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())

	if IsAlwaysTrust(ctx) {
		t.Errorf("AlwaysTrust should be false")
	}

	if !IsAlwaysTrust(WithAlwaysTrust(ctx, true)) {
		t.Errorf("AlwaysTrust should be true")
	}
}
