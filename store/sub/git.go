package sub

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
	"github.com/justwatchcom/gopass/store"
)

func (s *Store) gitCmd(name string, args ...string) error {
	cmd := exec.Command("git", args[0:]...)
	cmd.Dir = s.path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if s.debug {
		fmt.Printf("[DEBUG] store.%s: %s %+v\n", name, cmd.Path, cmd.Args)
	}
	return cmd.Run()
}

// GitInit initializes this store's git repo and
// recursively calls GitInit on all substores.
func (s *Store) GitInit(alias, signKey, userName, userEmail string) error {
	// the git repo may be empty (i.e. no branches, cloned from a fresh remote)
	// or already initialized. Only run git init if the folder is completely empty
	if !s.isGit() {
		if err := s.gitCmd("GitInit", "init"); err != nil {
			return fmt.Errorf("Failed to initialize git: %s", err)
		}
	}

	if err := s.gitCmd("GitInit", "config", "--local", "user.name", userName); err != nil {
		return err
	}
	if err := s.gitCmd("GitInit", "config", "--local", "user.email", userEmail); err != nil {
		return err
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

	// setup for proper diffs
	if err := s.gitCmd("GitInit", "config", "--local", "diff.gpg.binary", "true"); err != nil {
		color.Yellow("Error while initializing git: %s\n", err)
	}
	if err := s.gitCmd("GitInit", "config", "--local", "diff.gpg.textconv", "gpg --no-tty --decrypt"); err != nil {
		color.Yellow("Error while initializing git: %s\n", err)
	}

	// set GPG signkey
	if err := s.gitSetSignKey(signKey); err != nil {
		color.Yellow("Failed to configure Git GPG Commit signing: %s\n", err)
	}

	return nil
}

func (s *Store) gitSetSignKey(sk string) error {
	if sk == "" {
		return fmt.Errorf("SignKey not set")
	}

	if err := s.gitCmd("gitSetSignKey", "config", "--local", "user.signingkey", sk); err != nil {
		return err
	}

	return s.gitCmd("gitSetSignKey", "config", "--local", "commit.gpgsign", "true")
}

// Git runs arbitrary git commands on this store
func (s *Store) Git(args ...string) error {
	return s.gitCmd("Git", args...)
}

// isGit returns true if this stores has a .git folder
func (s *Store) isGit() bool {
	// TODO(dschulz) we may want to check if the folder actually contains
	// an initialized git setup
	return fsutil.IsDir(filepath.Join(s.path, ".git"))
}

// gitAdd adds the listed files to the git index
func (s *Store) gitAdd(files ...string) error {
	if !s.isGit() {
		return store.ErrGitNotInit
	}
	for i := range files {
		files[i] = strings.TrimPrefix(files[i], s.path+"/")
	}

	args := []string{"add", "--all"}
	args = append(args, files...)

	return s.gitCmd("gitAdd", args...)
}

// gitCommit creates a new git commit with the given commit message
func (s *Store) gitCommit(msg string) error {
	if !s.isGit() {
		return store.ErrGitNotInit
	}

	return s.gitCmd("gitCommit", "commit", "-m", msg)
}

func (s *Store) gitConfigValue(key string) (string, error) {
	if !s.isGit() {
		return "", store.ErrGitNotInit
	}

	buf := &bytes.Buffer{}

	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = s.path
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	if s.debug {
		fmt.Printf("store.gitConfigValue: %s %+v\n", cmd.Path, cmd.Args)
	}
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

// gitPush pushes the repo to it's origin.
// optional arguments: remote and branch
func (s *Store) gitPush(remote, branch string) error {
	if !s.isGit() {
		return store.ErrGitNotInit
	}

	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		branch = "master"
	}

	if v, err := s.gitConfigValue("remote." + remote + ".url"); err != nil || v == "" {
		return store.ErrGitNoRemote
	}

	if s.autoPull {
		if err := s.Git("pull", remote, branch); err != nil {
			fmt.Println(color.YellowString("Failed to pull before git push: %s", err))
		}
	}

	return s.Git("push", remote, branch)
}
