package ondisk

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

// Fsck checks store integrity and performs a compaction
func (o *OnDisk) Fsck(ctx context.Context) error {
	pcb := ctxutil.GetProgressCallback(ctx)

	if err := o.Compact(ctx); err != nil {
		return err
	}

	// build a list of existing files
	files := make(map[string]struct{}, len(o.idx.Entries)+1)
	files[idxFile] = struct{}{}
	files[idxBakFile] = struct{}{}
	for _, v := range o.idx.Entries {
		if v.IsDeleted() {
			continue
		}
		for _, r := range v.Revisions {
			files[r.Filename] = struct{}{}
		}
	}

	return filepath.Walk(o.dir, func(path string, fi os.FileInfo, err error) error {
		defer pcb()
		if err != nil {
			return err
		}
		if fi.IsDir() && len(fi.Name()) != 2 && path != o.dir {
			out.Print(ctx, "Skipping unknown dir: %s", path)
			return filepath.SkipDir
		}
		out.Debug(ctx, "Checking: %s", path)
		if fi.IsDir() {
			return o.fsckCheckDir(ctx, path, fi)
		}
		relPath := strings.TrimPrefix(path, o.dir+string(filepath.Separator))
		if err := o.fsckCheckFile(ctx, relPath, fi); err != nil {
			return err
		}
		_, found := files[relPath]
		if found {
			return nil
		}
		out.Yellow(ctx, "Found orphaned file in store. Removing: %s", relPath)
		return os.Remove(path)
	})
}

func (o *OnDisk) fsckCheckFile(ctx context.Context, relPath string, fi os.FileInfo) error {
	path := filepath.Join(o.dir, relPath)
	// check filename / hashsum
	fileHash, err := hashFromFile(path)
	if err != nil {
		return err
	}
	if len(fileHash) < 3 {
		return fmt.Errorf("invalid hash")
	}
	wantPath := filepath.Join(fileHash[0:2], fileHash[2:])
	if relPath != wantPath && !strings.Contains(relPath, idxFile) {
		wantFullPath := filepath.Join(o.dir, wantPath)
		out.Error(ctx, "  Invalid checksum / path: Want %s for %s", wantPath, relPath)
		if err := os.Rename(path, wantFullPath); err != nil {
			return err
		}
		out.Yellow(ctx, "  Renamed %s to %s", relPath, wantPath)
		path = wantFullPath
	}

	// check file modes
	if fi.Mode().Perm()&0177 == 0 {
		return nil
	}

	out.Yellow(ctx, "Permissions too wide: %s (%s)", path, fi.Mode().String())

	np := uint32(fi.Mode().Perm() & 0600)
	out.Green(ctx, "  Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String())
	if err := syscall.Chmod(path, np); err != nil {
		out.Error(ctx, "  Failed to set permissions for %s to rw-------: %s", path, err)
	}
	return nil
}

func (o *OnDisk) fsckCheckDir(ctx context.Context, path string, fi os.FileInfo) error {
	// check if any group or other perms are set,
	// i.e. check for perms other than rwx------
	if fi.Mode().Perm()&077 != 0 {
		out.Yellow(ctx, "Permissions too wide %s on dir %s", fi.Mode().Perm().String(), path)

		np := uint32(fi.Mode().Perm() & 0700)
		out.Green(ctx, "  Fixing permissions from %s to %s", fi.Mode().Perm().String(), os.FileMode(np).Perm().String())
		if err := syscall.Chmod(path, np); err != nil {
			out.Error(ctx, "  Failed to set permissions for %s to rwx------: %s", path, err)
		}
	}

	// check for empty folders
	isEmpty, err := fsutil.IsEmptyDir(path)
	if err != nil {
		return err
	}
	if isEmpty {
		out.Error(ctx, "Folder %s is empty. Removing", path)
		return os.Remove(path)
	}
	return nil
}

func hashFromFile(path string) (string, error) {
	fh, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fh.Close()

	h := sha256.New()
	if _, err := io.Copy(h, fh); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
