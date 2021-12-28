package termio

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
)

func TestPromptPass(t *testing.T) {
	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	_, err := promptPass(ctx, "foo")
	assert.NoError(t, err)
}
