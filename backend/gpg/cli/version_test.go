package cli

import (
	"sort"
	"testing"

	"github.com/blang/semver"
)

func TestSort(t *testing.T) {
	in := []gpgBin{
		gpgBin{
			path: "/usr/local/bin/gpg",
			ver: semver.Version{
				Major: 1,
				Minor: 9,
				Patch: 1,
			},
		},
		gpgBin{
			path: "/usr/bin/gpg",
			ver: semver.Version{
				Major: 2,
				Minor: 4,
			},
		},
		gpgBin{
			path: "/usr/local/bin/gpg2",
			ver: semver.Version{
				Major: 2,
				Minor: 1,
				Patch: 11,
			},
		},
	}
	sort.Sort(byVersion(in))
	t.Logf("Out: %+v", in)
	if in[len(in)-1].ver.LT(semver.Version{Major: 2, Minor: 4}) {
		t.Errorf("wrong sort order")
	}
}
