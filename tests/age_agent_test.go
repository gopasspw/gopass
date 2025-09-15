package tests

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgeAgent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows for now")
	}

	ts := newTester(t)
	defer ts.teardown()

	// create a new age identity
	out, err := ts.runCmd([]string{ts.Binary, "age", "identities", "keygen", "--password", "foo"}, []byte("test\ntest\n"))
	require.NoError(t, err, out)
}
