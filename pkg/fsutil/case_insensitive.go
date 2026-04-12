//go:build darwin || windows

package fsutil

import "strings"

// NormalizeSecretName returns a canonical lowercase version of the secret
// name. On case-insensitive filesystems (macOS and Windows) secret names that
// differ only in case refer to the same underlying file, so we normalize to
// lowercase to avoid phantom duplicates and silent overwrites.
func NormalizeSecretName(name string) string {
	return strings.ToLower(name)
}
