package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg/gpgconf"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	base, err := os.MkdirTemp("", "gopass-gpg-test-")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "create GnuPG test base directory: %v\n", err)
		os.Exit(1)
	}

	home, cleanup, err := prepareTestGPGHome(base)
	if err != nil {
		_ = os.RemoveAll(base)
		_, _ = fmt.Fprintf(os.Stderr, "prepare GnuPG test home: %v\n", err)
		os.Exit(1)
	}

	if err := os.Setenv("GNUPGHOME", gpgHomeEnv(home, runtime.GOOS)); err != nil {
		cleanup()
		_ = os.RemoveAll(base)
		_, _ = fmt.Fprintf(os.Stderr, "set GNUPGHOME for tests: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	cleanup()
	_ = os.RemoveAll(base)
	os.Exit(code)
}

func prepareTestGPGHome(base string) (string, func(), error) {
	home := filepath.Join(base, ".gnupg")
	if err := os.Mkdir(home, 0o700); err != nil {
		return "", func() {}, fmt.Errorf("create %s: %w", home, err)
	}

	cleanup := func() { _ = os.RemoveAll(home) }

	// Pre-initialize pubring.kbx and trustdb.gpg sequentially before any
	// t.Parallel() tests invoke gpg against the shared GNUPGHOME, otherwise
	// concurrent gpg processes race on initial trustdb creation.
	if err := initTestGPGHome(home); err != nil {
		cleanup()

		return "", func() {}, fmt.Errorf("initialize %s: %w", home, err)
	}

	return home, cleanup, nil
}

func initTestGPGHome(home string) error {
	bin, err := gpgconf.Binary(context.Background(), "")
	if err != nil {
		return nil
	}

	cmd := exec.CommandContext(context.Background(), bin, "--batch", "--list-keys")
	cmd.Env = withEnv(os.Environ(), "GNUPGHOME", gpgHomeEnv(home, runtime.GOOS))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s --batch --list-keys: %w (%s)", bin, err, strings.TrimSpace(string(out)))
	}

	return nil
}

func withEnv(environ []string, key, value string) []string {
	prefix := key + "="
	env := make([]string, 0, len(environ)+1)
	for _, entry := range environ {
		if strings.HasPrefix(entry, prefix) {
			continue
		}
		env = append(env, entry)
	}
	env = append(env, prefix+value)

	return env
}

func gpgHomeEnv(home, goos string) string {
	if goos != "windows" {
		return home
	}

	home = strings.ReplaceAll(home, `\`, "/")
	if len(home) >= 2 && home[1] == ':' {
		return "/" + strings.ToLower(home[:1]) + home[2:]
	}

	return home
}

func TestPrepareTestGPGHome(t *testing.T) {
	base := t.TempDir()
	home, cleanup, err := prepareTestGPGHome(base)
	require.NoError(t, err)
	t.Cleanup(cleanup)

	require.Equal(t, filepath.Join(base, ".gnupg"), home)

	info, err := os.Stat(home)
	require.NoError(t, err)
	require.True(t, info.IsDir())
	if runtime.GOOS != "windows" {
		require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	}

	if _, err := gpgconf.Binary(t.Context(), ""); err == nil {
		_, err := os.Stat(filepath.Join(home, "trustdb.gpg"))
		require.NoError(t, err)
	}
}

func TestGPGHomeEnv(t *testing.T) {
	require.Equal(t, "/c/msys64/tmp/gopass/.gnupg", gpgHomeEnv(`C:\msys64\tmp\gopass\.gnupg`, "windows"))
	require.Equal(t, "/tmp/gopass/.gnupg", gpgHomeEnv("/tmp/gopass/.gnupg", "linux"))
}
