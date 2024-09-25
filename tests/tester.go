package tests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/tests/can"
	"github.com/gopasspw/gopass/tests/gptest"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/stretchr/testify/require"
)

const (
	gopassConfig = `[core]
exportkeys = false
`
	keyID = "BE73F104"
)

// ErrNoCommand is returned when the command is missing.
var ErrNoCommand = fmt.Errorf("no command")

type tester struct {
	t *testing.T

	// Binary is the path to the gopass binary used for testing
	Binary    string
	sourceDir string
	tempDir   string
	resetFn   func()
}

// newTester is not compatible with t.Parallel because it uses t.Setenv.
func newTester(t *testing.T) *tester {
	t.Helper()

	sourceDir := "."
	if d := os.Getenv("GOPASS_TEST_DIR"); d != "" {
		sourceDir = d
	}

	gopassBin := ""
	if b := os.Getenv("GOPASS_BINARY"); b != "" {
		gopassBin = b
	}

	fi, err := os.Stat(gopassBin)
	if err != nil {
		t.Skipf("Failed to stat GOPASS_BINARY %s: %s", gopassBin, err)
	}

	if !strings.HasSuffix(gopassBin, ".exe") && fi.Mode()&0o111 == 0 {
		t.Fatalf("GOPASS_BINARY is not executeable")
	}

	t.Logf("Using gopass binary: %s", gopassBin)

	ts := &tester{
		t:         t,
		sourceDir: sourceDir,
		Binary:    gopassBin,
	}

	// create tempDir
	td, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)

	t.Logf("Tempdir: %s", td)
	ts.tempDir = td

	// prepare ENVIRONMENT
	ts.resetFn = gptest.UnsetVars("GNUPGHOME", "GOPASS_DEBUG", "NO_COLOR", "GOPASS_CONFIG", "GOPASS_NO_NOTIFY", "GOPASS_HOMEDIR")
	t.Setenv("GNUPGHOME", ts.gpgDir())
	t.Setenv("GOPASS_DEBUG", "")
	t.Setenv("NO_COLOR", "true")
	t.Setenv("GOPASS_CONFIG_NOSYSTEM", "true")
	t.Setenv("GOPASS_CONFIG_NO_MIGRATE", "true")
	t.Setenv("GOPASS_NO_NOTIFY", "true")
	t.Setenv("GOPASS_HOMEDIR", td)

	// write config
	require.NoError(t, os.MkdirAll(filepath.Dir(ts.gopassConfig()), 0o700))
	// we need to set the root path to something else than the root directory otherwise the mounts will show as regular entries
	if err := os.WriteFile(ts.gopassConfig(), []byte(gopassConfig+"\n[mounts]\npath = "+ts.storeDir("root")+"\n"), 0o600); err != nil {
		t.Fatalf("Failed to write gopass config to %s: %s", ts.gopassConfig(), err)
	}

	// copy gpg test files
	require.NoError(t, can.WriteTo(ts.gpgDir()))

	return ts
}

func (ts tester) gpgDir() string {
	return filepath.Join(ts.tempDir, ".gnupg")
}

func (ts tester) gopassConfig() string {
	return filepath.Join(ts.tempDir, ".config", "gopass", "config")
}

func (ts tester) storeDir(mount string) string {
	return filepath.Join(ts.tempDir, ".local", "share", "gopass", "stores", mount)
}

func (ts tester) workDir() string {
	return filepath.Dir(ts.tempDir)
}

func (ts tester) teardown() {
	ts.resetFn() // restore env vars

	if ts.tempDir == "" {
		return
	}

	err := os.RemoveAll(ts.tempDir)
	require.NoError(ts.t, err)
}

func (ts tester) runCmd(args []string, in []byte) (string, error) {
	ts.t.Helper()

	if len(args) < 1 {
		return "", fmt.Errorf("invalid args %v: %w", args, ErrNoCommand)
	}

	cmd := exec.CommandContext(context.Background(), args[0], args[1:]...)
	cmd.Dir = ts.workDir()
	cmd.Stdin = bytes.NewReader(in)

	ts.t.Logf("%+v", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return strings.TrimSpace(string(out)), nil
}

func (ts tester) run(arg string) (string, error) {
	ts.t.Helper()

	if runtime.GOOS == "windows" {
		arg = strings.ReplaceAll(arg, "\\", "\\\\")
	}

	args, err := shellquote.Split(arg)
	if err != nil {
		return "", fmt.Errorf("failed to split args %v: %w", arg, err)
	}

	cmd := exec.CommandContext(context.Background(), ts.Binary, args...)
	cmd.Dir = ts.workDir()

	ts.t.Logf("%+v", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}

	return strings.TrimSpace(string(out)), nil
}

func (ts tester) runWithInput(arg, input string) ([]byte, error) { //nolint:unused
	ts.t.Helper()

	reader := strings.NewReader(input)

	return ts.runWithInputReader(arg, reader)
}

func (ts tester) runWithInputReader(arg string, input io.Reader) ([]byte, error) { //nolint:unused
	ts.t.Helper()

	args, err := shellquote.Split(arg)
	if err != nil {
		return nil, fmt.Errorf("failed to split args %v: %w", arg, err)
	}

	cmd := exec.Command(ts.Binary, args...)
	cmd.Dir = ts.workDir()
	cmd.Stdin = input

	ts.t.Logf("%+v", cmd.Args)

	buf, err := cmd.CombinedOutput()
	if err != nil {
		return buf, fmt.Errorf("%s %v failed: %w", ts.Binary, args, err)
	}

	return buf, nil
}

func (ts *tester) initStore() {
	out, err := ts.run("init --crypto=gpgcli --storage=fs " + keyID)
	require.NoError(ts.t, err, "failed to init password store:\n%s", out)
}

func (ts *tester) initSecrets(prefix string) {
	out, err := ts.run("generate -p " + prefix + "foo/bar 20")
	require.NoError(ts.t, err, "failed to generate password:\n%s", out)

	out, err = ts.run("generate -p " + prefix + "baz 40")
	require.NoError(ts.t, err, "failed to generate password:\n%s", out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", prefix + "fixed/secret"}, []byte("moar"))
	require.NoError(ts.t, err, "failed to insert password:\n%s", out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", prefix + "fixed/twoliner"}, []byte("first line\nsecond line"))
	require.NoError(ts.t, err, "failed to insert password:\n%s", out)
}
