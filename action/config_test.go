package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/tests/gptest"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
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
  cryptobackend: gpg
  nocolor: false
  noconfirm: false
  nopager: false
  notifications: true
`
	want += "  path: " + u.StoreDir("") + "\n"
	want += `  safecontent: false
  storebackend: fs
  syncbackend: git
  usesymbols: false
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// action.setConfigValue
	assert.NoError(t, act.setConfigValue(ctx, "", "nopager", "true"))
	assert.Equal(t, "nopager: true", strings.TrimSpace(buf.String()))
	buf.Reset()

	// action.printConfigValues
	act.cfg.Mounts["foo"] = &config.StoreConfig{}
	act.printConfigValues(ctx, "", "nopager")
	want = `nopager: true
foo/nopager: false`
	assert.Equal(t, want, strings.TrimSpace(buf.String()))
	buf.Reset()

	// action.setConfigValue on substore
	assert.NoError(t, act.setConfigValue(ctx, "foo", "cliptimeout", "23"))
	assert.Equal(t, "foo/cliptimeout: 23", strings.TrimSpace(buf.String()))
	buf.Reset()

	// action.printConfigValues
	act.printConfigValues(ctx, "")
	want = `root store config:
  askformore: false
  autoimport: true
  autosync: true
  cliptimeout: 45
  cryptobackend: gpg
  nocolor: false
  noconfirm: false
  nopager: true
  notifications: true
`
	want += "  path: " + u.StoreDir("") + "\n"
	want += `  safecontent: false
  storebackend: fs
  syncbackend: git
  usesymbols: false
mount 'foo' config:
  autoimport: false
  autosync: false
  cliptimeout: 23
  nopager: false
  notifications: false
  path:`
	assert.Equal(t, want, strings.TrimSpace(buf.String()))
	buf.Reset()

	delete(act.cfg.Mounts, "foo")
	buf.Reset()

	// config autoimport
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"autoimport"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, act.Config(ctx, c))
	assert.Equal(t, "autoimport: true", strings.TrimSpace(buf.String()))
	buf.Reset()

	// config autoimport false
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"autoimport", "false"}))
	c = cli.NewContext(app, fs, nil)
	assert.NoError(t, act.Config(ctx, c))
	assert.Equal(t, "autoimport: false", strings.TrimSpace(buf.String()))
	buf.Reset()

	// action.ConfigComplete
	act.ConfigComplete(c)
	want = `askformore
autoimport
autosync
cliptimeout
cryptobackend
nocolor
noconfirm
nopager
notifications
path
safecontent
storebackend
syncbackend
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
