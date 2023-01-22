package gitfs

import (
	"context"
	"fmt"

	"github.com/gopasspw/gopass/pkg/debug"
)

// Get retrieves the named content.
func (g *Git) Get(ctx context.Context, name string) ([]byte, error) {
	return g.fs.Get(ctx, name)
}

// Set writes the given content.
func (g *Git) Set(ctx context.Context, name string, value []byte) error {
	return g.fs.Set(ctx, name, value)
}

// Delete removes the named entity.
func (g *Git) Delete(ctx context.Context, name string) error {
	return g.fs.Delete(ctx, name)
}

// Exists checks if the named entity exists.
func (g *Git) Exists(ctx context.Context, name string) bool {
	return g.fs.Exists(ctx, name)
}

// List returns a list of all entities
// e.g. foo, far/bar baz/.bang
// directory separator are normalized using `/`.
func (g *Git) List(ctx context.Context, prefix string) ([]string, error) {
	return g.fs.List(ctx, prefix)
}

// IsDir returns true if the named entity is a directory.
func (g *Git) IsDir(ctx context.Context, name string) bool {
	return g.fs.IsDir(ctx, name)
}

// Prune removes a named directory.
func (g *Git) Prune(ctx context.Context, prefix string) error {
	return g.fs.Prune(ctx, prefix)
}

// String implements fmt.Stringer.
func (g *Git) String() string {
	return fmt.Sprintf("gitfs(%s,path:%s)", g.Version(context.TODO()).String(), g.fs.Path())
}

// Path returns the path to this storage.
func (g *Git) Path() string {
	return g.fs.Path()
}

// Fsck checks the storage integrity.
func (g *Git) Fsck(ctx context.Context) error {
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

func (g *Git) addUntrackedFiles(ctx context.Context) error {
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

// Link creates a symlink.
func (g *Git) Link(ctx context.Context, from, to string) error {
	return g.fs.Link(ctx, from, to)
}

// IsSymlink returns true if the file is symlink.
func (g *Git) IsSymlink(ctx context.Context, fn string) bool {
	return g.fs.IsSymlink(ctx, fn)
}

// Move moves from src to dst.
func (g *Git) Move(ctx context.Context, src, dst string, del bool) error {
	return g.fs.Move(ctx, src, dst, del)
}
