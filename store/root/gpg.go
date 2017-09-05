package root

import (
	"context"

	"github.com/blang/semver"
)

// GPGVersion returns GPG version information
func (r *Store) GPGVersion(ctx context.Context) semver.Version {
	return r.store.GPGVersion(ctx)
}
