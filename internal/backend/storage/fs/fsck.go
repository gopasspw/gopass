package fs

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
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
		debug.Log("checking entry %q", entry)

		filename := filepath.Join(s.path, entry)
		dirs[filepath.Dir(filename)] = struct{}{}

		if err := s.fsckCheckFile(ctx, filename); err != nil {
			return err
		}
	}

	for dir := range dirs {
		debug.Log("checking dir %q", dir)
		if err := s.fsckCheckDir(ctx, dir); err != nil {
			return err
		}
	}

	if err := s.fsckCheckEmptyDirs(); err != nil {
		return err
	}

	debug.Log("checking root dir %q", s.path)
	return s.fsckCheckDir(ctx, s.path)
}

func (s *Store) fsckCheckFile(ctx context.Context, filename string) error {
	fi, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if fi.Mode().Perm()&0177 == 0 {
		return nil
	}

	out.Print(ctx, "Permissions too wide: %s (%s)", filename, fi.Mode().String())

	np := uint32(fi.Mode().Perm() & 0600)
	out.Print(ctx, "  Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String())
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
		out.Print(ctx, "Permissions too wide %s on dir %s", fi.Mode().Perm().String(), dirname)

		np := uint32(fi.Mode().Perm() & 0700)
		out.Print(ctx, "  Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String())
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

func (s *Store) fsckCheckEmptyDirs() error {
	v := []string{}
	if err := filepath.Walk(s.path, func(fp string, fi os.FileInfo, ferr error) error {
		if ferr != nil {
			return ferr
		}
		if !fi.IsDir() {
			return nil
		}
		if strings.HasPrefix(fi.Name(), ".") {
			return filepath.SkipDir
		}
		if fp == s.path {
			return nil
		}

		// add candidate
		debug.Log("adding candidate %q", fp)
		v = append(v, fp)
		return nil
	}); err != nil {
		return err
	}

	// start with longest path (deepest dir)
	sort.Slice(v, func(i, j int) bool {
		return len(v[i]) > len(v[j])
	})

	for _, d := range v {
		if err := fsckRemoveEmptyDir(d); err != nil {
			return err
		}
	}
	return nil
}

func fsckRemoveEmptyDir(fp string) error {
	ls, err := ioutil.ReadDir(fp)
	if err != nil {
		return err
	}
	if len(ls) > 0 {
		debug.Log("dir %q is not empty (%d)", fp, len(ls))
		return nil
	}

	debug.Log("removing %q ...", fp)
	return os.Remove(fp)
}
