package fossilfs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/debug"
)

func (f *Fossil) fixConfig(ctx context.Context) error {
	// enable autosync
	if err := f.ConfigSet(ctx, "autosync", "1"); err != nil {
		return fmt.Errorf("failed to set fossil config autosync: %w", err)
	}

	// binary-glob
	if err := f.ConfigSet(ctx, "binary-glob", "*.age,*.gpg"); err != nil {
		return fmt.Errorf("failed to set fossil config binary-glob: %w", err)
	}

	return nil
}

func (f *Fossil) InitConfig(ctx context.Context, _, _ string) error {
	// ensure a sane fossil config.
	if err := f.fixConfig(ctx); err != nil {
		return fmt.Errorf("failed to fix fossil config: %w", err)
	}

	return nil
}

// ConfigSet sets a local config value.
func (f *Fossil) ConfigSet(ctx context.Context, key, value string) error {
	return f.Cmd(ctx, "fossilConfigSet", "settings", "--exact", key, value)
}

// ConfigGet returns a given config value.
func (f *Fossil) ConfigGet(ctx context.Context, key string) (string, error) {
	if !f.IsInitialized() {
		return "", store.ErrGitNotInit
	}

	buf := &strings.Builder{}

	cmd := exec.CommandContext(ctx, "fossil", "settings", "--exact", key)
	cmd.Dir = f.fs.Path()
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	sv := strings.Fields(strings.TrimSpace(buf.String()))
	return sv[len(sv)-1], nil
}

// ConfigList returns all fossil config settings.
func (f *Fossil) ConfigList(ctx context.Context) (map[string]string, error) {
	if !f.IsInitialized() {
		return nil, store.ErrGitNotInit
	}

	buf := &strings.Builder{}

	cmd := exec.CommandContext(ctx, "fossil", "settings")
	cmd.Dir = f.fs.Path()
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr

	debug.Log("%s %+v", cmd.Path, cmd.Args)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(buf.String(), "\n")
	kv := make(map[string]string, len(lines))
	for _, line := range lines {
		p := strings.Fields(strings.TrimSpace(line))
		// only record settings with a value
		if len(p) < 2 {
			continue
		}
		kv[p[0]] = p[len(p)-1]
	}
	return kv, nil
}
