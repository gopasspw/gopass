// +build !windows

package cli

import (
	"context"
	"errors"
	"os/exec"

	"github.com/blang/semver"
)

func (g *GPG) detectBinary(bin string) error {
	bins := []string{"gpg", "gpg2", "gpg1", "gpg"}
	if bin != "" {
		bins = append([]string{bin}, bins...)
	}
	for i, b := range bins {
		if p, err := exec.LookPath(b); err == nil {
			g.binary = p
			// if we found a GPG 2.x binary we're good, otherwise we try the
			// others as well
			if g.Version(context.Background()).GTE(semver.Version{Major: 2}) {
				return nil
			}
			// in case gpg is version 1 and there may be a gpg2 for version 2 we
			// don't want to return immedeately on the first gpg try, but afterwards
			// we take the "best" / first available binary
			if b != "gpg" && i > len(bins)-2 {
				return nil
			}
		}
	}
	return errors.New("no gpg binary found")
}
