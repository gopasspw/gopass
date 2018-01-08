package root

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestMove(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)

	rs, err := createRootStore(ctx, tempdir)
	assert.NoError(t, err)

	assert.NoError(t, rs.Copy(ctx, "foo/bar/baz", "foo/bar/zab"))
	assert.NoError(t, rs.Move(ctx, "foo/bar/zab", "foo/bar/zab2"))
	assert.NoError(t, rs.Delete(ctx, "foo/bar/baz"))
	assert.NoError(t, rs.Prune(ctx, "foo"))
}
