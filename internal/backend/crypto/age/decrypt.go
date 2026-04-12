package age

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/agent"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Decrypt will attempt to decrypt the given payload.
func (a *Age) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if config.Bool(ctx, "age.agent-enabled") {
		plaintext, err := a.decryptWithAgent(ctx, ciphertext)
		if err == nil {
			return plaintext, nil
		}
		debug.Log("failed to decrypt with agent: %s", err)
		debug.Log("falling back to direct decryption")
	}

	ids, err := a.getAllIds(ctx)
	if err != nil {
		return nil, err
	}

	return a.decrypt(ciphertext, ids...)
}

func (a *Age) decryptWithAgent(ctx context.Context, ciphertext []byte) ([]byte, error) {
	client := agent.NewClient()
	plaintext, err := client.Decrypt(ciphertext)
	if err == nil {
		return plaintext, nil
	}

	if !strings.Contains(err.Error(), "agent is locked") {
		debug.Log("failed to decrypt with agent: %s", err)

		return nil, err
	}

	debug.Log("agent is locked, trying to unlock")
	// unlock the agent
	if err := client.Unlock(); err != nil {
		debug.Log("failed to unlock agent: %s", err)
	}
	// get identities
	ids, err := a.getAllIds(ctx)
	if err != nil {
		return nil, err
	}
	// send identities to agent
	sIds, err := a.identitiesToString(ids)
	if err != nil {
		return nil, err
	}
	if err := client.SendIdentities(sIds); err != nil {
		debug.Log("failed to send identities to agent: %s", err)
	}
	// set timeout
	if timeout := config.AsInt(config.String(ctx, "age.agent-timeout")); timeout > 0 {
		if err := client.SetTimeout(timeout); err != nil {
			debug.Log("failed to set agent timeout: %s", err)
		}
	}
	// retry decryption
	return client.Decrypt(ciphertext)
}

func (a *Age) decrypt(ciphertext []byte, ids ...age.Identity) ([]byte, error) {
	debug.V(1).Log("decrypting with %d ids", len(ids))

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
	debug.V(1).Log("Decrypted %d bytes of ciphertext to %d bytes of plaintext", len(ciphertext), n)

	return out.Bytes(), nil
}

// decryptFile is used to decrypt a scrypt encrypted age keyring/identity file.
// pwcb is called to obtain the passphrase; ppcb is invoked on a decrypt failure
// so cached passwords can be invalidated.
func (a *Age) decryptFile(_ context.Context, filename string, pwcb func(string, bool) ([]byte, error), ppcb func(string)) ([]byte, error) {
	ciphertext, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	debug.V(1).Log("read %d bytes from %s", len(ciphertext), filename)

	pw, err := pwcb(filename, false)
	if err != nil {
		return nil, err
	}

	// debug.Log("deriving scrypt identity from password: %q", out.Secret(pw))
	id, err := age.NewScryptIdentity(string(pw))
	if err != nil {
		return nil, err
	}

	plaintext, err := a.decrypt(ciphertext, id)
	if err != nil {
		ppcb(filename)
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
