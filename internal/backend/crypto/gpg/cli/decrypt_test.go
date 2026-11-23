package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gpgDecryptHelperEnv = "GOPASS_TEST_GPG_DECRYPT_HELPER"

func TestDecryptionSucceeded(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		status []string
		want   bool
	}{
		{
			name: "multi-ecdh throw-keyids success with gpg-exit failure",
			status: []string{
				"ENC_TO 0000000000000000 18 0",
				"KEY_CONSIDERED AAAA 0",
				"KEY_CONSIDERED BBBB 0",
				"DECRYPTION_KEY CCCC BBBB u",
				"BEGIN_DECRYPTION",
				"DECRYPTION_INFO 2 9 0",
				"PLAINTEXT 62 1795427797",
				"PLAINTEXT_LENGTH 15",
				"DECRYPTION_OKAY",
				"GOODMDC",
				"END_DECRYPTION",
				"FAILURE gpg-exit 33554433",
			},
			want: true,
		},
		{
			name: "multi-recipient with missing key for one recipient",
			status: []string{
				"ENC_TO 0000000000000000 18 0",
				"ENC_TO 0000000000000000 18 0",
				"NO_SECKEY 0000000000000000",
				"BEGIN_DECRYPTION",
				"DECRYPTION_OKAY",
				"GOODMDC",
				"END_DECRYPTION",
				"FAILURE gpg-exit 33554433",
			},
			want: true,
		},
		{
			name: "corrupted ciphertext with BADMDC",
			status: []string{
				"BEGIN_DECRYPTION",
				"PLAINTEXT 62 1795427797",
				"BADMDC",
				"DECRYPTION_FAILED",
				"END_DECRYPTION",
				"FAILURE gpg-exit 33554433",
			},
			want: false,
		},
		{
			name: "no matching secret key",
			status: []string{
				"NO_SECKEY 0000000000000000",
				"ERROR pkdecrypt_failed 16777226",
				"BEGIN_DECRYPTION",
				"DECRYPTION_FAILED",
				"END_DECRYPTION",
				"FAILURE gpg-exit 33554433",
			},
			want: false,
		},
		{
			name: "garbage input with NODATA and FAILURE decrypt",
			status: []string{
				"NODATA 1",
				"NODATA 2",
				"FAILURE decrypt 4294967295",
			},
			want: false,
		},
		{
			name: "bad signature is rejected",
			status: []string{
				"BEGIN_DECRYPTION",
				"BADSIG 1234567890ABCDEF",
				"DECRYPTION_OKAY",
				"GOODMDC",
				"END_DECRYPTION",
				"FAILURE gpg-exit 33554433",
			},
			want: false,
		},
		{
			name: "concatenated messages where second fails",
			status: []string{
				"BEGIN_DECRYPTION",
				"DECRYPTION_OKAY",
				"GOODMDC",
				"END_DECRYPTION",
				"BEGIN_DECRYPTION",
				"DECRYPTION_FAILED",
				"END_DECRYPTION",
			},
			want: false,
		},
		{
			name: "missing END_DECRYPTION",
			status: []string{
				"BEGIN_DECRYPTION",
				"DECRYPTION_OKAY",
				"GOODMDC",
			},
			want: false,
		},
		{
			name:   "empty status",
			status: nil,
			want:   false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, decryptionSucceeded(tc.status))
		})
	}
}

func TestGPGStatusFilter(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	f := &gpgStatusFilter{dst: &stderr}

	input := strings.Join([]string{
		"[GNUPG:] BEGIN_DECRYPTION",
		"gpg: ecdh failed in gcry_cipher_decrypt: Bad checksum",
		"[GNUPG:] DECRYPTION_OKAY\r",
		"[GNUPG:] END_DECRYPTION",
		"gpg: trailing line without newline",
	}, "\n")

	// Feed in small chunks to exercise mid-line buffering.
	for i := 0; i < len(input); i += 7 {
		end := min(i+7, len(input))
		n, err := f.Write([]byte(input[i:end]))
		require.NoError(t, err)
		require.Equal(t, end-i, n)
	}
	require.NoError(t, f.Flush())

	assert.Equal(t, []string{
		"BEGIN_DECRYPTION",
		"DECRYPTION_OKAY",
		"END_DECRYPTION",
	}, f.status)
	assert.Equal(t, "gpg: ecdh failed in gcry_cipher_decrypt: Bad checksum\ngpg: trailing line without newline", stderr.String())
}

