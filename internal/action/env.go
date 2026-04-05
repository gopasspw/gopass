package action

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/gopasspw/gopass/internal/action/exit"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/tempfile"
	"github.com/urfave/cli/v2"
)

// Env implements the env subcommand. It populates the environment of a subprocess with
// a set of environment variables corresponding to the secret subtree specified on the
// command line.
func (s *Action) Env(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	name := c.Args().First()
	args := c.Args().Tail()
	keepCase := c.Bool("keep-case")
	useStdin := c.Bool("stdin")
	useFile := c.Bool("file")
	useExec := c.Bool("exec")

	if len(args) == 0 {
		return exit.Error(exit.Usage, nil, "Missing subcommand to execute")
	}

	// At most one mode flag may be active at a time.
	modeCount := 0
	if useStdin {
		modeCount++
	}
	if useFile {
		modeCount++
	}
	if useExec {
		modeCount++
	}

	if modeCount > 1 {
		return exit.Error(exit.Usage, nil, "Only one of --stdin, --file or --exec may be specified")
	}

	if !s.Store.Exists(ctx, name) && !s.Store.IsDir(ctx, name) {
		return exit.Error(exit.NotFound, nil, "Secret %s not found", name)
	}

	keys, err := s.envKeys(ctx, name)
	if err != nil {
		return err
	}

	if useStdin {
		return s.envRunStdin(ctx, name, keys, args)
	}

	if useFile {
		return s.envRunFile(ctx, name, keys, args, keepCase)
	}

	return s.envRunDefault(ctx, name, keys, args, keepCase, useExec)
}

// envKeys resolves the set of store paths to operate on for name. If name is a
// directory the full list of entries under it is returned; otherwise a
// single-element slice is returned.
func (s *Action) envKeys(ctx context.Context, name string) ([]string, error) {
	if !s.Store.IsDir(ctx, name) {
		return []string{name}, nil
	}

	debug.Log("%q is a dir, adding its entries", name)

	l, err := s.Store.Tree(ctx)
	if err != nil {
		return nil, exit.Error(exit.List, err, "failed to list store: %s", err)
	}

	subtree, err := l.FindFolder(name)
	if err != nil {
		return nil, exit.Error(exit.NotFound, nil, "Entry %q not found", name)
	}

	return subtree.List(tree.INF), nil
}

// envRunStdin runs args with the secret's password written to the subprocess's
// stdin. No environment variable is set. Only valid for a single secret.
func (s *Action) envRunStdin(ctx context.Context, name string, keys []string, args []string) error {
	if s.Store.IsDir(ctx, name) {
		return exit.Error(exit.Usage, nil, "--stdin requires a single secret, not a directory")
	}

	sec, err := s.Store.Get(ctx, keys[0])
	if err != nil {
		return fmt.Errorf("failed to get entry %q: %w", name, err)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stdin = strings.NewReader(sec.Password())
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// envRunFile writes each secret to a ramdisk temp file and exports
// KEY_FILE=/path/to/file in the subprocess environment. All temp files are
// removed when the subprocess exits.
func (s *Action) envRunFile(ctx context.Context, name string, keys []string, args []string, keepCase bool) error {
	tfs := make([]*tempfile.File, 0, len(keys))
	defer func() {
		for _, tf := range tfs {
			_ = tf.Remove(ctx)
		}
	}()

	fileEnv := make([]string, 0, len(keys))

	for _, key := range keys {
		envEntry, tf, err := s.envWriteTempFile(ctx, name, key, keepCase)
		if err != nil {
			return err
		}

		tfs = append(tfs, tf)
		fileEnv = append(fileEnv, envEntry)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = append(os.Environ(), fileEnv...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// envWriteTempFile writes a single secret to a ramdisk temp file and returns
// the "KEY_FILE=path" env string and the open file handle for later cleanup.
func (s *Action) envWriteTempFile(ctx context.Context, name, key string, keepCase bool) (string, *tempfile.File, error) {
	sec, err := s.Store.Get(ctx, key)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get entry for env prefix %q: %w", name, err)
	}

	tf, err := tempfile.New(ctx, "gopass-env-")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file for %q: %w", key, err)
	}

	if _, err := fmt.Fprint(tf, sec.Password()); err != nil {
		_ = tf.Remove(ctx)

		return "", nil, fmt.Errorf("failed to write temp file for %q: %w", key, err)
	}

	if err := tf.Close(); err != nil {
		_ = tf.Remove(ctx)

		return "", nil, fmt.Errorf("failed to close temp file for %q: %w", key, err)
	}

	envKey := path.Base(key)
	if !keepCase {
		envKey = strings.ToUpper(envKey)
	}

	return fmt.Sprintf("%s_FILE=%s", envKey, tf.Name()), tf, nil
}

// envRunDefault runs args as a child process (or replaces the current process
// when useExec is true) with the resolved secrets injected as KEY=value
// environment variables.
func (s *Action) envRunDefault(ctx context.Context, name string, keys []string, args []string, keepCase, useExec bool) error {
	env := make([]string, 0, len(keys))

	for _, key := range keys {
		debug.Log("exporting to environment key: %s", key)

		sec, err := s.Store.Get(ctx, key)
		if err != nil {
			return fmt.Errorf("failed to get entry for env prefix %q: %w", name, err)
		}

		envKey := path.Base(key)
		if !keepCase {
			envKey = strings.ToUpper(envKey)
		}

		env = append(env, fmt.Sprintf("%s=%s", envKey, sec.Password()))
	}

	if useExec {
		return execReplace(args, append(os.Environ(), env...))
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
