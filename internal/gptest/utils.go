package gptest

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

// AllPathsToSlash converts a list of paths to their correct
// platform specific slash representation
func AllPathsToSlash(paths []string) []string {
	r := make([]string, len(paths))
	for i, p := range paths {
		r[i] = filepath.ToSlash(p)
	}
	return r
}

func setupEnv(em map[string]string) error {
	for k, v := range em {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func teardownEnv(em map[string]string) {
	for k := range em {
		_ = os.Unsetenv(k)
	}
}

// CliCtx create a new cli context with the given args parsed
func CliCtx(ctx context.Context, t *testing.T, args ...string) *cli.Context {
	return CliCtxWithFlags(ctx, t, nil, args...)
}

// CliCtxWithFlags creates a new cli context with the given args and flags parsed
func CliCtxWithFlags(ctx context.Context, t *testing.T, flags map[string]string, args ...string) *cli.Context {
	app := cli.NewApp()

	fs := flagset(t, flags, args)
	c := cli.NewContext(app, fs, nil)
	c.Context = ctx

	return c
}

func flagset(t *testing.T, flags map[string]string, args []string) *flag.FlagSet {
	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	for k, v := range flags {
		if v == "true" || v == "false" {
			f := cli.BoolFlag{
				Name:  k,
				Usage: k,
			}
			assert.NoError(t, f.Apply(fs))
		} else {
			f := cli.StringFlag{
				Name:  k,
				Usage: k,
			}
			assert.NoError(t, f.Apply(fs))
		}
	}
	argl := []string{}
	for k, v := range flags {
		argl = append(argl, "--"+k+"="+v)
	}
	argl = append(argl, args...)
	assert.NoError(t, fs.Parse(argl))

	return fs
}
