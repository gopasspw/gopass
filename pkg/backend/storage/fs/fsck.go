package fs

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	"github.com/justwatchcom/gopass/pkg/fsutil"
	"github.com/justwatchcom/gopass/pkg/out"
)

// Fsck checks the storage integrity
func (s *Store) Fsck(ctx context.Context) error {
	entries, err := s.List(ctx, "")
	if err != nil {
		return err
	}
	dirs := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		filename := filepath.Join(s.path, entry)
		dirs[filepath.Dir(filename)] = struct{}{}

		if err := s.fsckCheckFile(ctx, filename); err != nil {
			return err
		}
	}

	for dir := range dirs {
		if err := s.fsckCheckDir(ctx, dir); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) fsckCheckFile(ctx context.Context, filename string) error {
	fi, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if fi.Mode().Perm()&0177 == 0 {
		return nil
	}

	out.Yellow(ctx, "Permissions too wide: %s (%s)", filename, fi.Mode().String())

	np := uint32(fi.Mode().Perm() & 0600)
	out.Green(ctx, "Fixing permissions on %s from %s to %s", filename, fi.Mode().Perm().String(), os.FileMode(np).Perm().String())
	if err := syscall.Chmod(filename, np); err != nil {
		out.Red(ctx, "Failed to set permissions for %s to rw-------: %s", filename, err)
	}
	return nil
}

func (s *Store) fsckCheckDir(ctx context.Context, dirname string) error {
	fi, err := os.Stat(dirname)
	if err != nil {
		return err
	}

	// check if any group or other perms are set,
	// i.e. check for perms other than rwx------
	if fi.Mode().Perm()&077 != 0 {
		out.Yellow(ctx, "Permissions too wide %s on dir %s", fi.Mode().Perm().String(), dirname)

		np := uint32(fi.Mode().Perm() & 0700)
		out.Green(ctx, "Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String())
		if err := syscall.Chmod(dirname, np); err != nil {
			out.Red(ctx, "Failed to set permissions for %s to rwx------: %s", dirname, err)
		}
	}
	// check for empty folders
	isEmpty, err := fsutil.IsEmptyDir(dirname)
	if err != nil {
		return err
	}
	if isEmpty {
		out.Red(ctx, "WARNING: Folder %s is empty", dirname)
	}
	return nil
}
