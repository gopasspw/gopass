package sub

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/store"
)

const (
	keyDir   = ".gpg-keys"
	fileMode = 0600
	dirMode  = 0700
)

// Recipients returns the list of recipients of this store
func (s *Store) Recipients() []string {
	return s.recipients
}

// AddRecipient adds a new recipient to the list
func (s *Store) AddRecipient(id string) error {
	for _, k := range s.recipients {
		if k == id {
			return fmt.Errorf("Recipient already in store")
		}
	}

	s.recipients = append(s.recipients, id)

	if err := s.saveRecipients("Added Recipient " + id); err != nil {
		return err
	}

	return s.reencrypt("Added Recipient " + id)
}

// SaveRecipients persists the current recipients on disk
func (s *Store) SaveRecipients() error {
	return s.saveRecipients("Save Recipients")
}

// RemoveRecipient will remove the given recipient from the storefunc (s *Store) RemoveRecipient()id string) error {
func (s *Store) RemoveRecipient(id string) error {
	// but if this key is not available on this machine we
	// just try to remove it literally
	keys, err := s.gpg.FindPublicKeys(id)
	if err != nil {
		fmt.Printf("Failed to get GPG Key Info for %s: %s\n", id, err)
	}
	nk := make([]string, 0, len(s.recipients)-1)
	for _, k := range s.recipients {
		if k == id {
			continue
		}
		if len(keys) > 0 {
			// if the key is available locally we can also match the id against
			// the fingerprint
			if strings.HasSuffix(keys[0].Fingerprint, k) {
				continue
			}
		}
		nk = append(nk, k)
	}
	s.recipients = nk

	if err := s.saveRecipients("Removed Recipient " + id); err != nil {
		return err
	}

	return s.reencrypt("Removed Recipients " + id)
}

// Load all Recipients from the .gpg-id file into a list of Recipients.
func (s *Store) loadRecipients() ([]string, error) {
	// open recipient list (store/.gpg-id)
	f, err := os.Open(s.idFile())
	if err != nil {
		return []string{}, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Failed to close %s: %s\n", s.idFile(), err)
		}
	}()

	return unmarshalRecipients(f), nil
}

// ImportMissingPublicKeys will try to import any missing public keys from the
// .gpg-keys folder in the password store
func (s *Store) ImportMissingPublicKeys() error {
	for _, r := range s.recipients {
		if s.debug {
			fmt.Printf("[DEBUG] Checking recipients %s ...\n", r)
		}
		// check if this recipient is missing
		// we could list all keys outside the loop and just do the lookup here
		// but this way we ensure to use the exact same lookup logic as
		// gpg does on encryption
		kl, err := s.gpg.FindPublicKeys(r)
		if err != nil {
			fmt.Printf("[%s] Failed to get public key for %s: %s\n", s.alias, r, err)
		}
		if len(kl) > 0 {
			fmt.Println(color.CyanString("[%s] Keyring contains %d public keys for %s", s.alias, len(kl), r))
			continue
		}

		// we need to ask the user before importing
		// any key material into his keyring!
		if s.importFunc != nil {
			if !s.importFunc(r) {
				continue
			}
		}

		// try to load this recipient
		if err := s.importPublicKey(r); err != nil {
			fmt.Println(color.RedString("[%s] Failed to import public key for %s: %s", s.alias, r, err))
			continue
		}
		fmt.Println(color.GreenString("[%s] Imported public key for %s into Keyring", s.alias, r))
	}
	return nil
}

// Save all Recipients in memory to the .gpg-id file on disk.
func (s *Store) saveRecipients(msg string) error {
	// filepath.Dir(s.idFile()) should equal s.path, but better safe than sorry
	if err := os.MkdirAll(filepath.Dir(s.idFile()), dirMode); err != nil {
		return err
	}

	// save recipients to store/.gpg-id
	if err := ioutil.WriteFile(s.idFile(), marshalRecipients(s.recipients), fileMode); err != nil {
		return err
	}

	err := s.gitAdd(s.idFile())
	if err == nil {
		if err := s.gitCommit(msg); err != nil {
			if err != store.ErrGitNotInit && err != store.ErrGitNothingToCommit {
				return err
			}
		}
	} else {
		if err != store.ErrGitNotInit {
			return err
		}
	}

	// save recipients' public keys
	if err := os.MkdirAll(filepath.Join(s.path, keyDir), dirMode); err != nil {
		return err
	}

	// save all recipients public keys to the repo
	for _, r := range s.recipients {
		path, err := s.exportPublicKey(r)
		if err != nil {
			return err
		}
		if err := s.gitAdd(path); err != nil {
			if err == store.ErrGitNotInit {
				continue
			}
			return err
		}
		if err := s.gitCommit(fmt.Sprintf("Exported Public Keys %s", r)); err != nil && err != store.ErrGitNothingToCommit {
			fmt.Println(color.RedString("Failed to git commit: %s", err))
			continue
		}
	}

	// push to remote repo
	if err := s.gitPush("", ""); err != nil {
		if err == store.ErrGitNotInit {
			return nil
		}
		if err == store.ErrGitNoRemote {
			msg := "Warning: git has not remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			fmt.Println(color.YellowString(msg))
			return nil
		}
		return err
	}

	return nil
}

// marshal all in memory Recipients line by line to []byte.
func marshalRecipients(r []string) []byte {
	if len(r) == 0 {
		return []byte("\n")
	}

	m := make(map[string]struct{}, len(r))
	for _, k := range r {
		m[k] = struct{}{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := bytes.Buffer{}
	for _, k := range keys {
		_, _ = out.WriteString(k)
		_, _ = out.WriteString("\n")
	}

	return out.Bytes()
}

// unmarshal Recipients line by line from a io.Reader.
func unmarshalRecipients(reader io.Reader) []string {
	m := make(map[string]struct{}, 5)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			m[line] = struct{}{}
		}
	}

	lst := make([]string, 0, len(m))
	for k := range m {
		lst = append(lst, k)
	}
	sort.Strings(lst)

	return lst
}

// export an ASCII armored public key
func (s *Store) exportPublicKey(r string) (string, error) {
	filename := filepath.Join(s.path, keyDir, r)
	if fsutil.IsFile(filename) {
		return filename, nil
	}

	if err := s.gpg.ExportPublicKey(r, filename); err != nil {
		return filename, err
	}

	return filename, nil
}

// import an public key into the default keyring
func (s *Store) importPublicKey(r string) error {
	filename := filepath.Join(s.path, keyDir, r)
	if !fsutil.IsFile(filename) {
		return fmt.Errorf("Public Key %s not found at %s", r, filename)
	}

	return s.gpg.ImportPublicKey(filename)
}
