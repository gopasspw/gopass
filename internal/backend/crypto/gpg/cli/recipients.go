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

	alwaysTrust := gpg.IsAlwaysTrust(ctx)
	useable := kl.UseableKeys(alwaysTrust)
	if alwaysTrust {
		useable = kl
	}
	recp := useable.Recipients()
	debug.Log("recp before subkey check: %q", recp)

	// let us support the ! syntax that enforces specific subkey usage
	for _, val := range search {
		if !strings.HasSuffix(val, "!") {
			continue
		}
		str := strings.ToUpper(strings.TrimPrefix(strings.TrimPrefix(strings.TrimSuffix(val, "!"), "0x"), "0X"))
		if str == "" {
			continue
		}
		for _, key := range useable {
			sk, ok := key.MatchSubKey(val)
			if !ok {
				continue
			}
			// we remove the primary key from recp (even if the specific subkey
			// is unusable, we must not silently fall back to the primary key)
			recp = slices.DeleteFunc(recp, func(s string) bool {
				if strings.HasSuffix(s, "!") {
					return false
				}
				s = strings.ToUpper(strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X"))
				// because we use short fingerprints in recp
				return s != "" && strings.Contains(strings.ToUpper(key.Fingerprint), s)
			})
			if !sk.IsUseable(alwaysTrust) {
				debug.Log("subkey %q is not useable for encryption (validity=%q, caps=%+v, exp=%v)", val, sk.Validity, sk.Caps, sk.ExpirationDate)

				break
			}
			// and we add the specific subkey as is, keeping the !
			if !slices.Contains(recp, val) {
				recp = append(recp, val)
			}

			// we go to the next key in the odd case the subkey is part of multiple keys
			break
		}
	}

	debug.Log("found useable keys for %q: %q (all: %q)", search, recp, kl.Recipients())

	return recp, nil
}

// RecipientIDs returns a list of recipient IDs for a given encrypted blob.
func (g *GPG) RecipientIDs(ctx context.Context, buf []byte) ([]string, error) {
	ctx, cancel := withTimeout(ctx, timeoutFromContext(ctx))
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

		if sk, ok := kl[0].MatchSubKey(keyid); ok && sk.Fingerprint != "" {
			recp = append(recp, sk.Fingerprint)
		} else {
			recp = append(recp, kl[0].Fingerprint)
		}
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
