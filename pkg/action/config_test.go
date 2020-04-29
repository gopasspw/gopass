package action

import (
	"bytes"
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/tests/gptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	app := cli.NewApp()
	c := cli.NewContext(app, flag.NewFlagSet("default", flag.ContinueOnError), nil)
	c.Context = ctx

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	// action.Config
	assert.NoError(t, act.Config(c))
	want := `root store config:
  askformore: false
  autoclip: true
  autoimport: true
  autoprint: false
  autosync: true
  check_recipient_hash: false
  cliptimeout: 45
  concurrency: 1
  editrecipients: false
  nocolor: false
  noconfirm: false
  nopager: false
  notifications: true
`
	want += "  path: " + backend.FromPath(u.StoreDir("")).String() + "\n"
	want += `  safecontent: false
  usesymbols: false
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// action.setConfigValue
	assert.NoError(t, act.setConfigValue(ctx, "", "nopager", "true"))
	assert.Equal(t, "nopager: true", strings.TrimSpace(buf.String()), "action.setConfigValue")
	buf.Reset()

	// action.setConfigValue (invalid)
	assert.Error(t, act.setConfigValue(ctx, "", "foobar", "true"))
	buf.Reset()

	// action.printConfigValues
	act.cfg.Mounts["foo"] = &config.StoreConfig{}
	act.printConfigValues(ctx, "", "nopager")
	want = `nopager: true
foo/nopager: false`
	assert.Equal(t, want, strings.TrimSpace(buf.String()), "action.printConfigValues")
	buf.Reset()

	// action.setConfigValue on substore
	assert.NoError(t, act.setConfigValue(ctx, "foo", "cliptimeout", "23"))
	assert.Equal(t, "foo/cliptimeout: 23", strings.TrimSpace(buf.String()), "action.setConfigValue on substore")
	buf.Reset()

	// action.printConfigValues
	act.printConfigValues(ctx, "")
	want = `root store config:
  askformore: false
  autoclip: true
  autoimport: true
  autoprint: false
  autosync: true
  check_recipient_hash: false
  cliptimeout: 45
  concurrency: 1
  editrecipients: false
  nocolor: false
  noconfirm: false
  nopager: true
  notifications: true
`
	want += "  path: " + backend.FromPath(u.StoreDir("")).String() + "\n"
	want += `  safecontent: false
  usesymbols: false
mount 'foo' config:
  autoclip: false
  autoimport: false
  autosync: false
  cliptimeout: 23
  nopager: false
  notifications: false
`
	want += "  path: " + backend.FromPath("").String()
	assert.Equal(t, want, strings.TrimSpace(buf.String()), "action.printConfigValues")
	buf.Reset()

	delete(act.cfg.Mounts, "foo")
	buf.Reset()

	// config autoimport
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"autoimport"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx
	assert.NoError(t, act.Config(c))
	assert.Equal(t, "autoimport: true", strings.TrimSpace(buf.String()))
	buf.Reset()

	// config autoimport false
	fs = flag.NewFlagSet("default", flag.ContinueOnError)
	assert.NoError(t, fs.Parse([]string{"autoimport", "false"}))
	c = cli.NewContext(app, fs, nil)
	c.Context = ctx
	assert.NoError(t, act.Config(c))
	assert.Equal(t, "autoimport: false", strings.TrimSpace(buf.String()))
	buf.Reset()

	// action.ConfigComplete
	act.ConfigComplete(c)
	want = `askformore
autoclip
autoimport
autoprint
autosync
check_recipient_hash
cliptimeout
concurrency
editrecipients
nocolor
noconfirm
nopager
notifications
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
	c.Context = ctx
	assert.Error(t, act.Config(c))
	buf.Reset()
}
