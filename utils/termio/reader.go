package termio

import (
	"bytes"
	"io"
)

// LineReader is an unbuffered line reader
type LineReader struct {
	r io.Reader
}

// NewReader creates a new line reader
func NewReader(r io.Reader) *LineReader {
	return &LineReader{r: r}
}

// Read implements io.Reader
func (lr LineReader) Read(p []byte) (int, error) {
	return lr.r.Read(p)
}

// ReadLine reads one line w/o buffering
func (lr LineReader) ReadLine() (string, error) {
	out := &bytes.Buffer{}
	buf := make([]byte, 1) // important: we must only read one byte at a time!
	for {
		n, err := lr.r.Read(buf)
		for i := 0; i < n; i++ {
			if buf[i] == '\n' {
				return out.String(), nil
			}
			// err is always nil
			_ = out.WriteByte(buf[i])
		}
		// Callers should always process the n > 0 bytes returned before considering the error err.
		if err != nil {
			if err == io.EOF {
				return out.String(), nil
			}
			return out.String(), err
		}
	}
}
