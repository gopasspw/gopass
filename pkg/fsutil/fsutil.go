package fsutil

import (
	"io"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/pkg/errors"
)

var reCleanFilename = regexp.MustCompile(`[^\w\d@.-]`)

// CleanFilename strips all possibly suspicious characters from a filename
// WARNING: NOT suiteable for pathnames as slashes will be stripped as well!
func CleanFilename(in string) string {
	return strings.Trim(reCleanFilename.ReplaceAllString(in, "_"), "_ ")
}

// CleanPath resolves common aliases in a path and cleans it as much as possible
func CleanPath(path string) string {
	// http://stackoverflow.com/questions/17609732/expand-tilde-to-home-directory
	// TODO: We should consider if we really want to rewrite ~
	if len(path) > 1 && path[:2] == "~/" {
		usr, _ := user.Current()
		dir := usr.HomeDir
		path = strings.Replace(path, "~/", dir+"/", 1)
	}
	if p, err := filepath.Abs(path); err == nil {
		return p
	}
	return filepath.Clean(path)
}

// IsDir checks if a certain path exists and is a directory
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

// IsFile checks if a certain path is actually a file
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

// IsEmptyDir checks if a certain path is an empty directory
func IsEmptyDir(path string) (bool, error) {
	empty := true
	if err := filepath.Walk(path, func(fp string, fi os.FileInfo, ferr error) error {
		if ferr != nil {
			return ferr
		}
		if fi.IsDir() && (fi.Name() == "." || fi.Name() == "..") {
			return filepath.SkipDir
		}
		if fi.Mode().IsRegular() {
			empty = false
		}
		return nil
	}); err != nil {
		return false, err
	}
	return empty, nil
}

// Shred overwrite the given file any number of times
func Shred(path string, runs int) error {
	rand.Seed(time.Now().UnixNano())
	fh, err := os.OpenFile(path, os.O_WRONLY, 0600)
	if err != nil {
		return errors.Wrapf(err, "failed to open file '%s'", path)
	}
	buf := make([]byte, 1024)
	for i := 0; i < runs; i++ {
		// overwrite using pseudo-random data n-1 times and
		// use zeros in the last iteration
		if i < runs-1 {
			_, _ = rand.Read(buf)
		} else {
			buf = make([]byte, 1024)
		}
		if _, err := fh.Seek(0, 0); err != nil {
			return errors.Wrapf(err, "failed to seek to 0,0")
		}
		if _, err := fh.Write(buf); err != nil {
			if err != io.EOF {
				return errors.Wrapf(err, "failed to write to file")
			}
		}
		// if we fail to sync the written blocks to disk it'd be pointless
		// do any further loops
		if err := fh.Sync(); err != nil {
			return errors.Wrapf(err, "failed to sync to disk")
		}
	}
	if err := fh.Close(); err != nil {
		return errors.Wrapf(err, "failed to close file after writing")
	}

	return os.Remove(path)
}
