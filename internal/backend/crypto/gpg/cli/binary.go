package cli

import (
	"context"
	"errors"
	"os/exec"
	"sort"
)

// Binary returns the GPG binary location
func (g *GPG) Binary() string {
	if g == nil {
		return ""
	}
	return g.binary
}

// Binary reutrns the GGP binary location
func Binary(ctx context.Context, bin string) (string, error) {
	bins, err := detectBinaryCandidates(bin)
	if err != nil {
		return "", err
	}
	bv := make(byVersion, 0, len(bins))
	for _, b := range bins {
		//debug.Log("gpg.detectBinary - Looking for '%s' ...", b)
		if p, err := exec.LookPath(b); err == nil {
			gb := gpgBin{
				path: p,
				ver:  version(ctx, p),
			}
			//debug.Log("gpg.detectBinary - Found '%s' at '%s' (%s)", b, p, gb.ver.String())
			bv = append(bv, gb)
		}
	}
	if len(bv) < 1 {
		return "", errors.New("no gpg binary found")
	}
	sort.Sort(bv)
	binary := bv[len(bv)-1].path
	//debug.Log("gpg.detectBinary - using '%s'", binary)
	return binary, nil
}
