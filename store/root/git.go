package root

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/pkg/errors"
)

// GitInit initializes the git repo
func (r *Store) GitInit(ctx context.Context, name, sk, userName, userEmail string) error {
	ctx, store, _ := r.getStore(ctx, name)
	return store.GitInit(ctx, store.Alias(), sk, userName, userEmail)
}

// GitVersion returns git version information
func (r *Store) GitVersion(ctx context.Context) semver.Version {
	return r.store.GitVersion(ctx)
}

// Git runs arbitrary git commands on this store and all substores
func (r *Store) Git(ctx context.Context, name string, recurse, force bool, args ...string) error {
	ctx, sub, name := r.getStore(ctx, name)
	dispName := name
	if dispName == "" {
		dispName = "root"
	}

	fmt.Println(color.CyanString("[%s] Running git %s", dispName, strings.Join(args, " ")))
	if err := sub.Git(ctx, args...); err != nil {
		if errors.Cause(err) == store.ErrGitNoRemote {
			fmt.Println(color.YellowString("[%s] Has no remote. Skipping", dispName))
		} else {
			if !force {
				return errors.Wrapf(err, "failed to run git %s on sub store %s", strings.Join(args, " "), dispName)
			}
			fmt.Println(color.RedString("[%s] Failed to run 'git %s'", dispName, strings.Join(args, " ")))
		}
	}

	// TODO(dschulz) we could properly handle the "recurse to given substores"
	// case ...
	if !recurse || name != "" {
		return nil
	}

	for _, alias := range r.MountPoints() {
		fmt.Println(color.CyanString("[%s] Running 'git %s'", alias, strings.Join(args, " ")))
		if err := r.mounts[alias].Git(ctx, args...); err != nil {
			if errors.Cause(err) == store.ErrGitNoRemote {
				fmt.Println(color.YellowString("[%s] Has no remote. Skipping", alias))
				continue
			}
			if !force {
				return errors.Wrapf(err, "failed to perform git %s on %s", strings.Join(args, " "), alias)
			}
			fmt.Println(color.RedString("[%s] Failed to run 'git %s'", alias, strings.Join(args, " ")))
		}
	}

	return nil
}
