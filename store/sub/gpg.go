package sub

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/pkg/errors"
)

// GPGVersion returns parsed GPG version information
func (s *Store) GPGVersion(ctx context.Context) semver.Version {
	return s.gpg.Version(ctx)
}

// ImportMissingPublicKeys will try to import any missing public keys from the
// .gpg-keys folder in the password store
func (s *Store) ImportMissingPublicKeys(ctx context.Context) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to get recipients")
	}
	for _, r := range rs {
		if s.debug {
			fmt.Printf("[DEBUG] Checking recipients %s ...\n", r)
		}
		// check if this recipient is missing
		// we could list all keys outside the loop and just do the lookup here
		// but this way we ensure to use the exact same lookup logic as
		// gpg does on encryption
		kl, err := s.gpg.FindPublicKeys(ctx, r)
		if err != nil {
			fmt.Printf("[%s] Failed to get public key for %s: %s\n", s.alias, r, err)
		}
		if len(kl) > 0 {
			if s.debug {
				fmt.Println(color.CyanString("[%s] Keyring contains %d public keys for %s", s.alias, len(kl), r))
			}
			continue
		}

		// we need to ask the user before importing
		// any key material into his keyring!
		if s.importFunc != nil {
			if !s.importFunc(ctx, r) {
				continue
			}
		}

		// try to load this recipient
		if err := s.importPublicKey(ctx, r); err != nil {
			fmt.Println(color.RedString("[%s] Failed to import public key for %s: %s", s.alias, r, err))
			continue
		}
		fmt.Println(color.GreenString("[%s] Imported public key for %s into Keyring", s.alias, r))
	}
	return nil
}

// export an ASCII armored public key
func (s *Store) exportPublicKey(ctx context.Context, r string) (string, error) {
	filename := filepath.Join(s.path, keyDir, r)

	// do not overwrite existing keys
	if fsutil.IsFile(filename) {
		return "", nil
	}

	tmpFilename := filename + ".new"
	if err := s.gpg.ExportPublicKey(ctx, r, tmpFilename); err != nil {
		return "", err
	}

	defer func() {
		_ = os.Remove(tmpFilename)
	}()

	fi, err := os.Stat(tmpFilename)
	if err != nil {
		return "", err
	}

	if fi.Size() < 1024 {
		return "", errors.New("exported key too small")
	}

	if err := os.Rename(tmpFilename, filename); err != nil {
		return "", err
	}

	return filename, nil
}

// import an public key into the default keyring
func (s *Store) importPublicKey(ctx context.Context, r string) error {
	filename := filepath.Join(s.path, keyDir, r)
	if !fsutil.IsFile(filename) {
		return errors.Errorf("Public Key %s not found at %s", r, filename)
	}

	return s.gpg.ImportPublicKey(ctx, filename)
}
