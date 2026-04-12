package store

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateSecretName checks that a secret name is valid for use in the store.
// It rejects names that contain double slashes, path traversal components (..),
// or a leading slash, any of which could cause unexpected behavior or escapes
// from the store root.
func ValidateSecretName(name string) error {
	if strings.Contains(name, "//") {
		return fmt.Errorf("invalid secret name %q: must not contain consecutive slashes", name)
	}

	// Reject names that start with a slash (absolute—likely a mistake).
	if strings.HasPrefix(name, "/") {
		return fmt.Errorf("invalid secret name %q: must not start with a slash", name)
	}

	// Reject any path component that is "..", which could be used to escape
	// the store root when the name is joined with the store path.
	for _, part := range strings.Split(filepath.ToSlash(name), "/") {
		if part == ".." {
			return fmt.Errorf("invalid secret name %q: must not contain path traversal (..)", name)
		}
	}

	return nil
}
