package keyring

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gopasspw/gopass/internal/backend/crypto/xc/xcpb"

	"google.golang.org/protobuf/proto"
)

// Pubring is a public key ring
type Pubring struct {
	File string

	sync.Mutex
	data *xcpb.Pubring

	secring *Secring
}

// NewPubring initializes a new public key ring
func NewPubring(sec *Secring) *Pubring {
	return &Pubring{
		data: &xcpb.Pubring{
			PublicKeys: make([]*xcpb.PublicKey, 0, 10),
		},
		secring: sec,
	}
}

// LoadPubring loads an existing keyring from disk. If the file is not
// found an empty keyring is returned.
func LoadPubring(file string, sec *Secring) (*Pubring, error) {
	pr := NewPubring(sec)
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
func (p *Pubring) Save() error {
	buf, err := proto.Marshal(p.data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p.File, buf, 0600)
}

// Contains checks if a given key is in the keyring
func (p *Pubring) Contains(fp string) bool {
	p.Lock()
	defer p.Unlock()

	for _, pk := range p.data.PublicKeys {
		if pk.Fingerprint == fp {
			return true
		}
	}

	if p.secring == nil {
		return false
	}

	return p.secring.Contains(fp)
}

// KeyIDs returns a list of all key IDs
func (p *Pubring) KeyIDs() []string {
	p.Lock()
	defer p.Unlock()

	ids := make([]string, 0, len(p.data.PublicKeys))
	for _, pk := range p.data.PublicKeys {
		ids = append(ids, pk.Fingerprint)
	}
	if p.secring != nil {
		ids = append(ids, p.secring.KeyIDs()...)
	}
	sort.Strings(ids)
	return ids
}

// Export marshals a single key
func (p *Pubring) Export(id string) ([]byte, error) {
	p.Lock()
	defer p.Unlock()

	xpk := p.fetch(id)
	if xpk == nil {
		if p.secring != nil {
			return p.secring.Export(id, false)
		}
		return nil, fmt.Errorf("key not found")
	}

	return proto.Marshal(xpk)
}

// Get returns a single key
func (p *Pubring) Get(id string) *PublicKey {
	p.Lock()
	defer p.Unlock()

	xpk := p.fetch(id)
	if xpk == nil {
		if p.secring != nil {
			if pk := p.secring.Get(id); pk != nil {
				return &pk.PublicKey
			}
		}
		return nil
	}

	return pubPBToKR(xpk)
}

func (p *Pubring) fetch(id string) *xcpb.PublicKey {
	for _, pk := range p.data.PublicKeys {
		if pk.Fingerprint == id {
			return pk
		}
	}
	return nil
}

// Import unmarshals and inserts and previously exported key
func (p *Pubring) Import(buf []byte) error {
	pk := &xcpb.PublicKey{}
	if err := proto.Unmarshal(buf, pk); err != nil {
		return err
	}

	p.insert(pk)
	return nil
}

// Set inserts a key, possibly overwriting and existing entry
func (p *Pubring) Set(pk *PublicKey) error {
	p.Lock()
	defer p.Unlock()

	p.insert(pubKRToPB(pk))
	return nil
}

func (p *Pubring) insert(xpk *xcpb.PublicKey) {
	for i, e := range p.data.PublicKeys {
		if e.Fingerprint == xpk.Fingerprint {
			p.data.PublicKeys[i] = xpk
			return
		}
	}

	p.data.PublicKeys = append(p.data.PublicKeys, xpk)
}

// Remove deletes a single key
func (p *Pubring) Remove(id string) error {
	p.Lock()
	defer p.Unlock()

	match := -1
	for i, pk := range p.data.PublicKeys {
		if pk.Fingerprint == id {
			match = i
			break
		}
	}
	if match < 0 || match > len(p.data.PublicKeys) {
		return fmt.Errorf("not found")
	}
	p.data.PublicKeys = append(p.data.PublicKeys[:match], p.data.PublicKeys[match+1:]...)
	return nil
}

func pubPBToKR(xpk *xcpb.PublicKey) *PublicKey {
	if xpk == nil {
		return nil
	}

	pk := &PublicKey{
		PublicKey: [32]byte{},
	}
	pk.CreationTime = time.Unix(int64(xpk.CreationTime), 0)
	switch xpk.PubKeyAlgo {
	case xcpb.PublicKeyAlgorithm_NACL:
		pk.PubKeyAlgo = PubKeyNaCl
	}
	copy(pk.PublicKey[:], xpk.PublicKey)
	pk.Identity = xpk.Identity

	return pk
}

func pubKRToPB(pk *PublicKey) *xcpb.PublicKey {
	if pk == nil {
		return nil
	}

	xpk := &xcpb.PublicKey{
		CreationTime: uint64(pk.CreationTime.Unix()),
		Identity:     pk.Identity,
		Fingerprint:  pk.Fingerprint(),
	}
	switch pk.PubKeyAlgo {
	case PubKeyNaCl:
		xpk.PubKeyAlgo = xcpb.PublicKeyAlgorithm_NACL
	}
	copy(xpk.PublicKey[:], pk.PublicKey[:])

	return xpk
}
