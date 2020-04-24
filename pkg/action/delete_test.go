package action

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli"
)

func TestDelete(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()

	// delete
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)

	actName := "action.test"

	if runtime.GOOS == "windows" {
		actName = "action.test.exe"
	}

	if err := act.Delete(ctx, c); err == nil || err.Error() != fmt.Sprintf("Usage: %s rm name", actName) {
		t.Errorf("Should fail")
	}
	buf.Reset()

	// delete foo
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Delete(ctx, c))
	buf.Reset()

	// delete foo bar
	assert.NoError(t, act.Store.Set(ctx, "foo", secret.New("123", "---\nbar: zab")))
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo", "bar"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Delete(ctx, c))
	buf.Reset()

	// delete -r foo
	assert.NoError(t, act.Store.Set(ctx, "foo", secret.New("123", "---\nbar: zab")))
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	sf := cli.BoolFlag{
		Name:  "recursive",
		Usage: "recursive",
	}
	assert.NoError(t, sf.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--recursive=true", "foo"}))
	c = cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Delete(ctx, c))
	buf.Reset()

}
