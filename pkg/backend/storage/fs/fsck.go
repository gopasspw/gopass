package fs

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/out"
)

// Fsck checks the storage integrity
func (s *Store) Fsck(ctx context.Context) error {
	pcb := ctxutil.GetProgressCallback(ctx)

	entries, err := s.List(ctx, "")
	if err != nil {
		return err
	}
	dirs := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		pcb()
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
	out.Green(ctx, "  Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String())
	if err := syscall.Chmod(filename, np); err != nil {
		out.Error(ctx, "  Failed to set permissions for %s to rw-------: %s", filename, err)
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
		out.Green(ctx, "  Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String())
		if err := syscall.Chmod(dirname, np); err != nil {
			out.Error(ctx, "  Failed to set permissions for %s to rwx------: %s", dirname, err)
		}
	}

	// check for empty folders
	isEmpty, err := fsutil.IsEmptyDir(dirname)
	if err != nil {
		return err
	}
	if isEmpty {
		out.Error(ctx, "Folder %s is empty. Removing", dirname)
		return os.Remove(dirname)
	}
	return nil
}
