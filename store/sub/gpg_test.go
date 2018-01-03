package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGPG(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	sv := s.GPGVersion(ctx)
	t.Logf("GPG-Version: %s", sv.String())

	assert.NoError(t, s.ImportMissingPublicKeys(ctx))
}
