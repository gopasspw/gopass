package password

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/fsutil"
)

var (
	// ErrGitInit is returned if git is already initialized
	ErrGitInit = fmt.Errorf("git is already initialized")
	// ErrGitNotInit is returned if git is not initialized
	ErrGitNotInit = fmt.Errorf("git is not initialized")
	// ErrGitNoRemote is returned if git has no origin remote
	ErrGitNoRemote = fmt.Errorf("git has no remote origin")
)

// GitInit initializes this store's git repo and
// recursively calls GitInit on all substores.
func (s *Store) GitInit(signKey string) error {
	if s.isGit() {
		return ErrGitInit
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to initialize git: %s", err)
	}

	if err := s.gitAdd(s.path); err != nil {
		return err
	}
	if err := s.gitCommit("Add current contents of password store."); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(s.path, ".gitattributes"), []byte("*.gpg diff=gpg\n"), fileMode); err != nil {
		return fmt.Errorf("Failed to initialize git: %s", err)
	}
	if err := s.gitAdd(s.path + "/.gitattributes"); err != nil {
		fmt.Println(color.YellowString("Warning: Failed to add .gitattributes to git"))
	}
	if err := s.gitCommit("Configure git repository for gpg file diff."); err != nil {
		fmt.Println(color.YellowString("Warning: Failed to commit .gitattributes to git"))
	}

	cmd = exec.Command("git", "config", "--local", "diff.gpg.binary", "true")
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to initialize git: %s\n", err)
	}

	// set GPG signkey
	if err := s.gitSetSignKey(signKey); err != nil {
		fmt.Printf("Failed to configure Git GPG Commit signing: %s\n", err)
	}

	return nil
}

func (s *Store) gitSetSignKey(sk string) error {
	if sk == "" {
		return fmt.Errorf("SignKey not set")
	}

	cmd := exec.Command("git", "config", "--local", "user.signkey", sk)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "config", "--local", "commit.gpgsign", "true")
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Git runs arbitrary git commands on this store and all substores
func (s *Store) Git(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// isGit returns true if this stores has a .git folder
func (s *Store) isGit() bool {
	return fsutil.IsDir(filepath.Join(s.path, ".git"))
}

// gitAdd adds the listed files to the git index
func (s *Store) gitAdd(files ...string) error {
	if !s.isGit() {
		return ErrGitNotInit
	}

	args := []string{"add"}
	args = append(args, files...)
	cmd := exec.Command("git", args...)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %v", err)
	}

	return nil
}

// gitCommit creates a new git commit with the given commit message
func (s *Store) gitCommit(msg string) error {
	if !s.isGit() {
		return ErrGitNotInit
	}

	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit files to git: %v", err)
	}

	return nil
}

func (s *Store) gitConfigValue(key string) (string, error) {
	if !s.isGit() {
		return "", ErrGitNotInit
	}

	buf := &bytes.Buffer{}

	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = s.path
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// gitPush pushes the repo to it's origin.
// optional arguments: remote and branch
func (s *Store) gitPush(remote, branch string) error {
	if !s.isGit() {
		return ErrGitNotInit
	}

	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		branch = "master"
	}

	if v, err := s.gitConfigValue("remote." + remote + ".url"); err != nil || v == "" {
		return ErrGitNoRemote
	}

	if s.autoPull {
		if err := s.Git("pull", remote, branch); err != nil {
			return err
		}
	}

	return s.Git("push", remote, branch)
}
