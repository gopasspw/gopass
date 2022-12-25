package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("init")
	assert.NoError(t, err)
	assert.Contains(t, out, "Initializing a new password store ...")
	assert.Contains(t, out, "initialized")

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
