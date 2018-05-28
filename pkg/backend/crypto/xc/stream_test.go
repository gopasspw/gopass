package xc

import (
	"bytes"
	"context"
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	stdbin "encoding/binary"

	"github.com/alecthomas/binary"
	"github.com/golang/protobuf/proto"
	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/keyring"
	"github.com/gopasspw/gopass/pkg/backend/crypto/xc/xcpb"
	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	ctx := context.Background()

	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()
	assert.NoError(t, os.Setenv("GOPASS_CONFIG", filepath.Join(td, ".gopass.yml")))
	assert.NoError(t, os.Setenv("GOPASS_HOMEDIR", td))

	plainFile := filepath.Join(td, "data.bin")
	cryptFile := plainFile + ".xc"
	plainAgain := plainFile + ".dec"

	plainSum := sha512.New()
	againSum := sha512.New()

	passphrase := "test"

	k1, err := keyring.GenerateKeypair(passphrase)
	assert.NoError(t, err)

	skr := keyring.NewSecring()
	assert.NoError(t, skr.Set(k1))

	pkr := keyring.NewPubring(skr)

	xc := &XC{
		pubring: pkr,
		secring: skr,
		client:  &fakeAgent{passphrase},
	}

	pfh, err := os.OpenFile(plainFile, os.O_CREATE|os.O_WRONLY, 0600)
	assert.NoError(t, err)

	p := make([]byte, 1024)
	written := 0
	for i := 0; i < 64*1024; i++ {
		n, _ := rand.Read(p)
		n, err := pfh.Write(p[:n])
		written += n
		assert.NoError(t, err)
		_, _ = plainSum.Write(p[:n])
	}
	// add some more bytes to force an uneven block boundary
	p = make([]byte, 10)
	rand.Read(p)
	_, err = pfh.Write(p)
	assert.NoError(t, err)
	_, _ = plainSum.Write(p)

	assert.NoError(t, pfh.Close())
	t.Logf("Wrote %d bytes", written)

	pfh, err = os.Open(plainFile)
	assert.NoError(t, err)

	cfh, err := os.OpenFile(cryptFile, os.O_CREATE|os.O_WRONLY, 0600)
	assert.NoError(t, err)

	err = xc.EncryptStream(ctx, pfh, []string{k1.Fingerprint()}, cfh)
	assert.NoError(t, err)

	assert.NoError(t, pfh.Close())
	assert.NoError(t, cfh.Close())

	cfh, err = os.Open(cryptFile)
	assert.NoError(t, err)

	pfh, err = os.OpenFile(plainAgain, os.O_CREATE|os.O_WRONLY, 0600)
	assert.NoError(t, err)

	// check decryption works and yields exactly the input
	err = xc.DecryptStream(ctx, cfh, pfh)
	assert.NoError(t, err)

	assert.NoError(t, cfh.Close())
	assert.NoError(t, pfh.Close())

	pfh, err = os.Open(plainAgain)
	assert.NoError(t, err)
	buf := make([]byte, 1024)
	for {
		n, err := pfh.Read(buf)
		_, _ = againSum.Write(buf[:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			assert.NoError(t, err)
		}
	}
	assert.NoError(t, pfh.Close())

	assert.Equal(t, fmt.Sprintf("%X", plainSum.Sum(nil)), fmt.Sprintf("%X", againSum.Sum(nil)))
}

func BenchmarkEncryptDecrypt(b *testing.B) {
	b.StopTimer()
	ctx := context.Background()
	passphrase := "test"

	k1, err := keyring.GenerateKeypair(passphrase)
	assert.NoError(b, err)

	skr := keyring.NewSecring()
	assert.NoError(b, skr.Set(k1))

	pkr := keyring.NewPubring(skr)

	xc := &XC{
		pubring: pkr,
		secring: skr,
		client:  &fakeAgent{passphrase},
	}

	plaintext := &bytes.Buffer{}
	for i := 0; i < 1024*1024; i++ {
		plaintext.WriteString("a")
	}
	plainagain := &bytes.Buffer{}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		err := xc.EncryptStream(ctx, bytes.NewReader(plaintext.Bytes()), []string{k1.Fingerprint()}, buf)
		assert.NoError(b, err)

		err = xc.DecryptStream(ctx, bytes.NewReader(buf.Bytes()), plainagain)
		assert.NoError(b, err)

		buf.Reset()
		plainagain.Reset()
	}
}

