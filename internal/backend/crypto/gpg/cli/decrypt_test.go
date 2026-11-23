package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	gpgDecryptHelperEnv     = "GOPASS_TEST_GPG_DECRYPT_HELPER"
	gpgDecryptHintHelperEnv = "GOPASS_TEST_GPG_DECRYPT_HINT_HELPER"
)

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

	t.Run("mid-line chunking", func(t *testing.T) {
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
	})

	t.Run("crlf and error status", func(t *testing.T) {
		t.Parallel()

		var stderr bytes.Buffer
		f := &gpgStatusFilter{dst: &stderr}

		chunks := []string{
			"[GNUPG:] ENC_TO DE252335EC8A1EB2 18 0\n",
			"gpg: public key decryption failed: ",
			"Inappropriate ioctl for device\r\n",
			"[GNUPG:] ERROR pkdecrypt_failed ",
			"83918950\r\n",
			"gpg: decryption failed: Inappropriate ioctl for device",
		}
		for _, chunk := range chunks {
			n, err := f.Write([]byte(chunk))
			require.NoError(t, err)
			assert.Equal(t, len(chunk), n)
		}
		require.NoError(t, f.Flush())

		assert.Equal(t, []string{
			"ENC_TO DE252335EC8A1EB2 18 0",
			"ERROR pkdecrypt_failed 83918950",
		}, f.status)
		assert.Equal(t,
			"gpg: public key decryption failed: Inappropriate ioctl for device\r\ngpg: decryption failed: Inappropriate ioctl for device",
			stderr.String(),
		)
	})
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

func TestDecryptErrorHint(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name           string
		statusLines    []string
		wantSubstrings []string
		wantAbsent     []string
		wantEmpty      bool
	}{
		{
			name: "pinentry ENOTTY with unset GPG_TTY",
			statusLines: []string{
				"ENC_TO DE252335EC8A1EB2 18 0",
				"PINENTRY_LAUNCHED 290390 curses 1.3.3 - dumb - - 1000/1000 0",
				"NO_SECKEY DE252335EC8A1EB2",
				"ERROR pkdecrypt_failed 83918950",
				"BEGIN_DECRYPTION",
				"DECRYPTION_FAILED",
				"END_DECRYPTION",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"Pinentry", "flavor=curses", "tty=<unset>", "GPG_TTY"},
			wantAbsent:     []string{"SCD", "smartcard"},
		},
		{
			name: "pinentry ENOENT with broken GPG_TTY",
			statusLines: []string{
				"PINENTRY_LAUNCHED 290244 curses 1.3.3 /dev/foobar dumb - ? 1000/1000 0",
				"ERROR pkdecrypt_failed 83918929",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"Pinentry", "flavor=curses", "tty=/dev/foobar", "GPG_TTY"},
			wantAbsent:     []string{"SCD", "smartcard"},
		},
		{
			name: "pinentry error without PINENTRY_LAUNCHED",
			statusLines: []string{
				"ERROR pkdecrypt_failed 83918950",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"gpg error source: Pinentry;", "GPG_TTY"},
			wantAbsent:     []string{"flavor=", "tty="},
		},
		{
			name: "pinentry canceled by user produces no misleading hint",
			statusLines: []string{
				"PINENTRY_LAUNCHED 290390 curses 1.3.3 /dev/pts/1 xterm - - 1000/1000 0",
				"ERROR pkdecrypt_failed 83886179",
				"FAILURE gpg-exit 33554433",
			},
			wantEmpty: true,
		},
		{
			name: "smartcard daemon ENOTTY",
			statusLines: []string{
				"ENC_TO DE252335EC8A1EB2 18 0",
				"ERROR pkdecrypt_failed 100696166",
				"BEGIN_DECRYPTION",
				"DECRYPTION_FAILED",
				"END_DECRYPTION",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"SCD (smartcard daemon)", "smartcard/security key", "scdaemon"},
			wantAbsent:     []string{"Pinentry", "GPG_TTY"},
		},
		{
			name: "smartcard daemon CARD_NOT_PRESENT",
			statusLines: []string{
				"ERROR pkdecrypt_failed 100663408",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"SCD (smartcard daemon)", "smartcard/security key"},
			wantAbsent:     []string{"Pinentry", "GPG_TTY"},
		},
		{
			name: "smartcard daemon BAD_PIN omits insertion hint",
			statusLines: []string{
				"ERROR pkdecrypt_failed 100663383",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"SCD (smartcard daemon)"},
			wantAbsent:     []string{"inserted", "GPG_TTY"},
		},
		{
			name: "smartcard daemon canceled produces no hint",
			statusLines: []string{
				"ERROR pkdecrypt_failed 100663395",
				"FAILURE gpg-exit 33554433",
			},
			wantEmpty: true,
		},
		{
			name: "GPG agent ENOTTY",
			statusLines: []string{
				"ERROR pkdecrypt_failed 67141734",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"GPG Agent", "GPG_TTY"},
		},
		{
			name: "GPG agent general error",
			statusLines: []string{
				"ERROR pkdecrypt_failed 67108865",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"gpg error source: GPG Agent"},
			wantAbsent:     []string{"GPG_TTY"},
		},
		{
			name: "symmetric decrypt get_passphrase ENOTTY after PINENTRY_LAUNCHED",
			statusLines: []string{
				"NEED_PASSPHRASE_SYM 9 3 10",
				"PINENTRY_LAUNCHED 291217 curses 1.3.3 - dumb - - 1000/1000 0",
				"ERROR get_passphrase 32870",
				"ERROR symkey_decrypt.maybe_error 11_BAD_PASSPHRASE",
				"DECRYPTION_FAILED",
				"FAILURE gpg-exit 33554433",
			},
			wantSubstrings: []string{"Pinentry", "flavor=curses", "tty=<unset>", "GPG_TTY"},
		},
		{
			name: "garbage input produces no hint",
			statusLines: []string{
				"NODATA 1",
				"NODATA 2",
				"FAILURE decrypt 4294967295",
			},
			wantEmpty: true,
		},
		{
			name: "generic GPG source NO_SECKEY produces no hint",
			statusLines: []string{
				"NO_SECKEY DE252335EC8A1EB2",
				"ERROR pkdecrypt_failed 33554449",
				"DECRYPTION_FAILED",
				"FAILURE gpg-exit 33554433",
			},
			wantEmpty: true,
		},
		{
			name: "TPM2d source",
			statusLines: []string{
				"ERROR pkdecrypt_failed 268435457",
			},
			wantSubstrings: []string{"TPM2d"},
		},
		{
			name: "Assuan source",
			statusLines: []string{
				"ERROR pkdecrypt_failed 251658241",
			},
			wantSubstrings: []string{"Assuan"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := decryptErrorHint(tc.statusLines)
			if tc.wantEmpty {
				assert.Empty(t, got)

				return
			}
			for _, sub := range tc.wantSubstrings {
				assert.Contains(t, got, sub)
			}
			for _, sub := range tc.wantAbsent {
				assert.NotContains(t, got, sub)
			}
		})
	}
}

