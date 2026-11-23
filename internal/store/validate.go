package store

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateSecretName checks that a secret name is valid for use in the store.
// It rejects empty names, names that contain double slashes, path traversal
// components (..), current directory components (.), or a leading or trailing
// slash, any of which could cause unexpected behavior or escapes from the
// store root.
func ValidateSecretName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid secret name %q: must not be empty", name)
	}

	if strings.Contains(name, "//") {
		return fmt.Errorf("invalid secret name %q: must not contain consecutive slashes", name)
	}

	// Reject names that start with a slash (absolute—likely a mistake).
	if strings.HasPrefix(name, "/") {
		return fmt.Errorf("invalid secret name %q: must not start with a slash", name)
	}

	// Reject platform-native absolute paths (e.g. "C:\..." on Windows).
	if filepath.IsAbs(name) {
		return fmt.Errorf("invalid secret name %q: must not be an absolute path", name)
	}

	// Reject names that end with a slash, which would produce an empty basename
	// (e.g. "dir/.gpg").
	if strings.HasSuffix(name, "/") {
		return fmt.Errorf("invalid secret name %q: must not end with a slash", name)
	}

	// Reject any path component that is ".." or ".", which could be used to
	// escape the store root or produce hidden dotfile names like "..gpg".
	for _, part := range strings.Split(filepath.ToSlash(name), "/") {
		if part == ".." {
			return fmt.Errorf("invalid secret name %q: must not contain path traversal (..)", name)
		}
		if part == "." {
			return fmt.Errorf("invalid secret name %q: must not contain current directory component (.)", name)
		}
	}

	return nil
}
