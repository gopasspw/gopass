// Package gptest contains test helpers for gopass, including
// creating temporary directories, setting up environment variables,
// and creating CLI contexts for testing.
package gptest

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/urfave/cli/v3"
)

// AllPathsToSlash converts a list of paths to their correct
// platform specific slash representation.
func AllPathsToSlash(paths []string) []string {
	r := make([]string, len(paths))
	for i, p := range paths {
		r[i] = filepath.ToSlash(p)
	}

	return r
}

func setupEnv(t *testing.T, em map[string]string) {
	t.Helper()

	for k, v := range em {
		t.Setenv(k, v)
	}
}

// CliCtx create a new cli command with the given args.
func CliCtx(ctx context.Context, t *testing.T, args ...string) *cli.Command {
	t.Helper()

	return CliCtxWithFlags(ctx, t, nil, args...)
}

// CliCtxWithFlags creates a new cli command with the given args and flags.
func CliCtxWithFlags(ctx context.Context, t *testing.T, flags map[string]string, args ...string) *cli.Command {
	t.Helper()

	// Build the flag definitions and arg list.
	allArgs := make([]string, 0, len(flags)+len(args))
	cliFlags := make([]cli.Flag, 0, len(flags))

	for k, v := range flags {
		if k == "clip" {
			cliFlags = append(cliFlags, &cli.GenericFlag{
				Name:  k,
				Usage: k,
				Value: &optionalIntValue{},
			})
		} else if v == "true" || v == "false" {
			cliFlags = append(cliFlags, &cli.BoolFlag{Name: k})
		} else if _, err := strconv.Atoi(v); err == nil {
			cliFlags = append(cliFlags, &cli.IntFlag{Name: k})
		} else {
			cliFlags = append(cliFlags, &cli.StringFlag{Name: k})
		}

		allArgs = append(allArgs, "--"+k+"="+v)
	}

	allArgs = append(allArgs, args...)

	var captured *cli.Command

	root := &cli.Command{
		Flags: cliFlags,
		Action: func(c context.Context, cmd *cli.Command) error {
			captured = cmd

			return nil
		},
	}

	_ = root.Run(ctx, append([]string{"test"}, allArgs...))

	if captured == nil {
		// Fallback: return a bare command with no parsed flags.
		return root
	}

	return captured
}

// optionalIntValue is a test-only flag.Value that mirrors
// action.OptionalInt for use in test flag sets. It implements
// IsBoolFlag so the flag parser treats it as a boolean when
// no value is provided via the = syntax.
type optionalIntValue struct {
	raw string
}

func (o *optionalIntValue) Set(s string) error {
	o.raw = s

	return nil
}

func (o *optionalIntValue) String() string {
	return o.raw
}

func (o *optionalIntValue) IsBoolFlag() bool { return true }

func (o *optionalIntValue) Get() any { return o.raw }

// UnsetVars will unset the specified env vars and return a restore func.
func UnsetVars(ls ...string) func() {
	old := make(map[string]string, len(ls))
	for _, k := range ls {
		old[k] = os.Getenv(k)
		_ = os.Unsetenv(k)
	}

	return func() {
		for k, v := range old {
			_ = os.Setenv(k, v)
		}
	}
}
