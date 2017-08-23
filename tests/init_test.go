package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("init")
	assert.Error(t, err)
	assert.Equal(t, "\nError: failed to initialized store: failed to read user input: no interaction without terminal\n", out)

	ts = newTester(t)
	defer ts.teardown()

	out, err = ts.run("init " + keyID)
	assert.NoError(t, err)
	assert.Contains(t, out, "Please enter a user name for password store git config [John Doe]:")

	ts = newTester(t)
	defer ts.teardown()

	ts.initStore()
	// try to init again
	out, err = ts.run("init " + keyID)
	assert.Error(t, err)
	for _, o := range []string{
		"Found already initialized store at /tmp/gopass-",
		"You can add secondary stores with gopass init --path <path to secondary store> --store <mount name>",
	} {
		assert.Contains(t, out, o)
	}
}
