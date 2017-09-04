package sub

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/justwatchcom/gopass/gpg"
	"github.com/pkg/errors"
)

// Fsck checks this stores integrity
func (s *Store) Fsck(ctx context.Context, prefix string, check, force bool) (map[string]uint64, error) {
	rs, err := s.getRecipients(ctx, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get recipients")
	}

	storeRec, err := s.gpg.FindPublicKeys(ctx, rs...)
	if err != nil {
		fmt.Printf("Failed to list recipients: %s\n", err)
	}

	counts := make(map[string]uint64, 5)
	countFn := func(t string) {
		counts[t]++
	}
	err = filepath.Walk(s.path, s.mkStoreWalkerFsckFunc(ctx, prefix, check, force, storeRec, s.fsckFunc, countFn))
	return counts, err
}

// mkStoreFsckWalkerFunc create a func to walk a (sub)store, i.e. list it's content
func (s *Store) mkStoreWalkerFsckFunc(ctx context.Context, prefix string, check, force bool, storeRec gpg.KeyList, askFunc func(context.Context, string) bool, countFn func(string)) func(string, os.FileInfo, error) error {
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
			return s.fsckCheckDir(ctx, prefix, path, check, force, askFunc, shadowMap, countFn)
		}
		return s.fsckCheckFile(ctx, prefix, path, check, force, storeRec, askFunc, shadowMap, countFn)
	}
}

// fsckCheckDir checks a directory, mostly for it's permissions
func (s *Store) fsckCheckDir(ctx context.Context, prefix, fn string, check, force bool, askFunc func(context.Context, string) bool, sh map[string]struct{}, countFn func(string)) error {
	fi, err := os.Stat(fn)
	if err != nil {
		fmt.Println(color.RedString("[%s] Failed to check %s: %s\n", prefix, fn, err))
		countFn("err")
		return nil
	}
	// check for shadowing
	name := s.filenameToName(fn)
	if _, found := sh[name]; found {
		fmt.Println(color.YellowString("[%s] Shadowed %s by %s", name, fn))
		countFn("warn")
	}
	sh[name] = struct{}{}
	// check if any group or other perms are set,
	// i.e. check for perms other than rwx------
	if fi.Mode().Perm()&077 != 0 {
		fmt.Println(color.YellowString("[%s] Permissions too wide: %s (%s)", prefix, fn, fi.Mode().Perm().String()))
		countFn("warn")
		if !check && (force || askFunc == nil || askFunc(ctx, "Fix permissions?")) {
			np := uint32(fi.Mode().Perm() & 0700)
			fmt.Println(color.GreenString("[%s] Fixing permissions from %s to %s", prefix, fi.Mode().Perm().String(), os.FileMode(np).Perm().String()))
			countFn("fixed")
			if err := syscall.Chmod(fn, np); err != nil {
				fmt.Println(color.RedString("[%s] Failed to set permissions for %s to rwx------: %s", prefix, fn, err))
				countFn("err")
			}
		}
	}
	// check for empty folders
	isEmpty, err := fsutil.IsEmptyDir(fn)
	if err != nil {
		return errors.Wrapf(err, "failed to check if '%s' is empty", fn)
	}
	if isEmpty {
		fmt.Println(color.YellowString("[%s] Empty folder: %s", prefix, fn))
		countFn("warn")
		if !check && (force || askFunc == nil || askFunc(ctx, "Remove empty folder?")) {
			fmt.Println(color.GreenString("[%s] Removing empty folder %s", prefix, fn))
			if err := os.RemoveAll(fn); err != nil {
				fmt.Println(color.RedString("[%s] Failed to remove folder %s: %s", fn, err))
				countFn("err")
			} else {
				countFn("fixed")
			}
		}
		return filepath.SkipDir
	}
	return nil
}

func (s *Store) fsckCheckFile(ctx context.Context, prefix, fn string, check, force bool, storeRec gpg.KeyList, askFunc func(context.Context, string) bool, sh map[string]struct{}, countFn func(string)) error {
	fi, err := os.Stat(fn)
	if err != nil {
		fmt.Println(color.RedString("[%s] Failed to check %s: %s\n", prefix, fn, err))
		countFn("err")
		return nil
	}
	// check if any group or other perms are set,
	// i.e. check for perms other than rw-------
	if fi.Mode().Perm()&0177 != 0 {
		fmt.Println(color.YellowString("[%s] Permissions too wide: %s (%s)", prefix, fn, fi.Mode().String()))
		countFn("warn")
		if !check && (force || askFunc == nil || askFunc(ctx, "Fix permissions?")) {
			np := uint32(fi.Mode().Perm() & 0600)
			fmt.Println(color.GreenString("[%s] Fixing permissions from %s to %s", prefix, fi.Mode().Perm().String(), os.FileMode(np).Perm().String()))
			if err := syscall.Chmod(fn, np); err != nil {
				fmt.Println(color.RedString("[%s] Failed to set permissions for %s to rw-------: %s", prefix, fn, err))
				countFn("err")
			} else {
				countFn("fixed")
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
		fmt.Println(color.YellowString("[%s] Shadowed %s by %s", prefix, name, fn))
		countFn("warn")
	}
	sh[name] = struct{}{}
	// check that we can decrypt this file
	if err := s.fsckCheckSelfDecrypt(ctx, fn); err != nil {
		fmt.Println(color.RedString("[%s] Secret Key missing. Can't fix: %s", prefix, fn))
		countFn("err")
		return nil
	}
	// get the IDs this file was encrypted for
	fileRec, err := s.gpg.GetRecipients(ctx, fn)
	if err != nil {
		fmt.Println(color.RedString("[%s] Failed to check recipients: %s (%s)", prefix, fn, err))
		countFn("err")
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
		fmt.Println(color.YellowString("[%s] Extra recipient found %s: %s", prefix, fn, rec))
		countFn("warn")
		if !check && (force || askFunc == nil || askFunc(ctx, "Fix recipients?")) {
			if err := s.fsckFixRecipients(ctx, fn); err != nil {
				fmt.Println(color.RedString("[%s] Failed to fix recipients for %s: %s", prefix, fn, err))
				countFn("err")
			}
		}
	}
	// check that each recipient of the store can actually decrypt this file
	for _, key := range storeRec {
		if err := fsckCheckRecipientsInSubkeys(key, fileRec); err == nil {
			continue
		}
		fmt.Println(color.YellowString("[%s] Recipient missing %s: %s", prefix, name, key.ID()))
		countFn("warn")
		if !check && (force || askFunc == nil || askFunc(ctx, "Fix recipients?")) {
			if err := s.fsckFixRecipients(ctx, fn); err != nil {
				fmt.Println(color.RedString("[%s] Failed to fix recipients for %s: %s\n", prefix, fn, err))
				countFn("err")
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
	return errors.Errorf("None of the Recipients matches a subkey")
}

func (s *Store) fsckCheckSelfDecrypt(ctx context.Context, fn string) error {
	_, err := s.Get(ctx, s.filenameToName(fn))
	return errors.Wrapf(err, "failed to decode secret")
}

func (s *Store) fsckFixRecipients(ctx context.Context, fn string) error {
	name := s.filenameToName(fn)
	content, err := s.Get(ctx, s.filenameToName(fn))
	if err != nil {
		return errors.Wrapf(err, "failed to decode secret")
	}
	return s.Set(ctx, name, content, "fsck fix recipients")
}
