package root

import (
	"context"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/backend/crypto/gpg"
)

type gpger interface {
	Binary() string
	ListPublicKeys(context.Context) (gpg.KeyList, error)
	FindPublicKeys(context.Context, ...string) (gpg.KeyList, error)
	ListPrivateKeys(context.Context) (gpg.KeyList, error)
	FindPrivateKeys(context.Context, ...string) (gpg.KeyList, error)
	GetRecipients(context.Context, string) ([]string, error)
	Encrypt(context.Context, string, []byte, []string) error
	Decrypt(context.Context, string) ([]byte, error)
	ExportPublicKey(context.Context, string, string) error
	ImportPublicKey(context.Context, string) error
	Version(context.Context) semver.Version
}

// GPGVersion returns GPG version information
func (r *Store) GPGVersion(ctx context.Context) semver.Version {
	return r.store.GPGVersion(ctx)
}
