package xc

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"testing"

	"github.com/gopasspw/gopass/pkg/pwgen"
	"github.com/gopasspw/gopass/pkg/pwgen/xkcdgen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressPlain(t *testing.T) {
	t.Skip("not needed right now")

	for _, pwg := range []func(n int) string{
		func(n int) string { return pwgen.GeneratePasswordCharset(n+1, pwgen.CharAll) },
		func(n int) string {
			pw, _ := xkcdgen.RandomLength(n, "en")
			return pw
		},
	} {
		for i := 0; i < 1024; i++ {
			pw := pwg(i)
			buf := &bytes.Buffer{}
			gzw, err := gzip.NewWriterLevel(buf, gzip.BestCompression)
			require.NoError(t, err)
			_, _ = gzw.Write([]byte(pw))
			assert.NoError(t, gzw.Close())
			gzr, err := gzip.NewReader(bytes.NewReader(buf.Bytes()))
			require.NoError(t, err)
			out := &bytes.Buffer{}
			_, err = io.Copy(out, gzr)
			assert.NoError(t, err)
			assert.Equal(t, pw, out.String())
			t.Logf("len(raw): %d - len(gzip): %d - len(raw) < len(gzip): %t", len(pw), len(buf.Bytes()), len(pw) < len(buf.Bytes()))
		}
	}
}

func TestCompress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	for _, pwg := range []func(n int) string{
		func(n int) string { return pwgen.GeneratePasswordCharset(n+1, pwgen.CharAll) },
		func(n int) string {
			pw, _ := xkcdgen.RandomLength(n, "en")
			return pw
		},
	} {
		pwg := pwg // capture range variable
		t.Run(fmt.Sprintf("%p", pwg), func(t *testing.T) {
			t.Parallel()
			for i := 256; i < 512; i++ {
				pw := pwg(i)
				compPlain, compressed := compress([]byte(pw))
				decompPlain := []byte(pw)
				if compressed {
					var err error
					decompPlain, err = decompress(compPlain)
					assert.NoError(t, err)
				}
				assert.True(t, len(compPlain) <= len([]byte(pw)))
				assert.Equal(t, pw, string(decompPlain))
			}
		})
	}
}
