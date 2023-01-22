package fossilfs

import (
	"context"
	"fmt"
)

// Get retrieves the named content.
func (f *Fossil) Get(ctx context.Context, name string) ([]byte, error) {
	return f.fs.Get(ctx, name)
}

// Set writes the given content.
func (f *Fossil) Set(ctx context.Context, name string, value []byte) error {
	return f.fs.Set(ctx, name, value)
}

// Delete removes the named entity.
func (f *Fossil) Delete(ctx context.Context, name string) error {
	return f.fs.Delete(ctx, name)
}

// Exists checks if the named entity exists.
func (f *Fossil) Exists(ctx context.Context, name string) bool {
	return f.fs.Exists(ctx, name)
}

// List returns a list of all entities
// e.g. foo, far/bar baz/.bang
// directory separator are normalized using `/`.
func (f *Fossil) List(ctx context.Context, prefix string) ([]string, error) {
	return f.fs.List(ctx, prefix)
}

// IsDir returns true if the named entity is a directory.
func (f *Fossil) IsDir(ctx context.Context, name string) bool {
	return f.fs.IsDir(ctx, name)
}

// Prune removes a named directory.
func (f *Fossil) Prune(ctx context.Context, prefix string) error {
	return f.fs.Prune(ctx, prefix)
}

// String implements fmt.Stringer.
func (f *Fossil) String() string {
	return fmt.Sprintf("fossilfs(%s,path:%s)", f.Version(context.TODO()).String(), f.fs.Path())
}

// Path returns the path to this storage.
func (f *Fossil) Path() string {
	return f.fs.Path()
}

// Fsck checks the storage integrity.
func (f *Fossil) Fsck(ctx context.Context) error {
	// ensure sane fossil config.
	if err := f.fixConfig(ctx); err != nil {
		return fmt.Errorf("failed to fix fossil config: %w", err)
	}

	return f.fs.Fsck(ctx)
}

// Link creates a symlink.
func (f *Fossil) Link(ctx context.Context, from, to string) error {
	return f.fs.Link(ctx, from, to)
}

// IsSymlink returns true if the file is symlink.
func (f *Fossil) IsSymlink(ctx context.Context, fn string) bool {
	return f.fs.IsSymlink(ctx, fn)
}

// Move moves from src to dst.
func (f *Fossil) Move(ctx context.Context, src, dst string, del bool) error {
	return f.fs.Move(ctx, src, dst, del)
}
