package leaf

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGPG(t *testing.T) {
	ctx := config.NewContextInMemory()

	ctx, _, _ = captureOutput(t, ctx)

	s, err := createSubStore(t)
	require.NoError(t, err)

	require.NoError(t, s.ImportMissingPublicKeys(ctx))

	newRecp := "A3683834"
	err = s.AddRecipient(ctx, newRecp)
	require.NoError(t, err)

	require.NoError(t, s.ImportMissingPublicKeys(ctx))
}
