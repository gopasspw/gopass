package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/store/secret"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestDelete(t *testing.T) {
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

	// delete
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)

	if err := act.Delete(ctx, c); err == nil || err.Error() != "Usage: action.test rm name" {
		t.Errorf("Should fail")
	}
	buf.Reset()

	// delete foo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"foo"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Delete(ctx, c))
	buf.Reset()

	// delete foo bar
	assert.NoError(t, act.Store.Set(ctx, "foo", secret.New("123", "---\nbar: zab")))
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	if err := fs.Parse([]string{"foo", "bar"}); err != nil {
		t.Fatalf("Error: %s", err)
	}
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Delete(ctx, c))
	buf.Reset()
}
