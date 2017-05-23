package fsutil

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// CleanPath resolves common aliases in a path and cleans it as much as possible
func CleanPath(path string) string {
	if fi, err := os.Lstat(path); err == nil {
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			path, _ = filepath.EvalSymlinks(path)
		}
	}
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
