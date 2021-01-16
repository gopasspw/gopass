package updater

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/pkg/errors"
)

func extractFile(buf []byte, filename, dest string) error {
	var mode = os.FileMode(0755)

	// if overwriting an existing binary retain it's mode flags
	fi, err := os.Lstat(dest)
	if err == nil {
		mode = fi.Mode()
	}

	var rd io.Reader = bytes.NewReader(buf)
	switch filepath.Ext(filename) {
	case ".gz":
		gzr, err := gzip.NewReader(rd)
		if err != nil {
			return err
		}
		rd = gzr
	case ".bz2":
		rd = bzip2.NewReader(rd)
	case ".zip":
		return fmt.Errorf("zip archives are not supported, yet")
	}

	if err := os.Remove(dest); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to remove destination file: %q", err)
		}
	}

	dfh, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return errors.Wrapf(err, "Failed to open file: %s", dest)
	}
	defer func() {
		_ = dfh.Close()
	}()

	tarReader := tar.NewReader(rd)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrapf(err, "Failed to read from tar file")
		}
		name := filepath.Base(header.Name)
		if header.Typeflag == tar.TypeReg && name == "gopass" {
			n, err := io.Copy(dfh, tarReader)
			if err != nil {
				dfh.Close()
				os.Remove(dest)
				return errors.Wrapf(err, "Failed to read gopass from tar file")
			}
			// success
			debug.Log("wrote %d bytes to %v", n, dest)
			return nil
		}
	}
	return errors.Errorf("file not found in archive")
}
