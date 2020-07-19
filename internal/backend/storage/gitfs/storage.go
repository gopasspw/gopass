package gitfs

import (
	"context"
	"fmt"
)

// Get retrieves the named content
func (g *Git) Get(ctx context.Context, name string) ([]byte, error) {
	return g.fs.Get(ctx, name)
}

// Set writes the given content
func (g *Git) Set(ctx context.Context, name string, value []byte) error {
	return g.fs.Set(ctx, name, value)
}

// Delete removes the named entity
func (g *Git) Delete(ctx context.Context, name string) error {
	return g.fs.Delete(ctx, name)
}

// Exists checks if the named entity exists
func (g *Git) Exists(ctx context.Context, name string) bool {
	return g.fs.Exists(ctx, name)
}

// List returns a list of all entities
// e.g. foo, far/bar baz/.bang
// directory separator are normalized using `/`
func (g *Git) List(ctx context.Context, prefix string) ([]string, error) {
	return g.fs.List(ctx, prefix)
}

// IsDir returns true if the named entity is a directory
func (g *Git) IsDir(ctx context.Context, name string) bool {
	return g.fs.IsDir(ctx, name)
}

// Prune removes a named directory
func (g *Git) Prune(ctx context.Context, prefix string) error {
	return g.fs.Prune(ctx, prefix)
}

// String implements fmt.Stringer
func (g *Git) String() string {
	return fmt.Sprintf("gitfs(v0.1.0,path:%s)", g.fs.Path())
}

// Path returns the ondisk path
func (g *Git) Path() string {
	return g.fs.Path()
}

// Fsck checks the storage integrity
func (g *Git) Fsck(ctx context.Context) error {
	return g.fs.Fsck(ctx)
}
