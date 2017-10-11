package root

import (
	"context"
	"strings"

	"github.com/blang/semver"
	"github.com/justwatchcom/gopass/store"
	"github.com/justwatchcom/gopass/utils/out"
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
	ctxRoot := out.AddPrefix(ctx, "["+dispName+"]")

	out.Cyan(ctxRoot, "Running git %s", strings.Join(args, " "))
	if err := sub.Git(ctxRoot, args...); err != nil {
		if errors.Cause(err) == store.ErrGitNoRemote {
			out.Yellow(ctxRoot, "Has no remote. Skipping")
		} else {
			if !force {
				return errors.Wrapf(err, "failed to run git %s on sub store %s", strings.Join(args, " "), dispName)
			}
			out.Red(ctxRoot, "Failed to run 'git %s'", strings.Join(args, " "))
		}
	}

	// TODO(dschulz) we could properly handle the "recurse to given substores"
	// case ...
	if !recurse || name != "" {
		return nil
	}

	for _, alias := range r.MountPoints() {
		ctx := out.AddPrefix(ctx, "["+alias+"]")
		out.Cyan(ctx, "Running 'git %s'", strings.Join(args, " "))
		if err := r.mounts[alias].Git(ctx, args...); err != nil {
			if errors.Cause(err) == store.ErrGitNoRemote {
				out.Yellow(ctx, "Has no remote. Skipping")
				continue
			}
			if !force {
				return errors.Wrapf(err, "failed to perform git %s on %s", strings.Join(args, " "), alias)
			}
			out.Red(ctx, "Failed to run 'git %s'", strings.Join(args, " "))
		}
	}

	return nil
}
