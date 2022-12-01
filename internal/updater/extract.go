package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gopasspw/gopass/pkg/debug"
)

func extractFile(buf []byte, filename, dest string) error {
	mode := os.FileMode(0o755)
	dir := filepath.Dir(dest)

	// if overwriting an existing binary retain it's mode flags
	fi, err := os.Lstat(dest)
	if err == nil {
		mode = fi.Mode()
	}

	tfn, err := extractToTempFile(buf, filename, dest)
	if err != nil {
		return fmt.Errorf("failed to extract update to %s: %w", dest, err)
	}

	if err := removeOldBinary(dir, dest); err != nil {
		return fmt.Errorf("failed to remove old binary %s: %w", dest, err)
	}

	if err := os.Rename(tfn, dest); err != nil {
		return fmt.Errorf("failed to rename tempfile %s to %s: %w", tfn, dest, err)
	}

	return os.Chmod(dest, mode)
}

func extractToTempFile(buf []byte, filename, dest string) (string, error) {
	// open a temp file for writing
	dir := filepath.Dir(dest)
	dfh, err := ioutil.TempFile(dir, "gopass")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file in %s: %w", dir, err)
	}

	defer func() {
		_ = dfh.Sync()
		_ = dfh.Close()
	}()

	var rd io.Reader = bytes.NewReader(buf)

	switch filepath.Ext(filename) {
	case ".gz":
		gzr, err := gzip.NewReader(rd)
		if err != nil {
			return "", fmt.Errorf("failed to open gzip file: %w", err)
		}

		return extractTar(gzr, dfh, dfh.Name())
	case ".bz2":
		return extractTar(bzip2.NewReader(rd), dfh, dfh.Name())
	case ".zip":
		return extractZip(buf, dfh, dfh.Name())
	default:
		return "", fmt.Errorf("unsupported file extension: %q", filepath.Ext(filename))
	}
}

func extractZip(buf []byte, dfh io.WriteCloser, dest string) (string, error) {
	zrd, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}

	for i := 0; i < len(zrd.File); i++ {
		if zrd.File[i].Name != "gopass.exe" {
			continue
		}

		file, err := zrd.File[i].Open()
		if err != nil {
			return "", fmt.Errorf("failed to read from zip file: %w", err)
		}

		n, err := io.Copy(dfh, file)
		if err != nil {
			_ = dfh.Close()
			_ = os.Remove(dest)

			return "", fmt.Errorf("failed to read gopass.exe from zip file: %w", err)
		}
		// success
		debug.Log("wrote %d bytes to %v", n, dest)

		return dest, nil
	}

	return "", fmt.Errorf("file not found in archive")
}

func extractTar(rd io.Reader, dfh io.WriteCloser, dest string) (string, error) {
	tarReader := tar.NewReader(rd)

	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return "", fmt.Errorf("failed to read from tar file: %w", err)
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
			_ = dfh.Close()
			_ = os.Remove(dest)

			return "", fmt.Errorf("failed to read gopass from tar file: %w", err)
		}
		// success
		debug.Log("wrote %d bytes to %v", n, dest)

		return dest, nil
	}

	return "", fmt.Errorf("file not found in archive")
}
