package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) { //nolint:paralleltest
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("init")
	assert.Error(t, err)
	assert.Contains(t, out, "Initializing a new password store ...")
	assert.Contains(t, out, "Error: Failed to initialize store")
	assert.Contains(t, out, "failed to read user input: can not select private key without terminal\n")

	ts = newTester(t)
	defer ts.teardown()

	out, err = ts.run("init " + keyID)
	assert.NoError(t, err)
	assert.Contains(t, out, "initialized for")

	ts = newTester(t)
	defer ts.teardown()

	ts.initStore()
	// try to init again
	out, err = ts.run("init " + keyID)
	assert.Error(t, err)

	for _, o := range []string{
		"found already initialized store at ",
		"You can add secondary stores with 'gopass init --path <path to secondary store> --store <mount name>'",
	} {
		assert.Contains(t, out, o)
	}
}
