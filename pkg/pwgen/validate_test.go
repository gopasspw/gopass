package pwgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxConsec(t *testing.T) {
	// good
	for _, tc := range []string{
		"abcd",
		"foobar",
		"nope",
		"AaAa",
		"aaabbbaaa",
	} {
		assert.Equal(t, true, containsMaxConsecutive(tc, 4))
	}
	// bad
	for _, tc := range []string{
		"aaaa",
		"bbb",
		"fooobar",
		"AaaaA",
	} {
		assert.Equal(t, false, containsMaxConsecutive(tc, 3))
	}
}

func TestContainsOnly(t *testing.T) {
	// good
	for _, tc := range []string{
		"aBcDeF",
	} {
		assert.Equal(t, true, containsOnlyClasses(tc, Upper+Lower))
	}
}
