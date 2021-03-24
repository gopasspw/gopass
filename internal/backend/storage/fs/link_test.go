package fs

import (
	"testing"

	"gotest.tools/assert"
)

func TestLongestCommonPrefix(t *testing.T) {
	for _, tc := range []struct {
		Src    string
		Dst    string
		Prefix string
	}{
		{
			Src:    "foo/bar/baz/zab.txt",
			Dst:    "foo/baz/foo.txt",
			Prefix: "foo",
		},
	} {
		prefix := longestCommonPrefix(tc.Src, tc.Dst)
		assert.Equal(t, tc.Prefix, prefix)
	}
}

func TestAddRel(t *testing.T) {
	for _, tc := range []struct {
		Src string
		Dst string
		Out string
	}{
		{
			Src: "bar/baz.txt",
			Dst: "baz/foo.txt",
			Out: "../bar/baz.txt",
		},
	} {
		assert.Equal(t, tc.Out, addRel(tc.Src, tc.Dst))
	}
}
