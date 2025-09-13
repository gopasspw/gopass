package age

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Decrypt will attempt to decrypt the given payload.
func (a *Age) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if config.Bool(ctx, "age.agent-enabled") {
		client := agent.NewClient()
		plaintext, err := client.Decrypt(ciphertext)
		if err == nil {
			return plaintext, nil
		}
		debug.Log("failed to decrypt with agent: %s", err)
	}

	if !ctxutil.HasPasswordCallback(ctx) {
		debug.Log("no password callback found, redirecting to askPass")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
			pw, err := a.askPass.Passphrase(prompt, fmt.Sprintf("to load the keyring at %s", a.identity), false)

			return []byte(pw), err
		})
		ctx = ctxutil.WithPasswordPurgeCallback(ctx, a.askPass.Remove)
	}

	ids, err := a.getAllIds(ctx)
	if err != nil {
		return nil, err
	}

	return a.decrypt(ciphertext, ids...)
}

func (a *Age) decrypt(ciphertext []byte, ids ...age.Identity) ([]byte, error) {
	debug.Log("decrypting with %d ids", len(ids))

	out := &bytes.Buffer{}
	f := bytes.NewReader(ciphertext)
	r, err := age.Decrypt(f, ids...)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}
	n, err := io.Copy(out, r)
	if err != nil {
		return nil, fmt.Errorf("failed to write plaintext to buffer: %w", err)
	}
	debug.Log("Decrypted %d bytes of ciphertext to %d bytes of plaintext", len(ciphertext), n)

	return out.Bytes(), nil
}

// decryptFile is used to decrypt a scrypt encrypted age keyring/identity file.
func (a *Age) decryptFile(ctx context.Context, filename string) ([]byte, error) {
	ciphertext, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	debug.Log("read %d bytes from %s", len(ciphertext), filename)

	pw, err := ctxutil.GetPasswordCallback(ctx)(filename, false)
	if err != nil {
		return nil, err
	}

	id, err := age.NewScryptIdentity(string(pw))
	if err != nil {
		return nil, err
	}

	plaintext, err := a.decrypt(ciphertext, id)
	if err != nil {
		ctxutil.GetPasswordPurgeCallback(ctx)(filename)
	}

	return plaintext, err
}

func (a *Age) getAllIds(ctx context.Context) ([]age.Identity, error) {
	ids, err := a.getAllIdentities(ctx)
	if err != nil {
		return nil, err
	}
	idl := make([]age.Identity, 0, len(ids))
	for _, id := range ids {
		idl = append(idl, id)
	}

	return idl, nil
}
