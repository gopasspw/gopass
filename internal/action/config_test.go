package action

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithInteractive(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		stdout = os.Stdout
	}()

	t.Run("display config", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t)
		assert.NoError(t, act.Config(c))
		want := `core.autoimport = true
core.autopush = true
core.autosync = true
core.cliptimeout = 45
core.exportkeys = true
core.nopager = true
core.notifications = true
generate.autoclip = true
`
		want += "mounts.path = " + u.StoreDir("") + "\n"
		assert.Equal(t, want, buf.String())
	})

	t.Run("set valid config value", func(t *testing.T) {
		defer buf.Reset()

		assert.NoError(t, act.setConfigValue(ctx, "", "core.nopager", "true"))

		// should print accepted config value
		assert.Equal(t, "true", strings.TrimSpace(buf.String()), "action.setConfigValue")
	})

	t.Run("set invalid config value", func(t *testing.T) {
		defer buf.Reset()

		assert.Error(t, act.setConfigValue(ctx, "", "foobar", "true"))
	})

	t.Run("print single config value", func(t *testing.T) {
		defer buf.Reset()

		act.printConfigValues(ctx, "", "core.nopager")

		want := "true"
		assert.Equal(t, want, strings.TrimSpace(buf.String()), "action.printConfigValues")
	})

	t.Run("print all config values", func(t *testing.T) {
		defer buf.Reset()

		act.printConfigValues(ctx, "")
		want := `core.autoimport = true
core.autopush = true
core.autosync = true
core.cliptimeout = 45
core.exportkeys = true
core.nopager = true
core.notifications = true
generate.autoclip = true
`
		want += "mounts.path = " + u.StoreDir("")
		assert.Equal(t, want, strings.TrimSpace(buf.String()), "action.printConfigValues")
	})

	t.Run("show autoimport value", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "core.autoimport")
		assert.NoError(t, act.Config(c))
		assert.Equal(t, "true", strings.TrimSpace(buf.String()))
	})

	t.Run("disable autoimport", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "core.autoimport", "false")
		assert.NoError(t, act.Config(c))
		assert.Equal(t, "false", strings.TrimSpace(buf.String()))
	})

	t.Run("complete config items", func(t *testing.T) {
		defer buf.Reset()

		act.ConfigComplete(gptest.CliCtx(ctx, t))
		want := `core.autoimport
core.autopush
core.autosync
core.cliptimeout
core.exportkeys
core.nopager
core.notifications
generate.autoclip
mounts.path
`
		assert.Equal(t, want, buf.String())
	})

	t.Run("set autoimport to invalid value", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "autoimport", "false", "42")
		assert.Error(t, act.Config(c))
	})
}
