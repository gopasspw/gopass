package gpgconf

import (
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

// Version return the version of the gpg binary
func Version(ctx context.Context, binary string) semver.Version {
	v := semver.Version{}

	cmd := exec.CommandContext(ctx, binary, "--version")
	out, err := cmd.Output()
	if err != nil {
		return v
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "gpg ") {
			continue
		}

		p := strings.Fields(line)
		if len(p) < 1 {
			continue
		}

		sv, err := semver.Parse(p[len(p)-1])
		if err != nil {
			continue
		}

		return sv
	}
	return v
}
