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

func TestSplitPacket(t *testing.T) {
	t.Parallel()

	for in, out := range map[string]map[string]string{
		"": {},
		":pubkey enc packet: version 3, algo 1, keyid 00F0FF00FFC00F0F": {
			"algo":    "1",
			"keyid":   "00F0FF00FFC00F0F",
			"version": "3",
		},
		":encrypted data packet:": {},
	} {
		assert.Equal(t, out, splitPacket(in))
	}
}

type subkeyTestKeyring struct {
	root       string
	gpgHome    string
	primaryFpr string
	sub1Fpr    string
	sub2Fpr    string
	signSubFpr string
}

func setupSubkeyTestKeyring(t *testing.T) subkeyTestKeyring {
	t.Helper()

	root := shortGPGTempDir(t)
	home := filepath.Join(root, "h")
	require.NoError(t, os.MkdirAll(home, 0o700))

	gpgHome := gpgHomeEnv(home, runtime.GOOS)

	primaryFpr, sub1Fpr := genGPGTestKey(t, home, "Subkey User <subkey@example.org>")
	require.Len(t, sub1Fpr, 40, "expected initial encryption subkey")

	// Add a second encryption subkey and a signing-only subkey.
	addEncCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--batch", "--passphrase", "", "--quick-add-key", primaryFpr, "cv25519", "encr", "never")
	addOut, err := addEncCmd.CombinedOutput()
	require.NoError(t, err, "quick-add-key encr failed: %s", string(addOut))

	addSignCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--batch", "--passphrase", "", "--quick-add-key", primaryFpr, "ed25519", "sign", "never")
	addOut, err = addSignCmd.CombinedOutput()
	require.NoError(t, err, "quick-add-key sign failed: %s", string(addOut))

	// Collect all subkey fingerprints.
	listCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--with-colons", "--with-fingerprint", "--list-public-keys", primaryFpr)
	listOut, err := listCmd.Output()
	require.NoError(t, err)

	var sub2Fpr, signSubFpr, curSubCap string
	for line := range strings.SplitSeq(string(listOut), "\n") {
		fields := strings.Split(line, ":")
		if len(fields) < 10 {
			continue
		}
		switch fields[0] {
		case "sub":
			curSubCap = fields[11]
		case "fpr":
			if curSubCap == "" {
				continue
			}
			fp := fields[9]
			switch {
			case strings.Contains(curSubCap, "e") && fp != sub1Fpr:
				sub2Fpr = fp
			case strings.Contains(curSubCap, "s") && !strings.Contains(curSubCap, "e"):
				signSubFpr = fp
			}
			curSubCap = ""
		}
	}
	require.Len(t, sub2Fpr, 40, "expected second encryption subkey")
	require.Len(t, signSubFpr, 40, "expected signing-only subkey")

	return subkeyTestKeyring{
		root:       root,
		gpgHome:    gpgHome,
		primaryFpr: primaryFpr,
		sub1Fpr:    sub1Fpr,
		sub2Fpr:    sub2Fpr,
		signSubFpr: signSubFpr,
	}
}

