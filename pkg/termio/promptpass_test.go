package termio

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/require"
)

func TestPromptPass(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	_, err := promptPass(ctx, "foo")
	require.NoError(t, err)
}
