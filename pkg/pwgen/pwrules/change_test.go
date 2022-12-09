package pwrules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupChangeURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	assert.Equal(t, "https://account.gmx.net/ciss/security/edit/passwordChange", LookupChangeURL(ctx, "gmx.net"))
}
