package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"

	_ "github.com/gopasspw/gopass/internal/backend/crypto"
	_ "github.com/gopasspw/gopass/internal/backend/rcs"
	_ "github.com/gopasspw/gopass/internal/backend/storage"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)

	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	color.NoColor = true
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	// display empty template tree
	assert.NoError(t, act.TemplatesPrint(clictx(ctx, t, "foo")))
	assert.Equal(t, "gopass\n\n", buf.String())
	buf.Reset()

	// add template
	assert.NoError(t, act.Store.SetTemplate(ctx, "foo", []byte("foobar")))
	assert.NoError(t, act.TemplatesPrint(clictx(ctx, t, "foo")))
	want := `Pushed changes to git remote
gopass
└── foo

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// complete templates
	act.TemplatesComplete(clictx(ctx, t, "foo"))
	assert.Equal(t, "foo\n", buf.String())
	buf.Reset()

	// print template
	assert.NoError(t, act.TemplatePrint(clictx(ctx, t, "foo")))
	assert.Equal(t, "foobar\n", buf.String())
	buf.Reset()

	// edit template
	assert.Error(t, act.TemplateEdit(clictx(ctx, t, "foo")))
	buf.Reset()

	// remove template
	assert.NoError(t, act.TemplateRemove(clictx(ctx, t, "foo")))
	buf.Reset()
}
