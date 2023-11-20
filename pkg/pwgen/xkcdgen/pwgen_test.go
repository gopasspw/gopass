package xkcdgen

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandom(t *testing.T) {
	t.Parallel()

	pw := Random()
	if len(pw) < 4 {
		t.Errorf("too short")
	}

	if len(strings.Fields(pw)) < 4 {
		t.Errorf("too few words")
	}
}

func TestRandomLengthDelim(t *testing.T) {
	t.Parallel()

	_, err := RandomLengthDelim(10, " ", "cn_ZH", false, false)
	require.Error(t, err)
}