func TestDecryptErrorSourcePropagation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mode         string
		wantContains []string
		wantAbsent   []string
	}{
		{
			name:         "pinentry ENOTTY",
			mode:         "pinentry-enotty",
			wantContains: []string{"Pinentry", "flavor=curses", "tty=<unset>", "GPG_TTY"},
			wantAbsent:   []string{"SCD"},
		},
		{
			name:         "smartcard SCD ENOTTY",
			mode:         "scd-enotty",
			wantContains: []string{"SCD (smartcard daemon)", "smartcard/security key"},
			wantAbsent:   []string{"Pinentry", "GPG_TTY"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(gpgDecryptHintHelperEnv, tc.mode)

			ctx := config.NewContextInMemory()
			g := &GPG{
				binary: os.Args[0],
				args:   []string{"-test.run=^TestGPGDecryptHintHelperProcess$", "--"},
			}

			_, err := g.Decrypt(ctx, []byte("ciphertext"))
			require.Error(t, err)

			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr, "wrapped error should preserve *exec.ExitError")
			assert.Equal(t, 2, exitErr.ExitCode())

			for _, sub := range tc.wantContains {
				assert.Contains(t, err.Error(), sub)
			}
			for _, sub := range tc.wantAbsent {
				assert.NotContains(t, err.Error(), sub)
			}
		})
	}
}

