package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCopy(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	app := cli.NewApp()
	// copy foo bar
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"foo", "bar"}))
	c := cli.NewContext(app, fs, nil)

	assert.NoError(t, act.Copy(ctx, c))
	buf.Reset()

	// copy not-found still-not-there
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"not-found", "still-not-there"}))
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Copy(ctx, c))
	buf.Reset()

	// copy
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	c = cli.NewContext(app, fs, nil)

	assert.Error(t, act.Copy(ctx, c))
	buf.Reset()
}
