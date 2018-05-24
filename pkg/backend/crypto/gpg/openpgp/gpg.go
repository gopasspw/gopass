package openpgp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
	"golang.org/x/crypto/openpgp/packet"

	"github.com/blang/semver"
	"github.com/pkg/errors"
)

// GPG is a pure-Go GPG backend
type GPG struct {
	pubfn   string
	pubring openpgp.EntityList
	secfn   string
	secring openpgp.EntityList
	client  agentClient
}

// New creates a new pure-Go GPG backend
func New(ctx context.Context) (*GPG, error) {
	pubfn := filepath.Join(gpgHome(ctx), "pubring.gpg")
	pubring, err := readKeyring(pubfn)
	if err != nil {
		return nil, err
	}
	secfn := filepath.Join(gpgHome(ctx), "secring.gpg")
	secring, err := readKeyring(secfn)
	if err != nil {
		return nil, err
	}
	g := &GPG{
		pubring: pubring,
		secring: secring,
		pubfn:   pubfn,
		secfn:   secfn,
	}
	return g, nil
}

// RecipientIDs returns the recipients of the encrypted message
func (g *GPG) RecipientIDs(ctx context.Context, ciphertext []byte) ([]string, error) {
	recps := make([]string, 0, 1)
	packets := packet.NewReader(bytes.NewReader(ciphertext))
	for {
		p, err := packets.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch p := p.(type) {
		case *packet.EncryptedKey:
			for _, key := range g.pubring {
				if key.PrimaryKey == nil {
					continue
				}
				if key.PrimaryKey.KeyId == p.KeyId {
					recps = append(recps, key.PrimaryKey.KeyIdString())
				}
			}
		}
	}
	return recps, nil
}

// Encrypt encrypts the plaintext for the given recipients
func (g *GPG) Encrypt(ctx context.Context, plaintext []byte, recipients []string) ([]byte, error) {
	ciphertext := &bytes.Buffer{}
	ents := g.recipientsToEntities(recipients)
	wc, err := openpgp.Encrypt(ciphertext, ents, nil, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encrypt")
	}
	if _, err := io.Copy(wc, bytes.NewReader(plaintext)); err != nil {
		return nil, errors.Wrapf(err, "failed to write plaintext to encoder")
	}
	if err := wc.Close(); err != nil {
		return nil, errors.Wrapf(err, "failed to finalize encryption")
	}
	return ciphertext.Bytes(), nil
}

// Decrypt decryptes the ciphertext
// see https://gist.github.com/stuart-warren/93750a142d3de4e8fdd2
func (g *GPG) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	md, err := openpgp.ReadMessage(bytes.NewReader(ciphertext), g, g.mkPromptFunc(), nil)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, md.UnverifiedBody); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExportPublicKey does nothing
func (g *GPG) ExportPublicKey(ctx context.Context, id string) ([]byte, error) {
	ent := g.findEntity(id)
	if ent == nil {
		return nil, fmt.Errorf("key not found")
	}

	buf := &bytes.Buffer{}
	err := ent.PrimaryKey.Serialize(buf)
	return buf.Bytes(), err
}

// ImportPublicKey does nothing
func (g *GPG) ImportPublicKey(ctx context.Context, buf []byte) error {
	el, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(buf))
	if err != nil {
		return err
	}
	g.pubring = append(g.pubring, el...)
	return nil
}

// Version returns dummy version info
func (g *GPG) Version(context.Context) semver.Version {
	return semver.Version{Major: 1}
}

// Binary always returns ''
func (g *GPG) Binary() string {
	return ""
}

// Sign is not implemented
func (g *GPG) Sign(ctx context.Context, in string, sigf string) error {
	signKeys := g.SigningKeys()
	if len(signKeys) < 1 {
		return fmt.Errorf("no signing keys available")
	}

	sigfh, err := os.OpenFile(sigf, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer sigfh.Close()

	wc, err := clearsign.Encode(sigfh, signKeys[0].PrivateKey, nil)
	if err != nil {
		return err
	}
	infh, err := os.Open(in)
	if err != nil {
		return err
	}
	defer infh.Close()

	if _, err := io.Copy(wc, infh); err != nil {
		return err
	}
	return wc.Close()
}

// Verify is not implemented
func (g *GPG) Verify(ctx context.Context, sigf string, in string) error {
	sig, err := ioutil.ReadFile(sigf)
	if err != nil {
		return err
	}
	b, _ := clearsign.Decode(sig)
	infh, err := os.Open(in)
	if err != nil {
		return err
	}
	defer infh.Close()
	_, err = openpgp.CheckDetachedSignature(g.pubring, infh, bytes.NewReader(b.Bytes))
	if err != nil {
		return err
	}
	return nil
}

// CreatePrivateKey is not implemented
func (g *GPG) CreatePrivateKey(ctx context.Context) error {
	return fmt.Errorf("not yet implemented")
}

// CreatePrivateKeyBatch is not implemented
func (g *GPG) CreatePrivateKeyBatch(ctx context.Context, name, email, pw string) error {
	ent, err := openpgp.NewEntity(name, "", email, &packet.Config{
		RSABits: 4096,
	})
	if err != nil {
		return err
	}
	g.secring = append(g.secring, ent)
	return g.saveSecring()
}

// EmailFromKey returns the email for this key
func (g *GPG) EmailFromKey(ctx context.Context, id string) string {
	ent := g.findEntity(id)
	if ent == nil || ent.Identities == nil {
		return ""
	}
	for name, id := range ent.Identities {
		if id.UserId == nil {
			return name
		}
		return id.UserId.Email
	}
	return ""
}

// NameFromKey is returns the name for this key
func (g *GPG) NameFromKey(ctx context.Context, id string) string {
	ent := g.findEntity(id)
	if ent == nil || ent.Identities == nil {
		return ""
	}
	for name, id := range ent.Identities {
		if id.UserId == nil {
			return name
		}
		return id.UserId.Name
	}
	return ""
}

// FormatKey returns the id
func (g *GPG) FormatKey(ctx context.Context, id string) string {
	ent := g.findEntity(id)
	if ent == nil || ent.Identities == nil {
		return ""
	}
	for name := range ent.Identities {
		return name
	}
	return ""
}

// Fingerprint returns the full-length native fingerprint
func (g *GPG) Fingerprint(ctx context.Context, id string) string {
	ent := g.findEntity(id)
	if ent == nil || ent.PrimaryKey == nil {
		return ""
	}
	return fmt.Sprintf("%x", ent.PrimaryKey.Fingerprint)
}

// Initialized returns nil
func (g *GPG) Initialized(context.Context) error {
	return nil
}

// Name returns openpgp
func (g *GPG) Name() string {
	return "openpgp"
}

// Ext returns gpg
func (g *GPG) Ext() string {
	return "gpg"
}

// IDFile returns .gpg-id
func (g *GPG) IDFile() string {
	return ".gpg-id"
}

// ReadNamesFromKey unmarshals and returns the names associated with the given public key
func (g *GPG) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	el, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(buf))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read key ring")
	}
	if len(el) != 1 {
		return nil, errors.Errorf("Public Key must contain exactly one Entity")
	}
	names := make([]string, 0, len(el[0].Identities))
	for _, v := range el[0].Identities {
		names = append(names, v.Name)
	}
	return names, nil
}
