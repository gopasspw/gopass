package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os/exec"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/utils/out"
)

type gpgBin struct {
	path string
	ver  semver.Version
}

type byVersion []gpgBin

func (v byVersion) Len() int {
	return len(v)
}

func (v byVersion) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v byVersion) Less(i, j int) bool {
	return v[i].ver.LT(v[j].ver)
}

// Version will returns GPG version information
func (g *GPG) Version(ctx context.Context) semver.Version {
	return version(ctx, g.Binary())
}

func version(ctx context.Context, binary string) semver.Version {
	v := semver.Version{}

	cmd := exec.CommandContext(ctx, binary, "--version")
	out, err := cmd.Output()
	if err != nil {
		return v
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "gpg ") {
			p := strings.Fields(line)
			sv, err := semver.Parse(p[len(p)-1])
			if err != nil {
				continue
			}
			return sv
		}
	}
	return v
}

func (g *GPG) detectBinary(ctx context.Context, bin string) error {
	bins, err := g.detectBinaryCandidates(bin)
	if err != nil {
		return err
	}
	bv := make(byVersion, 0, len(bins))
	for _, b := range bins {
		out.Debug(ctx, "gpg.detectBinary - Looking for '%s' ...", b)
		if p, err := exec.LookPath(b); err == nil {
			gb := gpgBin{
				path: p,
				ver:  version(ctx, p),
			}
			out.Debug(ctx, "gpg.detectBinary - Found '%s' at '%s' (%s)", b, p, gb.ver.String())
			bv = append(bv, gb)
		}
	}
	if len(bv) < 1 {
		return errors.New("no gpg binary found")
	}
	sort.Sort(bv)
	g.binary = bv[len(bv)-1].path
	out.Debug(ctx, "gpg.detectBinary - using '%s'", g.binary)
	return nil
}
