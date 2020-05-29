package xc

import (
	"bytes"
	"context"
	"sort"
	"strings"
	"text/template"

	"github.com/gopasspw/gopass/internal/backend/crypto/xc/keyring"
	"github.com/gopasspw/gopass/internal/backend/crypto/xc/xcpb"
	"github.com/gopasspw/gopass/internal/debug"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// RecipientIDs reads the header of the given file and extracts the
// recipients IDs
func (x *XC) RecipientIDs(ctx context.Context, ciphertext []byte) ([]string, error) {
	msg := &xcpb.Message{}
	if err := proto.Unmarshal(ciphertext, msg); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(msg.Header.Recipients))
	for k := range msg.Header.Recipients {
		ids = append(ids, k)
	}
	sort.Strings(ids)
	return ids, nil
}

// ReadNamesFromKey unmarshals the given public key and returns the identities name
func (x *XC) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	pk := &xcpb.PublicKey{}
	if err := proto.Unmarshal(buf, pk); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal public key: %s", err)
	}

	return []string{pk.Identity.Name}, nil
}

// ListRecipients lists all public key IDs
func (x *XC) ListRecipients(ctx context.Context) ([]string, error) {
	return x.pubring.KeyIDs(), nil
}

// ListIdentities lists all private key IDs
func (x *XC) ListIdentities(ctx context.Context) ([]string, error) {
	return x.secring.KeyIDs(), nil
}

// FindRecipients finds all matching public keys
func (x *XC) FindRecipients(ctx context.Context, search ...string) ([]string, error) {
	ids := make([]string, 0, 1)
	candidates, _ := x.ListRecipients(ctx)
	for _, needle := range search {
		for _, fp := range candidates {
			if strings.HasSuffix(fp, needle) {
				ids = append(ids, fp)
			}
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// FindIdentities finds all matching private keys
func (x *XC) FindIdentities(ctx context.Context, search ...string) ([]string, error) {
	ids := make([]string, 0, 1)
	candidates, _ := x.ListIdentities(ctx)
	for _, needle := range search {
		for _, fp := range candidates {
			if strings.HasSuffix(fp, needle) {
				ids = append(ids, fp)
			}
		}
	}
	sort.Strings(ids)
	return ids, nil
}

func (x *XC) findID(id string) *xcpb.Identity {
	if key := x.pubring.Get(id); key != nil {
		return key.Identity
	}
	if key := x.secring.Get(id); key != nil {
		return key.PublicKey.Identity
	}
	return &xcpb.Identity{}
}

// Fingerprint returns the id
func (x *XC) Fingerprint(ctx context.Context, id string) string {
	return id
}

// FormatKey formats a key
func (x *XC) FormatKey(ctx context.Context, id, tpl string) string {
	if tpl == "" {
		tpl = "{{ .ID }} - {{ .Name }} <{{ .Email }}>"
	}

	tmpl, err := template.New(tpl).Parse(tpl)
	if err != nil {
		return ""
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, x.findID(id)); err != nil {
		debug.Log("Failed to render template '%s': %s", tpl, err)
		return ""
	}

	return buf.String()
}

// GenerateIdentity creates a new keypair
func (x *XC) GenerateIdentity(ctx context.Context, name, email, passphrase string) error {
	k, err := keyring.GenerateKeypair(passphrase)
	if err != nil {
		return errors.Wrapf(err, "failed to generate keypair: %s", err)
	}
	k.Identity.Name = name
	k.Identity.Email = email
	if err := x.secring.Set(k); err != nil {
		return errors.Wrapf(err, "failed to set %v to secring: %s", k, err)
	}
	return x.secring.Save()
}
