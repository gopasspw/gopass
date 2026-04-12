package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExitCodesFlag(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.run("--exit-codes")
	require.NoError(t, err)

	// Verify a representative selection of codes and names appear in the output.
	assert.Contains(t, out, "0")
	assert.Contains(t, out, "OK")
	assert.Contains(t, out, "10")
	assert.Contains(t, out, "NotFound")
	assert.Contains(t, out, "21")
	assert.Contains(t, out, "Doctor")
}
