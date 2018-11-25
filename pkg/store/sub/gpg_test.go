package sub

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/out"

	"github.com/muesli/goprogressbar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGPG(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	goprogressbar.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
		goprogressbar.Stdout = os.Stdout
	}()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	assert.NoError(t, s.ImportMissingPublicKeys(ctx))

	newRecp := "A3683834"
	err = s.AddRecipient(ctx, newRecp)
	assert.NoError(t, err)

	assert.NoError(t, s.ImportMissingPublicKeys(ctx))
}
