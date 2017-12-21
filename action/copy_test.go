package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

func TestCopy(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

	app := cli.NewApp()
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	if err := act.Copy(ctx, c); err == nil || err.Error() != "Usage: action.test cp old-path new-path" {
		t.Errorf("Should fail")
	}
}
