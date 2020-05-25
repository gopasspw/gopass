package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("generate")
	assert.Error(t, err)
	assert.Equal(t, "\nError: please provide a password name\n", out)

	out, err = ts.run("generate foo 0")
	assert.Error(t, err)
	assert.Equal(t, "\nError: password length must not be zero\n", out)

	out, err = ts.run("generate -p baz 42")
	assert.NoError(t, err)
	lines := strings.Split(out, "\n")
	assert.Len(t, lines, 3)
	assert.Contains(t, out, "The generated password is:")
	assert.Len(t, lines[2], 42)

	_ = os.Setenv("GOPASS_CHARACTER_SET", "a")
	out, err = ts.run("generate -p zab 4")
	assert.NoError(t, err)
	lines = strings.Split(out, "\n")
	assert.Len(t, lines, 3)
	assert.Contains(t, out, "The generated password is:")
	assert.Equal(t, lines[2], "aaaa")
}
