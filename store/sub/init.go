package sub

import (
	"fmt"
	"strings"

	"github.com/justwatchcom/gopass/fsutil"
	"github.com/pkg/errors"
)

// Initialized returns true if the store is properly initialized
func (s *Store) Initialized() bool {
	return fsutil.IsFile(s.idFile())
}

// Init tries to initalize a new password store location matching the object
func (s *Store) Init(path string, ids ...string) error {
	if s.Initialized() {
		return errors.Errorf(`Found already initialized store at %s.
You can add secondary stores with gopass init --path <path to secondary store> --store <mount name>`, path)
	}

	// initialize recipient list
	s.recipients = make([]string, 0, len(ids))

	for _, id := range ids {
		if id == "" {
			continue
		}
		kl, err := s.gpg.FindPublicKeys(id)
		if err != nil || len(kl) < 1 {
			fmt.Println("Failed to fetch public key:", id)
			continue
		}
		s.recipients = append(s.recipients, kl[0].Fingerprint)
	}

	if len(s.recipients) < 1 {
		return errors.Errorf("failed to initialize store: no valid recipients given")
	}

	kl, err := s.gpg.FindPrivateKeys(s.recipients...)
	if err != nil {
		return errors.Errorf("Failed to get available private keys: %s", err)
	}

	if len(kl) < 1 {
		return errors.Errorf("None of the recipients has a secret key. You will not be able to decrypt the secrets you add")
	}

	if err := s.saveRecipients("Initialized Store for " + strings.Join(s.recipients, ", ")); err != nil {
		return errors.Errorf("failed to initialize store: %v", err)
	}

	return nil
}
