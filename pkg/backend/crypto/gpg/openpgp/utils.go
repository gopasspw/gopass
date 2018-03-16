package openpgp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/justwatchcom/gopass/pkg/out"
	homedir "github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/openpgp"
)

type agentClient interface {
	Ping() error
	Passphrase(string, string) (string, error)
	Remove(string) error
}

var maxUnlockAttempts = 3

func (g *GPG) mkPromptFunc() func([]openpgp.Key, bool) ([]byte, error) {
	attempt := 0
	return func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		attempt++
		if attempt > maxUnlockAttempts {
			return nil, fmt.Errorf("out of retries")
		}
		for i, key := range keys {
			if key.PublicKey == nil || key.PrivateKey == nil {
				continue
			}
			fp := key.PublicKey.KeyIdString()
			passphrase, err := g.client.Passphrase(fp, fmt.Sprintf("Unlock private key %s", fp))
			if err != nil {
				continue
			}
			if err := keys[i].PrivateKey.Decrypt([]byte(passphrase)); err == nil {
				return []byte(passphrase), nil
			}
			if err := g.client.Remove(fp); err != nil {
				return nil, err
			}
			time.Sleep(10 * time.Millisecond)
		}
		return nil, nil
	}
}

func (g *GPG) findEntity(id string) *openpgp.Entity {
	return g.findEntityInLists(id, g.secring, g.pubring)
}

func (g *GPG) findEntityInLists(id string, els ...openpgp.EntityList) *openpgp.Entity {
	id = strings.TrimPrefix(id, "0x")
	for _, el := range els {
		for _, ent := range el {
			if ent.PrimaryKey == nil {
				continue
			}
			fp := fmt.Sprintf("%X", ent.PrimaryKey.Fingerprint)
			if strings.HasSuffix(fp, id) {
				return ent
			}
		}
	}
	return nil
}

func (g *GPG) recipientsToEntities(recipients []string) []*openpgp.Entity {
	ents := make([]*openpgp.Entity, 0, len(recipients))
	for _, key := range g.pubring {
		if key.PrimaryKey == nil {
			continue
		}
		fp := fmt.Sprintf("%X", key.PrimaryKey.Fingerprint)
		for _, recp := range recipients {
			recp = strings.TrimPrefix(recp, "0x")
			if strings.HasSuffix(fp, recp) {
				ents = append(ents, key)
			}
		}
	}
	return ents
}

func listKeyIDs(el openpgp.EntityList) []string {
	ids := make([]string, 0, len(el))
	for _, key := range el {
		if key.PrimaryKey == nil {
			continue
		}
		ids = append(ids, key.PrimaryKey.KeyIdString())
	}
	return ids
}

func readKeyring(fn string) (openpgp.EntityList, error) {
	fh, err := os.Open(fn)
	if err != nil {
		if os.IsNotExist(err) {
			return openpgp.EntityList{}, nil
		}
		return nil, err
	}
	defer fh.Close()

	return openpgp.ReadKeyRing(fh)
}

// gpgHome returns the gnupg homedir
func gpgHome(ctx context.Context) string {
	if gh := os.Getenv("GNUPGHOME"); gh != "" {
		return gh
	}
	hd, err := homedir.Dir()
	if err != nil {
		out.Debug(ctx, "Failed to get homedir: %s", err)
		return ""
	}
	return filepath.Join(hd, ".gnupg")
}
