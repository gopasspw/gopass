package action

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/fsutil"
	"github.com/urfave/cli"
)

// Clone will fetch and mount a new password store from a git repo
func (s *Action) Clone(c *cli.Context) error {
	if len(c.Args()) < 1 {
		return fmt.Errorf("Usage: gopass clone repo [mount]")
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
		return fmt.Errorf("Can not clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `gopass clone %s sub`", repo, repo)
	}

	// clone repo
	if err := gitClone(repo, path); err != nil {
		return err
	}

	// add mount
	if mount != "" {
		if !s.Store.Initialized() {
			return fmt.Errorf("Root-Store is not initialized. Clone or init root store first")
		}
		if err := s.Store.AddMount(mount, path); err != nil {
			return fmt.Errorf("Failed to add mount: %s", err)
		}
		fmt.Printf("Mounted password store %s at mount point `%s` ...\n", path, mount)
	}

	// save new mount in config file
	if err := s.Store.Config().Save(); err != nil {
		return fmt.Errorf("Failed to update config: %s", err)
	}

	fmt.Printf("Your password store is ready to use! Has a look around: `gopass %s`\n", mount)

	return nil
}

func gitClone(repo, path string) error {
	if fsutil.IsDir(path) {
		return fmt.Errorf("%s is a directory that already exists", path)
	}

	fmt.Printf("Cloning repository %s to %s ...\n", repo, path)

	cmd := exec.Command("git", "clone", repo, path)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
