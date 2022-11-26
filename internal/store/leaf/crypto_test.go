package leaf

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGPG(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tempdir := t.TempDir()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	assert.NoError(t, s.ImportMissingPublicKeys(ctx))

	newRecp := "A3683834"
	err = s.AddRecipient(ctx, newRecp)
	assert.NoError(t, err)

	assert.NoError(t, s.ImportMissingPublicKeys(ctx))
}
