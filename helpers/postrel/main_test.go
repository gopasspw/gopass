//go:build linux || darwin

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test mustCheckEnv function
func TestMustCheckEnv(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "mock-token")
	os.Setenv("GITHUB_USER", "mock-user")
	os.Setenv("GITHUB_FORK", "mock-fork")

	assert.NotPanics(t, mustCheckEnv)
}

// Test createMilestones function
// TODO: Add test for createMilestones function
// func TestCreateMilestones(t *testing.T) {
// 	ctx := t.Context()
// 	ghCl := newMockGHClient(ctx)
// 	version := semver.MustParse("1.2.3")

// 	err := ghCl.createMilestones(ctx, version)
// 	assert.NoError(t, err)
// }

// Test versionFile function
func TestVersionFile(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.2.3"), 0o644)
	assert.NoError(t, err)

	cwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	version, err := versionFile()
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3", version.String())
}

// Test goVersion function
func TestGoVersion(t *testing.T) {
	for _, v := range []string{"1.15", "1.16", "1.17", "1.25rc2"} {
		t.Logf("Testing goVersion with %s", v)
		assert.NotEmpty(t, goVersion(v), v)
	}
}

// Test syncWorkflows function
func TestSyncWorkflows(t *testing.T) {
	// set up a fake gopass repository holding the source workflows
	src := t.TempDir()
	srcWorkflows := filepath.Join(src, ".github", "workflows")
	assert.NoError(t, os.MkdirAll(srcWorkflows, 0o755))

	for _, wf := range sharedWorkflows {
		content := "name: " + wf.tmpl + "\non:\n  push:\n"
		assert.NoError(t, os.WriteFile(filepath.Join(srcWorkflows, wf.name), []byte(content), 0o644))
	}

	// set up a fake integration repository
	dst := filepath.Join(t.TempDir(), "gopass-hibp")
	assert.NoError(t, os.MkdirAll(dst, 0o755))

	// syncWorkflows resolves the source relative to the current working directory
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, os.Chdir(src))
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	updater := &inUpdater{}
	assert.NoError(t, updater.syncWorkflows(t.Context(), dst))

	for _, wf := range sharedWorkflows {
		content, err := os.ReadFile(filepath.Join(dst, ".github", "workflows", wf.name))
		assert.NoError(t, err)

		want := "name: " + strings.ReplaceAll(wf.tmpl, "%s", "gopass-hibp")
		assert.Contains(t, string(content), want)
	}
}
