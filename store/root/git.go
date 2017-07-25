package root

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/fatih/color"
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
	store := r.getStore(name)
	fmt.Println(color.CyanString("Running git %s on store %s", strings.Join(args, " "), name))
	if err := store.Git(args...); err != nil {
		if !force {
			return err
		}
		fmt.Println(color.RedString("Failed to run git %+v on store %s", args, name))
	}

	// TODO(dschulz) we could properly handle the "recurse to given substores"
	// case ...
	if !recurse || name != "" {
		return nil
	}

	for _, alias := range r.MountPoints() {
		fmt.Println(color.CyanString("Running git %s on store %s", strings.Join(args, " "), alias))
		if err := r.mounts[alias].Git(args...); err != nil {
			if !force {
				return err
			}
			fmt.Println(color.RedString("Failed to run git %+v on store %s", args, alias))
		}
	}

	return nil
}
