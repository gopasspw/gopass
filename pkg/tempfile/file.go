// Package tempfile is a wrapper around os.MkdirTemp, providing an OO pattern
// as well as secure placement on a temporary ramdisk.
package tempfile

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// globalPrefix is prefixed to all temporary dirs.
var globalPrefix string

// File is a temporary file.
type File struct {
	dir string
	dev string
	fh  *os.File
}

// New returns a new tempfile wrapper.
func New(ctx context.Context, prefix string) (*File, error) {
	td, err := os.MkdirTemp(tempdirBase(), globalPrefix+prefix)
	if err != nil {
		return nil, err
	}

	tf := &File{
		dir: td,
	}

	if err := tf.mount(ctx); err != nil {
		_ = os.RemoveAll(tf.dir)
		return nil, err
	}

	fn := filepath.Join(tf.dir, "secret")
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fn, err)
	}
	tf.fh = fh

	return tf, nil
}

// Name returns the name of the tempfile.
func (t *File) Name() string {
	if t.fh == nil {
		return ""
	}
	return t.fh.Name()
}

// Write implements io.Writer.
func (t *File) Write(p []byte) (int, error) {
	if t.fh == nil {
		return 0, fmt.Errorf("not initialized")
	}
	return t.fh.Write(p)
}

// Close implements io.WriteCloser.
func (t *File) Close() error {
	if t.fh == nil {
		return nil
	}
	return t.fh.Close()
}

// Remove attempts to remove the tempfile.
func (t *File) Remove(ctx context.Context) error {
	_ = t.Close()
	if err := t.unmount(ctx); err != nil {
		return fmt.Errorf("failed to unmount %s from %s: %w", t.dev, t.dir, err)
	}
	if t.dir == "" {
		return nil
	}
	return os.RemoveAll(t.dir)
}
