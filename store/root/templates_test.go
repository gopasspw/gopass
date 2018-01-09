package root

import (
	"context"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, u)
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
