package age

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"filippo.io/age"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Encrypt will encrypt the given payload.
func (a *Age) Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error) {
	// add our own public keys to the recipients to ensure we can decrypt it later.
	idRecps, err := a.IdentityRecipients(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch identity recipients for encryption: %w", err)
	}
	// parse the most specific recipients file and add it to the final
	// recipients, too.
	recp, err := a.parseRecipients(ctx, recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recipients file for encryption: %w", err)
	}

	recp = dedupe(append(recp, idRecps...))

	return a.encrypt(plaintext, recp...)
}

// dedupe the recipients, only works for native age recipients.
func dedupe(recp []age.Recipient) []age.Recipient {
	out := make([]age.Recipient, 0, len(recp))
	set := make(map[string]age.Recipient, len(recp))
	for _, r := range recp {
		k, ok := r.(fmt.Stringer)
		// handle non-native recipients.
		if !ok {
			out = append(out, r)

			continue
		}
		set[k.String()] = r
	}

	for _, r := range set {
		out = append(out, r)
	}
	debug.Log("in: %+v - out: %+v", recp, out)

	return out
}

func (a *Age) encrypt(plaintext []byte, recp ...age.Recipient) ([]byte, error) {
	out := &bytes.Buffer{}
	w, err := age.Encrypt(out, recp...)
	if err != nil {
		return nil, err
	}
	n, err := w.Write(plaintext)
	if err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	debug.Log("Encrypted %d bytes of plaintext to %d bytes of ciphertext for %q", n, out.Len(), recp)

	return out.Bytes(), nil
}

func (a *Age) encryptFile(ctx context.Context, filename string, plaintext []byte, confirm bool) error {
	pw, err := ctxutil.GetPasswordCallback(ctx)(filename, confirm)
	if err != nil {
		return err
	}

	id, err := age.NewScryptRecipient(string(pw))
	if err != nil {
		return err
	}
	buf, err := a.encrypt(plaintext, id)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, buf, 0o600)
}
