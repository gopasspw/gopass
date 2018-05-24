package gptest

import (
	"os"
	"path/filepath"
)

// AllPathsToSlash converts a list of paths to their correct
// platform specific slash representation
func AllPathsToSlash(paths []string) []string {
	r := make([]string, len(paths))
	for i, p := range paths {
		r[i] = filepath.ToSlash(p)
	}
	return r
}

func setupEnv(em map[string]string) error {
	for k, v := range em {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func teardownEnv(em map[string]string) {
	for k := range em {
		_ = os.Unsetenv(k)
	}
}
