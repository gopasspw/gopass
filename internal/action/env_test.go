package action

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvLeafHappyPath(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	// Command-line would be: "gopass env foo env", where "foo" is an existing
	// secret with value "secret". We expect to see the key/value in the output
	// of the /usr/bin/env utility in the form "BAZ=secret".
	pw := pwgen.GeneratePassword(24, false)
	require.NoError(t, act.insertStdin(ctx, "baz", []byte(pw), false))
	buf.Reset()

	require.NoError(t, act.Env(gptest.CliCtx(ctx, t, "baz", "env")))
	assert.Contains(t, buf.String(), fmt.Sprintf("BAZ=%s\n", pw))
}

func TestEnvLeafHappyPathKeepCase(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf
	stdout = buf
	defer func() {
		out.Stdout = os.Stdout
		out.Stderr = os.Stderr
		stdout = os.Stdout
	}()

	// Command-line would be: "gopass env --keep-case BaZ env", where
	// "foo" is an existing secret with value "secret". We expect to see the
	// key/value in the output of the /usr/bin/env utility in the form
	// "BaZ=secret".
	pw := pwgen.GeneratePassword(24, false)
	require.NoError(t, act.insertStdin(ctx, "BaZ", []byte(pw), false))
	buf.Reset()

	flags := make(map[string]string, 1)
	flags["keep-case"] = "true"
	require.NoError(t, act.Env(gptest.CliCtxWithFlags(ctx, t, flags, "BaZ", "env")))
	assert.Contains(t, buf.String(), fmt.Sprintf("BaZ=%s\n", pw))
}

func TestEnvSecretNotFound(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	// Command-line would be: "gopass env non-existing true".
	require.EqualError(t, act.Env(gptest.CliCtx(ctx, t, "non-existing", "true")),
		"Secret non-existing not found")
}

func TestEnvProgramNotFound(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	wanted := "exec: \"non-existing\": executable file not found in "
	if runtime.GOOS == "windows" {
		wanted += "%PATH%"
	} else {
		wanted += "$PATH"
	}

	// Command-line would be: "gopass env foo non-existing".
	require.EqualError(t, act.Env(gptest.CliCtx(ctx, t, "foo", "non-existing")),
		wanted)
}

// Crash regression.
func TestEnvProgramNotSpecified(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	// Command-line would be: "gopass env foo".
	require.EqualError(t, act.Env(gptest.CliCtx(ctx, t, "foo")),
		"Missing subcommand to execute")
}
