package gpg

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
)

func TestAlwaysTrust(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()

	if IsAlwaysTrust(ctx) {
		t.Errorf("AlwaysTrust should be false")
	}

	if !IsAlwaysTrust(WithAlwaysTrust(ctx, true)) {
		t.Errorf("AlwaysTrust should be true")
	}
}
