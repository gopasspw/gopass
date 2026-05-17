//go:build linux

package main

import (
	"os"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/helpers/gitutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func chdir(t *testing.T, dir string) {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(cwd))
	})
}

func withArgs(t *testing.T, args ...string) {
	t.Helper()

	oldArgs := os.Args
	os.Args = args
	t.Cleanup(func() {
		os.Args = oldArgs
	})
}

func TestParseReleaseArgs(t *testing.T) {
	args := parseReleaseArgs([]string{"release", "--dry-run", "v1.3.0-rc.2", "v1.3.0-rc.1"})

	assert.True(t, args.dryRun)
	assert.Equal(t, "1.3.0-rc.2", args.nextVersion)
	assert.Equal(t, "1.3.0-rc.1", args.prevVersion)
}

// TestGetVersions tests the getVersions function.
func TestGetVersions(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir := t.TempDir()

	dir := gitutils.InitGitDirWithRemote(t, tempDir)
	// Change the working directory to the temporary directory.
	chdir(t, dir)
	withArgs(t, "release")

	// Create a mock VERSION file.
	err := os.WriteFile("VERSION", []byte("1.2.3\n"), 0o644)
	assert.NoError(t, err)

	// Create a git tag.
	require.NoError(t, gitutils.GitTagAndPush(dir, "v1.2.3"))

	// Call the getVersions function.
	prevVer, nextVer := getVersions()

	// Assert the versions.
	assert.Equal(t, "1.2.3", prevVer.String())
	assert.Equal(t, "1.2.4", nextVer.String())
}

func TestGetVersionsReleaseCandidateStartsFromLatestStable(t *testing.T) {
	tempDir := t.TempDir()
	dir := gitutils.InitGitDirWithRemote(t, tempDir)
	chdir(t, dir)
	withArgs(t, "release", "v1.3.0-rc.1")

	require.NoError(t, os.WriteFile("VERSION", []byte("1.2.3\n"), 0o644))
	require.NoError(t, gitutils.GitTagAndPush(dir, "v1.2.3"))

	prevVer, nextVer := getVersions()

	assert.Equal(t, "1.2.3", prevVer.String())
	assert.Equal(t, "1.3.0-rc.1", nextVer.String())
}

func TestGetVersionsReleaseCandidateUsesPreviousRC(t *testing.T) {
	tempDir := t.TempDir()
	dir := gitutils.InitGitDirWithRemote(t, tempDir)
	chdir(t, dir)
	withArgs(t, "release", "v1.3.0-rc.2")

	require.NoError(t, os.WriteFile("VERSION", []byte("1.3.0-rc.1\n"), 0o644))
	require.NoError(t, gitutils.GitTagAndPush(dir, "v1.2.3"))
	require.NoError(t, gitutils.GitTagAndPush(dir, "v1.3.0-rc.1"))

	prevVer, nextVer := getVersions()

	assert.Equal(t, "1.3.0-rc.1", prevVer.String())
	assert.Equal(t, "1.3.0-rc.2", nextVer.String())
}

func TestGetVersionsForArgsIgnoresDryRun(t *testing.T) {
	tempDir := t.TempDir()
	dir := gitutils.InitGitDirWithRemote(t, tempDir)
	chdir(t, dir)

	require.NoError(t, os.WriteFile("VERSION", []byte("1.2.3\n"), 0o644))
	require.NoError(t, gitutils.GitTagAndPush(dir, "v1.2.3"))

	prevVer, nextVer := getVersionsForArgs(releaseArgs{dryRun: true})

	assert.Equal(t, "1.2.3", prevVer.String())
	assert.Equal(t, "1.2.4", nextVer.String())
}

// TestWriteVersion tests the writeVersion function.
func TestWriteVersion(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir := t.TempDir()
	// Change the working directory to the temporary directory.
	chdir(t, tempDir)

	// Call the writeVersion function.
	err := writeVersion(semver.MustParse("1.2.3"))
	assert.NoError(t, err)

	// Read the VERSION file.
	data, err := os.ReadFile("VERSION")
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3\n", string(data))
}

// TestWriteVersionGo tests the writeVersionGo function.
func TestWriteVersionGo(t *testing.T) {
	// Create a temporary directory for the test.
	tempDir := t.TempDir()
	// Change the working directory to the temporary directory.
	chdir(t, tempDir)

	// Call the writeVersionGo function.
	err := writeVersionGo(semver.MustParse("1.2.3"))
	assert.NoError(t, err)

	// Read the version.go file.
	data, err := os.ReadFile("version.go")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Major: 1")
	assert.Contains(t, string(data), "Minor: 2")
	assert.Contains(t, string(data), "Patch: 3")
}
