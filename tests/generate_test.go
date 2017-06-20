package tests

import (
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
	assert.Equal(t, "Which name do you want to use? []: \nError: provide a password name\n", out)

	out, err = ts.run("generate foo 0")
	assert.Error(t, err)
	assert.Equal(t, "\nError: password length must be bigger than 0\n", out)

	out, err = ts.run("generate baz 42")
	assert.NoError(t, err)
	lines := strings.Split(out, "\n")
	assert.Len(t, lines, 2)
	assert.Equal(t, "The generated password for baz is:", lines[0])
	assert.Len(t, lines[1], 42)
}
