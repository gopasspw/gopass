package termio

import (
	"context"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadLines(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	ctx := context.Background()
	stdin := strings.NewReader("fo")
	lr := NewReader(ctx, iotest.TimeoutReader(stdin))

	line, err := lr.ReadLine()
	require.Error(t, err)
	assert.Equal(t, "f", line)
}

func TestRead(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	stdin := strings.NewReader(`foobarbazzabzabzab`)
	lr := NewReader(ctx, stdin)

	b := make([]byte, 10)
	n, err := lr.Read(b)
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, "foobarbazz", string(b))
}

func mustReadLine(r io.Reader) string {
	ctx := context.Background()

	line, err := NewReader(ctx, r).ReadLine()
	if err != nil {
		panic(err)
	}

	return line
}
