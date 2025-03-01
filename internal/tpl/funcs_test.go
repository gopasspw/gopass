package tpl

import (
	"strings"
	"testing"
	"time"

	"github.com/gopasspw/gopass/internal/hashsum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMd5sum(t *testing.T) {
	result, err := md5sum()("test")
	require.NoError(t, err)
	assert.Equal(t, hashsum.MD5Hex("test"), result)
}

func TestSha1sum(t *testing.T) {
	result, err := sha1sum()("test")
	require.NoError(t, err)
	assert.Equal(t, hashsum.SHA1Hex("test"), result)
}

func TestSha256sum(t *testing.T) {
	result, err := sha256sum()("test")
	require.NoError(t, err)
	assert.Equal(t, hashsum.SHA256Hex("test"), result)
}

func TestSha512sum(t *testing.T) {
	result, err := sha512sum()("test")
	require.NoError(t, err)
	assert.Equal(t, hashsum.SHA512Hex("test"), result)
}

func TestBlake3sum(t *testing.T) {
	result, err := blake3sum()("test")
	require.NoError(t, err)
	assert.Equal(t, hashsum.Blake3Hex("test"), result)
}

func TestMd5cryptFunc(t *testing.T) {
	result, err := md5cryptFunc()("salt", "password")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "{MD5-CRYPT}"), result)
}

func TestSshaFunc(t *testing.T) {
	result, err := sshaFunc()("salt", "password")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "{SSHA}"), result)
}

func TestSsha256Func(t *testing.T) {
	result, err := ssha256Func()("salt", "password")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "{SSHA256}"), result)
}

func TestSsha512Func(t *testing.T) {
	result, err := ssha512Func()("salt", "password")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "{SSHA512}"), result)
}

func TestArgon2iFunc(t *testing.T) {
	result, err := argon2iFunc()("salt", "password")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "{ARGON2I}$argon2i$"), result)
}

func TestArgon2idFunc(t *testing.T) {
	result, err := argon2idFunc()("salt", "password")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "{ARGON2ID}$argon2id$"), result)
}

func TestBcryptFunc(t *testing.T) {
	result, err := bcryptFunc()("password")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(result, "{BLF-CRYPT}$2a$"))
}

func TestRoundDuration(t *testing.T) {
	assert.Equal(t, "1h", roundDuration(time.Hour))
	assert.Equal(t, "1m", roundDuration(time.Minute))
	assert.Equal(t, "1s", roundDuration(time.Second))
	assert.Equal(t, "1d", roundDuration(time.Hour*24))
	assert.Equal(t, "1mo", roundDuration(time.Hour*24*30))
	assert.Equal(t, "1y", roundDuration(time.Hour*24*365))
}

func TestDate(t *testing.T) {
	ts := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2023-10-01", date(ts))
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello...", truncate(8, "hello world"))
	assert.Equal(t, "hello", truncate(8, "hello"))
	assert.Equal(t, "...", truncate(0, "hello"))
	assert.Equal(t, "h...", truncate(1, "hello"))
	assert.Equal(t, "he...", truncate(2, "hello"))
	assert.Equal(t, "hel...", truncate(3, "hello"))
	assert.Equal(t, "hell...", truncate(4, "hello"))
	assert.Equal(t, "hel", truncate(3, "hel"))
	assert.Equal(t, "he...", truncate(2, "hel"))
}

func TestJoin(t *testing.T) {
	assert.Equal(t, "a,b,c", join(",", []string{"a", "b", "c"}))
	assert.Equal(t, "1,2,3", join(",", []int{1, 2, 3}))
}
