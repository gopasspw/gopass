package agent

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
	"filippo.io/age/plugin"
)

// parseIdentity parses a single identity string, supporting AGE-PLUGIN-* prefixed plugin
// identities in addition to native age keys. It mirrors the parsing logic in the parent
// age package but does not need the wrappedIdentity/wrappedRecipient types since the agent
// only performs decryption, not encryption.
//
// It supports gopass's custom format: `<identity>"|"<recipient>` — the recipient suffix is
// simply stripped before parsing.
func parseIdentity(s string) (age.Identity, error) {
	switch {
	case strings.HasPrefix(s, "AGE-PLUGIN-"):
		sp := strings.Split(s, "|")
		id, err := plugin.NewIdentity(sp[0], nil)
		if err != nil {
			return nil, fmt.Errorf("unable to parse plugin identity: %w", err)
		}

		return id, nil
	case strings.HasPrefix(s, "AGE-SECRET-KEY-PQ-1"):
		sp := strings.Split(s, "|")

		return age.ParseHybridIdentity(sp[0])
	case strings.HasPrefix(s, "AGE-SECRET-KEY-1"):
		sp := strings.Split(s, "|")

		return age.ParseX25519Identity(sp[0])
	default:
		return nil, fmt.Errorf("unknown identity type: %.12s", s)
	}
}

// parseIdentities parses multiple age identities from a reader, supporting plugin identities
// in addition to native age keys. It replaces age.ParseIdentities() which only handles
// native keys and rejects AGE-PLUGIN-* identities (causing "malformed secret key: mixed
// case" errors with e.g. age-plugin-yubikey).
func parseIdentities(f io.Reader) ([]age.Identity, error) {
	const privateKeySizeLimit = 1 << 24 // 16 MiB
	var ids []age.Identity
	scanner := bufio.NewScanner(io.LimitReader(f, privateKeySizeLimit))
	var n int
	for scanner.Scan() {
		n++
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		i, err := parseIdentity(line)
		if err != nil {
			return nil, fmt.Errorf("error at line %d: %w", n, err)
		}
		ids = append(ids, i)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read secret keys file: %w", err)
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no secret keys found")
	}

	return ids, nil
}
