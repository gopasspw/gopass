//go:build !darwin && !linux && !solaris && !windows && !freebsd

package clipboard

func killPrecedessors() error {
	return nil
}