// TestGPGDecryptHintHelperProcess is only executed as a subprocess by
// TestDecryptErrorSourcePropagation to simulate GPG --status-fd 2 output.
func TestGPGDecryptHintHelperProcess(t *testing.T) {
	mode := os.Getenv(gpgDecryptHintHelperEnv)
	if mode == "" {
		t.Skip("not the helper process")
	}

	switch mode {
	case "pinentry-enotty":
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] PINENTRY_LAUNCHED 12345 curses 1.3.3 - dumb - - 1000/1000 0")
		_, _ = fmt.Fprintln(os.Stderr, "gpg: public key decryption failed: Inappropriate ioctl for device")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] ERROR pkdecrypt_failed 83918950")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] DECRYPTION_FAILED")
		_, _ = fmt.Fprintln(os.Stderr, "gpg: decryption failed: Inappropriate ioctl for device")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] FAILURE gpg-exit 33554433")
	case "scd-enotty":
		_, _ = fmt.Fprintln(os.Stderr, "gpg: public key decryption failed: Inappropriate ioctl for device")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] ERROR pkdecrypt_failed 100696166")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] DECRYPTION_FAILED")
		_, _ = fmt.Fprintln(os.Stderr, "gpg: decryption failed: Inappropriate ioctl for device")
		_, _ = fmt.Fprintln(os.Stderr, "[GNUPG:] FAILURE gpg-exit 33554433")
	}

	os.Exit(2)
}

func TestDecryptRealGPGPinentryError(t *testing.T) {
	if testing.Short() || runtime.GOOS != "linux" {
		t.Skip("skipping real-gpg pinentry test")
	}

	gpgBin, err := exec.LookPath("gpg")
	if err != nil {
		t.Skip("gpg not found in PATH")
	}
	gpgconfBin, err := exec.LookPath("gpgconf")
	if err != nil {
		t.Skip("gpgconf not found in PATH")
	}
	pinentryBin, err := exec.LookPath("pinentry-curses")
	if err != nil {
		t.Skip("pinentry-curses not found in PATH")
	}

	gnupgHome := t.TempDir()
	require.NoError(t, os.Chmod(gnupgHome, 0o700))

	t.Setenv("GNUPGHOME", gnupgHome)
	t.Setenv("LC_ALL", "C")
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")

	agentConf := filepath.Join(gnupgHome, "gpg-agent.conf")
	require.NoError(t, os.WriteFile(agentConf, []byte("pinentry-program "+pinentryBin+"\n"), 0o600))

	t.Cleanup(func() {
		_ = exec.Command(gpgconfBin, "--kill", "gpg-agent").Run()
	})

	// Generate a passphrase-protected key so decryption requires pinentry.
	genCmd := exec.Command(gpgBin, "--batch", "--gen-key")
	genCmd.Stdin = bytes.NewBufferString(
		"Key-Type: eddsa\n" +
			"Key-Curve: ed25519\n" +
			"Subkey-Type: ecdh\n" +
			"Subkey-Curve: cv25519\n" +
			"Name-Real: Pinentry Hint Test\n" +
			"Name-Email: pinentry-hint@example.com\n" +
			"Passphrase: test-passphrase\n" +
			"%commit\n",
	)
	out, err := genCmd.CombinedOutput()
	require.NoError(t, err, "gpg --gen-key failed: %s", string(out))

	// Encrypt a test secret to the generated key.
	encCmd := exec.Command(gpgBin, "--batch", "--yes", "--trust-model", "always", "-e", "-r", "pinentry-hint@example.com")
	encCmd.Stdin = bytes.NewBufferString("top secret\n")
	ciphertext, err := encCmd.Output()
	require.NoError(t, err)

	ctx := config.NewContextInMemory()
	g := &GPG{
		binary: gpgBin,
		args:   defaultArgs,
	}

	// 1. Unset GPG_TTY triggers "Inappropriate ioctl for device" from Pinentry.
	t.Run("unset GPG_TTY", func(t *testing.T) {
		t.Setenv("GPG_TTY", "")
		require.NoError(t, exec.Command(gpgconfBin, "--kill", "gpg-agent").Run())

		_, err := g.Decrypt(ctx, ciphertext)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Pinentry")
		assert.Contains(t, err.Error(), "tty=<unset>")
		assert.Contains(t, err.Error(), "GPG_TTY")
	})

	// 2. Broken GPG_TTY path triggers ENOENT from Pinentry.
	t.Run("broken GPG_TTY path", func(t *testing.T) {
		badTTY := "/dev/nonexistent-gopass-test-tty"
		t.Setenv("GPG_TTY", badTTY)
		require.NoError(t, exec.Command(gpgconfBin, "--kill", "gpg-agent").Run())

		_, err := g.Decrypt(ctx, ciphertext)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Pinentry")
		assert.Contains(t, err.Error(), "tty="+badTTY)
		assert.Contains(t, err.Error(), "GPG_TTY")
	})
}
