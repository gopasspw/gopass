package sub

import "github.com/justwatchcom/gopass/gpg"

// GPGVersion returns parsed GPG version information
func (s *Store) GPGVersion() gpg.Version {
	return s.gpg.Version()
}
