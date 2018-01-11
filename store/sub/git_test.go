package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGit(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	assert.NoError(t, s.Git(ctx, "status"))
	assert.NoError(t, s.GitPush(ctx, "origin", "master"))

	t.Skip("flaky")
	assert.NoError(t, s.GitInit(ctx, "", "", ""))
	assert.NoError(t, s.Git(ctx, "status"))
}
