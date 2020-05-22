package dump

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testHibpSampleSorted = `000000005AD76BD555C1D6D771DE417A4B87E4B4
00000000A8DAE4228F821FB418F59826079BF368:42
00000000DD7F2A1C68A35673713783CA390C9E93:42
00000001E225B908BAC31C56DB04D892E47536E0:42
00000008CD1806EB7B9B46A8F87690B2AC16F617:42
0000000A0E3B9F25FF41DE4B5AC238C2D545C7A8:42
0000000A1D4B746FAA3FD526FF6D5BC8052FDB38:42
0000000CAEF405439D57847A8657218C618160B2:42
0000000FC1C08E6454BED24F463EA2129E254D43:42
00000010F4B38525354491E099EB1796278544B1`
const testHibpSampleUnsorted = `000000005AD76BD555C1D6D771DE417A4B87E4B4
00000000A8DAE4228F821FB418F59826079BF368:42
00000008CD1806EB7B9B46A8F87690B2AC16F617:42
0000000A0E3B9F25FF41DE4B5AC238C2D545C7A8:42
0000000A1D4B746FAA3FD526FF6D5BC8052FDB38:42
0000000CAEF405439D57847A8657218C618160B2:42
0000000FC1C08E6454BED24F463EA2129E254D43:42
00000000DD7F2A1C68A35673713783CA390C9E93:42
00000001E225B908BAC31C56DB04D892E47536E0:42
00000010F4B38525354491E099EB1796278544B1`

func Example() {
	ctx := context.Background()
	scanner, err := New("path/to/hibp-dump")
	if err != nil {
		panic(err)
	}
	matches := scanner.LookupBatch(ctx, []string{
		"list",
		"of",
		"sha1",
		"hashes",
	})
	fmt.Println(matches)
}

func TestScanner(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()

	// no hibp dump, no scanner
	_, err = New()
	assert.Error(t, err)

	// setup file and env (sorted)
	fn := filepath.Join(td, "dump.txt")
	assert.NoError(t, ioutil.WriteFile(fn, []byte(testHibpSampleSorted), 0644))

	scanner, err := New(fn)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, scanner.LookupBatch(ctx, []string{"foobar"}))

	// setup file and env (unsorted)
	fn = filepath.Join(td, "dump.txt")
	assert.NoError(t, ioutil.WriteFile(fn, []byte(testHibpSampleUnsorted), 0644))

	scanner, err = New(fn)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, scanner.LookupBatch(ctx, []string{"foobar"}))

	// gzip
	fn = filepath.Join(td, "dump.txt.gz")
	assert.NoError(t, testWriteGZ(fn, []byte(testHibpSampleSorted)))

	scanner, err = New(fn)
	assert.NoError(t, err)
	assert.Equal(t, []string{}, scanner.LookupBatch(ctx, []string{"foobar"}))
}

func testWriteGZ(fn string, buf []byte) error {
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		_ = fh.Close()
	}()

	gzw := gzip.NewWriter(fh)
	defer func() {
		_ = gzw.Close()
	}()

	_, err = gzw.Write(buf)
	return err
}
