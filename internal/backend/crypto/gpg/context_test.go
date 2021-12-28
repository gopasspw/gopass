package gpg

import (
	"context"
	"testing"
)

// nolint:ifshort
func TestAlwaysTrust(t *testing.T) {
	ctx := context.Background()

	if IsAlwaysTrust(ctx) {
		t.Errorf("AlwaysTrust should be false")
	}

	if !IsAlwaysTrust(WithAlwaysTrust(ctx, true)) {
		t.Errorf("AlwaysTrust should be true")
	}
}
