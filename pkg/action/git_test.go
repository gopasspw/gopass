package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestGit(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithInteractive(ctx, false)

	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	app := cli.NewApp()

	// git init
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	un := cli.StringFlag{
		Name:  "username",
		Usage: "username",
	}
	assert.NoError(t, un.ApplyWithError(fs))
	ue := cli.StringFlag{
		Name:  "useremail",
		Usage: "useremail",
	}
	assert.NoError(t, ue.ApplyWithError(fs))
	assert.NoError(t, fs.Parse([]string{"--username", "foobar", "--useremail", "foo.bar@example.org"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.GitInit(ctx, c))
	buf.Reset()

	// getUserData
	name, email := act.getUserData(ctx, "", "", "")
	assert.Equal(t, "", name)
	assert.Equal(t, "", email)

	// GitAddRemote
	assert.Error(t, act.GitAddRemote(ctx, c))
	buf.Reset()

	// GitRemoveRemote
	assert.Error(t, act.GitRemoveRemote(ctx, c))
	buf.Reset()

	// GitPull
	assert.Error(t, act.GitPull(ctx, c))
	buf.Reset()

	// GitPush
	assert.Error(t, act.GitPush(ctx, c))
	buf.Reset()
}
