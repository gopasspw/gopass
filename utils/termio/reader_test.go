package termio

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLines(t *testing.T) {
	for _, tc := range [][]string{
		{"foo", "bar"},
		{"foo", "bar", "", "baz"},
		{"foo", "µ"},
		{"µ", "ĸ", "aŧ", "¶a"},
	} {
		stdin := strings.NewReader(strings.Join(tc, "\n"))
		for _, s := range tc {
			assert.Equal(t, s, mustReadLine(stdin))
		}
	}
}

func mustReadLine(r io.Reader) string {
	line, err := NewReader(r).ReadLine()
	if err != nil {
		panic(err)
	}
	return line
}
