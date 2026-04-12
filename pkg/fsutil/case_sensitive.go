//go:build !darwin && !windows

package fsutil

// NormalizeSecretName returns the secret name unchanged.
// On case-sensitive filesystems (Linux and others) no normalization is needed.
func NormalizeSecretName(name string) string {
	return name
}
