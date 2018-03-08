package sub

import (
	"bufio"
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/backend"
	gpgmock "github.com/justwatchcom/gopass/backend/crypto/gpg/mock"
	"github.com/justwatchcom/gopass/backend/store/fs"
	gitmock "github.com/justwatchcom/gopass/backend/sync/git/mock"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestGetRecipientsDefault(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	genRecs, _, err := createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := &Store{
		alias:  "",
		url:    backend.FromPath(tempdir),
		crypto: gpgmock.New(),
		sync:   gitmock.New(),
		store:  fs.New(tempdir),
	}

	assert.Equal(t, genRecs, s.Recipients(ctx))
	recs, err := s.GetRecipients(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, genRecs, recs)
}

func TestGetRecipientsSubID(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	genRecs, _, err := createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := &Store{
		alias:  "",
		url:    backend.FromPath(tempdir),
		crypto: gpgmock.New(),
		sync:   gitmock.New(),
		store:  fs.New(tempdir),
	}

	recs, err := s.GetRecipients(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, genRecs, recs)

	err = ioutil.WriteFile(filepath.Join(tempdir, "foo", "bar", s.crypto.IDFile()), []byte("john.doe\n"), 0600)
	assert.NoError(t, err)

	recs, err = s.GetRecipients(ctx, "foo/bar/baz")
	assert.NoError(t, err)
	assert.Equal(t, []string{"john.doe"}, recs)
}

func TestSaveRecipients(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	_, _, err = createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	recp := []string{"john.doe"}
	s := &Store{
		alias:  "",
		url:    backend.FromPath(tempdir),
		crypto: gpgmock.New(),
		sync:   gitmock.New(),
		store:  fs.New(tempdir),
	}

	// remove recipients
	_ = os.Remove(filepath.Join(tempdir, s.crypto.IDFile()))

	assert.NoError(t, s.saveRecipients(ctx, recp, "test-save-recipients", true))
	assert.Error(t, s.saveRecipients(ctx, nil, "test-save-recipients", true))

	buf, err := s.store.Get(ctx, s.idFile(ctx, ""))
	assert.NoError(t, err)

	foundRecs := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		foundRecs = append(foundRecs, strings.TrimSpace(scanner.Text()))
	}
	sort.Strings(foundRecs)

	for i := 0; i < len(recp); i++ {
		if i >= len(foundRecs) {
			t.Errorf("Read too few recipients")
			break
		}
		if recp[i] != foundRecs[i] {
			t.Errorf("Mismatch at %d: %s vs %s", i, recp[i], foundRecs[i])
		}
	}
}

func TestAddRecipient(t *testing.T) {
	ctx := context.Background()
	ctx = out.WithHidden(ctx, true)

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	genRecs, _, err := createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	s := &Store{
		alias:  "",
		url:    backend.FromPath(tempdir),
		crypto: gpgmock.New(),
		sync:   gitmock.New(),
		store:  fs.New(tempdir),
	}

	newRecp := "A3683834"

	err = s.AddRecipient(ctx, newRecp)
	assert.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, append(genRecs, newRecp), rs)

	err = s.SaveRecipients(ctx)
	assert.NoError(t, err)
}

func TestRemoveRecipient(t *testing.T) {
	ctx := context.Background()
	ctx = out.WithHidden(ctx, true)

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	_, _, err = createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	s := &Store{
		alias:  "",
		url:    backend.FromPath(tempdir),
		crypto: gpgmock.New(),
		sync:   gitmock.New(),
		store:  fs.New(tempdir),
	}

	err = s.RemoveRecipient(ctx, "0xDEADBEEF")
	assert.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"0xFEEDBEEF"}, rs)
}

func TestListRecipients(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	genRecs, _, err := createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	obuf := &bytes.Buffer{}
	out.Stdout = obuf
	defer func() {
		out.Stdout = os.Stdout
	}()

	ctx = backend.WithCryptoBackendString(ctx, "gpgmock")
	ctx = backend.WithSyncBackendString(ctx, "gitmock")
	s, err := New(
		ctx,
		"",
		tempdir,
		tempdir,
	)
	assert.NoError(t, err)

	rs, err := s.GetRecipients(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, genRecs, rs)

	assert.Equal(t, "0xDEADBEEF", s.OurKeyID(ctx))
}
