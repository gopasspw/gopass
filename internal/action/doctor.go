package action

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/backend/storage/gitfs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/urfave/cli/v2"
)

// Doctor checks the gopass installation for common issues and prints a
// diagnostic report. It exits with a non-zero status if any check fails.
func (s *Action) Doctor(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	verbose := c.Bool("verbose")

	type check struct {
		name string
		fn   func(context.Context) error
	}

	checks := []check{
		{"GPG binary", s.doctorCheckGPG},
		{"age binary", s.doctorCheckAge},
		{"git binary", s.doctorCheckGit},
		{"git identity", s.doctorCheckGitIdentity},
		{"store permissions", s.doctorCheckStorePermissions},
		{"recipient keys", s.doctorCheckRecipients},
	}

	failed := 0
	for _, ch := range checks {
		if err := ch.fn(ctx); err != nil {
			out.Errorf(ctx, "%s: %s", ch.name, err)
			failed++
		} else if verbose {
			out.OKf(ctx, "%s", ch.name)
		}
	}

	// Remote check is advisory — warns but does not fail the command.
	s.doctorCheckGitRemote(ctx)

	if failed > 0 {
		return exit.Error(exit.Doctor, nil, "doctor found %d failing check(s)", failed)
	}

	out.OKf(ctx, "All checks passed")

	return nil
}

// doctorMountPoints returns the root mount ("") followed by all sub-store mount points.
func (s *Action) doctorMountPoints() []string {
	return append([]string{""}, s.Store.MountPoints()...)
}

// doctorCheckGPG fails if any store uses GPG encryption but the gpg binary is not found.
func (s *Action) doctorCheckGPG(_ context.Context) error {
	for _, mp := range s.doctorMountPoints() {
		sub, err := s.Store.GetSubStore(mp)
		if err != nil || sub == nil {
			continue
		}

		if sub.Crypto().Name() == "gpg" {
			if _, err := exec.LookPath("gpg"); err != nil {
				return fmt.Errorf("gpg binary not found in PATH (required by store %q)", doctorStoreLabel(mp))
			}

			return nil
		}
	}

	return nil
}

// doctorCheckAge fails if any store uses age encryption but the age binary is not found.
func (s *Action) doctorCheckAge(_ context.Context) error {
	for _, mp := range s.doctorMountPoints() {
		sub, err := s.Store.GetSubStore(mp)
		if err != nil || sub == nil {
			continue
		}

		if sub.Crypto().Name() == "age" {
			if _, err := exec.LookPath("age"); err != nil {
				return fmt.Errorf("age binary not found in PATH (required by store %q)", doctorStoreLabel(mp))
			}

			return nil
		}
	}

	return nil
}

// doctorCheckGit fails if any store uses the git storage backend but the git binary is not found.
func (s *Action) doctorCheckGit(_ context.Context) error {
	for _, mp := range s.doctorMountPoints() {
		sub, err := s.Store.GetSubStore(mp)
		if err != nil || sub == nil {
			continue
		}

		if _, ok := sub.Storage().(*gitfs.Git); ok {
			if _, err := exec.LookPath("git"); err != nil {
				return fmt.Errorf("git binary not found in PATH (required by store %q)", doctorStoreLabel(mp))
			}

			return nil
		}
	}

	return nil
}

// doctorCheckGitIdentity fails if any git-backed store is missing user.name or user.email in its git config.
func (s *Action) doctorCheckGitIdentity(ctx context.Context) error {
	for _, mp := range s.doctorMountPoints() {
		sub, err := s.Store.GetSubStore(mp)
		if err != nil || sub == nil {
			continue
		}

		g, ok := sub.Storage().(*gitfs.Git)
		if !ok {
			continue
		}

		for _, key := range []string{"user.name", "user.email"} {
			v, err := g.ConfigGet(ctx, key)
			if err != nil || v == "" {
				return fmt.Errorf("git config %q not set for store %q", key, doctorStoreLabel(mp))
			}
		}
	}

	return nil
}

// doctorCheckStorePermissions fails if any store directory is missing or world-writable.
func (s *Action) doctorCheckStorePermissions(_ context.Context) error {
	for _, mp := range s.doctorMountPoints() {
		sub, err := s.Store.GetSubStore(mp)
		if err != nil || sub == nil {
			continue
		}

		path := sub.Path()
		if !fsutil.IsDir(path) {
			return fmt.Errorf("store %q: path %q does not exist or is not a directory", doctorStoreLabel(mp), path)
		}

		info, statErr := os.Stat(path)
		if statErr != nil {
			return fmt.Errorf("store %q: cannot stat path %q: %w", doctorStoreLabel(mp), path, statErr)
		}

		if info.Mode().Perm()&0o002 != 0 {
			return fmt.Errorf("store %q: path %q is world-writable (mode %04o)", doctorStoreLabel(mp), path, info.Mode().Perm())
		}
	}

	return nil
}

// doctorCheckRecipients fails if any store has invalid or expired recipient keys.
func (s *Action) doctorCheckRecipients(ctx context.Context) error {
	for _, mp := range s.doctorMountPoints() {
		if err := s.Store.CheckRecipients(ctx, mp); err != nil {
			return fmt.Errorf("store %q: %w", doctorStoreLabel(mp), err)
		}
	}

	return nil
}

// doctorCheckGitRemote warns if any git-backed store has no remote configured.
// It does not return an error because a local-only store without sync is valid.
func (s *Action) doctorCheckGitRemote(ctx context.Context) {
	for _, mp := range s.doctorMountPoints() {
		sub, err := s.Store.GetSubStore(mp)
		if err != nil || sub == nil {
			continue
		}

		g, ok := sub.Storage().(*gitfs.Git)
		if !ok {
			continue
		}

		v, err := g.ConfigGet(ctx, "remote.origin.url")
		if err != nil || v == "" {
			out.Warningf(ctx, "store %q has no git remote configured (sync will be unavailable)", doctorStoreLabel(mp))
		}
	}
}

// doctorStoreLabel returns a human-readable label for a store mount point.
func doctorStoreLabel(mp string) string {
	if mp == "" {
		return "<root>"
	}

	return mp
}
