package keyring

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/xcpb"

	"google.golang.org/protobuf/proto"
)

// Secring is private key ring
type Secring struct {
	File string

	sync.Mutex
	data *xcpb.Secring
}

// NewSecring initializes and a new secring
func NewSecring() *Secring {
	return &Secring{
		data: &xcpb.Secring{
			PrivateKeys: make([]*xcpb.PrivateKey, 0, 10),
		},
	}
}

// LoadSecring loads an existing secring from disk. If the file is not found
// an empty keyring is returned
func LoadSecring(file string) (*Secring, error) {
	pr := NewSecring()
	pr.File = file

	buf, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return pr, nil
	}
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(buf, pr.data); err != nil {
		return nil, err
	}

	return pr, nil
}

// Save writes the keyring to the previously set location on disk
func (p *Secring) Save() error {
	buf, err := proto.Marshal(p.data)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p.File), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(p.File, buf, 0600)
}

// Contains returns true if the given key is found in the keyring
func (p *Secring) Contains(fp string) bool {
	p.Lock()
	defer p.Unlock()

	for _, pk := range p.data.PrivateKeys {
		if pk.PublicKey.Fingerprint == fp {
			return true
		}
	}
	return false
}

// KeyIDs returns a list of key IDs
func (p *Secring) KeyIDs() []string {
	p.Lock()
	defer p.Unlock()

	ids := make([]string, 0, len(p.data.PrivateKeys))
	for _, pk := range p.data.PrivateKeys {
		ids = append(ids, pk.PublicKey.Fingerprint)
	}
	sort.Strings(ids)
	return ids
}

// Export marshals a single private key
func (p *Secring) Export(id string, withPrivate bool) ([]byte, error) {
	p.Lock()
	defer p.Unlock()

	xpk := p.fetch(id)
	if xpk == nil {
		return nil, fmt.Errorf("key not found")
	}

	if withPrivate {
		return proto.Marshal(xpk)
	}
	return proto.Marshal(xpk.PublicKey)
}

// Get returns a single key
func (p *Secring) Get(id string) *PrivateKey {
	p.Lock()
	defer p.Unlock()

	xpk := p.fetch(id)
	if xpk == nil {
		return nil
	}

	return secPBToKR(xpk)
}

func (p *Secring) fetch(id string) *xcpb.PrivateKey {
	for _, pk := range p.data.PrivateKeys {
		if pk.PublicKey.Fingerprint == id {
			return pk
		}
	}
	return nil
}

// Import unmarshals and imports a previously exported key
func (p *Secring) Import(buf []byte) error {
	pk := &xcpb.PrivateKey{}
	if err := proto.Unmarshal(buf, pk); err != nil {
		return err
	}

	p.insert(pk)
	return nil
}

// Set inserts a single key
func (p *Secring) Set(pk *PrivateKey) error {
	if !pk.Encrypted {
		return fmt.Errorf("private key must be encrypted")
	}

	p.Lock()
	defer p.Unlock()

	p.insert(secKRToPB(pk))
	return nil
}

func (p *Secring) insert(xpk *xcpb.PrivateKey) {
	for i, e := range p.data.PrivateKeys {
		if e.PublicKey.Fingerprint == xpk.PublicKey.Fingerprint {
			p.data.PrivateKeys[i] = xpk
		}
	}

	p.data.PrivateKeys = append(p.data.PrivateKeys, xpk)
}

// Remove deletes the given key
func (p *Secring) Remove(id string) error {
	p.Lock()
	defer p.Unlock()

	match := -1
	for i, pk := range p.data.PrivateKeys {
		if pk.PublicKey.Fingerprint == id {
			match = i
			break
		}
	}
	if match < 0 || match > len(p.data.PrivateKeys) {
		return fmt.Errorf("not found")
	}
	p.data.PrivateKeys = append(p.data.PrivateKeys[:match], p.data.PrivateKeys[match+1:]...)
	return nil
}

func secPBToKR(xpk *xcpb.PrivateKey) *PrivateKey {
	pk := &PrivateKey{
		PublicKey: PublicKey{
			PublicKey: [32]byte{},
		},
		EncryptedData: make([]byte, len(xpk.Ciphertext)),
		Nonce:         [nonceLength]byte{},
		Salt:          make([]byte, len(xpk.Salt)),
	}

	// public part
	pk.PublicKey.CreationTime = time.Unix(int64(xpk.PublicKey.CreationTime), 0)
	switch xpk.PublicKey.PubKeyAlgo {
	case xcpb.PublicKeyAlgorithm_NACL:
		pk.PublicKey.PubKeyAlgo = PubKeyNaCl
	}
	copy(pk.PublicKey.PublicKey[:], xpk.PublicKey.PublicKey)
	pk.PublicKey.Identity = xpk.PublicKey.Identity

	// private part
	pk.Encrypted = true
	copy(pk.EncryptedData, xpk.Ciphertext)
	copy(pk.Nonce[:], xpk.Nonce)
	copy(pk.Salt, xpk.Salt)

	return pk
}

func secKRToPB(pk *PrivateKey) *xcpb.PrivateKey {
	xpk := &xcpb.PrivateKey{
		PublicKey: &xcpb.PublicKey{
			CreationTime: uint64(pk.CreationTime.Unix()),
			Identity:     pk.Identity,
			Fingerprint:  pk.Fingerprint(),
			PublicKey:    make([]byte, len(pk.PublicKey.PublicKey)),
		},
		Ciphertext: make([]byte, len(pk.EncryptedData)),
		Nonce:      make([]byte, len(pk.Nonce)),
		Salt:       make([]byte, len(pk.Salt)),
	}

	// public key
	switch pk.PubKeyAlgo {
	case PubKeyNaCl:
		xpk.PublicKey.PubKeyAlgo = xcpb.PublicKeyAlgorithm_NACL
	}
	copy(xpk.PublicKey.PublicKey, pk.PublicKey.PublicKey[:])

	// private key
	copy(xpk.Ciphertext, pk.EncryptedData)
	copy(xpk.Nonce, pk.Nonce[:])
	copy(xpk.Salt, pk.Salt)
	return xpk
}
