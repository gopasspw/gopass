package fsutil

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
)

var reCleanFilename = regexp.MustCompile(`[^\w\d@.-]`)

// CleanFilename strips all possibly suspicious characters from a filename
// WARNING: NOT suiteable for pathnames as slashes will be stripped as well!
func CleanFilename(in string) string {
	return strings.Trim(reCleanFilename.ReplaceAllString(in, "_"), "_ ")
}

// ExpandHomedir expands the tilde to the users home dir (if present).
func ExpandHomedir(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		dir := filepath.Clean(appdir.UserHome() + path[1:])
		debug.Log("Expanding %s to %s", path, dir)

		return dir
	}

	debug.Log("No tilde found in %s", path)

	return path
}

// CleanPath resolves common aliases in a path and cleans it as much as possible.
func CleanPath(path string) string {
	// Only replace ~ if GOPASS_HOMEDIR is set. In that case we do expect any reference
	// to the users homedir to be replaced by the value of GOPASS_HOMEDIR. This is mainly
	// for testing and experiments. In all other cases we do want to leave ~ as-is.
	if len(path) > 1 && path[:2] == "~/" {
		if hd := os.Getenv("GOPASS_HOMEDIR"); hd != "" {
			return filepath.Clean(hd + path[2:])
		}
	}

	if p, err := filepath.Abs(path); err == nil && !strings.HasPrefix(path, "~") {
		return p
	}

	return filepath.Clean(path)
}

// IsDir checks if a certain path exists and is a directory.
// https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// not found
			return false
		}

		debug.Log("failed to check dir %s: %s\n", path, err)

		return false
	}

	return fi.IsDir()
}

// IsFile checks if a certain path is actually a file.
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// not found
			return false
		}

		debug.Log("failed to check file %s: %s\n", path, err)

		return false
	}

	return fi.Mode().IsRegular()
}

// IsNonEmptyFile checks if a certain path is a regular file and
// non-zero in size.
func IsNonEmptyFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// not found
			return false
		}

		debug.Log("failed to check file %s: %s\n", path, err)

		return false
	}

	if !fi.Mode().IsRegular() {
		return false
	}

	return fi.Size() > 0
}

// IsEmptyDir checks if a certain path is an empty directory.
func IsEmptyDir(path string) (bool, error) {
	empty := true

	if err := filepath.Walk(path, func(fp string, fi os.FileInfo, ferr error) error {
		if ferr != nil {
			return ferr
		}
		if fi.IsDir() && (fi.Name() == "." || fi.Name() == "..") {
			return filepath.SkipDir
		}
		if !fi.IsDir() {
			empty = false
		}

		return nil
	}); err != nil {
		return false, fmt.Errorf("failed to walk %s: %w", path, err)
	}

	return empty, nil
}

// Shred overwrite the given file any number of times.
func Shred(path string, runs int) error {
	fh, err := os.OpenFile(path, os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", path, err)
	}

	// ignore the error. this is only taking effect if we error out.
	defer func() {
		_ = fh.Close()
	}()

	fi, err := fh.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file %q: %w", path, err)
	}

	flen := fi.Size()

	// overwrite using pseudo-random data n-1 times and
	// use zeros in the last iteration
	bufFn := func() []byte {
		buf := make([]byte, 1024)
		_, _ = rand.Read(buf)

		return buf
	}

	for i := 0; i < runs; i++ {
		if i >= runs-1 {
			bufFn = func() []byte {
				return make([]byte, 1024)
			}
		}

		if _, err := fh.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek to 0,0: %w", err)
		}

		var written int64

		for {
			// end of file
			if written >= flen {
				break
			}

			buf := bufFn()

			n, err := fh.Write(buf[0:min(flen-written, int64(len(buf)))])
			if err != nil {
				if !errors.Is(err, io.EOF) {
					return fmt.Errorf("failed to write to file: %w", err)
				}
				// end of file, should not happen
				break
			}

			written += int64(n)
		}
		// if we fail to sync the written blocks to disk it'd be pointless
		// do any further loops
		if err := fh.Sync(); err != nil {
			return fmt.Errorf("failed to sync to disk: %w", err)
		}
	}

	if err := fh.Close(); err != nil {
		return fmt.Errorf("failed to close file after writing: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}

	return nil
}

// FileContains searches the given file for the search string and returns true
// iff it's an exact (substring) match.
func FileContains(path, needle string) bool {
	fh, err := os.Open(path)
	if err != nil {
		debug.Log("failed to open %q for reading: %s", path, err)

		return false
	}

	defer func() {
		_ = fh.Close()
	}()

	s := bufio.NewScanner(fh)
	for s.Scan() {
		if strings.Contains(s.Text(), needle) {
			return true
		}
	}

	return false
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}

	return b
}

// CopyFile copies a file from src to dst. Permissions will be preserved. It is expected to
// fail if the destination does exist but is not writeable.
func CopyFile(from, to string) error {
	rdr, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("failed to open file %q for reading: %w", from, err)
	}
	defer func() {
		_ = rdr.Close()
	}()

	rdrStat, err := rdr.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat open file %q: %w", from, err)
	}

	wrt, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, rdrStat.Mode())
	if err != nil {
		return fmt.Errorf("failed to open file %q for writing: %w", to, err)
	}
	defer func() {
		_ = wrt.Close()
	}()

	n, err := io.Copy(wrt, rdr)
	if err != nil {
		return fmt.Errorf("failed to copy content of %q to %q: %w", from, to, err)
	}

	debug.Log("copied %d bytes from %q to %q", n, from, to)

	// sync permission, applies in case the destination did exist but had different perms
	if err := os.Chmod(to, rdrStat.Mode()); err != nil {
		return fmt.Errorf("failed to sync permissions to %q: %w", to, err)
	}

	return nil
}

// CopyFileForce copies a file from src to dst. Permissions will be preserved. The destination
// if removed before copying to avoid permission issues.
func CopyFileForce(from, to string) error {
	if IsFile(to) {
		if err := os.Remove(to); err != nil {
			return fmt.Errorf("failed to remove %q: %w", to, err)
		}
	}

	return CopyFile(from, to)
}
