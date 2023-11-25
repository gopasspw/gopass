package pwrules

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLookupChangeURL(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())
	assert.Equal(t, "https://account.gmx.net/ciss/security/edit/passwordChange", LookupChangeURL(ctx, "gmx.net"))
}
