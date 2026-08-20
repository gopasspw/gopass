package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

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

	if err := os.Setenv("GNUPGHOME", home); err != nil {
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

	return home, func() { _ = os.RemoveAll(home) }, nil
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
	require.Equal(t, os.FileMode(0o700), info.Mode().Perm())
}
