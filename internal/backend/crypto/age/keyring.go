package age

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/termio"
)

// Keyring is an age keyring
type Keyring []Keypair

// Keypair is a public / private keypair
type Keypair struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Identity string `json:"identity"`
}

func (a *Age) pkself(ctx context.Context) (age.Recipient, error) {
	kr, err := a.loadKeyring()

	var id *age.X25519Identity
	if err != nil || len(kr) < 1 {
		id, err = a.genKey(ctx)
	} else {
		id, err = age.ParseX25519Identity(kr[len(kr)-1].Identity)
	}
	if err != nil {
		return nil, err
	}
	return id.Recipient(), nil
}

func (a *Age) genKey(ctx context.Context) (*age.X25519Identity, error) {
	debug.Log("No native age key found. Generating ...")
	id, err := a.generateIdentity(termio.DetectName(ctx, nil), termio.DetectEmail(ctx, nil))
	if err != nil {
		return nil, err
	}
	return id, nil
}

// GenerateIdentity will create a new native private key
func (a *Age) GenerateIdentity(ctx context.Context, name, email, _ string) error {
	_, err := a.generateIdentity(name, email)
	return err
}

func (a *Age) generateIdentity(name, email string) (*age.X25519Identity, error) {
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return id, err
	}

	kr, err := a.loadKeyring()
	if err != nil {
		debug.Log("Warning: Failed to load keyring from %s: %s", a.keyring, err)
	}

	kr = append(kr, Keypair{
		Name:     name,
		Email:    email,
		Identity: id.String(),
	})

	return id, a.saveKeyring(kr)
}

func (a *Age) loadKeyring() (Keyring, error) {
	kr := make(Keyring, 1)
	buf, err := a.decryptFile(a.keyring)
	if err != nil {
		debug.Log("can't decrypt keyring at %s: %s", a.keyring, err)
		return kr, err
	}
	if err := json.Unmarshal(buf, &kr); err != nil {
		debug.Log("can't parse keyring at %s: %s", a.keyring, err)
		return kr, err
	}
	// remove invalid IDs
	valid := make(Keyring, 0, len(kr))
	for _, k := range kr {
		if k.Identity == "" {
			continue
		}
		valid = append(valid, k)
	}
	debug.Log("loaded keyring from %s", a.keyring)
	return valid, nil
}

func (a *Age) saveKeyring(k Keyring) error {
	if err := os.MkdirAll(filepath.Dir(a.keyring), 0700); err != nil {
		return err
	}

	// encrypt final keyring
	buf, err := json.Marshal(k)
	if err != nil {
		return err
	}
	if err := a.encryptFile(a.keyring, buf); err != nil {
		return err
	}
	debug.Log("saved encrypted keyring to %s", a.keyring)
	return nil
}
