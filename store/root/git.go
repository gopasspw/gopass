package root

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/store"
	"github.com/pkg/errors"
)

// GitInit initializes the git repo
func (r *Store) GitInit(name, sk, userName, userEmail string) error {
	store := r.getStore(name)
	return store.GitInit(store.Alias(), sk, userName, userEmail)
}

// GitVersion returns git version information
func (r *Store) GitVersion() semver.Version {
	return r.store.GitVersion()
}

// Git runs arbitrary git commands on this store and all substores
func (r *Store) Git(name string, recurse, force bool, args ...string) error {
	sub := r.getStore(name)
	dispName := name
	if dispName == "" {
		dispName = "root"
	}
	fmt.Println(color.CyanString("[%s] Running git %s", dispName, strings.Join(args, " ")))
	if err := sub.Git(args...); err != nil {
		if err == store.ErrGitNoRemote {
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
		if err := r.mounts[alias].Git(args...); err != nil {
			if err == store.ErrGitNoRemote {
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
