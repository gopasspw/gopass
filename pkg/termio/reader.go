package termio

import (
	"bytes"
	"context"
	"io"
)

// LineReader is an unbuffered line reader.
type LineReader struct {
	r   io.Reader
	ctx context.Context
}

// NewReader creates a new line reader.
func NewReader(ctx context.Context, r io.Reader) *LineReader {
	return &LineReader{r: r, ctx: ctx}
}

// Read implements io.Reader.
func (lr LineReader) Read(p []byte) (int, error) {
	return lr.r.Read(p)
}

// rr is a composite value to transport the result of Read through a channel.
type rr struct {
	n   int
	err error
}

// ReadLine reads one line w/o buffering.
func (lr LineReader) ReadLine() (string, error) {
	out := &bytes.Buffer{}
	buf := make([]byte, 1) // important: we must only read one byte at a time!
	for {
		// we wait for the user input in the background so we can use the
		// select statement below to be able to immediately quit when the
		// user presses Ctrl+C
		msg := make(chan rr, 1)
		go func() {
			n, err := lr.r.Read(buf)
			msg <- rr{n, err}
		}()

		var n int
		var err error
		// wait for a user input (or a signal to abort)
		select {
		case <-lr.ctx.Done():
			return "", ErrAborted
		case s := <-msg:
			n = s.n
			err = s.err
		}

		// process the user input
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
