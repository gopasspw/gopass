package age

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	"filippo.io/age"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

// Encrypt will encrypt the given payload
func (a *Age) Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error) {
	// add our own public key
	pks, err := a.pkself(ctx)
	if err != nil {
		return nil, err
	}
	recp, err := a.parseRecipients(ctx, recipients)
	if err != nil {
		return nil, err
	}
	recp = append(recp, pks)
	recp = dedupe(recp)
	return a.encrypt(plaintext, recp...)
}

// dedupe the recipients, only works for native age recipients
func dedupe(recp []age.Recipient) []age.Recipient {
	out := make([]age.Recipient, 0, len(recp))
	set := make(map[string]age.Recipient, len(recp))
	for _, r := range recp {
		k, ok := r.(fmt.Stringer)
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
	debug.Log("Wrote %d bytes of plaintext for %+v", n, recp)
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

	return ioutil.WriteFile(filename, buf, 0600)
}
