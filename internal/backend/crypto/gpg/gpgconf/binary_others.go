//go:build !windows
// +build !windows

package gpgconf

import (
	"context"
	"os/exec"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

func detectBinary(_ context.Context, name string) (string, error) {
	// user supplied binaries take precedence
	if name != "" {
		return exec.LookPath(name)
	}
	// try to get the proper binary from gpgconf(1)
	p, err := Path("gpg")
	if err != nil || p == "" || !fsutil.IsFile(p) {
		debug.Log("gpgconf failed (%q), falling back to path lookup: %q", p, err)
		// otherwise fall back to the default and try
		// to look up "gpg"
		return exec.LookPath("gpg")
	}

	debug.Log("gpgconf returned %q for gpg", p)
	return p, nil
}
