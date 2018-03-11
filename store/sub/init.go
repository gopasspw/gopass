package sub

import (
	"context"
	"strings"

	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

// Initialized returns true if the store is properly initialized
func (s *Store) Initialized(ctx context.Context) bool {
	if s == nil || s.storage == nil {
		return false
	}
	return s.storage.Exists(ctx, s.idFile(ctx, ""))
}

// Init tries to initialize a new password store location matching the object
func (s *Store) Init(ctx context.Context, path string, ids ...string) error {
	if s.Initialized(ctx) {
		return errors.Errorf(`Found already initialized store at %s.
You can add secondary stores with gopass init --path <path to secondary store> --store <mount name>`, path)
	}

	// initialize recipient list
	recipients := make([]string, 0, len(ids))

	for _, id := range ids {
		if id == "" {
			continue
		}
		kl, err := s.crypto.FindPublicKeys(ctx, id)
		if err != nil {
			out.Red(ctx, "Failed to fetch public key for '%s': %s", id, err)
			continue
		}
		if len(kl) < 1 {
			out.Red(ctx, "No useable keys for '%s'", id)
			continue
		}
		recipients = append(recipients, kl[0])
	}

	if len(recipients) < 1 {
		return errors.Errorf("failed to initialize store: no valid recipients given in %+v", ids)
	}

	kl, err := s.crypto.FindPrivateKeys(ctx, recipients...)
	if err != nil {
		return errors.Errorf("Failed to get available private keys: %s", err)
	}

	if len(kl) < 1 {
		return errors.Errorf("None of the recipients has a secret key. You will not be able to decrypt the secrets you add")
	}

	if err := s.saveRecipients(ctx, recipients, "Initialized Store for "+strings.Join(recipients, ", "), true); err != nil {
		return errors.Wrapf(err, "failed to initialize store: %s", err)
	}

	return nil
}
