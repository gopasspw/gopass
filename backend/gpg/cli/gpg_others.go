// +build !windows

package cli

import (
	"context"
	"os/exec"

	"github.com/blang/semver"
)

func (g *GPG) detectBinary(bin string) error {
	for _, b := range []string{bin, "gpg", "gpg2", "gpg1", "gpg"} {
		if p, err := exec.LookPath(b); err == nil {
			g.binary = p
			// if we found a GPG 2.x binary we're good, otherwise we try the
			// others as well
			if g.Version(context.Background()).GTE(semver.Version{Major: 2}) {
				break
			}
		}
	}
	return nil
}
