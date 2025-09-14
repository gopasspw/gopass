package action

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
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
		require.NoError(t, act.Config(c))
		want := `age.agent-enabled = false
core.autoimport = true
core.autopush = true
core.autosync = true
core.cliptimeout = 45
core.exportkeys = true
core.follow-references = false
core.nopager = true
core.notifications = true
generate.autoclip = true
`
		want += "mounts.path = " + u.StoreDir("") + "\n" +
			"pwgen.xkcd-lang = en\n"
		assert.Equal(t, want, buf.String())
	})

	t.Run("set valid config value", func(t *testing.T) {
		defer buf.Reset()

		require.NoError(t, act.setConfigValue(ctx, "", "core.nopager", "true"))

		// should print accepted config value
		assert.Equal(t, "true", strings.TrimSpace(buf.String()), "action.setConfigValue")
	})

	t.Run("set invalid config value", func(t *testing.T) {
		defer buf.Reset()

		require.Error(t, act.setConfigValue(ctx, "", "foobar", "true"))
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
		want := `age.agent-enabled = false
core.autoimport = true
core.autopush = true
core.autosync = true
core.cliptimeout = 45
core.exportkeys = true
core.follow-references = false
core.nopager = true
core.notifications = true
generate.autoclip = true
`
		want += "mounts.path = " + u.StoreDir("") + "\n" +
			"pwgen.xkcd-lang = en\n"

		assert.Equal(t, want, buf.String(), "action.printConfigValues")
	})

	t.Run("show autoimport value", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "core.autoimport")
		require.NoError(t, act.Config(c))
		assert.Equal(t, "true", strings.TrimSpace(buf.String()))
	})

	t.Run("disable autoimport", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "core.autoimport", "false")
		require.NoError(t, act.Config(c))
		assert.Equal(t, "false", strings.TrimSpace(buf.String()))
	})

	t.Run("complete config items", func(t *testing.T) {
		defer buf.Reset()

		act.ConfigComplete(gptest.CliCtx(ctx, t))
		want := `age.agent-enabled
core.autoimport
core.autopush
core.autosync
core.cliptimeout
core.exportkeys
core.follow-references
core.nopager
core.notifications
generate.autoclip
mounts.path
pwgen.xkcd-lang
`
		assert.Equal(t, want, buf.String())
	})

	t.Run("set autoimport to invalid value", func(t *testing.T) {
		defer buf.Reset()

		c := gptest.CliCtx(ctx, t, "autoimport", "false", "42")
		require.Error(t, act.Config(c))
	})
}