var benchInner = 1

func BenchmarkByteSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		data := []byte("foobar")

		for j := 0; j < benchInner; j++ {
			enc := binary.NewEncoder(buf)
			_ = enc.Encode(data)
		}
		dec := binary.NewDecoder(bytes.NewReader(buf.Bytes()))
		for {
			if err := dec.Decode(&data); err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("Error: %s", err)
			}
		}
	}
}

func BenchmarkEncodeByteSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		data := []byte("foobar")

		for j := 0; j < benchInner; j++ {
			enc := binary.NewEncoder(buf)
			_ = enc.Encode(data)
		}
	}
}

func BenchmarkDecodeByteSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buf := &bytes.Buffer{}
		data := []byte("foobar")

		for j := 0; j < benchInner; j++ {
			enc := binary.NewEncoder(buf)
			_ = enc.Encode(data)
		}
		b.StartTimer()
		dec := binary.NewDecoder(bytes.NewReader(buf.Bytes()))
		for {
			if err := dec.Decode(&data); err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("Error: %s", err)
			}
		}
	}
}

func BenchmarkChunk(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		data := &xcpb.Chunk{
			Body: []byte("foobar"),
		}

		for j := 0; j < benchInner; j++ {
			enc := binary.NewEncoder(buf)
			_ = enc.Encode(data)
		}
		dec := binary.NewDecoder(bytes.NewReader(buf.Bytes()))
		for {
			if err := dec.Decode(data); err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("Error: %s", err)
			}
		}
	}
}

type chunkEncoder struct {
	d *xcpb.Chunk
}

func (c *chunkEncoder) MarshalBinary() ([]byte, error) {
	return proto.Marshal(c.d)
}

func (c *chunkEncoder) UnmarshalBinary(data []byte) error {
	return proto.Unmarshal(data, c.d)
}

func BenchmarkChunkEncoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		data := &chunkEncoder{
			d: &xcpb.Chunk{Body: []byte("foobar")},
		}

		for j := 0; j < benchInner; j++ {
			enc := binary.NewEncoder(buf)
			_ = enc.Encode(data)
		}
		dec := binary.NewDecoder(bytes.NewReader(buf.Bytes()))
		for {
			if err := dec.Decode(data); err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("Error: %s", err)
			}
		}
	}
}

func BenchmarkProtoEncoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		data := &xcpb.Chunk{Body: []byte("foobar")}

		for j := 0; j < benchInner; j++ {
			b, _ := proto.Marshal(data)
			buf.Write(b)
			d2 := &xcpb.Chunk{}
			_ = proto.Unmarshal(b, d2)
		}
	}
}

func BenchmarkCustomEncoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		data := []byte("foobar")
		encbuf := make([]byte, 8)

		for j := 0; j < benchInner; j++ {
			l := stdbin.PutUvarint(encbuf, uint64(len(data)))
			buf.Write(encbuf[:l])
			buf.Write(data)
		}
		r := bytes.NewReader(buf.Bytes())
		for {
			l, err := stdbin.ReadUvarint(r)
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("read error: %s", err)
			}
			encbuf = make([]byte, l)
			_, err = r.Read(encbuf)
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("read error: %s", err)
			}
			if string(data) != string(encbuf) {
				b.Logf("invalid data: '%s' vs. '%s'", data, encbuf)
			}
		}
	}
}

func BenchmarkCustomDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		data := []byte("foobar")
		var encbuf []byte

		for j := 0; j < benchInner; j++ {
			enc := binary.NewEncoder(buf)
			_ = enc.Encode(data)
		}
		r := bytes.NewReader(buf.Bytes())
		for {
			l, err := stdbin.ReadUvarint(r)
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("read error: %s", err)
			}
			encbuf = make([]byte, l)
			_, err = r.Read(encbuf)
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("read error: %s", err)
			}
			if string(data) != string(encbuf) {
				b.Logf("invalid data: '%s' vs. '%s'", data, encbuf)
			}
		}
	}
}