func TestSubkeysFindEncryptAndRecipientIDs(t *testing.T) {
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not found in PATH")
	}

	kr := setupSubkeyTestKeyring(t)
	root, gpgHome, primaryFpr, sub1Fpr, sub2Fpr, signSubFpr := kr.root, kr.gpgHome, kr.primaryFpr, kr.sub1Fpr, kr.sub2Fpr, kr.signSubFpr

	t.Setenv("GNUPGHOME", gpgHome)

	ctx := config.NewContextInMemory()
	ctx = gpg.WithAlwaysTrust(ctx, false)
	ctx = gpg.WithUseCache(ctx, false)

	g, err := New(ctx, Config{})
	require.NoError(t, err)

	// 1. 40-char subkey fingerprint with '!' is preserved.
	recps, err := g.FindRecipients(ctx, sub1Fpr+"!")
	require.NoError(t, err)
	assert.Equal(t, []string{sub1Fpr + "!"}, recps)

	// 2. Lowercase 0x-prefixed 16-char subkey ID with '!' is preserved.
	sub1Lower := "0x" + strings.ToLower(sub1Fpr[len(sub1Fpr)-16:]) + "!"
	recps, err = g.FindRecipients(ctx, sub1Lower)
	require.NoError(t, err)
	assert.Equal(t, []string{sub1Lower}, recps)

	// 3. Two forced subkeys of the same primary key are both preserved.
	recps, err = g.FindRecipients(ctx, sub1Fpr+"!", sub2Fpr+"!")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{sub1Fpr + "!", sub2Fpr + "!"}, recps)

	// 4. Signing-only subkey with '!' is rejected and does NOT fall back to primary key.
	recps, err = g.FindRecipients(ctx, signSubFpr+"!")
	require.NoError(t, err)
	assert.Empty(t, recps)

	// 5. Fingerprint preserves subkey fingerprint + '!' for forced subkeys, primary for non-forced.
	assert.Equal(t, sub1Fpr+"!", g.Fingerprint(ctx, sub1Lower))
	assert.Equal(t, primaryFpr, g.Fingerprint(ctx, sub1Fpr))

	// 6. Encrypting to a single forced subkey encrypts ONLY to that subkey.
	ct1, err := g.Encrypt(ctx, []byte("secret-one"), []string{sub1Fpr + "!"})
	require.NoError(t, err)
	ids1, err := g.RecipientIDs(ctx, ct1)
	require.NoError(t, err)
	assert.Equal(t, []string{sub1Fpr}, ids1)

	pt1, err := g.Decrypt(ctx, ct1)
	require.NoError(t, err)
	assert.Equal(t, []byte("secret-one"), pt1)

	// 7. Encrypting to both forced subkeys encrypts to BOTH subkeys.
	ctBoth, err := g.Encrypt(ctx, []byte("secret-both"), []string{sub1Fpr + "!", sub2Fpr + "!"})
	require.NoError(t, err)
	idsBoth, err := g.RecipientIDs(ctx, ctBoth)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{sub1Fpr, sub2Fpr}, idsBoth)

	// 8. Encrypting to a signing-only forced subkey fails.
	_, err = g.Encrypt(ctx, []byte("nope"), []string{signSubFpr + "!"})
	require.Error(t, err)

	// 9. Offline primary key (sec#) with only sub1 secret subkey can still find identity & decrypt.
	offlineHome := filepath.Join(root, "offline")
	require.NoError(t, os.MkdirAll(offlineHome, 0o700))
	offlineGPGHome := gpgHomeEnv(offlineHome, runtime.GOOS)
	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", offlineGPGHome, "--kill", "gpg-agent").Run()
	})

	expSecCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--export-secret-subkeys", sub1Fpr+"!")
	sub1Sec, err := expSecCmd.Output()
	require.NoError(t, err)

	expTrustCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", gpgHome, "--export-ownertrust")
	otrust, err := expTrustCmd.Output()
	require.NoError(t, err)

	impSecCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", offlineGPGHome, "--import")
	impSecCmd.Stdin = strings.NewReader(string(sub1Sec))
	impOut, err := impSecCmd.CombinedOutput()
	require.NoError(t, err, "import secret subkey failed: %s", string(impOut))

	impTrustCmd := exec.CommandContext(t.Context(), "gpg", "--homedir", offlineGPGHome, "--import-ownertrust")
	impTrustCmd.Stdin = strings.NewReader(string(otrust))
	impOut, err = impTrustCmd.CombinedOutput()
	require.NoError(t, err, "import ownertrust failed: %s", string(impOut))

	t.Setenv("GNUPGHOME", offlineGPGHome)
	gOffline, err := New(ctx, Config{})
	require.NoError(t, err)

	ids, err := gOffline.FindIdentities(ctx, sub1Fpr+"!")
	require.NoError(t, err)
	assert.Equal(t, []string{"0x" + primaryFpr[len(primaryFpr)-16:]}, ids)

	ptOffline, err := gOffline.Decrypt(ctx, ct1)
	require.NoError(t, err)
	assert.Equal(t, []byte("secret-one"), ptOffline)
}