func TestDecryptExitCodeHandling(t *testing.T) {
	for _, tc := range []struct {
		name      string
		mode      string
		wantPlain string
		wantErr   bool
	}{
		{
			name:      "exit 2 with DECRYPTION_OKAY succeeds",
			mode:      "exit2_ok",
			wantPlain: "secret-plaintext\n",
			wantErr:   false,
		},
		{
			name:    "exit 2 with BADMDC fails",
			mode:    "exit2_badmdc",
			wantErr: true,
		},
		{
			name:    "exit 1 with DECRYPTION_OKAY still fails",
			mode:    "exit1_ok",
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(gpgDecryptHelperEnv, tc.mode)

			ctx := config.NewContextInMemory()
			g := &GPG{
				binary: os.Args[0],
				args:   []string{"-test.run=^TestGPGDecryptHelperProcess$", "--"},
			}

			got, err := g.Decrypt(ctx, []byte("ciphertext"))
			if tc.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantPlain, string(got))
		})
	}
}

func TestGPGDecryptHelperProcess(t *testing.T) {
	mode := os.Getenv(gpgDecryptHelperEnv)
	if mode == "" {
		t.Skip("not the helper process")
	}

	switch mode {
	case "exit2_ok":
		_, _ = fmt.Fprint(os.Stdout, "secret-plaintext\n")
		_, _ = fmt.Fprintln(os.Stderr, "gpg: ecdh failed in gcry_cipher_decrypt: Bad checksum")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] BEGIN_DECRYPTION")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] DECRYPTION_OKAY")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] GOODMDC")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] END_DECRYPTION")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] FAILURE gpg-exit 33554433")
		os.Exit(2)
	case "exit2_badmdc":
		_, _ = fmt.Fprint(os.Stdout, "tampered-plaintext\n")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] BEGIN_DECRYPTION")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] BADMDC")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] DECRYPTION_FAILED")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] END_DECRYPTION")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] FAILURE gpg-exit 33554433")
		os.Exit(2)
	case "exit1_ok":
		_, _ = fmt.Fprint(os.Stdout, "secret-plaintext\n")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] BEGIN_DECRYPTION")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] DECRYPTION_OKAY")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] END_DECRYPTION")
		os.Exit(1)
	}
}

func TestDecryptMultiECDHThrowKeyids(t *testing.T) {
	if testing.Short() || runtime.GOOS != "linux" {
		t.Skip("skipping real gpg test in short mode or non-linux")
	}

	gpgBin, err := exec.LookPath("gpg")
	if err != nil {
		t.Skip("gpg binary not found")
	}

	gnupgHome := t.TempDir()
	require.NoError(t, os.Chmod(gnupgHome, 0o700))
	t.Setenv("GNUPGHOME", gnupgHome)

	runGPG := func(stdin []byte, args ...string) []byte {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), gpgBin, args...)
		if stdin != nil {
			cmd.Stdin = bytes.NewReader(stdin)
		}
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "gpg %v failed: %s", args, string(out))

		return out
	}

	// Generate two ECDH (ed25519 + cv25519) keys so gpg must try the wrong key
	// first when decrypting an anonymous-recipient (--throw-keyids) message for
	// the second key. On GnuPG 2.4.4+, that causes gpg to exit with status 2
	// even though decryption succeeds.
	for _, user := range []string{"one@example.com", "two@example.com"} {
		runGPG(nil, "--batch", "--passphrase", "", "--quick-generate-key", user, "ed25519", "sign", "never")
		out := runGPG(nil, "--with-colons", "--list-keys", user)
		var fpr string
		for line := range strings.SplitSeq(string(out), "\n") {
			if strings.HasPrefix(line, "fpr:") {
				fields := strings.Split(line, ":")
				if len(fields) > 9 {
					fpr = fields[9]

					break
				}
			}
		}
		require.NotEmpty(t, fpr)
		runGPG(nil, "--batch", "--passphrase", "", "--quick-add-key", fpr, "cv25519", "encr", "never")
	}

	wantPlain := []byte("multi-ecdh-throw-keyids-secret\n")
	encCmd := exec.CommandContext(t.Context(), gpgBin, "--batch", "--yes", "--trust-model", "always", "--throw-keyids", "--encrypt", "-r", "two@example.com")
	encCmd.Stdin = bytes.NewReader(wantPlain)
	ciphertext, err := encCmd.Output()
	require.NoError(t, err)

	ctx := config.NewContextInMemory()
	g, err := New(ctx, Config{Binary: gpgBin})
	require.NoError(t, err)

	gotPlain, err := g.Decrypt(ctx, ciphertext)
	require.NoError(t, err)
	assert.Equal(t, wantPlain, gotPlain)

	// Verify that a tampered ciphertext still fails decryption even though
	// multi-ECDH key trials also produce exit status 2.
	corrupted := bytes.Clone(ciphertext)
	corrupted[len(corrupted)-5] ^= 0xff
	_, err = g.Decrypt(ctx, corrupted)
	require.Error(t, err)
}
