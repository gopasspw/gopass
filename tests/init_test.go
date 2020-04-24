package tests

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows.")
	}
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("init")
	assert.Error(t, err)
	assert.Equal(t, "[init] Initializing a new password store ...\n\nError: failed to initialize store: failed to read user input: can not select private key without terminal\n", out)

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
		"Found already initialized store at ",
		"You can add secondary stores with gopass init --path <path to secondary store> --store <mount name>",
	} {
		assert.Contains(t, out, o)
	}
}
