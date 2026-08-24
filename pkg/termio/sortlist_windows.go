//go:build windows

package termio

// makeRaw is a no-op on Windows. The interactive sort list falls back to
// line-based input there.
func makeRaw() (*termState, error) {
	return nil, nil
}

// restore is a no-op on Windows.
func restore(*termState) error {
	return nil
}

// termState is a placeholder for the terminal state on Windows.
type termState struct{}
