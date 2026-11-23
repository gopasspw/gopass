package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func shortGPGTempDir(t *testing.T) string {
	t.Helper()

	if runtime.GOOS == "windows" {
		return t.TempDir()
	}

	// macOS has a 104-byte UNIX socket path limit (sun_path), which t.TempDir()
	// under /var/folders/... easily exceeds for gpg-agent sockets.
	dir, err := os.MkdirTemp("/tmp", "gp-*") //nolint:usetesting // need short path under /tmp for macOS gpg-agent socket
	if err != nil {
		return t.TempDir()
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir
}

func genGPGTestKey(t *testing.T, homedir, uid string) (string, string) {
	t.Helper()

	gpgHome := gpgHomeEnv(homedir, runtime.GOOS)

	cmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--batch", "--passphrase", "", "--quick-gen-key", uid, "default", "default", "never")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "quick-gen-key failed: %s", string(out))

	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", gpgHome, "--kill", "gpg-agent").Run()
	})

	listCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--with-colons", "--with-fingerprint", "--list-secret-keys")
	listOut, err := listCmd.Output()
	require.NoError(t, err)

	var fpr, subFpr string

	for line := range strings.SplitSeq(string(listOut), "\n") {
		if !strings.HasPrefix(line, "fpr:") {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) < 10 {
			continue
		}
		if fpr == "" {
			fpr = fields[9]
		} else if subFpr == "" {
			subFpr = fields[9]
		}
	}

	require.Len(t, fpr, 40, "expected 40-char primary fingerprint")

	return fpr, subFpr
}

func TestFindIdentitiesFormatsAndLocale(t *testing.T) {
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not found in PATH")
	}

	root := shortGPGTempDir(t)
	userHome := filepath.Join(root, "u")
	teamHome := filepath.Join(root, "t")

	// Pre-create homedirs with 0700 permissions for gpg.
	require.NoError(t, os.MkdirAll(userHome, 0o700))
	require.NoError(t, os.MkdirAll(teamHome, 0o700))

	myFpr, mySubFpr := genGPGTestKey(t, userHome, "Alice Tester <alice@example.org>")
	teamFpr, _ := genGPGTestKey(t, teamHome, "Bob Teammate <bob@example.org>")

	userGPGHome := gpgHomeEnv(userHome, runtime.GOOS)
	teamGPGHome := gpgHomeEnv(teamHome, runtime.GOOS)

	// Export teammate's public key and import it into user's keyring (mirrors JoinTeam step 1).
	exportCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", teamGPGHome, "--armor", "--export", teamFpr)
	teamPub, err := exportCmd.Output()
	require.NoError(t, err)

	importCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", userGPGHome, "--import")
	importCmd.Stdin = strings.NewReader(string(teamPub))
	importOut, err := importCmd.CombinedOutput()
	require.NoError(t, err, "import teammate pubkey failed: %s", string(importOut))

	t.Setenv("GNUPGHOME", userGPGHome)
	// Force a non-English locale to verify listKeys still recognizes "No secret key".
	t.Setenv("LC_ALL", "de_DE.UTF-8")
	t.Setenv("LANGUAGE", "de")

	ctx := config.NewContextInMemory()
	ctx = gpg.WithAlwaysTrust(ctx, false)
	ctx = gpg.WithUseCache(ctx, false)

	g, err := New(ctx, Config{})
	require.NoError(t, err)

	longID := myFpr[len(myFpr)-16:]
	shortID := myFpr[len(myFpr)-8:]
	wantID := "0x" + longID

	cases := []struct {
		name    string
		queries []string
		want    []string
	}{
		{
			name:    "full fingerprint",
			queries: []string{myFpr},
			want:    []string{wantID},
		},
		{
			name:    "0x-prefixed full fingerprint",
			queries: []string{"0x" + myFpr},
			want:    []string{wantID},
		},
		{
			name:    "long key ID without 0x",
			queries: []string{longID},
			want:    []string{wantID},
		},
		{
			name:    "long key ID with 0x",
			queries: []string{"0x" + longID},
			want:    []string{wantID},
		},
		{
			name:    "short key ID without 0x",
			queries: []string{shortID},
			want:    []string{wantID},
		},
		{
			name:    "short key ID with 0x",
			queries: []string{"0x" + shortID},
			want:    []string{wantID},
		},
		{
			name:    "email address",
			queries: []string{"alice@example.org"},
			want:    []string{wantID},
		},
		{
			name:    "angle-bracketed email",
			queries: []string{"<alice@example.org>"},
			want:    []string{wantID},
		},
		{
			name:    "mixed team recipients with full fingerprint",
			queries: []string{teamFpr, myFpr},
			want:    []string{wantID},
		},
		{
			name:    "mixed team recipients with email",
			queries: []string{"bob@example.org", "alice@example.org"},
			want:    []string{wantID},
		},
		{
			name:    "only teammate key returns empty without error under non-English locale",
			queries: []string{teamFpr},
			want:    []string{},
		},
		{
			name:    "unknown key returns empty without error under non-English locale",
			queries: []string{"0xDEADBEEFDEADBEEF"},
			want:    []string{},
		},
		{
			name:    "flag-like recipient string is not interpreted as option",
			queries: []string{"--version"},
			want:    []string{},
		},
	}

	if mySubFpr != "" {
		cases = append(cases,
			struct {
				name    string
				queries []string
				want    []string
			}{
				name:    "subkey fingerprint",
				queries: []string{mySubFpr},
				want:    []string{wantID},
			},
			struct {
				name    string
				queries []string
				want    []string
			}{
				name:    "subkey long ID with bang",
				queries: []string{"0x" + mySubFpr[len(mySubFpr)-16:] + "!"},
				want:    []string{wantID},
			},
		)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := g.FindIdentities(ctx, tc.queries...)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
