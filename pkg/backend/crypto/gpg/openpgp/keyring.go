package openpgp

import (
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"

	"github.com/pkg/errors"
)

// ListPublicKeyIDs does nothing
func (g *GPG) ListPublicKeyIDs(context.Context) ([]string, error) {
	if g.pubring == nil {
		return nil, fmt.Errorf("pubring is not initialized")
	}
	ids := listKeyIDs(g.pubring)
	if g.secring != nil {
		ids = append(ids, listKeyIDs(g.secring)...)
	}
	return ids, nil
}

// FindPublicKeys does nothing
func (g *GPG) FindPublicKeys(ctx context.Context, keys ...string) ([]string, error) {
	kl, err := g.ListPublicKeyIDs(ctx)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		if strings.HasPrefix(key, "0x") {
			key = strings.TrimPrefix(key, "0x")
		}
		keys[i] = strings.ToUpper(key)
	}
	matches := make([]string, 0, len(keys))
	for _, key := range kl {
		for _, needle := range keys {
			if strings.HasSuffix(key, needle) {
				matches = append(matches, key)
			}
		}
	}
	if m, err := g.FindPrivateKeys(ctx, keys...); err == nil {
		matches = append(matches, m...)
	}
	return matches, nil
}

// ListPrivateKeyIDs does nothing
func (g *GPG) ListPrivateKeyIDs(context.Context) ([]string, error) {
	if g.secring == nil {
		return nil, fmt.Errorf("secring is not initialized")
	}
	return listKeyIDs(g.secring), nil
}

// FindPrivateKeys does nothing
func (g *GPG) FindPrivateKeys(ctx context.Context, keys ...string) ([]string, error) {
	kl, err := g.ListPrivateKeyIDs(ctx)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		if strings.HasPrefix(key, "0x") {
			key = strings.TrimPrefix(key, "0x")
		}
		keys[i] = strings.ToUpper(key)
	}
	matches := make([]string, 0, len(keys))
	for _, key := range kl {
		for _, needle := range keys {
			if strings.HasSuffix(key, needle) {
				matches = append(matches, key)
			}
		}
	}
	return matches, nil
}

func (g *GPG) savePubring() error {
	pubfnTmp := g.pubfn + ".tmp"
	if err := g.writePubring(pubfnTmp); err != nil {
		os.Remove(pubfnTmp)
		return errors.Wrapf(err, "failed to write pubring")
	}
	return os.Rename(pubfnTmp, g.pubfn)
}

func (g *GPG) writePubring(fn string) error {
	fh, err := os.OpenFile(fn, os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to save pubring, failed to open file: %s", err)
	}
	defer fh.Close()
	for _, e := range g.pubring {
		if err := e.Serialize(fh); err != nil {
			return err
		}
	}
	return nil
}

func (g *GPG) saveSecring() error {
	secfnTmp := g.secfn + ".tmp"
	if err := g.writeSecring(secfnTmp); err != nil {
		os.Remove(secfnTmp)
		return errors.Wrapf(err, "failed to write secring")
	}
	return os.Rename(secfnTmp, g.secfn)
}

func (g *GPG) writeSecring(fn string) error {
	fh, err := os.OpenFile(fn, os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "unable to save secring, failed to open file: %s", err)
	}
	defer fh.Close()
	for _, e := range g.secring {
		if err := e.SerializePrivate(fh, nil); err != nil {
			return err
		}
	}
	return nil
}

// KeysById implements openpgp.Keyring
func (g *GPG) KeysById(id uint64) []openpgp.Key {
	return append(g.secring.KeysById(id), g.pubring.KeysById(id)...)
}

// KeysByIdUsage implements openpgp.Keyring
func (g *GPG) KeysByIdUsage(id uint64, requiredUsage byte) []openpgp.Key {
	return append(g.secring.KeysByIdUsage(id, requiredUsage), g.pubring.KeysByIdUsage(id, requiredUsage)...)
}

// DecryptionKeys implements openpgp.Keyring
func (g *GPG) DecryptionKeys() []openpgp.Key {
	return append(g.secring.DecryptionKeys(), g.pubring.DecryptionKeys()...)
}

// SigningKeys returns a list of signing keys
func (g *GPG) SigningKeys() []openpgp.Key {
	keys := []openpgp.Key{}
	for _, e := range g.secring {
		for _, subKey := range e.Subkeys {
			if subKey.PrivateKey != nil && (!subKey.Sig.FlagsValid || subKey.Sig.FlagSign) {
				keys = append(keys, openpgp.Key{
					Entity:        e,
					PublicKey:     subKey.PublicKey,
					PrivateKey:    subKey.PrivateKey,
					SelfSignature: subKey.Sig,
				})
			}
		}
	}
	return keys
}
