package xc

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/debug"
)

const (
	name = "xc"
)

func init() {
	backend.RegisterCrypto(backend.XC, name, &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	debug.Log("Using Crypto Backend: %s (EXPERIMENTAL)", name)
	return New(config.Directory(), nil), nil
}

func (l loader) Handles(s backend.Storage) error {
	if s.Exists(context.TODO(), IDFile) {
		return nil
	}
	return fmt.Errorf("not supported")
}

func (l loader) Priority() int {
	return 11
}
func (l loader) String() string {
	return name
}
