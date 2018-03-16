package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/ctxutil"
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

	assert.NotNil(t, s.RCS())
	assert.Equal(t, "noop", s.RCS().Name())
	assert.NoError(t, s.RCS().InitConfig(ctx, "foo", "bar@baz.com"))
	assert.Equal(t, semver.Version{}, s.RCS().Version(ctx))
	assert.NoError(t, s.RCS().AddRemote(ctx, "foo", "bar"))
	assert.NoError(t, s.RCS().Pull(ctx, "origin", "master"))
	assert.NoError(t, s.RCS().Push(ctx, "origin", "master"))

	assert.NoError(t, s.GitInit(ctx, "", ""))
	assert.NoError(t, s.GitInit(backend.WithRCSBackend(ctx, backend.Noop), "", ""))
	assert.NoError(t, s.GitInit(backend.WithRCSBackend(ctx, backend.GoGit), "", ""))
	assert.Error(t, s.GitInit(backend.WithRCSBackend(ctx, -1), "", ""))

	ctx = ctxutil.WithDebug(ctx, true)
	assert.NoError(t, s.GitInit(backend.WithRCSBackend(ctx, backend.GitCLI), "Foo Bar", "foo.bar@example.org"))
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

	assert.NotNil(t, s.RCS())
	assert.Equal(t, "noop", s.RCS().Name())
	assert.NoError(t, s.RCS().InitConfig(ctx, "foo", "bar@baz.com"))

	_, err = s.ListRevisions(ctx, "foo")
	assert.Error(t, err)

	sec, err := s.GetRevision(ctx, "foo", "bar")
	assert.NoError(t, err)
	assert.Equal(t, "", sec.Password())
}
