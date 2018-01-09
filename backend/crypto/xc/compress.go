package xc

import (
	"bytes"
	"compress/gzip"
	"io"
)

func compress(in []byte) ([]byte, bool) {
	buf := &bytes.Buffer{}
	gzw, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
	if err != nil {
		return in, false
	}
	if _, err := gzw.Write(in); err != nil {
		return in, false
	}
	if err := gzw.Close(); err != nil {
		return in, false
	}
	if len(buf.Bytes()) >= len(in) {
		return in, false
	}
	return buf.Bytes(), true
}

func decompress(in []byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	gzr, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(buf, gzr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
