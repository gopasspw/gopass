package pwgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxConsec(t *testing.T) {
	t.Parallel()

	// good
	for _, tc := range []string{
		"abcd",
		"foobar",
		"nope",
		"AaAa",
		"aaabbbaaa",
	} {
		assert.True(t, containsMaxConsecutive(tc, 4))
	}
	// bad
	for _, tc := range []string{
		"aaaa",
		"bbb",
		"fooobar",
		"AaaaA",
	} {
		assert.False(t, containsMaxConsecutive(tc, 3))
	}
}

func TestContainsOnly(t *testing.T) {
	t.Parallel()

	// good
	for _, tc := range []string{
		"aBcDeF",
	} {
		assert.True(t, containsOnlyClasses(tc, Upper+Lower))
	}

	// bad
	for _, tc := range []string{
		"aBcDeF3",
	} {
		assert.False(t, containsOnlyClasses(tc, Upper+Lower))
	}
}
