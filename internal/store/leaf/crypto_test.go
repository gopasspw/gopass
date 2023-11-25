package leaf

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/stretchr/testify/require"
)

func TestGPG(t *testing.T) {
	ctx := config.NewNoWrites().WithConfig(context.Background())

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	s, err := createSubStore(t)
	require.NoError(t, err)

	require.NoError(t, s.ImportMissingPublicKeys(ctx))

	newRecp := "A3683834"
	err = s.AddRecipient(ctx, newRecp)
	require.NoError(t, err)

	require.NoError(t, s.ImportMissingPublicKeys(ctx))
}
