package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestTemplates(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	color.NoColor = true
	defer func() {
		stdout = os.Stdout
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// display empty template tree
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.TemplatesPrint(ctx, c))
	assert.Equal(t, "gopass\n\n", buf.String())
	buf.Reset()

	// add template
	assert.NoError(t, act.Store.SetTemplate(ctx, "foo", []byte("foobar")))
	assert.NoError(t, act.TemplatesPrint(ctx, c))
	want := `gopass
└── foo

`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// complete templates
	act.TemplatesComplete(ctx, c)
	assert.Equal(t, "foo\n", buf.String())
	buf.Reset()

	// print template
	assert.NoError(t, act.TemplatePrint(ctx, c))
	assert.Equal(t, "foobar\n", buf.String())

	// remove template
	assert.NoError(t, act.TemplateRemove(ctx, c))
	buf.Reset()
}
