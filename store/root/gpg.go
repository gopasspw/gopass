package root

import "github.com/blang/semver"

// GPGVersion returns GPG version information
func (r *Store) GPGVersion() semver.Version {
	return r.store.GPGVersion()
}
