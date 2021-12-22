package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/debug"
)

func extractFile(buf []byte, filename, dest string) error {
	var mode = os.FileMode(0755)

	// if overwriting an existing binary retain it's mode flags
	fi, err := os.Lstat(dest)
	if err == nil {
		mode = fi.Mode()
	}

	if err := os.Remove(dest); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to remove destination file: %q", err)
		}
	}

	// open the destination file for writing
	dfh, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", dest, err)
	}
	defer func() {
		_ = dfh.Close()
	}()

	var rd io.Reader = bytes.NewReader(buf)
	switch filepath.Ext(filename) {
	case ".gz":
		gzr, err := gzip.NewReader(rd)
		if err != nil {
			return err
		}
		return extractTar(gzr, dfh, dest)
	case ".bz2":
		return extractTar(bzip2.NewReader(rd), dfh, dest)
	case ".zip":
		return extractZip(buf, dfh, dest)
	default:
		return fmt.Errorf("unsupported file extension: %q", filepath.Ext(filename))
	}
}

func extractZip(buf []byte, dfh io.WriteCloser, dest string) error {
	zrd, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return err
	}

	for i := 0; i < len(zrd.File); i++ {
		if zrd.File[i].Name != "gopass.exe" {
			continue
		}

		file, err := zrd.File[i].Open()
		if err != nil {
			return fmt.Errorf("failed to read from zip file: %w", err)
		}

		n, err := io.Copy(dfh, file)
		if err != nil {
			dfh.Close()
			os.Remove(dest)
			return fmt.Errorf("failed to read gopass.exe from zip file: %w", err)
		}
		// success
		debug.Log("wrote %d bytes to %v", n, dest)
		return nil
	}

	return fmt.Errorf("file not found in archive")
}

func extractTar(rd io.Reader, dfh io.WriteCloser, dest string) error {
	tarReader := tar.NewReader(rd)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read from tar file: %w", err)
		}
		name := filepath.Base(header.Name)
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if name != "gopass" {
			continue
		}

		n, err := io.Copy(dfh, tarReader)
		if err != nil {
			dfh.Close()
			os.Remove(dest)
			return fmt.Errorf("failed to read gopass from tar file: %w", err)
		}
		// success
		debug.Log("wrote %d bytes to %v", n, dest)
		return nil
	}
	return fmt.Errorf("file not found in archive")
}
