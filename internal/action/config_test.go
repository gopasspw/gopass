package action

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

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
  autoclip: true
  autoimport: true
  cliptimeout: 45
  confirm: false
  editrecipients: false
  exportkeys: false
  nocolor: false
  nopager: false
  notifications: true
`
	want += "  path: " + u.StoreDir("") + "\n"
	want += `  safecontent: false
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
	act.printConfigValues(ctx, "", "nopager")
	want = "nopager: true"
	assert.Equal(t, want, strings.TrimSpace(buf.String()), "action.printConfigValues")
	buf.Reset()

	// action.printConfigValues
	act.printConfigValues(ctx, "")
	want = `root store config:
  autoclip: true
  autoimport: true
  cliptimeout: 45
  confirm: false
  editrecipients: false
  exportkeys: false
  nocolor: false
  nopager: true
  notifications: true
`
	want += "  path: " + u.StoreDir("") + "\n"
	want += `  safecontent: false`
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
	want = `autoclip
autoimport
cliptimeout
confirm
editrecipients
exportkeys
nocolor
nopager
notifications
path
safecontent
`
	assert.Equal(t, want, buf.String())
	buf.Reset()

	// config autoimport false 42
	c = gptest.CliCtx(ctx, t, "autoimport", "false", "42")
	assert.Error(t, act.Config(c))
	buf.Reset()
}
