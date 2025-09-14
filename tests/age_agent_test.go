package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgeAgent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows for now")
	}

	ts := newTester(t)
	defer ts.teardown()

	pinentryScript := `#!/bin/sh
echo "D test"
echo "OK"
`
	pinentryPath := filepath.Join(ts.tempDir, "pinentry.sh")
	require.NoError(t, os.WriteFile(pinentryPath, []byte(pinentryScript), 0755))

	gpgAgentConf := `pinentry-program ` + pinentryPath
	require.NoError(t, os.WriteFile(filepath.Join(ts.gpgDir(), "gpg-agent.conf"), []byte(gpgAgentConf), 0644))

	// create a new age identity
	out, err := ts.runCmd([]string{ts.Binary, "age", "identities", "keygen"}, []byte("test\ntest"))
	require.NoError(t, err, out)
	lines := strings.Split(out, "\n")
	require.True(t, len(lines) > 0)
	recipientLine := lines[0]
	recipient := strings.TrimPrefix(recipientLine, "New age identity created: ")

	// initialize a new store with the age backend
	out, err = ts.run("init --crypto age --storage fs " + recipient)
	require.NoError(t, err, out)

	// enable the age agent
	out, err = ts.run("config age.agent-enabled true")
	require.NoError(t, err, out)

	// insert a new secret
	out, err = ts.runCmd([]string{ts.Binary, "insert", "foo"}, []byte("bar\nbar"))
	require.NoError(t, err, out)

	// show the secret, this should start the agent
	out, err = ts.runCmd([]string{ts.Binary, "show", "foo"}, []byte("test\n"))
	require.NoError(t, err, out)
	assert.Equal(t, "bar", out)

	// show the secret again, this should use the cached passphrase from the agent
	out, err = ts.run("show foo")
	require.NoError(t, err, out)
	assert.Equal(t, "bar", out)

	// lock the agent
	out, err = ts.run("age lock")
	require.NoError(t, err, out)

	// show the secret again, this should prompt for the passphrase again
	out, err = ts.runCmd([]string{ts.Binary, "show", "foo"}, []byte("test\n"))
	require.NoError(t, err, out)
	assert.Equal(t, "bar", out)
}
