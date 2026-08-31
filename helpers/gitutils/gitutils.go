package gitutils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var Verbose = false

func InitGitDirWithRemote(t *testing.T, baseDir string) string {
	t.Helper()

	remoteDir := filepath.Join(baseDir, "remote")
	InitGitBare(t, remoteDir)

	cmd := exec.Command("git", "clone", remoteDir, "repo")
	cmd.Dir = baseDir
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	require.NoError(t, cmd.Run())

	dir := filepath.Join(baseDir, "repo")
	PopulateGitDir(t, dir)

	return dir
}

func initGitWithArgs(t *testing.T, dir string, extraArgs ...string) string {
	t.Helper()

	// make sure the directory exists
	require.NoError(t, os.MkdirAll(dir, 0o755))

	// git init -b master
	args := []string{
		"init",
		"-b",
		"master",
	}
	args = append(args, extraArgs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	require.NoError(t, cmd.Run())

	return dir
}

func InitGitDir(t *testing.T, dir string) string {
	t.Helper()

	dir = initGitWithArgs(t, dir)
	PopulateGitDir(t, dir)

	return dir
}

func InitGitBare(t *testing.T, dir string) string {
	t.Helper()

	return initGitWithArgs(t, dir, "--bare")
}

func PopulateGitDir(t *testing.T, dir string) {
	t.Helper()
	// Create a file in the repo so we have something to commit and create a root commit from.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("test content"), 0o644))

	// Add the file to the index.
	cmd := exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	require.NoError(t, cmd.Run())

	// Commit the file.
	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	require.NoError(t, cmd.Run())
}

func IsGitClean(dir string) bool {
	if sv := os.Getenv("GOPASS_FORCE_CLEAN"); sv != "" {
		return true
	}

	cmd := exec.Command("git", "diff", "--stat")
	cmd.Dir = dir
	buf, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	if strings.TrimSpace(string(buf)) != "" {
		fmt.Printf("❌ Git in %s is not clean: %q\n", dir, string(buf))

		return false
	}

	return true
}

func GitCoMaster(dir string) error {
	cmd := exec.Command("git", "checkout", "master")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	return cmd.Run()
}

func GitCoBranch(dir, branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	return cmd.Run()
}

func GitDelBranch(dir, branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	return cmd.Run()
}

func GitPom(dir string) error {
	cmd := exec.Command("git", "pull", "origin", "master")
	cmd.Dir = dir
	// hide long pull output unless an error occurs
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	if err := cmd.Run(); err != nil {
		fmt.Println(buf.String())

		return err
	}

	return nil
}

// GitSyncMaster makes sure the repo at dir is on master and in sync with
// origin/master. It fetches the remote and hard resets the local master
// to origin/master. It refuses to run on a dirty worktree.
func GitSyncMaster(dir string) error {
	if !IsGitClean(dir) {
		return fmt.Errorf("git worktree at %s is dirty, refusing to sync", dir)
	}

	steps := [][]string{
		{"checkout", "master"},
		{"fetch", "origin"},
		{"reset", "--hard", "origin/master"},
	}
	for _, args := range steps {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		buf := &bytes.Buffer{}
		cmd.Stdout = buf
		cmd.Stderr = buf
		if Verbose {
			fmt.Printf("Running command: %s\n", cmd)
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git %s failed: %s: %w", strings.Join(args, " "), strings.TrimSpace(buf.String()), err)
		}
	}

	return nil
}

func GitAdd(dir string, files ...string) error {
	args := []string{"add"}
	args = append(args, files...)

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	return cmd.Run()
}

func GitCommitAndPush(dir, tag string) error {
	cmd := exec.Command("git", "commit", "-a", "-s", "-m", "Update to "+tag)
	cmd.Dir = dir
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if Verbose {
		fmt.Printf("Running command: %s\n", cmd)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %s: %w", strings.TrimSpace(buf.String()), err)
	}

	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = dir
	buf = &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if Verbose {
		fmt.Printf("Running command: %s\n", cmd)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %s: %w", strings.TrimSpace(buf.String()), err)
	}

	return nil
}

func GitCommit(dir, commitMsg string, files ...string) error {
	args := []string{"add"}
	args = append(args, files...)

	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}

	cmd.Dir = dir
	if Verbose {
		fmt.Printf("Running command: %s\n", cmd)
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-s", "-m", commitMsg)
	cmd.Stderr = os.Stderr
	if Verbose {
		cmd.Stdout = os.Stdout
		fmt.Printf("Running command: %s\n", cmd)
	}
	cmd.Dir = dir

	return cmd.Run()
}

func GitPush(remote, branch string) error {
	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		branch = "master"
	}

	cmd := exec.Command("git", "push", remote, branch)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if Verbose {
		fmt.Printf("Running command: %s\n", cmd)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push %s to %s: %s: %w", branch, remote, strings.TrimSpace(buf.String()), err)
	}

	return nil
}

func GitTagAndPush(dir string, tag string) error {
	cmd := exec.Command("git", "tag", "-m", "'Tag "+tag+"'", tag)
	cmd.Dir = dir
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if Verbose {
		fmt.Printf("Running command: %s\n", cmd)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to tag: %s: %w", strings.TrimSpace(buf.String()), err)
	}

	cmd = exec.Command("git", "push", "origin", tag)
	cmd.Dir = dir
	buf = &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	if Verbose {
		fmt.Printf("Running command: %s\n", cmd)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push tag: %s: %w", strings.TrimSpace(buf.String()), err)
	}

	return nil
}

func GitHasTag(dir string, tag string) bool {
	cmd := exec.Command("git", "rev-parse", tag)
	cmd.Dir = dir
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if Verbose {
		fmt.Printf("Running command: %s\n", cmd)
	}

	return cmd.Run() == nil
}
