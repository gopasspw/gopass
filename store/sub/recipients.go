package sub

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
)

const (
	keyDir   = ".gpg-keys"
	fileMode = 0600
	dirMode  = 0700
)

// Recipients returns the list of recipients of this store
func (s *Store) Recipients(ctx context.Context) []string {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		out.Red(ctx, "failed to read recipient list: %s", err)
	}
	return rs
}

// AddRecipient adds a new recipient to the list
func (s *Store) AddRecipient(ctx context.Context, id string) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to read recipient list")
	}

	for _, k := range rs {
		if k == id {
			return errors.Errorf("Recipient already in store")
		}
	}

	rs = append(rs, id)

	if err := s.saveRecipients(ctx, rs, "Added Recipient "+id, true); err != nil {
		return errors.Wrapf(err, "failed to save recipients")
	}

	out.Cyan(ctx, "Reencrypting existing secrets. This may take some time ...")
	return s.reencrypt(WithReason(ctx, "Added Recipient "+id))
}

// SaveRecipients persists the current recipients on disk
func (s *Store) SaveRecipients(ctx context.Context) error {
	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed get recipients")
	}
	return s.saveRecipients(ctx, rs, "Save Recipients", true)
}

// RemoveRecipient will remove the given recipient from the storefunc (s *Store) RemoveRecipient()id string) error {
func (s *Store) RemoveRecipient(ctx context.Context, id string) error {
	// but if this key is not available on this machine we
	// just try to remove it literally
	keys, err := s.gpg.FindPublicKeys(ctx, id)
	if err != nil {
		out.Cyan(ctx, "Warning: Failed to get GPG Key Info for %s: %s", id, err)
	}

	rs, err := s.GetRecipients(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to read recipient list")
	}

	nk := make([]string, 0, len(rs)-1)
RECIPIENTS:
	for _, k := range rs {
		if k == id {
			continue RECIPIENTS
		}
		// if the key is available locally we can also match the id against
		// the fingerprint
		for _, key := range keys {
			if strings.HasSuffix(key.Fingerprint, k) {
				continue RECIPIENTS
			}
		}
		nk = append(nk, k)
	}

	if len(rs) == len(nk) {
		return errors.Errorf("recipient not in store")
	}

	if err := s.saveRecipients(ctx, nk, "Removed Recipient "+id, true); err != nil {
		return errors.Wrapf(err, "failed to save recipients")
	}

	return s.reencrypt(WithReason(ctx, "Removed Recipient "+id))
}

// GetRecipients will load all Recipients from the .gpg-id file for the given
// secret path
func (s *Store) GetRecipients(ctx context.Context, file string) ([]string, error) {
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

// ExportMissingPublicKeys will export any possibly missing public keys to the
// stores .gpg-keys directory
func (s *Store) ExportMissingPublicKeys(ctx context.Context, rs []string) (bool, error) {
	ok := true
	exported := false
	for _, r := range rs {
		path, err := s.exportPublicKey(ctx, r)
		if err != nil {
			ok = false
			out.Red(ctx, "failed to export public key for '%s': %s", r, err)
			continue
		}
		if path == "" {
			continue
		}
		// at least one key has been exported
		exported = true
		if err := s.gitAdd(ctx, path); err != nil {
			if errors.Cause(err) == store.ErrGitNotInit {
				continue
			}
			ok = false
			out.Red(ctx, "failed to add public key for '%s' to git: %s", r, err)
			continue
		}
		if err := s.gitCommit(ctx, fmt.Sprintf("Exported Public Keys %s", r)); err != nil && err != store.ErrGitNothingToCommit {
			ok = false
			out.Red(ctx, "Failed to git commit: %s", err)
			continue
		}
	}
	if !ok {
		return exported, errors.New("some keys failed")
	}
	return exported, nil
}

// Save all Recipients in memory to the .gpg-id file on disk.
func (s *Store) saveRecipients(ctx context.Context, rs []string, msg string, exportKeys bool) error {
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

	if err := s.gitAdd(ctx, idf); err != nil {
		if err != store.ErrGitNotInit {
			return errors.Wrapf(err, "failed to add file '%s' to git", idf)
		}
	}

	if err := s.gitCommit(ctx, msg); err != nil {
		if err != store.ErrGitNotInit && err != store.ErrGitNothingToCommit {
			return errors.Wrapf(err, "failed to commit changes to git")
		}
	}

	// save recipients' public keys
	if err := os.MkdirAll(filepath.Join(s.path, keyDir), dirMode); err != nil {
		return errors.Wrapf(err, "failed to create key dir '%s'", keyDir)
	}

	// save all recipients public keys to the repo
	if exportKeys {
		if _, err := s.ExportMissingPublicKeys(ctx, rs); err != nil {
			out.Red(ctx, "Failed to export missing public keys: %s", err)
		}
	}

	// push to remote repo
	if err := s.GitPush(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			out.Yellow(ctx, msg)
			return nil
		}
		return errors.Wrapf(err, "failed to push changes to git")
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
