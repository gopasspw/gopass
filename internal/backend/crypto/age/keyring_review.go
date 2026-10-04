package age

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/termio"
)

// keyringReview is trusted machine-local state, not an integrity guarantee
// against an attacker who can also change this record or the running program.
// Binding it to ciphertext prevents stale approval after keyring replacement.
type keyringReview struct {
	Version          int      `json:"version"`
	Recipients       []string `json:"recipients"`
	CiphertextSHA256 string   `json:"ciphertext_sha256"`
}

func keyringDigest(ciphertext []byte) string {
	digest := sha256.Sum256(ciphertext)

	return hex.EncodeToString(digest[:])
}

func normalizedKeyringRecipients(ctx context.Context) []string {
	recipients := strings.Split(config.String(ctx, "age.keyring-recipients"), ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}
	slices.Sort(recipients)

	return slices.Compact(recipients)
}

// readKeyringReview only accepts a record for the current recipient-encrypted
// ciphertext. Missing or invalid records require review rather than guessing
// recipients from age's recipient stanzas, which need not expose public keys.
func readKeyringReview(filename string) *keyringReview {
	data, err := os.ReadFile(filename + ".review.json")
	if err != nil {
		return nil
	}
	var review keyringReview
	if json.Unmarshal(data, &review) != nil || review.Version != 1 || len(review.Recipients) == 0 {
		return nil
	}
	ciphertext, err := os.ReadFile(filename)
	if err != nil || keyringDigest(ciphertext) != review.CiphertextSHA256 {
		return nil
	}
	passphrase, err := keyringUsesPassphrase(ciphertext)
	if err != nil || passphrase {
		return nil
	}
	if !slices.IsSorted(review.Recipients) {
		return nil
	}
	for i, recipient := range review.Recipients {
		if recipient == "" || strings.TrimSpace(recipient) != recipient || (i > 0 && recipient == review.Recipients[i-1]) {
			return nil
		}
	}

	return &review
}

// reviewKeyringRecipients runs at the shared write boundary, including identity
// mutations. Automatic yes must not authorize a new set of keyring decryptors.
func reviewKeyringRecipients(ctx context.Context, filename string, recipients []string) error {
	previous := readKeyringReview(filename)
	if previous != nil && slices.Equal(previous.Recipients, recipients) {
		return nil
	}
	if !ctxutil.IsInteractive(ctx) || !ctxutil.IsTerminal(ctx) {
		return fmt.Errorf("keyring recipients require manual review in an interactive terminal before writing")
	}
	var message strings.Builder
	message.WriteString("Identity keyring recipients:\n")
	for _, recipient := range recipients {
		fmt.Fprintf(&message, "  %s\n", recipient)
	}
	if previous != nil {
		for _, recipient := range recipients {
			if !slices.Contains(previous.Recipients, recipient) {
				fmt.Fprintf(&message, "Added: %s\n", recipient)
			}
		}
		for _, recipient := range previous.Recipients {
			if !slices.Contains(recipients, recipient) {
				fmt.Fprintf(&message, "Removed: %s\n", recipient)
			}
		}
	}
	message.WriteString("Any corresponding private identity can unlock the keyring and access its software identities.\nProtect the keyring with these recipients?")
	// AskForConfirmation normally honors automatic yes. Override it only for
	// this security-sensitive review; cancellation defaults to preserving the file.
	if !termio.AskForConfirmation(ctxutil.WithAlwaysYes(ctx, false), message.String()) {
		return fmt.Errorf("keyring recipient review cancelled: %w", termio.ErrAborted)
	}

	return nil
}

func saveKeyringReview(filename string, recipients []string, ciphertext []byte) error {
	review := keyringReview{Version: 1, Recipients: recipients, CiphertextSHA256: keyringDigest(ciphertext)}
	data, err := json.Marshal(review)
	if err != nil {
		return fmt.Errorf("failed to encode keyring review: %w", err)
	}
	// Reuse the same atomic, restricted-permission file writer. This record
	// contains public recipients and a ciphertext hash, never private identities.
	if err := writeEncryptedKeyring(filename+".review.json", data); err != nil {
		return fmt.Errorf("keyring updated, but failed to save recipient review; the next write requires manual review: %w", err)
	}

	return nil
}
