package action

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"

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

	t.Run("display empty template tree", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.TemplatesPrint(gptest.CliCtx(ctx, t, "foo")))
		assert.Equal(t, "gopass\n\n", buf.String())
	})

	t.Run("add template", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.Store.SetTemplate(ctx, "foo", []byte("foobar")))
		assert.NoError(t, act.TemplatesPrint(gptest.CliCtx(ctx, t, "foo")))
		want := `Pushed changes to git remote
gopass
└── foo

`
		assert.Equal(t, want, buf.String())
	})

	t.Run("complete templates", func(t *testing.T) {
		defer buf.Reset()
		act.TemplatesComplete(gptest.CliCtx(ctx, t, "foo"))
		assert.Equal(t, "foo\n", buf.String())
	})

	t.Run("print template", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.TemplatePrint(gptest.CliCtx(ctx, t, "foo")))
		assert.Equal(t, "foobar\n", buf.String())
	})

	t.Run("edit template", func(t *testing.T) {
		defer buf.Reset()
		assert.Error(t, act.TemplateEdit(gptest.CliCtx(ctx, t, "foo")))
	})

	t.Run("remove template", func(t *testing.T) {
		defer buf.Reset()
		assert.NoError(t, act.TemplateRemove(gptest.CliCtx(ctx, t, "foo")))
	})
}
