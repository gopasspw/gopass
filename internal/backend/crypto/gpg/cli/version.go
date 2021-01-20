package cli

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/blang/semver/v4"
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
