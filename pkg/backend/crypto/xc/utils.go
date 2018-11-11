package xc

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/keyring"
	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/xcpb"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
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

// ListPublicKeyIDs lists all public key IDs
func (x *XC) ListPublicKeyIDs(ctx context.Context) ([]string, error) {
	return x.pubring.KeyIDs(), nil
}

// ListPrivateKeyIDs lists all private key IDs
func (x *XC) ListPrivateKeyIDs(ctx context.Context) ([]string, error) {
	return x.secring.KeyIDs(), nil
}

// FindPublicKeys finds all matching public keys
func (x *XC) FindPublicKeys(ctx context.Context, search ...string) ([]string, error) {
	ids := make([]string, 0, 1)
	candidates, _ := x.ListPublicKeyIDs(ctx)
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

// FindPrivateKeys finds all matching private keys
func (x *XC) FindPrivateKeys(ctx context.Context, search ...string) ([]string, error) {
	ids := make([]string, 0, 1)
	candidates, _ := x.ListPrivateKeyIDs(ctx)
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

// FormatKey formats a key
func (x *XC) FormatKey(ctx context.Context, id string) string {
	if key := x.pubring.Get(id); key != nil {
		return id + " - " + key.Identity.ID()
	}
	if key := x.secring.Get(id); key != nil {
		return id + " - " + key.PublicKey.Identity.ID()
	}
	return id
}

// NameFromKey extracts the name from a key
func (x *XC) NameFromKey(ctx context.Context, id string) string {
	if key := x.pubring.Get(id); key != nil {
		return key.Identity.Name
	}
	if key := x.secring.Get(id); key != nil {
		return key.PublicKey.Identity.Name
	}
	return id
}

// EmailFromKey extracts the email from a key
func (x *XC) EmailFromKey(ctx context.Context, id string) string {
	if key := x.pubring.Get(id); key != nil {
		return key.Identity.Email
	}
	if key := x.secring.Get(id); key != nil {
		return key.PublicKey.Identity.Email
	}
	return id
}

// Fingerprint returns the full-length native fingerprint
func (x *XC) Fingerprint(ctx context.Context, id string) string {
	return id
}

// CreatePrivateKeyBatch creates a new keypair
func (x *XC) CreatePrivateKeyBatch(ctx context.Context, name, email, passphrase string) error {
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

// CreatePrivateKey is not implemented
func (x *XC) CreatePrivateKey(ctx context.Context) error {
	return fmt.Errorf("not yet implemented")
}
