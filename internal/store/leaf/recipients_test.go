package leaf

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/backend"
	plain "github.com/gopasspw/gopass/internal/backend/crypto/plain"
	"github.com/gopasspw/gopass/internal/backend/storage/fs"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRecipientsDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tempdir := t.TempDir()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	assert.Equal(t, genRecs, s.Recipients(ctx))
	recs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)

	ids := recs.IDs()
	sort.Strings(ids)
	assert.Equal(t, genRecs, ids)
}

func TestGetRecipientsSubID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tempdir := t.TempDir()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	recs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, genRecs, recs.IDs())

	err = os.WriteFile(filepath.Join(tempdir, "foo", "bar", s.crypto.IDFile()), []byte("john.doe\n"), 0o600)
	require.NoError(t, err)

	recs, err = s.GetRecipients(ctx, "foo/bar/baz")
	require.NoError(t, err)
	assert.Equal(t, []string{"john.doe"}, recs.IDs())
}

func TestSaveRecipients(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	// remove recipients
	_ = os.Remove(filepath.Join(tempdir, s.crypto.IDFile()))

	rs := recipients.New()
	rs.Add("john.doe")

	require.NoError(t, s.saveRecipients(ctx, rs, "test-save-recipients"))
	require.Error(t, s.saveRecipients(ctx, nil, "test-save-recipients"))

	buf, err := s.storage.Get(ctx, s.idFile(ctx, ""))
	require.NoError(t, err)

	foundRecs := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(buf))

	for scanner.Scan() {
		foundRecs = append(foundRecs, strings.TrimSpace(scanner.Text()))
	}

	sort.Strings(foundRecs)

	ids := rs.IDs()
	for i := 0; i < len(ids); i++ {
		if i >= len(foundRecs) {
			t.Errorf("Read too few recipients")

			break
		}

		if ids[i] != foundRecs[i] {
			t.Errorf("Mismatch at %d: %s vs %s", i, ids[i], foundRecs[i])
		}
	}
}

func TestAddRecipient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = ctxutil.WithHidden(ctx, true)

	tempdir := t.TempDir()

	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	newRecp := "A3683834"

	err = s.AddRecipient(ctx, newRecp)
	require.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, append(genRecs, newRecp), rs.IDs())

	err = s.SaveRecipients(ctx, false)
	require.NoError(t, err)
}

func TestRemoveRecipient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = ctxutil.WithHidden(ctx, true)

	tempdir := t.TempDir()

	_, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

	s := &Store{
		alias:   "",
		path:    tempdir,
		crypto:  plain.New(),
		storage: fs.New(tempdir),
	}

	err = s.RemoveRecipient(ctx, "0xDEADBEEF")
	require.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"0xFEEDBEEF"}, rs.IDs())
}

func TestListRecipients(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tempdir := t.TempDir()

	genRecs, _, err := createStore(tempdir, nil, nil)
	require.NoError(t, err)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

	ctx = backend.WithCryptoBackendString(ctx, "plain")
	s, err := New(
		ctx,
		"",
		tempdir,
	)
	require.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, genRecs, rs.IDs())

	assert.Equal(t, "0xDEADBEEF", s.OurKeyID(ctx))
}

func TestCheckRecipients(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test setup not supported on Windows")
	}

	u := gptest.NewGUnitTester(t)

	ctx := context.Background()
	ctx = ctxutil.WithTerminal(ctx, false)
	ctx = backend.WithCryptoBackend(ctx, backend.GPGCLI)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf

	defer func() {
		out.Stdout = os.Stdout
	}()

	s, err := New(ctx, "", u.StoreDir(""))
	require.NoError(t, err)

	require.NoError(t, s.CheckRecipients(ctx))

	u.AddExpiredRecipient()
	require.Error(t, s.CheckRecipients(ctx))
}
