//go:build linux
// +build linux

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/helpers/gitutils"
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
// 	ctx := context.Background()
// 	ghCl := newMockGHClient(ctx)
// 	version := semver.MustParse("1.2.3")

// 	err := ghCl.createMilestones(ctx, version)
// 	assert.NoError(t, err)
// }

// Test updateGopasspw function
func TestUpdateGopasspw(t *testing.T) {
	td := t.TempDir()
	version := semver.MustParse("1.2.3")

	dir := gitutils.InitGitDirWithRemote(t, td)

	err := os.WriteFile(filepath.Join(dir, "index.tpl"), []byte("Version: {{.Version}}"), 0o644)
	assert.NoError(t, err)

	err = updateGopasspw(dir, version)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "index.html"))
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Version: 1.2.3")
}

// Test versionFile function
func TestVersionFile(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.2.3"), 0o644)
	assert.NoError(t, err)

	os.Chdir(dir)
	version, err := versionFile()
	assert.NoError(t, err)
	assert.Equal(t, "1.2.3", version.String())
}

// Test goVersion function
func TestGoVersion(t *testing.T) {
	version := goVersion()
	assert.NotEmpty(t, version)
}

// Test updateWorkflows function
func TestUpdateWorkflows(t *testing.T) {
	dir := t.TempDir()
	gitutils.InitGitDir(t, dir)

	err := os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0o755)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, ".github", "workflows", "test.yml"), []byte("go-version: 1.15"), 0o644)
	assert.NoError(t, err)

	updater := &inUpdater{
		goVer: "1.16",
	}

	err = updater.updateWorkflows(context.Background(), dir)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, ".github", "workflows", "test.yml"))
	assert.NoError(t, err)
	assert.Contains(t, string(content), "go-version: 1.16")
}
