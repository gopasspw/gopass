package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/gopasspw/gopass/internal/backend/crypto/gpg"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/debug"
)

// ListRecipients returns a parsed list of GPG public keys.
func (g *GPG) ListRecipients(ctx context.Context) ([]string, error) {
	if g.pubKeys == nil {
		kl, err := g.listKeys(ctx, "public")
		if err != nil {
			return nil, err
		}
		g.pubKeys = kl
	}

	if gpg.IsAlwaysTrust(ctx) {
		return g.pubKeys.Recipients(), nil
	}

	return g.pubKeys.UseableKeys(gpg.IsAlwaysTrust(ctx)).Recipients(), nil
}

// FindRecipients searches for the given public keys.
func (g *GPG) FindRecipients(ctx context.Context, search ...string) ([]string, error) {
	kl, err := g.listKeys(ctx, "public", search...)
	if err != nil {
		return nil, err
	}
	if kl == nil {
		return nil, fmt.Errorf("no keys found for %v", search)
	}

	recp := kl.UseableKeys(gpg.IsAlwaysTrust(ctx)).Recipients()
	if gpg.IsAlwaysTrust(ctx) {
		recp = kl.Recipients()
	}
	debug.Log("recp before subkey check: %q", recp)

	// let us support the ! syntax that enforces specific subkey usage
	for _, val := range search {
		str := strings.TrimPrefix(strings.TrimSuffix(val, "!"), "0x")
		if !strings.HasSuffix(val, "!") || str == "" {
			continue
		}
		for _, key := range kl {
			for sub := range key.SubKeys {
				if !strings.Contains(sub, str) {
					continue
				}
				// we remove that key from the recp
				recp = slices.DeleteFunc(recp, func(s string) bool {
					s = strings.TrimPrefix(s, "0x")
					// because we use short fingerprints in recp
					return strings.Contains(key.Fingerprint, s)
				})
				// and we add the specific subkey as is, keeping the !
				recp = append(recp, val)

				// we go to the next key in the odd case the subkey is part of multiple keys
				break
			}
		}
	}

	debug.Log("found useable keys for %q: %q (all: %q)", search, recp, kl.Recipients())

	return recp, nil
}

// RecipientIDs returns a list of recipient IDs for a given encrypted blob.
func (g *GPG) RecipientIDs(ctx context.Context, buf []byte) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	recp := make([]string, 0, 5)

	// extract recipients from gpg output
	args := []string{"--batch", "--list-only", "--list-packets", "--no-default-keyring", "--secret-keyring", "/dev/null"}
	cmd := exec.CommandContext(ctx, g.binary, args...)
	cmd.Stdin = bytes.NewReader(buf)

	// switch to LANG C for more predictable output, switch back later
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "LANGUAGE=") {
			continue
		}
		cmd.Env = append(cmd.Env, env)
	}
	cmd.Env = append(cmd.Env, "LANGUAGE=C")

	debug.Log("%s %+v", cmd.Path, cmd.Args)

	cmdout, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	// parse the output
	scanner := bufio.NewScanner(bytes.NewBuffer(cmdout))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		debug.Log("GPG Output: %s", line)
		if !strings.HasPrefix(line, ":pubkey enc packet:") {
			continue
		}

		m := splitPacket(line)
		keyid, found := m["keyid"]
		if !found {
			continue
		}

		kl, err := g.listKeys(ctx, "public", keyid)
		if err != nil || len(kl) < 1 {
			continue
		}

		recp = append(recp, kl[0].Fingerprint)
	}

	if g.throwKids {
		out.Warningf(ctx, "gpg option throw-keyids is set. some features might not work.")
	}

	return recp, nil
}

func splitPacket(in string) map[string]string {
	m := make(map[string]string, 3)
	p := strings.Split(in, ":")
	if len(p) < 3 {
		return m
	}

	p = strings.Split(strings.TrimSpace(p[2]), " ")
	for i := 0; i+1 < len(p); i += 2 {
		m[p[i]] = strings.Trim(p[i+1], ",")
	}

	return m
}
