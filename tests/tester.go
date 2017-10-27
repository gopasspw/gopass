package tests

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"io"

	shellquote "github.com/kballard/go-shellquote"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	gopassConfig = `askformore: false
autoimport: true
autosync: true
cliptimeout: 45
noconfirm: true
safecontent: true`
	keyID = "BE73F104"
)

type tester struct {
	t *testing.T

	// Binary is the path to the gopass binary used for testing
	Binary    string
	sourceDir string
	tempDir   string
}

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
		t.Fatalf("Failed to stat GOPASS_BINARY %s: %s", gopassBin, err)
	}
	if !strings.HasSuffix(gopassBin, ".exe") && fi.Mode()&0111 == 0 {
		t.Fatalf("GOPASS_BINARY is not executeable")
	}
	t.Logf("Using gopass binary: %s", gopassBin)

	ts := &tester{
		t:         t,
		sourceDir: sourceDir,
		Binary:    gopassBin,
	}
	// create tempDir
	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)

	t.Logf("Tempdir: %s", td)
	ts.tempDir = td

	// prepare ENVIRONMENT
	_ = os.Setenv("GNUPGHOME", ts.gpgDir())
	_ = os.Setenv("GOPASS_DEBUG", "")
	_ = os.Setenv("GOPASS_NOCOLOR", "true")
	_ = os.Setenv("GOPASS_CONFIG", ts.gopassConfig())

	// write config
	if err := ioutil.WriteFile(ts.gopassConfig(), []byte(gopassConfig+"\npath: "+ts.storeDir("")+"\n"), 0600); err != nil {
		t.Fatalf("Failed to write gopass config to %s: %s", ts.gopassConfig(), err)
	}

	// copy gpg test files
	files := map[string]string{
		ts.sourceDir + "/can/gnupg/pubring.gpg": ts.gpgDir() + "/pubring.gpg",
		ts.sourceDir + "/can/gnupg/random_seed": ts.gpgDir() + "/random_seed",
		ts.sourceDir + "/can/gnupg/secring.gpg": ts.gpgDir() + "/secring.gpg",
		ts.sourceDir + "/can/gnupg/trustdb.gpg": ts.gpgDir() + "/trustdb.gpg",
	}
	for from, to := range files {
		buf, err := ioutil.ReadFile(from)
		require.NoError(t, err, "Failed to read file %s", from)

		err = os.MkdirAll(filepath.Dir(to), 0700)
		require.NoError(t, err, "Failed to create dir for %s", to)

		err = ioutil.WriteFile(to, buf, 0600)
		require.NoError(t, err, "Failed to write file %s", to)
	}

	return ts
}

func (ts tester) gpgDir() string {
	return filepath.Join(ts.tempDir, ".gnupg")
}

func (ts tester) gopassConfig() string {
	return filepath.Join(ts.tempDir, ".gopass.yml")
}

func (ts tester) storeDir(mount string) string {
	if mount != "" {
		mount = "-" + mount
	}
	return filepath.Join(ts.tempDir, ".password-store"+mount)
}

func (ts tester) workDir() string {
	return filepath.Dir(ts.tempDir)
}

func (ts tester) teardown() {
	if ts.tempDir == "" {
		return
	}
	err := os.RemoveAll(ts.tempDir)
	require.NoError(ts.t, err)
}

func (ts tester) runCmd(args []string, in []byte) (string, error) {
	if len(args) < 1 {
		return "", errors.Errorf("no command")
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
	args, err := shellquote.Split(arg)
	if err != nil {
		return "", err
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

func (ts tester) runWithInput(arg, input string) ([]byte, error) {
	reader := strings.NewReader(input)
	out, err := ts.runWithInputReader(arg, reader)
	return out, err
}

func (ts tester) runWithInputReader(arg string, input io.Reader) ([]byte, error) {
	args, err := shellquote.Split(arg)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ts.Binary, args...)
	cmd.Dir = ts.workDir()
	cmd.Stdin = input

	ts.t.Logf("%+v", cmd.Args)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, err
	}

	return out, nil
}

func (ts *tester) initStore() {
	out, err := ts.run("init --nogit " + keyID)
	require.NoError(ts.t, err, "failed to init password store:\n%s", out)
}

func (ts *tester) initSecrets(prefix string) {
	out, err := ts.run("generate -p " + prefix + "foo/bar 20")
	require.NoError(ts.t, err, "failed to generate password:\n%s", out)

	out, err = ts.run("generate -p " + prefix + "baz 40")
	require.NoError(ts.t, err, "failed to generate password:\n%s", out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", prefix + "fixed/secret"}, []byte("moar"))
	require.NoError(ts.t, err, "failed to insert password:\n%s", out)

	out, err = ts.runCmd([]string{ts.Binary, "insert", prefix + "fixed/twoliner"}, []byte("and\nmore stuff"))
	require.NoError(ts.t, err, "failed to insert password:\n%s", out)
}
