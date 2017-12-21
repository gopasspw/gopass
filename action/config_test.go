package action

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/urfave/cli"
)

func TestConfig(t *testing.T) {
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

	// action.Config
	if err := act.Config(ctx, c); err != nil {
		t.Errorf("Error: %s", err)
	}
	want := `root store config:
  askformore: false
  autoimport: true
  autosync: true
  cliptimeout: 45
  nocolor: false
  noconfirm: false
  nopager: false
`
	want += "  path: " + filepath.Join(td, "store") + "\n"
	want += `  safecontent: false
  usesymbols: false
`
	if buf.String() != want {
		t.Errorf("'%s' != '%s'", buf.String(), want)
	}
	buf.Reset()

	// action.setConfigValue
	if err := act.setConfigValue(ctx, "", "nopager", "true"); err != nil {
		t.Errorf("Error: %s", err)
	}
	sv := strings.TrimSpace(buf.String())
	want = "nopager: true"
	if sv != want {
		t.Errorf("Wrong config echo: '%s' != '%s'", sv, want)
	}
	buf.Reset()

	// action.printConfigValues
	if err := act.printConfigValues(ctx, "", "nopager"); err != nil {
		t.Errorf("Error: %s", err)
	}
	sv = strings.TrimSpace(buf.String())
	if sv != want {
		t.Errorf("Wrong config result: '%s' != '%s'", sv, want)
	}
	buf.Reset()
}
