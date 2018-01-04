package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestTemplates(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// display empty template tree
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c := cli.NewContext(app, fs, nil)

	out := capture(t, func() error {
		return act.TemplatesPrint(ctx, c)
	})
	want := `gopass`
	if out != want {
		t.Errorf("'%s' != '%s'", want, out)
	}
	buf.Reset()

	// add template
	if err := act.Store.SetTemplate(ctx, "foo", []byte("foobar")); err != nil {
		t.Errorf("Failed to add template: %s", err)
	}
	out = capture(t, func() error {
		return act.TemplatesPrint(ctx, c)
	})
	want = `gopass
└── foo`
	if out != want {
		t.Errorf("'%s' != '%s'", want, out)
	}
	buf.Reset()

	// complete templates
	out = capture(t, func() error {
		act.TemplatesComplete(c)
		return nil
	})
	assert.Equal(t, out, "foo")
	buf.Reset()

	// remove template
	assert.NoError(t, act.TemplateRemove(ctx, c))
	buf.Reset()
}
