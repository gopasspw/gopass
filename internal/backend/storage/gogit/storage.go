package gogit

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Get retrieves the named content.
func (g *GoGit) Get(ctx context.Context, name string) ([]byte, error) {
	return g.fs.Get(ctx, name)
}

// Set writes the given content.
func (g *GoGit) Set(ctx context.Context, name string, value []byte) error {
	return g.fs.Set(ctx, name, value)
}

// Delete removes the named entity.
func (g *GoGit) Delete(ctx context.Context, name string) error {
	return g.fs.Delete(ctx, name)
}

// Exists checks if the named entity exists.
func (g *GoGit) Exists(ctx context.Context, name string) bool {
	return g.fs.Exists(ctx, name)
}

// List returns a list of all entities.
func (g *GoGit) List(ctx context.Context, prefix string) ([]string, error) {
	return g.fs.List(ctx, prefix)
}

// IsDir returns true if the named entity is a directory.
func (g *GoGit) IsDir(ctx context.Context, name string) bool {
	return g.fs.IsDir(ctx, name)
}

// Prune removes a named directory.
func (g *GoGit) Prune(ctx context.Context, prefix string) error {
	return g.fs.Prune(ctx, prefix)
}

// Link creates a symlink.
func (g *GoGit) Link(ctx context.Context, from, to string) error {
	return g.fs.Link(ctx, from, to)
}

// Move moves from src to dst.
func (g *GoGit) Move(ctx context.Context, src, dst string, del bool) error {
	return g.fs.Move(ctx, src, dst, del)
}

// LinkTarget returns the relative target of a symlinked secret.
func (g *GoGit) LinkTarget(ctx context.Context, name string) (string, bool, error) {
	return g.fs.LinkTarget(ctx, name)
}

// Fsck checks the storage integrity.
func (g *GoGit) Fsck(ctx context.Context) error {
	// ensure sane git config.
	if err := g.fixConfig(ctx); err != nil {
		return fmt.Errorf("failed to fix git config: %w", err)
	}

	// add any untracked files.
	if err := g.addUntrackedFiles(ctx); err != nil {
		return fmt.Errorf("failed to add untracked files: %w", err)
	}

	return g.fs.Fsck(ctx)
}

func (g *GoGit) addUntrackedFiles(ctx context.Context) error {
	ut := g.ListUntrackedFiles(ctx)
	if len(ut) < 1 && !g.HasStagedChanges(ctx) {
		debug.Log("no untracked or staged files found")

		return nil
	}

	debug.Log("untracked files found: %v", ut)
	if err := g.Add(ctx, ut...); err != nil {
		return fmt.Errorf("failed to add untracked files: %w", err)
	}

	return g.Commit(ctx, "fsck")
}
