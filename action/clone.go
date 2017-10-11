package action

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/justwatchcom/gopass/config"
	"github.com/justwatchcom/gopass/utils/fsutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// Clone will fetch and mount a new password store from a git repo
func (s *Action) Clone(ctx context.Context, c *cli.Context) error {
	if len(c.Args()) < 1 {
		return errors.Errorf("Usage: %s clone repo [mount]", s.Name)
	}

	repo := c.Args()[0]
	mount := ""
	if len(c.Args()) > 1 {
		mount = c.Args()[1]
	}

	path := c.String("path")

	return s.clone(ctx, repo, mount, path)
}

func (s *Action) clone(ctx context.Context, repo, mount, path string) error {
	if path == "" {
		path = config.PwStoreDir(mount)
	}
	if mount == "" && s.Store.Initialized() {
		return s.exitError(ctx, ExitAlreadyInitialized, nil, "Can not clone %s to the root store, as this store is already initialized. Please try cloning to a submount: `%s clone %s sub`", repo, s.Name, repo)
	}

	// clone repo
	if err := gitClone(ctx, repo, path); err != nil {
		return s.exitError(ctx, ExitGit, err, "failed to clone repo '%s' to '%s'", repo, path)
	}

	// add mount
	if mount != "" {
		if !s.Store.Initialized() {
			return s.exitError(ctx, ExitNotInitialized, nil, "Root-Store is not initialized. Clone or init root store first")
		}
		if err := s.Store.AddMount(ctx, mount, path); err != nil {
			return s.exitError(ctx, ExitMount, err, "Failed to add mount: %s", err)
		}
		fmt.Printf("Mounted password store %s at mount point `%s` ...\n", path, mount)
	}

	// save new mount in config file
	if err := s.cfg.Save(); err != nil {
		return s.exitError(ctx, ExitIO, err, "Failed to update config: %s", err)
	}

	out.Green(ctx, "Your password store is ready to use! Have a look around: `%s %s`\n", s.Name, mount)

	return nil
}

func gitClone(ctx context.Context, repo, path string) error {
	if fsutil.IsDir(path) {
		return errors.Errorf("%s is a directory that already exists", path)
	}

	fmt.Printf("Cloning repository %s to %s ...\n", repo, path)

	cmd := exec.CommandContext(ctx, "git", "clone", repo, path)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
