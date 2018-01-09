package termio

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"

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

func TestReadLineError(t *testing.T) {
	stdin := strings.NewReader("fo")
	lr := NewReader(iotest.TimeoutReader(stdin))

	line, err := lr.ReadLine()
	assert.Error(t, err)
	assert.Equal(t, "f", line)
}

func TestRead(t *testing.T) {
	stdin := strings.NewReader(`foobarbazzabzabzab`)
	lr := NewReader(stdin)

	b := make([]byte, 10)
	n, err := lr.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, "foobarbazz", string(b))
}

func mustReadLine(r io.Reader) string {
	line, err := NewReader(r).ReadLine()
	if err != nil {
		panic(err)
	}
	return line
}
