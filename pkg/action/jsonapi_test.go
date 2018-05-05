package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestJSONAPI(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	act, err := newMock(ctx, u)
	assert.NoError(t, err)

	app := cli.NewApp()
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	assert.NoError(t, act.JSONAPI(ctx, c))
}

func TestStringInSlice(t *testing.T) {
	if stringInSlice("foo", []string{"bar", "baz"}) {
		t.Errorf("Should not contain foo")
	}
	if !stringInSlice("foo", []string{"foo", "bar", "baz"}) {
		t.Errorf("Should contain foo")
	}
}
