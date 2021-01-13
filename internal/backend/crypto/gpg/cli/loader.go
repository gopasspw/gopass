package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

const (
	name = "gpgcli"
)

func init() {
	backend.RegisterCrypto(backend.GPGCLI, name, &loader{})
}

type loader struct{}

// New implements backend.CryptoLoader.
func (l loader) New(ctx context.Context) (backend.Crypto, error) {
	debug.Log("Using Crypto Backend: %s", name)
	return New(ctx, Config{
		Umask:  fsutil.Umask(),
		Args:   GPGOpts(),
		Binary: os.Getenv("GOPASS_GPG_BINARY"),
	})
}

func (l loader) Handles(s backend.Storage) error {
	if s.Exists(context.TODO(), IDFile) {
		return nil
	}
	return fmt.Errorf("not supported")
}

func (l loader) Priority() int {
	return 1
}

func (l loader) String() string {
	return name
}
