package action

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	// action.Config
	c := gptest.CliCtx(ctx, t)
	assert.NoError(t, act.Config(c))
	want := `root store config:
  askformore: false
  autoclip: true
  autoimport: true
  autosync: true
  check_recipient_hash: false
  cliptimeout: 45
  concurrency: 1
  editrecipients: false
  exportkeys: true
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
  autosync: true
  check_recipient_hash: false
  cliptimeout: 45
  concurrency: 1
  editrecipients: false
  exportkeys: true
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
  exportkeys: false
  nopager: false
  notifications: false
`
	want += "  path: " + backend.FromPath("").String()
	assert.Equal(t, want, strings.TrimSpace(buf.String()), "action.printConfigValues")
	buf.Reset()

	delete(act.cfg.Mounts, "foo")
	buf.Reset()

	// config autoimport
	c = gptest.CliCtx(ctx, t, "autoimport")
	assert.NoError(t, act.Config(c))
	assert.Equal(t, "autoimport: true", strings.TrimSpace(buf.String()))
	buf.Reset()

	// config autoimport false
	c = gptest.CliCtx(ctx, t, "autoimport", "false")
	assert.NoError(t, act.Config(c))
	assert.Equal(t, "autoimport: false", strings.TrimSpace(buf.String()))
	buf.Reset()

	// action.ConfigComplete
	act.ConfigComplete(c)
	want = `askformore
autoclip
autoimport
autosync
check_recipient_hash
cliptimeout
concurrency
editrecipients
exportkeys
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
	c = gptest.CliCtx(ctx, t, "autoimport", "false", "42")
	assert.Error(t, act.Config(c))
	buf.Reset()
}
