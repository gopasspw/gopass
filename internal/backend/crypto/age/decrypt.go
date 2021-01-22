package age

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"filippo.io/age"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Decrypt will attempt to decrypt the given payload
func (a *Age) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	if !ctxutil.HasPasswordCallback(ctx) {
		debug.Log("no password callback found, redirecting to askPass")
		ctx = ctxutil.WithPasswordCallback(ctx, func(prompt string, _ bool) ([]byte, error) {
			pw, err := a.askPass.Passphrase(prompt, fmt.Sprintf("to load the keyring at %s", a.keyring), false)
			return []byte(pw), err
		})
	}
	ids, err := a.getAllIds(ctx)
	if err != nil {
		return nil, err
	}
	return a.decrypt(ciphertext, ids...)
}

func (a *Age) decrypt(ciphertext []byte, ids ...age.Identity) ([]byte, error) {
	out := &bytes.Buffer{}
	f := bytes.NewReader(ciphertext)
	r, err := age.Decrypt(f, ids...)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (a *Age) decryptFile(ctx context.Context, filename string) ([]byte, error) {
	ciphertext, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	pw, err := ctxutil.GetPasswordCallback(ctx)(filename, false)
	if err != nil {
		return nil, err
	}
	id, err := age.NewScryptIdentity(string(pw))
	if err != nil {
		return nil, err
	}
	return a.decrypt(ciphertext, id)
}
