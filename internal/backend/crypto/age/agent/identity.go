package agent

import (
	"fmt"
	"io"
	"strings"

	"filippo.io/age"
	"filippo.io/age/plugin"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/identityfile"
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

// parseIdentities parses native and plugin identities using the shared bounded
// identity-file scanner and the agent-specific single-line parser.
func parseIdentities(reader io.Reader) ([]age.Identity, error) {
	return identityfile.Parse(reader, parseIdentity)
}
