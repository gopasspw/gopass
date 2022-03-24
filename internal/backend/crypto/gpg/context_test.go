package gpg

import (
	"context"
	"testing"
)

func TestAlwaysTrust(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	if IsAlwaysTrust(ctx) {
		t.Errorf("AlwaysTrust should be false")
	}

	if !IsAlwaysTrust(WithAlwaysTrust(ctx, true)) {
		t.Errorf("AlwaysTrust should be true")
	}
}
