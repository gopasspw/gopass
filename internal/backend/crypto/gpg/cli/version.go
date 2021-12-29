package cli

import (
	"context"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/internal/backend/crypto/gpg/gpgconf"
)

// Version will return GPG version information.
func (g *GPG) Version(ctx context.Context) semver.Version {
	return gpgconf.Version(ctx, g.Binary())
}
