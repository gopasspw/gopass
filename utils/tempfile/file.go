package tempfile

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/pkg/errors"
)

// File is a temporary file
type File struct {
	dir string
	dev string
	fh  *os.File
	dbg bool
}

// New returns a new tempfile wrapper
func New(ctx context.Context, prefix string) (*File, error) {
	td, err := ioutil.TempDir(tempdirBase(), prefix)
	if err != nil {
		return nil, err
	}

	tf := &File{
		dir: td,
		dbg: ctxutil.IsDebug(ctx),
	}

	if err := tf.mount(ctx); err != nil {
		_ = os.RemoveAll(tf.dir)
		return nil, err
	}

	fn := filepath.Join(tf.dir, "secret")
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, errors.Errorf("Failed to open file %s: %s", fn, err)
	}
	tf.fh = fh

	return tf, nil
}

// Name returns the name of the tempfile
func (t *File) Name() string {
	if t.fh == nil {
		return ""
	}
	return t.fh.Name()
}

// Write implement io.Writer
func (t *File) Write(p []byte) (int, error) {
	if t.fh == nil {
		return 0, errors.Errorf("not initialized")
	}
	return t.fh.Write(p)
}

// Close implements io.WriteCloser
func (t *File) Close() error {
	if t.fh == nil {
		return nil
	}
	return t.fh.Close()
}

// Remove attempts to remove the tempfile
func (t *File) Remove(ctx context.Context) error {
	_ = t.Close()
	if err := t.unmount(ctx); err != nil {
		return errors.Errorf("Failed to unmount %s from %s: %s", t.dev, t.dir, err)
	}
	if t.dir == "" {
		return nil
	}
	return os.RemoveAll(t.dir)
}
