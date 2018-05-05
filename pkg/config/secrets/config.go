package secrets

import (
	crypto_rand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/nacl/secretbox"

	"github.com/justwatchcom/gopass/pkg/fsutil"

	"github.com/pkg/errors"
)

const (
	saltLength  = 16
	nonceLength = 24
	keyLength   = 32
	filename    = "config.enc"
)

// Config is an encrypted config store
type Config struct {
	filename   string
	passphrase string
}

// New will load the given file from disk and try to unseal it
func New(dir, passphrase string) (*Config, error) {
	if dir == "" || dir == "." {
		return nil, fmt.Errorf("dir must not be empty")
	}
	fn := filepath.Join(dir, filename)

	c := &Config{
		filename:   fn,
		passphrase: passphrase,
	}

	if !fsutil.IsFile(fn) {
		err := save(c.filename, c.passphrase, map[string]string{})
		return c, err
	}

	_, err := c.Get("")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open existing secrects config %s: %s", fn, err)
	}
	return c, nil
}

// Get loads the requested key from disk
func (c *Config) Get(key string) (string, error) {
	data, err := load(c.filename, c.passphrase)
	return data[key], err
}

// Set writes the requested key to disk
func (c *Config) Set(key, value string) error {
	data, err := load(c.filename, c.passphrase)
	if err != nil {
		return errors.Wrapf(err, "failed to read secrects config %s: %s", c.filename, err)
	}

	old := data[key]
	if value == old {
		return nil
	}

	data[key] = value
	return save(c.filename, c.passphrase, data)
}

// Unset removes the key from the storage
func (c *Config) Unset(key string) error {
	data, err := load(c.filename, c.passphrase)
	if err != nil {
		return errors.Wrapf(err, "failed to read secrects config %s: %s", c.filename, err)
	}

	_, found := data[key]
	if !found {
		return nil
	}

	delete(data, key)
	return save(c.filename, c.passphrase, data)
}

func load(fn, passphrase string) (map[string]string, error) {
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return open(buf, passphrase)
}

// open will try to unseal the given buffer
func open(buf []byte, passphrase string) (map[string]string, error) {
	salt := make([]byte, saltLength)
	copy(salt, buf[:saltLength])
	var nonce [nonceLength]byte
	copy(nonce[:], buf[saltLength:nonceLength+saltLength])
	secretKey := deriveKey(passphrase, salt)
	decrypted, ok := secretbox.Open(nil, buf[nonceLength+saltLength:], &nonce, &secretKey)
	if !ok {
		return nil, fmt.Errorf("failed to decrypt")
	}
	data := map[string]string{}
	if err := json.Unmarshal(decrypted, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// save will try to marshal, seal and write to disk
func save(filename, passphrase string, data map[string]string) error {
	buf, err := seal(data, passphrase)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf, 0600)
}

// seal will try to marshal and seal the given data
func seal(data map[string]string, passphrase string) ([]byte, error) {
	jstr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var nonce [nonceLength]byte
	if _, err := io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return nil, err
	}
	salt := make([]byte, saltLength)
	if _, err := crypto_rand.Read(salt); err != nil {
		return nil, err
	}
	secretKey := deriveKey(passphrase, salt)
	prefix := append(salt, nonce[:]...)
	return secretbox.Seal(prefix, jstr, &nonce, &secretKey), nil
}

// parameters chosen as per https://godoc.org/golang.org/x/crypto/argon2#IDKey
func deriveKey(passphrase string, salt []byte) [keyLength]byte {
	secretKeyBytes := argon2.IDKey([]byte(passphrase), salt, 4, 64*1024, 4, 32)
	var secretKey [keyLength]byte
	copy(secretKey[:], secretKeyBytes)
	return secretKey
}
