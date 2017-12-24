package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

func TestEdit(t *testing.T) {
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

	if err := act.Edit(ctx, c); err == nil || err.Error() != "Usage: action.test edit secret" {
		t.Errorf("Should fail")
	}
}

func TestEditor(t *testing.T) {
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

	touch, err := exec.LookPath("touch")
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	want := "foobar"
	out, err := act.editor(ctx, touch, []byte(want))
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if string(out) != want {
		t.Errorf("'%s' != '%s'", string(out), want)
	}
}
