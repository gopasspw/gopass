package tests

import (
	"os"
	"runtime"
	"testing"

	"github.com/gopasspw/gopass/tests/agecan"
	"github.com/stretchr/testify/require"
)

func TestAgeAgent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows for now")
	}

	ts := newTester(t)
	defer ts.teardown()

	out, err := ts.runCmd([]string{ts.Binary, "age", "identities", "keygen", "--password", "foo"}, []byte("test\ntest\n"))
	require.NoError(t, err, out)
}

func TestAgeAgentUnlockAndDecrypt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows for now")
	}

	ts := newAgeTester(t)
	defer ts.teardown()

	// Initialize with passphrase-protected identity.
	// GOPASS_AGE_PASSWORD and GOPASS_AGE_STDIN_PASSPHRASE are already set by newAgeTester.
	// initAgeStore also starts the age agent via tryStartAgent.
	ts.initAgeStore(true)

	// 1. Insert a secret (agent already has identities in memory from init).
	out, err := ts.runCmd([]string{ts.Binary, "insert", "-m", "test/secret"}, []byte("my secret password\n"))
	require.NoError(t, err, out)

	// 2. Verify baseline works without passphrase (agent is unlocked).
	require.NoError(t, os.Unsetenv("GOPASS_AGE_PASSWORD"))
	out, err = ts.runCmd([]string{ts.Binary, "show", "test/secret"}, nil)
	require.NoError(t, err, out)
	require.Contains(t, out, "my secret password")

	// 3. Lock the agent.
	out, err = ts.runCmd([]string{ts.Binary, "age", "agent", "lock"}, nil)
	require.NoError(t, err, out)

	// 4. Verify agent is locked.
	out, err = ts.runCmd([]string{ts.Binary, "age", "agent", "status"}, nil)
	require.NoError(t, err, out)
	require.Contains(t, out, "locked")

	// 4a. Verify gopass show works with passphrase via stdin when agent is locked.
	//     GOPASS_AGE_PASSWORD is unset (from step 2), so the fallback path
	//     uses GOPASS_AGE_STDIN_PASSPHRASE to read from stdin.
	t.Setenv("GOPASS_AGE_STDIN_PASSPHRASE", "1")
	out, err = ts.runCmd([]string{ts.Binary, "show", "test/secret"}, []byte(agecan.TestPin+"\n"))
	require.NoError(t, err, out)
	require.Contains(t, out, "my secret password")
	require.NoError(t, os.Unsetenv("GOPASS_AGE_STDIN_PASSPHRASE"))

	// 5. Unlock with passphrase via stdin (GOPASS_AGE_STDIN_PASSPHRASE forces CLI fallback).
	t.Setenv("GOPASS_AGE_STDIN_PASSPHRASE", "1")
	out, err = ts.runCmd([]string{ts.Binary, "age", "agent", "unlock"}, []byte(agecan.TestPin+"\n"))
	require.NoError(t, err, out)
	require.Contains(t, out, "Age agent unlocked and identities reloaded")
	require.NoError(t, os.Unsetenv("GOPASS_AGE_STDIN_PASSPHRASE"))

	// 6. Verify show works without re-prompting for PIN.
	require.NoError(t, os.Unsetenv("GOPASS_AGE_PASSWORD"))
	out, err = ts.runCmd([]string{ts.Binary, "show", "test/secret"}, nil)
	require.NoError(t, err, out)
	require.Contains(t, out, "my secret password")

	// 7. Cleanup.
	out, err = ts.runCmd([]string{ts.Binary, "age", "agent", "stop"}, nil)
	require.NoError(t, err, out)
}
