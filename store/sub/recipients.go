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
	"github.com/justwatchcom/gopass/store"
	"github.com/pkg/errors"
)

const (
	keyDir   = ".gpg-keys"
	fileMode = 0600
	dirMode  = 0700
)

// Recipients returns the list of recipients of this store
func (s *Store) Recipients() []string {
	rs, err := s.getRecipients("")
	if err != nil {
		fmt.Println(color.RedString("failed to read recipient list: %s", err))
	}
	return rs
}

// AddRecipient adds a new recipient to the list
func (s *Store) AddRecipient(id string) error {
	rs, err := s.getRecipients("")
	if err != nil {
		return errors.Wrapf(err, "failed to read recipient list")
	}

	for _, k := range rs {
		if k == id {
			return errors.Errorf("Recipient already in store")
		}
	}

	rs = append(rs, id)

	if err := s.saveRecipients(rs, "Added Recipient "+id, true); err != nil {
		return errors.Wrapf(err, "failed to save recipients")
	}

	return s.reencrypt("Added Recipient " + id)
}

// SaveRecipients persists the current recipients on disk
func (s *Store) SaveRecipients() error {
	rs, err := s.getRecipients("")
	if err != nil {
		return errors.Wrapf(err, "failed get recipients")
	}
	return s.saveRecipients(rs, "Save Recipients", true)
}

// RemoveRecipient will remove the given recipient from the storefunc (s *Store) RemoveRecipient()id string) error {
func (s *Store) RemoveRecipient(id string) error {
	// but if this key is not available on this machine we
	// just try to remove it literally
	keys, err := s.gpg.FindPublicKeys(id)
	if err != nil {
		fmt.Printf("Failed to get GPG Key Info for %s: %s\n", id, err)
	}

	rs, err := s.getRecipients("")
	if err != nil {
		return errors.Wrapf(err, "failed to read recipient list")
	}

	nk := make([]string, 0, len(rs)-1)
	for _, k := range rs {
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

	if err := s.saveRecipients(nk, "Removed Recipient "+id, true); err != nil {
		return errors.Wrapf(err, "failed to save recipients")
	}

	return s.reencrypt("Removed Recipients " + id)
}

// Load all Recipients from the .gpg-id file into a list of Recipients.
func (s *Store) getRecipients(file string) ([]string, error) {
	idf := s.idFile(file)
	// open recipient list (store/.gpg-id)
	f, err := os.Open(idf)
	if err != nil {
		return []string{}, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Failed to close %s: %s\n", idf, err)
		}
	}()

	return unmarshalRecipients(f), nil
}

// Save all Recipients in memory to the .gpg-id file on disk.
func (s *Store) saveRecipients(rs []string, msg string, exportKeys bool) error {
	if len(rs) < 1 {
		return errors.New("can not remove all recipients")
	}

	idf := s.idFile("")
	// filepath.Dir(s.idFile()) should equal s.path, but better safe than sorry
	if err := os.MkdirAll(filepath.Dir(idf), dirMode); err != nil {
		return errors.Wrapf(err, "failed to create directory for recipients")
	}

	// save recipients to store/.gpg-id
	if err := ioutil.WriteFile(idf, marshalRecipients(rs), fileMode); err != nil {
		return errors.Wrapf(err, "failed to write recipients file")
	}

	err := s.gitAdd(idf)
	if err == nil {
		if err := s.gitCommit(msg); err != nil {
			if err != store.ErrGitNotInit && err != store.ErrGitNothingToCommit {
				return errors.Wrapf(err, "failed to commit changes to git")
			}
		}
	} else {
		if err != store.ErrGitNotInit {
			return errors.Wrapf(err, "failed to add file '%s' to git", idf)
		}
	}

	// save recipients' public keys
	if err := os.MkdirAll(filepath.Join(s.path, keyDir), dirMode); err != nil {
		return errors.Wrapf(err, "failed to create key dir '%s'", keyDir)
	}

	// save all recipients public keys to the repo
	if exportKeys {
		if err := s.exportPublicKeys(rs); err != nil {
			return errors.Wrapf(err, "failed to export public keys: %s", err)
		}
	}

	// push to remote repo
	if err := s.gitPush("", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has not remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			fmt.Println(color.YellowString(msg))
			return nil
		}
		return errors.Wrapf(err, "failed to push changes to git")
	}

	return nil
}

func (s *Store) exportPublicKeys(rs []string) error {
	for _, r := range rs {
		path, err := s.exportPublicKey(r)
		if err != nil {
			fmt.Println(color.RedString("failed to export public keys for '%s': %s", r, err))
			continue
		}

		if err := s.gitAdd(path); err != nil {
			if errors.Cause(err) == store.ErrGitNotInit {
				continue
			}
			fmt.Println(color.RedString("failed to add public key for '%s' to git: %s", r, err))
			continue
		}
	}

	if err := s.gitCommit(fmt.Sprintf("Exported Public Keys %v", rs)); err != nil && errors.Cause(err) != store.ErrGitNothingToCommit && errors.Cause(err) != store.ErrGitNotInit {
		fmt.Println(color.RedString("Failed to git commit: %s", err))
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
