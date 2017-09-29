package fsutil

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type tempfile struct {
	dir string
	dev string
	fh  *os.File
	dbg bool
}

// TempFiler is a tempfile interface
type TempFiler interface {
	io.WriteCloser
	Name() string
	Remove(context.Context) error
}

// TempFile returns a new tempfile wrapper
func TempFile(ctx context.Context, prefix string) (TempFiler, error) {
	td, err := ioutil.TempDir(tempdirBase(), prefix)
	if err != nil {
		return nil, err
	}
	tf := &tempfile{
		dir: td,
	}
	if gdb := os.Getenv("GOPASS_DEBUG"); gdb == "true" {
		tf.dbg = true
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

func (t *tempfile) Name() string {
	if t.fh == nil {
		return ""
	}
	return t.fh.Name()
}

func (t *tempfile) Write(p []byte) (int, error) {
	if t.fh == nil {
		return 0, errors.Errorf("not initialized")
	}
	return t.fh.Write(p)
}

func (t *tempfile) Close() error {
	if t.fh == nil {
		return nil
	}
	return t.fh.Close()
}

func (t *tempfile) Remove(ctx context.Context) error {
	_ = t.Close()
	if err := t.unmount(ctx); err != nil {
		return errors.Errorf("Failed to unmount %s from %s: %s", t.dev, t.dir, err)
	}
	return os.RemoveAll(t.dir)
}
