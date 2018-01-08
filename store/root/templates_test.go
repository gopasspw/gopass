package root

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, tempdir)
	assert.NoError(t, err)

	tt, err := rs.TemplateTree(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "gopass\n", tt.Format(0))

	assert.Equal(t, false, rs.HasTemplate(ctx, "foo"))
	_, err = rs.GetTemplate(ctx, "foo")
	assert.Error(t, err)
	assert.Error(t, rs.RemoveTemplate(ctx, "foo"))

	assert.NoError(t, rs.SetTemplate(ctx, "foo", []byte("foobar")))
	assert.Equal(t, true, rs.HasTemplate(ctx, "foo"))
	b, err := rs.GetTemplate(ctx, "foo")
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(b))
	b, found := rs.LookupTemplate(ctx, "foo/bar")
	assert.Equal(t, true, found)
	assert.Equal(t, "foobar", string(b))
	assert.NoError(t, rs.RemoveTemplate(ctx, "foo"))
}
