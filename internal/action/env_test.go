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

func TestEnvFlagsMutuallyExclusive(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	flags := map[string]string{"stdin": "true", "file": "true"}
	require.EqualError(t, act.Env(gptest.CliCtxWithFlags(ctx, t, flags, "foo", "true")),
		"Only one of --stdin, --file or --exec may be specified")
}

func TestEnvStdinRequiresLeaf(t *testing.T) {
	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	require.NoError(t, act.insertStdin(ctx, "dir/key1", []byte("pw1"), false))

	flags := map[string]string{"stdin": "true"}
	require.EqualError(t, act.Env(gptest.CliCtxWithFlags(ctx, t, flags, "dir", "cat")),
		"--stdin requires a single secret, not a directory")
}

func TestEnvStdin(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires unix cat")
	}

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

	pw := pwgen.GeneratePassword(24, false)
	require.NoError(t, act.insertStdin(ctx, "baz", []byte(pw), false))
	buf.Reset()

	// cat reads from stdin and writes to stdout; we expect the raw password value.
	flags := map[string]string{"stdin": "true"}
	require.NoError(t, act.Env(gptest.CliCtxWithFlags(ctx, t, flags, "baz", "cat")))
	assert.Equal(t, pw, buf.String())
}

func TestEnvFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires unix env utility")
	}

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

	pw := pwgen.GeneratePassword(24, false)
	require.NoError(t, act.insertStdin(ctx, "baz", []byte(pw), false))
	buf.Reset()

	// env prints all env vars; we expect BAZ_FILE=<tempfile_path> in the output.
	flags := map[string]string{"file": "true"}
	require.NoError(t, act.Env(gptest.CliCtxWithFlags(ctx, t, flags, "baz", "env")))
	assert.Contains(t, buf.String(), "BAZ_FILE=")

	// Verify the temp file is cleaned up after the command returns.
	for _, line := range bytes.Split(buf.Bytes(), []byte("\n")) {
		if bytes.HasPrefix(line, []byte("BAZ_FILE=")) {
			path := string(bytes.TrimPrefix(line, []byte("BAZ_FILE=")))
			_, err := os.Stat(path)
			assert.True(t, os.IsNotExist(err), "temp file should be removed after command: %s", path)
		}
	}
}

func TestEnvExecNotFound(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("--exec on Windows returns 'not supported' error; covered by TestEnvExecUnsupportedOnWindows")
	}

	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	flags := map[string]string{"exec": "true"}
	err = act.Env(gptest.CliCtxWithFlags(ctx, t, flags, "foo", "non-existing-gopass-test-cmd"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-existing-gopass-test-cmd")
}

func TestEnvExecUnsupportedOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	u := gptest.NewUnitTester(t)

	ctx := config.NewContextInMemory()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = ctxutil.WithTerminal(ctx, false)
	act, err := newMock(ctx, u.StoreDir(""))
	require.NoError(t, err)
	require.NotNil(t, act)
	ctx = act.cfg.WithConfig(ctx)

	flags := map[string]string{"exec": "true"}
	require.EqualError(t, act.Env(gptest.CliCtxWithFlags(ctx, t, flags, "foo", "cmd")),
		"--exec is not supported on Windows")
}

// Ensure the original (default) behaviour is undisturbed when no mode flag is set.
func TestEnvDefaultModeUnchanged(t *testing.T) {
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

	pw := pwgen.GeneratePassword(24, false)
	require.NoError(t, act.insertStdin(ctx, "baz", []byte(pw), false))
	buf.Reset()

	require.NoError(t, act.Env(gptest.CliCtx(ctx, t, "baz", "env")))
	assert.Contains(t, buf.String(), fmt.Sprintf("BAZ=%s\n", pw))
}
