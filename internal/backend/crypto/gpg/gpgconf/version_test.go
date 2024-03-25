package gpgconf

import (
	"sort"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/require"
)

func TestSort(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   []gpgBin
		out  []semver.Version
	}{
		{
			name: "simple",
			in: []gpgBin{
				{
					path: "/usr/local/bin/gpg",
					ver:  semver.MustParse("1.9.1"),
				},
				{
					path: "/usr/bin/gpg",
					ver:  semver.MustParse("2.4.0"),
				},
				{
					path: "/usr/local/bin/gpg2",
					ver:  semver.MustParse("2.1.11"),
				},
			},
			out: []semver.Version{
				semver.MustParse("1.9.1"),
				semver.MustParse("2.1.11"),
				semver.MustParse("2.4.0"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sort.Sort(byVersion(tc.in))

			require.Equal(t, len(tc.in), len(tc.out))
			for i, v := range tc.out {
				if !tc.in[i].ver.Equals(v) {
					t.Errorf("wrong sort order at %d: %s != %s", i, tc.in[i].ver, v)
				}
			}
		})
	}
}
