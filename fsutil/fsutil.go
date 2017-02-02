package fsutil

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// CleanPath resolves common aliases in a path and cleans it as much as possible
func CleanPath(path string) string {
	// http://stackoverflow.com/questions/17609732/expand-tilde-to-home-directory
	if path[:2] == "~/" {
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
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// not found
			return false
		}
		fmt.Printf("failed to check dir %s: %s\n", path, err)
		return false
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		fmt.Printf("dir %s is a symlink. ignoring", path)
		return false
	}

	return fi.IsDir()
}

// IsFile checks if a certain path is actually a file
func IsFile(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// not found
			return false
		}
		fmt.Printf("failed to check dir %s: %s\n", path, err)
		return false
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		fmt.Printf("dir %s is a symlink. ignoring", path)
		return false
	}

	return fi.Mode().IsRegular()
}

// Tempdir returns a temporary directory suiteable for sensitive data. It tries
// /dev/shm but if this isn't working it will return an empty string. Using
// this with ioutil.Tempdir will ensure that we're getting the "best" tempdir.
func Tempdir() string {
	shmDir := "/dev/shm"
	if fi, err := os.Stat(shmDir); err == nil {
		if fi.IsDir() {
			if unix.Access(shmDir, unix.W_OK) == nil {
				return shmDir
			}
		}
	}
	return ""
}
