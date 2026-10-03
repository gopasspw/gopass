package age

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/backend/crypto/age/identityfile"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// keyringUsesPassphrase inspects only the age header. Authentication and full
// format validation remain the responsibility of age.Decrypt.
func keyringUsesPassphrase(ciphertext []byte) (bool, error) {
	scanner := bufio.NewScanner(bytes.NewReader(ciphertext))
	if !scanner.Scan() || scanner.Text() != "age-encryption.org/v1" {
		return false, fmt.Errorf("invalid age keyring header")
	}
	passphrase := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "--- ") {
			return passphrase, nil
		}
		if strings.HasPrefix(line, "-> scrypt ") {
			passphrase = true
		}
	}

	return false, fmt.Errorf("incomplete age keyring header")
}

// keyringRecipients requires every configured entry to be an explicit public
// recipient. In particular, never silently drop an invalid recovery recipient.
func (a *Age) keyringRecipients(ctx context.Context) ([]age.Recipient, error) {
	value := strings.TrimSpace(config.String(ctx, "age.keyring-recipients"))
	if value == "" {
		return nil, nil
	}
	var recipients []age.Recipient
	for index, entry := range strings.Split(value, ",") {
		entry = strings.TrimSpace(entry)
		if !strings.HasPrefix(entry, "age1") && !strings.HasPrefix(entry, "ssh-") {
			return nil, fmt.Errorf("age.keyring-recipients: entry %d must be an explicit public recipient", index+1)
		}
		parsed, err := a.parseRecipients(ctx, []string{entry})
		if err != nil || len(parsed) != 1 {
			return nil, fmt.Errorf("age.keyring-recipients: invalid public recipient at entry %d", index+1)
		}
		recipients = append(recipients, parsed[0])
	}

	return recipients, nil
}

// keyringUnlockIdentities loads bootstrap identities without consulting the
// keyring, agent or SSH discovery, avoiding a recursive unlock dependency.
func (a *Age) keyringUnlockIdentities(ctx context.Context) ([]age.Identity, error) {
	path := strings.TrimSpace(config.String(ctx, "age.keyring-identities"))
	if path == "" {
		return nil, fmt.Errorf("recipient-encrypted age keyring requires age.keyring-identities")
	}
	path = fsutil.CleanPath(path)
	if filepath.Clean(path) == filepath.Clean(a.identity) {
		return nil, fmt.Errorf("age.keyring-identities must be independent of the encrypted keyring")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring unlock identities: %w", err)
	}
	defer func() { _ = file.Close() }()

	ids, err := identityfile.Parse(file, parseIdentity)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keyring unlock identities: %w", err)
	}

	return ids, nil
}

// writeEncryptedKeyring replaces the file only after encryption completes. The
// temporary file contains ciphertext only and must share the destination's
// filesystem for atomic rename; the ramdisk tempfile helper cannot ensure that.
func writeEncryptedKeyring(filename string, ciphertext []byte) error {
	file, err := os.CreateTemp(filepath.Dir(filename), ".age-keyring-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()
	if _, err := file.Write(ciphertext); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(file.Name(), filename)
}

// reencryptIdentities migrates protection without parsing or rewriting identity
// lines (plugins in the keyring need not be present just to change protection).
func (a *Age) reencryptIdentities(ctx context.Context) error {
	contents, err := a.loadIdentityFile(ctx)
	if err != nil {
		return fmt.Errorf("failed to read identity keyring: %w", err)
	}
	if err := a.saveIdentities(ctx, []string{contents}, false); err != nil {
		return err
	}
	if err := a.recpCache.Remove(idRecpCacheKey); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to invalidate identity recipient cache: %w", err)
	}

	return nil
}

// recipientKeyring detects protection without prompting for credentials.
func (a *Age) recipientKeyring() (bool, error) {
	ciphertext, err := os.ReadFile(a.identity)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	passphrase, err := keyringUsesPassphrase(ciphertext)
	if err != nil {
		return false, err
	}

	return !passphrase, nil
}

// verifyKeyringProtection refuses to replace the keyring with ciphertext that
// the configured bootstrap identities cannot unlock.
func (a *Age) verifyKeyringProtection(ctx context.Context, ciphertext, plaintext []byte) error {
	ids, err := a.keyringUnlockIdentities(ctx)
	if err != nil {
		return err
	}
	restored, err := a.decrypt(ciphertext, ids...)
	if err != nil {
		return fmt.Errorf("cannot unlock the new identity keyring; original keyring unchanged: %w", err)
	}
	defer clear(restored)
	if !bytes.Equal(restored, plaintext) {
		return fmt.Errorf("new identity keyring verification failed")
	}

	return nil
}

// missingIdentityFile distinguishes an absent keyring from a missing bootstrap
// identity file, whose error must not be mistaken for an empty keyring.
func missingIdentityFile(filename string, err error) bool {
	if !errors.Is(err, os.ErrNotExist) {
		return false
	}
	_, statErr := os.Stat(filename)

	return errors.Is(statErr, os.ErrNotExist)
}
