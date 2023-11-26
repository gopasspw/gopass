package cui

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestGetSelection(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextReadOnly()
	ctx = ctxutil.WithInteractive(ctx, false)

	act, sel := GetSelection(ctx, "foo", []string{"foo", "bar"})
	assert.Equal(t, "impossible", act)
	assert.Equal(t, 0, sel)
}
