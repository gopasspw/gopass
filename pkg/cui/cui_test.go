package cui

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
)

func TestGetSelection(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)

	act, sel := GetSelection(ctx, "foo", "bar", []string{"foo", "bar"})
	assert.Equal(t, "impossible", act)
	assert.Equal(t, 0, sel)
}
