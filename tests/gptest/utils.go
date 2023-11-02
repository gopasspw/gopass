package gptest

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
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

// CliCtx create a new cli context with the given args parsed.
func CliCtx(ctx context.Context, t *testing.T, args ...string) *cli.Context {
	t.Helper()

	return CliCtxWithFlags(ctx, t, nil, args...)
}

// CliCtxWithFlags creates a new cli context with the given args and flags parsed.
func CliCtxWithFlags(ctx context.Context, t *testing.T, flags map[string]string, args ...string) *cli.Context {
	t.Helper()

	app := cli.NewApp()

	fs := flagset(t, flags, args)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	return c
}

func flagset(t *testing.T, flags map[string]string, args []string) *flag.FlagSet {
	t.Helper()

	fs := flag.NewFlagSet("default", flag.ContinueOnError)

	for k, v := range flags {
		if v == "true" || v == "false" {
			f := cli.BoolFlag{
				Name:  k,
				Usage: k,
			}
			require.NoError(t, f.Apply(fs))
		} else if _, err := strconv.Atoi(v); err == nil {
			f := cli.IntFlag{
				Name:  k,
				Usage: k,
			}
			require.NoError(t, f.Apply(fs))
		} else {
			f := cli.StringFlag{
				Name:  k,
				Usage: k,
			}
			require.NoError(t, f.Apply(fs))
		}
	}

	argl := []string{}
	for k, v := range flags {
		argl = append(argl, "--"+k+"="+v)
	}

	argl = append(argl, args...)
	require.NoError(t, fs.Parse(argl))

	return fs
}

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
