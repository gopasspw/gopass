package cli

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Decrypt will try to decrypt the given file
func (g *GPG) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	args := append(g.args, "--decrypt")
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	return cmd.Output()
}

// RecipientIDs returns a list of recipient IDs for a given file
func (g *GPG) RecipientIDs(ctx context.Context, buf []byte) ([]string, error) {
	_ = os.Setenv("LANGUAGE", "C")
	recp := make([]string, 0, 5)

	args := []string{"--batch", "--list-only", "--list-packets", "--no-default-keyring", "--secret-keyring", "/dev/null"}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf)
	debug.Log("%s %+v", cmd.Path, cmd.Args)

	cmdout, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(cmdout))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		debug.Log("GPG Output: %s", line)
		if !strings.HasPrefix(line, ":pubkey enc packet:") {
			continue
		}
		m := splitPacket(line)
		if keyid, found := m["keyid"]; found {
			kl, err := g.listKeys(ctx, "public", keyid)
			if err != nil || len(kl) < 1 {
				continue
			}
			recp = append(recp, kl[0].Fingerprint)
		}
	}

	if g.throwKids {
		// TODO shouldn't log here
		out.Warningf(ctx, "gpg option throw-keyids is set. some features might not work.")
	}
	return recp, nil
}
