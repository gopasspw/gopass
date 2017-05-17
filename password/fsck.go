package password

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/gpg"
)

// Fsck checks this stores integrity
func (s *Store) Fsck(check, force bool) error {
	storeRec, err := gpg.ListPublicKeys(s.recipients...)
	if err != nil {
		fmt.Printf("Failed to list recipients: %s\n", err)
	}

	if err := filepath.Walk(s.path, s.mkStoreWalkerFsckFunc(check, force, storeRec, s.fsckFunc)); err != nil {
		return err
	}

	return nil
}

// mkStoreFsckWalkerFunc create a func to walk a (sub)store, i.e. list it's content
func (s *Store) mkStoreWalkerFsckFunc(check, force bool, storeRec gpg.KeyList, askFunc func(string) bool) func(string, os.FileInfo, error) error {
	shadowMap := make(map[string]struct{}, 100)
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && path != s.path {
			return filepath.SkipDir
		}
		if info.IsDir() && (info.Name() == "." || info.Name() == "..") {
			return filepath.SkipDir
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return s.fsckCheckDir(path, check, force, askFunc, shadowMap)
		}
		return s.fsckCheckFile(path, check, force, storeRec, askFunc, shadowMap)
	}
}

// fsckCheckDir checks a directory, mostly for it's permissions
func (s *Store) fsckCheckDir(fn string, check, force bool, askFunc func(string) bool, sh map[string]struct{}) error {
	fi, err := os.Stat(fn)
	if err != nil {
		fmt.Printf("Failed to check %s: %s\n", fn, err)
		return nil
	}
	// check for shadowing
	name := s.filenameToName(fn)
	if _, found := sh[name]; found {
		fmt.Printf("%s is shadowed by %s", name, fn)
	}
	sh[name] = struct{}{}
	// check if any group or other perms are set,
	// i.e. check for perms other than rwx------
	if fi.Mode().Perm()&077 != 0 {
		fmt.Println(color.CyanString("Wrong permissions for dir %s: %s", fn, fi.Mode().Perm().String()))
		if !check && (force || askFunc == nil || askFunc("Fix permissions?")) {
			np := uint32(fi.Mode().Perm() & 0700)
			fmt.Println(color.GreenString("Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String()))
			if err := syscall.Chmod(fn, np); err != nil {
				fmt.Println(color.RedString("Failed to set permissions for %s to rwx------: %s", fn, err))
			}
		}
	}
	return nil
}

func (s *Store) fsckCheckFile(fn string, check, force bool, storeRec gpg.KeyList, askFunc func(string) bool, sh map[string]struct{}) error {
	fi, err := os.Stat(fn)
	if err != nil {
		fmt.Printf("Failed to check %s: %s\n", fn, err)
		return nil
	}
	// check if any group or other perms are set,
	// i.e. check for perms other than rw-------
	if fi.Mode().Perm()&0177 != 0 {
		fmt.Println(color.CyanString("Wrong permissions for file %s: %s", fn, fi.Mode().String()))
		if !check && (force || askFunc == nil || askFunc("Fix permissions?")) {
			np := uint32(fi.Mode().Perm() & 0600)
			fmt.Println(color.GreenString("Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String()))
			if err := syscall.Chmod(fn, np); err != nil {
				fmt.Println(color.RedString("Failed to set permissions for %s to rw-------: %s", fn, err))
			}
		}
	}
	// we check all files (secrets and meta-data) for permissions,
	// but all other checks are only applied to secrets (which end in .gpg)
	if !strings.HasSuffix(fn, ".gpg") {
		return nil
	}
	// check for shadowing
	name := s.filenameToName(fn)
	if _, found := sh[name]; found {
		fmt.Println(color.CyanString("%s is shadowed by %s", name, fn))
	}
	sh[name] = struct{}{}
	// check that we can decrypt this file
	if err := s.fsckCheckSelfDecrypt(fn); err != nil {
		fmt.Println(color.RedString("No secret key available to decrypt %s. Can not fix", fn))
		return nil
	}
	// get the IDs this file was encrypted for
	fileRec, err := gpg.GetRecipients(fn)
	if err != nil {
		fmt.Println(color.RedString("Failed to get recipients of %s: %s\n", fn, err))
		return nil
	}
	// check that each recipient of the file is in the current
	// recipient list
	for _, rec := range fileRec {
		if _, err := storeRec.FindKey(rec); err == nil {
			// the recipient is (still) present in the recipients file of the store
			continue
		}
		// the recipient is not present in the recipients file of the store
		fmt.Println(color.CyanString("Extra recipient found for %s: %s\n", fn, rec))
		if !check && (force || askFunc == nil || askFunc("Fix recipients?")) {
			if err := s.fsckFixRecipients(fn); err != nil {
				fmt.Println(color.RedString("Failed to fix recipients for %s: %s\n", fn, err))
			}
		}
	}
	// check that each recipient of the store can actually decrypt this file
	for _, key := range storeRec {
		if err := fsckCheckRecipientsInSubkeys(key, fileRec); err == nil {
			continue
		}
		fmt.Println(color.CyanString("Missing recipient on %s: %s\n", s, key.OneLine()))
		if !check && (force || askFunc == nil || askFunc("Fix recipients?")) {
			if err := s.fsckFixRecipients(fn); err != nil {
				fmt.Println(color.RedString("Failed to fix recipients for %s: %s\n", fn, err))
			}
		}
	}
	return nil
}

func fsckCheckRecipientsInSubkeys(key gpg.Key, recipients []string) error {
	for _, rec := range recipients {
		for k := range key.SubKeys {
			if strings.HasSuffix(k, rec) {
				return nil
			}
		}
	}
	return fmt.Errorf("None of the Recipients matches a subkey")
}

func (s *Store) fsckCheckSelfDecrypt(fn string) error {
	_, err := s.Get(s.filenameToName(fn))
	return err
}

func (s *Store) fsckFixRecipients(fn string) error {
	name := s.filenameToName(fn)
	content, err := s.Get(s.filenameToName(fn))
	if err != nil {
		return err
	}
	return s.Set(name, content, "fsck fix recipients")
}
