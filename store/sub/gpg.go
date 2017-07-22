package sub

import "github.com/blang/semver"

// GPGVersion returns parsed GPG version information
func (s *Store) GPGVersion() semver.Version {
	return s.gpg.Version()
}
