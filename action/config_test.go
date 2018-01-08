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

	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestConfig(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	assert.NoError(t, err)

	app := cli.NewApp()
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	// action.Config
	assert.NoError(t, act.Config(ctx, c))
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
	assert.NoError(t, act.setConfigValue(ctx, "", "nopager", "true"))
	sv := strings.TrimSpace(buf.String())
	want = "nopager: true"
	if sv != want {
		t.Errorf("Wrong config echo: '%s' != '%s'", sv, want)
	}
	buf.Reset()

	// action.printConfigValues
	act.cfg.Mounts["foo"] = &config.StoreConfig{}
	act.printConfigValues(ctx, "", "nopager")
	want = `nopager: true
foo/nopager: false`
	sv = strings.TrimSpace(buf.String())
	if sv != want {
		t.Errorf("Wrong config result: '%s' != '%s'", sv, want)
	}

	delete(act.cfg.Mounts, "foo")
	buf.Reset()

	// config autoimport
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"autoimport"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, act.Config(ctx, c))
	want = `autoimport: true`
	sv = strings.TrimSpace(buf.String())
	if sv != want {
		t.Errorf("Wrong config result: '%s' != '%s'", sv, want)
	}
	buf.Reset()

	// config autoimport false
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"autoimport", "false"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, act.Config(ctx, c))
	want = `autoimport: false`
	sv = strings.TrimSpace(buf.String())
	if sv != want {
		t.Errorf("Wrong config result: '%s' != '%s'", sv, want)
	}
	buf.Reset()

	// action.ConfigComplete
	act.ConfigComplete(c)
	want = `askformore
autoimport
autosync
cliptimeout
nocolor
noconfirm
nopager
path
safecontent
usesymbols
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// config autoimport false 42
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"autoimport", "false", "42"}))
	c = cli.NewContext(app, fs, nil)
	assert.Error(t, act.Config(ctx, c))
	buf.Reset()
}
