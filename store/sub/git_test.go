package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend"
	"github.com/justwatchcom/gopass/utils/ctxutil"
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

	assert.NotNil(t, s.Sync())
	assert.Equal(t, "git-mock", s.Sync().Name())
	assert.NoError(t, s.GitInitConfig(ctx, "foo", "bar@baz.com"))
	assert.Equal(t, semver.Version{}, s.GitVersion(ctx))
	assert.NoError(t, s.GitAddRemote(ctx, "foo", "bar"))
	assert.NoError(t, s.GitPull(ctx, "origin", "master"))
	assert.NoError(t, s.GitPush(ctx, "origin", "master"))

	assert.NoError(t, s.GitInit(ctx, "", ""))
	assert.NoError(t, s.GitInit(backend.WithSyncBackend(ctx, backend.GitMock), "", ""))
	assert.NoError(t, s.GitInit(backend.WithSyncBackend(ctx, backend.GoGit), "", ""))
	assert.Error(t, s.GitInit(backend.WithSyncBackend(ctx, -1), "", ""))

	ctx = ctxutil.WithDebug(ctx, true)
	assert.NoError(t, s.GitInit(backend.WithSyncBackend(ctx, backend.GitCLI), "Foo Bar", "foo.bar@example.org"))
}

func TestGitRevisions(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	assert.NotNil(t, s.Sync())
	assert.Equal(t, "git-mock", s.Sync().Name())
	assert.NoError(t, s.GitInitConfig(ctx, "foo", "bar@baz.com"))

	_, err = s.ListRevisions(ctx, "foo")
	assert.Error(t, err)

	sec, err := s.GetRevision(ctx, "foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "", sec.Password())
}
