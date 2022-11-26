package gitconfig

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrim(t *testing.T) {
	t.Parallel()

	for _, tc := range [][]string{
		{" a ", "b       ", "\tc\n"},
	} {
		trim(tc)
		for _, e := range tc {
			assert.Equal(t, strings.TrimSpace(e), e)
		}
	}
}

func TestSplitKey(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in         string
		section    string
		subsection string
		key        string
	}{
		{
			in:         "url.git@gist.github.com:.pushinsteadof",
			section:    "url",
			subsection: "git@gist.github.com:",
			key:        "pushinsteadof",
		},
		{
			in:      "gc.auto",
			section: "gc",
			key:     "auto",
		},
	} {
		sec, sub, key := splitKey(tc.in)
		assert.Equal(t, tc.section, sec, sec)
		assert.Equal(t, tc.subsection, sub, sub)
		assert.Equal(t, tc.key, key, key)
	}
}
