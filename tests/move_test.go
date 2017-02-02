package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMove(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	_, err := ts.run("move")
	assert.Error(t, err)

	ts.initializeStore()

	out, err := ts.run("move")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: gopass mv old-path new-path\n", out)

	out, err = ts.run("move foo")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: gopass mv old-path new-path\n", out)

	out, err = ts.run("move foo bar")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Entry is not in the password store\n", out)

	ts.initializeSecrets()

	out, err = ts.run("move foo bar")
	assert.NoError(t, err)

	out, err = ts.run("move foo/bar foo/baz")
	assert.Equal(t, "\nError: Entry is not in the password store\n", out)

	_, err = ts.run("show bar/bar")
	assert.NoError(t, err)

	_, err = ts.run("show baz")
	assert.NoError(t, err)
}
