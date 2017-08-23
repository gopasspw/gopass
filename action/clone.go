package action

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Clone will fetch and mount a new password store from a git repo
func (s *Action) Clone(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return errors.Errorf("Usage: %s clone repo [mount]", s.Name)
	}

	repo := c.Args()[0]
	mount := ""
	if len(c.Args()) > 1 {
		mount = c.Args()[1]
	}

	path := c.String("path")
	if path == "" {
		path = config.PwStoreDir(mount)
	}

	if mount == "" && s.Store.Initialized() {
		return s.exitError(ExitAlreadyInitialized, nil, "Can not clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `%s clone %s sub`", repo, s.Name, repo)
	}

	// clone repo
	if err := gitClone(repo, path); err != nil {
		return s.exitError(ExitGit, err, "failed to clone repo '%s' to '%s'", repo, path)
	}

	// add mount
	if mount != "" {
		if !s.Store.Initialized() {
			return s.exitError(ExitNotInitialized, nil, "Root-Store is not initialized. Clone or init root store first")
		}
		if err := s.Store.AddMount(mount, path); err != nil {
			return s.exitError(ExitMount, err, "Failed to add mount: %s", err)
		}
		fmt.Printf("Mounted password store %s at mount point `%s` ...\n", path, mount)
	}

	// save new mount in config file
	if err := s.Store.Config().Save(); err != nil {
		return s.exitError(ExitIO, err, "Failed to update config: %s", err)
	}

	fmt.Println(color.GreenString("Your password store is ready to use! Have a look around: `%s %s`\n", s.Name, mount))

	return nil
}

func gitClone(repo, path string) error {
	if fsutil.IsDir(path) {
		return errors.Errorf("%s is a directory that already exists", path)
	}

	fmt.Printf("Cloning repository %s to %s ...\n", repo, path)

	cmd := exec.Command("git", "clone", repo, path)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
