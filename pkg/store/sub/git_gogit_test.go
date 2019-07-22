// +build gogit

package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoGit(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	require.NoError(t, err)

	assert.NotNil(t, s.RCS())
	assert.Equal(t, "noop", s.RCS().Name())
	assert.NoError(t, s.RCS().InitConfig(ctx, "foo", "bar@baz.com"))

	assert.NoError(t, s.GitInit(ctx, "", ""))
	assert.NoError(t, s.GitInit(backend.WithRCSBackend(ctx, backend.Noop), "", ""))
	assert.NoError(t, s.GitInit(backend.WithRCSBackend(ctx, backend.GoGit), "", ""))
	assert.Error(t, s.GitInit(backend.WithRCSBackend(ctx, -1), "", ""))

	ctx = ctxutil.WithDebug(ctx, true)
	assert.NoError(t, s.GitInit(backend.WithRCSBackend(ctx, backend.GitCLI), "Foo Bar", "foo.bar@example.org"))
}
