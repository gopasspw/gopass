package pwrules

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestLookupChangeURL(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextReadOnly()
	assert.Equal(t, "https://account.gmx.net/ciss/security/edit/passwordChange", LookupChangeURL(ctx, "gmx.net"))
}
