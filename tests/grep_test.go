package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrep(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("grep")
	assert.Error(t, err)
	assert.Equal(t, "\nError: Usage: gopass grep arg\n", out)

	out, err = ts.run("grep BOOM")
	assert.NoError(t, err)
	assert.Zero(t, out)

	ts.initSecrets("")

	out, err = ts.run("grep moar")
	assert.NoError(t, err)
	assert.Equal(t, "fixed/secret:\nmoar", out)
}
