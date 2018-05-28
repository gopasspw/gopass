package termwiz

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
)

func TestGetSelection(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	act, sel := GetSelection(ctx, "foo", "bar", []string{"foo", "bar"})
	assert.Equal(t, act, "impossible")
	assert.Equal(t, sel, 0)
}
